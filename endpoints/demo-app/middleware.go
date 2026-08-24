// HTTP middleware for the Demo App.
package demoapp

import (
	"crypto/subtle"
	"database/sql"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

const (
	// HeaderInternalAPIKey is the header used by demo-app-plugin to
	// authenticate against the internal callback endpoints.
	HeaderInternalAPIKey = "X-Plugin-Api-Key"

	// CtxKeyUserID is the gin context key under which the authenticated
	// user ID is stored by AuthRequired. Handlers can fetch it via
	// c.MustGet(CtxKeyUserID).
	CtxKeyUserID = "demoapp.userID"
)

// AuthRequired redirects unauthenticated browser requests to /login and
// rejects API requests with 401. It also looks up the user record and
// exposes the numeric ID at CtxKeyUserID for downstream handlers.
func AuthRequired(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		uid, ok := LoggedInUserID(c)
		if !ok || uid <= 0 {
			if wantsJSON(c) {
				c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "not authenticated"})
				return
			}
			c.Redirect(http.StatusFound, "/login?next="+c.Request.URL.Path)
			c.Abort()
			return
		}

		// Touch the DB to confirm the user still exists (handles the case
		// where an admin deletes a user without invalidating sessions).
		if _, err := FindByID(c.Request.Context(), db, uid); err != nil {
			_ = Logout(c)
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "session user no longer exists"})
			return
		}
		c.Set(CtxKeyUserID, uid)
		c.Next()
	}
}

// InternalAPIKey returns middleware that rejects requests missing or
// carrying a wrong X-Plugin-Api-Key header. The comparison is constant-time.
func InternalAPIKey(expected string) gin.HandlerFunc {
	expectedBytes := []byte(expected)
	return func(c *gin.Context) {
		got := c.GetHeader(HeaderInternalAPIKey)
		if expected == "" || subtle.ConstantTimeCompare([]byte(got), expectedBytes) != 1 {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid internal api key"})
			return
		}
		c.Next()
	}
}

// wantsJSON returns true when the request prefers a JSON error over a
// HTML redirect. /api/* always counts as JSON; otherwise we honor the
// Accept header.
func wantsJSON(c *gin.Context) bool {
	if strings.HasPrefix(c.Request.URL.Path, "/api/") ||
		strings.HasPrefix(c.Request.URL.Path, "/internal/") {
		return true
	}
	return strings.Contains(c.GetHeader("Accept"), "application/json")
}
