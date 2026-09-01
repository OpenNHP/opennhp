package main

import (
	"crypto/rand"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
)

// registerRequest is the JSON body for POST /api/register.
//
// Username is the local identifier; email is what we hand to NHP as userId
// and what OIDC subjects are linked by. Password is required only when the
// user is creating a password-based account; OIDC users skip this endpoint.
//
// ServerName + CipherScheme select which nhp-server identity to register
// against and which cipher scheme to derive the public key under. The
// private key itself is scheme-agnostic; only the derived public key +
// the server/scheme binding are fixed at registration. Both are optional
// and default to the first server + its relay-registered scheme.
type registerRequest struct {
	Username     string `json:"username"`
	Password     string `json:"password"`
	Email        string `json:"email"`
	ServerName   string `json:"serverName"`
	CipherScheme string `json:"cipherScheme"`
}

// registerResponse is the JSON body returned from /api/register. The
// privateKey field is the user's NHP private key, base64-encoded, and is
// only ever delivered over TLS to an authenticated session. It is the
// only time the browser sees the private key during registration — once
// the Noise handshake completes, the browser must drop it.
type registerResponse struct {
	RegToken   string            `json:"regToken"`
	PrivateKey string            `json:"privateKey"`
	PublicKey  string            `json:"publicKey"`
	DeviceID   string            `json:"deviceId"`
	NHP        nhpEndpointConfig `json:"nhp"`
}

// nhpEndpointConfig is the relay/server config the SPA needs to build a
// NHPAgent instance — service id, server public key, relay URL, scheme.
//
// ServerPubKey is the server's public key under the chosen CipherScheme
// (used for ECDH). RelayPubKey is the key the relay registered the server
// under; js-agent fingerprints it for the /relay/<id> route. It is empty
// when the chosen scheme matches the relay-registered scheme (the ECDH
// key already produces the right fingerprint), and set otherwise so a
// cross-scheme knock still routes to the correct server.
type nhpEndpointConfig struct {
	ServiceID      string `json:"serviceId"`
	ServerPubKey   string `json:"serverPublicKey"`
	RelayPubKey    string `json:"relayPublicKey,omitempty"`
	RelayURL       string `json:"relayUrl"`
	CipherScheme   string `json:"cipherScheme"`
	UserID         string `json:"userId"`
	OrganizationID string `json:"organizationId"`
	ServerName     string `json:"serverName"`
}

// nhpEndpointFor builds the nhpEndpointConfig the SPA needs for a given
// (server, scheme) binding. It assembles ServerPubKey + RelayPubKey per
// the rule on ServerEntry.relayPublicKeyFor. Returns an error when the
// chosen scheme's public key is not configured for the server (i.e. the
// operator didn't fill both keys, so cross-scheme knock is unavailable).
func (a *App) nhpEndpointFor(srv *ServerEntry, scheme CipherScheme, userID string) (nhpEndpointConfig, error) {
	pub := srv.publicKeyFor(scheme)
	if pub == "" {
		return nhpEndpointConfig{}, fmt.Errorf("server %q has no %s public key configured", srv.Name, scheme)
	}
	return nhpEndpointConfig{
		ServiceID:      srv.ServiceID,
		ServerPubKey:   pub,
		RelayPubKey:    srv.relayPublicKeyFor(scheme),
		RelayURL:       a.Cfg.PublicRelayURL,
		CipherScheme:   string(scheme),
		UserID:         userID,
		OrganizationID: srv.OrganizationID,
		ServerName:     srv.Name,
	}, nil
}

// resolveServerScheme normalizes a (serverName, scheme) pair from a request
// against the configured servers, applying defaults when either is empty.
// Returns the resolved server + scheme, or an error when the named server
// doesn't exist or the scheme is unsupported.
func (a *App) resolveServerScheme(serverName, schemeStr string) (*ServerEntry, CipherScheme, error) {
	srv := a.Cfg.FindServer(serverName)
	if srv == nil {
		srv = a.Cfg.DefaultServer()
		if srv == nil {
			return nil, "", errors.New("no [[Servers]] configured")
		}
	}
	scheme := CipherScheme(schemeStr)
	if scheme == "" {
		scheme = srv.RelayRegisteredScheme
	}
	if scheme == "" {
		scheme = a.Cfg.CipherScheme
	}
	if scheme != CipherSchemeCurve25519 && scheme != CipherSchemeGMSM {
		return nil, "", fmt.Errorf("unsupported cipherScheme %q (expected curve25519 or gmsm)", scheme)
	}
	if srv.publicKeyFor(scheme) == "" {
		return nil, "", fmt.Errorf("server %q does not support scheme %q (missing public key)", srv.Name, scheme)
	}
	return srv, scheme, nil
}

