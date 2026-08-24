// Internal callbacks invoked by demo-app-plugin (running inside
// nhp-server). They are gated by the X-Plugin-Api-Key middleware.
package demoapp

import (
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

// OTPDeliverRequest is the JSON body posted by demo-app-plugin after
// RequestOTP generated a one-time password.
type OTPDeliverRequest struct {
	UserID  string `json:"userId" binding:"required"`
	OtpCode string `json:"otpCode" binding:"required"`
	Email   string `json:"email" binding:"required,email"`
}

// MarkRegisteredRequest is the JSON body posted by demo-app-plugin after
// RegisterAgent successfully stored the user's public key.
type MarkRegisteredRequest struct {
	UserID    string `json:"userId" binding:"required"`
	PublicKey string `json:"publicKey" binding:"required"`
}

// handleOTPDeliver sends the OTP code to the user's email address.
//
// The plugin generates the OTP via helper.GenerateOTPFunc() and stores
// it in nhp-server's SQLite keystore; we have no record of the code on
// the demo-app side, we just email whatever string the plugin passes.
func (a *App) handleOTPDeliver(c *gin.Context) {
	var req OTPDeliverRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	subject := a.Cfg.SMTP.Subject
	if a.Cfg.SMTP.CodeInSubject == nil || *a.Cfg.SMTP.CodeInSubject {
		subject = strings.TrimRight(subject, ": ") + ": " + req.OtpCode
	}

	body := fmt.Sprintf(
		"Hello,\n\n"+
			"Your OpenNHP Demo verification code is: %s\n\n"+
			"This code expires in %d minutes. If you didn't request this, "+
			"please ignore this email — your account is safe.\n\n"+
			"Learn more about OpenNHP at https://opennhp.org\n",
		req.OtpCode, a.Cfg.OtpTTLMinutes,
	)

	if err := a.Mailer.Send(req.Email, subject, body); err != nil {
		c.JSON(http.StatusBadGateway, gin.H{"error": "delivery failed: " + err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"ok": true})
}

// handleMarkRegistered records that the user has successfully registered
// their NHP public key with nhp-server. The browser displays
// "Registration complete" only after both this call and the matching
// NHP_RAK have landed.
//
// We accept the publicKey from the request body and require it to match
// the one we stored during /api/users/register. This prevents a stale
// plugin call (e.g., from a previous registration attempt) from
// spuriously marking the just-registered user as complete.
func (a *App) handleMarkRegistered(c *gin.Context) {
	var req MarkRegisteredRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	deadline := time.Now().Add(5 * time.Second)
	for {
		err := MarkNHPRegistered(c.Request.Context(), a.DB, req.UserID, req.PublicKey)
		if err == nil {
			c.JSON(http.StatusOK, gin.H{"ok": true})
			return
		}
		// SQLite can transiently report "database is locked" under
		// concurrent writer load; one short retry is enough on the
		// contended-but-tiny SQLite we use here.
		if time.Now().After(deadline) || !strings.Contains(err.Error(), "locked") {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		time.Sleep(20 * time.Millisecond)
	}
}

// handleLookupEmail returns the email address on file for the given
// userId. Used by demo-app-plugin's RequestOTP fallback when the
// browser didn't supply an email (e.g., OIDC re-keying). The plugin
// already authenticated via X-Plugin-Api-Key, so this endpoint just
// does the lookup.
func (a *App) handleLookupEmail(c *gin.Context) {
	userId := c.Query("userId")
	if userId == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "userId is required"})
		return
	}
	u, err := FindByUsername(c.Request.Context(), a.DB, userId)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"email": u.Email})
}
