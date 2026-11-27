//go:build cgo

package main

import (
	"bytes"
	"crypto/rand"
	"encoding/base64"
	"encoding/binary"
	"errors"
	"strings"
	"sync"
	"testing"

	core "github.com/OpenNHP/opennhp/nhp/core"
	"github.com/OpenNHP/opennhp/nhp/core/scheme/curve"
)

func newTestDevice(t *testing.T) uintptr {
	t.Helper()
	key := core.NewECDH(core.ECC_CURVE25519).PrivateKeyBase64()
	handle, res := testDeviceInit(core.NHP_AGENT, key)
	if res.code != 0 {
		t.Fatalf("nhp_device_init: %d %s", res.code, res.msg)
	}
	t.Cleanup(func() { testDeviceClose(handle) })
	return handle
}

// A malformed packet from an unauthenticated peer must come back as an
// error code, never as a Go panic crossing into the host process. These
// inputs make the core return plain (non-*core.Error) errors, which the
// old single-value type assertion turned into a panic.
func TestDecryptMalformedPacketReturnsError(t *testing.T) {
	handle := newTestDevice(t)

	random := make([]byte, 200)
	if _, err := rand.Read(random); err != nil {
		t.Fatal(err)
	}
	for name, pkt := range map[string][]byte{
		"empty":  {},
		"short":  {0x01, 0x02, 0x03},
		"random": random,
		// The reported case: an allowed header type whose length field
		// disagrees with the packet ("packet total size is incorrect").
		"inconsistent length": craftHeader(core.NHP_ACK, 50, 300),
		// A header type this device does not accept.
		"wrong type": craftHeader(core.NHP_KNK, 0, 300),
	} {
		res := testDecrypt(handle, pkt, len(pkt))
		if res.code == 0 {
			t.Errorf("%s: malformed packet decrypted successfully", name)
		}
		// It must be an ordinary error code, not a panic caught by the
		// recover guard (which only exists as a last line of defense).
		if res.code == testErrRuntimePanic {
			t.Errorf("%s: handled by panicking: %s", name, res.msg)
		}
	}
}

// craftHeader builds a packet of total bytes whose header announces type t
// and body size size (type and size are XOR-masked with the preamble).
func craftHeader(t, size, total int) []byte {
	pkt := make([]byte, total)
	binary.BigEndian.PutUint32(pkt[0:4], 0x5a5a5a5a)
	binary.BigEndian.PutUint32(pkt[4:8], 0x5a5a5a5a^uint32(t<<16|size))
	return pkt
}

// Lengths and pointers from the C caller are checked before they are used.
func TestDecryptRejectsBadBuffers(t *testing.T) {
	handle := newTestDevice(t)
	for name, tc := range map[string]struct {
		pkt []byte
		n   int
	}{
		"negative length":  {[]byte{1, 2, 3}, -1},
		"NULL with length": {nil, 16},
		"empty":            {[]byte{}, 0},
		"short":            {[]byte{1, 2, 3}, 3},
		"one below header": {make([]byte, curve.HeaderSize-1), curve.HeaderSize - 1},
		"over buffer size": {make([]byte, core.PacketBufferSize+1), core.PacketBufferSize + 1},
	} {
		res := testDecrypt(handle, tc.pkt, tc.n)
		if res.code != testErrIncorrectPacketSize {
			t.Errorf("%s: got code %d (%s), want %d", name, res.code, res.msg, testErrIncorrectPacketSize)
		}
	}
}

