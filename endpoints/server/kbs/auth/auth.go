package auth

import (
	"crypto/rand"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"

	kbsError "github.com/OpenNHP/opennhp/endpoints/server/kbs/error"
)

type AuthRequest struct {
	Version     string         `json:"version"`
	Tee         string         `json:"tee"`
	ExtraParams map[string]any `json:"extra-params"`
}

type AuthResponse struct {
	Nonce       string `json:"nonce"`
	ExtraParams string `json:"extra-params"`
}

func generateNonce() (string, error) {
	b := make([]byte, 32)
	_, err := rand.Read(b)
	if err != nil {
		return "", err
	}

	return base64.StdEncoding.EncodeToString(b), nil
}

func generateSecureSessionID() (string, error) {
	randomBytes := make([]byte, 32)
	_, err := rand.Read(randomBytes)
	if err != nil {
		return "", fmt.Errorf("fail to generate session id: %w", err)

	}

	return hex.EncodeToString(randomBytes), nil
}

func Auth(c *gin.Context) {
	var req AuthRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, kbsError.InvalidRequest(err))
		return
	}

	nonce, err := generateNonce()
	if err != nil {
		c.JSON(http.StatusInternalServerError, kbsError.NonceGenerationFailed(err))
		return
	}

	kbsSessionId, err := generateSecureSessionID()
	if err != nil {
		c.JSON(http.StatusInternalServerError, kbsError.SessionIDGenerationFailed(err))
		return
	}

	c.SetCookie(
		"kbs-session-id",
		kbsSessionId,
		3600, // the unit is second
		"/",
		"",
		true, // Secure: only send over HTTPS
		true,
	)

	c.JSON(http.StatusOK, AuthResponse{
		Nonce:       nonce,
		ExtraParams: "",
	})
}
