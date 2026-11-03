package agent

import (
	"bytes"
	"crypto/ed25519"
	"crypto/rand"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	toml "github.com/pelletier/go-toml/v2"

	"github.com/OpenNHP/opennhp/nhp/log"
)

// DHP web console hardening.
//
// The console exposes privileged operations (rewriting the nhp-server trust
// anchor, rotating keys, restarting the agent, uploading and executing
// trusted applications). It therefore:
//
//   - listens on loopback by default (ConsoleConfig.ListenAddr);
//   - requires a bearer token on every route, including the dynamic TA
//     routes served through NoRoute;
//   - only executes uploaded TAs that carry a valid ed25519 signature from
//     an operator key listed in ConsoleConfig.TrustedTAPublicKeys.
//
// The agent's own TEE posture (safeTee) is still enforced on TA routes, but
// only as an additional check after the caller has been authenticated —
// device posture is not caller identity.

const (
	consoleConfigFileName    = "console.toml"
	consoleTokenFileName     = "console.token"
	defaultConsoleListenAddr = "127.0.0.1:443"

	// consoleTokenBytes is the amount of entropy in a generated token.
	consoleTokenBytes = 32

	// consoleShutdownTimeout bounds how long CloseDHPWebConsole waits for
	// in-flight console requests before closing their connections.
	consoleShutdownTimeout = 5 * time.Second

	// minConsoleTokenLen is the shortest operator-supplied token accepted.
	// A generated token is 2*consoleTokenBytes hex characters.
	minConsoleTokenLen = 32
)

// maxTAUploadSize caps the request body accepted by the TA register
// endpoint so an authenticated-but-careless client cannot fill the disk.
// A variable only so tests can lower it.
var maxTAUploadSize int64 = 256 << 20

// maxConsoleBodySize caps the request body of every other console route
// (JSON: the server peer table, TA call parameters). The console server
// sets no ReadTimeout (TA uploads may be slow), so without a cap a token
// holder could stream an arbitrarily large JSON body into memory.
var maxConsoleBodySize int64 = 1 << 20

var (
	errTASignatureMissing = errors.New("trusted application signature is required")
	errTASignatureInvalid = errors.New("trusted application signature is invalid")
	errTARevoked          = errors.New("trusted application is revoked (RevokedTADigests in etc/console.toml)")
	errTANoTrustedKeys    = errors.New("trusted application registration is disabled: no TrustedTAPublicKeys configured in etc/console.toml")
)

// ConsoleConfig is loaded from etc/console.toml. The file is optional; every
// field has a secure default.
type ConsoleConfig struct {
	// ListenAddr is the host:port the HTTPS console binds to. Defaults to
	// loopback so the console is never reachable from the network unless an
	// operator opts in explicitly.
	ListenAddr string `toml:"ListenAddr"`

	// TrustedTAPublicKeys lists base64-encoded ed25519 public keys whose
	// signatures are accepted on uploaded trusted applications. When empty,
	// TA registration is refused.
	TrustedTAPublicKeys []string `toml:"TrustedTAPublicKeys"`

	// RevokedTADigests lists hex SHA-256 digests of TA binaries that must
	// no longer be registered even though they carry a valid signature
	// (e.g. an old, vulnerable version).
	RevokedTADigests []string `toml:"RevokedTADigests"`

	trustedTAKeys []ed25519.PublicKey
	revokedTA     map[string]struct{}
}

// isTARevoked reports whether a TA with this SHA-256 digest is revoked.
func (c *ConsoleConfig) isTARevoked(digest []byte) bool {
	if c == nil {
		return false
	}
	_, revoked := c.revokedTA[hex.EncodeToString(digest)]
	return revoked
}

