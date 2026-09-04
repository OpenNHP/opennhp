package main

import (
	"context"
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/coreos/go-oidc/v3/oidc"
	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	"golang.org/x/oauth2"
)

// OIDCRelyingParty is the Demo app's application-side OIDC client. It is
// independent of the nhp-server ASP plugin (which uses the basic plugin
// for NHP-side OTP/REG/LST); the demo owns its own user identity.
type OIDCRelyingParty struct {
	Provider *oidc.Provider
	Verifier *oidc.IDTokenVerifier
	OAuth    *oauth2.Config
	Issuer   string
}

// NewOIDCRelyingParty discovers the IdP via the issuer URL, builds the
// oauth2.Config and returns a verifier ready to validate ID tokens.
func NewOIDCRelyingParty(oc *OIDCConfig) (*OIDCRelyingParty, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	provider, err := oidc.NewProvider(ctx, oc.IssuerURL)
	if err != nil {
		return nil, err
	}
	cfg := &oauth2.Config{
		ClientID:     oc.ClientID,
		ClientSecret: oc.ClientSecret,
		RedirectURL:  oc.RedirectURL,
		Endpoint:     provider.Endpoint(),
		Scopes:       []string{oidc.ScopeOpenID, "profile", "email"},
	}
	verifier := provider.Verifier(&oidc.Config{ClientID: oc.ClientID})
	return &OIDCRelyingParty{
		Provider: provider,
		Verifier: verifier,
		OAuth:    cfg,
		Issuer:   oc.IssuerURL,
	}, nil
}

// handleOIDCLogin initiates the OIDC code flow. State is generated and
// stored in the session; on callback we verify it matches before doing
// anything dangerous. The browser is then redirected to the IdP's
// authorize URL.
func (a *App) handleOIDCLogin(c *gin.Context) {
	if a.OIDCVefifier == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "OIDC not enabled"})
		return
	}
	state, err := randomToken()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "state gen failed"})
		return
	}
	nonce, err := randomToken()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "nonce gen failed"})
		return
	}
	s := sessions.Default(c)
	s.Set(sessKeyOIDCState, state)
	s.Set(sessKeyOIDCNonce, nonce)
	if err := s.Save(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "session save failed"})
		return
	}
	url := a.OIDCVefifier.OAuth.AuthCodeURL(state, oidc.Nonce(nonce))
	c.Redirect(http.StatusTemporaryRedirect, url)
}

// handleOIDCCallback completes the OIDC code flow. It validates state,
// exchanges the code for tokens, verifies the ID token, and either
// links the OIDC subject to an existing password user with the same
// email or creates a new OIDC-only user.
func (a *App) handleOIDCCallback(c *gin.Context) {
	if a.OIDCVefifier == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "OIDC not enabled"})
		return
	}
	s := sessions.Default(c)
	stateRaw := s.Get(sessKeyOIDCState)
	nonceRaw := s.Get(sessKeyOIDCNonce)
	clearSession(s) // always clear OIDC session bits, even on error
	if stateRaw == nil || nonceRaw == nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "missing OIDC state"})
		return
	}
	state, _ := stateRaw.(string)
	nonce, _ := nonceRaw.(string)
	if c.Query("state") != state {
		c.JSON(http.StatusBadRequest, gin.H{"error": "OIDC state mismatch"})
		return
	}
	code := c.Query("code")
	if code == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "missing code"})
		return
	}

	ctx := contextOf(c)
	tok, err := a.OIDCVefifier.OAuth.Exchange(ctx, code)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "token exchange failed: " + err.Error()})
		return
	}
	// Extra returns any; a token response without id_token must not panic
	// the (unauthenticated) callback. Use the comma-ok form and 401.
	idTokStr, ok := tok.Extra("id_token").(string)
	if !ok || idTokStr == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "token response missing id_token"})
		return
	}
	idTok, err := a.OIDCVefifier.Verifier.Verify(ctx, idTokStr)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "id token verification failed: " + err.Error()})
		return
	}
	if idTok.Nonce != nonce {
		c.JSON(http.StatusBadRequest, gin.H{"error": "OIDC nonce mismatch"})
		return
	}

	var claims struct {
		Sub           string `json:"sub"`
		Email         string `json:"email"`
		EmailVerified bool   `json:"email_verified"`
	}
	if err := idTok.Claims(&claims); err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "id token claims parse failed"})
		return
	}
	if claims.Sub == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "id token missing sub"})
		return
	}

	email := strings.ToLower(strings.TrimSpace(claims.Email))
	if email == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "id token missing email claim"})
		return
	}
	// Require a verified email before we trust it for account merge.
	// Some IdPs permit an arbitrary/unverified email claim; feeding such
	// a claim into upsertExternalUser's email-merge path lets an attacker
	// take over a victim's local account. GitHub filters on verified
	// upstream; the OIDC path must do the same here. Absent claim is
	// treated as unverified (safe default).
	if !claims.EmailVerified {
		c.JSON(http.StatusForbidden, gin.H{"error": "id token email not verified"})
		return
	}

	user, err := a.upsertExternalUser(ctx, claims.Sub, email, "oidc")
	if err != nil {
		if isUpsertConflict(err) {
			c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "user upsert failed: " + err.Error()})
		return
	}

	s.Set(sessKeyUserID, user.ID)
	if err := s.Save(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "session save failed"})
		return
	}

	// Redirect back to the SPA. The SPA's me/config endpoints will pick
	// up the session cookie automatically.
	c.Redirect(http.StatusTemporaryRedirect, "/")
}

