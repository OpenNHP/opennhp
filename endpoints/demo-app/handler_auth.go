// Authentication handlers: register / login / logout / OIDC / page renders.
package demoapp

import (
	"context"
	"crypto/rand"
	"database/sql"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"

	"github.com/coreos/go-oidc/v3/oidc"
	"golang.org/x/oauth2"
)

// RegisterPageRequest and RegisterRequest.
//
// RegisterRequest is the JSON body the browser POSTs after generating
// the NHP keypair. We don't trust the password for any crypto material
// — the encryptedPrivateKey already contains the PBKDF2/AES-GCM blob
// — but we DO need it for bcrypt (login).
type RegisterRequest struct {
	Username            string `json:"username" binding:"required,min=3,max=64"`
	Password            string `json:"password" binding:"required,min=8"`
	Email               string `json:"email" binding:"required,email"`
	EncryptedPrivateKey string `json:"encryptedPrivateKey" binding:"required"`
	NhpPublicKey        string `json:"nhpPublicKey" binding:"required"`
}

// LoginRequest is the JSON body the browser POSTs from /login.
type LoginRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

// handleIndex renders the landing page. The login state decides whether
// the user sees "Sign in" or "Go to dashboard".
func (a *App) handleIndex(c *gin.Context) {
	uid, ok := LoggedInUserID(c)
	if ok && uid > 0 {
		c.Redirect(http.StatusFound, "/dashboard")
		return
	}
	a.renderTemplate(c, "index.html", gin.H{
		"OIDCEnabled": a.Cfg.OIDC.Enabled,
	})
}

// handleLoginPage renders the login form.
func (a *App) handleLoginPage(c *gin.Context) {
	a.renderTemplate(c, "login.html", gin.H{
		"OIDCEnabled": a.Cfg.OIDC.Enabled,
		"Next":        c.Query("next"),
	})
}

// handleRegisterPage renders the (multi-step) registration form. The
// browser-side wizard lives in web/static/register.js.
func (a *App) handleRegisterPage(c *gin.Context) {
	a.renderTemplate(c, "register.html", gin.H{
		"AuthServiceId": a.Cfg.NHP.AuthServiceId,
		"CipherScheme":  a.Cfg.NHP.CipherScheme,
	})
}

// handleAPIRegister stores a new Demo App user. It does NOT log the user
// in — the browser still needs to complete the NHP key registration via
// js-agent, which is signalled by setting nhp_registered_at.
//
// The user's password is bcrypt-hashed (cost=12) for later login. The
// encryptedPrivateKey is stored verbatim; it's a self-describing blob
// that contains its own PBKDF2 salt + AES-GCM IV (see web/static/crypto.js).
func (a *App) handleAPIRegister(c *gin.Context) {
	var req RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), 12)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "password hash failed"})
		return
	}

	tx, err := a.DB.BeginTx(c.Request.Context(), nil)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "db error"})
		return
	}
	defer tx.Rollback() //nolint:errcheck

	u := &User{
		Username:            req.Username,
		Email:               req.Email,
		NhpPublicKey:        sql.NullString{String: req.NhpPublicKey, Valid: true},
		EncryptedPrivateKey: sql.NullString{String: req.EncryptedPrivateKey, Valid: true},
	}
	u.PasswordHash = sql.NullString{String: string(hash), Valid: true}

	id, err := CreateUser(c.Request.Context(), tx, u)
	if err != nil {
		if errors.Is(err, ErrUserAlreadyExists) {
			c.JSON(http.StatusConflict, gin.H{"error": "username already taken"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if err := tx.Commit(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"ok": true, "userId": id})
}

// handleAPILogin authenticates the user via username + bcrypt password.
// It rejects users whose NHP key isn't yet registered — they must
// finish NHP_REG first.
func (a *App) handleAPILogin(c *gin.Context) {
	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	u, err := FindByUsername(c.Request.Context(), a.DB, req.Username)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid username or password"})
		return
	}

	if !u.PasswordHash.Valid {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "this account uses OIDC; sign in via SSO"})
		return
	}
	if err := bcrypt.CompareHashAndPassword([]byte(u.PasswordHash.String), []byte(req.Password)); err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid username or password"})
		return
	}

	// For OIDC accounts (no password_hash) we'd never get here; for
	// password accounts, nhp_registered_at is the gate that says "the
	// browser has finished the NHP key registration". Without it the
	// dashboard will fail to load resources, so we block at login.
	if !u.NhpRegisteredAt.Valid {
		c.JSON(http.StatusForbidden, gin.H{
			"error":      "NHP key not registered yet",
			"next":       "/register?resume=1",
			"nhpPending": true,
		})
		return
	}

	if err := LoginUser(c, u.ID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "session save failed"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"ok": true, "userId": u.ID, "username": u.Username})
}

