package main

import (
	"bufio"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"os"
	"os/signal"
	"path/filepath"
	"runtime"
	"strings"
	"syscall"
	"time"

	"github.com/urfave/cli/v2"

	"github.com/OpenNHP/opennhp/endpoints/agent"
	"github.com/OpenNHP/opennhp/nhp/common"
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
			&cli.BoolFlag{Name: "json", Value: false, DisableDefaultText: true, Usage: "output in JSON format"},
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

	pubkeyCmd := &cli.Command{
		Name:  "pubkey",
		Usage: "get public key from private key",
		Flags: []cli.Flag{
			&cli.BoolFlag{Name: "curve", Value: false, DisableDefaultText: true, Usage: "get curve25519 key"},
			&cli.BoolFlag{Name: "sm2", Value: false, DisableDefaultText: true, Usage: "get sm2 key (default)"},
			&cli.BoolFlag{Name: "json", Value: false, DisableDefaultText: true, Usage: "output in JSON format"},
		},
		Action: func(c *cli.Context) error {
			privKey, err := base64.StdEncoding.DecodeString(c.Args().First())
			if err != nil {
				if c.Bool("json") {
					json.NewEncoder(os.Stdout).Encode(map[string]interface{}{
						"error": err.Error(),
					})
					return nil
				}
				return err
			}
			eccType := core.ECC_SM2
			if c.Bool("curve") {
				eccType = core.ECC_CURVE25519
			}
			e := core.ECDHFromKey(eccType, privKey)
			if e == nil {
				err := fmt.Errorf("invalid input key")
				if c.Bool("json") {
					json.NewEncoder(os.Stdout).Encode(map[string]interface{}{
						"error": err.Error(),
					})
					return nil
				}
				return err
			}
			pub := e.PublicKeyBase64()
			if c.Bool("json") {
				json.NewEncoder(os.Stdout).Encode(map[string]string{
					"publicKey": pub,
				})
			} else {
				fmt.Println("Public key: ", pub)
			}
			return nil
		},
	}
	dhpCmd := &cli.Command{
		Name:  "dhp",
		Usage: "create dhp agent process for NHP protocol",
		Action: func(c *cli.Context) error {
			return runDHPApp()
		},
	}

	registerCmd := &cli.Command{
		Name:  "register",
		Usage: "register agent public key with nhp-server (interactive OTP → REG two-step flow)",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:  "email",
				Usage: "your email address (UserId); prompted interactively if not provided",
				Value: "",
			},
			&cli.StringFlag{
				Name:  "asp-id",
				Usage: "AuthServiceId to register with; prompted interactively if not provided",
				Value: "",
			},
			&cli.StringFlag{
				Name:  "res-id",
				Usage: "ResourceId (optional); prompted interactively if not provided",
				Value: "",
			},
			&cli.StringFlag{
				Name:  "server",
				Usage: "cluster name from server.toml (optional, overrides resource.toml binding)",
				Value: "",
			},
			&cli.StringFlag{
				Name:  "device-id",
				Usage: "device identifier (optional)",
				Value: "",
			},
			&cli.StringFlag{
				Name:  "org-id",
				Usage: "organization identifier (optional)",
				Value: "",
			},
			&cli.StringFlag{
				Name:  "otp",
				Usage: "OTP code (optional; if provided, skips the OTP request step)",
				Value: "",
			},
		},
		Action: func(c *cli.Context) error {
			return runRegisterApp(
				c.String("email"),
				c.String("asp-id"),
				c.String("res-id"),
				c.String("server"),
				c.String("device-id"),
				c.String("org-id"),
				c.String("otp"),
			)
		},
	}

	app.Commands = []*cli.Command{
		runCmd,
		keygenCmd,
		pubkeyCmd,
		dhpCmd,
		registerCmd,
	}

	if err := app.Run(os.Args); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
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

