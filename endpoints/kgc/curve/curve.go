// This package provides the interface and custom implementation of elliptic curve operations.
// It defines interfaces and structures to perform standard elliptic curve cryptographic operations
// such as point addition, scalar multiplication, and checking whether a point lies on the curve.
// we can create own custom curve or wrap official curve to follow the interface of this package.

package curve

import (
	"crypto/elliptic"
	"math/big"

	"github.com/emmansun/gmsm/sm2"
)

type CurveParams struct {
	P       *big.Int // the order of the underlying field
	N       *big.Int // the order of the base point
	A       *big.Int // the constant of the curve equation
	B       *big.Int // the constant of the curve equation
	Gx, Gy  *big.Int // (x,y) of the base point
	BitSize int      // the size of the underlying field
	Name    string   // the canonical name of the curve
}

type Curve interface {
	Params() *CurveParams
	IsOnCurve(x, y *big.Int) bool
	Add(x1, y1, x2, y2 *big.Int) (x, y *big.Int)
	ScalarMult(x1, y1 *big.Int, k []byte) (x, y *big.Int)
	ScalarBaseMult(k []byte) (x, y *big.Int)
}

type CustomStandardCurve struct {
	*CurveParams
}

func (c *CustomStandardCurve) Params() *CurveParams {
	return c.CurveParams
}

func (c *CustomStandardCurve) IsOnCurve(x, y *big.Int) bool {
	if x.Sign() == 0 && y.Sign() == 0 {
		return true // infinite point
	}

	// calculate y² mod p
	ySquare := new(big.Int).Exp(y, big.NewInt(2), c.P)

	// calculate x³ + ax + b mod p
	x3 := new(big.Int).Exp(x, big.NewInt(3), c.P)
	ax := new(big.Int).Mul(c.A, x)
	ax.Mod(ax, c.P)

	rhs := new(big.Int).Add(x3, ax)
	rhs.Add(rhs, c.B)
	rhs.Mod(rhs, c.P)

	return ySquare.Cmp(rhs) == 0
}

func (c *CustomStandardCurve) Add(x1, y1, x2, y2 *big.Int) (x, y *big.Int) {
	if x1.Sign() == 0 && y1.Sign() == 0 {
		return x2, y2
	}

	if x2.Sign() == 0 && y2.Sign() == 0 {
		return x1, y1
	}

	// handle inverse point (P + (-P) = infinity point)
	if x1.Cmp(x2) == 0 && y1.Cmp(new(big.Int).Sub(c.P, y2)) == 0 {
		return new(big.Int), new(big.Int)
	}

	var lambda *big.Int

	// point doubling (P == Q)
	if x1.Cmp(x2) == 0 && y1.Cmp(y2) == 0 {
		// λ = (3x² + a) / (2y) mod p
		num := new(big.Int).Mul(big.NewInt(3), new(big.Int).Exp(x1, big.NewInt(2), nil))
		num.Add(num, c.A)
		num.Mod(num, c.P)

		den := new(big.Int).Mul(big.NewInt(2), y1)
		den.Mod(den, c.P)

		// calculate modular inverse
		denInv := new(big.Int).ModInverse(den, c.P)
		lambda = new(big.Int).Mul(num, denInv)
		lambda.Mod(lambda, c.P)
	} else {
		// normal addition (P ≠ Q)
		// λ = (y₂ - y₁) / (x₂ - x₁) mod p
		num := new(big.Int).Sub(y2, y1)
		num.Mod(num, c.P)

		den := new(big.Int).Sub(x2, x1)
		den.Mod(den, c.P)

		denInv := new(big.Int).ModInverse(den, c.P)
		lambda = new(big.Int).Mul(num, denInv)
		lambda.Mod(lambda, c.P)
	}

	// calculate x₃ = λ² - x₁ - x₂ mod p
	x3 := new(big.Int).Exp(lambda, big.NewInt(2), nil)
	x3.Sub(x3, x1)
	x3.Sub(x3, x2)
	x3.Mod(x3, c.P)

	// calculate y₃ = λ(x₁ - x₃) - y₁ mod p
	y3 := new(big.Int).Sub(x1, x3)
	y3.Mul(y3, lambda)
	y3.Sub(y3, y1)
	y3.Mod(y3, c.P)


	return x3, y3
}

