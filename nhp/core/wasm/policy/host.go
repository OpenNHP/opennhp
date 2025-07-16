// this package is used to import host function

package policy

import (
	"runtime"

	"github.com/OpenNHP/opennhp/nhp/core/wasm/policy/memory"
)

// _log is a WebAssembly import which prints a string (linear memory offset,
// byteCount) to the console.
//
//go:wasmimport env log
func _log(ptr, size uint32)

func Log(message string) {
	ptr, size := memory.StringToPtr(message)
	_log(ptr, size)
	runtime.KeepAlive(message)
}