func printAgentInfo() {
	// Safely get commit ID (first 12 chars or full if shorter)
	commitId := version.CommitId
	if len(commitId) > 12 {
		commitId = commitId[:12]
	}

	fmt.Println(colorGreen + "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━" + colorReset)
	fmt.Println()
	fmt.Printf("  %s🔐 NHP-Agent%s is running!\n", colorBold, colorReset)
	fmt.Println()
	fmt.Printf("  %sVersion:%s    %s\n", colorYellow, colorReset, version.Version)
	fmt.Printf("  %sCommit:%s     %s\n", colorYellow, colorReset, commitId)
	fmt.Printf("  %sBuild:%s      %s\n", colorYellow, colorReset, version.BuildTime)
	fmt.Printf("  %sPlatform:%s   %s/%s\n", colorYellow, colorReset, runtime.GOOS, runtime.GOARCH)
	fmt.Println()
	fmt.Printf("  %sMode:%s       %sKnock Client%s\n", colorBlue, colorReset, colorCyan, colorReset)
	fmt.Printf("  %sStarted:%s    %s\n", colorBlue, colorReset, time.Now().Format("2006-01-02 15:04:05"))
	fmt.Println()
	fmt.Println(colorGreen + "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━" + colorReset)
	fmt.Println()
	fmt.Printf("  %sPress Ctrl+C to stop the agent%s\n", colorDim, colorReset)
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

	a := &agent.UdpAgent{}

	err = a.Start(exeDirPath, 4)
	if err != nil {
		fmt.Printf("\n  %s❌ Failed to start agent:%s %v\n\n", colorYellow, colorReset, err)
		return err
	}

	// Print agent info after successful start
	printAgentInfo()

	a.StartKnockLoop()
	// react to terminate signals
	termCh := make(chan os.Signal, 1)
	signal.Notify(termCh, syscall.SIGTERM, os.Interrupt, syscall.SIGABRT)

	// block until terminated
	<-termCh

	fmt.Printf("\n  %s🛑 Shutting down agent...%s\n", colorYellow, colorReset)
	a.Stop()
	fmt.Printf("  %s✅ Agent stopped gracefully%s\n\n", colorGreen, colorReset)

	return nil
}

func printDHPAgentInfo() {
	// Safely get commit ID (first 12 chars or full if shorter)
	commitId := version.CommitId
	if len(commitId) > 12 {
		commitId = commitId[:12]
	}

	fmt.Println(colorGreen + "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━" + colorReset)
	fmt.Println()
	fmt.Printf("  %s🛡️  NHP-Agent (DHP Mode)%s is running!\n", colorBold, colorReset)
	fmt.Println()
	fmt.Printf("  %sVersion:%s    %s\n", colorYellow, colorReset, version.Version)
	fmt.Printf("  %sCommit:%s     %s\n", colorYellow, colorReset, commitId)
	fmt.Printf("  %sBuild:%s      %s\n", colorYellow, colorReset, version.BuildTime)
	fmt.Printf("  %sPlatform:%s   %s/%s\n", colorYellow, colorReset, runtime.GOOS, runtime.GOARCH)
	fmt.Println()
	fmt.Printf("  %sMode:%s       %sDHP (Data Hiding Protocol)%s\n", colorBlue, colorReset, colorCyan, colorReset)
	fmt.Printf("  %sStarted:%s    %s\n", colorBlue, colorReset, time.Now().Format("2006-01-02 15:04:05"))
	fmt.Println()
	fmt.Println(colorGreen + "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━" + colorReset)
	fmt.Println()
	fmt.Printf("  %sPress Ctrl+C to stop the agent%s\n", colorDim, colorReset)
	fmt.Println()
}

