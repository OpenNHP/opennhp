package main

/*
#include <stdlib.h>
*/
import "C"

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"strings"
	"unsafe"

	"github.com/OpenNHP/opennhp/agent"
	"github.com/OpenNHP/opennhp/common"
	"github.com/OpenNHP/opennhp/core"
)

var gAgentInstance *agent.UdpAgent
var gWorkingDir string
var gLogLevel int

func deepCopyCString(c_str *C.char) string {
	if c_str == nil {
		return ""
	}
	goStr := C.GoString(c_str)
	return strings.Clone(goStr)
}

// 释放NHPSDK产生的字符串缓冲区内存
//
//export nhp_free_cstring
func nhp_free_cstring(ptr *C.char) {
	C.free(unsafe.Pointer(ptr))
}

// 初始化nhp_agent实例的工作目录路径：workingdir/etc/下为需要读取的配置文件，workingdir/logs下将生成日志文件
// 输入：
// workingDir：agent工作目录路径
// logLevel: 0: silent, 1: error, 2: info, 3: debug, 4: verbose
// 返回：
// agent实例是否初始化成功
//
//export nhp_agent_init
func nhp_agent_init(workingDir *C.char, logLevel C.int) bool {
	if gAgentInstance != nil {
		return true
	}

	gAgentInstance = &agent.UdpAgent{}
	err := gAgentInstance.Start(deepCopyCString(workingDir), int(logLevel))
	if err != nil {
		return false
	}

	return true
}

// 同步停止并释放nhp_agent
//
//export nhp_agent_close
func nhp_agent_close() {
	if gAgentInstance == nil {
		return
	}

	gAgentInstance.Stop()
	gAgentInstance = nil
}

// 读取workingdir/etc下所写入的用户信息、资源信息、服务器信息等配置文件，并异步启动循环敲门子线程
// 输入：无
// 返回：
// -1：未初始化错误，>=0：调用时敲门子线程请求敲门的资源个数（敲门资源会随workingdir/etc/resource.toml中的配置改动而同步变更）
//
//export nhp_agent_knockloop_start
func nhp_agent_knockloop_start() C.int {
	if gAgentInstance == nil {
		return -1
	}

	count := gAgentInstance.StartKnockLoop()
	return C.int(count)
}

// 同步停止循环敲门子线程
//
//export nhp_agent_knockloop_stop
func nhp_agent_knockloop_stop() {
	if gAgentInstance == nil {
		return
	}

	gAgentInstance.StopKnockLoop()
}

// 设置agent所代表的用户信息
// 输入：
// userId：用户名标识（可选，但不建议留空）
// devId：设备标识（可选）
// orgId：公司或组织标识（可选）
// userData：与后端服务需要对接的额外字段（json格式字符串，可选）
// 返回：
// 用户信息是否设置成功
//
//export nhp_agent_set_knock_user
func nhp_agent_set_knock_user(userId *C.char, devId *C.char, orgId *C.char, userData *C.char) bool {
	if gAgentInstance == nil {
		return false
	}
	jsonStr := deepCopyCString(userData)
	var data map[string]any
	if len(jsonStr) > 0 {
		err := json.Unmarshal([]byte(jsonStr), &data)
		if err != nil {
			return false
		}
	}

	gAgentInstance.SetDeviceId(deepCopyCString(devId))
	gAgentInstance.SetKnockUser(deepCopyCString(userId), deepCopyCString(orgId), data)
	return true
}

// 向agent添加一个NHP服务器信息，供敲门使用（agent可以向多个NHP服务器发起不同的敲门请求）
// 输入：
// pubkey：NHP服务器公钥
// ip：NHP服务器ip地址
// host：NHP服务器域名（如果设置域名，则ip项为可选）
// port：NHP服务器工作端口号（如果设置为0，将采用默认62206端口）
// expire：NHP服务器公钥过期时间（单位：epoch秒，设为0则为永久）
// 返回：
// 服务器信息是否添加成功
//
//export nhp_agent_add_server
func nhp_agent_add_server(pubkey *C.char, ip *C.char, host *C.char, port C.int, expire int64) bool {
	if gAgentInstance == nil {
		return false
	}

	if pubkey == nil || (ip == nil && host == nil) {
		return false
	}

	serverPort := int(port)
	if serverPort == 0 {
		serverPort = 62206 // use default server listening port
	}

	serverPeer := &core.UdpPeer{
		Type:         core.NHP_SERVER,
		PubKeyBase64: deepCopyCString(pubkey),
		Ip:           deepCopyCString(ip),
		Port:         serverPort,
		Hostname:     deepCopyCString(host),
		ExpireTime:   expire,
	}
	gAgentInstance.AddServer(serverPeer)
	return true
}

