package main

import (
	"encoding/base64"
	"fmt"
	"os"
	"os/signal"
	"path/filepath"
	"slices"
	"syscall"

	"github.com/OpenNHP/opennhp/endpoints/db"
	"github.com/OpenNHP/opennhp/nhp/common"
	"github.com/OpenNHP/opennhp/nhp/core"
	ztdolib "github.com/OpenNHP/opennhp/nhp/core/ztdo"
	"github.com/OpenNHP/opennhp/nhp/log"
	"github.com/OpenNHP/opennhp/nhp/version"
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
			&cli.StringFlag{Name: "source", Value: "", Usage: "source file to be encrypted, this is not required for streaming mode"},
			&cli.StringFlag{Name: "dataSourceType", Value: "", Usage: "type of data source, the default value is online, supported values are online, offline and stream"},
			&cli.StringFlag{Name: "smartPolicy", Value: "", Usage: "The wasm policy file"},
			&cli.StringFlag{Name: "output", Value: "", Usage: "Save path of the ztdo file or decrypted file"},
			&cli.StringFlag{Name: "accessUrl", Value: "", Usage: "ZTDO access url for online or offline mode or API url for streaming mode"},
			&cli.StringFlag{Name: "ztdo", Value: "", Usage: "path to the ztdo file"},
			&cli.StringFlag{Name: "dataPrivateKey", Value: "", Usage: "data private key with base64 format"},
			&cli.StringFlag{Name: "providerPublicKey", Value: "", Usage: "provider public key with base64 format"},
		},
		Before: func(c *cli.Context) error {
			if c.String("mode") == "encrypt" {
				if c.String("dataSourceType") == "" {
					return fmt.Errorf("--dataSourceType is required in encrypt mode")
				} else {
					if !slices.Contains([]string{"online", "offline", "stream"}, c.String("dataSourceType")) {
						return fmt.Errorf("invalid --dataSourceType, allowed values are online, offline and stream")
					}

					if c.String("dataSourceType") != "stream" {
						if c.String("source") == "" {
							return fmt.Errorf("--source is required when --dataSourceType is not stream")
						}
					} else {
						if c.String("accessUrl") == "" {
							return fmt.Errorf("--accessUrl is required when --dataSourceType is stream")
						}
					}

					if c.String("smartPolicy") == "" {
						return fmt.Errorf("--smartPolicy is required in encrypt mode")
					}
				}

				if c.String("ztdo") != "" || c.String("dataPrivateKey") != "" || c.String("providerPublicKey") != "" {
					return fmt.Errorf("--ztdo, --dataPrivateKey and --providerPublicKey are only allowed in decrypt mode")
				}
			} else if c.String("mode") == "decrypt" {
				if c.String("source") != "" || c.String("smartPolicy") != "" || c.String("accessUrl") != "" {
					return fmt.Errorf("--source, --smartPolicy and --accessUrl are only allowed in encrypt mode")
				}

				if c.String("ztdo") == "" || c.String("output") == "" || c.String("dataPrivateKey") == "" || c.String("providerPublicKey") == ""{
					return fmt.Errorf("--ztdo, --output, --dataPrivateKey and --providerPublicKey are required in decrypt mode")
				}
			} else {
				return nil
			}

			return nil
		},
		Action: func(c *cli.Context) error {
			mode := c.String("mode")
			source := c.String("source")
			dsType := c.String("dataSourceType")
			smartPolicy := c.String("smartPolicy")
			output := c.String("output")
			ztdo := c.String("ztdo")
			dataPrivateKey := c.String("dataPrivateKey")
			accessUrl := c.String("accessUrl")
			providerPublicKeyBase64 := c.String("providerPublicKey")
			return runApp(mode, source, dsType, output, smartPolicy, ztdo, dataPrivateKey, accessUrl, providerPublicKeyBase64)
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
			eccType := core.ECC_SM2
			if c.Bool("curve") {
				eccType = core.ECC_CURVE25519
			}
			e = core.NewECDH(eccType)
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
			cipherType := core.ECC_SM2
			if c.Bool("curve") {
				cipherType = core.ECC_CURVE25519
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


func runApp(mode string, source string, dsType string, output string, smartPolicy string, ztdoFilePath string, dataPrivateKeyBase64 string, accessUrl string, providerPublicKeyBase64 string) error {
	exeFilePath, err := os.Executable()
	if err != nil {
		return err
	}
	exeDirPath := filepath.Dir(exeFilePath)
	a := &db.UdpDevice{}
	if mode == "none" {
		a.EnableOnlineReport = true
	}
	err = a.Start(exeDirPath, 4)
	if err != nil {
		return err
	}

	if mode == "none" {
		termCh := make(chan os.Signal, 1)
		signal.Notify(termCh, syscall.SIGTERM, os.Interrupt, syscall.SIGABRT)

		// block until terminated
		<-termCh
		a.Stop()

		return nil
	}

	ztdo := ztdolib.NewZtdo()
	dataMsgPattern := [][]ztdolib.MessagePattern{
		{ztdolib.MessagePatternS, ztdolib.MessagePatternDHSS},
		{ztdolib.MessagePatternRS, ztdolib.MessagePatternDHSS},
	}

	dataKeyPairEccMode := ztdolib.CURVE25519
	if a.GetCipherSchema() == 0 {
		dataKeyPairEccMode = ztdolib.SM2
	}

	if mode == "encrypt" {
		outputFilePath := output
		policyFile := smartPolicy
		smartPolicy, err := db.ReadPolicyFile(policyFile)
		if err != nil {
			log.Error("failed to read policy file:%s\n", err)
			return err
		}

		if !(dsType == "stream") {
			if dsType == "online" {
				ztdo.SetNhpServer(a.GetServerPeer().SendAddr().String())
			} else { // offline
				ztdo.SetNhpServer("")
			}

			dataPrkStore := db.NewDataPrivateKeyStore(a.GetOwnEcdh().PublicKeyBase64())
			dataPrk := dataPrkStore.Generate(dataKeyPairEccMode)

			dataPbk := core.ECDHFromKey(dataKeyPairEccMode.ToEccType(), dataPrk).PublicKey()
			sa := ztdolib.NewSymmetricAgreement(dataKeyPairEccMode, true)
			sa.SetMessagePatterns(dataMsgPattern)

			sa.SetStaticKeyPair(a.GetOwnEcdh())
			sa.SetRemoteStaticPublicKey(dataPbk)

			gcmKey, ad :=sa.AgreeSymmetricKey()

			symmetricCipherMode, err := ztdolib.NewSymmetricCipherMode(a.GetSymmetricCipherMode())
			if err != nil {
				log.Error("failed to create symmetric cipher mode:%s\n", err)
				return err
			}
			ztdo.SetCipherConfig(true, symmetricCipherMode, dataKeyPairEccMode)
			zoId := ztdo.GetObjectID()

			log.Info("Encrypt ztdo file(file name: %s and ztdo id: %s) with cipher settings: ECC mode(%s) and Symmetric Cipher Mode(%s)\n", source, zoId, dataKeyPairEccMode, symmetricCipherMode)

			if err := ztdo.EncryptZtdoFile(source, outputFilePath, gcmKey[:], ad); err != nil {
				log.Error("failed to encrypt ztdo file: %s\n", err)
				return err
			}

			// Save data private key after success encryption
			dataPrkStore.Save(zoId)
		}

		if dsType != "offline" {
			drgMsg := common.DRGMsg{
				DoType: db.DoType_Default,
				DoId:   ztdo.GetObjectID(),
				DbId:   a.GetDataBrokerId(),
				DataSourceType: dsType,
				AccessUrl: accessUrl,
				AccessByNHP: false,
				Spo: smartPolicy,
			}

			a.SendDHPRegister(drgMsg)
		}

		os.Exit(0)
	} else if mode == "decrypt" {
		if err := ztdo.ParseHeader(ztdoFilePath); err != nil {
			log.Error("failed to parse ztdo header:%s\n", err)
			fmt.Printf("failed to parse ztdo header:%s\n", err)
			os.Exit(1)
		}

		dataKeyPairEccMode := ztdo.GetECCMode()

		dataPrk, _ := base64.StdEncoding.DecodeString(dataPrivateKeyBase64)
		sa := ztdolib.NewSymmetricAgreement(dataKeyPairEccMode, false)
		sa.SetMessagePatterns(dataMsgPattern)
		sa.SetStaticKeyPair(core.ECDHFromKey(dataKeyPairEccMode.ToEccType(), dataPrk))

		providerPublicKey, _ := base64.StdEncoding.DecodeString(providerPublicKeyBase64)
		sa.SetRemoteStaticPublicKey(providerPublicKey)

		gcmKey, ad := sa.AgreeSymmetricKey()

		log.Info("Decrypting ztdo file(file name: %s and ztdo id: %s) with cipher settings: ECC mode(%s) and Symmetric Cipher Mode(%s)\n", ztdoFilePath, ztdo.GetObjectID(), dataKeyPairEccMode, ztdo.GetCipherMode())

		if err := ztdo.DecryptZtdoFile(ztdoFilePath, output, gcmKey[:], ad); err != nil {
			log.Error("failed to decrypt ztdo file:%s\n", err)
			fmt.Printf("failed to decrypt ztdo file:%s\n", err)
			os.Exit(1)
		} else {
			log.Info("Decrypt ztdo file successfully\n")
			fmt.Printf("Decrypt ztdo file successfully\n")
		}

		os.Exit(0)
	}

	return nil
}
