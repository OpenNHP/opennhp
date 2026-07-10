package core

import (
	"bytes"
	"testing"
)

func TestKeyGenReusesSameKeyHMAC(t *testing.T) {
	for _, hashType := range []HashTypeEnum{HASH_BLAKE2S, HASH_SM3, HASH_SHA256} {
		t.Run(hashTypeName(hashType), func(t *testing.T) {
			noise := NoiseFactory{HashType: hashType}
			key := []byte("key material")
			input := []byte("input material")

			var got0, got1, got2 [HashSize]byte
			var want0, want1, want2 [HashSize]byte
			noise.KeyGen3(&got0, &got1, &got2, key, input)
			referenceKeyGen3(&noise, &want0, &want1, &want2, key, input)
			if !bytes.Equal(got0[:], want0[:]) || !bytes.Equal(got1[:], want1[:]) || !bytes.Equal(got2[:], want2[:]) {
				t.Fatal("KeyGen3 output changed after HMAC reuse")
			}

			optimized2 := testing.AllocsPerRun(1000, func() {
				noise.KeyGen2(&got0, &got1, key, input)
			})
			reference2 := testing.AllocsPerRun(1000, func() {
				referenceKeyGen2(&noise, &want0, &want1, key, input)
			})
			if optimized2 >= reference2 {
				t.Fatalf("KeyGen2 allocations = %.0f, reference = %.0f; HMAC reuse saved nothing", optimized2, reference2)
			}

			optimized3 := testing.AllocsPerRun(1000, func() {
				noise.KeyGen3(&got0, &got1, &got2, key, input)
			})
			reference3 := testing.AllocsPerRun(1000, func() {
				referenceKeyGen3(&noise, &want0, &want1, &want2, key, input)
			})
			if optimized3 >= reference3 {
				t.Fatalf("KeyGen3 allocations = %.0f, reference = %.0f; HMAC reuse saved nothing", optimized3, reference3)
			}
		})
	}
}

func referenceKeyGen2(n *NoiseFactory, dst0, dst1 *[HashSize]byte, key, input []byte) {
	var prk [HashSize]byte
	n.HMAC1(&prk, key, input)
	n.HMAC1(dst0, prk[:], []byte{0x1})
	n.HMAC2(dst1, prk[:], dst0[:], []byte{0x2})
	SetZero(prk[:])
}

func referenceKeyGen3(n *NoiseFactory, dst0, dst1, dst2 *[HashSize]byte, key, input []byte) {
	var prk [HashSize]byte
	n.HMAC1(&prk, key, input)
	n.HMAC1(dst0, prk[:], []byte{0x1})
	n.HMAC2(dst1, prk[:], dst0[:], []byte{0x2})
	n.HMAC2(dst2, prk[:], dst1[:], []byte{0x3})
	SetZero(prk[:])
}

func hashTypeName(hashType HashTypeEnum) string {
	switch hashType {
	case HASH_BLAKE2S:
		return "BLAKE2s"
	case HASH_SM3:
		return "SM3"
	case HASH_SHA256:
		return "SHA256"
	default:
		return "unknown"
	}
}
