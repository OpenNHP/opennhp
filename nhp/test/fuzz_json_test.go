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
