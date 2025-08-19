package resource

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/rsa"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math/big"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/sigstore/cosign/v2/pkg/cosign"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"

	"github.com/OpenNHP/opennhp/endpoints/server/kbs/attest"
	kbsError "github.com/OpenNHP/opennhp/endpoints/server/kbs/error"
)

var (
	baseDir = "/opt/confidential-containers/kbs/repository"
)

func init() {
	err := generateCosignKeyPair(
		filepath.Join(baseDir, "cosign.key"),
		filepath.Join(baseDir, "/default/cosign-key/pub"),
	)

	if err != nil {
		panic(err)
	}
}

func generateCosignKeyPair(privateKeyPath, publicKeyPath string) error {
	if _, err := os.Stat(privateKeyPath); err == nil {
		if _, err := os.Stat(publicKeyPath); err == nil {
			return nil
		}
	}

	keys, err := cosign.GenerateKeyPair(nil)
	if err != nil {
		return err
	}

	if err := os.MkdirAll(filepath.Dir(privateKeyPath), 0755); err != nil {
		return fmt.Errorf("fail to create private key directory: %w", err)
	}

	if err := os.MkdirAll(filepath.Dir(publicKeyPath), 0755); err != nil {
		return fmt.Errorf("fail to create public key directory: %w", err)
	}

	if err := os.WriteFile(privateKeyPath, keys.PrivateBytes, 0600); err != nil {
		return fmt.Errorf("fail to write private key file: %w", err)
	}

	if err := os.WriteFile(publicKeyPath, keys.PublicBytes, 0644); err != nil {
		return fmt.Errorf("fail to write public key file: %w", err)
	}

	return nil
}

func GetResource(c *gin.Context) {
	path := c.Param("path")
	if path == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "resource path is empty"})
		return
	}

	authHeader := c.GetHeader("Authorization")
	if authHeader == "" || !strings.HasPrefix(authHeader, "Bearer ") {
		c.JSON(http.StatusUnauthorized, kbsError.TokenNotFound())
		return
	}
	tokenStr := strings.TrimPrefix(authHeader, "Bearer ")

	verifiedToken, err := VerifyJWT(tokenStr)
	if err != nil {
		c.JSON(http.StatusUnauthorized, kbsError.TokenInvalid())
		return
	}

	if !verifiedToken.Valid {
		c.JSON(http.StatusUnauthorized, kbsError.TokenInvalid())
		return
	}

	if claims, ok := verifiedToken.Claims.(*attest.CustomClaims); ok {
		if !claims.CosignAuthorized {
			c.JSON(http.StatusForbidden, kbsError.PolicyDeny())
			return
		}
	}

	teePubKey, err := attest.GetTeePubKeyByToken(tokenStr)
	if err != nil {
		c.JSON(http.StatusForbidden, kbsError.TeePubKeyNotFound(err))
		return
	}

	resourceData, err := loadResource(path)
	if err != nil {
		c.JSON(http.StatusNotFound, kbsError.ResourceNotFound(err))
		return
	}

	contentKey := make([]byte, 32)
	if _, err := rand.Read(contentKey); err != nil {
		c.JSON(http.StatusInternalServerError, kbsError.KeyGenerationFailed(err))
		return
	}

	encryptedKey, err := rsa.EncryptPKCS1v15(
		rand.Reader,
		teePubKey,
		contentKey,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, kbsError.KeyEncryptionFailed(err))
		return
	}

	encryptedContent, iv, _, err := encryptWithA256GCM(contentKey, resourceData)
	if err != nil {
		c.JSON(http.StatusInternalServerError, kbsError.ContentEncryptionFailed(err))
		return
	}

	protected := map[string]string{
		"alg": "RSA1_5",
		"enc": "A256GCM",
	}
	protectedJSON, _ := json.Marshal(protected)

	response := map[string]string{
		"protected":     string(protectedJSON),
		"encrypted_key": base64.RawURLEncoding.EncodeToString(encryptedKey),
		"iv":            base64.RawURLEncoding.EncodeToString(iv),
		"ciphertext":    base64.RawURLEncoding.EncodeToString(encryptedContent),
		"tag":           "",
	}

	c.JSON(http.StatusOK, response)
}

func loadResource(resourceID string) ([]byte, error) {
	absBaseDir, err := filepath.Abs(baseDir)
	if err != nil {
		return nil, fmt.Errorf("fail to get base directory absolute path: %w", err)
	}

	fullPath := filepath.Join(absBaseDir, resourceID)

	absFullPath, err := filepath.Abs(fullPath)
	if err != nil {
		return nil, fmt.Errorf("fail to get resource absolute path: %w", err)
	}

	// Check if the path is within the base directory to avoid path traversal attack.
	if !strings.HasPrefix(absFullPath, absBaseDir) {
		return nil, errors.New("invalid resource ID: potential path traversal attack")
	}

	if _, err := os.Stat(absFullPath); err != nil {
		if os.IsNotExist(err) {
			return nil, errors.New("resource not found")
		}
		return nil, fmt.Errorf("fail to check resource: %w", err)
	}

	data, err := os.ReadFile(absFullPath)
	if err != nil {
		return nil, fmt.Errorf("fail to read resource: %w", err)
	}
	return data, nil
}

func encryptWithA256GCM(key, plaintext []byte) (ciphertext, iv, tag []byte, err error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, nil, nil, err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, nil, nil, err
	}

	iv = make([]byte, gcm.NonceSize())
	if _, err = io.ReadFull(rand.Reader, iv); err != nil {
		return nil, nil, nil, err
	}

	ciphertext = gcm.Seal(nil, iv, plaintext, nil)

	return ciphertext, iv, nil, nil
}

func VerifyJWT(tokenString string) (*jwt.Token, error) {
	// First parse the token without verification to get the header
	parser := jwt.NewParser()
	unverifiedToken, _, err := parser.ParseUnverified(tokenString, jwt.MapClaims{})
	if err != nil {
		return nil, fmt.Errorf("failed to parse token: %w", err)
	}

	// Extract JWK from header
	jwkHeader, ok := unverifiedToken.Header["jwk"].(map[string]any)
	if !ok {
		return nil, fmt.Errorf("missing or invalid jwk in header")
	}

	// Convert JWK back to ECDSA public key
	xStr, ok := jwkHeader["x"].(string)
	if !ok {
		return nil, fmt.Errorf("missing x coordinate in jwk")
	}
	yStr, ok := jwkHeader["y"].(string)
	if !ok {
		return nil, fmt.Errorf("missing y coordinate in jwk")
	}

	xBytes, err := base64.RawURLEncoding.DecodeString(xStr)
	if err != nil {
		return nil, fmt.Errorf("invalid x coordinate: %w", err)
	}
	yBytes, err := base64.RawURLEncoding.DecodeString(yStr)
	if err != nil {
		return nil, fmt.Errorf("invalid y coordinate: %w", err)
	}

	publicKey := &ecdsa.PublicKey{
		Curve: elliptic.P256(),
		X:     new(big.Int).SetBytes(xBytes),
		Y:     new(big.Int).SetBytes(yBytes),
	}

	// Now verify the token with the extracted public key
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (any, error) {
		// Check signing method
		if _, ok := token.Method.(*jwt.SigningMethodECDSA); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return publicKey, nil
	})

	if err != nil {
		return nil, err
	}

	return token, nil
}

