package engine

import (
	"context"
	"log"

	"github.com/tetratelabs/wazero"
	"github.com/tetratelabs/wazero/api"
	"github.com/tetratelabs/wazero/imports/wasi_snapshot_preview1"
)

type Engine struct {
	ctx    context.Context
	r      wazero.Runtime
	mod    api.Module
	malloc api.Function
	free   api.Function
}

func NewEngine() *Engine {
	return &Engine{}
}

// LoadWasm loads and instantiates a WASM module from the given byte slice.
// It sets up the necessary host functions (including logging) and WASI environment.
// The instantiated module is stored in the Engine for later execution.
// Returns an error if the WASM module fails to instantiate.
func (e *Engine) LoadWasm(wasmBytes []byte) error {
	ctx := context.Background()

	r := wazero.NewRuntime(ctx)

	_, err := r.NewHostModuleBuilder("env").
		NewFunctionBuilder().WithFunc(logString).Export("log").
		Instantiate(ctx)
	if err != nil {
		log.Panicln(err)
	}

	// Instantiate WASI
	wasi_snapshot_preview1.MustInstantiate(ctx, r)

	// Configure the module to initialize the reactor
	config := wazero.NewModuleConfig().WithStartFunctions("_initialize")

	// Instantiate the wasm
	mod, err := r.InstantiateWithConfig(ctx, wasmBytes, config)
	if err != nil {
		return err
	}

	e.ctx = ctx
	e.r = r
	e.mod = mod
	e.malloc = mod.ExportedFunction("malloc")
	e.free = mod.ExportedFunction("free")

	return nil
}

// Close terminates the engine's resources by closing the underlying runner.
// It should be called to clean up resources when the engine is no longer needed.
func (e *Engine) Close() {
	e.r.Close(e.ctx)
}

func (e *Engine) ReadContentFromVMMemory(memPos uint32, memLen uint32) []byte {
	if bytes, ok := e.mod.Memory().Read(memPos, memLen); !ok {
		log.Panicf("error reading memory at 0x%x with length %d", memPos, memLen)
		return nil
	} else {
		return bytes
	}
}

func (e *Engine) WriteContentToVMMemory(content string) (memPos uint64, memLen uint64) {
	memLen = uint64(len(content))

	results, err := e.malloc.Call(e.ctx, memLen)
	if err != nil {
		log.Panicln(err)
	}

	memPos = results[0]

	if !e.mod.Memory().Write(uint32(memPos), []byte(content)) {
		log.Panicln("out of range of memory size")
	}

	return
}

func (e *Engine) OnAttestationCollect() (attestation string) {
	onAttestationCollectFunc := e.mod.ExportedFunction("onAttestationCollect")

	result, err := onAttestationCollectFunc.Call(e.ctx)
	if err != nil {
		log.Panicln(err)
	}

	memPos := uint32(result[0] >> 32)
	memLen := uint32(result[0])

	attestationBytes := e.ReadContentFromVMMemory(memPos, memLen)
	if attestationBytes == nil {
		log.Panicln("error reading attestation bytes")
	}

	return string(attestationBytes)
}

func (e *Engine) OnAttestationVerify(attestation string) bool {
	onAttestationVerifyFunc := e.mod.ExportedFunction("onAttestationVerify")

	attestationPos, attestationLen := e.WriteContentToVMMemory(attestation)

	result, err := onAttestationVerifyFunc.Call(e.ctx, attestationPos, attestationLen)
	if err != nil {
		log.Panicln(err)
	}

	if result[0] == 1 {
		return true
	}

	return false
}

func (e *Engine) OnDataPreprocess(metadata string, rawData string, filter string) (processedData string) {
	onDataPreprocessFunc := e.mod.ExportedFunction("onDataPreprocess")

	metadataPos, metadataLen := e.WriteContentToVMMemory(metadata)
	rawDataPos, rawDataLen := e.WriteContentToVMMemory(rawData)
	filterPos, filterLen := e.WriteContentToVMMemory(filter)

	result, err := onDataPreprocessFunc.Call(e.ctx, metadataPos, metadataLen, rawDataPos, rawDataLen, filterPos, filterLen)
	if err != nil {
		log.Panicln(err)
	}

	memPos := uint32(result[0] >> 32)
	memLen := uint32(result[0])

	processedDataBytes := e.ReadContentFromVMMemory(memPos, memLen)
	if processedDataBytes == nil {
		log.Panicln("error reading processed data bytes")
	}

	return string(processedDataBytes)
}

func (e *Engine) OnDataPostprocess(rawOutput string) (processedOutput string) {
	onDataPostprocessFunc := e.mod.ExportedFunction("onDataPostprocess")

	rawOutputPos, rawOutputLen := e.WriteContentToVMMemory(rawOutput)

	result, err := onDataPostprocessFunc.Call(e.ctx, rawOutputPos, rawOutputLen)
	if err != nil {
		log.Panicln(err)
	}

	memPos := uint32(result[0] >> 32)
	memLen := uint32(result[0])

	processedOutputBytes := e.ReadContentFromVMMemory(memPos, memLen)
	if processedOutputBytes == nil {
		log.Panicln("error reading processed output bytes")
	}

	return string(processedOutputBytes)
}