func TestEncryptRejectsBadBuffers(t *testing.T) {
	handle := newTestDevice(t)
	peer := core.NewECDH(core.ECC_CURVE25519).PublicKey()
	for name, tc := range map[string]struct {
		pk      []byte
		pkLen   int
		data    []byte
		dataLen int
	}{
		"negative data length": {peer, len(peer), []byte("x"), -5},
		"NULL data":            {peer, len(peer), nil, 10},
		"negative key length":  {peer, -1, []byte("x"), 1},
		"oversized key":        {make([]byte, 1024), 1024, []byte("x"), 1},
	} {
		res := testEncrypt(handle, core.NHP_KNK, tc.pk, tc.pkLen, tc.data, tc.dataLen)
		if res.code == 0 {
			t.Errorf("%s: accepted", name)
		}
	}

	// A valid message still encrypts.
	res := testEncrypt(handle, core.NHP_KNK, peer, len(peer), []byte(`{"hello":"world"}`), 17)
	if res.code != 0 || len(res.out) == 0 {
		t.Fatalf("valid encrypt: %d %s", res.code, res.msg)
	}
}

// An unusable key must fail initialization instead of registering a nil
// device under handle 0 and reporting success.
func TestInitRejectsUnusableKey(t *testing.T) {
	bad := base64.StdEncoding.EncodeToString([]byte("too short"))
	handle, res := testDeviceInit(core.NHP_AGENT, bad)
	if res.code != testErrCreateDeviceFailed {
		t.Fatalf("unusable key: got %d (%s), want %d", res.code, res.msg, testErrCreateDeviceFailed)
	}
	if got := testDecrypt(handle, []byte{1}, 1); got.code != testErrDeviceNotInitialized {
		t.Fatalf("handle from a failed init is usable: got %d (%s)", got.code, got.msg)
	}
	if got := testSetOverload(0, true); got.code != testErrDeviceNotInitialized {
		t.Fatalf("handle 0 is registered: got %d (%s)", got.code, got.msg)
	}
}

func TestSM4AeadValidatesInputs(t *testing.T) {
	key := bytes.Repeat([]byte{7}, 16)
	nonce := bytes.Repeat([]byte{9}, 12)
	plain := []byte("sm4 plaintext")
	aad := []byte("aad")

	enc := testSM4Encrypt(key, len(key), nonce, len(nonce), plain, len(plain), aad, len(aad))
	if enc.code != 0 {
		t.Fatalf("valid SM4 encrypt: %d %s", enc.code, enc.msg)
	}
	dec := testSM4Decrypt(key, len(key), nonce, len(nonce), enc.out, len(enc.out), aad, len(aad))
	if dec.code != 0 || !bytes.Equal(dec.out, plain) {
		t.Fatalf("SM4 round trip: %d %s %q", dec.code, dec.msg, dec.out)
	}

	short := nonce[:8]
	for name, res := range map[string]sdkResult{
		// cipher.AEAD.Seal/Open panic on a wrong nonce size.
		"encrypt short nonce": testSM4Encrypt(key, len(key), short, len(short), plain, len(plain), aad, len(aad)),
		"decrypt short nonce": testSM4Decrypt(key, len(key), short, len(short), enc.out, len(enc.out), aad, len(aad)),
		"encrypt short key":   testSM4Encrypt(key[:8], 8, nonce, len(nonce), plain, len(plain), aad, len(aad)),
		"encrypt NULL key":    testSM4Encrypt(nil, 16, nonce, len(nonce), plain, len(plain), aad, len(aad)),
		"encrypt negative":    testSM4Encrypt(key, len(key), nonce, len(nonce), plain, -1, aad, len(aad)),
		"decrypt negative":    testSM4Decrypt(key, len(key), nonce, len(nonce), enc.out, -1, aad, len(aad)),
		"decrypt NULL aad":    testSM4Decrypt(key, len(key), nonce, len(nonce), enc.out, len(enc.out), nil, 4),
	} {
		if res.code == 0 {
			t.Errorf("%s: accepted", name)
		}
	}

	// A NULL key or nonce is reported as such, not as a length problem
	// ("need at least 16 bytes, got 16").
	for name, res := range map[string]sdkResult{
		"key":   testSM4Encrypt(nil, len(key), nonce, len(nonce), plain, len(plain), aad, len(aad)),
		"nonce": testSM4Decrypt(key, len(key), nil, len(nonce), enc.out, len(enc.out), aad, len(aad)),
	} {
		if res.code == 0 || !strings.Contains(res.msg, name+": NULL buffer") {
			t.Errorf("NULL %s: got %d %q", name, res.code, res.msg)
		}
	}
}

