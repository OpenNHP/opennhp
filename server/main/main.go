package main

import (
	"fmt"
	"os"
	"os/signal"
	"path/filepath"
	"runtime/pprof"
	"syscall"

	"github.com/OpenNHP/opennhp/nhp"
	"github.com/OpenNHP/opennhp/server"
	"github.com/OpenNHP/opennhp/version"
	"github.com/urfave/cli/v2"
)

func main() {
	app := cli.NewApp()
	app.Name = "nhp-server"
	app.Usage = "server entity for NHP protocol"
	app.Version = version.Version

	runCmd := &cli.Command{
		Name:  "run",
		Usage: "create and run server process for NHP protocol",
		Flags: []cli.Flag{
			&cli.BoolFlag{Name: "prof", Value: false, DisableDefaultText: true, Usage: "running profiling for the server"},
		},
		Action: func(c *cli.Context) error {
			return runApp(c.Bool("prof"))
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

func runApp(enableProfiling bool) error {
	exeFilePath, err := os.Executable()
	if err != nil {
		return err
	}
	exeDirPath := filepath.Dir(exeFilePath)

	if enableProfiling {
		// Start profiling
		f, err := os.Create(filepath.Join(exeDirPath, "cpu.prf"))
		if err == nil {
			pprof.StartCPUProfile(f)
			defer pprof.StopCPUProfile()
		}
	}

	us := server.UdpServer{}
	err = us.Start(exeDirPath, 4)
	if err != nil {
		return err
	}

	// react to terminate signals
	termCh := make(chan os.Signal, 1)
	signal.Notify(termCh, syscall.SIGTERM, os.Interrupt)

	// block until terminated
	<-termCh
	us.Stop()

	return nil
}