// handleAPILogout clears the session. Always returns 200 even if the
// cookie was missing — idempotent.
func (a *App) handleAPILogout(c *gin.Context) {
	_ = Logout(c)
	c.JSON(http.StatusOK, gin.H{"ok": true})
}

// ── OIDC ───────────────────────────────────────────────────────────────

// oidcProvider bundles an OIDC verifier and the matching OAuth2 config.
// Either both are nil (OIDC disabled) or both are set.
type oidcProvider struct {
	verifier *oidc.IDTokenVerifier
	config   *oauth2.Config
}

// oidcProviderCache is a process-wide singleton so the discovery doc is
// fetched exactly once. It's safe to read without a lock because the
// field is set once at startup (from /auth/oidc/start) and never
// mutated afterwards.
var oidcProviderCache *oidcProvider

// oidcLazy loads (and caches) the OIDC provider on first use.
func (a *App) oidcLazy() (*oidcProvider, error) {
	if !a.Cfg.OIDC.Enabled {
		return nil, errors.New("OIDC not enabled")
	}
	if oidcProviderCache != nil {
		return oidcProviderCache, nil
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	provider, err := oidc.NewProvider(ctx, a.Cfg.OIDC.IssuerURL)
	if err != nil {
		return nil, fmt.Errorf("oidc discovery: %w", err)
	}
	cfg := &oauth2.Config{
		ClientID:     a.Cfg.OIDC.ClientID,
		ClientSecret: a.Cfg.OIDC.ClientSecret,
		RedirectURL:  a.Cfg.OIDC.RedirectURL,
		Endpoint:     provider.Endpoint(),
		Scopes:       strings.Split(a.Cfg.OIDC.Scopes, " "),
	}
	verifier := provider.Verifier(&oidc.Config{ClientID: a.Cfg.OIDC.ClientID})
	oidcProviderCache = &oidcProvider{verifier: verifier, config: cfg}
	return oidcProviderCache, nil
}

// handleOIDCStart redirects the browser to the IdP's authorization
// endpoint. The state nonce is stored in the session so the callback
// can verify it.
func (a *App) handleOIDCStart(c *gin.Context) {
	p, err := a.oidcLazy()
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	state := randomURLSafe(24)
	if err := SetOIDCState(c, state); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "session save failed"})
		return
	}
	c.Redirect(http.StatusFound, p.config.AuthCodeURL(state, oidc.Nonce(randomURLSafe(16))))
}

// handleOIDCCallback completes the OIDC code exchange. Two paths are
// supported:
//   1. First-time OIDC sign-in: create a Demo user keyed on (provider, sub).
//      If nhp_registered_at is still NULL, the user is sent to /register
//      to complete NHP key onboarding.
//   2. Returning user: log them in and redirect to /dashboard.
func (a *App) handleOIDCCallback(c *gin.Context) {
	state, err := PopOIDCState(c)
	if err != nil || state == "" || state != c.Query("state") {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid oauth state"})
		return
	}

	p, err := a.oidcLazy()
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	code := c.Query("code")
	if code == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "missing code"})
		return
	}
	token, err := p.config.Exchange(c.Request.Context(), code)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "code exchange failed"})
		return
	}
	rawIDToken, ok := token.Extra("id_token").(string)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "no id_token in response"})
		return
	}
	idToken, err := p.verifier.Verify(c.Request.Context(), rawIDToken)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "id_token verify failed"})
		return
	}

	var claims struct {
		Sub     string `json:"sub"`
		Email   string `json:"email"`
		Name    string `json:"name"`
		Picture string `json:"picture"`
	}
	if err := idToken.Claims(&claims); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "claims parse failed"})
		return
	}
	if claims.Sub == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "id_token missing sub"})
		return
	}

	provider := idToken.Issuer
	if provider == "" {
		provider = a.Cfg.OIDC.IssuerURL
	}

	// First-time sign-in: create the user if necessary.
	u, err := FindByOIDCSub(c.Request.Context(), a.DB, provider, claims.Sub)
	if errors.Is(err, ErrUserNotFound) {
		if claims.Email == "" {
			// Without an email we can't deliver OTPs. Force the user back
			// through /register with a real password account instead.
			c.Redirect(http.StatusFound, "/register?oidcNeedsEmail=1")
			return
		}
		// Synthesize a username from the email local-part. If the chosen
		// username is taken we append a short hex suffix.
		username := usernameFromEmail(claims.Email)
		username, err = a.uniqueUsername(c.Request.Context(), username)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		u = &User{
			Username:     username,
			Email:        claims.Email,
			OIDCSub:      sql.NullString{String: claims.Sub, Valid: true},
			OIDCProvider: sql.NullString{String: provider, Valid: true},
		}
		id, err := CreateUser(c.Request.Context(), a.DB, u)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		u.ID = id
	} else if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// The OIDC user must still complete NHP key onboarding before they
	// can knock. We log them in immediately so the browser can POST to
	// /api/oidc/nhp-onboard with the encrypted private key, but the
	// dashboard will gate them until nhp_registered_at is set.
	if err := LoginUser(c, u.ID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "session save failed"})
		return
	}

	if !u.NhpRegisteredAt.Valid {
		c.Redirect(http.StatusFound, "/register?oidc=1")
		return
	}
	c.Redirect(http.StatusFound, "/dashboard")
}

