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
	s.Set(sessKeyUsername, user.Username)
	s.Set(sessKeyOIDCSubject, user.OIDCSubject)
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
//  2. If an IdP-only user has the same email, link this subject to that
//     user. We do NOT auto-link a password account: /api/register
//     performs no email verification, so a password account's email is
//     unproven. Pre-review-#2 the function refused the IdP sign-in
//     entirely in that case (errEmailHeldByPasswordAccount), which let
//     an attacker who pre-registered victim@example.com permanently
//     lock the victim out of GitHub sign-in (no link UI existed).
//     Instead, we now SKIP the password-holder row and create a
//     SEPARATE IdP-only account for the genuine IdP user. The email
//     UNIQUE constraint has been relaxed in userstore.dropEmailUnique so
//     the two rows can coexist. The squatter's row is untouched.
//  3. Otherwise (no email match, or the only match was a password
//     holder that we skipped in step 2) create a new password-less
//     user with the external subject, the given provider, and freshly
//     generated NHP keys, in the pending state. External-IdP users
//     skip the password step but NOT the NHP_REG handshake — the
//     public key must still be registered with nhp-server. They land
//     on the complete-registration view, which runs the handshake via
//     /api/credentials and flips them active.
//
// Sentinel errors (callers map these to 409, not 500):
//   - errEmailLinkedToDifferentSubject: email matches an IdP-only
//     account already bound to a different subject — refuse to overwrite.
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

	// Step 2: if an IdP-only user already has this email, merge onto
	// that row. Password-holding rows are skipped (not merged, not
	// rejected) — the new-user branch below creates a separate
	// IdP-only account so the genuine IdP user is not locked out by a
	// password squat. Review #2.
	if u, err := a.Store.GetUserByEmail(ctx, email); err == nil {
		if u.PasswordHash == "" {
			// IdP-only row — safe to merge.
			// Don't silently overwrite an existing different subject.
			if u.OIDCSubject != "" && u.OIDCSubject != key {
				return nil, errEmailLinkedToDifferentSubject
			}
			u.OIDCSubject = key
			if u.NHPPrivateKeyEnc == "" || u.NHPPublicKey == "" {
				if err := a.generateAndStoreNHPKeys(ctx, u); err != nil {
					return nil, err
				}
			}
			if err := a.Store.UpdateUserOIDCSubject(ctx, u.ID, provider, subject); err != nil {
				return nil, err
			}
			return u, nil
		}
		// Password-holding row: fall through to create a separate
		// IdP-only account. The squatter's row is left intact; we do
		// not absorb its identity (which would let the password
		// holder harvest a key the genuine user later registers) and
		// we do not reject the sign-in (which would lock the genuine
		// user out indefinitely).
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

// generateAndStoreNHPKeys fills in the NHP material on an existing user
// (currently only used for OIDC-linked password users). Caller must
// have already verified the user exists.
func (a *App) generateAndStoreNHPKeys(ctx context.Context, u *User) error {
	priv, pub, serverName, scheme, err := a.generateDefaultNHPKeyMaterial()
	if err != nil {
		return err
	}
	deviceID, err := randomDeviceID()
	if err != nil {
		return err
	}
	privEnc, err := sealKey(a.MasterKey, priv)
	if err != nil {
		return err
	}
	return a.Store.UpdateUserKeys(ctx, u.ID, pub, privEnc, deviceID, serverName, scheme)
}

// Sentinel errors for upsertExternalUser. Callers (the OIDC and GitHub
// callbacks) surface these as 409 Conflict rather than 500, so a user
// who hits the relink guard gets a meaningful message instead of an
// opaque server error.
//
// errEmailHeldByPasswordAccount was removed in review #2: the pre-hijack
// guard no longer rejects IdP sign-ins on a password-holder email match
// — it now SKIPS the squat row and creates a separate IdP-only account.
// The IdP-only branch still raises errEmailLinkedToDifferentSubject when
// the email match is itself an IdP-only account bound to a different
// subject.
var (
	errEmailLinkedToDifferentSubject = errors.New("email already linked to a different IdP identity")
)

// isUpsertConflict reports whether an upsertExternalUser error should map
// to 409 Conflict rather than 500. Covers the relink guard
// (errEmailLinkedToDifferentSubject) and the UNIQUE-constraint sentinels
// (ErrUsernameTaken / ErrSubjectTaken) that fire when a namespaced
// external username or oidc_subject already belongs to another account
// (review #3/#4).
func isUpsertConflict(err error) bool {
	return errors.Is(err, errEmailLinkedToDifferentSubject) ||
		errors.Is(err, ErrUsernameTaken) ||
		errors.Is(err, ErrSubjectTaken)
}
