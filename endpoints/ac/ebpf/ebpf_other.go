//go:build !linux

package ebpf

import (
	utilsebpf "github.com/OpenNHP/opennhp/nhp/utils/ebpf"

	"github.com/OpenNHP/opennhp/nhp/log"
)

var ErrEBPFSupportedOnlyOnLinux = utilsebpf.ErrEBPFSupportedOnlyOnLinux

var (
	DenyLogger *log.Logger
	AcLogger   *log.Logger
)

func EbpfEngineLoad(dirPath string, logLevel int, acId string) error {
	_, err := utilsebpf.EngineLoad(utilsebpf.EngineLoadParams{
		Variant:     utilsebpf.VariantAC,
		ComponentId: acId,
		LogDirPath:  dirPath,
		LogLevel:    logLevel,
	})
	return err
}

// clean eBPF map file
func CleanupBPFFiles() {
	utilsebpf.CleanupBPFFiles(utilsebpf.VariantAC)
}
