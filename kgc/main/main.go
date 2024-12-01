package main

import (
	"fmt"
	"os"
	"regexp"
	"strings"
	"github.com/OpenNHP/opennhp/kgc"
	"github.com/OpenNHP/opennhp/kgc/user"
	"github.com/urfave/cli/v2"
)

func main() {

	app := cli.NewApp()
	app.Name = "kgc"
	app.Usage = "kgc command-line tool for CL-PKC"
	app.Version = "1.0.0"

	keygenKgcCmd := &cli.Command{
		Name:  "keygenkgc",
		Usage: "Generate partial private key and declared public key",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:     "userid",
				Aliases:  []string{"u"},
				Usage:    "User ID for generating partial private key and declared public key",
				Required: false, 
			},
		},
		Action: func(c *cli.Context) error {
			userID := c.String("userid") // Get userid from command line argument
    	if userID == "" {            // If not provided on the command line, calls getUserID()
        var err error
        userID, err = getUserID()
        if err != nil {
            return fmt.Errorf("failed to get user ID: %v", err)
        }
    	}
			return generateKgcKey(userID)
		},
	}

	// Define keygen command
	keygenCmd := &cli.Command{
		Name:  "keygen",
		Usage: "Generate complete key pair for a user using user ID",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name: "userid", 
				Aliases: []string{"u"}, 
				Usage: "User ID for key generation", 
				Required: false},
		},
		Action: func(c *cli.Context) error {
			userID := c.String("userid") 
    	if userID == "" {            
        var err error
        userID, err = getUserID()
        if err != nil {
            return fmt.Errorf("failed to get user ID: %v", err)
        }
    	}
			return generateUserKey(userID)
		},
	}

	// Define sign command
	signCmd := &cli.Command{
		Name:  "sign",
		Usage: "Sign a plaintext message",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name: "message", 
				Aliases: []string{"m"}, 
				Usage: "Plaintext message to sign", 
				Required: false},
		},
		Action: func(c *cli.Context) error {
			userID := c.String("userid") 
    	if userID == "" {            
        var err error
        userID, err = getUserID()
        if err != nil {
            return fmt.Errorf("failed to get user ID: %v", err)
        }
    	}
			return signMessage(userID)
		},
	}

	// Define verifysign command
	verifySignCmd := &cli.Command{
		Name:  "verifysign",
		Usage: "Verify a signed message",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name: "signature", 
				Aliases: []string{"s"}, 
				Usage: "Signed message to verify", 
				Required: false},
		},
		Action: func(c *cli.Context) error {
			userID, err := getUserID()
			if err != nil {
				return fmt.Errorf("failed to get user ID: %v", err)
			}
			return verifySignature(userID)
		},
	}

	// Register all commands
	app.Commands = []*cli.Command{
		keygenKgcCmd,
		keygenCmd,
		signCmd,
		verifySignCmd,
	}

	// Run application
	if err := app.Run(os.Args); err != nil {
		fmt.Fprintln(os.Stderr, err)
	}

	//Verify that the public key is valid
	/* user.ComputePAPrime(user.DA)
	fmt.Printf("PAX_': %X\n", user.PAx_)
	fmt.Printf("PAY_': %X\n", user.PAy_)
	 if user.PAx.Cmp(user.PAx_) != 0 {
		fmt.Printf("Verification of PAX failed\n")
	} else {
		fmt.Printf("Verify PAX success\n")
	}
	if user.PAy.Cmp(user.PAy_) != 0 {
		fmt.Printf("Verification of PAY failed\n")
	} else {
		fmt.Printf("Verify PAY success\n")
	} */
}

func getUserID() (string, error) {
	var userID string
	for {
		fmt.Print("Please enter user ID (e.g., email address or phone number): ")
		fmt.Scanln(&userID)
		if strings.Contains(userID, "@") {
			if !isValidEmail(userID) {
				fmt.Println("Invalid email address. Try again.")
				continue
			}
		} else {
			if !isValidPhoneNumber(userID) {
				fmt.Println("Invalid phone number. Try again.")
				continue
			}
		}
		fmt.Println("Valid user ID.")
		break
	}
	return userID, nil
}

// isValidEmail is used to verify whether the email address is valid
func isValidEmail(email string) bool {
	regex := `^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`
	re, err := regexp.Compile(regex)
	if err != nil {
		fmt.Println("Regular compilation error:", err)
		return false
	}
	return re.MatchString(email)
}

