package test

import (
	"fmt"
	"os"
	"strings"
	"testing"

	log "github.com/OpenNHP/opennhp/nhp/log"
	utils "github.com/OpenNHP/opennhp/nhp/utils"
)

func TestUUID(t *testing.T) {
	uuid, err := utils.NewUUID()
	if err != nil {
		fmt.Println("error: ", err)
		return
	}

	fmt.Println("uuid: ", uuid)
}

func TestGenerateUUIDv4(t *testing.T) {
	uuid, err := utils.GenerateUUIDv4()
	if err != nil {
		fmt.Println("error: ", err)
		return
	}

	fmt.Println("uuid: ", uuid)
}

func TestIPTables(t *testing.T) {
	iptables, err := utils.NewIPTables()

	if err != nil {
		fmt.Printf("error: %v\n", err)
	}

	fmt.Printf("iptables: %+v", iptables)
}

func TestPanicCatch(t *testing.T) {
	tlog := log.NewLogger("NHP-LogTest", log.LogLevelDebug, "", "logtest")
	log.SetGlobalLogger(tlog)

	func() {
		defer func() {
			fmt.Println("defer function returns 0")
		}()

		defer utils.CatchPanicThenRun(func() {
			fmt.Println("panic caught###")
		})

		defer func() {
			fmt.Println("defer function returns 1")
		}()

		fmt.Println("function starts")

		panic(1)
	}()
}

func TestUpdateTomlConfig(t *testing.T) {
	tempFile, err := os.CreateTemp("", "config-*.toml")
	if err != nil {
		t.Fatalf("can't create temporary file: %v", err)
	}
	defer os.Remove(tempFile.Name())

	initialContent := `# NHP-Agent base config
# field with (-) does not support dynamic update

# PrivateKeyBase64 (-): agent private key in base64 format.
# TEEPrivateKeyBase64 (-): TEE private key in base64 format.
# DefaultCipherScheme: 0: curve25519, 1: gmsm.
# UserId: specify the user id this agent represents.
# OrganizationId: specify the organization id this agent represents.
# LogLevel: 0: silent, 1: error, 2: info, 3: audit, 4: debug, 5: trace.
PrivateKeyBase64 = "lDaE1EKKyIJG4A28IZup/GDBZWYWEPZqGFaoV4Rlnn0="
DefaultCipherScheme = 0
UserId = "agent-0"
OrganizationId = "opennhp.org"
LogLevel = 4
# UserData: a customized user entry for flexibility.
# Its key-value pairs will be send to server along with knock message.
[UserData]
"ExampleKey0" = "StringValue"
"ExampleKey1" = 1
"ExampleKey2" = true
`
	_, writeErr := tempFile.WriteString(initialContent)
	if writeErr != nil {
		t.Fatalf("can't write into temporary file: %v", writeErr)
	}
	tempFile.Close()

	updateErr := utils.UpdateTomlConfig(tempFile.Name(), "PrivateKeyBase64", "+Jnee2lP6Kn47qzSaqwSmWxORsBkkCV6YHsRqXM23Vo=")
	if updateErr != nil {
		t.Fatalf("can't update toml config: %v", updateErr)
	}

	content, err := os.ReadFile(tempFile.Name())
	if err != nil {
		t.Fatalf("can't read temporary file: %v", err)
	}

	if !strings.Contains(string(content), "PrivateKeyBase64 = \"+Jnee2lP6Kn47qzSaqwSmWxORsBkkCV6YHsRqXM23Vo=\"") {
		t.Fatalf("can't find updated value in temporary file")
	}
}
