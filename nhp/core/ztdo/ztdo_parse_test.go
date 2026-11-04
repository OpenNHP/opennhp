package ztdo

import (
	"bytes"
	"crypto/rand"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"testing"
)

var (
	testGcmKey = bytes.Repeat([]byte{0x42}, 32)
	testAD     = []byte("ztdo-parse-test")
)

// writeZtdo encrypts plaintext into a ZTDO file and returns its path and the
// length of its header, so tests can find the first payload.
func writeZtdo(t testing.TB, plaintext []byte, metadata string, signed bool) (string, int) {
	t.Helper()
	dir := t.TempDir()
	plainPath := filepath.Join(dir, "plain")
	if err := os.WriteFile(plainPath, plaintext, 0o600); err != nil {
		t.Fatal(err)
	}
	z := NewZtdo()
	if err := z.SetNhpServer("nhp.example:62206"); err != nil {
		t.Fatal(err)
	}
	if err := z.SetMetadata(metadata); err != nil {
		t.Fatal(err)
	}
	z.SetCipherConfig(signed, AES256GCM128Tag, CURVE25519)
	cipherPath := filepath.Join(dir, "cipher.ztdo")
	if err := z.EncryptZtdoFile(plainPath, cipherPath, testGcmKey, testAD); err != nil {
		t.Fatal(err)
	}
	return cipherPath, toBuffer(z.header).Len()
}

func decrypt(t testing.TB, path string) ([]byte, error) {
	t.Helper()
	out := filepath.Join(t.TempDir(), "out")
	if err := NewZtdo().DecryptZtdoFile(path, out, testGcmKey, testAD); err != nil {
		return nil, err
	}
	return os.ReadFile(out)
}

// patchFile overwrites bytes at off in the file at path.
func patchFile(t testing.TB, path string, off int, b []byte) {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	copy(data[off:], b)
	if err := os.WriteFile(path, data, 0o600); err != nil {
		t.Fatal(err)
	}
}

// payloadLength encodes n as the 3-byte little-endian payload length field.
func payloadLength(n uint32) []byte {
	return []byte{byte(n), byte(n >> 8), byte(n >> 16)}
}

// allocatedDuring reports how many bytes f allocated.
func allocatedDuring(f func()) uint64 {
	var before, after runtime.MemStats
	runtime.GC()
	runtime.ReadMemStats(&before)
	f()
	runtime.ReadMemStats(&after)
	return after.TotalAlloc - before.TotalAlloc
}

// Valid files with multi-chunk metadata still round-trip.
func TestZtdoRoundTrip(t *testing.T) {
	plaintext := make([]byte, 100_000)
	_, _ = rand.Read(plaintext)
	longMetadata := strings.Repeat("m", MetadataChunkMaxSize+10) // exercises lengthContinue
	for _, size := range []int{1, len(plaintext)} {
		for _, metadata := range []string{"", "short", longMetadata} {
			path, _ := writeZtdo(t, plaintext[:size], metadata, true)
			got, err := decrypt(t, path)
			if err != nil {
				t.Fatalf("size=%d metadata=%d bytes: %v", size, len(metadata), err)
			}
			if !bytes.Equal(got, plaintext[:size]) {
				t.Fatalf("size=%d metadata=%d bytes: plaintext mismatch", size, len(metadata))
			}
			z := NewZtdo()
			if err := z.ParseHeader(path); err != nil {
				t.Fatalf("ParseHeader: %v", err)
			}
			if string(z.header.GetMetadata()) != metadata {
				t.Fatalf("metadata mismatch: got %d bytes, want %d", len(z.header.GetMetadata()), len(metadata))
			}
		}
	}
}

// A payload length shorter than the IV used to wrap around to about 4 GiB
// and be allocated before any authentication. It is now refused.
func TestZtdoPayloadLengthUnderflow(t *testing.T) {
	for _, declared := range []uint32{0, 1, IVSize - 1} {
		path, headerLen := writeZtdo(t, []byte("secret data"), "", true)
		patchFile(t, path, headerLen, payloadLength(declared))

		var err error
		alloc := allocatedDuring(func() { _, err = decrypt(t, path) })
		if err == nil || !strings.Contains(err.Error(), "smaller than") {
			t.Fatalf("declared length %d: err = %v, want an underflow refusal", declared, err)
		}
		if alloc > 16<<20 {
			t.Fatalf("declared length %d: allocated %d bytes", declared, alloc)
		}
	}
}

