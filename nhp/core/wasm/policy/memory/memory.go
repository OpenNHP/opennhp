package memory

import (
	"unsafe"
)

// ReadBufferFromMemory returns a buffer from memory of WebAssembly VM
func ReadBufferFromMemory(bufferPosition *uint32, length uint32) []byte {
	subjectBuffer := make([]byte, length)
	// Convert the pointer to a byte slice pointer first
	//nolint:gosec // G103: Required for WASM memory access - pointer comes from WASM runtime
	data := unsafe.Slice((*byte)(unsafe.Pointer(bufferPosition)), length)
	for i := range subjectBuffer {
		subjectBuffer[i] = data[i]
	}
	return subjectBuffer
}

// CopyBufferToMemory returns a single value
// (a kind of pair with position and length)
func CopyBufferToMemory(buffer []byte) uint64 {
	bufferPtr := &buffer[0]
	//nolint:gosec // G103: Required for WASM memory access - returns pointer to Go memory for WASM
	unsafePtr := uintptr(unsafe.Pointer(bufferPtr))

	//nolint:gosec // G115: Pointer truncation is intentional for 32-bit WASM address space
	ptr := uint32(unsafePtr)
	//nolint:gosec // G115: Buffer length is always within uint32 range for WASM
	size := uint32(len(buffer))

	return (uint64(ptr) << uint64(32)) | uint64(size)
}

func StringToPtr(s string) (uint32, uint32) {
	//nolint:gosec // G103: Required for WASM string passing - pointer to string data for WASM
	ptr := unsafe.Pointer(unsafe.StringData(s))
	//nolint:gosec // G115: Pointer truncation is intentional for 32-bit WASM address space
	return uint32(uintptr(ptr)), uint32(len(s))
}
