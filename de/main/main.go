package main

import (
	"encoding/base64"
	"fmt"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"

	"github.com/OpenNHP/opennhp/core"
	"github.com/OpenNHP/opennhp/de"
	"github.com/OpenNHP/opennhp/version"
	"github.com/urfave/cli/v2"
)

func main() {
	initApp()
}
func initApp() {
	app := cli.NewApp()
	app.Name = "nhp-device"
	app.Usage = "device entity for NHP protocol"
	app.Version = version.Version

	runCmd := &cli.Command{
		Name:  "run",
		Usage: "create and run device process for NHP protocol",
		Flags: []cli.Flag{
			&cli.StringFlag{Name: "mode", Value: "none", Usage: "encrypt;decrypt"},
			&cli.StringFlag{Name: "source", Value: "sample.txt", Usage: "source file"},
			&cli.StringFlag{Name: "policy", Value: "policyinfo.json", Usage: "The policy file contains the public key information of the data accessor"},
			&cli.StringFlag{Name: "output", Value: "output.txt", Usage: "Save path of the ztdo file"},
			&cli.StringFlag{Name: "meta", Value: "meta.json", Usage: "meta.json"},
			&cli.StringFlag{Name: "ztdo", Value: "", Usage: "path to the ztdo file"},
			&cli.StringFlag{Name: "decodeKey", Value: "", Usage: "decrypt key"},
		},
		Action: func(c *cli.Context) error {
			mode := c.String("mode")
			source := c.String("source")
			policy := c.String("policy")
			output := c.String("output")
			ztdo := c.String("ztdo")
			decodeKey := c.String("decodeKey")
			meta := c.String("meta")
			return runApp(mode, source, output, policy, ztdo, decodeKey, meta)
		},
	}

	keygenCmd := &cli.Command{
		Name:  "keygen",
		Usage: "generate key pairs for NHP devices",
		Flags: []cli.Flag{
			&cli.BoolFlag{Name: "curve", Value: false, DisableDefaultText: true, Usage: "generate curve25519 keys"},
			&cli.BoolFlag{Name: "sm2", Value: false, DisableDefaultText: true, Usage: "generate sm2 keys"},
		},
		Action: func(c *cli.Context) error {
			var e core.Ecdh
			if c.Bool("sm2") {
				e = core.NewECDH(core.ECC_SM2)
			} else {
				e = core.NewECDH(core.ECC_CURVE25519)
			}
			pub := e.PublicKeyBase64()
			priv := e.PrivateKeyBase64()
			fmt.Println("Private key: ", priv)
			fmt.Println("Public key: ", pub)
			return nil
		},
	}

	pubkeyCmd := &cli.Command{
		Name:  "pubkey",
		Usage: "get public key from private key",
		Flags: []cli.Flag{
			&cli.BoolFlag{Name: "curve", Value: false, DisableDefaultText: true, Usage: "get curve25519 key"},
			&cli.BoolFlag{Name: "sm2", Value: false, DisableDefaultText: true, Usage: "get sm2 key"},
		},
		Action: func(c *cli.Context) error {
			privKey, err := base64.StdEncoding.DecodeString(c.Args().First())
			if err != nil {
				return err
			}
			cipherType := core.ECC_CURVE25519
			if c.Bool("sm2") {
				cipherType = core.ECC_SM2
			}
			e := core.ECDHFromKey(cipherType, privKey)
			if e == nil {
				return fmt.Errorf("invalid input key")
			}
			pub := e.PublicKeyBase64()
			fmt.Println("Public key: ", pub)
			return nil
		},
	}

	app.Commands = []*cli.Command{
		runCmd,
		keygenCmd,
		pubkeyCmd,
	}

	if err := app.Run(os.Args); err != nil {
		fmt.Fprintln(os.Stderr, err)
	}
}

/*
*
decodeKey:Data Decryption Key
decodeSavePath:Save Directory Path
*/
func runApp(mode string, source string, output string, policy string, ztdo string, decodeKey string, meta string) error {
	fmt.Println("mode=" + mode)

	exeFilePath, err := os.Executable()
	if err != nil {
		return err
	}
	exeDirPath := filepath.Dir(exeFilePath)
	a := &de.UdpDevice{}
	err = a.Start(exeDirPath, 4)
	if err != nil {
		return err
	}

	// react to terminate signals
	termCh := make(chan os.Signal, 1)
	signal.Notify(termCh, syscall.SIGTERM, os.Interrupt, syscall.SIGABRT)

	if mode == "encrypt" {
		fmt.Println("policy=" + policy)
		fmt.Println("source=" + source)
		fmt.Println("output=" + output)
		outputFilePath := output
		policyFile := policy
		dhpPolicy, err := de.ReadPolicyFile(policyFile)
		if err != nil {
			fmt.Printf("failed to read policy file:%s\n", err)
			return err
		}
		ztdoMetainfo, err := de.ReadMetaFile((meta))
		if err != nil {
			fmt.Printf("failed to read meta file:%s\n", err)
			return err
		}
		zoId, encodedKey := de.EncodeToZtoFile(source, outputFilePath, ztdoMetainfo)
		if zoId != "" {

			fmt.Printf("Encryption Key for Data Content,key:%s\n", encodedKey)
			eccKey, err := core.SM2Encrypt(dhpPolicy.ConsumerPublicKey, encodedKey)
			if err != nil {
				fmt.Printf("Data encryption failed：%s\n", err)
				return err
			}
			a.SendDHPRegister(zoId, dhpPolicy, eccKey)
		} else {
			fmt.Printf("failed to read source file")
		}
		os.Exit(0)
	} else if mode == "decrypt" {
		fmt.Println("ztdo=" + ztdo)
		fmt.Println("decodeKey=" + decodeKey)
		fmt.Println("output=" + output)
		de.DecodeZtoFile(ztdo, decodeKey, output)
		os.Exit(0)
	}
	// block until terminated
	<-termCh
	a.Stop()
	return nil
}