func runDHPApp() error {
	exeFilePath, err := os.Executable()
	if err != nil {
		return err
	}
	exeDirPath := filepath.Dir(exeFilePath)

	common.ExeDirPath = exeDirPath
	agent.ExeDirPath = exeDirPath

	// Print banner before starting
	printBanner()

	a := &agent.UdpAgent{}

	err = a.InitializeSecret()
	if err != nil {
		fmt.Printf("\n  %s❌ Failed to initialize secret:%s %v\n\n", colorYellow, colorReset, err)
		return err
	}

	err = a.Start(exeDirPath, 4)
	if err != nil {
		fmt.Printf("\n  %s❌ Failed to start DHP agent:%s %v\n\n", colorYellow, colorReset, err)
		return err
	}

	// Print DHP agent info after successful start
	printDHPAgentInfo()

	a.CreateDHPWebConsole()
	a.StartDHPKnockLoop()

	// react to terminate signals
	termCh := make(chan os.Signal, 1)
	signal.Notify(termCh, syscall.SIGTERM, os.Interrupt, syscall.SIGABRT)

	// block until terminated
	<-termCh

	fmt.Printf("\n  %s🛑 Shutting down DHP agent...%s\n", colorYellow, colorReset)
	a.Stop()
	fmt.Printf("  %s✅ DHP Agent stopped gracefully%s\n\n", colorGreen, colorReset)

	return nil
}

// promptInput prints a prompt and reads a trimmed line from stdin.
// If defaultVal is non-empty it is shown in brackets and returned when the
// user presses Enter without typing anything.
func promptInput(reader *bufio.Reader, label, defaultVal string) (string, error) {
	if defaultVal != "" {
		fmt.Printf("  %s%s%s [%s%s%s]: ", colorBold, label, colorReset, colorDim, defaultVal, colorReset)
	} else {
		fmt.Printf("  %s%s%s: ", colorBold, label, colorReset)
	}
	line, err := reader.ReadString('\n')
	if err != nil {
		return "", err
	}
	val := strings.TrimSpace(line)
	if val == "" {
		return defaultVal, nil
	}
	return val, nil
}