// upsertExternalUser implements the external-IdP (OIDC or OAuth) to
// password user merge logic. It is called from the OIDC callback and
// the GitHub OAuth callback with the IdP's stable subject + email.
// provider names the IdP ("oidc" or "github") and is stamped on newly
// created rows so the post-login UI can show the account's origin.
//
//  1. If a user already has this external subject, return it. Its
//     auth_provider is left untouched (it reflects how the account was
//     originally created).
//  2. If a row already has this verified email — whether a password
//     holder or an IdP-only row from a different provider — SKIP it
//     and fall through to step 3. Silently merging onto a password
//     holder would let the squatter harvest a key the genuine user
//     later registers; silently merging onto a cross-provider IdP
//     row would let an early GitHub user take over a later OIDC user
//     (or vice versa) just by guessing the email. The earlier
//     errEmailLinkedToDifferentSubject guard rejected the cross-
//     provider case outright, which let either user permanently lock
//     the other out — same class of indefinite lockout review #2
//     removed for password squats. Falling through creates a separate
//     IdP-only account for the genuine IdP user; the email UNIQUE
//     constraint was relaxed in userstore.dropEmailUnique so the
//     rows can coexist. The squatter's row is untouched.
//  3. Otherwise (no email match) create a new password-less user with
//     the external subject, the given provider, and freshly generated
//     NHP keys, in the pending state. External-IdP users skip the
//     password step but NOT the NHP_REG handshake — the public key
//     must still be registered with nhp-server. They land on the
//     complete-registration view, which runs the handshake via
//     /api/credentials and flips them active.
//
// Sentinel errors (callers map these to 409, not 500):
//   - ErrUsernameTaken / ErrSubjectTaken / ErrEmailTaken: the
//     database-level UNIQUE constraints that fire on the insert
//     when the namespaced external username/oidc_subject/email
//     collides with an existing row.
func (a *App) upsertExternalUser(ctx context.Context, subject, email, provider string) (*User, error) {
	// Namespace the subject by provider so a GitHub numeric id and an
	// OIDC sub that happen to share the same string cannot resolve to
	// the same row (cross-provider takeover — review #5).
	key := externalSubjectKey(provider, subject)
	if u, err := a.Store.GetUserByOIDCSubject(ctx, provider, subject); err == nil {
		return u, nil
	} else if !errors.Is(err, ErrUserNotFound) {
		return nil, err
	}

	// Step 2: an existing row with the same verified email is left
	// alone — fall through to step 3 so the new-user branch creates a
	// SEPARATE IdP-only account. We don't merge onto a password-holder
	// (review #2: the password squat row is unverified, so the genuine
	// IdP user deserves their own account, and absorbing the squat
	// row's identity would let the squatter harvest a key the genuine
	// user later registers). We also don't merge onto an IdP-only row
	// from a DIFFERENT provider: every IdP-only row the app produced
	// has OIDCSubject == key, so this branch's "u.OIDCSubject != key"
	// guard was always true for cross-provider lookups — meaning the
	// merge body was unreachable for any row we made. We now drop the
	// branch entirely so a GitHub user with verified email v@x is no
	// longer permanently locked out when an OIDC user later signs in
	// for the same email (and vice versa). Each gets their own row;
	// the email UNIQUE constraint was relaxed in
	// userstore.dropEmailUnique so the rows can coexist.
	if _, err := a.Store.GetUserByEmail(ctx, email); err == nil {
		// fall through to step 3
	} else if !errors.Is(err, ErrUserNotFound) {
		return nil, err
	}

	// New OIDC user — no password, freshly generated NHP keys. OIDC has no
	// interactive scheme/server selection at callback time, so bind to the
	// default server (Servers[0]) under its relay-registered scheme. The
	// user can still switch via the complete-registration view later
	// (private key is scheme-agnostic, only the public key is re-derived).
	priv, pub, serverName, scheme, err := a.generateDefaultNHPKeyMaterial()
	if err != nil {
		return nil, err
	}
	deviceID, err := randomDeviceID()
	if err != nil {
		return nil, err
	}
	privEnc, err := sealKey(a.MasterKey, priv)
	if err != nil {
		return nil, err
	}
	u := &User{
		// External users identify by email, but email is a free-form value
		// a password user can register as their username. Setting the
		// password-style Username to the victim's email would collide
		// with such a squatter on the UNIQUE username column and 500 the
		// IdP sign-in — a trivially-triggered lockout (review #3). Use
		// the provider-namespaced subject as the username instead: it is
		// unique by construction, "|" is reserved for external identities
		// (handleRegister rejects it for password users), and the IdP
		// subject is unknowable to a squatter so it cannot be targeted.
		Username:         key,
		Email:            email,
		OIDCSubject:      key, // provider-namespaced (review #5)
		NHPPublicKey:     pub,
		NHPPrivateKeyEnc: privEnc,
		NHPDeviceID:      deviceID,
		ServerName:       serverName,
		CipherScheme:     scheme,
		AuthProvider:     provider,
		Status:           UserStatusPending, // complete NHP_REG via the resume view before activating
	}
	if err := a.Store.CreateUser(ctx, u); err != nil {
		return nil, err
	}
	return u, nil
}