// isValidPhoneNumber validates if the input is a valid phone number using regex
func isValidPhoneNumber(phone string) bool {
	// This regex matches phone numbers like:
	// 1234567890 (10 digits)
	// +1-234-567-8901 (+1 followed by a 10-digit number)
	regex := `^(?:\+?\d{1,3})?[-.\s]?\(?\d{1,4}?\)?[-.\s]?\d{1,4}[-.\s]?\d{1,4}[-.\s]?\d{1,4}$`
	re, err := regexp.Compile(regex)
	if err != nil {
		fmt.Println("Regex compilation error:", err)
		return false
	}
	return re.MatchString(phone)
}

// Command implementation logic
func generateKgcKey(userID string) error {
	fmt.Println("Generating partial private key and declared public key...")
	user.ProcessUserEmail(userID)  
	user.GenerateUserKeyPairSM2()
	kgc.GenerateMasterKeyPairSM2()
	kgc.GenerateWA(user.UAX, user.UAY)
	kgc.CalculateHA(kgc.EntlA,kgc.IdA,kgc.A,kgc.B,kgc.Gx,kgc.Gy,kgc.PpubX,kgc.PpubY)
	kgc.ComputeL(kgc.WAx,kgc.WAy,kgc.HA,kgc.N)
	kgc.ComputeTA(kgc.W,kgc.L,kgc.Ms,kgc.N)
	fmt.Printf("User partial private key: %X\n", kgc.TA)
	fmt.Printf("User’s declares public key x coordinate: %X\nUser’s declares public key y coordinate: %X\n",kgc.WAx,kgc.WAy )
	return nil
}

func generateUserKey(userID string) error {
	fmt.Printf("Generating complete key pair for user ID: %s...\n", userID)
	user.ProcessUserEmail(userID)  
	user.GenerateUserKeyPairSM2()
	kgc.GenerateMasterKeyPairSM2()
	kgc.GenerateWA(user.UAX, user.UAY)
	kgc.CalculateHA(kgc.EntlA,kgc.IdA,kgc.A,kgc.B,kgc.Gx,kgc.Gy,kgc.PpubX,kgc.PpubY)
	kgc.ComputeL(kgc.WAx,kgc.WAy,kgc.HA,kgc.N)
	kgc.ComputeTA(kgc.W,kgc.L,kgc.Ms,kgc.N)
	user.ComputeDA(kgc.TA,user.DA_,kgc.N)
	user.ComputePA(kgc.WAx,kgc.WAy,kgc.PpubX,kgc.PpubY,kgc.L)
	fmt.Printf("User private key: %X\n", user.DA)
	fmt.Printf("User’s actual public key x coordinate: %X\nUser’s actual public key y coordinate: %X\n", user.PAx, user.PAy)
	return nil
}

func signMessage(userID string) error {
	user.ProcessUserEmail(userID)  
	user.GenerateUserKeyPairSM2()
	kgc.GenerateMasterKeyPairSM2()
	kgc.GenerateWA(user.UAX, user.UAY)
	kgc.CalculateHA(kgc.EntlA,kgc.IdA,kgc.A,kgc.B,kgc.Gx,kgc.Gy,kgc.PpubX,kgc.PpubY)
	kgc.ComputeL(kgc.WAx,kgc.WAy,kgc.HA,kgc.N)
	kgc.ComputeTA(kgc.W,kgc.L,kgc.Ms,kgc.N)
	user.ComputeDA(kgc.TA,user.DA_,kgc.N)
	user.SignMessageAndEncrypt(user.DA)
	fmt.Println("Message successfully signed!")
	fmt.Printf("Signature:\n")
	fmt.Printf("r: %X\ns: %X\n", user.R, user.S)
	return nil
}

func verifySignature(userID string) error {
	user.ProcessUserEmail(userID)  
	user.GenerateUserKeyPairSM2()
	kgc.GenerateMasterKeyPairSM2()
	kgc.GenerateWA(user.UAX, user.UAY)
	kgc.CalculateHA(kgc.EntlA,kgc.IdA,kgc.A,kgc.B,kgc.Gx,kgc.Gy,kgc.PpubX,kgc.PpubY)
	kgc.ComputeL(kgc.WAx,kgc.WAy,kgc.HA,kgc.N)
	kgc.ComputeTA(kgc.W,kgc.L,kgc.Ms,kgc.N)
	user.ComputeDA(kgc.TA,user.DA_,kgc.N)
	user.ComputePA(kgc.WAx,kgc.WAy,kgc.PpubX,kgc.PpubY,kgc.L)
	user.SignMessageAndEncrypt(user.DA)
	user.VerifySignature(user.R,user.S,user.PAx,user.PAy,[32]byte(user.MessageHash))
	return nil
}