func runRegisterApp(email, aspId, resId, serverCluster, deviceId, orgId, otpCode string) error {
	exeFilePath, err := os.Executable()
	if err != nil {
		return err
	}
	exeDirPath := filepath.Dir(exeFilePath)

	printBanner()

	fmt.Println(colorGreen + "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━" + colorReset)
	fmt.Printf("\n  %s🔑 NHP-Agent Key Registration%s\n", colorBold, colorReset)
	fmt.Printf("  %sPlease provide the following information:%s\n\n", colorDim, colorReset)

	reader := bufio.NewReader(os.Stdin)

	// --- Interactive prompts for any values not supplied via flags ---

	// Email is the primary identity (UserId). No default — must be provided.
	for email == "" {
		val, readErr := promptInput(reader, "Email address (UserId)", "")
		if readErr != nil {
			fmt.Printf("\n  %s❌ Failed to read input:%s %v\n\n", colorYellow, colorReset, readErr)
			return readErr
		}
		if val == "" {
			fmt.Printf("  %s  Email is required. Please try again.%s\n", colorYellow, colorReset)
			continue
		}
		email = val
	}

	// Step 0: generate a fresh key pair — let the user choose the cipher
	// scheme. Labels match the reg.opennhp.org registration page.
	fmt.Printf("\n  %sCipher scheme:%s\n", colorBold, colorReset)
	fmt.Printf("    %s[1]%s Curve25519 (X25519 + AES-256-GCM + BLAKE2s)\n", colorCyan, colorReset)
	fmt.Printf("    %s[2]%s GMSM / SM2 (SM2 + SM4-GCM + SM3)\n", colorCyan, colorReset)
	schemeInput, _ := promptInput(reader, "Choose cipher scheme", "1")
	schemeInput = strings.TrimSpace(schemeInput)

	var eccType core.EccTypeEnum
	var cipherScheme int
	switch schemeInput {
	case "2":
		eccType = core.ECC_SM2
		cipherScheme = common.CIPHER_SCHEME_GMSM
		fmt.Printf("  %s✔  GMSM / SM2 selected%s\n\n", colorGreen, colorReset)
	default:
		eccType = core.ECC_CURVE25519
		cipherScheme = common.CIPHER_SCHEME_CURVE
		fmt.Printf("  %s✔  Curve25519 selected%s\n\n", colorGreen, colorReset)
	}

	ecdh := core.NewECDH(eccType)
	privKeyBytes := ecdh.PrivateKey()
	fmt.Printf("  %sGenerated key pair:%s\n", colorYellow, colorReset)
	fmt.Printf("    Private key:       %s%s%s\n", colorDim, ecdh.PrivateKeyBase64(), colorReset)
	// The private key is scheme-agnostic: the same bytes derive both an SM2
	// and a Curve25519 public key. Show both regardless of the selected
	// scheme (the one matching the selection is registered), so it's clear
	// one key can serve either cipher. Mark the active one.
	sm2Pub := core.ECDHFromKey(core.ECC_SM2, privKeyBytes).PublicKeyBase64()
	curvePub := core.ECDHFromKey(core.ECC_CURVE25519, privKeyBytes).PublicKeyBase64()
	activeMark := func(gmsm bool) string {
		if (cipherScheme == common.CIPHER_SCHEME_GMSM) == gmsm {
			return colorGreen + "  ← selected" + colorReset
		}
		return ""
	}
	fmt.Printf("    Curve25519 pubkey: %s%s\n", curvePub, activeMark(false))
	fmt.Printf("    SM2 public key:    %s%s\n", sm2Pub, activeMark(true))
	fmt.Println()

	// Start agent with the existing config (for peer/resource resolution), then
	// immediately replace its device with the freshly generated key pair.
	a := &agent.UdpAgent{}
	// register bootstraps an agent before any config.toml exists, so allow a
	// missing base config (empty config + throwaway key, replaced below via
	// ReinitWithKey). run/dhp leave this false and fail loudly on a missing
	// config.
	a.SetAllowMissingConfig(true)
	if err = a.Start(exeDirPath, 2); err != nil {
		fmt.Printf("\n  %s❌ Failed to start agent:%s %v\n\n", colorYellow, colorReset, err)
		return err
	}
	defer a.Stop()

	if err = a.ReinitWithKey(privKeyBytes, cipherScheme); err != nil {
		fmt.Printf("\n  %s❌ Failed to initialize new key:%s %v\n\n", colorYellow, colorReset, err)
		return err
	}

	// asp-id: default to the first resource entry's AuthServiceId if available,
	// falling back to "example" for the standard demo deployment.
	defaultAspId := "example"
	if aspId == "" {
		if kt := a.FirstKnockTarget(); kt != nil && kt.AuthServiceId != "" {
			defaultAspId = kt.AuthServiceId
		}
	}
	if aspId == "" {
		val, readErr := promptInput(reader, "AuthServiceId (asp-id)", defaultAspId)
		if readErr != nil {
			fmt.Printf("\n  %s❌ Failed to read input:%s %v\n\n", colorYellow, colorReset, readErr)
			return readErr
		}
		aspId = strings.TrimSpace(val)
		if aspId == "" {
			fmt.Printf("\n  %s❌ AuthServiceId is required%s\n\n", colorYellow, colorReset)
			return fmt.Errorf("AuthServiceId is required")
		}
	}

	// res-id: default from the matching resource.toml entry for this asp-id,
	// falling back to "demo" for the standard demo deployment.
	defaultResId := "demo"
	if resId == "" {
		if kt := a.FindKnockTargetByAspId(aspId); kt != nil && kt.ResourceId != "" {
			defaultResId = kt.ResourceId
		}
	}
	if resId == "" {
		val, readErr := promptInput(reader, "ResourceId (res-id)", defaultResId)
		if readErr != nil {
			fmt.Printf("\n  %s❌ Failed to read input:%s %v\n\n", colorYellow, colorReset, readErr)
			return readErr
		}
		resId = strings.TrimSpace(val)
	}

	// Server cluster — default to "default" for the standard demo deployment.
	if serverCluster == "" {
		defaultCluster := "default"
		if kt := a.FindKnockTargetByAspId(aspId); kt != nil {
			if sc := kt.GetServerCluster(); sc != nil && sc.Name != "" {
				defaultCluster = sc.Name
			}
		}
		val, readErr := promptInput(reader, "Server cluster name", defaultCluster)
		if readErr != nil {
			fmt.Printf("\n  %s❌ Failed to read input:%s %v\n\n", colorYellow, colorReset, readErr)
			return readErr
		}
		serverCluster = strings.TrimSpace(val)
	}

	// Device ID — optional.
	if deviceId == "" {
		val, readErr := promptInput(reader, "Device ID (optional, Enter to skip)", "")
		if readErr != nil {
			fmt.Printf("\n  %s❌ Failed to read input:%s %v\n\n", colorYellow, colorReset, readErr)
			return readErr
		}
		deviceId = strings.TrimSpace(val)
	}

	// Org ID — optional, default from existing config.
	defaultOrgId := a.OrganizationId()
	if orgId == "" {
		val, readErr := promptInput(reader, "Organization ID (optional)", defaultOrgId)
		if readErr != nil {
			fmt.Printf("\n  %s❌ Failed to read input:%s %v\n\n", colorYellow, colorReset, readErr)
			return readErr
		}
		orgId = strings.TrimSpace(val)
	}

	fmt.Println()

	// Apply collected identity overrides to the running agent before
	// sending. Carry the email in UserData under the "email" key — the ASP
	// plugin reads UserData["email"] as the OTP recipient (same as the
	// js-agent web flow). Without it the plugin falls back to UserId, which
	// works only when UserId happens to be a deliverable address; passing
	// it explicitly matches reg.opennhp.org and avoids the fallback path.
	a.SetKnockUser(email, orgId, map[string]any{"email": email})
	if deviceId != "" {
		a.SetDeviceId(deviceId)
	}

	// Build the KnockResource for registration.
	res := &agent.KnockResource{
		AuthServiceId: aspId,
		ResourceId:    resId,
	}

	// Resolve target cluster: --server flag (or interactive input) takes
	// precedence over resource.toml entries.
	if serverCluster != "" {
		res.Cluster = serverCluster
	} else {
		kt := a.FindKnockTarget(aspId, resId)
		if kt == nil {
			kt = a.FindKnockTargetByAspId(aspId)
		}
		if kt != nil {
			sc := kt.GetServerCluster()
			if sc != nil {
				res.Cluster = sc.Name
			}
		}
	}

	if res.Cluster == "" && res.ServerPubKey == "" {
		fmt.Printf("  %s❌ No server cluster found for asp-id=%q%s\n", colorYellow, aspId, colorReset)
		fmt.Printf("  %sHint: configure a matching [[Resources]] entry in resource.toml or pass --server <cluster-name>%s\n\n", colorDim, colorReset)
		return fmt.Errorf("no server cluster: set Cluster in resource.toml or use --server")
	}

	sc, err := a.FindServerClusterFromResource(res)
	if err != nil {
		fmt.Printf("  %s❌ Cannot resolve server cluster %q: %v%s\n\n", colorYellow, res.Cluster, err, colorReset)
		fmt.Printf("  %sHint: ensure the cluster name matches a [[Servers]] Name in server.toml%s\n\n", colorDim, colorReset)
		return err
	}

	target := &agent.KnockTarget{
		KnockResource: *res,
		ServerPeer:    sc.RepresentativePeer(),
		ServerCluster: sc,
	}

	// Step 1: OTP request (skipped when --otp is provided).
	if otpCode == "" {
		fmt.Printf("  %sStep 1/2:%s Requesting OTP from server...\n", colorBlue, colorReset)
		if err = a.RequestOtp(target); err != nil {
			fmt.Printf("\n  %s❌ OTP request failed:%s %v\n\n", colorYellow, colorReset, err)
			return err
		}
		fmt.Printf("  %s✔  OTP sent.%s Please check your inbox at %s%s%s for the verification code.\n\n",
			colorGreen, colorReset, colorCyan, email, colorReset)

		var readErr error
		for otpCode == "" {
			val, readErr2 := promptInput(reader, "Enter OTP", "")
			readErr = readErr2
			if readErr != nil {
				break
			}
			if val == "" {
				fmt.Printf("  %s  OTP cannot be empty. Please try again.%s\n", colorYellow, colorReset)
				continue
			}
			otpCode = val
		}
		if readErr != nil {
			fmt.Printf("\n  %s❌ Failed to read OTP:%s %v\n\n", colorYellow, colorReset, readErr)
			return readErr
		}
		fmt.Println()
	} else {
		fmt.Printf("  %sStep 1/2:%s Using provided OTP (skipping request step)\n\n", colorBlue, colorReset)
	}

	// Step 2: Register public key with OTP.
	fmt.Printf("  %sStep 2/2:%s Sending NHP-REG to server...\n", colorBlue, colorReset)
	rakMsg, regErr := a.RegisterPublicKey(otpCode, target)

	fmt.Println()
	fmt.Println(colorGreen + "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━" + colorReset)

	if regErr != nil {
		errCode := ""
		errMsgStr := regErr.Error()
		if rakMsg != nil {
			errCode = rakMsg.ErrCode
			if rakMsg.ErrMsg != "" {
				errMsgStr = rakMsg.ErrMsg
			}
		}
		fmt.Println()
		fmt.Printf("  %s❌ Registration failed%s\n", colorYellow, colorReset)
		if errCode != "" {
			fmt.Printf("  %sError code:%s  %s\n", colorYellow, colorReset, errCode)
		}
		fmt.Printf("  %sError:%s        %s\n", colorYellow, colorReset, errMsgStr)
		fmt.Println()
		// Return the error instead of os.Exit so the deferred a.Stop()
		// (and any other cleanup) still runs. main() maps a non-nil error
		// to a non-zero exit code.
		if errMsgStr != "" {
			return fmt.Errorf("registration failed: %s", errMsgStr)
		}
		return fmt.Errorf("registration failed")
	}

	// Registration succeeded — print results.
	pubKey := a.PublicKeyBase64ByCipherScheme()
	privKey := a.PrivateKeyBase64()
	fmt.Println()
	fmt.Printf("  %s✅ Registration successful!%s\n", colorGreen, colorReset)
	fmt.Println()
	fmt.Printf("  %sEmail (UserId):%s   %s\n", colorYellow, colorReset, email)
	fmt.Printf("  %sAuthServiceId:%s    %s\n", colorYellow, colorReset, aspId)
	if orgId != "" {
		fmt.Printf("  %sOrganization:%s     %s\n", colorYellow, colorReset, orgId)
	}
	if deviceId != "" {
		fmt.Printf("  %sDevice ID:%s        %s\n", colorYellow, colorReset, deviceId)
	}
	fmt.Printf("  %sCipher scheme:%s    %s\n", colorYellow, colorReset, func() string {
		if cipherScheme == common.CIPHER_SCHEME_GMSM {
			return "GMSM / SM2"
		}
		return "Curve25519"
	}())
	fmt.Printf("  %sPrivate key:%s      %s%s%s\n", colorYellow, colorReset, colorDim, privKey, colorReset)
	fmt.Printf("  %sPublic key:%s       %s\n", colorYellow, colorReset, pubKey)
	if rakMsg != nil && rakMsg.ExpiresAt != nil {
		expTime := time.Unix(*rakMsg.ExpiresAt, 0)
		fmt.Printf("  %sKey expires at:%s   %s  %s(after this date the server will reject knock requests)%s\n",
			colorYellow, colorReset, expTime.Format("2006-01-02 15:04:05 MST"), colorDim, colorReset)
	}
	fmt.Println()

	// Offer to write config.toml + resource.toml. These writers emit a
	// minimal template, so overwriting an already-configured agent would
	// discard fields the register flow doesn't know about (UserData,
	// TEEPrivateKeyBase64, custom LogLevel, and every additional
	// [[Resources]] entry). Detect existing files: warn, list what would
	// be overwritten, and DEFAULT THE PROMPT TO NO so pressing Enter does
	// not silently destroy config. Existing files are backed up to
	// <name>.bak before being replaced.
	cfgPath := filepath.Join(exeDirPath, "etc", "config.toml")
	resPath := filepath.Join(exeDirPath, "etc", "resource.toml")
	var existing []string
	if _, err := os.Stat(cfgPath); err == nil {
		existing = append(existing, "config.toml")
	}
	if _, err := os.Stat(resPath); err == nil {
		existing = append(existing, "resource.toml")
	}

	defaultYes := len(existing) == 0
	if defaultYes {
		fmt.Printf("  %sWrite config.toml and resource.toml with these settings? [Y/n]:%s ", colorBold, colorReset)
	} else {
		fmt.Printf("  %s⚠  Existing %s found — overwriting keeps only the fields shown above%s\n",
			colorYellow, strings.Join(existing, " and "), colorReset)
		fmt.Printf("  %s(a .bak copy is kept). Overwrite? [y/N]:%s ", colorDim, colorReset)
	}
	choice, _ := reader.ReadString('\n')
	choice = strings.TrimSpace(strings.ToLower(choice))

	confirmed := (choice == "y" || choice == "yes") || (defaultYes && choice == "")
	if confirmed {
		if err := writeRegistrationConfig(exeDirPath, privKey, email, orgId, cipherScheme); err != nil {
			fmt.Printf("  %s⚠  Failed to write config.toml:%s %v\n\n", colorYellow, colorReset, err)
		} else {
			fmt.Printf("  %s✔  config.toml written to:%s %s\n\n", colorGreen, colorReset, cfgPath)
		}
		if err := writeResourceConfig(exeDirPath, aspId, resId, serverCluster); err != nil {
			fmt.Printf("  %s⚠  Failed to write resource.toml:%s %v\n\n", colorYellow, colorReset, err)
		} else {
			fmt.Printf("  %s✔  resource.toml written to:%s %s\n\n", colorGreen, colorReset, resPath)
		}
	} else if len(existing) > 0 {
		fmt.Printf("  %sKept existing config unchanged. Your new private key and public key are shown above.%s\n\n",
			colorDim, colorReset)
	}

	return nil
}

