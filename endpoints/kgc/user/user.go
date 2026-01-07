package user

import (
	"encoding/base64"
	"fmt"
	"hash"
	"math/big"

	"github.com/OpenNHP/opennhp/endpoints/kgc"
	"github.com/OpenNHP/opennhp/endpoints/kgc/curve"
)

type UserPartialKey struct {
	PrivateKey *big.Int
	PubX       *big.Int
	PubY       *big.Int
}

type UserFullKey struct {
	PrivateKey *big.Int
	PubX       *big.Int
	PubY       *big.Int
}

func (u UserFullKey) String() string {
	return fmt.Sprintf("UserFullKey{PrivateKey: %x, PubX: %x, PubY: %x}", u.PrivateKey.Bytes(), u.PubX.Bytes(), u.PubY.Bytes())
}

func NewUserFullKeyWithPrivateKey(prkBase64 string) (*UserFullKey, error) {
	prk, err := base64.StdEncoding.DecodeString(prkBase64)
	if err != nil {
		return nil, err
	}

	userFullKey := &UserFullKey{
		PrivateKey: new(big.Int).SetBytes(prk),
	}

	return userFullKey, nil
}

type User interface {
	Params() *curve.Curve
	GenerateUserPartialKey() (*UserPartialKey, error)
	GenerateUserFullKey(kgcUserPartialKey *kgc.KGCUserPartialKey) (*UserFullKey, error)
	VerifyFullKey() error
}

type UserImpl struct {
	curve.Curve
	h         hash.Hash
	masterKey *kgc.MasterKey
}

func NewUser(params curve.Curve, hash hash.Hash, masterKey *kgc.MasterKey) *UserImpl {
	return &UserImpl{
		Curve:     params,
		h:         hash,
		masterKey: masterKey,
	}
}

func (u *UserImpl) Params() *curve.CurveParams {
	return u.Curve.Params()
}

func (u *UserImpl) GenerateUserPartialKey() (*UserPartialKey, error) {
	partialPrivateKey, err := kgc.GenerateRandomNumber(u.Params().N)
	if err != nil {
		return nil, err
	}

	partialPubX, partialPubY := u.Curve.ScalarBaseMult(partialPrivateKey.Bytes())

	return &UserPartialKey{
		PrivateKey: partialPrivateKey,
		PubX:       partialPubX,
		PubY:       partialPubY,
	}, nil
}

func (u *UserImpl) GenerateUserFullKey(kgcUserPartialKey *kgc.KGCUserPartialKey, userPartialKey *UserPartialKey) (*UserFullKey, error) {
	fullPrivateKey := new(big.Int).Add(kgcUserPartialKey.T, userPartialKey.PrivateKey)
	fullPrivateKey.Mod(fullPrivateKey, u.Params().N)
	if fullPrivateKey.Cmp(big.NewInt(0)) == 0 {
		return nil, fmt.Errorf("full key generation failed, please regenerate user partial key")
	}

	return &UserFullKey{
		PrivateKey: fullPrivateKey,
		PubX:       kgcUserPartialKey.Wx,
		PubY:       kgcUserPartialKey.Wy,
	}, nil
}

// CalculateFullPublicKey computes the full public key by combining a declared public key with a master public key.
// It takes a base64-encoded declared public key and user ID as input, and returns the derived public key coordinates (X,Y).
// The calculation involves hashing user-specific information and applying elliptic curve operations.
// Returns error if the input public key is invalid.
func (u *UserImpl) CalculateFullPublicKey(declaredPbkBase64, userId string) (*big.Int, *big.Int, error) {
	byteLen := u.Curve.Params().BitSize / 8
	declaredPbk, err := base64.StdEncoding.DecodeString(declaredPbkBase64)
	if err != nil {
		return nil, nil, err
	}

	declaredPbkX := new(big.Int).SetBytes(declaredPbk[:byteLen])
	declaredPbkY := new(big.Int).SetBytes(declaredPbk[byteLen:])

	info := []byte{}
	lenUserId := len(userId)
	info = append(info, byte(lenUserId>>8))
	info = append(info, byte(lenUserId))
	info = append(info, []byte(userId)...)
	info = append(info, u.Params().A.Bytes()...)
	info = append(info, u.Params().B.Bytes()...)
	info = append(info, u.Params().Gx.Bytes()...)
	info = append(info, u.Params().Gy.Bytes()...)
	info = append(info, u.masterKey.PpubX.Bytes()...)
	info = append(info, u.masterKey.PpubY.Bytes()...)

	u.h.Write(info)
	infoHash := u.h.Sum(nil)
	u.h.Reset()

	lamdaInfo := []byte{}
	lamdaInfo = append(lamdaInfo, declaredPbkX.Bytes()...)
	lamdaInfo = append(lamdaInfo, declaredPbkY.Bytes()...)
	lamdaInfo = append(lamdaInfo, infoHash...)

	u.h.Write(lamdaInfo)
	lamdaHash := u.h.Sum(nil)
	u.h.Reset()
	lamda := new(big.Int).SetBytes(lamdaHash)
	lamda.Mod(lamda, u.Params().N)

	scaleMasterPubX, scaleMasterPubY := u.Curve.ScalarMult(u.masterKey.PpubX, u.masterKey.PpubY, lamda.Bytes())

	userPubX, userPubY := u.Curve.Add(
		declaredPbkX, declaredPbkY,
		scaleMasterPubX, scaleMasterPubY,
	)

	return userPubX, userPubY, nil
}

