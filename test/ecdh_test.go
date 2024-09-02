package test

import (
	"encoding/base64"
	"fmt"
	"testing"

	"github.com/OpenNHP/opennhp/nhp"
)

func TestCurve25519Keys(t *testing.T) {
	e := nhp.NewECDH(nhp.ECC_CURVE25519)

	fmt.Printf("Private key: %s\n", e.PrivateKeyBase64())
	fmt.Printf("Public key: %s\n", e.PublicKeyBase64())
}

func TestSM2Keys(t *testing.T) {
	e := nhp.NewECDH(nhp.ECC_SM2)

	fmt.Printf("Private key: %s\n", e.PrivateKeyBase64())
	fmt.Printf("Public key: %s\n", e.PublicKeyBase64())
}

func TestPublicKeys(t *testing.T) {
	//prk, err := base64.StdEncoding.DecodeString("kgvvQaBGfHNWCbZMkFWS1K07BgRXlnOo7CHTZF1bsmI=") // server
	//prk, err := base64.StdEncoding.DecodeString("2kRXjwV9zAUMc0Vf0jl984q2p9EiyjbAMUPKNu517z4=") // agent
	prk, err := base64.StdEncoding.DecodeString("D2bieOaJarsM9euBBfSs/Ky8g/X6lBQ73NmP55CMgds=") // door
	if err != nil {
		fmt.Printf("Private key decode error\n")
		return
	}
	curvee := nhp.ECDHFromKey(nhp.ECC_CURVE25519, prk)
	if curvee == nil {
		fmt.Printf("Wrong private key\n")
		return
	}
	sm2e := nhp.ECDHFromKey(nhp.ECC_SM2, prk)
	if sm2e == nil {
		fmt.Printf("Wrong private key\n")
		return
	}

	fmt.Printf("Curve25519 public key: %s\n", curvee.PublicKeyBase64())
	fmt.Printf("SM2 public key: %s\n", sm2e.PublicKeyBase64())
}

func TestPeer(t *testing.T) {
	server := &nhp.UdpPeer{
		Ip:           "192.168.2.27",
		Port:         62206,
		PubKeyBase64: "c0HALYy3433SqJmfN0JpRk1Q6H7xh84MAg89jYtRrQM=",
		ExpireTime:   1716345064,
		Type:         nhp.NHP_SERVER,
	}

	var p *nhp.UdpPeer = (*nhp.UdpPeer)(server)

	var peer nhp.Peer = p

	fmt.Printf("Pub key %s, addr %s, name %s\n", peer.PublicKeyBase64(), peer.SendAddr().String(), peer.Name())
}
