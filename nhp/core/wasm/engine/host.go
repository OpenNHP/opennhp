// Package engine provides host functions that can be called from within a WebAssembly (WASM) virtual machine.
// These functions enable interaction between the WASM runtime and the host environment,
// such as logging operations or other system-level interactions.

package engine

import (
	"context"
	"fmt"
	"log"

	"github.com/tetratelabs/wazero/api"
)

func logString(_ context.Context, m api.Module, offset, byteCount uint32) {
	buf, ok := m.Memory().Read(offset, byteCount)
	if !ok {
		log.Panicf("Memory.Read(%d, %d) out of range", offset, byteCount)
	}
	fmt.Println(string(buf))
}
