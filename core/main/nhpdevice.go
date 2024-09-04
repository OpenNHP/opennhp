package main

/*
#include "nhpdevicedef.h"
*/
import "C"

import (
	"encoding/base64"
	"fmt"
	"unsafe"

	"github.com/OpenNHP/opennhp/core"
)

var devices map[string]*core.Device = make(map[string]*core.Device)
var handles map[uintptr]*core.Device = make(map[uintptr]*core.Device)

// 释放NHPSDK产生的字符串缓冲区内存
//
//export nhp_free_cstring
func nhp_free_cstring(ptr *C.char) {
	C.free(unsafe.Pointer(ptr))
}

// 释放NHPSDK产生的nhpresult结构体缓冲区内存
//
//export nhp_free_NhpResult
func nhp_free_NhpResult(ptr *C.NhpResult) {
	C.free(unsafe.Pointer(ptr.errMsg))
	C.free(unsafe.Pointer(ptr))
}

// 释放NHPSDK产生的NhpEncryptResult结构体缓冲区内存
//
//export nhp_free_NhpEncryptResult
func nhp_free_NhpEncryptResult(ptr *C.NhpEncryptResult) {
	C.free(unsafe.Pointer(ptr.errMsg))
	C.free(unsafe.Pointer(ptr.packet))
	C.free(unsafe.Pointer(ptr))
}

// 释放NHPSDK产生的NhpDecryptResult结构体缓冲区内存
//
//export nhp_free_NhpDecryptResult
func nhp_free_NhpDecryptResult(ptr *C.NhpDecryptResult) {
	C.free(unsafe.Pointer(ptr.errMsg))
	C.free(unsafe.Pointer(ptr.msgId))
	C.free(unsafe.Pointer(ptr.data))
	C.free(unsafe.Pointer(ptr))
}

// 初始化nhp_device实例
// 输入:
// deviceType:此device所代表的NHP组件: 1: nhp-agent, 2: nhp-server, 3: nhp-ac
// privateKeyBase64: 私钥的base64编码
// 返回:
// NhpResult指针: errCode 0: nhp device实例初始化成功; 非0：nhp device实例初始化失败, errMsg: 错误描述
// 调用者：调用完成后释放返回数据，见nhp_free_NhpResult函数
//
//export nhp_device_init
func nhp_device_init(deviceType C.int, privateKeyBase64 string) *C.NhpResult {
	resultPtr := (*C.NhpResult)(C.malloc(C.sizeof_NhpResult))
	C.memset(unsafe.Pointer(resultPtr), 0, C.sizeof_NhpResult)

	if _, found := devices[privateKeyBase64]; found {
		resultPtr.errCode = C.ERR_NHP_DEVICE_ALREADY_CREATED
		resultPtr.errMsg = C.CString("nhp device already created with the given key")
		return resultPtr
	}

	privateKey, err := base64.StdEncoding.DecodeString(privateKeyBase64)
	if err != nil {
		resultPtr.errCode = C.ERR_NHP_CREATE_DEVICE_FAILED
		resultPtr.errMsg = C.CString("key invalid")
		return resultPtr
	}

	device := core.NewDevice(int(deviceType), privateKey, nil)

	handle := uintptr(unsafe.Pointer(device))
	devices[privateKeyBase64] = device
	handles[handle] = device
	resultPtr.handle = C.size_t(handle)
	return resultPtr
}

// 关闭nhp_device
// 输入：
// handle: NHP device句柄
//
//export nhp_device_close
func nhp_device_close(handle uintptr) {
	device, found := handles[handle]
	if found {
		delete(handles, handle)
		for k, v := range devices {
			if v == device {
				delete(devices, k)
				return
			}
		}
	}
}

