package main

import (
	"fmt"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"

	"github.com/OpenNHP/opennhp/ac"
	"github.com/OpenNHP/opennhp/nhp"
	"github.com/OpenNHP/opennhp/version"
	"github.com/urfave/cli/v2"
)

func main() {
	app := cli.NewApp()
	app.Name = "nhp-ac"
	app.Usage = "door entity for NHP protocol"
	app.Version = version.Version

	runCmd := &cli.Command{
		Name:  "run",
		Usage: "create and run door process for NHP protocol",
		Action: func(c *cli.Context) error {
			return runApp()
		},
	}

	keygenCmd := &cli.Command{
		Name:  "keygen",
		Usage: "generate key pairs for NHP devices",
		Flags: []cli.Flag{
			&cli.BoolFlag{Name: "curve", Value: false, DisableDefaultText: true, Usage: "generate curve25519 keys (default)"},
			&cli.BoolFlag{Name: "sm2", Value: false, DisableDefaultText: true, Usage: "generate sm2 keys"},
		},
		Action: func(c *cli.Context) error {
			var e nhp.Ecdh
			if c.Bool("sm2") {
				e = nhp.NewECDH(nhp.ECC_SM2)
			} else {
				e = nhp.NewECDH(nhp.ECC_CURVE25519)
			}
			pub := e.PublicKeyBase64()
			priv := e.PrivateKeyBase64()
			fmt.Println("Private key: ", priv)
			fmt.Println("Public key: ", pub)
			return nil
		},
	}

	app.Commands = []*cli.Command{
		runCmd,
		keygenCmd,
	}

	if err := app.Run(os.Args); err != nil {
		panic(err)
	}
}

func runApp() error {
	exeFilePath, err := os.Executable()
	if err != nil {
		return err
	}
	exeDirPath := filepath.Dir(exeFilePath)

	d := &ac.UdpDoor{}
	err = d.Start(exeDirPath, 4)
	if err != nil {
		return err
	}

	// react to terminate signals
	termCh := make(chan os.Signal, 1)
	signal.Notify(termCh, syscall.SIGTERM, os.Interrupt, syscall.SIGABRT)

	// block until terminated
	<-termCh
	d.Stop()

	return nil
}
