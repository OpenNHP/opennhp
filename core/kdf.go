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
		return NewHash(n.HashType)
	}
	mac := hmac.New(newHash, key)
	mac.Write(in0)
	mac.Sum(dst[:0])
	mac.Reset()
}

func (n *NoiseFactory) HMAC2(dst *[HashSize]byte, key, in0, in1 []byte) {
	newHash := func() hash.Hash {
		return NewHash(n.HashType)
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
	hash := NewHash(n.HashType)
	hash.Write(key)
	hash.Write(input)
	hash.Sum(dst[:0])
	hash.Reset()
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
