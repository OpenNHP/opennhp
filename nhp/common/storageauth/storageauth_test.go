package storageauth

import (
	"net/http"
	"testing"
)

func TestSignVerify(t *testing.T) {
	secret := []byte("0123456789abcdef0123456789abcdef")
	r, err := NewRequest(http.MethodPost, "/storage/upload", []byte("body"), "pubkey", "serverkey")
	if err != nil {
		t.Fatal(err)
	}
	if r.ContentLength != 4 || r.BodySHA256 != BodyHash([]byte("body")) || !ValidBodyHash(r.BodySHA256) {
		t.Fatalf("NewRequest body fields: %d %q", r.ContentLength, r.BodySHA256)
	}
	if !ValidNonce(r.Nonce) {
		t.Fatalf("NewNonce produced an invalid nonce %q", r.Nonce)
	}
	sig := Sign(secret, r)
	if !Verify(secret, r, sig) {
		t.Fatal("valid signature rejected")
	}

	// Every covered field changes the signature.
	for name, mutate := range map[string]func(*Request){
		"method":         func(r *Request) { r.Method = http.MethodGet },
		"path":           func(r *Request) { r.Path = "/storage/metadata/x" },
		"content length": func(r *Request) { r.ContentLength++ },
		"public key":     func(r *Request) { r.PublicKeyBase64 = "other" },
		"server key":     func(r *Request) { r.ServerPublicKeyBase64 = "other server" },
		"timestamp":      func(r *Request) { r.Timestamp++ },
		"nonce":          func(r *Request) { r.Nonce = "00000000000000000000000000000000" },
		"body":           func(r *Request) { r.BodySHA256 = BodyHash([]byte("other body")) },
	} {
		m := r
		mutate(&m)
		if Verify(secret, m, sig) {
			t.Errorf("signature still valid after changing the %s", name)
		}
	}
	if Verify([]byte("another shared secret"), r, sig) {
		t.Error("signature valid under another secret")
	}
	if Verify(secret, r, "not base64!") {
		t.Error("garbage signature accepted")
	}
}

func TestSetHeaders(t *testing.T) {
	r := Request{PublicKeyBase64: "pk", Timestamp: 42, Nonce: "n", BodySHA256: "h"}
	h := http.Header{}
	r.SetHeaders(h, "sig")
	for k, want := range map[string]string{
		HeaderPublicKey: "pk", HeaderTimestamp: "42", HeaderNonce: "n", HeaderBodyHash: "h", HeaderSignature: "sig",
	} {
		if got := h.Get(k); got != want {
			t.Errorf("%s = %q, want %q", k, got, want)
		}
	}
}

func TestValidNonce(t *testing.T) {
	for s, want := range map[string]bool{
		"0123456789abcdef0123456789abcdef":  true,
		"":                                  false,
		"0123":                              false,
		"zz23456789abcdef0123456789abcdef":  false,
		"0123456789abcdef0123456789abcdef0": false,
	} {
		if got := ValidNonce(s); got != want {
			t.Errorf("ValidNonce(%q) = %v, want %v", s, got, want)
		}
	}
}