// backupIfExists copies path to path+".bak" when path already exists, so a
// subsequent truncating write does not irreversibly destroy the operator's
// prior config. A failure to back up aborts the write (returned error).
func backupIfExists(path string) error {
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}
	return os.WriteFile(path+".bak", data, 0o600)
}

// writeResourceConfig writes a minimal resource.toml binding the registered asp-id/res-id to a cluster.
func writeResourceConfig(exeDirPath, aspId, resId, cluster string) error {
	etcDir := filepath.Join(exeDirPath, "etc")
	if err := os.MkdirAll(etcDir, 0o755); err != nil {
		return err
	}
	resPath := filepath.Join(etcDir, "resource.toml")

	content := fmt.Sprintf(`# NHP-Agent resource config — generated by nhp-agentd register

[[Resources]]
AuthServiceId = %q
ResourceId    = %q
Cluster       = %q
`, aspId, resId, cluster)

	if err := backupIfExists(resPath); err != nil {
		return err
	}
	return os.WriteFile(resPath, []byte(content), 0o600)
}

// writeRegistrationConfig writes a minimal config.toml using the registered identity.
func writeRegistrationConfig(exeDirPath, privKey, userId, orgId string, cipherScheme int) error {
	etcDir := filepath.Join(exeDirPath, "etc")
	if err := os.MkdirAll(etcDir, 0o755); err != nil {
		return err
	}
	cfgPath := filepath.Join(etcDir, "config.toml")

	content := fmt.Sprintf(`# NHP-Agent config — generated by nhp-agentd register
# DefaultCipherScheme: 0 = curve25519, 1 = gmsm
PrivateKeyBase64 = %q
DefaultCipherScheme = %d
UserId = %q
OrganizationId = %q
LogLevel = 2
`, privKey, cipherScheme, userId, orgId)

	if err := backupIfExists(cfgPath); err != nil {
		return err
	}
	return os.WriteFile(cfgPath, []byte(content), 0o600)
}
