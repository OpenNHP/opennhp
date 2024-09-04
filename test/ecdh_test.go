package test

import (
	"encoding/base64"
	"fmt"
	"testing"

	"github.com/OpenNHP/opennhp/core"
)

func TestCurve25519Keys(t *testing.T) {
	e := core.NewECDH(core.ECC_CURVE25519)

	fmt.Printf("Private key: %s\n", e.PrivateKeyBase64())
	fmt.Printf("Public key: %s\n", e.PublicKeyBase64())
}

func TestSM2Keys(t *testing.T) {
	e := core.NewECDH(core.ECC_SM2)

	fmt.Printf("Private key: %s\n", e.PrivateKeyBase64())
	fmt.Printf("Public key: %s\n", e.PublicKeyBase64())
}

func TestPublicKeys(t *testing.T) {
	//prk, err := base64.StdEncoding.DecodeString("kgvvQaBGfHNWCbZMkFWS1K07BgRXlnOo7CHTZF1bsmI=") // server
	//prk, err := base64.StdEncoding.DecodeString("2kRXjwV9zAUMc0Vf0jl984q2p9EiyjbAMUPKNu517z4=") // agent
	prk, err := base64.StdEncoding.DecodeString("D2bieOaJarsM9euBBfSs/Ky8g/X6lBQ73NmP55CMgds=") // ac
	if err != nil {
		fmt.Printf("Private key decode error\n")
		return
	}
	curvee := core.ECDHFromKey(core.ECC_CURVE25519, prk)
	if curvee == nil {
		fmt.Printf("Wrong private key\n")
		return
	}
	sm2e := core.ECDHFromKey(core.ECC_SM2, prk)
	if sm2e == nil {
		fmt.Printf("Wrong private key\n")
		return
	}

	fmt.Printf("Curve25519 public key: %s\n", curvee.PublicKeyBase64())
	fmt.Printf("SM2 public key: %s\n", sm2e.PublicKeyBase64())
}

func TestPeer(t *testing.T) {
	server := &core.UdpPeer{
		Ip:           "192.168.2.27",
		Port:         62206,
		PubKeyBase64: "c0HALYy3433SqJmfN0JpRk1Q6H7xh84MAg89jYtRrQM=",
		ExpireTime:   1716345064,
		Type:         core.NHP_SERVER,
	}

	var p *core.UdpPeer = (*core.UdpPeer)(server)

	var peer core.Peer = p

	fmt.Printf("Pub key %s, addr %s, name %s\n", peer.PublicKeyBase64(), peer.SendAddr().String(), peer.Name())
}