func (c *CustomStandardCurve) ScalarMult(x1, y1 *big.Int, k []byte) (x, y *big.Int) {
	x = new(big.Int)
	y = new(big.Int)

	kCopy := new(big.Int).SetBytes(k)

	// binary expansion (Montgomery ladder algorithm)
	for kCopy.Sign() > 0 {
		if kCopy.Bit(0) == 1 {
			x, y = c.Add(x, y, x1, y1)
		}

		x1, y1 = c.Add(x1, y1, x1, y1) // point doubling
		kCopy.Rsh(kCopy, 1)  // right shift by one bit
	}

	return x, y
}

func (c *CustomStandardCurve) ScalarBaseMult(k []byte) (x, y *big.Int) {
	return c.ScalarMult(c.Gx, c.Gy, k)
}

func NewCustomSM2Curve() *CustomStandardCurve {
	p, _ := new(big.Int).SetString("FFFFFFFEFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFF00000000FFFFFFFFFFFFFFFF", 16)
	a, _ := new(big.Int).SetString("FFFFFFFEFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFF00000000FFFFFFFFFFFFFFFC", 16)
	b, _ := new(big.Int).SetString("28E9FA9E9D9F5E344D5A9E4BCF6509A7F39789F515AB8F92DDBCBD414D940E93", 16)
	n, _ := new(big.Int).SetString("FFFFFFFEFFFFFFFFFFFFFFFFFFFFFFFF7203DF6B21C6052B53BBF40939D54123", 16)

	gx, _ := new(big.Int).SetString("32C4AE2C1F1981195F9904466A39C9948FE30BBFF2660BE1715A4589334C74C7", 16)
	gy, _ := new(big.Int).SetString("BC3736A2F4F6779C59BDCEE36B692153D0A9877CC62A474002DF32E52139F0A0", 16)

	return &CustomStandardCurve{
		&CurveParams{
			P:       p,
			N:       n,
			A:       a,
			B:       b,
			Gx:      gx,
			Gy:      gy,
			BitSize: 256,
			Name:    "Custom SM2",
		},
	}
}

func NewCustomSecp256k1Curve() *CustomStandardCurve {
	p, _ := new(big.Int).SetString("FFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFEFFFFFC2F", 16)
	a, _ := new(big.Int).SetString("000000000000000000000000000000000000000000000000000000000000", 16)
	b, _ := new(big.Int).SetString("000000000000000000000000000000000000000000000000000000000007", 16)
	n, _ := new(big.Int).SetString("FFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFEBAAEDCE6AF48A03BBFD25E8CD0364141", 16)

	gx, _ := new(big.Int).SetString("79BE667EF9DCBBAC55A06295CE870B07029BFCDB2DCE28D959F2815B16F81798", 16)
	gy, _ := new(big.Int).SetString("483ADA7726A3C4655DA4FBFC0E1108A8FD17B448A68554199C47D08FFB10D4B8", 16)

	return &CustomStandardCurve{
		&CurveParams{
			P:       p,
			N:       n,
			A:       a,
			B:       b,
			Gx:      gx,
			Gy:      gy,
			BitSize: 256,
			Name:    "Custom Secp256k1",
		},
	}
}

type OfficialSM2Curve struct {
	curve elliptic.Curve
}

func (c *OfficialSM2Curve) Params() *CurveParams {
	A := new(big.Int).Mod(big.NewInt(-3), c.curve.Params().P)

	return &CurveParams{
		P:      c.curve.Params().P,
		N:      c.curve.Params().N,
		A:      A,
		B:      c.curve.Params().B,
		Gx:     c.curve.Params().Gx,
		Gy:     c.curve.Params().Gy,
		BitSize: c.curve.Params().BitSize,
		Name:    c.curve.Params().Name,
	}
}

func (c *OfficialSM2Curve) IsOnCurve(x, y *big.Int) bool {
	return c.curve.IsOnCurve(x, y)
}

func (c *OfficialSM2Curve) Add(x1, y1, x2, y2 *big.Int) (x, y *big.Int) {
	return c.curve.Add(x1, y1, x2, y2)
}

func (c *OfficialSM2Curve) ScalarMult(x1, y1 *big.Int, k []byte) (x, y *big.Int) {
	return c.curve.ScalarMult(x1, y1, k)
}

func (c *OfficialSM2Curve) ScalarBaseMult(k []byte) (x, y *big.Int) {
	return c.curve.ScalarBaseMult(k)
}

func NewOfficialSM2Curve() *OfficialSM2Curve {
	return &OfficialSM2Curve{
		curve: sm2.P256(),
	}
}
