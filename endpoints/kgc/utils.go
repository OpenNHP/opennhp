package kgc

import (
	"crypto/rand"
	"fmt"
	"math/big"
	"os"
	"path/filepath"
)

func GenerateRandomNumber(N *big.Int) (*big.Int, error) {
	r, err := rand.Int(rand.Reader, N)
	if err != nil {
		return nil, err
	}

	if r.Cmp(big.NewInt(0)) == 0 {
		return nil, fmt.Errorf("random number is zero")
	}

	return r, nil
}

func GetExeDirPath() (string, error) {
	exeFilePath, err := os.Executable()
	if err != nil {
		return "", err
	}

	return filepath.Dir(exeFilePath), nil
}
