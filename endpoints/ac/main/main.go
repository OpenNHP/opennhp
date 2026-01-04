package main

import (
	"fmt"
	"os"
	"os/signal"
	"path/filepath"
	"runtime"
	"syscall"
	"time"

	"github.com/OpenNHP/opennhp/endpoints/ac"
	"github.com/OpenNHP/opennhp/endpoints/ac/ebpf"
	"github.com/OpenNHP/opennhp/nhp/core"
	"github.com/OpenNHP/opennhp/nhp/version"
	"github.com/urfave/cli/v2"
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
	app.Name = "nhp-ac"
	app.Usage = "ac entity for NHP protocol"
	app.Version = version.Version

	runCmd := &cli.Command{
		Name:  "run",
		Usage: "create and run ac process for NHP protocol",
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

	app.Commands = []*cli.Command{
		runCmd,
		keygenCmd,
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

func getFilterModeName(mode int) string {
	switch mode {
	case ac.FilterMode_IPTABLES:
		return "IPTables"
	case ac.FilterMode_EBPFXDP:
		return "eBPF/XDP"
	default:
		return "Unknown"
	}
}

func printACInfo(cfg *ac.Config) {
	// Safely get commit ID (first 12 chars or full if shorter)
	commitId := version.CommitId
	if len(commitId) > 12 {
		commitId = commitId[:12]
	}

	fmt.Println(colorGreen + "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━" + colorReset)
	fmt.Println()
	fmt.Printf("  %s🛡️  NHP-AC%s (Access Controller) is running!\n", colorBold, colorReset)
	fmt.Println()
	fmt.Printf("  %sVersion:%s    %s\n", colorYellow, colorReset, version.Version)
	fmt.Printf("  %sCommit:%s     %s\n", colorYellow, colorReset, commitId)
	fmt.Printf("  %sBuild:%s      %s\n", colorYellow, colorReset, version.BuildTime)
	fmt.Printf("  %sPlatform:%s   %s/%s\n", colorYellow, colorReset, runtime.GOOS, runtime.GOARCH)
	fmt.Println()
	fmt.Printf("  %sFilter:%s     %s%s%s\n", colorBlue, colorReset, colorCyan, getFilterModeName(cfg.FilterMode), colorReset)
	if cfg.ACId != "" {
		fmt.Printf("  %sAC ID:%s      %s\n", colorBlue, colorReset, cfg.ACId)
	}
	fmt.Printf("  %sStarted:%s    %s\n", colorBlue, colorReset, time.Now().Format("2006-01-02 15:04:05"))
	fmt.Println()
	fmt.Println(colorGreen + "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━" + colorReset)
	fmt.Println()
	fmt.Printf("  %sPress Ctrl+C to stop the access controller%s\n", colorDim, colorReset)
	fmt.Println()
}

func runApp() error {
	exeFilePath, err := os.Executable()
	if err != nil {
		return err
	}
	exeDirPath := filepath.Dir(exeFilePath)

	// Print banner before starting
	printBanner()

	d := &ac.UdpAC{}
	err = d.Start(exeDirPath, 4)
	if err != nil {
		fmt.Printf("\n  %s❌ Failed to start AC:%s %v\n\n", colorYellow, colorReset, err)
		return err
	}
	cfg := d.GetConfig()

	// Print AC info after successful start
	printACInfo(cfg)

	// react to terminate signals
	termCh := make(chan os.Signal, 1)
	signal.Notify(termCh, syscall.SIGTERM, os.Interrupt, syscall.SIGABRT)
	if cfg.FilterMode == ac.FilterMode_EBPFXDP {
		defer ebpf.CleanupBPFFiles()
	}
	// block until terminated
	<-termCh

	fmt.Printf("\n  %s🛑 Shutting down access controller...%s\n", colorYellow, colorReset)
	d.Stop()
	fmt.Printf("  %s✅ Access controller stopped gracefully%s\n\n", colorGreen, colorReset)

	return nil
}
