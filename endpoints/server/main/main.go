package main

import (
	"fmt"
	"os"
	"os/signal"
	"path/filepath"
	"runtime"
	"runtime/pprof"
	"syscall"
	"time"

	"github.com/OpenNHP/opennhp/endpoints/server"
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
		f, err := os.Create(filepath.Join(exeDirPath, "cpu.prf"))
		if err == nil {
			pprof.StartCPUProfile(f)
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
