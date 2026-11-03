// Package storageauth signs and verifies nhp-db requests to the nhp-server
// /storage HTTP API.
//
// There is no new secret to distribute: the key is the ECDH shared secret of
// the nhp-db's static NHP key and the nhp-server's static NHP key, both of
// which the two sides already trust (the server lists the nhp-db in
// db.toml, the nhp-db lists the server in server.toml). A request carries
// the nhp-db public key, a timestamp, a nonce and the SHA-256 of its body,
// and an HMAC over those, the request line and the server's public key (so a
// signature is bound to one server). The HMAC key is derived from the
// shared secret with HKDF-SHA256; the server recomputes it and checks the
// body against the digest.
//
// The signature never reveals the shared secret, so a request observed on a
// plain-HTTP connection does not yield a reusable credential; the timestamp
// window and nonce cache on the server stop it from being replayed, and the
// body digest stops an on-path attacker from swapping the uploaded content.
// The response is not authenticated, so HTTPS is still recommended.
package storageauth

import (
	"crypto/hkdf"
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"net/http"
	"strconv"
	"strings"
	"time"
)

const (
	HeaderPublicKey = "X-NHP-DB-PubKey"
	HeaderTimestamp = "X-NHP-Timestamp"
	HeaderNonce     = "X-NHP-Nonce"
	HeaderSignature = "X-NHP-Signature"
	HeaderBodyHash  = "X-NHP-Content-SHA256"

	// MaxClockSkew is how far a request's timestamp may be from the
	// server's clock, in either direction.
	MaxClockSkew = 5 * time.Minute

	// nonceBytes is the size of a nonce before hex encoding.
	nonceBytes = 16

	// context separates these signatures from any other use of the same
	// ECDH shared secret.
	context = "opennhp-storage-auth-v1"
)

// Request holds the parts of an HTTP request that the signature covers.
type Request struct {
	Method          string
	Path            string // URL path, without query
	ContentLength   int64  // 0 for a request without a body
	PublicKeyBase64 string // the signer's NHP public key
	// ServerPublicKeyBase64 is the NHP public key of the server the request
	// is for. It is signed but not sent: each side knows it.
	ServerPublicKeyBase64 string
	Timestamp             int64  // Unix seconds
	Nonce                 string // hex, see NewNonce
	BodySHA256            string // hex SHA-256 of the request body (EmptyBodySHA256 for none)
}

// EmptyBodySHA256 is the BodySHA256 of a request without a body.
var EmptyBodySHA256 = BodyHash(nil)

// BodyHash returns the hex SHA-256 of body, for Request.BodySHA256.
func BodyHash(body []byte) string {
	sum := sha256.Sum256(body)
	return hex.EncodeToString(sum[:])
}

// ValidBodyHash reports whether s has the form BodyHash produces.
func ValidBodyHash(s string) bool {
	if len(s) != 2*sha256.Size {
		return false
	}
	_, err := hex.DecodeString(s)
	return err == nil
}

// NewRequest fills in a timestamp, a fresh nonce and the body digest.
func NewRequest(method, path string, body []byte, publicKeyBase64, serverPublicKeyBase64 string) (Request, error) {
	nonce, err := NewNonce()
	if err != nil {
		return Request{}, err
	}
	return Request{
		Method:                method,
		Path:                  path,
		ContentLength:         int64(len(body)),
		PublicKeyBase64:       publicKeyBase64,
		ServerPublicKeyBase64: serverPublicKeyBase64,
		Timestamp:             time.Now().Unix(),
		Nonce:                 nonce,
		BodySHA256:            BodyHash(body),
	}, nil
}

// NewNonce returns a random nonce.
func NewNonce() (string, error) {
	b := make([]byte, nonceBytes)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}

// ValidNonce reports whether s has the form NewNonce produces.
func ValidNonce(s string) bool {
	if len(s) != 2*nonceBytes {
		return false
	}
	_, err := hex.DecodeString(s)
	return err == nil
}

func (r Request) canonical() []byte {
	return []byte(strings.Join([]string{
		context,
		strings.ToUpper(r.Method),
		r.Path,
		strconv.FormatInt(r.ContentLength, 10),
		r.PublicKeyBase64,
		r.ServerPublicKeyBase64,
		strconv.FormatInt(r.Timestamp, 10),
		r.Nonce,
		r.BodySHA256,
	}, "\n"))
}

// macKey derives the HMAC key from the raw ECDH output, rather than using
// that output directly.
func macKey(sharedSecret []byte) []byte {
	// hkdf.Key only fails for an output longer than 255 hash blocks.
	key, _ := hkdf.Key(sha256.New, sharedSecret, nil, context, sha256.Size)
	return key
}

func mac(sharedSecret []byte, r Request) []byte {
	m := hmac.New(sha256.New, macKey(sharedSecret))
	m.Write(r.canonical())
	return m.Sum(nil)
}

// Sign returns the base64 signature of r under sharedSecret.
func Sign(sharedSecret []byte, r Request) string {
	return base64.StdEncoding.EncodeToString(mac(sharedSecret, r))
}

// Verify reports whether sig is r's signature under sharedSecret. It does
// not check the timestamp or the nonce; the caller does.
func Verify(sharedSecret []byte, r Request, sig string) bool {
	got, err := base64.StdEncoding.DecodeString(sig)
	if err != nil {
		return false
	}
	return hmac.Equal(got, mac(sharedSecret, r))
}

// SetHeaders adds r and its signature to h.
func (r Request) SetHeaders(h http.Header, sig string) {
	h.Set(HeaderPublicKey, r.PublicKeyBase64)
	h.Set(HeaderTimestamp, strconv.FormatInt(r.Timestamp, 10))
	h.Set(HeaderNonce, r.Nonce)
	h.Set(HeaderBodyHash, r.BodySHA256)
	h.Set(HeaderSignature, sig)
}