// A payload length larger than what is left in the file is refused before
// the buffer is allocated.
func TestZtdoPayloadLengthBeyondFile(t *testing.T) {
	path, headerLen := writeZtdo(t, []byte("secret data"), "", true)
	patchFile(t, path, headerLen, payloadLength(1<<24-1)) // the largest a 3-byte length can claim

	var err error
	alloc := allocatedDuring(func() { _, err = decrypt(t, path) })
	if err == nil || !strings.Contains(err.Error(), "remain") {
		t.Fatalf("err = %v, want a remaining-bytes refusal", err)
	}
	if alloc > 4<<20 {
		t.Fatalf("allocated %d bytes for a %d-byte claim in a tiny file", alloc, 1<<24-1)
	}
}

// Truncated input is an error, never silently zero-filled.
func TestZtdoTruncated(t *testing.T) {
	path, headerLen := writeZtdo(t, bytes.Repeat([]byte("x"), 5000), "some metadata", true)
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	for _, cut := range []int{3, headerLen - 1, headerLen + 2, headerLen + 20, len(data) - 1} {
		short := filepath.Join(t.TempDir(), "short.ztdo")
		if err := os.WriteFile(short, data[:cut], 0o600); err != nil {
			t.Fatal(err)
		}
		if _, err := decrypt(t, short); err == nil {
			t.Fatalf("file cut to %d of %d bytes decrypted without error", cut, len(data))
		}
		if cut < headerLen {
			if err := NewZtdo().ParseHeader(short); err == nil {
				t.Fatalf("header cut to %d of %d bytes parsed without error", cut, headerLen)
			}
		}
	}
}

// A metadata length that runs past the end of the file is refused.
func TestZtdoMetadataLengthBeyondFile(t *testing.T) {
	path, _ := writeZtdo(t, []byte("x"), "", true)
	// Header layout: magic(4) id(16) version(2) nhpServerLen(1) nhpServer(n) cipherConfig(1) metadataLen(2)...
	metadataLenOff := MagicNumberSize + ObjectIDSize + VersionSize + NhpServerLenSize + len("nhp.example:62206") + CipherConfigSize
	patchFile(t, path, metadataLenOff, []byte{0xff, 0x7f}) // 32767, no continuation
	if err := NewZtdo().ParseHeader(path); err == nil || !strings.Contains(err.Error(), "remain") {
		t.Fatalf("err = %v, want a remaining-bytes refusal", err)
	}
}

// A continued metadata chunk that is not full is refused: the encoder never
// writes one, and a run of empty continued chunks would otherwise turn every
// 2 bytes of input into a whole slice element.
func TestZtdoContinuedMetadataChunkMustBeFull(t *testing.T) {
	path, _ := writeZtdo(t, []byte("x"), "", true)
	metadataLenOff := MagicNumberSize + ObjectIDSize + VersionSize + NhpServerLenSize + len("nhp.example:62206") + CipherConfigSize
	patchFile(t, path, metadataLenOff, []byte{0x00, 0x80}) // length 0, continuation set
	if err := NewZtdo().ParseHeader(path); err == nil || !strings.Contains(err.Error(), "continued") {
		t.Fatalf("err = %v, want a continued-chunk refusal", err)
	}
}

// A file that fails to decrypt or verify leaves no plaintext behind.
func TestZtdoDecryptFailureRemovesOutput(t *testing.T) {
	path, _ := writeZtdo(t, []byte("secret data"), "", true)
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	patchFile(t, path, len(data)-1, []byte{data[len(data)-1] ^ 0xff}) // corrupt the signature only
	out := filepath.Join(t.TempDir(), "out")
	if err := NewZtdo().DecryptZtdoFile(path, out, testGcmKey, testAD); err == nil {
		t.Fatal("corrupted signature verified")
	}
	if _, err := os.Stat(out); !os.IsNotExist(err) {
		t.Fatalf("plaintext left behind after failed verification: stat err = %v", err)
	}
}

