package main

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"os"
	"os/signal"
	"path/filepath"
	"runtime"
	"runtime/pprof"
	"syscall"
	"time"

	"github.com/urfave/cli/v2"

	"github.com/OpenNHP/opennhp/endpoints/server"
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
			&cli.BoolFlag{Name: "curve", Value: false, DisableDefaultText: true, Usage: "generate curve25519 keys only"},
			&cli.BoolFlag{Name: "sm2", Value: false, DisableDefaultText: true, Usage: "generate sm2 keys only (default)"},
			&cli.BoolFlag{Name: "both", Value: false, DisableDefaultText: true, Usage: "generate both SM2 and Curve25519 keys from one private key"},
			&cli.BoolFlag{Name: "json", Value: false, DisableDefaultText: true, Usage: "output in JSON format"},
		},
		Action: func(c *cli.Context) error {
			bothSchemes := c.Bool("both")
			curveOnly := c.Bool("curve") && !bothSchemes

			if bothSchemes {
				// Generate one private key and derive both public keys from it.
				e := core.NewECDH(core.ECC_SM2)
				priv := e.PrivateKeyBase64()
				sm2Pub := e.PublicKeyBase64()
				privBytes := e.PrivateKey()
				curvePub := core.ECDHFromKey(core.ECC_CURVE25519, privBytes).PublicKeyBase64()
				if c.Bool("json") {
					output := map[string]string{
						"privateKey":          priv,
						"sm2PublicKey":        sm2Pub,
						"curve25519PublicKey": curvePub,
					}
					json.NewEncoder(os.Stdout).Encode(output)
				} else {
					fmt.Println("Private key:          ", priv)
					fmt.Println("SM2 public key:       ", sm2Pub)
					fmt.Println("Curve25519 public key:", curvePub)
				}
				return nil
			}

			eccType := core.ECC_SM2
			if curveOnly {
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

	// pubkey derives public keys from an EXISTING private key. Used by
	// scripts/generate-nhp-keys.sh to backfill the SM2/Curve25519 public
	// keys for a legacy secret that only stored one of them, WITHOUT
	// rotating the (scheme-agnostic) private key — the same 32 bytes yield
	// both an SM2 and a Curve25519 public key.
	pubkeyCmd := &cli.Command{
		Name:  "pubkey",
		Usage: "derive public key(s) from an existing base64 private key",
		Flags: []cli.Flag{
			&cli.BoolFlag{Name: "curve", Value: false, DisableDefaultText: true, Usage: "output curve25519 public key"},
			&cli.BoolFlag{Name: "sm2", Value: false, DisableDefaultText: true, Usage: "output sm2 public key (default)"},
			&cli.BoolFlag{Name: "both", Value: false, DisableDefaultText: true, Usage: "output both SM2 and Curve25519 public keys"},
			&cli.BoolFlag{Name: "json", Value: false, DisableDefaultText: true, Usage: "output in JSON format"},
		},
		Action: func(c *cli.Context) error {
			emitErr := func(err error) error {
				if c.Bool("json") {
					json.NewEncoder(os.Stdout).Encode(map[string]string{"error": err.Error()})
					return nil
				}
				return err
			}
			privBytes, err := base64.StdEncoding.DecodeString(c.Args().First())
			if err != nil {
				return emitErr(fmt.Errorf("decode private key: %w", err))
			}

			if c.Bool("both") {
				sm2 := core.ECDHFromKey(core.ECC_SM2, privBytes)
				curve := core.ECDHFromKey(core.ECC_CURVE25519, privBytes)
				if sm2 == nil || curve == nil {
					return emitErr(fmt.Errorf("invalid input key"))
				}
				if c.Bool("json") {
					json.NewEncoder(os.Stdout).Encode(map[string]string{
						"sm2PublicKey":        sm2.PublicKeyBase64(),
						"curve25519PublicKey": curve.PublicKeyBase64(),
					})
				} else {
					fmt.Println("SM2 public key:       ", sm2.PublicKeyBase64())
					fmt.Println("Curve25519 public key:", curve.PublicKeyBase64())
				}
				return nil
			}

			eccType := core.ECC_SM2
			if c.Bool("curve") {
				eccType = core.ECC_CURVE25519
			}
			e := core.ECDHFromKey(eccType, privBytes)
			if e == nil {
				return emitErr(fmt.Errorf("invalid input key"))
			}
			if c.Bool("json") {
				json.NewEncoder(os.Stdout).Encode(map[string]string{"publicKey": e.PublicKeyBase64()})
			} else {
				fmt.Println("Public key: ", e.PublicKeyBase64())
			}
			return nil
		},
	}

	app.Commands = []*cli.Command{
		runCmd,
		keygenCmd,
		pubkeyCmd,
	}

	if err := app.Run(os.Args); err != nil {
		panic(err)
	}
}

func printBanner() {
	banner := `
` + colorCyan + colorBold + `
   ____                   _   _ _   _ ____  
  / __ \                 | \ | | | | |  _ \ 
 | |  | |_ __   ___ _ __ |  \| | |_| | |_) |
 | |  | | '_ \ / _ \ '_ \| . ' |  _  |  __/ 
 | |__| | |_) |  __/ | | | |\  | | | | |    
  \____/| .__/ \___|_| |_|_| \_|_| |_|_|    
        | |                                  
        |_|  ` + colorReset + colorDim + `Network-infrastructure Hiding Protocol` + colorReset + `
` + colorPurple + `
  ⭐ GitHub: ` + colorReset + `https://github.com/OpenNHP/opennhp
` + colorYellow + `  💡 Star us & Join the community! Contributors welcome!` + colorReset + `

`
	fmt.Print(banner)
}

func printServerInfo(us *server.UdpServer) {
	// Safely get commit ID (first 12 chars or full if shorter)
	commitId := version.CommitId
	if len(commitId) > 12 {
		commitId = commitId[:12]
	}

	fmt.Println(colorGreen + "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━" + colorReset)
	fmt.Println()
	fmt.Printf("  %s🚀 NHP-Server%s is running!\n", colorBold, colorReset)
	fmt.Println()
	fmt.Printf("  %sVersion:%s    %s\n", colorYellow, colorReset, version.Version)
	fmt.Printf("  %sCommit:%s     %s\n", colorYellow, colorReset, commitId)
	fmt.Printf("  %sBuild:%s      %s\n", colorYellow, colorReset, version.BuildTime)
	fmt.Printf("  %sPlatform:%s   %s/%s\n", colorYellow, colorReset, runtime.GOOS, runtime.GOARCH)
	fmt.Println()
	fmt.Printf("  %sUDP Port:%s   %s%d%s\n", colorBlue, colorReset, colorCyan, us.GetListenPort(), colorReset)

	// Display HTTP status
	httpPort, httpEnabled := us.GetHttpPort()
	if httpEnabled {
		fmt.Printf("  %sHTTP Port:%s  %s%d%s (TLS: %s)\n", colorBlue, colorReset, colorCyan, httpPort, colorReset, us.GetHttpTLSStatus())
	} else {
		fmt.Printf("  %sHTTP:%s       %sdisabled%s\n", colorBlue, colorReset, colorDim, colorReset)
	}

	fmt.Printf("  %sStarted:%s    %s\n", colorBlue, colorReset, time.Now().Format("2006-01-02 15:04:05"))
	fmt.Println()
	fmt.Println(colorGreen + "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━" + colorReset)
	fmt.Println()
	fmt.Printf("  %sPress Ctrl+C to stop the server%s\n", colorDim, colorReset)
	fmt.Println()
}

func runApp(enableProfiling bool) error {
	exeFilePath, err := os.Executable()
	if err != nil {
		return err
	}
	exeDirPath := filepath.Dir(exeFilePath)

	if enableProfiling {
		// Start profiling
		f, createErr := os.Create(filepath.Join(exeDirPath, "cpu.prf"))
		if createErr == nil {
			_ = pprof.StartCPUProfile(f)
			defer pprof.StopCPUProfile()
		}
	}

	// Print banner before starting
	printBanner()

	us := server.UdpServer{}
	err = us.Start(exeDirPath, 4)
	if err != nil {
		fmt.Printf("\n  %s❌ Failed to start server:%s %v\n\n", colorYellow, colorReset, err)
		return err
	}

	// Print server info after successful start
	printServerInfo(&us)

	// react to terminate signals
	termCh := make(chan os.Signal, 1)
	signal.Notify(termCh, syscall.SIGTERM, os.Interrupt)

	// block until terminated
	<-termCh

	fmt.Printf("\n  %s🛑 Shutting down server...%s\n", colorYellow, colorReset)
	us.Stop()
	fmt.Printf("  %s✅ Server stopped gracefully%s\n\n", colorGreen, colorReset)

	return nil
}
