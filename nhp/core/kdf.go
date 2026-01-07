package core

import (
	"crypto/hmac"
	"crypto/subtle"
	"hash"
)

type NoiseFactory struct {
	HashType HashTypeEnum
}

func (n *NoiseFactory) HMAC1(dst *[HashSize]byte, key, in0 []byte) {
	newHash := func() hash.Hash {
		h, err := NewHash(n.HashType)
		if err != nil {
			panic("failed to create hash for HMAC: " + err.Error())
		}
		return h
	}
	mac := hmac.New(newHash, key)
	mac.Write(in0)
	mac.Sum(dst[:0])
	mac.Reset()
}

func (n *NoiseFactory) HMAC2(dst *[HashSize]byte, key, in0, in1 []byte) {
	newHash := func() hash.Hash {
		h, err := NewHash(n.HashType)
		if err != nil {
			panic("failed to create hash for HMAC: " + err.Error())
		}
		return h
	}
	mac := hmac.New(newHash, key)
	mac.Write(in0)
	mac.Write(in1)
	mac.Sum(dst[:0])
	mac.Reset()
}

func (n *NoiseFactory) KeyGen1(dst0 *[HashSize]byte, key, input []byte) {
	n.HMAC1(dst0, key, input)
	n.HMAC1(dst0, dst0[:], []byte{0x1})
}

func (n *NoiseFactory) KeyGen2(dst0, dst1 *[HashSize]byte, key, input []byte) {
	var prk [HashSize]byte
	n.HMAC1(&prk, key, input)
	n.HMAC1(dst0, prk[:], []byte{0x1})
	n.HMAC2(dst1, prk[:], dst0[:], []byte{0x2})
	SetZero(prk[:])
}

func (n *NoiseFactory) KeyGen3(dst0, dst1, dst2 *[HashSize]byte, key, input []byte) {
	var prk [HashSize]byte
	n.HMAC1(&prk, key, input)
	n.HMAC1(dst0, prk[:], []byte{0x1})
	n.HMAC2(dst1, prk[:], dst0[:], []byte{0x2})
	n.HMAC2(dst2, prk[:], dst1[:], []byte{0x3})
	SetZero(prk[:])
}

func (n *NoiseFactory) MixKey(dst *[SymmetricKeySize]byte, key []byte, input []byte) {
	n.KeyGen1(dst, key, input)
}

func (n *NoiseFactory) MixHash(dst *[HashSize]byte, key []byte, input []byte) {
	h, err := NewHash(n.HashType)
	if err != nil {
		panic("failed to create hash for MixHash: " + err.Error())
	}
	h.Write(key)
	h.Write(input)
	h.Sum(dst[:0])
	h.Reset()
}

func SetZero(arr []byte) {
	for i := range arr {
		arr[i] = 0
	}
}

func IsZero(arr []byte) bool {
	for _, b := range arr {
		r := subtle.ConstantTimeByteEq(b, 0)
		if r != 1 {
			return false
		}
	}
	return true
}