// 将载荷数据（明文消息）进行NHP噪声加密并生成NHP加密报文
// 输入:
// handle: NHP device句柄
// msgType: NHP消息类型
// peerPbk: 对端peer公钥
// peerPbkLen: 对端peer公钥长度
// data: 明文载荷
// dataLen: 明文载荷长度
// params: 加密参数，见NhpEncryptParams结构体
// 返回:
// NhpEncryptResult指针，见NhpEncryptResult结构体
// 调用者：调用完成后释放返回数据，见nhp_free_NhpEncryptResult函数
//
//export nhp_device_encrypt_data
func nhp_device_encrypt_data(handle uintptr, msgType C.int, peerPbk *C.uchar, peerPbkLen C.int, data *C.uchar, dataLen C.int, params C.NhpEncryptParams) *C.NhpEncryptResult {
	resultPtr := (*C.NhpEncryptResult)(C.malloc(C.sizeof_NhpEncryptResult))
	C.memset(unsafe.Pointer(resultPtr), 0, C.sizeof_NhpEncryptResult)

	device, found := handles[handle]
	if !found {
		resultPtr.errCode = C.ERR_NHP_DEVICE_NOT_INITIALIZED
		resultPtr.errMsg = C.CString("nhp device not initialized")
		return resultPtr
	}

	md := &core.MsgData{
		HeaderType:     int(msgType),
		CipherScheme:   int(params.cipherScheme),
		PeerPk:         C.GoBytes(unsafe.Pointer(peerPbk), peerPbkLen),
		Message:        C.GoBytes(unsafe.Pointer(data), dataLen),
		Compress:       params.compress != 0,
		TransactionId:  uint64(params.assignTransactionId),
		ExternalCookie: (*[core.CookieSize]byte)(unsafe.Pointer(&params.cookie)),
	}

	mad, err := device.MsgToPacket(md)
	if err != nil {
		nhpError := err.(*core.Error)
		if nhpError != nil {
			resultPtr.errCode = C.int(nhpError.ErrorNumber())
		} else {
			resultPtr.errCode = -1
		}
		resultPtr.errMsg = C.CString(err.Error())
	} else {
		resultPtr.transactionId = C.ulonglong(mad.TransactionId)
		resultPtr.packet = (*C.uchar)(C.CBytes(mad.BasePacket.Content))
		resultPtr.packetLen = C.int(len(mad.BasePacket.Content))
	}

	return resultPtr
}

// 将载荷数据（明文消息）进行NHP噪声加密并生成NHP报文
// 输入:
// handle: NHP device句柄
// packet: 待解析的NHP加密报文
// packetLen: NHP加密报文长度
// 返回:
// NhpDecryptResult指针，见NhpDecryptResult结构体
// 调用者：调用完成后释放返回数据，见nhp_free_NhpDecryptResult函数
//
//export nhp_device_decrypt_packet
func nhp_device_decrypt_packet(handle uintptr, packet *C.uchar, packetLen C.int, context C.NhpConnContext) *C.NhpDecryptResult {
	resultPtr := (*C.NhpDecryptResult)(C.malloc(C.sizeof_NhpDecryptResult))
	C.memset(unsafe.Pointer(resultPtr), 0, C.sizeof_NhpDecryptResult)

	device, found := handles[handle]
	if !found {
		resultPtr.errCode = C.ERR_NHP_DEVICE_NOT_INITIALIZED
		resultPtr.errMsg = C.CString("nhp device not initialized")
		return resultPtr
	}

	var cookieStore *core.CookieStore
	if context.cookieStore != nil {
		cookieStore = (*core.CookieStore)(unsafe.Pointer(context.cookieStore))
	}

	var lastPeerSendTime *int64
	if context.lastPeerSendTime != nil {
		lastPeerSendTime = (*int64)(unsafe.Pointer(context.lastPeerSendTime))
	}

	var peerPbk *[core.PublicKeySizeEx]byte
	if context.peerPbk != nil {
		peerPbk = (*[core.PublicKeySizeEx]byte)(unsafe.Pointer(context.peerPbk))
	}

	pd := &core.PacketData{
		BasePacket:             &core.Packet{Content: C.GoBytes(unsafe.Pointer(packet), packetLen)},
		ConnLastRemoteSendTime: lastPeerSendTime,
		ConnCookieStore:        cookieStore,
		ConnPeerPublicKey:      peerPbk,
	}

	ppd, err := device.PacketToMsg(pd)

	if err != nil {
		nhpError := err.(*core.Error)
		if nhpError != nil {
			resultPtr.errCode = C.int(nhpError.ErrorNumber())
		} else {
			resultPtr.errCode = -1
		}
		resultPtr.errMsg = C.CString(err.Error())
	} else {
		resultPtr.msgType = C.int(ppd.HeaderType)
		if ppd.HeaderType != core.NHP_KPL {
			resultPtr.msgTransactionId = C.ulonglong(ppd.SenderTrxId)
			resultPtr.msgId = (*C.uchar)(C.CBytes(ppd.SenderIdentity))
			resultPtr.msgIdLen = C.int(len(ppd.SenderIdentity))
			resultPtr.data = (*C.uchar)(C.CBytes(ppd.BodyMessage))
			resultPtr.dataLen = C.int(len(ppd.BodyMessage))
		}
	}

	return resultPtr
}