// loadConsoleConfig reads etc/console.toml from etcDir. A missing file yields
// the defaults; a malformed file or an invalid key is an error so that a
// typo never silently widens exposure.
func loadConsoleConfig(etcDir string) (*ConsoleConfig, error) {
	conf := &ConsoleConfig{}

	content, err := os.ReadFile(filepath.Join(etcDir, consoleConfigFileName))
	switch {
	case err == nil:
		// Reject unknown keys: a misspelled setting (e.g. "ListenAdr") must
		// fail loudly rather than silently fall back to a default.
		dec := toml.NewDecoder(bytes.NewReader(content)).DisallowUnknownFields()
		if decodeErr := dec.Decode(conf); decodeErr != nil {
			var strictErr *toml.StrictMissingError
			if errors.As(decodeErr, &strictErr) {
				keys := make([]string, 0, len(strictErr.Errors))
				for _, e := range strictErr.Errors {
					keys = append(keys, strings.Join(e.Key(), "."))
				}
				return nil, fmt.Errorf("parse %s: unknown key(s) %s", consoleConfigFileName, strings.Join(keys, ", "))
			}
			return nil, fmt.Errorf("parse %s: %w", consoleConfigFileName, decodeErr)
		}
	case errors.Is(err, os.ErrNotExist):
		// defaults below
	default:
		return nil, fmt.Errorf("read %s: %w", consoleConfigFileName, err)
	}

	conf.ListenAddr = strings.TrimSpace(conf.ListenAddr)
	if conf.ListenAddr == "" {
		conf.ListenAddr = defaultConsoleListenAddr
	}

	for i, keyB64 := range conf.TrustedTAPublicKeys {
		raw, decErr := base64.StdEncoding.DecodeString(strings.TrimSpace(keyB64))
		if decErr != nil || len(raw) != ed25519.PublicKeySize {
			return nil, fmt.Errorf("%s: TrustedTAPublicKeys[%d] is not a base64 ed25519 public key", consoleConfigFileName, i)
		}
		conf.trustedTAKeys = append(conf.trustedTAKeys, ed25519.PublicKey(raw))
	}

	conf.revokedTA = make(map[string]struct{}, len(conf.RevokedTADigests))
	for i, d := range conf.RevokedTADigests {
		d = strings.ToLower(strings.TrimSpace(d))
		if raw, decErr := hex.DecodeString(d); decErr != nil || len(raw) != sha256.Size {
			return nil, fmt.Errorf("%s: RevokedTADigests[%d] is not a hex SHA-256 digest", consoleConfigFileName, i)
		}
		conf.revokedTA[d] = struct{}{}
	}

	return conf, nil
}

// consoleNotice is a startup message produced before the agent's file
// logger exists (PrepareDHPWebConsole runs before Start). They are logged
// by ServeDHPWebConsole so they reach the persistent agent log.
type consoleNotice struct {
	warning bool
	msg     string
}

func (n consoleNotice) log() {
	if n.warning {
		log.Warning("%s", n.msg)
	} else {
		log.Info("%s", n.msg)
	}
}