// handleRegister is the entry point for the two-phase registration flow.
//
// Phase 1 (this handler):
//   - validates username/email uniqueness and password strength
//   - generates the NHP key pair in the BACKEND (the browser never sees
//     the keygen step — this is the demo's whole point)
//   - encrypts the private key with AES-256-GCM under KeyEnvelopeKey
//   - persists both keys + a fresh deviceId with status=pending
//   - issues a reg_token the SPA uses to retry / confirm
//   - returns the plain private key in the response so the browser can
//     drive the NHP_REG handshake. The SPA holds it in memory only.
//
// The browser then runs requestOtp -> registerPublicKey -> RAK, and calls
// /api/register/confirm to flip status=active. If the user clears their
// browser before confirm, they hit /api/register/retry with the reg_token
// to get the same private key back without re-typing a password.
func (a *App) handleRegister(c *gin.Context) {
	var req registerRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}
	req.Username = strings.TrimSpace(req.Username)
	req.Email = strings.TrimSpace(strings.ToLower(req.Email))
	if req.Username == "" || req.Email == "" || req.Password == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "username, password, email required"})
		return
	}
	if len(req.Password) < 8 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "password too short (min 8 chars)"})
		return
	}

	ctx := contextOf(c)
	if _, err := a.Store.GetUserByUsername(ctx, req.Username); err == nil {
		c.JSON(http.StatusConflict, gin.H{"error": "username already exists"})
		return
	}
	if _, err := a.Store.GetUserByEmail(ctx, req.Email); err == nil {
		c.JSON(http.StatusConflict, gin.H{"error": "email already exists"})
		return
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "password hashing failed"})
		return
	}

	// Resolve the target server + scheme BEFORE keygen so the public key is
	// derived under the scheme the user actually chose. The private key is
	// scheme-agnostic; only the derived public key + binding are persisted.
	srv, scheme, err := a.resolveServerScheme(req.ServerName, req.CipherScheme)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	priv, err := GenerateSchemeAgnosticPrivateKey()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "key generation failed"})
		return
	}
	pub, err := DerivePublicKey(priv, scheme)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "public key derivation failed"})
		return
	}
	deviceID, err := randomDeviceID()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "device id generation failed"})
		return
	}
	privEnc, err := sealKey(a.MasterKey, priv)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "key sealing failed"})
		return
	}

	u := &User{
		Username:         req.Username,
		PasswordHash:     string(hash),
		Email:            req.Email,
		NHPPublicKey:     pub,
		NHPPrivateKeyEnc: privEnc,
		NHPDeviceID:      deviceID,
		ServerName:       srv.Name,
		CipherScheme:     string(scheme),
		AuthProvider:     "password",
		Status:           UserStatusPending,
	}
	if err := a.Store.CreateUser(ctx, u); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "user create failed"})
		return
	}

	regToken, err := randomToken()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "token generation failed"})
		return
	}
	if err := a.Store.SaveRegToken(ctx, &RegToken{
		Token:     regToken,
		UserID:    u.ID,
		ExpiresAt: time.Now().UTC().Add(30 * time.Minute),
	}); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "token save failed"})
		return
	}

	// Persist the reg_token in the session too, so /api/register/confirm
	// can be called without the SPA having to round-trip the token.
	s := sessions.Default(c)
	s.Set(sessKeyRegToken, regToken)
	s.Set(sessKeyUserID, u.ID)
	s.Set(sessKeyUsername, u.Username)
	_ = s.Save()

	nhp, err := a.nhpEndpointFor(srv, scheme, u.Email)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, registerResponse{
		RegToken:   regToken,
		PrivateKey: priv,
		PublicKey:  pub,
		DeviceID:   deviceID,
		NHP:        nhp,
	})
}

// registerRetryRequest is the JSON body for POST /api/register/retry.
type registerRetryRequest struct {
	RegToken string `json:"regToken"`
}