// handleOIDCOnboard completes the NHP key onboarding for an OIDC user.
// It's analogous to step 3 of the password registration wizard: the
// browser POSTs the keypair + the encrypted private key (keyed on a
// password the user just chose) and we update the row.
func (a *App) handleOIDCOnboard(c *gin.Context) {
	var req struct {
		EncryptedPrivateKey string `json:"encryptedPrivateKey" binding:"required"`
		NhpPublicKey        string `json:"nhpPublicKey" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	uid, _ := LoggedInUserID(c)
	if uid == 0 {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "not authenticated"})
		return
	}

	username := currentUsername(c, a.DB, uid)
	if username == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "user no longer exists"})
		return
	}
	if err := SetOIDCKeys(c.Request.Context(), a.DB, username, req.NhpPublicKey, req.EncryptedPrivateKey); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// The user still needs to run NHP_OTP + NHP_REG from the browser
	// before we mark them complete. The /register wizard in oidc mode
	// handles that; we just return success here.
	c.JSON(http.StatusOK, gin.H{"ok": true})
}

// ── helpers ────────────────────────────────────────────────────────────

// usernameFromEmail strips the @-domain and replaces dots/plus signs
// that are awkward in usernames. We still call uniqueUsername to make
// sure the result is free.
func usernameFromEmail(email string) string {
	at := strings.IndexByte(email, '@')
	local := email
	if at > 0 {
		local = email[:at]
	}
	local = strings.ReplaceAll(local, ".", "_")
	local = strings.ReplaceAll(local, "+", "_")
	return local
}

// uniqueUsername appends a short random suffix until the chosen name is
// not taken. We cap at a handful of attempts to avoid spinning.
func (a *App) uniqueUsername(ctx context.Context, base string) (string, error) {
	if base == "" {
		base = "user"
	}
	for i := 0; i < 8; i++ {
		candidate := base
		if i > 0 {
			candidate = fmt.Sprintf("%s_%s", base, randomHex(3))
		}
		_, err := FindByUsername(ctx, a.DB, candidate)
		if errors.Is(err, ErrUserNotFound) {
			return candidate, nil
		}
		if err != nil {
			return "", err
		}
	}
	return "", errors.New("could not generate a unique username")
}

// currentUsername looks up the username behind the session user ID. We
// keep this off the hot path because it's only used by handleOIDCOnboard.
func currentUsername(c *gin.Context, db *sql.DB, uid int64) string {
	u, err := FindByID(c.Request.Context(), db, uid)
	if err != nil || u == nil {
		return ""
	}
	return u.Username
}

// randomURLSafe returns a base64url-encoded random string of n bytes
// (the output is ~1.33 * n characters long).
func randomURLSafe(n int) string {
	buf := make([]byte, n)
	if _, err := rand.Read(buf); err != nil {
		// crypto/rand should never fail; fall back to uuid for paranoia.
		return base64.RawURLEncoding.EncodeToString([]byte(uuid.New().String()))
	}
	return base64.RawURLEncoding.EncodeToString(buf)
}

// randomHex returns hex(n) of random bytes.
func randomHex(n int) string {
	buf := make([]byte, n)
	if _, err := rand.Read(buf); err != nil {
		return uuid.New().String()[:n]
	}
	return hex.EncodeToString(buf)
}

// safeRedirect ensures the `next` query parameter points to a local path.
// We accept anything starting with "/" that does not begin with "//" or
// "/\\", which are common bypass vectors.
func safeRedirect(next string) string {
	if next == "" || !strings.HasPrefix(next, "/") || strings.HasPrefix(next, "//") || strings.HasPrefix(next, "/\\") {
		return "/dashboard"
	}
	if u, err := url.Parse(next); err == nil && u.Host != "" {
		return "/dashboard"
	}
	return next
}
