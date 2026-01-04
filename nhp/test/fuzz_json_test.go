package test

import (
	"encoding/json"
	"testing"

	"github.com/OpenNHP/opennhp/nhp/common"
)

// FuzzAgentKnockMsg tests JSON parsing of AgentKnockMsg.
// This is a security-critical message type used for authentication.
func FuzzAgentKnockMsg(f *testing.F) {
	// Seed corpus with valid JSON structures
	f.Add([]byte(`{"userId":"test","deviceId":"dev1"}`))
	f.Add([]byte(`{"userId":"","deviceId":""}`))
	f.Add([]byte(`{}`))
	f.Add([]byte(`{"nested":{"deep":"value"}}`))
	f.Add([]byte(`[]`))
	f.Add([]byte(`null`))
	f.Add([]byte(``))

	f.Fuzz(func(t *testing.T, data []byte) {
		var msg common.AgentKnockMsg
		// Should not panic on any JSON input
		_ = json.Unmarshal(data, &msg)
	})
}

// FuzzServerKnockAckMsg tests JSON parsing of server acknowledgment messages.
func FuzzServerKnockAckMsg(f *testing.F) {
	f.Add([]byte(`{"errCode":0,"errMsg":"success"}`))
	f.Add([]byte(`{"errCode":-1,"errMsg":"error"}`))
	f.Add([]byte(`{}`))

	f.Fuzz(func(t *testing.T, data []byte) {
		var msg common.ServerKnockAckMsg
		_ = json.Unmarshal(data, &msg)
	})
}

// FuzzACOpsResultMsg tests JSON parsing of AC operation result messages.
func FuzzACOpsResultMsg(f *testing.F) {
	f.Add([]byte(`{"errCode":0,"preAccessAction":{}}`))
	f.Add([]byte(`{}`))

	f.Fuzz(func(t *testing.T, data []byte) {
		var msg common.ACOpsResultMsg
		_ = json.Unmarshal(data, &msg)
	})
}

// FuzzDARMsg tests JSON parsing of Data Access Request messages.
// Security-critical for DHP (Data Hiding Protocol).
func FuzzDARMsg(f *testing.F) {
	f.Add([]byte(`{"doId":"test-id","dbId":"db1"}`))
	f.Add([]byte(`{}`))

	f.Fuzz(func(t *testing.T, data []byte) {
		var msg common.DARMsg
		_ = json.Unmarshal(data, &msg)
	})
}

// FuzzServerCookieMsg tests JSON parsing of cookie messages.
// Cookie handling is security-critical for session management.
func FuzzServerCookieMsg(f *testing.F) {
	f.Add([]byte(`{"cookie":"abc123","expireTime":1234567890}`))
	f.Add([]byte(`{"cookie":""}`))
	f.Add([]byte(`{}`))

	f.Fuzz(func(t *testing.T, data []byte) {
		var msg common.ServerCookieMsg
		_ = json.Unmarshal(data, &msg)
	})
}

// FuzzAgentOTPMsg tests JSON parsing of OTP (One-Time Password) messages.
// Contains authentication secrets.
func FuzzAgentOTPMsg(f *testing.F) {
	f.Add([]byte(`{"userId":"user1","organizationId":"org1","authServiceId":"auth1"}`))
	f.Add([]byte(`{}`))

	f.Fuzz(func(t *testing.T, data []byte) {
		var msg common.AgentOTPMsg
		_ = json.Unmarshal(data, &msg)
	})
}

// FuzzAgentRegisterMsg tests JSON parsing of registration messages.
// Registration is security-critical for identity validation.
func FuzzAgentRegisterMsg(f *testing.F) {
	f.Add([]byte(`{"userId":"user1","deviceId":"dev1","organizationId":"org1"}`))
	f.Add([]byte(`{}`))

	f.Fuzz(func(t *testing.T, data []byte) {
		var msg common.AgentRegisterMsg
		_ = json.Unmarshal(data, &msg)
	})
}

// FuzzServerRegisterAckMsg tests JSON parsing of registration acknowledgment.
func FuzzServerRegisterAckMsg(f *testing.F) {
	f.Add([]byte(`{"errCode":0,"errMsg":"success"}`))
	f.Add([]byte(`{}`))

	f.Fuzz(func(t *testing.T, data []byte) {
		var msg common.ServerRegisterAckMsg
		_ = json.Unmarshal(data, &msg)
	})
}

// FuzzServerACOpsMsg tests JSON parsing of AC operations messages.
// Contains complex nested NetAddress arrays.
func FuzzServerACOpsMsg(f *testing.F) {
	f.Add([]byte(`{"userId":"user1","deviceId":"dev1","srcAddrs":[{"ip":"192.168.1.1","port":8080}]}`))
	f.Add([]byte(`{"srcAddrs":[],"dstAddrs":[]}`))
	f.Add([]byte(`{}`))

	f.Fuzz(func(t *testing.T, data []byte) {
		var msg common.ServerACOpsMsg
		_ = json.Unmarshal(data, &msg)
	})
}

// FuzzACOnlineMsg tests JSON parsing of AC online status messages.
func FuzzACOnlineMsg(f *testing.F) {
	f.Add([]byte(`{"acId":"ac1"}`))
	f.Add([]byte(`{}`))

	f.Fuzz(func(t *testing.T, data []byte) {
		var msg common.ACOnlineMsg
		_ = json.Unmarshal(data, &msg)
	})
}