// Any panic inside an exported function is reported as
// ERR_NHP_SDK_RUNTIME_PANIC instead of unwinding into the C caller.
func TestRecoverToResult(t *testing.T) {
	res := testRecoverToResult(func() { panic("boom") })
	if res.code != testErrRuntimePanic || res.msg == "" {
		t.Fatalf("recovered panic: got %d %q", res.code, res.msg)
	}
	if ok := testRecoverToResult(func() {}); ok.code != 0 || ok.msg != "" {
		t.Fatalf("no panic: got %d %q", ok.code, ok.msg)
	}
}

func TestFreeFunctionsAcceptNil(t *testing.T) {
	testFreeNil()
}

// A multi-threaded host calls the SDK concurrently; the device registry
// must not be a concurrent map access (a fatal, unrecoverable error).
func TestConcurrentDeviceUse(t *testing.T) {
	var wg sync.WaitGroup
	for i := 0; i < 8; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			key := core.NewECDH(core.ECC_CURVE25519).PrivateKeyBase64()
			handle, res := testDeviceInit(core.NHP_AGENT, key)
			if res.code != 0 {
				t.Errorf("init: %d %s", res.code, res.msg)
				return
			}
			for j := 0; j < 20; j++ {
				testDecrypt(handle, []byte{1, 2, 3}, 3)
				testSetOverload(handle, j%2 == 0)
			}
			testDeviceClose(handle)
		}()
	}
	wg.Wait()
}

// Handles are never reused: a closed device's handle stays invalid even
// after another device is created, instead of reaching that new device.
func TestHandlesAreNotReused(t *testing.T) {
	first := newTestDevice(t)
	testDeviceClose(first)
	second := newTestDevice(t)
	if second == first {
		t.Fatalf("handle %d reused after close", first)
	}
	if res := testSetOverload(first, true); res.code != testErrDeviceNotInitialized {
		t.Fatalf("closed handle still usable: %d %s", res.code, res.msg)
	}
	if res := testSetOverload(second, true); res.code != 0 {
		t.Fatalf("new handle unusable: %d %s", res.code, res.msg)
	}
}

// The registry keeps its own copy of the key string: a C caller may free
// or reuse the buffer it passed as soon as nhp_device_init returns.
func TestInitCopiesCallerKey(t *testing.T) {
	key := core.NewECDH(core.ECC_CURVE25519).PrivateKeyBase64()
	handle, res := testInitWithCallerOwnedKey(core.NHP_AGENT, key)
	if handle != 0 {
		t.Cleanup(func() { testDeviceClose(handle) })
	}
	if res.code != testErrDeviceAlreadyCreated {
		t.Fatalf("re-init with the same key after the caller reused its buffer: got %d (%s), want %d",
			res.code, res.msg, testErrDeviceAlreadyCreated)
	}
}

// A panic the core recovers itself must not hand its stack trace to the C
// caller: on the decrypt path an unauthenticated packet can trigger it.
func TestErrorMessageHidesCorePanicDetails(t *testing.T) {
	core.ErrRuntimePanic.SetExtraError(errors.New("!!!recovered from panic: boom\ngoroutine 1 [running]:\nmain.secret()"))
	t.Cleanup(func() { core.ErrRuntimePanic.SetExtraError(nil) })

	if msg := errorMessage(core.ErrRuntimePanic); strings.Contains(msg, "goroutine") || strings.Contains(msg, "boom") {
		t.Fatalf("core panic details reach the C caller: %q", msg)
	}
	if msg := errorMessage(errors.New("packet total size is incorrect")); msg != "packet total size is incorrect" {
		t.Fatalf("ordinary error message changed: %q", msg)
	}
}
