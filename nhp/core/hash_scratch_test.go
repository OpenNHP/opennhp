package core

import (
	"hash"
	"testing"

	"github.com/OpenNHP/opennhp/nhp/common"
)

func TestHashSumScratchBufferDoesNotAllocate(t *testing.T) {
	tests := []struct {
		name     string
		hashType HashTypeEnum
	}{
		{name: "BLAKE2s", hashType: HASH_BLAKE2S},
		{name: "SM3", hashType: HASH_SM3},
		{name: "SHA256", hashType: HASH_SHA256},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			h, err := NewHash(test.hashType)
			if err != nil {
				t.Fatalf("NewHash: %v", err)
			}
			var scratch [HashSize]byte
			input := []byte("allocation regression")

			allocs := testing.AllocsPerRun(1000, func() {
				h.Reset()
				_, _ = h.Write(input)
				if got := len(h.Sum(scratch[:0])); got != h.Size() {
					t.Fatalf("sum length = %d, want %d", got, h.Size())
				}
			})
			if allocs != 0 {
				t.Fatalf("hash.Sum with scratch buffer allocated %.0f times, want 0", allocs)
			}
		})
	}
}

func TestCheckHMACUsesScratchAndRejectsTampering(t *testing.T) {
	var packetBuf PacketBuffer
	packet := &Packet{Buf: &packetBuf, Content: packetBuf[:]}
	header := packet.HeaderWithCipherScheme(common.CIPHER_SCHEME_CURVE)
	header.SetCounter(1)
	header.SetTypeAndPayloadSize(NHP_KNK, 0)

	seed := []byte("hmac test seed")
	prefixLen := header.Size() - HashSize
	expected := mustTestHash(t, HASH_BLAKE2S)
	_, _ = expected.Write(seed)
	_, _ = expected.Write(header.Bytes()[:prefixLen])
	expected.Sum(header.HMACBytes()[:0])

	parser := &PacketParserData{
		header:       header,
		CipherScheme: common.CIPHER_SCHEME_CURVE,
	}
	runCheck := func() bool {
		h := mustTestHash(t, HASH_BLAKE2S)
		_, _ = h.Write(seed)
		parser.hmacHash = h
		return parser.checkHMAC(false)
	}

	if !runCheck() {
		t.Fatal("checkHMAC rejected a valid digest")
	}
	header.HMACBytes()[0] ^= 0xff
	if runCheck() {
		t.Fatal("checkHMAC accepted a tampered digest")
	}
}

func mustTestHash(t *testing.T, hashType HashTypeEnum) hash.Hash {
	t.Helper()
	h, err := NewHash(hashType)
	if err != nil {
		t.Fatalf("NewHash: %v", err)
	}
	return h
}
