//go:build wasip1 || wasm

package main

import (
	"github.com/OpenNHP/opennhp/nhp/core/wasm/policy/impl"
	"github.com/OpenNHP/opennhp/nhp/core/wasm/policy/memory"
)

var PolicyInstance = impl.NewPolicy()

// Use `tinygo build -o policy.wasm -scheduler=none --no-debug -buildmode=c-shared -target=wasi main.go` to build the wasm file

// entrypoint of policy which will be compiled to wasm, in general, don't change this file. Write the policy in policy.go.
func main() {}

//go:wasmexport onAttestationCollect
func onAttestationCollect() uint64 {
	attestationJson := PolicyInstance.OnAttestationCollect()

	posSizePair := memory.CopyBufferToMemory([]byte(attestationJson))

	return posSizePair
}

//go:wasmexport onAttestationVerify
func onAttestationVerify(attestationPosition *uint32, attestationLength uint32) bool {
	attestationJson := memory.ReadBufferFromMemory(attestationPosition, attestationLength)

	return PolicyInstance.OnAttestationVerify(string(attestationJson))
}

//go:wasmexport onDataPreprocess
func onDataPreprocess(metaDataPosition *uint32, metaDataLength uint32, rawDataPosition *uint32, rawDataLength uint32, filterPosition *uint32, filterLength uint32) uint64 {
	metaDataJson := memory.ReadBufferFromMemory(metaDataPosition, metaDataLength)
	rawDataJson := memory.ReadBufferFromMemory(rawDataPosition, rawDataLength)
	filterJson := memory.ReadBufferFromMemory(filterPosition, filterLength)

	processed_data := PolicyInstance.OnDataPreprocess(string(metaDataJson), string(rawDataJson), string(filterJson))

	return memory.CopyBufferToMemory([]byte(processed_data))
}

//go:wasmexport onDataPostprocess
func onDataPostprocess(rawDataPosition *uint32, rawDataLength uint32) uint64 {
	rawData := memory.ReadBufferFromMemory(rawDataPosition, rawDataLength)

	processedData := PolicyInstance.OnDataPostprocess(string(rawData))

	return memory.CopyBufferToMemory([]byte(processedData))
}