// FuzzServerACAckMsg tests JSON parsing of server AC acknowledgment.
func FuzzServerACAckMsg(f *testing.F) {
	f.Add([]byte(`{"errCode":0}`))
	f.Add([]byte(`{}`))

	f.Fuzz(func(t *testing.T, data []byte) {
		var msg common.ServerACAckMsg
		_ = json.Unmarshal(data, &msg)
	})
}

// FuzzDRGMsg tests JSON parsing of data object registration messages.
// DHP critical - registers data objects with the server.
func FuzzDRGMsg(f *testing.F) {
	f.Add([]byte(`{"doType":1,"doId":"do1","dbId":"db1","dataSourceType":"online"}`))
	f.Add([]byte(`{}`))

	f.Fuzz(func(t *testing.T, data []byte) {
		var msg common.DRGMsg
		_ = json.Unmarshal(data, &msg)
	})
}

// FuzzDAKMsg tests JSON parsing of data registration acknowledgment.
func FuzzDAKMsg(f *testing.F) {
	f.Add([]byte(`{"errCode":0,"errMsg":"success"}`))
	f.Add([]byte(`{}`))

	f.Fuzz(func(t *testing.T, data []byte) {
		var msg common.DAKMsg
		_ = json.Unmarshal(data, &msg)
	})
}

// FuzzDAGMsg tests JSON parsing of data access grant messages.
// Contains complex nested KeyAccessObject with base64 key material.
func FuzzDAGMsg(f *testing.F) {
	f.Add([]byte(`{"errCode":0,"doId":"do1","dbId":"db1","kao":{"wrappedKey":"base64data"}}`))
	f.Add([]byte(`{}`))

	f.Fuzz(func(t *testing.T, data []byte) {
		var msg common.DAGMsg
		_ = json.Unmarshal(data, &msg)
	})
}

// FuzzDWRMsg tests JSON parsing of data wrapping request messages.
func FuzzDWRMsg(f *testing.F) {
	f.Add([]byte(`{"doId":"do1","consumerPbk":"base64pubkey"}`))
	f.Add([]byte(`{}`))

	f.Fuzz(func(t *testing.T, data []byte) {
		var msg common.DWRMsg
		_ = json.Unmarshal(data, &msg)
	})
}

// FuzzDWAMsg tests JSON parsing of data wrapping acknowledgment.
func FuzzDWAMsg(f *testing.F) {
	f.Add([]byte(`{"errCode":0,"wrappedKey":"base64data"}`))
	f.Add([]byte(`{}`))

	f.Fuzz(func(t *testing.T, data []byte) {
		var msg common.DWAMsg
		_ = json.Unmarshal(data, &msg)
	})
}

// FuzzDSAMsg tests JSON parsing of self-attestation request messages.
func FuzzDSAMsg(f *testing.F) {
	f.Add([]byte(`{"nonce":"randomnonce","smartPolicy":{"policy":"base64wasm"}}`))
	f.Add([]byte(`{}`))

	f.Fuzz(func(t *testing.T, data []byte) {
		var msg common.DSAMsg
		_ = json.Unmarshal(data, &msg)
	})
}

// FuzzDAVMsg tests JSON parsing of attestation verification messages.
// Handles base64 evidence data.
func FuzzDAVMsg(f *testing.F) {
	f.Add([]byte(`{"doId":"do1","evidence":"base64evidence","claims":"base64claims"}`))
	f.Add([]byte(`{}`))

	f.Fuzz(func(t *testing.T, data []byte) {
		var msg common.DAVMsg
		_ = json.Unmarshal(data, &msg)
	})
}

// FuzzDBOnlineMsg tests JSON parsing of database online messages.
func FuzzDBOnlineMsg(f *testing.F) {
	f.Add([]byte(`{"dbId":"db1"}`))
	f.Add([]byte(`{}`))

	f.Fuzz(func(t *testing.T, data []byte) {
		var msg common.DBOnlineMsg
		_ = json.Unmarshal(data, &msg)
	})
}

// FuzzDHPKnockMsg tests JSON parsing of DHP knock messages.
// Contains base64 evidence for data access.
func FuzzDHPKnockMsg(f *testing.F) {
	f.Add([]byte(`{"doId":"do1","evidence":"base64evidence"}`))
	f.Add([]byte(`{}`))

	f.Fuzz(func(t *testing.T, data []byte) {
		var msg common.DHPKnockMsg
		_ = json.Unmarshal(data, &msg)
	})
}

// FuzzAgentListMsg tests JSON parsing of agent list request messages.
func FuzzAgentListMsg(f *testing.F) {
	f.Add([]byte(`{"userId":"user1","organizationId":"org1"}`))
	f.Add([]byte(`{}`))

	f.Fuzz(func(t *testing.T, data []byte) {
		var msg common.AgentListMsg
		_ = json.Unmarshal(data, &msg)
	})
}

// FuzzServerListResultMsg tests JSON parsing of server list result messages.
// Contains maps of resources.
func FuzzServerListResultMsg(f *testing.F) {
	f.Add([]byte(`{"errCode":0,"resources":{"app1":{"name":"App1"}}}`))
	f.Add([]byte(`{}`))

	f.Fuzz(func(t *testing.T, data []byte) {
		var msg common.ServerListResultMsg
		_ = json.Unmarshal(data, &msg)
	})
}
