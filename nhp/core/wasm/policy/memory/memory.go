package memory

import (
	"unsafe"
)

// ReadBufferFromMemory returns a buffer from memory of WebAssembly VM
func ReadBufferFromMemory(bufferPosition *uint32, length uint32) []byte {
	subjectBuffer := make([]byte, length)
	// Convert the pointer to a byte slice pointer first
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
    unsafePtr := uintptr(unsafe.Pointer(bufferPtr))

    ptr := uint32(unsafePtr)
    size := uint32(len(buffer))

    return (uint64(ptr) << uint64(32)) | uint64(size)
}

func StringToPtr(s string) (uint32, uint32) {
	ptr := unsafe.Pointer(unsafe.StringData(s))
	return uint32(uintptr(ptr)), uint32(len(s))
}
