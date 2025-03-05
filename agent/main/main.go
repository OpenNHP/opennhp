package main

import (
	"encoding/base64"
	"fmt"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"

	"github.com/OpenNHP/opennhp/agent"
	"github.com/OpenNHP/opennhp/core"
	"github.com/OpenNHP/opennhp/version"
	"github.com/urfave/cli/v2"
)

func main() {
	app := cli.NewApp()
	app.Name = "nhp-agent"
	app.Usage = "agent entity for NHP protocol"
	app.Version = version.Version

	runCmd := &cli.Command{
		Name:  "run",
		Usage: "create and run agent process for NHP protocol",
		Action: func(c *cli.Context) error {
			return runApp()
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
	dhpCmd := &cli.Command{
		Name:  "dhp",
		Usage: "create and dhp agent process for NHP protocol",
		Flags: []cli.Flag{
			&cli.StringFlag{Name: "ztdo", Value: "", Usage: "Ztdo file path"},
			&cli.StringFlag{Name: "output", Value: "", Usage: "Decrypted file output path"},
		},
		Action: func(c *cli.Context) error {
			ztdo := c.String("ztdo")
			output := c.String("output")
			return runDHPApp(ztdo, output)
		},
	}

	app.Commands = []*cli.Command{
		runCmd,
		keygenCmd,
		pubkeyCmd,
		dhpCmd,
	}

	if err := app.Run(os.Args); err != nil {
		fmt.Fprintln(os.Stderr, err)
	}
}

func runApp() error {
	exeFilePath, err := os.Executable()
	if err != nil {
		return err
	}
	exeDirPath := filepath.Dir(exeFilePath)

	a := &agent.UdpAgent{}

	err = a.Start(exeDirPath, 4)
	if err != nil {
		return err
	}
	a.StartKnockLoop()
	// react to terminate signals
	termCh := make(chan os.Signal, 1)
	signal.Notify(termCh, syscall.SIGTERM, os.Interrupt, syscall.SIGABRT)

	// block until terminated
	<-termCh
	a.Stop()
	return nil
}
func runDHPApp(ztdo string, output string) error {
	exeFilePath, err := os.Executable()
	if err != nil {
		return err
	}
	exeDirPath := filepath.Dir(exeFilePath)

	a := &agent.UdpAgent{}

	err = a.Start(exeDirPath, 4)
	if err != nil {
		return err
	}
	if ztdo != "" {
		//request ztdo file
		a.StartDecodeZtdo(ztdo, output)
	} else {
		a.StartKnockLoop()
	}

	// react to terminate signals
	termCh := make(chan os.Signal, 1)
	signal.Notify(termCh, syscall.SIGTERM, os.Interrupt, syscall.SIGABRT)

	// block until terminated
	<-termCh
	a.Stop()
	return nil
}
