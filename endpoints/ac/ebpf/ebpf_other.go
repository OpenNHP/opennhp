//go:build !linux

package ebpf

import (
	// "log"

	"fmt"

	"github.com/OpenNHP/opennhp/nhp/log"
)

var ErrEBPFSupportedOnlyOnLinux = fmt.Errorf("eBPF functionality is only supported on Linux, current platform is not Linux")
var (
	DenyLogger *log.Logger
	AcLogger   *log.Logger
)

func EbpfEngineLoad(dirPath string, logLevel int, acId string) error {
	log.Info("eBPF function must be compiled on Linux OS")
	return ErrEBPFSupportedOnlyOnLinux
}

// clean eBPF map file
func CleanupBPFFiles() {
	log.Info("ebpf func must be compile based linux os")
}
