package attest

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"encoding/base64"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// A token is only valid if THIS process signed it.
//
// The check used to take its verification key from the token's own "jwk"
// header, so anyone could generate a keypair, sign a token claiming
// cosign_authorized, publish the public half in the header, and have it
// verify. The signature proved nothing; this test is what says so.
func TestVerifyTokenRejectsForeignKeys(t *testing.T) {
	foreign, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		t.Fatal(err)
	}

	tok := jwt.NewWithClaims(jwt.SigningMethodES256, CustomClaims{CosignAuthorized: true})
	// The header the old code trusted.
	tok.Header["jwk"] = map[string]any{
		"kty": "EC", "crv": "P-256",
		"x": base64.RawURLEncoding.EncodeToString(foreign.X.Bytes()),
		"y": base64.RawURLEncoding.EncodeToString(foreign.Y.Bytes()),
	}
	signed, err := tok.SignedString(foreign)
	if err != nil {
		t.Fatal(err)
	}

	if _, verr := VerifyToken(signed); verr == nil {
		t.Fatal("a token signed with a caller-chosen key was accepted")
	}

	// Without the jwk header either — nothing about the token may select
	// the key it is checked against.
	plain := jwt.NewWithClaims(jwt.SigningMethodES256, CustomClaims{CosignAuthorized: true})
	signed, err = plain.SignedString(foreign)
	if err != nil {
		t.Fatal(err)
	}
	if _, verr := VerifyToken(signed); verr == nil {
		t.Fatal("a token signed with a foreign key was accepted")
	}
}

// A token this process issued verifies, and its claims come back typed.
//
// The typing is the point: the caller used to type-assert jwt.Parse's
// result to *CustomClaims, which never succeeded, so the
// cosign_authorized check it guarded was dead.
func TestVerifyTokenAcceptsOurOwnAndTypesTheClaims(t *testing.T) {
	issued, err := generateJWT()
	if err != nil {
		t.Fatalf("generateJWT: %v", err)
	}
	claims, err := VerifyToken(issued)
	if err != nil {
		t.Fatalf("a token this process issued was refused: %v", err)
	}
	if !claims.CosignAuthorized {
		t.Error("cosign_authorized did not survive the round trip")
	}

	// A token signed by us but asserting nothing must be distinguishable,
	// so the caller can refuse it.
	unauth := jwt.NewWithClaims(jwt.SigningMethodES256, CustomClaims{CosignAuthorized: false})
	signed, err := unauth.SignedString(jwtSigningKey)
	if err != nil {
		t.Fatal(err)
	}
	claims, err = VerifyToken(signed)
	if err != nil {
		t.Fatalf("VerifyToken: %v", err)
	}
	if claims.CosignAuthorized {
		t.Error("cosign_authorized reported true for a token that does not carry it")
	}
}

// Only ES256. Without an algorithm allowlist, "alg":"none" and the HMAC
// confusion family become candidates.
func TestVerifyTokenRejectsOtherAlgorithms(t *testing.T) {
	tok := jwt.NewWithClaims(jwt.SigningMethodNone, CustomClaims{CosignAuthorized: true})
	signed, err := tok.SignedString(jwt.UnsafeAllowNoneSignatureType)
	if err != nil {
		t.Fatal(err)
	}
	if _, verr := VerifyToken(signed); verr == nil {
		t.Fatal(`a token with alg "none" was accepted`)
	}
}

// An expired token is refused. ParseWithClaims into *CustomClaims is what
// brings jwt/v5's RegisteredClaims validation back; the previous path
// yielded jwt.MapClaims, so nothing downstream looked at exp either.
func TestVerifyTokenRejectsExpiredTokens(t *testing.T) {
	past := time.Now().Add(-2 * time.Hour)
	expired := jwt.NewWithClaims(jwt.SigningMethodES256, CustomClaims{
		CosignAuthorized: true,
		RegisteredClaims: jwt.RegisteredClaims{
			IssuedAt:  jwt.NewNumericDate(past),
			ExpiresAt: jwt.NewNumericDate(past.Add(5 * time.Minute)),
		},
	})
	signed, err := expired.SignedString(jwtSigningKey)
	if err != nil {
		t.Fatal(err)
	}
	if _, verr := VerifyToken(signed); verr == nil {
		t.Fatal("an expired token was accepted")
	}
}
