package user

import (
	"crypto/rand"
	"crypto/sha256"
	"fmt"
	"math/big"
	"github.com/emmansun/gmsm/sm2"
)

var (
	ID  string    //User ID
	DA_ *big.Int  //User secret value
	UAX *big.Int  // User public value
	UAY *big.Int  // User public value
	DA  *big.Int  // User private key
	PAx, PAy *big.Int
	PAx_, PAy_ *big.Int
	K *big.Int
	R, S *big.Int
	MessageHash []byte
)

//receives the email and processes it
func ProcessUserEmail(id string)  {
	ID = id
}

// GenerateUserKeyPairSM2,Generate user private key dA_ and public key UA based on SM2
func GenerateUserKeyPairSM2() ( error) {
	// Using the SM2 Curve
	curve := sm2.P256()
	// Randomly generate user private key dA_, the range is [1, n-1]
	dA_, err := rand.Int(rand.Reader, curve.Params().N)
	if err != nil {
		return fmt.Errorf("failed to generate private key: %v", err)
	}
	if dA_.Cmp(big.NewInt(0)) == 0 {
		dA_, err = rand.Int(rand.Reader, curve.Params().N)
		if err != nil {
			return fmt.Errorf("failed to regenerate private key: %v", err)
		}
	}
	UAx, UAy := curve.ScalarBaseMult(dA_.Bytes())
	DA_ = dA_
  UAX = UAx
  UAY = UAy
  return nil
}

// Calculate dA
func ComputeDA(tA, dA_ *big.Int, n *big.Int) *big.Int {
	dA := new(big.Int).Set(tA)      
	dA.Add(dA, dA_)                
	dA.Mod(dA, n)   
	DA=dA              
	return DA
}

// Calculate PA = WA + [l]Ppub
func ComputePA(WAx, WAy, PpubX, PpubY *big.Int, lInteger *big.Int) (*big.Int, *big.Int) {
	curve := sm2.P256()
	PpubXl, PpubYl := curve.ScalarMult(PpubX, PpubY, lInteger.Bytes())
	PAX, PAY := curve.Add(WAx, WAy, PpubXl, PpubYl)
	PAx = PAX
	PAy = PAY
	return PAx, PAy
}

// Calculate P'A = [dA]G
func ComputePAPrime(dA *big.Int) (*big.Int, *big.Int) {
	curve := sm2.P256()
	PAX_, PAY_ := curve.ScalarBaseMult(dA.Bytes())
	PAx_ = PAX_
	PAy_ = PAY_
	return PAx_, PAy_
}

//sign
func SignMessageAndEncrypt(dA *big.Int) error {
	var message string
	fmt.Print("Enter the plaintext message to sign: ")
	fmt.Scanln(&message)
	messageHash := sha256.Sum256([]byte(message))
	r, s, err := SignWithPrivateKey(dA, messageHash[:])
	if err != nil {
		return fmt.Errorf("failed to sign message: %v", err)
	}
	R=r
	S=s
	MessageHash = messageHash[:]
	//fmt.Printf("messageHash: %X\n",messageHash)
	return nil
}

// Signature logic: generate signature based on private key and message hash
func SignWithPrivateKey(dA *big.Int, messageHash []byte) (*big.Int, *big.Int, error) {
	curve := sm2.P256() 
	k, err := GenerateRandomScalar()
	if err != nil {
		return nil, nil, fmt.Errorf("failed to generate random scalar: %v", err)
	}

	// r = (k * G).x mod n
	rX, _ := curve.ScalarBaseMult(k.Bytes())
	r := new(big.Int).Mod(rX, curve.Params().N)
	if r.Sign() == 0 {
		return nil, nil, fmt.Errorf("invalid r value, retrying")
	}

	// s = (k^-1 * (hash + r * dA)) mod n
	n := curve.Params().N
	kInv := new(big.Int).ModInverse(k, n)
	hash := new(big.Int).SetBytes(messageHash)
	rdA := new(big.Int).Mul(r, dA)
	s := new(big.Int).Mul(kInv, new(big.Int).Add(hash, rdA))
	s.Mod(s, n)
	if s.Sign() == 0 {
		return nil, nil, fmt.Errorf("invalid s value, retrying")
	}
	R=r
	S=s
	return R, S, nil
}

// Generate a random scalar k for signature calculation
func GenerateRandomScalar() (*big.Int, error) {
	curve := sm2.P256()
	n := curve.Params().N
	k, err := rand.Int(rand.Reader, n)
	if err != nil {
		return nil, fmt.Errorf("failed to generate random scalar: %v", err)
	}
	if k.Sign() == 0 {
		k, err = rand.Int(rand.Reader, n)
		if err != nil {
			return nil, fmt.Errorf("failed to regenerate random scalar: %v", err)
		}
	}
	K=k
	return K, nil
}

//Verify signature
func VerifySignature(R, S, PAx, PAy *big.Int, messageHash [32]byte) bool {
	curve := sm2.P256()
	
	// Check R and S validity
	n := curve.Params().N
	if R.Sign() <= 0 || R.Cmp(n) >= 0 {
		fmt.Println("Invalid R: out of range.")
		return false
	}
	if S.Sign() <= 0 || S.Cmp(n) >= 0 {
		fmt.Println("Invalid S: out of range.")
		return false
	}

	// Verify the public key is on the curve
	if !curve.IsOnCurve(PAx, PAy) {
		fmt.Println("Invalid public key: not on curve.")
		return false
	}

	// Compute w = S^-1 mod n
	w := new(big.Int).ModInverse(S, n)
	if w == nil {
		fmt.Println("Failed to compute modular inverse of S.")
		return false
	}

	// Compute u1 = (messageHash * w) mod n and u2 = (R * w) mod n
	u1 := new(big.Int).Mul(new(big.Int).SetBytes(messageHash[:]), w)
	u1.Mod(u1, n)
	u2 := new(big.Int).Mul(R, w)
	u2.Mod(u2, n)

	// Compute (x1, y1) = u1*G + u2*(PAx, PAy)
	x1, y1 := curve.ScalarBaseMult(u1.Bytes())       // u1*G
	x2, y2 := curve.ScalarMult(PAx, PAy, u2.Bytes()) // u2*PA
	x, y := curve.Add(x1, y1, x2, y2)               // u1*G + u2*PA

	// Verify if R == x mod n
	if x == nil || y == nil {
		fmt.Println("Invalid point during verification.")
		return false
	}
	v := new(big.Int).Mod(x, n)
	if v.Cmp(R) == 0 {
		fmt.Println("Signature verification succeeded!")
		return true
	} else {
		fmt.Println("Signature verification failed!")
		return false
	}
}