// 删除agent中的NHP服务器信息
// 输入：
// pubkey：NHP服务器公钥
//
//export nhp_agent_remove_server
func nhp_agent_remove_server(pubkey *C.char) {
	if gAgentInstance == nil {
		return
	}
	if pubkey == nil {
		return
	}

	gAgentInstance.RemoveServer(deepCopyCString(pubkey))
}

// 向agent添加一个资源信息，供敲门使用（agent可以对不同资源发起敲门请求）
// 输入：
// aspId：认证服务商标识
// resId：资源标识
// serverAddr：NHP服务器ip地址或域名（管理该资源的NHP服务器）
// 返回：
// 资源器信息是否添加成功
//
//export nhp_agent_add_resource
func nhp_agent_add_resource(aspId *C.char, resId *C.char, serverAddr *C.char) bool {
	if gAgentInstance == nil {
		return false
	}

	if aspId == nil || resId == nil || serverAddr == nil {
		return false
	}

	resource := &agent.KnockResource{
		AuthServiceId: deepCopyCString(aspId),
		ResourceId:    deepCopyCString(resId),
		ServerAddr:    deepCopyCString(serverAddr),
	}
	err := gAgentInstance.AddResource(resource)
	return err == nil
}

// 删除agent中的资源信息
// 输入：
// aspId：认证服务商标识
// resId：资源标识
//
//export nhp_agent_remove_resource
func nhp_agent_remove_resource(aspId *C.char, resId *C.char) {
	if gAgentInstance == nil {
		return
	}

	if aspId == nil || resId == nil {
		return
	}

	gAgentInstance.RemoveResource(deepCopyCString(aspId), deepCopyCString(resId))
}

// agent为访问资源向资源所在服务器发起单次敲门请求
// 输入：
// aspId：认证服务商标识
// resId：资源标识
// serverAddr：NHP服务器ip地址或域名（管理该资源的NHP服务器）
// 返回：
// 服务器的回应消息（json格式字符串缓冲区指针）：
// "errCode"：错误码（字符串，"0"表示成功）
// "errMsg"：错误消息（字符串）
// "resHost"：资源服务器地址（"resHost"：{"服务器名称1":"服务器主机名1", "服务器名称2":"服务器主机名2", ...}）
// "opnTime"：开门时长（整数，单位秒）
// "aspToken"：认证服务商认证后产生的token（可选）
// "agentAddr"：NHP服务器视角下agent的ip地址
// "preActs"：与资源的预连接信息（可选）
// "redirectUrl"：http跳转链接（可选）
//
// 调用前需要调用nhp_agent_add_server，将NHP服务器公钥、地址等信息加入agent
// 调用者负责调用nhp_free_cstring释放返回的char*指针
//
//export nhp_agent_knock_resource
func nhp_agent_knock_resource(aspId *C.char, resId *C.char, serverAddr *C.char) *C.char {
	ackMsg := &common.ServerKnockAckMsg{}

	func() {
		if gAgentInstance == nil {
			ackMsg.ErrCode = common.ErrNoAgentInstance.ErrorCode()
			ackMsg.ErrMsg = common.ErrNoAgentInstance.Error()
			return
		}

		if aspId == nil || resId == nil || serverAddr == nil {
			ackMsg.ErrCode = common.ErrInvalidInput.ErrorCode()
			ackMsg.ErrMsg = common.ErrInvalidInput.Error()
			return
		}

		resource := &agent.KnockResource{
			AuthServiceId: deepCopyCString(aspId),
			ResourceId:    deepCopyCString(resId),
			ServerAddr:    deepCopyCString(serverAddr),
		}

		peer := gAgentInstance.FindServerPeerFromResource(resource)
		if peer == nil {
			ackMsg.ErrCode = common.ErrKnockServerNotFound.ErrorCode()
			ackMsg.ErrMsg = common.ErrKnockServerNotFound.Error()
			return
		}

		target := &agent.KnockTarget{
			KnockResource: *resource,
			ServerPeer:    peer,
		}

		ackMsg, _ = gAgentInstance.Knock(target)
	}()

	bytes, _ := json.Marshal(ackMsg)
	ret := C.CString(string(bytes))

	return ret
}

