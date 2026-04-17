package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/signal"
	"path/filepath"
	"runtime"
	"syscall"
	"time"

	"github.com/urfave/cli/v2"

	"github.com/OpenNHP/opennhp/endpoints/relay"
	"github.com/OpenNHP/opennhp/nhp/core"
	"github.com/OpenNHP/opennhp/nhp/version"
)

// ANSI color codes
const (
	colorReset  = "\033[0m"
	colorCyan   = "\033[36m"
	colorGreen  = "\033[32m"
	colorYellow = "\033[33m"
	colorBlue   = "\033[34m"
	colorPurple = "\033[35m"
	colorBold   = "\033[1m"
	colorDim    = "\033[2m"
)

func main() {
	app := cli.NewApp()
	app.Name = "nhp-relay"
	app.Usage = "HTTP relay for NHP protocol (bridges HTTPS browser clients to UDP NHP Server)"
	app.Version = version.Version

	runCmd := &cli.Command{
		Name:  "run",
		Usage: "start the NHP relay service",
		Action: func(c *cli.Context) error {
			return runApp()
		},
	}

	keygenCmd := &cli.Command{
		Name:  "keygen",
		Usage: "generate key pairs for NHP devices",
		Flags: []cli.Flag{
			&cli.BoolFlag{Name: "curve", Value: false, DisableDefaultText: true, Usage: "generate curve25519 keys"},
			&cli.BoolFlag{Name: "sm2", Value: false, DisableDefaultText: true, Usage: "generate sm2 keys (default)"},
			&cli.BoolFlag{Name: "json", Value: false, DisableDefaultText: true, Usage: "output in JSON format"},
		},
		Action: func(c *cli.Context) error {
			eccType := core.ECC_SM2
			if c.Bool("curve") {
				eccType = core.ECC_CURVE25519
			}
			e := core.NewECDH(eccType)
			pub := e.PublicKeyBase64()
			priv := e.PrivateKeyBase64()
			if c.Bool("json") {
				output := map[string]string{
					"privateKey": priv,
					"publicKey":  pub,
				}
				json.NewEncoder(os.Stdout).Encode(output)
			} else {
				fmt.Println("Private key: ", priv)
				fmt.Println("Public key: ", pub)
			}
			return nil
		},
	}

	app.Commands = []*cli.Command{
		runCmd,
		keygenCmd,
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
	relay.ExeDirPath = exeDirPath

	printBanner()

	cfg, err := relay.LoadConfig("")
	if err != nil {
		fmt.Fprintf(os.Stderr, "\n  %s❌ Failed to load config: %v%s\n\n", colorYellow, err, colorReset)
		return err
	}

	rs, err := relay.New(cfg)
	if err != nil {
		fmt.Fprintf(os.Stderr, "\n  %s❌ Failed to create relay: %v%s\n\n", colorYellow, err, colorReset)
		return err
	}

	printRelayInfo(cfg)

	termCh := make(chan os.Signal, 1)
	signal.Notify(termCh, syscall.SIGTERM, os.Interrupt, syscall.SIGABRT)

	errCh := make(chan error, 1)
	go func() {
		if err := rs.Start(); err != nil {
			errCh <- err
		}
	}()

	select {
	case sig := <-termCh:
		fmt.Printf("\n  %s🛑 Received signal %s, shutting down relay...%s\n", colorYellow, sig, colorReset)
	case err := <-errCh:
		fmt.Printf("\n  %s❌ Relay error: %v%s\n", colorYellow, err, colorReset)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := rs.Stop(ctx); err != nil {
		fmt.Fprintf(os.Stderr, "  relay shutdown error: %v\n", err)
	}
	fmt.Printf("  %s✅ Relay stopped gracefully%s\n\n", colorGreen, colorReset)
	return nil
}

func printBanner() {
	fmt.Print(colorCyan + colorBold + `
  _   _ _   _ ____     ____      _
 | \ | | | | |  _ \   |  _ \ ___| | __ _ _   _
 |  \| | |_| | |_) |  | |_) / _ \ |/ _' | | | |
 | |\  |  _  |  __/   |  _ <  __/ | (_| | |_| |
 |_| \_|_| |_|_|      |_| \_\___|_|\__,_|\__, |
                                           |___/ ` + colorReset + colorDim + `HTTP→UDP Bridge` + colorReset + colorPurple + `

  ⭐ GitHub: ` + colorReset + `https://github.com/OpenNHP/opennhp
`)
}

func printRelayInfo(cfg *relay.Config) {
	commitId := version.CommitId
	if len(commitId) > 12 {
		commitId = commitId[:12]
	}

	listenScheme := "http"
	if cfg.EnableTLS {
		listenScheme = "https"
	}

	fmt.Println(colorGreen + "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━" + colorReset)
	fmt.Printf("  %s🔁 NHP-Relay%s is running!\n\n", colorBold, colorReset)
	fmt.Printf("  %sVersion:%s    %s\n", colorYellow, colorReset, version.Version)
	fmt.Printf("  %sCommit:%s     %s\n", colorYellow, colorReset, commitId)
	fmt.Printf("  %sBuild:%s      %s\n", colorYellow, colorReset, version.BuildTime)
	fmt.Printf("  %sPlatform:%s   %s/%s\n\n", colorYellow, colorReset, runtime.GOOS, runtime.GOARCH)
	fmt.Printf("  %sListen:%s     %s://%s:%d/relay\n", colorBlue, colorReset, listenScheme, cfg.ListenIP, cfg.ListenPort)
	fmt.Printf("  %sUpstream:%s   udp://%s:%d\n", colorBlue, colorReset, cfg.NHPServerHost, cfg.NHPServerPort)
	fmt.Printf("  %sUDP timeout:%s %dms\n", colorBlue, colorReset, cfg.UDPTimeoutMs)
	fmt.Printf("  %sStarted:%s    %s\n\n", colorBlue, colorReset, time.Now().Format("2006-01-02 15:04:05"))
	fmt.Println(colorGreen + "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━" + colorReset)
	fmt.Printf("\n  %sPress Ctrl+C to stop%s\n\n", colorDim, colorReset)
}