// 设置nhp device的过载模式
// 输入:
// handle: NHP device句柄
// overload: 是否开启过载模式
// 返回:
// NhpResult指针，见NhpResult结构体
// 调用者：调用完成后释放返回数据，见nhp_free_NhpResult函数
//
//export nhp_device_set_overload
func nhp_device_set_overload(handle uintptr, overload bool) *C.NhpResult {
	resultPtr := (*C.NhpResult)(C.malloc(C.sizeof_NhpResult))
	C.memset(unsafe.Pointer(resultPtr), 0, C.sizeof_NhpResult)

	device, found := handles[handle]
	if !found {
		resultPtr.errCode = C.ERR_NHP_DEVICE_NOT_INITIALIZED
		resultPtr.errMsg = C.CString("nhp device not initialized")
		return resultPtr
	}

	device.SetOverload(overload)
	return resultPtr
}

// 进行SM4 AEAD加密
// 输入:
// key: 密钥缓冲
// keyLen: 密钥长度（取前16字节）
// nonce: 计数器计数值缓冲
// nonceLen: 计数器计数值长度（须12字节）
// plain: 明文数据
// plainLen: 明文长度
// addtionalData: 附带验证数据
// addtionalDataLen: 附带验证数据长度
// 返回:
// NhpEncryptResult指针，密文由packet与packetLen表示
// 调用者：调用完成后释放返回数据，见nhp_free_NhpEncryptResult函数
//
//export nhp_sm4_aead_encrypt
func nhp_sm4_aead_encrypt(key *C.uchar, keyLen C.int, nonce *C.uchar, nonceLen C.int, plain *C.uchar, plainLen C.int, additionalData *C.uchar, additionalDataLen C.int) *C.NhpEncryptResult {
	resultPtr := (*C.NhpEncryptResult)(C.malloc(C.sizeof_NhpEncryptResult))
	C.memset(unsafe.Pointer(resultPtr), 0, C.sizeof_NhpEncryptResult)

	var aeadKey [core.SymmetricKeySize]byte
	buf := make([]byte, plainLen+16)
	copy(aeadKey[:], C.GoBytes(unsafe.Pointer(key), keyLen))

	aead := core.AeadFromKey(core.GCM_SM4, &aeadKey)
	cipher := aead.Seal(buf[:0], C.GoBytes(unsafe.Pointer(nonce), nonceLen), C.GoBytes(unsafe.Pointer(plain), plainLen), C.GoBytes(unsafe.Pointer(additionalData), additionalDataLen))
	if cipher == nil {
		resultPtr.errCode = 1
		resultPtr.errMsg = C.CString("GCM encryption failed")
		return resultPtr
	}

	resultPtr.packet = (*C.uchar)(C.CBytes(cipher))
	resultPtr.packetLen = C.int(len(cipher))
	return resultPtr
}

// 进行SM4 AEAD解密
// 输入:
// key: 密钥缓冲
// keyLen: 密钥长度（取前16字节）
// nonce: 计数器计数值缓冲（须与加密时计数值相同）
// nonceLen: 计数器计数值长度（须12字节）
// cipher: 密文数据
// cipherLen: 密文长度
// addtionalData: 附带验证数据
// addtionalDataLen: 附带验证数据长度
// 返回:
// NhpDecryptResult指针，明文由data与dataLen表示
// 调用者：调用完成后释放返回数据，见nhp_free_NhpEncryptResult函数
//
//export nhp_sm4_aead_decrypt
func nhp_sm4_aead_decrypt(key *C.uchar, keyLen C.int, nonce *C.uchar, nonceLen C.int, cipher *C.uchar, cipherLen C.int, additionalData *C.uchar, additionalDataLen C.int) *C.NhpDecryptResult {
	resultPtr := (*C.NhpDecryptResult)(C.malloc(C.sizeof_NhpDecryptResult))
	C.memset(unsafe.Pointer(resultPtr), 0, C.sizeof_NhpDecryptResult)

	var aeadKey [core.SymmetricKeySize]byte
	buf := make([]byte, cipherLen)
	copy(aeadKey[:], C.GoBytes(unsafe.Pointer(key), keyLen))

	aead := core.AeadFromKey(core.GCM_SM4, &aeadKey)
	plain, err := aead.Open(buf[:0], C.GoBytes(unsafe.Pointer(nonce), nonceLen), C.GoBytes(unsafe.Pointer(cipher), cipherLen), C.GoBytes(unsafe.Pointer(additionalData), additionalDataLen))
	if err != nil {
		resultPtr.errCode = 1
		resultPtr.errMsg = C.CString(fmt.Sprintf("GCM decryption failed: %s", err))
		return resultPtr
	}

	resultPtr.data = (*C.uchar)(C.CBytes(plain))
	resultPtr.dataLen = C.int(len(plain))
	return resultPtr
}