// generateDefaultNHPKeyMaterial produces a scheme-agnostic private key plus
// the public key derived under the default server's relay-registered scheme.
// Shared by the OIDC upsert paths that have no interactive scheme choice.
// Returns (privB64, pubB64, serverName, scheme, err).
func (a *App) generateDefaultNHPKeyMaterial() (priv, pub, serverName, scheme string, err error) {
	srv := a.Cfg.DefaultServer()
	if srv == nil {
		return "", "", "", "", errors.New("no [[Servers]] configured")
	}
	s := string(srv.RelayRegisteredScheme)
	if s == "" {
		s = string(a.Cfg.CipherScheme)
	}
	priv, err = GenerateSchemeAgnosticPrivateKey()
	if err != nil {
		return "", "", "", "", err
	}
	pub, err = DerivePublicKey(priv, CipherScheme(s))
	if err != nil {
		return "", "", "", "", err
	}
	return priv, pub, srv.Name, s, nil
}

// generateAndStoreNHPKeys was removed when the OIDC merge branch was
// dropped (every IdP-only row the app produced had OIDCSubject == key,
// so the cross-provider guard "u.OIDCSubject != key" was always true for
// the rows this branch targeted — the merge body was unreachable). New
// IdP-only accounts now mint NHP material inline in upsertExternalUser
// via generateDefaultNHPKeyMaterial.

// isUpsertConflict reports whether an upsertExternalUser error should map
// to 409 Conflict rather than 500. Covers the UNIQUE-constraint sentinels
// (ErrUsernameTaken / ErrSubjectTaken / ErrEmailTaken) that fire when a
// namespaced external username, oidc_subject, or email already belongs
// to another account.
//
// ErrEmailTaken is included for the case where the demoapp is still
// running against a pre-review-#2 database — the email column UNIQUE
// is dropped by UserStore.migrate on re-open, but on a DB created
// before this fix landed the column UNIQUE survives and a duplicate
// email INSERT surfaces as ErrEmailTaken. Surfacing it as 500 "user
// upsert failed" is the exact lockout review #2 set out to remove
// (review #6).
func isUpsertConflict(err error) bool {
	return errors.Is(err, ErrUsernameTaken) ||
		errors.Is(err, ErrSubjectTaken) ||
		errors.Is(err, ErrEmailTaken)
}