// loadOrCreateConsoleToken returns the console bearer token stored in
// etcDir/console.token, generating a random one (mode 0600) on first start.
// A token supplied by the operator must be at least minConsoleTokenLen
// characters.
func loadOrCreateConsoleToken(etcDir string) (string, []consoleNotice, error) {
	tokenPath := filepath.Join(etcDir, consoleTokenFileName)

	content, err := os.ReadFile(tokenPath)
	if err == nil {
		token := strings.TrimSpace(string(content))
		if token == "" {
			return "", nil, fmt.Errorf("%s is empty; delete it to have a new token generated", tokenPath)
		}
		if len(token) < minConsoleTokenLen {
			return "", nil, fmt.Errorf("%s holds a %d-character token, at least %d are required; delete it to have a strong token generated",
				tokenPath, len(token), minConsoleTokenLen)
		}
		var notices []consoleNotice
		// Windows reports 0666/0444 regardless of ACLs, so the POSIX
		// group/other check is meaningless there.
		if info, statErr := os.Stat(tokenPath); runtime.GOOS != "windows" && statErr == nil && info.Mode().Perm()&0o077 != 0 {
			notices = append(notices, consoleNotice{warning: true, msg: fmt.Sprintf(
				"DHP console token file %s is accessible by group/other (mode %v); restrict it to 0600", tokenPath, info.Mode().Perm())})
		}
		return token, notices, nil
	}
	if !errors.Is(err, os.ErrNotExist) {
		return "", nil, fmt.Errorf("read %s: %w", tokenPath, err)
	}

	buf := make([]byte, consoleTokenBytes)
	if _, randErr := rand.Read(buf); randErr != nil {
		return "", nil, fmt.Errorf("generate console token: %w", randErr)
	}
	token := hex.EncodeToString(buf)

	if mkdirErr := os.MkdirAll(etcDir, 0o755); mkdirErr != nil {
		return "", nil, mkdirErr
	}
	// O_EXCL: never clobber a token another process created concurrently.
	f, err := os.OpenFile(tokenPath, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0o600)
	if err != nil {
		return "", nil, fmt.Errorf("create %s: %w", tokenPath, err)
	}
	if _, writeErr := f.WriteString(token + "\n"); writeErr != nil {
		closeErr := f.Close()
		os.Remove(tokenPath)
		return "", nil, fmt.Errorf("write %s: %w", tokenPath, errors.Join(writeErr, closeErr))
	}
	// A failed Close may mean the data never reached disk; do not leave a
	// possibly truncated token behind for the next start to accept.
	if closeErr := f.Close(); closeErr != nil {
		os.Remove(tokenPath)
		return "", nil, fmt.Errorf("write %s: %w", tokenPath, closeErr)
	}

	return token, []consoleNotice{{msg: "generated DHP console token in " + tokenPath}}, nil
}

// consoleAuthMiddleware rejects any request that does not present
// "Authorization: Bearer <token>". Both sides are hashed before the
// constant-time compare so the token length is not leaked either.
func consoleAuthMiddleware(token string) gin.HandlerFunc {
	want := sha256.Sum256([]byte(token))

	return func(c *gin.Context) {
		const prefix = "Bearer "
		header := c.GetHeader("Authorization")
		if len(header) <= len(prefix) || !strings.EqualFold(header[:len(prefix)], prefix) {
			c.Header("WWW-Authenticate", `Bearer realm="nhp-agent"`)
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "missing bearer token"})
			return
		}

		// Leading/trailing whitespace is ignored on both sides (the token
		// file is trimmed too); internal whitespace is significant.
		got := sha256.Sum256([]byte(strings.TrimSpace(header[len(prefix):])))
		if subtle.ConstantTimeCompare(got[:], want[:]) != 1 {
			c.Header("WWW-Authenticate", `Bearer realm="nhp-agent", error="invalid_token"`)
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid bearer token"})
			return
		}

		c.Next()
	}
}

// taSignatureContext domain-separates TA signatures: the signed message is
// taSignatureContext || sha256(file), so a signature made with an operator
// key for some other purpose (e.g. over a bare SHA-256 digest) is never
// accepted as a TA signature, and the format can be versioned.
const taSignatureContext = "opennhp-dhp-ta-v1\x00"

// taSignedMessage returns the message an operator signs for a TA whose
// file has the given SHA-256 digest.
func taSignedMessage(digest []byte) []byte {
	msg := make([]byte, 0, len(taSignatureContext)+len(digest))
	msg = append(msg, taSignatureContext...)
	return append(msg, digest...)
}

// verifyTASignature checks that sigB64 is an ed25519 signature over
// taSignedMessage(digest) by one of the trusted operator keys.
func verifyTASignature(trusted []ed25519.PublicKey, digest []byte, sigB64 string) error {
	if len(trusted) == 0 {
		return errTANoTrustedKeys
	}

	sigB64 = strings.TrimSpace(sigB64)
	if sigB64 == "" {
		return errTASignatureMissing
	}

	sig, err := base64.StdEncoding.DecodeString(sigB64)
	if err != nil || len(sig) != ed25519.SignatureSize {
		return errTASignatureInvalid
	}

	msg := taSignedMessage(digest)
	for _, pk := range trusted {
		if ed25519.Verify(pk, msg, sig) {
			return nil
		}
	}

	return errTASignatureInvalid
}