// VerifyFullKey verifies the user's full key by checking if the derived public key matches
// the public key generated from the private key. It combines user information with system
// parameters to compute a hash, then uses this hash to scale the master public key.
// The scaled key is added to declared public key and compared against the public key
// generated from the full key's private key. Returns an error if verification fails.
func (u *UserImpl) VerifyFullKey(fullKey *UserFullKey, userId string) error {
	declaredUserPubBytes := fullKey.PubX.Bytes()
	declaredUserPubBytes = append(declaredUserPubBytes, fullKey.PubY.Bytes()...)

	declaredPbkBase64 := base64.StdEncoding.EncodeToString(declaredUserPubBytes)

	userPubX, userPubY, err := u.CalculateFullPublicKey(declaredPbkBase64, userId)
	if err != nil {
		return err
	}

	userPubXFromPrk, userPubYFromPrk := u.Curve.ScalarBaseMult(fullKey.PrivateKey.Bytes())

	if userPubX.Cmp(userPubXFromPrk) != 0 || userPubY.Cmp(userPubYFromPrk) != 0 {
		return fmt.Errorf("full key verification failed")
	}

	return nil
}

// Sign generates an ECDSA signature for the given message using the user's private key.
// It returns the signature components (r, s) and any error encountered during signing.
// The signature is computed using the standard ECDSA algorithm:
// 1. Hashes the message
// 2. Generates a random nonce k
// 3. Computes r = (k*G).x mod N
// 4. Computes s = k⁻¹·(e + r·dA) mod N
// where e is the message hash, dA is the private key, and N is the curve order.
func (u *UserImpl) Sign(prkBase64 string, message string) (r, s *big.Int, err error) {
	u.h.Write([]byte(message))
	msgHash := u.h.Sum(nil)
	u.h.Reset()

	k, err := kgc.GenerateRandomNumber(u.Params().N)
	if err != nil {
		return nil, nil, err
	}

	// k*G
	kGx, _ := u.Curve.ScalarBaseMult(k.Bytes())

	// r = kGx mod N
	r = new(big.Int).Mod(kGx, u.Params().N)
	if r.Sign() == 0 {
		return nil, nil, fmt.Errorf("invalid r value")
	}

	// k⁻¹
	kInv := new(big.Int).ModInverse(k, u.Params().N)

	prk, err := base64.StdEncoding.DecodeString(prkBase64)
	if err != nil {
		return nil, nil, err
	}

	// s = k⁻¹·(e + r·dA) mod n
	rda := new(big.Int).Mul(r, new(big.Int).SetBytes(prk))
	ePlusRda := new(big.Int).Add(rda, new(big.Int).SetBytes(msgHash))
	s = new(big.Int).Mod(new(big.Int).Mul(ePlusRda, kInv), u.Params().N)

	if s.Sign() == 0 {
		return nil, nil, fmt.Errorf("invalid s value")
	}

	return r, s, nil
}

// Verify checks the validity of an ECDSA signature (r, s) for the given message using the user's public key.
// It returns true if the signature is valid, false otherwise, and any error encountered during verification.
// The verification is performed using the standard ECDSA algorithm:
// 1. Hashes the message to get e (same hash function as used in signing)
// 2. Checks that r and s are in the range [1, N-1] where N is the curve order
// 3. Computes w = s⁻¹ mod N
// 4. Computes u1 = e·w mod N and u2 = r·w mod N
// 5. Computes (x1, y1) = u1*G + u2*Q where Q is the public key point
// 6. Verifies that r ≡ x1 mod N
// If any step fails, the signature is considered invalid.
func (u *UserImpl) Verify(declaredPbkBase64, userId, message, sigBase64 string) bool {
	n := u.Params().N

	sig, err := base64.StdEncoding.DecodeString(sigBase64)
	if err != nil {
		return false
	}

	r := new(big.Int).SetBytes(sig[:len(sig)/2])
	s := new(big.Int).SetBytes(sig[len(sig)/2:])

	if r.Sign() <= 0 || r.Cmp(n) >= 0 {
		return false
	}
	if s.Sign() <= 0 || s.Cmp(n) >= 0 {
		return false
	}

	u.h.Write([]byte(message))
	msgHash := u.h.Sum(nil)
	u.h.Reset()

	// w = s⁻¹ mod n
	w := new(big.Int).ModInverse(s, n)

	// u1 = e·w mod n
	u1 := new(big.Int).Mul(new(big.Int).SetBytes(msgHash), w)
	u1.Mod(u1, n)

	// u2 = r·w mod n
	u2 := new(big.Int).Mul(r, w)
	u2.Mod(u2, n)

	userPubX, userPubY, err := u.CalculateFullPublicKey(declaredPbkBase64, userId)
	if err != nil {
		return false
	}

	// (x, y) = u1·G + u2·P
	x1, y1 := u.Curve.ScalarBaseMult(u1.Bytes())
	x2, y2 := u.Curve.ScalarMult(userPubX, userPubY, u2.Bytes())
	x, _ := u.Curve.Add(x1, y1, x2, y2)

	// r' = x1 mod n
	rPrime := new(big.Int).Mod(x, n)

	return r.Cmp(rPrime) == 0
}
