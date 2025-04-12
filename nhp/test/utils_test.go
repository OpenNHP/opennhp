package test

import (
	"fmt"
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
