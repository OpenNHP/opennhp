package resource

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/rsa"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"syscall"

	"github.com/gin-gonic/gin"
	"github.com/sigstore/cosign/v2/pkg/cosign"

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

	if err := os.WriteFile(publicKeyPath, keys.PublicBytes, 0644); err != nil { //nolint:gosec // G306: Public keys are intentionally world-readable
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

	claims, err := VerifyJWT(tokenStr)
	if err != nil {
		c.JSON(http.StatusUnauthorized, kbsError.TokenInvalid())
		return
	}

	// Unconditional, not "if the claims happen to assert". The previous
	// shape wrapped this in a type assertion that could not succeed, so a
	// token without the claim was served anyway.
	if !claims.CosignAuthorized {
		c.JSON(http.StatusForbidden, kbsError.PolicyDeny())
		return
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
	if _, randErr := rand.Read(contentKey); randErr != nil {
		c.JSON(http.StatusInternalServerError, kbsError.KeyGenerationFailed(randErr))
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
	// Gin catch-all parameters include one leading slash. Keep the remainder
	// relative so os.Root still rejects traversal and escaping symlinks.
	resourceID = strings.TrimPrefix(resourceID, "/")
	if !filepath.IsLocal(resourceID) {
		return nil, errors.New("invalid resource ID: potential path traversal attack")
	}
	root, err := os.OpenRoot(baseDir)
	if err != nil {
		return nil, fmt.Errorf("fail to open resource directory: %w", err)
	}
	defer func() { _ = root.Close() }()
	f, err := root.OpenFile(resourceID, os.O_RDONLY|syscall.O_NONBLOCK, 0)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, errors.New("resource not found")
		}
		return nil, fmt.Errorf("fail to open resource: %w", err)
	}
	defer func() { _ = f.Close() }()

	fileInfo, err := f.Stat()
	if err != nil {
		return nil, fmt.Errorf("fail to stat resource: %w", err)
	}
	if !fileInfo.Mode().IsRegular() {
		return nil, errors.New("resource is not a regular file")
	}

	data, err := io.ReadAll(f)
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

// VerifyJWT verifies a KBS token against the signing key this process
// generated at startup and returns its typed claims.
//
// It used to read the verification key out of the token's own "jwk"
// header, which made every self-signed token valid. See
// attest.VerifyToken.
func VerifyJWT(tokenString string) (*attest.CustomClaims, error) {
	return attest.VerifyToken(tokenString)
}