// agent明确告知NHP服务器退出自身到该资源的访问权限
// 输入：
// aspId：认证服务商标识
// resId：资源标识
// serverAddr：NHP服务器ip地址或域名（管理该资源的NHP服务器）
// 返回：
// 是否成功退出
// 调用前需要调用nhp_agent_add_server，将NHP服务器公钥、地址等信息加入agent
//
//export nhp_agent_exit_resource
func nhp_agent_exit_resource(aspId *C.char, resId *C.char, serverAddr *C.char) bool {
	var err error
	ackMsg := &common.ServerKnockAckMsg{}

	func() {
		if gAgentInstance == nil {
			ackMsg.ErrCode = common.ErrNoAgentInstance.ErrorCode()
			ackMsg.ErrMsg = common.ErrNoAgentInstance.Error()
			err = common.ErrNoAgentInstance
			return
		}

		if aspId == nil || resId == nil || serverAddr == nil {
			ackMsg.ErrCode = common.ErrInvalidInput.ErrorCode()
			ackMsg.ErrMsg = common.ErrInvalidInput.Error()
			err = common.ErrInvalidInput
			return
		}

		resource := &agent.KnockResource{
			AuthServiceId: deepCopyCString(aspId),
			ResourceId:    deepCopyCString(resId),
			ServerAddr:    deepCopyCString(serverAddr),
		}

		peer := gAgentInstance.FindServerPeerFromResource(resource)
		if peer == nil {
			ackMsg.ErrCode = common.ErrKnockServerNotFound.ErrorCode()
			ackMsg.ErrMsg = common.ErrKnockServerNotFound.Error()
			err = common.ErrKnockServerNotFound
			return
		}

		target := &agent.KnockTarget{
			KnockResource: *resource,
			ServerPeer:    peer,
		}

		ackMsg, err = gAgentInstance.ExitKnockRequest(target)
	}()

	return err == nil
}

// cipherType: 0-curve25519; 1-sm2
// result: "privatekey"|"publickey"
// caller is responsible to free the returned char* pointer
//
//export nhp_generate_keys
func nhp_generate_keys(cipherType C.int) *C.char {
	var e core.Ecdh
	switch core.EccTypeEnum(cipherType) {
	case core.ECC_SM2:
		e = core.NewECDH(core.ECC_SM2)
	case core.ECC_CURVE25519:
		fallthrough
	default:
		e = core.NewECDH(core.ECC_CURVE25519)
	}
	pub := e.PublicKeyBase64()
	priv := e.PrivateKeyBase64()

	res := fmt.Sprintf("%s|%s", priv, pub)
	pRes := C.CString(res)

	return pRes
}

// cipherType: 0-curve25519; 1-sm2
// privateBase64: private key in base64 format
// result: "publickey"
// caller is responsible to free the returned char* pointer
//
//export nhp_privkey_to_pubkey
func nhp_privkey_to_pubkey(cipherType C.int, privateBase64 *C.char) *C.char {
	privKey := deepCopyCString(privateBase64)
	privKeyBytes, err := base64.StdEncoding.DecodeString(privKey)
	if err != nil {
		return nil
	}

	e := core.ECDHFromKey(core.EccTypeEnum(cipherType), privKeyBytes)
	if e == nil {
		return nil
	}
	pub := e.PublicKeyBase64()
	pPub := C.CString(pub)

	return pPub
}