// handleRegisterRetry returns the same registration material as the
// original /api/register call (private key, deviceId, NHP config) so a
// user who lost their browser session can re-drive the NHP_REG flow
// without re-typing their password. Gated by the reg_token to avoid
// trivially exposing private keys.
func (a *App) handleRegisterRetry(c *gin.Context) {
	var req registerRetryRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}
	ctx := contextOf(c)
	_, user, err := a.Store.LookupRegToken(ctx, req.RegToken)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if user.NHPPrivateKeyEnc == "" || user.NHPPublicKey == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "user has no NHP keys; restart registration"})
		return
	}
	priv, err := openKey(a.MasterKey, user.NHPPrivateKeyEnc)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "key unseal failed"})
		return
	}
	srv, scheme, err := a.userServerScheme(user)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	nhp, err := a.nhpEndpointFor(srv, scheme, user.Email)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, registerResponse{
		RegToken:   req.RegToken,
		PrivateKey: priv,
		PublicKey:  user.NHPPublicKey,
		DeviceID:   user.NHPDeviceID,
		NHP:        nhp,
	})
}

// userServerScheme resolves a user's stored (server_name, cipher_scheme)
// binding to a configured ServerEntry + CipherScheme. Legacy users with
// empty bindings fall back to the default server + its registered scheme.
// Returns an error only when no servers are configured at all.
func (a *App) userServerScheme(u *User) (*ServerEntry, CipherScheme, error) {
	srv := a.Cfg.FindServer(u.ServerName)
	if srv == nil {
		srv = a.Cfg.DefaultServer()
	}
	if srv == nil {
		return nil, "", errors.New("no [[Servers]] configured")
	}
	scheme := CipherScheme(u.CipherScheme)
	if scheme == "" {
		scheme = srv.RelayRegisteredScheme
	}
	if scheme == "" {
		scheme = a.Cfg.CipherScheme
	}
	if scheme == "" {
		scheme = CipherSchemeCurve25519
	}
	return srv, scheme, nil
}

// bindRequest is the JSON body for POST /api/register/bind. It lets a
// signed-in pending user (re)pick their nhp-server cluster + cipher
// scheme before completing NHP_REG. Used by the complete-registration
// view, where external-IdP (GitHub/OIDC) users land with a default
// binding they never chose and password users may want to change.
type bindRequest struct {
	ServerName   string `json:"serverName"`
	CipherScheme string `json:"cipherScheme"`
}

// handleRebind re-derives the user's NHP public key under the chosen
// server + scheme and persists the new binding, returning the reg
// material the SPA needs to drive NHP_REG. The private key is
// scheme-agnostic, so it is NOT rotated — only the derived public key
// + the server/scheme binding change. If the user somehow has no key
// yet (edge case), one is generated here. The deviceId is kept so a
// resumed registration keeps the same device identity.
//
// Identity is proven by the requireSession middleware (the caller is
// already logged in); this is the session counterpart to the
// regToken-gated /api/register/retry path.
func (a *App) handleRebind(c *gin.Context) {
	user, ok := c.MustGet("user").(*User)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "user context missing"})
		return
	}
	// An active user re-deriving their public key under a new scheme would
	// persist a key the nhp-server never saw in NHP_REG, so every
	// subsequent knock/listServices fails against the stale server-side
	// key while /api/me still reports active — bricking the account with
	// no UI path to re-register. Only pending users (mid-registration)
	// may rebind; they re-run NHP_REG before /api/register/confirm flips
	// them active.
	if user.Status != UserStatusPending {
		c.JSON(http.StatusConflict, gin.H{"error": "account already active; rebinding requires re-registration"})
		return
	}
	var req bindRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}
	srv, scheme, err := a.resolveServerScheme(req.ServerName, req.CipherScheme)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	ctx := contextOf(c)

	// Unseal the existing private key, or generate one if none exists.
	var priv string
	if user.NHPPrivateKeyEnc != "" {
		priv, err = openKey(a.MasterKey, user.NHPPrivateKeyEnc)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "key unseal failed"})
			return
		}
	} else {
		priv, err = GenerateSchemeAgnosticPrivateKey()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "key generation failed"})
			return
		}
	}
	pub, err := DerivePublicKey(priv, scheme)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "public key derivation failed"})
		return
	}

	// Keep the existing deviceId so a resumed registration does not
	// change device identity mid-flow; mint one only if absent.
	deviceID := user.NHPDeviceID
	if deviceID == "" {
		deviceID, err = randomDeviceID()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "device id generation failed"})
			return
		}
	}
	privEnc, err := sealKey(a.MasterKey, priv)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "key sealing failed"})
		return
	}
	if err := a.Store.UpdateUserKeys(ctx, user.ID, pub, privEnc, deviceID, srv.Name, string(scheme)); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "update binding failed"})
		return
	}

	nhp, err := a.nhpEndpointFor(srv, scheme, user.Email)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	// No regToken: the caller is already in a session and completes via
	// the session path in /api/register/confirm.
	c.JSON(http.StatusOK, registerResponse{
		RegToken:   "",
		PrivateKey: priv,
		PublicKey:  pub,
		DeviceID:   deviceID,
		NHP:        nhp,
	})
}

