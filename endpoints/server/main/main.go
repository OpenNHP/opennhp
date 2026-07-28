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
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/urfave/cli/v2"

	"github.com/OpenNHP/opennhp/endpoints/server"
	"github.com/OpenNHP/opennhp/nhp/audit"
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

	// audit verifies the integrity of a security audit ledger produced by
	// the server's tamper-evident audit log. It walks the hash chain and
	// reports the first break, if any.
	auditCmd := &cli.Command{
		Name:  "audit",
		Usage: "tools for the tamper-evident security audit ledger",
		Subcommands: []*cli.Command{
			{
				Name:      "verify",
				Usage:     "verify the hash chain of an audit ledger file",
				ArgsUsage: "<ledgerFile>",
				Flags: []cli.Flag{
					&cli.StringFlag{Name: "key", Usage: "base64 HMAC signing key, if the ledger was signed"},
					&cli.BoolFlag{Name: "strict", Usage: "exit non-zero (2) if verification is incomplete: damaged (skipped) lines, or signed entries left unchecked because no --key was given"},
				},
				Action: func(c *cli.Context) error {
					path := c.Args().First()
					if path == "" {
						return fmt.Errorf("usage: audit verify <ledgerFile>")
					}
					var hmacKey []byte
					if k := c.String("key"); k != "" {
						decoded, err := base64.StdEncoding.DecodeString(k)
						if err != nil {
							return fmt.Errorf("decode --key: %w", err)
						}
						hmacKey = decoded
					}
					res, err := verifyLedgerFile(path, hmacKey)
					if err != nil {
						return err
					}

					if res.Err != nil {
						fmt.Printf("FAILED: %v\n", res.Err)
						fmt.Printf("%d entr%s verified before the break.\n", res.Count, plural(res.Count))
						// Exit non-zero with a clean message rather than a
						// panic stack trace — this is a verification tool and
						// a failed check is an expected, reportable outcome.
						os.Exit(1)
					}
					fmt.Printf("OK: %d entr%s, hash chain intact.\n", res.Count, plural(res.Count))
					if res.UncheckedSigs > 0 {
						// Do not let a signed ledger checked without its key
						// read as fully verified. The hash chain alone is
						// forgeable by anyone who can rewrite the file; the
						// signatures are the part that isn't, and they were
						// not checked here.
						fmt.Printf("warning: %d signed entr%s NOT verified — no --key given, so only the hash chain was checked.\n",
							res.UncheckedSigs, plural(res.UncheckedSigs))
					}
					if res.Skipped > 0 {
						// Damage, not tampering: the chain still links up, so
						// no committed entry was altered or removed. Say so
						// explicitly rather than leaving it to look like a
						// clean bill of health.
						fmt.Printf("note: skipped %d unparseable line(s)%s — likely a torn write from an unclean shutdown, not tampering.\n",
							res.Skipped, formatSkippedLines(res.SkippedLines, res.Skipped))
					}
					// In --strict mode an incomplete verification is a failure
					// for gating purposes (CI/cron): a keyless check of a
					// signed ledger, or a file with damaged lines, is not the
					// same as a full clean pass. Distinct exit code 2 so a
					// caller can tell it apart from a chain break (1).
					if c.Bool("strict") && (res.Skipped > 0 || res.UncheckedSigs > 0) {
						fmt.Println("strict: verification incomplete (see warnings above).")
						os.Exit(2)
					}
					return nil
				},
			},
		},
	}

	app.Commands = []*cli.Command{
		runCmd,
		keygenCmd,
		pubkeyCmd,
		auditCmd,
	}

	if err := app.Run(os.Args); err != nil {
		panic(err)
	}
}

// verifyLedgerFile opens the ledger and walks its chain. It exists so the
// file handle is closed before the caller decides whether to exit — a
// deferred Close in the command action would never run on the os.Exit path.
func verifyLedgerFile(path string, hmacKey []byte) (audit.VerifyResult, error) {
	f, err := os.Open(filepath.Clean(path))
	if err != nil {
		return audit.VerifyResult{}, err
	}
	defer f.Close()
	return audit.VerifyChain(f, hmacKey), nil
}

// plural returns the English plural suffix for "entry"/"entries" counts.
func plural(n uint64) string {
	if n == 1 {
		return "y"
	}
	return "ies"
}

// formatSkippedLines renders the line numbers of skipped lines for the note,
// e.g. " (lines 3, 7)". It returns "" when none were captured, and marks the
// list as partial when the reported slice is shorter than the total skipped.
func formatSkippedLines(lines []uint64, total uint64) string {
	if len(lines) == 0 {
		return ""
	}
	parts := make([]string, len(lines))
	for i, ln := range lines {
		parts[i] = strconv.FormatUint(ln, 10)
	}
	list := strings.Join(parts, ", ")
	if uint64(len(lines)) < total {
		list += ", …"
	}
	return " (line" + plural2(len(lines)) + " " + list + ")"
}

// plural2 returns "" for one item and "s" otherwise, for "line"/"lines".
func plural2(n int) string {
	if n == 1 {
		return ""
	}
	return "s"
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