// Parses run concurrently without shared state; run with -race.
func TestZtdoConcurrentParse(t *testing.T) {
	paths := make([]string, 4)
	for i := range paths {
		paths[i], _ = writeZtdo(t, bytes.Repeat([]byte{byte(i)}, 3000), strings.Repeat("m", 100*i), true)
	}
	var wg sync.WaitGroup
	for i := 0; i < 16; i++ {
		wg.Add(1)
		go func(path string) {
			defer wg.Done()
			if err := NewZtdo().ParseHeader(path); err != nil {
				t.Error(err)
			}
			if _, err := decrypt(t, path); err != nil {
				t.Error(err)
			}
		}(paths[i%len(paths)])
	}
	wg.Wait()
}

// The parser never panics, never consumes more than it was given, and never
// allocates more than the input could hold, whatever the bytes are.
func FuzzZtdoParse(f *testing.F) {
	path, _ := writeZtdo(f, []byte("seed plaintext"), "seed metadata", true)
	seed, err := os.ReadFile(path)
	if err != nil {
		f.Fatal(err)
	}
	f.Add(seed)
	f.Add([]byte{})
	f.Add(payloadLength(0))
	f.Fuzz(func(t *testing.T, data []byte) {
		for _, target := range []any{NewZtdoHeader(), &ZtdoPayload{}, &ZtdoSignature{}} {
			r := bytes.NewReader(data)
			if err := toStructure(r, int64(len(data)), target); err == nil {
				// What was parsed must re-serialize to exactly the bytes
				// consumed: DecryptZtdoFile relies on that for both the
				// signature hash and its remaining-size bookkeeping.
				consumed := len(data) - r.Len()
				if got := toBuffer(target).Bytes(); !bytes.Equal(got, data[:consumed]) {
					t.Fatalf("parsed %d bytes but re-serialized to %d different bytes", consumed, len(got))
				}
			}
		}
	})
}

// An empty plaintext encrypts to a header and signature with no payloads;
// it decrypts to an empty file, and its signature is still checked.
func TestZtdoEmptyPlaintext(t *testing.T) {
	path, headerLen := writeZtdo(t, nil, "meta", true)
	got, err := decrypt(t, path)
	if err != nil || len(got) != 0 {
		t.Fatalf("empty plaintext: got %d bytes, err %v", len(got), err)
	}
	patchFile(t, path, headerLen-1, []byte{'X'}) // tamper with the header's metadata
	if _, err := decrypt(t, path); err == nil || !strings.Contains(err.Error(), "signature") {
		t.Fatalf("tampered empty file: err = %v, want a signature failure", err)
	}
}

// Files whose header says unsigned are refused: clearing the flag and
// dropping the signature must not skip the whole-file check. That includes
// a header-only file, which would otherwise decrypt to an empty plaintext.
func TestZtdoUnsignedRefused(t *testing.T) {
	cipherConfigOff := MagicNumberSize + ObjectIDSize + VersionSize + NhpServerLenSize + len("nhp.example:62206")
	for _, plaintext := range [][]byte{[]byte("secret data"), nil} {
		path, _ := writeZtdo(t, plaintext, "", true)
		data, err := os.ReadFile(path)
		if err != nil {
			t.Fatal(err)
		}
		// The downgrade: clear the signature flag and drop the signature.
		data[cipherConfigOff] &^= 0x80
		data = data[:len(data)-SIGNATURELenSize]
		if err := os.WriteFile(path, data, 0o600); err != nil {
			t.Fatal(err)
		}
		if _, err := decrypt(t, path); err == nil || !strings.Contains(err.Error(), "unsigned") {
			t.Fatalf("plaintext %d bytes: err = %v, want an unsigned-file refusal", len(plaintext), err)
		}
	}

	// A file written unsigned by the encryptor is refused the same way.
	path, _ := writeZtdo(t, []byte("x"), "", false)
	if _, err := decrypt(t, path); err == nil || !strings.Contains(err.Error(), "unsigned") {
		t.Fatalf("encrypted unsigned: err = %v, want an unsigned-file refusal", err)
	}
}
