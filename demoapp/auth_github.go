package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	"golang.org/x/oauth2"
)

// GitHub endpoints are fixed (GitHub does not implement OIDC discovery),
// so they live here as constants rather than being discovered at boot.
const (
	githubAuthURL   = "https://github.com/login/oauth/authorize"
	githubTokenURL  = "https://github.com/login/oauth/access_token"
	githubUserURL   = "https://api.github.com/user"
	githubEmailsURL = "https://api.github.com/user/emails"
)

// GitHubProvider is the demoapp's GitHub OAuth client. GitHub is plain
// OAuth 2.0 (no id_token), so unlike OIDCRelyingParty there is no
// discovery step and no ID-token verifier — identity comes from the
// /user API after the code exchange.
type GitHubProvider struct {
	OAuth *oauth2.Config
}

// NewGitHubProvider builds the oauth2.Config from a single set of fixed
// GitHub endpoints. Scopes request read access to the user's profile and
// email: read:user covers id/login/name, user:email is required for the
// /user/emails fallback when the primary email is private.
func NewGitHubProvider(oc *OAuthConfig) (*GitHubProvider, error) {
	if oc.ClientID == "" || oc.ClientSecret == "" || oc.RedirectURL == "" {
		return nil, errors.New("GitHub OAuth block is Enabled but missing ClientID/ClientSecret/RedirectURL")
	}
	cfg := &oauth2.Config{
		ClientID:     oc.ClientID,
		ClientSecret: oc.ClientSecret,
		RedirectURL:  oc.RedirectURL,
		Endpoint: oauth2.Endpoint{
			AuthURL:  githubAuthURL,
			TokenURL: githubTokenURL,
		},
		Scopes: []string{"read:user", "user:email"},
	}
	return &GitHubProvider{OAuth: cfg}, nil
}

// githubUser is the subset of GET /user fields we need. Email is nil when
// the user's primary email is private, in which case we fall back to
// /user/emails.
type githubUser struct {
	ID    int64  `json:"id"`
	Login string `json:"login"`
	Name  string `json:"name"`
	Email string `json:"email"`
}

// githubEmail is one entry from GET /user/emails. GitHub returns a list;
// we pick the primary+verified one.
type githubEmail struct {
	Email    string `json:"email"`
	Primary  bool   `json:"primary"`
	Verified bool   `json:"verified"`
}

// handleGitHubLogin initiates the OAuth code flow. A random state is stored
// in the session and verified on callback to prevent CSRF. The browser is
// redirected to GitHub's authorize endpoint.
func (a *App) handleGitHubLogin(c *gin.Context) {
	if a.GitHub == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "GitHub login not enabled"})
		return
	}
	state, err := randomToken()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "state gen failed"})
		return
	}
	s := sessions.Default(c)
	s.Set(sessKeyOAuthState, state)
	if err := s.Save(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "session save failed"})
		return
	}
	url := a.GitHub.OAuth.AuthCodeURL(state)
	c.Redirect(http.StatusTemporaryRedirect, url)
}

// handleGitHubCallback completes the OAuth code flow. It validates state,
// exchanges the code for an access token, fetches the user's GitHub identity
// (falling back to /user/emails when the primary email is private), and
// upserts the user via the shared external-IdP merge logic.
func (a *App) handleGitHubCallback(c *gin.Context) {
	if a.GitHub == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "GitHub login not enabled"})
		return
	}
	s := sessions.Default(c)
	stateRaw := s.Get(sessKeyOAuthState)
	clearSession(s) // always clear the OAuth state bit, even on error
	if stateRaw == nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "missing OAuth state"})
		return
	}
	state, _ := stateRaw.(string)
	if c.Query("state") != state {
		c.JSON(http.StatusBadRequest, gin.H{"error": "OAuth state mismatch"})
		return
	}
	code := c.Query("code")
	if code == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "missing code"})
		return
	}

	ctx := contextOf(c)
	tok, err := a.GitHub.OAuth.Exchange(ctx, code)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "token exchange failed: " + err.Error()})
		return
	}

	ghUser, err := fetchGitHubUser(ctx, tok)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "github user fetch failed: " + err.Error()})
		return
	}

	// GitHub's stable subject is the numeric account id (login can be
	// renamed). Stringify so it slots into the OIDCSubject column used by
	// the shared upsert path.
	subject := strconv.FormatInt(ghUser.ID, 10)
	email := strings.ToLower(strings.TrimSpace(ghUser.Email))
	if email == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "GitHub account has no usable email"})
		return
	}

	user, err := a.upsertExternalUser(ctx, subject, email, "github")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "user upsert failed: " + err.Error()})
		return
	}

	s.Set(sessKeyUserID, user.ID)
	s.Set(sessKeyUsername, user.Username)
	s.Set(sessKeyOIDCSubject, user.OIDCSubject)
	if err := s.Save(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "session save failed"})
		return
	}

	// Redirect back to the SPA; the me/config endpoints pick up the
	// session cookie automatically.
	c.Redirect(http.StatusTemporaryRedirect, "/")
}

// fetchGitHubUser calls GET /user with the access token and, when the
// primary email is absent (private email), falls back to GET /user/emails
// to pick the primary+verified address. A 10s timeout guards both calls so
// a stalled GitHub API does not hang the callback.
func fetchGitHubUser(ctx context.Context, tok *oauth2.Token) (*githubUser, error) {
	reqCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	client := oauth2.NewClient(reqCtx, oauth2.StaticTokenSource(tok))

	req, err := http.NewRequestWithContext(reqCtx, http.MethodGet, githubUserURL, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Accept", "application/vnd.github+json")
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(io.LimitReader(resp.Body, 512))
		return nil, fmt.Errorf("/user returned %d: %s", resp.StatusCode, strings.TrimSpace(string(body)))
	}
	var u githubUser
	if err := json.NewDecoder(resp.Body).Decode(&u); err != nil {
		return nil, fmt.Errorf("decode /user: %w", err)
	}
	if u.ID == 0 {
		return nil, errors.New("github user response missing id")
	}
	if u.Email != "" {
		return &u, nil
	}

	// Primary email is private — fall back to /user/emails.
	req, err = http.NewRequestWithContext(reqCtx, http.MethodGet, githubEmailsURL, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Accept", "application/vnd.github+json")
	resp, err = client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(io.LimitReader(resp.Body, 512))
		return nil, fmt.Errorf("/user/emails returned %d: %s", resp.StatusCode, strings.TrimSpace(string(body)))
	}
	var emails []githubEmail
	if err := json.NewDecoder(resp.Body).Decode(&emails); err != nil {
		return nil, fmt.Errorf("decode /user/emails: %w", err)
	}
	for _, e := range emails {
		if e.Primary && e.Verified {
			u.Email = e.Email
			return &u, nil
		}
	}
	// No primary+verified address; accept any verified one as a last resort
	// so the user is not blocked, but surface the compromise via the missing
	// email check at the call site.
	for _, e := range emails {
		if e.Verified {
			u.Email = e.Email
			return &u, nil
		}
	}
	return &u, nil
}
