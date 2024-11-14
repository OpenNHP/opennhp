package main

import (
	_ "crypto/rand"
	"fmt"
	_ "log"
	"math/big"

	_"github.com/OpenNHP/opennhp/core/scheme/curve"
	"github.com/OpenNHP/opennhp/kgc"
	"github.com/emmansun/gmsm/sm2"
)

func main() {
	idA := []byte("example@163.com") // Example user ID
	entlA := len(idA)*8
	//fmt.Print("User ID: ", idA)
	//fmt.Printf("User ID length: %d\n", entlA)
	curve := sm2.P256() // Use SM2 curves
	//n := new(big.Int)
  //n.SetString("FFFFFFFEFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFF00000000FFFFFFFFFFFFFFFC", 16)
	n := curve.Params().N

	gx := new(big.Int)
	gx.SetString("32C4AE2C1F1981195F9904466A79E36EAA3C8E4B1EAC73D2C2FB7D3290ECF8A5", 16)

	gy := new(big.Int)
	gy.SetString("BC3E1A3D1F1694E1A3D1C3E34D35E4C510DC53D6A57E44D2B38C6B8C8AD7C1A1", 16)
	/* gx := curve.Params().Gx   // x coordinate of base point G
  gy := curve.Params().Gy   // y coordinate of base point G */
	a := new(big.Int)
	a.SetString("FFFFFFFEFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFF00000000FFFFFFFFFFFFFFFF", 16)

	b := new(big.Int)
	b.SetString("28E9FA9E9D9A3E004F2018D6F8C7193A3197D1A9F23A5C24B545C1E3A09F31A4", 16)
  /* a :=big.NewInt(0)      // a value of the curve
	b := curve.Params().B     // b value of the curve */


	dA_, UAx, UAy, err := kgc.GenerateUserKeyPairSM2()
if err != nil {
    fmt.Printf("User partial key pair generation failed: %v\n", err)
    return
}
fmt.Printf("User partial key pair generated successfully:\n  dA_ = %X \n UAx = %X \n UAy = %X\n", dA_, UAx, UAy) 
fmt.Printf("-------------------------------------------------------------------------\n")



	// Generate KGC master key pair
	ms, XPpub, YPpub, err := kgc.GenerateMasterKeyPairSM2()
	if err != nil {
		fmt.Printf("Failed to generate kgc master key pair: %v\n", err)
		return
	}
	fmt.Printf("KGC master key pair generated successfully:\n  ms = %X \n XPpub = %X \n YPPub = %X\n", ms, XPpub, YPpub)
	fmt.Printf("-------------------------------------------------------------------------\n")
	

	WAX, WAY, w, err := kgc.GenerateWA(UAx, UAy)
	if err != nil {
		fmt.Printf("Failed to generate WA: %v\n", err)
		return
	}
	fmt.Printf("WA calculation successful:\n  WAX = %X \n WAY = %X \n w = %X\n", WAX, WAY, w)
	fmt.Printf("-------------------------------------------------------------------------\n")
	ha := kgc.CalculateHA(entlA, idA, a,b,gx, gy, XPpub, YPpub) 
	fmt.Printf("HA calculation successful:\n ha = %x\n", ha)
	fmt.Printf("-------------------------------------------------------------------------\n")
	lInteger, err := kgc.ComputeL(WAX, WAY, ha, n) 
	if err != nil {
		fmt.Printf("Failed to calculate L: %v\n", err)
		return
	}
	fmt.Printf("Calculation success: \n L = %X\n", lInteger)
	fmt.Printf("-------------------------------------------------------------------------\n")
	TA := kgc.ComputeTA(w, lInteger, ms, n) 
	fmt.Printf("Calculate TA success: \n TA = %X\n",TA )
	fmt.Printf("-------------------------------------------------------------------------\n")
	dA := kgc.ComputeDA(TA, dA_, n) 
	fmt.Printf("User's actual private key: \n dA = %X\n", dA)
	fmt.Printf("-------------------------------------------------------------------------\n")
	PAX, PAY := kgc.ComputePA(WAX, WAY, XPpub, YPpub, lInteger)
	fmt.Printf("User's actual public key: \n PAX = %X \n PAY = %X\n", PAX, PAY)
	fmt.Printf("-------------------------------------------------------------------------\n")
	PAX_, PAY_ := kgc.ComputePAPrime(dA)
	fmt.Printf("Calculate PA' successfully:\n PAX_ = %X \n PAY_ = %X\n", PAX_, PAY_)
	fmt.Printf("-------------------------------------------------------------------------\n")
	if PAX.Cmp(PAX_) != 0 {
		fmt.Printf("Verification of PAX failed\n")
	} else {
		fmt.Printf("Verify PAX success\n")
	}
	if PAY.Cmp(PAY_) != 0 {
		fmt.Printf("Verification of PAY failed\n")
	} else {
		fmt.Printf("Verify PAY success\n")
	}
	
  
}
