package kgc

import (
	"crypto/rand"
	"encoding/binary"
	"fmt"
	"math/big"
	"github.com/emmansun/gmsm/sm2"
	"github.com/emmansun/gmsm/sm3"
	"github.com/BurntSushi/toml"
	"os"
	"path/filepath"
	"github.com/OpenNHP/opennhp/kgc/user"
)

var(
	id = user.ID
	curve = sm2.P256()
	N = curve.Params().N
	Gx = curve.Params().Gx
	Gy = curve.Params().Gy
	IdA = []byte(id)
	EntlA = len(IdA) * 8
	Ms *big.Int 
	PpubX *big.Int 
	PpubY *big.Int 
	curveParams CurveParams
	A, B *big.Int
	WAx, WAy, W *big.Int
	HA []byte
	L *big.Int
	TA *big.Int
)

// A structure for storing configuration
type CurveParams struct {
    A string `toml:"a"`
    B string `toml:"b"`
}

//InitConfig loads the configuration and initializes global variables
func InitConfig() error {
    // Get the current working directory path
    wd, err := os.Getwd()
    if err != nil {
        return fmt.Errorf("error getting current directory: %v", err)
    }

   // Path to splice TOML files
    tomlFilePath := filepath.Join(wd, "kgc", "main", "etc", "Curve.toml")

   // Read and parse TOML files
    _, err = toml.DecodeFile(tomlFilePath, &curveParams)
    if err != nil {
        return fmt.Errorf("error loading TOML file: %v", err)
    }

    // Convert a and b from strings in TOML file to big.Int type
    A = new(big.Int)
    A.SetString(curveParams.A, 16)
    B = new(big.Int)
    B.SetString(curveParams.B, 16)
    return nil
}

func GetA() *big.Int {
    return A
}

func GetB() *big.Int {
    return B
}

// GenerateMasterKeyPairSM2,Generate the system's master private key ms and master public key Ppub
func GenerateMasterKeyPairSM2() (*big.Int, *big.Int, error) {
	curve := sm2.P256()
	ms, err := rand.Int(rand.Reader, curve.Params().N)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to generate system master private key ms: %v", err)
	}
	if ms.Cmp(big.NewInt(0)) == 0 {
		ms, err = rand.Int(rand.Reader, curve.Params().N)
		if err != nil {
			return nil, nil, fmt.Errorf("regeneration of system master private key ms failed: %v", err)
		}
	}
	Ppubx, Ppuby := curve.ScalarBaseMult(ms.Bytes())
	Ms = ms
	PpubX = Ppubx
	PpubY = Ppuby
	return Ppubx, Ppuby, nil
}

// GenerateWA,Calculate WA = [w]G + UA
func GenerateWA(UAx, UAy *big.Int) (*big.Int, *big.Int, *big.Int, error) {
	curve := sm2.P256()
	// Generate a random number w in the range [1, n-1]
	w, err := rand.Int(rand.Reader, curve.Params().N)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("failed to generate random number w: %v", err)
	}

	// Make sure w is not 0
	if w.Cmp(big.NewInt(0)) == 0 {
		w, err = rand.Int(rand.Reader, curve.Params().N)
		if err != nil {
			return nil, nil, nil, fmt.Errorf("failed to regenerate random number w: %v", err)
		}
	}
	Wx, Wy := curve.ScalarBaseMult(w.Bytes())
	wAx, wAy := curve.Add(Wx, Wy, UAx, UAy)
	WAx = wAx
	WAy = wAy
	W = w
	return WAx, WAy, w, nil
}

// Calculate HA = H256(entlA || idA || a || b || xG || yG || xPub || yPub)
func CalculateHA(entlA int, idA []byte, a, b, xG, yG, xPub, yPub *big.Int) ([]byte,error) {
	if a == nil || b == nil || xG == nil || yG == nil || xPub == nil || yPub == nil {
		return nil, fmt.Errorf("one or more big.Int parameters passed in were nil")
	}
	entlABytes := make([]byte, 2)
	binary.BigEndian.PutUint16(entlABytes, uint16(entlA))
	data := append(entlABytes, idA...)
	data = append(data, a.Bytes()...)
	data = append(data, b.Bytes()...)
	data = append(data, xG.Bytes()...)
	data = append(data, yG.Bytes()...)
	data = append(data, xPub.Bytes()...)
	data = append(data, yPub.Bytes()...)
	hash := sm3.New()
	hash.Write(data)
	ha := hash.Sum(nil)
	HA = ha
	return HA,nil
}

// ComputeL  l = H256(xWA‖yWA‖HA) mod n
func ComputeL(xWA, yWA *big.Int, HA []byte, n *big.Int) (*big.Int, error) {
	xBits := intToBitString(xWA)
	yBits := intToBitString(yWA)
	hashData := append(xBits, yBits...)
	hashData = append(hashData, HA...)
	hash := sm3.Sum(hashData)
	l := new(big.Int).SetBytes(hash[:])
	l.Mod(l, n)  
	if l.Cmp(big.NewInt(0)) < 0 {
		return nil, fmt.Errorf("the calculated result l is a negative number")
	}
	k := (n.BitLen() + 7) / 8
	lBytes := intToBytes(l, k)
	lInteger := new(big.Int).SetBytes(lBytes)
	L = lInteger
	return L, nil
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


//Calculate tA= w + (l * ms)
func ComputeTA(w, lInteger, ms, n *big.Int) *big.Int {
	tA := new(big.Int).Set(w)
	lMod := new(big.Int).Mod(lInteger, n)
	msMod := new(big.Int).Mod(ms, n)
	lMulMs := new(big.Int).Mul(lMod, msMod)
	lMulMs.Mod(lMulMs, n)
	tA.Add(tA, lMulMs)
	tA.Mod(tA, n)
	TA = tA
	return TA
}






