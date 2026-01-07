package main

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"os"

	"github.com/OpenNHP/opennhp/endpoints/kgc"
	"github.com/OpenNHP/opennhp/endpoints/kgc/user"

	"github.com/OpenNHP/opennhp/nhp/version"

	"github.com/urfave/cli/v2"
)

func main() {
	app := cli.NewApp()
	app.Name = "kgc"
	app.Usage = "kgc is used to distribute the key to the user"
	app.Version = version.Version

	masterCmd := &cli.Command{
		Name:  "setup",
		Usage: "generate the system parameters and the master public and private key pair in kgc",
		Action: func(c *cli.Context) error {
			return setUp()
		},
	}

	userCmd := &cli.Command{
		Name:  "keygen",
		Usage: "generate the user private key and declared user public key.",
		Flags: []cli.Flag{
			&cli.StringFlag{Name: "user-id", Usage: "specify the user identifier that can be email address, phone number or other unique identifier", Required: true},
			&cli.BoolFlag{Name: "json", Value: false, DisableDefaultText: true, Usage: "output in JSON format"},
		},
		Before: func(c *cli.Context) error {
			if len(c.String("user-id")) > 64 {
				return fmt.Errorf("userId is too long, should be not longer than 64")
			}
			return nil
		},
		Action: func(c *cli.Context) error {
			userId := c.String("user-id")
			return GenerateUserFullKey(userId, c.Bool("json"))
		},
	}

	signCmd := &cli.Command{
		Name:  "sign",
		Usage: "sign the message with the user's private key",
		Flags: []cli.Flag{
			&cli.StringFlag{Name: "private-key", Usage: "specify private key with base64 format", Required: true},
			&cli.StringFlag{Name: "message", Usage: "specify the message to be signed", Required: true},
			&cli.BoolFlag{Name: "json", Value: false, DisableDefaultText: true, Usage: "output in JSON format"},
		},
		Action: func(c *cli.Context) error {
			privateKey := c.String("private-key")
			message := c.String("message")
			return Sign(privateKey, message, c.Bool("json"))
		},
	}

	verifyCmd := &cli.Command{
		Name:  "verify",
		Usage: "verify the signature with the user's declared public key",
		Flags: []cli.Flag{
			&cli.StringFlag{Name: "declared-public-key", Usage: "specify the declared public key with base64 format", Required: true},
			&cli.StringFlag{Name: "user-id", Usage: "specify the user identifier that can be email address, phone number or other unique identifier", Required: true},
			&cli.StringFlag{Name: "message", Usage: "specify the message to be signed", Required: true},
			&cli.StringFlag{Name: "signature", Usage: "specify the signature with base64 format", Required: true},
			&cli.BoolFlag{Name: "json", Value: false, DisableDefaultText: true, Usage: "output in JSON format"},
		},
		Action: func(c *cli.Context) error {
			declaredPbkBase64 := c.String("declared-public-key")
			userId := c.String("user-id")
			message := c.String("message")
			signatureBase64 := c.String("signature")
			return Verify(declaredPbkBase64, userId, message, signatureBase64, c.Bool("json"))
		},
	}

	app.Commands = []*cli.Command{masterCmd, userCmd, signCmd, verifyCmd}

	if err := app.Run(os.Args); err != nil {
		panic(err)
	}
}

func setUp() error {
	k := kgc.SM2.NewKGCImpl()
	return k.GenerateMasterKey()
}

// GenerateUserFullKey generates a complete key pair for a user based on the provided userId.
// It initializes a KGC instance from the configuration and creates a new user with parameters derived from the KGC.
// Then it proceeds to generate the user's partial key, KGC-assisted partial key, and finally the full key.
// The function validates the generated full key and prints the private and public keys in Base64 format.
func GenerateUserFullKey(userId string, outputJson bool) error {
	k, err := kgc.NewKGCImplFromConfig()
	if err != nil {
		if outputJson {
			json.NewEncoder(os.Stdout).Encode(map[string]interface{}{
				"error": err.Error(),
			})
			return nil
		}
		return err
	}

	user := user.NewUser(k.Curve, k.GetHash(), k.GetMasterKey())

	userPartialKey, err := user.GenerateUserPartialKey()
	if err != nil {
		if outputJson {
			json.NewEncoder(os.Stdout).Encode(map[string]interface{}{
				"error": err.Error(),
			})
			return nil
		}
		return err
	}

	kgcUserPartialKey, err := k.GenerateKGCUserPartialKey(userId, userPartialKey.PubX, userPartialKey.PubY)
	if err != nil {
		if outputJson {
			json.NewEncoder(os.Stdout).Encode(map[string]interface{}{
				"error": err.Error(),
			})
			return nil
		}
		return err
	}

	userFullKey, err := user.GenerateUserFullKey(kgcUserPartialKey, userPartialKey)
	if err != nil {
		if outputJson {
			json.NewEncoder(os.Stdout).Encode(map[string]interface{}{
				"error": err.Error(),
			})
			return nil
		}
		return err
	}

	declaredUserPubBytes := userFullKey.PubX.Bytes()
	declaredUserPubBytes = append(declaredUserPubBytes, userFullKey.PubY.Bytes()...)

	if err := user.VerifyFullKey(userFullKey, userId); err != nil {
		if outputJson {
			json.NewEncoder(os.Stdout).Encode(map[string]interface{}{
				"error": err.Error(),
			})
			return nil
		}
		return err
	}

	priv := base64.StdEncoding.EncodeToString(userFullKey.PrivateKey.Bytes())
	pub := base64.StdEncoding.EncodeToString(declaredUserPubBytes)

	if outputJson {
		json.NewEncoder(os.Stdout).Encode(map[string]string{
			"privateKey":        priv,
			"declaredPublicKey": pub,
		})
	} else {
		fmt.Println("Private key: ", priv)
		fmt.Println("Declared Public key: ", pub)
	}
	return nil
}

func Sign(privateKey string, message string, outputJson bool) error {
	k, err := kgc.NewKGCImplFromConfig()
	if err != nil {
		if outputJson {
			json.NewEncoder(os.Stdout).Encode(map[string]interface{}{
				"error": err.Error(),
			})
			return nil
		}
		return err
	}

	user := user.NewUser(k.Curve, k.GetHash(), k.GetMasterKey())

	r, s, err := user.Sign(privateKey, message)
	if err != nil {
		if outputJson {
			json.NewEncoder(os.Stdout).Encode(map[string]interface{}{
				"error": err.Error(),
			})
			return nil
		}
		return err
	}

	sig := base64.StdEncoding.EncodeToString(append(r.Bytes(), s.Bytes()...))

	if outputJson {
		json.NewEncoder(os.Stdout).Encode(map[string]string{
			"signature": sig,
		})
	} else {
		fmt.Println("base64 encoded signature: ", sig)
	}

	return nil
}

func Verify(declaredPbkBase64, userId, message, signatureBase64 string, outputJson bool) error {
	k, err := kgc.NewKGCImplFromConfig()
	if err != nil {
		if outputJson {
			json.NewEncoder(os.Stdout).Encode(map[string]interface{}{
				"error": err.Error(),
			})
			return nil
		}
		return err
	}

	user := user.NewUser(k.Curve, k.GetHash(), k.GetMasterKey())

	valid := user.Verify(declaredPbkBase64, userId, message, signatureBase64)

	if outputJson {
		json.NewEncoder(os.Stdout).Encode(map[string]bool{
			"valid": valid,
		})
	} else {
		if valid {
			fmt.Println("Signature is valid")
		} else {
			fmt.Println("Signature is invalid")
		}
	}
	return nil
}