// registerConfirmRequest is the JSON body for POST /api/register/confirm.
// RegToken is required for the fresh-registration flow but may be empty
// when the caller is already signed in (the resume path proves identity
// via the session instead).
type registerConfirmRequest struct {
	RegToken  string `json:"regToken"`
	DeviceID  string `json:"deviceId"`
	ExpiresAt int64  `json:"expiresAt"` // unix-millis from the server RAK, if any
	RakOK     bool   `json:"rakOk"`
}

// handleRegisterConfirm finalizes registration. We trust RakOK here
// because the SPA already received the RAK from nhp-server in the
// browser session. Identity is proven one of two ways:
//   - regToken: the token issued by /api/register (ties the call to the
//     original account creation; used by the fresh-registration flow).
//   - session: when regToken is empty, the caller is already logged in
//     (sessKeyUserID). This is the resume path — a pending user who
//     logged in and is completing the NHP_REG handshake after the fact.
//
// Both paths require the request's deviceId to match the user's stored
// deviceId before activating. Once confirmed the user is "active".
func (a *App) handleRegisterConfirm(c *gin.Context) {
	var req registerConfirmRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}
	if !req.RakOK {
		c.JSON(http.StatusBadRequest, gin.H{"error": "RakOK must be true to confirm"})
		return
	}
	ctx := contextOf(c)

	var (
		user   *User
		regTok string
	)
	if strings.TrimSpace(req.RegToken) != "" {
		// Fresh-registration path: prove identity via the reg token.
		rt, u, err := a.Store.LookupRegToken(ctx, req.RegToken)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		regTok = rt.Token
		user = u
	} else {
		// Resume path: prove identity via the logged-in session.
		s := sessions.Default(c)
		uidRaw := s.Get(sessKeyUserID)
		if uidRaw == nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "regToken required when not signed in"})
			return
		}
		uid, ok := uidRaw.(int64)
		if !ok {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid session"})
			return
		}
		u, err := a.Store.GetUserByID(ctx, uid)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "user not found"})
			return
		}
		user = u
	}

	if user.NHPDeviceID != req.DeviceID {
		c.JSON(http.StatusBadRequest, gin.H{"error": "deviceId mismatch"})
		return
	}
	if err := a.Store.ActivateUser(ctx, user.ID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "activate failed"})
		return
	}
	// Best-effort cleanup; ignore failure so a stale token doesn't
	// prevent confirm.
	if regTok != "" {
		_ = a.Store.DeleteRegToken(ctx, regTok)
	}

	// Set a fresh session so the SPA can immediately fetch /api/credentials.
	s := sessions.Default(c)
	s.Set(sessKeyUserID, user.ID)
	s.Set(sessKeyUsername, user.Username)
	_ = s.Save()

	c.JSON(http.StatusOK, gin.H{"success": true})
}

// loginRequest is the JSON body for POST /api/login.
type loginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

// handleLogin verifies the bcrypt hash and creates a session.
func (a *App) handleLogin(c *gin.Context) {
	var req loginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}
	req.Username = strings.TrimSpace(req.Username)
	if req.Username == "" || req.Password == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "username and password required"})
		return
	}
	ctx := contextOf(c)
	user, err := a.Store.GetUserByUsername(ctx, req.Username)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid credentials"})
		return
	}
	// Pending users may log in — their account exists and NHP keys are
	// sealed, but the NHP_REG handshake with nhp-server never completed.
	// The SPA inspects `status` and routes them to the complete-registration
	// view instead of resources.
	if user.PasswordHash == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "this account uses OIDC; please log in with the IdP"})
		return
	}
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password)); err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid credentials"})
		return
	}

	s := sessions.Default(c)
	s.Set(sessKeyUserID, user.ID)
	s.Set(sessKeyUsername, user.Username)
	if err := s.Save(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "session save failed"})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"username": user.Username,
		"email":    user.Email,
		"status":   string(user.Status),
	})
}

// handleLogout clears the session cookie.
func (a *App) handleLogout(c *gin.Context) {
	s := sessions.Default(c)
	s.Clear()
	_ = s.Save()
	c.JSON(http.StatusOK, gin.H{"success": true})
}

