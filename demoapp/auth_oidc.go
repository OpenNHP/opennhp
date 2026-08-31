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
	s.Clear() // always clear OIDC session bits, even on error
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
	idTok, err := a.OIDCVefifier.Verifier.Verify(ctx, tok.Extra("id_token").(string))
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "id token verification failed: " + err.Error()})
		return
	}
	if idTok.Nonce != nonce {
		c.JSON(http.StatusBadRequest, gin.H{"error": "OIDC nonce mismatch"})
		return
	}

	var claims struct {
		Sub   string `json:"sub"`
		Email string `json:"email"`
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

	user, err := a.upsertExternalUser(ctx, claims.Sub, email)
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

	// Redirect back to the SPA. The SPA's me/config endpoints will pick
	// up the session cookie automatically.
	c.Redirect(http.StatusTemporaryRedirect, "/")
}

// upsertExternalUser implements the external-IdP (OIDC or OAuth) to
// password user merge logic. It is called from the OIDC callback and
// the GitHub OAuth callback with the IdP's stable subject + email:
//
//  1. If a user already has this external subject, return it.
//  2. If a password user has the same email, link this subject to that
//     user (we treat them as the same person). If the linked user has
//     no NHP keys yet, generate them now so the user can still use the
//     demo.
//  3. Otherwise create a new password-less user with the external
//     subject and freshly generated NHP keys, in the pending state.
//     External-IdP users skip the password step but NOT the NHP_REG
//     handshake — the public key must still be registered with
//     nhp-server. They land on the complete-registration view, which
//     runs the handshake via /api/credentials and flips them active.
func (a *App) upsertExternalUser(ctx context.Context, subject, email string) (*User, error) {
	if u, err := a.Store.GetUserByOIDCSubject(ctx, subject); err == nil {
		return u, nil
	} else if !errors.Is(err, ErrUserNotFound) {
		return nil, err
	}

	if u, err := a.Store.GetUserByEmail(ctx, email); err == nil {
		u.OIDCSubject = subject
		if u.NHPPrivateKeyEnc == "" || u.NHPPublicKey == "" {
			if err := a.generateAndStoreNHPKeys(ctx, u); err != nil {
				return nil, err
			}
		}
		if err := a.Store.UpdateUserOIDCSubject(ctx, u.ID, subject); err != nil {
			return nil, err
		}
		return u, nil
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
		Username:         email, // OIDC users identify by email
		Email:            email,
		OIDCSubject:      subject,
		NHPPublicKey:     pub,
		NHPPrivateKeyEnc: privEnc,
		NHPDeviceID:      deviceID,
		ServerName:       serverName,
		CipherScheme:     scheme,
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
