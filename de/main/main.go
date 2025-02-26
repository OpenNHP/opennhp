package main

import (
	"encoding/base64"
	"fmt"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"

	"github.com/OpenNHP/opennhp/core"
	"github.com/OpenNHP/opennhp/de"
	"github.com/OpenNHP/opennhp/version"
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
			&cli.StringFlag{Name: "mode", Value: "none", Usage: "encrypt:加密;decrypt:解密"},
			&cli.StringFlag{Name: "source", Value: "sample.txt", Usage: "源文件文件"},
			&cli.StringFlag{Name: "policy", Value: "policyinfo.json", Usage: "策略文件包含数据访问者的公钥信息"},
			&cli.StringFlag{Name: "output", Value: "output.txt", Usage: "加密码后.ztdo文件保存地址"},
			&cli.StringFlag{Name: "meta", Value: "meta.json", Usage: "meta.json源文件的部分概要信息"},
			&cli.StringFlag{Name: "ztdo", Value: "", Usage: "需要解密的ztdo文件路径"},
			&cli.StringFlag{Name: "decodeKey", Value: "", Usage: "解密密钥"},
		},
		Action: func(c *cli.Context) error {
			mode := c.String("mode")
			source := c.String("source")
			policy := c.String("policy")
			output := c.String("output")
			ztdo := c.String("ztdo")
			decodeKey := c.String("decodeKey")
			return runApp(mode, source, output, policy, ztdo, decodeKey)
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

	app.Commands = []*cli.Command{
		runCmd,
		keygenCmd,
		pubkeyCmd,
	}

	if err := app.Run(os.Args); err != nil {
		fmt.Fprintln(os.Stderr, err)
	}
}

/*
*
decodeKey:数据解密密钥
decodeSavePath:解密文件存储目录路径
*/
func runApp(mode string, source string, output string, policy string, ztdo string, decodeKey string) error {
	fmt.Println("mode=" + mode)

	exeFilePath, err := os.Executable()
	if err != nil {
		return err
	}
	exeDirPath := filepath.Dir(exeFilePath)
	a := &de.UdpDevice{}
	err = a.Start(exeDirPath, 4)
	if err != nil {
		return err
	}

	// react to terminate signals
	termCh := make(chan os.Signal, 1)
	signal.Notify(termCh, syscall.SIGTERM, os.Interrupt, syscall.SIGABRT)

	if mode == "encrypt" {
		fmt.Println("policy=" + policy)
		fmt.Println("source=" + source)
		fmt.Println("output=" + output)
		outputFilePath := output
		policyFile := policy
		dhpPolicy, err := de.ReadPolicyFile(policyFile)
		if err != nil {
			fmt.Printf("读取PolicyFile失败：%s\n", err)
			return err
		}
		zoId, encodedKey := de.EncodeToZtoFile(source, outputFilePath)
		if zoId != "" {

			fmt.Printf("数据内容加密的密钥,明文:%s\n", encodedKey)
			eccKey, err := core.ECCEncryption(dhpPolicy.ConsumerPublicKey, encodedKey)
			if err != nil {
				fmt.Printf("对数据数据内容密钥进行ECC算法加密失败：%s\n", err)
				return err
			}
			fmt.Printf("数据内容加密的密钥进ECC后的结果：%s \n", eccKey)
			a.SendDHPRegister(zoId, dhpPolicy, eccKey)
		} else {
			fmt.Printf("读取源文件失败")
		}
		os.Exit(0)
	} else if mode == "decrypt" {
		fmt.Println("ztdo=" + ztdo)
		fmt.Println("decodeKey=" + decodeKey)
		fmt.Println("output=" + output)
		de.DecodeZtoFile(ztdo, decodeKey, output)
		os.Exit(0)
	}
	// block until terminated
	<-termCh
	a.Stop()
	return nil
}
