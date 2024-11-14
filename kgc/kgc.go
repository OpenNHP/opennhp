package kgc

import (
	_ "bytes"
	_ "crypto/ecdsa"
	_ "crypto/elliptic"
	"crypto/rand"
	_ "crypto/sha256"
	_"encoding/base64"
	"encoding/binary"
	_ "encoding/binary"
	"fmt"
	"math/big"

	"github.com/emmansun/gmsm/sm2"
	"github.com/emmansun/gmsm/sm3"
)

// GenerateUserKeyPairSM2,Generate user private key dA_ and public key UA based on SM2
func GenerateUserKeyPairSM2() (*big.Int, *big.Int, *big.Int, error) {
	// Using the SM2 Curve
	curve := sm2.P256()
	// Randomly generate user private key dA_, the range is [1, n-1]
	dA_, err := rand.Int(rand.Reader, curve.Params().N)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("Failed to generate user private key dA_: %v", err)
	}

	// Make sure dA_ is not 0
	if dA_.Cmp(big.NewInt(0)) == 0 {
		dA_, err = rand.Int(rand.Reader, curve.Params().N)
		if err != nil {
			return nil, nil, nil, fmt.Errorf("Failed to regenerate user private key: %v", err)
		}
	}
	// Use the curve base point G to calculate the user public key UA = [dA_]G
	UAx, UAy := curve.ScalarBaseMult(dA_.Bytes())
	return dA_, UAx, UAy, nil
}

// GenerateMasterKeyPairSM2,Generate the system's master private key ms and master public key Ppub
func GenerateMasterKeyPairSM2() (*big.Int, *big.Int, *big.Int, error) {
	// Using the SM2 Curve
	curve := sm2.P256()

	// Generate the system master private key ms, the range is [1, n-1]
	ms, err := rand.Int(rand.Reader, curve.Params().N)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("Failed to generate system master private key ms: %v", err)
	}

	// Make sure ms is not 0
	if ms.Cmp(big.NewInt(0)) == 0 {
		ms, err = rand.Int(rand.Reader, curve.Params().N)
		if err != nil {
			return nil, nil, nil, fmt.Errorf("Regeneration of system master private key ms failed: %v", err)
		}
	}
	// Use the curve base point G to calculate the system master public key Ppub = [ms]G
	PpubX, PpubY := curve.ScalarBaseMult(ms.Bytes())
	return ms, PpubX, PpubY, nil
}

// GenerateWA,Calculate WA = [w]G + UA
func GenerateWA(UAx, UAy *big.Int) (*big.Int, *big.Int, *big.Int, error) {
	// Using the SM2 curve
	curve := sm2.P256()

	// Generate a random number w in the range [1, n-1]
	w, err := rand.Int(rand.Reader, curve.Params().N)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("Failed to generate random number w: %v", err)
	}

	// Make sure w is not 0
	if w.Cmp(big.NewInt(0)) == 0 {
		w, err = rand.Int(rand.Reader, curve.Params().N)
		if err != nil {
			return nil, nil, nil, fmt.Errorf("Failed to regenerate random number w: %v", err)
		}
	}
	Wx, Wy := curve.ScalarBaseMult(w.Bytes())
	//fmt.Printf("wx=%x, wy=%x\n", Wx, Wy)
	WAx, WAy := curve.Add(Wx, Wy, UAx, UAy)
	return WAx, WAy, w, nil
}

func CalculateHA(entlA int, idA []byte, a, b, xG, yG, xPub, yPub *big.Int) []byte {
	// Convert entlA to a 2-byte bit string
	entlABytes := make([]byte, 2)
	binary.BigEndian.PutUint16(entlABytes, uint16(entlA))
	// Concatenate all parameters
	data := append(entlABytes, idA...) 
	data = append(data, a.Bytes()...) 
	data = append(data, b.Bytes()...) 
	data = append(data, xG.Bytes()...) 
	data = append(data, yG.Bytes()...) 
	data = append(data, xPub.Bytes()...) 
	data = append(data, yPub.Bytes()...) 
	// Calculating SM3 hash value
	hash := sm3.New()
	hash.Write(data)
	HA := hash.Sum(nil)
	return HA
}

// ComputeL  l = H256(xWA‖yWA‖HA) mod n
func ComputeL(xWA, yWA *big.Int, HA []byte, n *big.Int) (*big.Int, error) {
	//  Convert the coordinates of WA to a bit string
	xBits := intToBitString(xWA)
	yBits := intToBitString(yWA)
	//  Concatenating bit strings and HA
	hashData := append(xBits, yBits...)
	hashData = append(hashData, HA...)
	//  Calculating hashes using SM3
	hash := sm3.Sum(hashData)
	// Convert the hash value to a big.Int
	l := new(big.Int).SetBytes(hash[:]) 
	l.Mod(l, n)
	if l.Cmp(big.NewInt(0)) < 0 {
		return nil, fmt.Errorf("The calculated result l is a negative number")
	}

	//  Convert l to a byte string
	k := (n.BitLen() + 7) / 8 
	lBytes := intToBytes(l, k)

	// Convert l to an integer, ensuring compliance with the standard
	lInteger := new(big.Int).SetBytes(lBytes)

	return lInteger, nil
}

// intToBitString 
func intToBitString(x *big.Int) []byte {

	bitLen := x.BitLen()
	
	byteLen := (bitLen + 7) / 8
	
	bitString := make([]byte, byteLen)
	
	xBytes := x.Bytes()
	
	copy(bitString[byteLen-len(xBytes):], xBytes)

	return bitString
}

// intToBytes 
func intToBytes(x *big.Int, k int) []byte {
	m := make([]byte, k)
	xBytes := x.Bytes()
	copy(m[k-len(xBytes):], xBytes)
	return m
}

// 计算 tA
func ComputeTA(w, lInteger, ms, n *big.Int) *big.Int {
	tA := new(big.Int).Set(w) 
	lMod := new(big.Int).Mod(lInteger, n) 
	msMod := new(big.Int).Mod(ms, n)      
	lMulMs := new(big.Int).Mul(lMod, msMod) 
	lMulMs.Mod(lMulMs, n) 
	tA.Add(tA, lMulMs) // tA = w + (l * ms)
	tA.Mod(tA, n)      // tA = (w + (l * ms)) mod n
	return tA
}

// 计算 dA
func ComputeDA(tA, dA_ *big.Int, n *big.Int) *big.Int {
	dA := new(big.Int).Set(tA)      
	dA.Add(dA, dA_)                
	dA.Mod(dA, n)                 
	return dA
}

// Calculate PA = WA + [l]Ppub
func ComputePA(WAx, WAy, PpubX, PpubY *big.Int, lInteger *big.Int) (*big.Int, *big.Int) {
	curve := sm2.P256()
	PpubXl, PpubYl := curve.ScalarMult(PpubX, PpubY, lInteger.Bytes())
	//fmt.Printf("PpubXl=%x, PpubYl=%x\n", PpubXl, PpubYl)
	PAx, PAy := curve.Add(WAx, WAy, PpubXl, PpubYl)
	return PAx, PAy
}

// Calculate P'A = [dA]G
func ComputePAPrime(dA *big.Int) (*big.Int, *big.Int) {
	curve := sm2.P256()
	PAX_, PAY_ := curve.ScalarBaseMult(dA.Bytes())
	return PAX_, PAY_
}




