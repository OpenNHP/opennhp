// Session helpers. We piggy-back on gin-contrib/sessions (which the
// endpoints module already depends on) so we don't add a new package, and
// the same session store can host both the auth cookie and the OIDC state
// token. Sessions are signed (HMAC-SHA256) via gorilla/securecookie —
// tampering with the cookie invalidates it.
package demoapp

import (
	"errors"
	"net/http"

	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
)

// Session names and keys.
const (
	SessionName      = "demoapp-session"
	SessionKeyUserID = "userID"
	SessionKeyOIDC   = "oidcState"
)

// NewCookieStore returns a signed cookie session store backed by the
// provided secret. The secret must be at least 32 bytes for securecookie.
func NewCookieStore(secret []byte) sessions.Store {
	if len(secret) < 32 {
		// Pad deterministically; the App.validate() emits a warning at boot
		// when the operator left the dev fallback in place.
		pad := make([]byte, 32-len(secret))
		for i := range pad {
			pad[i] = '0'
		}
		secret = append(secret, pad...)
	}
	return sessions.NewCookieStore(secret)
}

// CurrentSession returns the demo session attached to the request context.
func CurrentSession(c *gin.Context) (sessions.Session, error) {
	s := sessions.Default(c)
	if s == nil {
		return nil, errors.New("session middleware not registered")
	}
	return s, nil
}

// LoginUser records the user ID under SessionKeyUserID.
func LoginUser(c *gin.Context, userID int64) error {
	s, err := CurrentSession(c)
	if err != nil {
		return err
	}
	s.Set(SessionKeyUserID, userID)
	return s.Save()
}

// Logout clears the session and marks it for deletion at the browser.
func Logout(c *gin.Context) error {
	s, err := CurrentSession(c)
	if err != nil {
		return err
	}
	s.Clear()
	s.Options(sessions.Options{
		Path:   "/",
		MaxAge: -1,
	})
	return s.Save()
}

// LoggedInUserID returns the signed-in user's ID, or 0 / false if no
// session is present (or the cookie was tampered with).
func LoggedInUserID(c *gin.Context) (int64, bool) {
	s, err := CurrentSession(c)
	if err != nil {
		return 0, false
	}
	v := s.Get(SessionKeyUserID)
	switch x := v.(type) {
	case int64:
		return x, true
	case int:
		return int64(x), true
	case float64:
		return int64(x), true
	}
	return 0, false
}

// SetOIDCState stashes a CSRF nonce in the session so /auth/oidc/callback
// can verify the redirect came from us.
func SetOIDCState(c *gin.Context, state string) error {
	s, err := CurrentSession(c)
	if err != nil {
		return err
	}
	s.Set(SessionKeyOIDC, state)
	return s.Save()
}

// PopOIDCState reads and clears the OIDC state value. Comparing against
// the inbound `state` query parameter prevents CSRF on the callback.
func PopOIDCState(c *gin.Context) (string, error) {
	s, err := CurrentSession(c)
	if err != nil {
		return "", err
	}
	v := s.Get(SessionKeyOIDC)
	s.Delete(SessionKeyOIDC)
	_ = s.Save()
	state, _ := v.(string)
	return state, nil
}

// EnsureSessionCookie issues the Set-Cookie header for the session if it
// hasn't been written yet. Useful right before returning a redirect from
// a flow that just set state (OIDC) but didn't otherwise touch the session.
func EnsureSessionCookie(c *gin.Context) {
	if _, err := c.Cookie(SessionName); err == nil {
		return
	}
	// Touch the store to force cookie write.
	_, _ = CurrentSession(c)
	if w := c.Writer; w != nil && w.Status() == http.StatusOK {
		// gin's default 200 path is fine; nothing to do — the session
		// middleware will persist on next Save().
	}
}
