// Dashboard page render + the JSON APIs it depends on.
package demoapp

import (
	"database/sql"
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
)

// handleDashboard renders the dashboard page. The page is mostly client-
// driven: it fetches /api/user/profile, decrypts the private key with
// the user's password, and uses js-agent to list + knock on resources.
func (a *App) handleDashboard(c *gin.Context) {
	uid, ok := LoggedInUserID(c)
	if !ok || uid == 0 {
		c.Redirect(http.StatusFound, "/login")
		return
	}
	u, err := FindByID(c.Request.Context(), a.DB, uid)
	if err != nil {
		_ = Logout(c)
		c.Redirect(http.StatusFound, "/login")
		return
	}
	if !u.NhpRegisteredAt.Valid {
		// Bounce back into the wizard to finish key onboarding.
		next := "/register?resume=1"
		c.Redirect(http.StatusFound, next)
		return
	}

	a.renderTemplate(c, "dashboard.html", gin.H{
		"Username":      u.Username,
		"Email":         u.Email,
		"OIDCProvider":  nullIfEmpty(u.OIDCProvider.String),
		"AuthServiceId": a.Cfg.NHP.AuthServiceId,
		"CipherScheme":  a.Cfg.NHP.CipherScheme,
	})
}

// handleAPIProfile returns the user's profile data including the
// password-encrypted NHP private key. The browser needs this blob to
// decrypt the key on every page load.
func (a *App) handleAPIProfile(c *gin.Context) {
	uid, _ := LoggedInUserID(c)
	u, err := FindByID(c.Request.Context(), a.DB, uid)
	if err != nil {
		if errors.Is(err, ErrUserNotFound) {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "user not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"userId":              u.Username,
		"email":               u.Email,
		"oidcProvider":        nullIfEmpty(u.OIDCProvider.String),
		"nhpPublicKey":        nullIfEmpty(u.NhpPublicKey.String),
		"encryptedPrivateKey": nullIfEmpty(u.EncryptedPrivateKey.String),
		"nhpRegisteredAt":     nullableUnix(u.NhpRegisteredAt),
	})
}

// handleAPINHPConfig returns the public NHP configuration the browser
// needs to construct js-agent: the relay URL, server public key, ASP id,
// and cipher scheme. This is intentionally unauthenticated for the
// resource list path — the public values are not sensitive.
func (a *App) handleAPINHPConfig(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"relayUrl":      a.Cfg.NHP.RelayUrl,
		"serverPubKey":  a.Cfg.NHP.ServerPublicKey,
		"authServiceId": a.Cfg.NHP.AuthServiceId,
		"cipherScheme":  a.Cfg.NHP.CipherScheme,
	})
}

func nullIfEmpty(s string) any {
	if s == "" {
		return nil
	}
	return s
}

func nullableUnix(v sql.NullInt64) any {
	if !v.Valid {
		return nil
	}
	return v.Int64
}
