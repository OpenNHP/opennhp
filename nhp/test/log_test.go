package test

import (
	"testing"
	"time"

	log "github.com/OpenNHP/opennhp/nhp/log"
)

func TestLog(t *testing.T) {
	// init logger
	//tlog := log.NewLogger("NHP-LogTest", log.LogLevelDebug, "", "logtest")
	//log.SetGlobalLogger(tlog)

	for i := 0; i < 3; i++ {
		log.Info("Info log test")
		//log.Debug("Debug log test")
		//log.Critical("Critical log test")
		time.Sleep(5 * time.Second)
	}
	log.Close()
}
