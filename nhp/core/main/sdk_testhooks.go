package main

/*
#include "nhpdevicedef.h"
*/
import "C"

import "unsafe"

// Go-typed wrappers around the exported C ABI, for nhpdevice_test.go: test
// files cannot use cgo directly. The Makefile builds the shared library
// from main.go and nhpdevice.go only, so none of this is shipped.

// sdkResult is what a test needs from any of the C result structs.
type sdkResult struct {
	code int
	msg  string
	out  []byte
}

// cBuf returns a C copy of b (NULL for a nil b) and a function freeing it.
func cBuf(b []byte) (*C.uchar, func()) {
	if b == nil {
		return nil, func() {}
	}
	p := C.CBytes(b)
	return (*C.uchar)(p), func() { C.free(p) }
}

func goStringOrEmpty(p *C.char) string {
	if p == nil {
		return ""
	}
	return C.GoString(p)
}

func testDeviceInit(deviceType int, privateKeyBase64 string) (uintptr, sdkResult) {
	r := nhp_device_init(C.int(deviceType), privateKeyBase64)
	defer nhp_free_NhpResult(r)
	return uintptr(r.handle), sdkResult{code: int(r.errCode), msg: goStringOrEmpty(r.errMsg)}
}

func testDeviceClose(handle uintptr) { nhp_device_close(handle) }

// testInitWithCallerOwnedKey passes the key to nhp_device_init the way a C
// caller does: as a string over C memory. It then overwrites that memory,
// as a caller that frees or reuses its buffer would, and initializes again
// with the same key. It returns the second result, which must be
// ERR_NHP_DEVICE_ALREADY_CREATED if the registry kept its own copy.
func testInitWithCallerOwnedKey(deviceType int, key string) (uintptr, sdkResult) {
	n := len(key)
	cKey := C.CBytes([]byte(key))
	defer C.free(cKey)
	borrowed := unsafe.String((*byte)(cKey), n)

	handle, first := testDeviceInit(deviceType, borrowed)
	if first.code != 0 {
		return 0, first
	}
	C.memset(cKey, 'A', C.size_t(n))
	_, second := testDeviceInit(deviceType, key)
	return handle, second
}

func testEncrypt(handle uintptr, msgType int, peerPk []byte, peerPkLen int, data []byte, dataLen int) sdkResult {
	pk, freePk := cBuf(peerPk)
	defer freePk()
	d, freeD := cBuf(data)
	defer freeD()
	var params C.NhpEncryptParams
	r := nhp_device_encrypt_data(handle, C.int(msgType), pk, C.int(peerPkLen), d, C.int(dataLen), params)
	defer nhp_free_NhpEncryptResult(r)
	res := sdkResult{code: int(r.errCode), msg: goStringOrEmpty(r.errMsg)}
	if r.packet != nil {
		res.out = C.GoBytes(unsafe.Pointer(r.packet), r.packetLen)
	}
	return res
}

func testDecrypt(handle uintptr, packet []byte, packetLen int) sdkResult {
	p, freeP := cBuf(packet)
	defer freeP()
	var ctx C.NhpConnContext
	r := nhp_device_decrypt_packet(handle, p, C.int(packetLen), ctx)
	defer nhp_free_NhpDecryptResult(r)
	return sdkResult{code: int(r.errCode), msg: goStringOrEmpty(r.errMsg)}
}

func testSetOverload(handle uintptr, overload bool) sdkResult {
	r := nhp_device_set_overload(handle, overload)
	defer nhp_free_NhpResult(r)
	return sdkResult{code: int(r.errCode), msg: goStringOrEmpty(r.errMsg)}
}

func testSM4Encrypt(key []byte, keyLen int, nonce []byte, nonceLen int, plain []byte, plainLen int, aad []byte, aadLen int) sdkResult {
	k, fk := cBuf(key)
	defer fk()
	n, fn := cBuf(nonce)
	defer fn()
	p, fp := cBuf(plain)
	defer fp()
	a, fa := cBuf(aad)
	defer fa()
	r := nhp_sm4_aead_encrypt(k, C.int(keyLen), n, C.int(nonceLen), p, C.int(plainLen), a, C.int(aadLen))
	defer nhp_free_NhpEncryptResult(r)
	res := sdkResult{code: int(r.errCode), msg: goStringOrEmpty(r.errMsg)}
	if r.packet != nil {
		res.out = C.GoBytes(unsafe.Pointer(r.packet), r.packetLen)
	}
	return res
}

func testSM4Decrypt(key []byte, keyLen int, nonce []byte, nonceLen int, ciphertext []byte, cipherLen int, aad []byte, aadLen int) sdkResult {
	k, fk := cBuf(key)
	defer fk()
	n, fn := cBuf(nonce)
	defer fn()
	c, fc := cBuf(ciphertext)
	defer fc()
	a, fa := cBuf(aad)
	defer fa()
	r := nhp_sm4_aead_decrypt(k, C.int(keyLen), n, C.int(nonceLen), c, C.int(cipherLen), a, C.int(aadLen))
	defer nhp_free_NhpDecryptResult(r)
	res := sdkResult{code: int(r.errCode), msg: goStringOrEmpty(r.errMsg)}
	if r.data != nil {
		res.out = C.GoBytes(unsafe.Pointer(r.data), r.dataLen)
	}
	return res
}

// testRecoverToResult runs f inside the same recover guard the exported
// functions use and reports what the C caller would receive.
func testRecoverToResult(f func()) sdkResult {
	r := (*C.NhpResult)(C.malloc(C.sizeof_NhpResult))
	C.memset(unsafe.Pointer(r), 0, C.sizeof_NhpResult)
	defer nhp_free_NhpResult(r)
	func() {
		defer recoverToResult(&r.errCode, &r.errMsg)
		f()
	}()
	return sdkResult{code: int(r.errCode), msg: goStringOrEmpty(r.errMsg)}
}

func testFreeNil() {
	nhp_free_NhpResult(nil)
	nhp_free_NhpEncryptResult(nil)
	nhp_free_NhpDecryptResult(nil)
	nhp_free_cstring(nil)
}

const testErrRuntimePanic = int(C.ERR_NHP_SDK_RUNTIME_PANIC)
const testErrCreateDeviceFailed = int(C.ERR_NHP_CREATE_DEVICE_FAILED)
const testErrDeviceAlreadyCreated = int(C.ERR_NHP_DEVICE_ALREADY_CREATED)
const testErrIncorrectPacketSize = int(C.ERR_NHP_INCORRECT_PACKET_SIZE)
const testErrDeviceNotInitialized = int(C.ERR_NHP_DEVICE_NOT_INITIALIZED)
