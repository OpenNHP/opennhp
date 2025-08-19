package attest

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/rsa"
	"encoding/base64"
	"errors"
	"fmt"
	"math/big"
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"

	kbsError "github.com/OpenNHP/opennhp/endpoints/server/kbs/error"
)

var (
	teePubKeys = struct {
		sync.RWMutex
		data map[string]*rsa.PublicKey
	}{data: make(map[string]*rsa.PublicKey)}

	jwtSigningKey *ecdsa.PrivateKey
	initKeyOnce   sync.Once
)

type AttestRequest struct {
	TeePubkey   TeePubkey   `json:"tee-pubkey"`
	TeeEvidence map[string]any `json:"tee-evidence"`
}

type TeePubkey struct {
	Kty string `json:"kty"`
	Alg string `json:"alg"`
	N   string `json:"n"`
	E   string `json:"e"`
}

type CustomClaims struct {
	CosignAuthorized bool            `json:"cosign_authorized"`
	jwt.RegisteredClaims
}

func init() {
	initKeyOnce.Do(func() {
		var err error
		jwtSigningKey, err = ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
		if err != nil {
			panic(fmt.Sprintf("fail to generate jwt signing key: %v", err))

		}
	})
}

func Attest(c *gin.Context) {
	sessionID, err := c.Cookie("kbs-session-id")
	if err != nil || sessionID == "" {
		c.JSON(http.StatusUnauthorized, kbsError.MissingOrInvalidSessionID())
		return
	}

	var req AttestRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, kbsError.InvalidRequest(err))
		return
	}

	teePubKey, err := parseTeePubkey(req.TeePubkey)
	if err != nil {
		c.JSON(http.StatusBadRequest, kbsError.TeePubKeyNotFound(err))
		return
	}

	token, err := generateJWT()
	if err != nil {
		c.JSON(http.StatusInternalServerError, kbsError.TokenGenerationFailed(err))
		return
	}

	teePubKeys.Lock()
	teePubKeys.data[token] = teePubKey
	teePubKeys.Unlock()

	c.SetCookie(
		"kbs-session-id",
		sessionID,
		3600,
		"/", "", false, true,
	)

	c.JSON(http.StatusOK, gin.H{
		"token": token,
	})
}

func parseTeePubkey(pubkey TeePubkey) (*rsa.PublicKey, error) {
	if pubkey.Kty != "RSA" {
		return nil, errors.New("unsupported key type, expect RSA")
	}

	nBytes, err := base64.RawURLEncoding.DecodeString(pubkey.N)
	if err != nil {
		return nil, fmt.Errorf("invalid n: %w", err)
	}

	eBytes, err := base64.RawURLEncoding.DecodeString(pubkey.E)
	if err != nil {
		return nil, fmt.Errorf("invalid e: %w", err)
	}

	e := 0
	for _, b := range eBytes {
		e = e<<8 | int(b)
	}

	return &rsa.PublicKey{
		N: new(big.Int).SetBytes(nBytes),
		E: e,
	}, nil
}

func generateJWT() (string, error) {
	if jwtSigningKey == nil {
		return "", errors.New("JWT signing key is not initialized")
	}

	claims := CustomClaims{
		CosignAuthorized: true,
		RegisteredClaims: jwt.RegisteredClaims{
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(5 * time.Minute)),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodES256, claims)

	publicKey := jwtSigningKey.PublicKey
	token.Header["jwk"] = map[string]any{
		"alg": "ES256",
		"crv": "P-256",
		"kty": "EC",
		"x":   base64.RawURLEncoding.EncodeToString(publicKey.X.Bytes()),
		"y":   base64.RawURLEncoding.EncodeToString(publicKey.Y.Bytes()),
	}

	return token.SignedString(jwtSigningKey)
}

func GetTeePubKeyByToken(token string) (*rsa.PublicKey, error) {
	teePubKeys.RLock()
	defer teePubKeys.RUnlock()

	pubKey, exists := teePubKeys.data[token]
	if !exists {
		return nil, errors.New("TEE public key is not found for specified token")

	}
	return pubKey, nil
}