// handleDeleteAccount deregisters the signed-in user: deletes their local
// row (username, password hash, sealed NHP private key, deviceId, etc.) and
// clears the session. reg_tokens cascade via the FK on the reg_tokens table.
//
// The NHP-server-side agent public key is NOT removed here — it is left to
// expire via the server's AgentKeyTTLSeconds (24h default). The deleted
// private key is gone, so the orphaned server key is unusable; and
// FindAgentByPublicKey already filters on expires_at. Re-registration mints a
// fresh keypair + deviceId, so it never collides with the orphaned row.
//
// Identity is proven by the requireSession middleware, which loaded the user
// into the context. A 404 here means a concurrent delete already removed the
// row — clear the session and report success so the SPA routes to login.
func (a *App) handleDeleteAccount(c *gin.Context) {
	user, ok := c.MustGet("user").(*User)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "user context missing"})
		return
	}
	ctx := contextOf(c)
	if err := a.Store.DeleteUser(ctx, user.ID); err != nil {
		if errors.Is(err, ErrUserNotFound) {
			// Row already gone (concurrent delete). Still clear the session
			// so the client ends logged out and can re-register.
			s := sessions.Default(c)
			s.Clear()
			_ = s.Save()
			c.JSON(http.StatusOK, gin.H{"success": true})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "delete failed"})
		return
	}
	s := sessions.Default(c)
	s.Clear()
	_ = s.Save()
	c.JSON(http.StatusOK, gin.H{"success": true})
}

// handleMe returns the current user's basic profile (no secrets).
func (a *App) handleMe(c *gin.Context) {
	user, ok := c.MustGet("user").(*User)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "user context missing"})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"username":     user.Username,
		"email":        user.Email,
		"status":       string(user.Status),
		"cipherScheme": user.CipherScheme,
		"serverName":   user.ServerName,
		"authProvider": user.AuthProvider,
	})
}

// handleGetCredentials returns the wrapped private key material to an
// authenticated session. The SPA uses this for /api/list and /api/knock.
// This is the only endpoint that returns the unwrapped private key
// during login.
func (a *App) handleGetCredentials(c *gin.Context) {
	user := c.MustGet("user").(*User)
	if user.NHPPrivateKeyEnc == "" || user.NHPPublicKey == "" || user.NHPDeviceID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "user has no NHP credentials; please re-register"})
		return
	}
	priv, err := openKey(a.MasterKey, user.NHPPrivateKeyEnc)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "credential unseal failed"})
		return
	}
	srv, scheme, err := a.userServerScheme(user)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	nhp, err := a.nhpEndpointFor(srv, scheme, user.Email)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"privateKey": priv,
		"publicKey":  user.NHPPublicKey,
		"deviceId":   user.NHPDeviceID,
		"userId":     user.Email,
		"nhp":        nhp,
	})
}

// handleGetConfig returns the catalog of resources the SPA can show
// alongside the dynamic list from listServices. The SPA intersects
// the list with this catalog so it can render titles/URLs/AC hosts.
// Resources are filtered to the user's registered server (global ones
// with no ServerName are shown to all).
func (a *App) handleGetConfig(c *gin.Context) {
	user := c.MustGet("user").(*User)
	srv, scheme, err := a.userServerScheme(user)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	nhp, err := a.nhpEndpointFor(srv, scheme, user.Email)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	resources := a.Cfg.ResourcesFor(srv.Name)
	out := gin.H{
		"serviceId":      nhp.ServiceID,
		"serverPubKey":   nhp.ServerPubKey,
		"relayPubKey":    nhp.RelayPubKey,
		"relayUrl":       nhp.RelayURL,
		"cipherScheme":   nhp.CipherScheme,
		"organizationId": nhp.OrganizationID,
		"serverName":     nhp.ServerName,
		"userId":         user.Email,
		"resources":      make([]gin.H, 0, len(resources)),
	}
	for _, r := range resources {
		out["resources"] = append(out["resources"].([]gin.H), gin.H{
			"id":     r.ID,
			"title":  r.Title,
			"url":    r.URL,
			"acHost": r.ACHost,
		})
	}
	c.JSON(http.StatusOK, out)
}

// randomDeviceID returns a 32-character hex string. We use this for the
// NHP deviceId; it's stored alongside the public key so the SPA can pick
// it up on retry.
func randomDeviceID() (string, error) {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}

// randomToken returns a 32-byte URL-safe base64 token, used for reg_tokens
// and OIDC state.
func randomToken() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", fmt.Errorf("rand: %w", err)
	}
	return base64.RawURLEncoding.EncodeToString(b), nil
}

// errInvalidSession is used by handlers that need to return 401 cleanly.
var errInvalidSession = errors.New("invalid session")
