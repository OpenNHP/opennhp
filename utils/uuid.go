package utils

import (
	"errors"
	"fmt"
	"math/rand"
	"time"
)

func NewUUID() (string, error) {
	rng := rand.New(rand.NewSource(time.Now().UnixNano()))
	uuid := make([]byte, 16)
	_, err := rng.Read(uuid)
	if err != nil {
		return "", errors.New("failed to generate UUID: " + err.Error())
	}

	// Set version bits (version 4)
	uuid[6] = (uuid[6] & 0x0F) | 0x40
	// Set variant bits (variant 1)
	uuid[8] = (uuid[8] & 0x3F) | 0x80

	return fmt.Sprintf("%x-%x-%x-%x-%x", uuid[0:4], uuid[4:6], uuid[6:8], uuid[8:10], uuid[10:]), nil
}

func RandNumber() int {
	rng := rand.New(rand.NewSource(time.Now().UnixNano()))
	randomNumber := rng.Intn(10000)
	if randomNumber < 1000 {
		randomNumber += 1000
	}

	return randomNumber
}
