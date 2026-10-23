package agent

import (
	"encoding/base64"
	"encoding/json"
	"reflect"
	"testing"
	"time"

	"github.com/OpenNHP/opennhp/nhp/common"
	"github.com/OpenNHP/opennhp/nhp/core"
	"github.com/OpenNHP/opennhp/nhp/log"
)

func TestSDKRegistrationPreservesIdentity(t *testing.T) {
	log.SetGlobalLogger(log.NewLogger("", log.LogLevelSilent, "", ""))
	for _, operation := range []string{"otp", "register"} {
		for _, outcome := range []string{"success", "unknown-cluster", "send-error", "rejected"} {
			if operation == "otp" && outcome == "rejected" {
				continue
			}
			for _, deviceID := range []string{"registration-device", ""} {
				t.Run(operation+"/"+outcome+"/"+deviceID, func(t *testing.T) {
					key := make([]byte, core.PrivateKeySize)
					key[0] = 1
					privateKey := base64.StdEncoding.EncodeToString(key)
					original := &KnockUser{UserId: "original-user", OrganizationId: "original-org", UserData: map[string]any{"role": "original"}}
					a := &UdpAgent{
						config:    &Config{PrivateKeyBase64: privateKey, DefaultCipherScheme: common.CIPHER_SCHEME_CURVE},
						device:    core.NewDevice(core.NHP_AGENT, key, nil),
						knockUser: &KnockUser{UserId: original.UserId, OrganizationId: original.OrganizationId, UserData: original.UserData},
						deviceId:  "original-device", sendMsgCh: make(chan *core.MsgData),
						serverClusterByName: map[string]*ServerCluster{"c1": buildTwoInstanceCluster(t)},
					}
					a.running.Store(true)
					a.signals.stop = make(chan struct{})
					cluster := "c1"
					if outcome == "unknown-cluster" {
						cluster = "missing"
					}
					if outcome == "send-error" {
						close(a.signals.stop)
					}
					done := make(chan RegistrationResult, 1)
					go func() {
						if operation == "otp" {
							done <- RequestRegistrationOtp(a, "", 0, privateKey, 0, "registration@example.com", deviceID, "registration-org", `{"role":"registration"}`, "asp", "resource", cluster)
						} else {
							done <- RegisterDevice(a, "", 0, privateKey, 0, "registration@example.com", deviceID, "registration-org", `{"role":"registration"}`, "asp", "resource", cluster, "123456")
						}
					}()
					assertIdentity := func() {
						t.Helper()
						a.knockUserMutex.RLock()
						defer a.knockUserMutex.RUnlock()
						if a.deviceId != "original-device" || !reflect.DeepEqual(a.knockUser, original) {
							t.Errorf("agent identity changed: device=%q user=%+v", a.deviceId, a.knockUser)
						}
					}
					if outcome == "success" || outcome == "rejected" {
						select {
						case msg := <-a.sendMsgCh:
							var payload common.AgentRegisterMsg
							if err := json.Unmarshal(msg.Message, &payload); err != nil {
								t.Fatal(err)
							}
							if payload.UserId != "registration@example.com" || payload.DeviceId != deviceID || payload.OrganizationId != "registration-org" || payload.AuthServiceId != "asp" || payload.PublicKey != a.PublicKeyBase64ByCipherScheme() {
								t.Errorf("wrong request identity: %+v", payload)
							}
							if !reflect.DeepEqual(payload.UserData, map[string]any{"role": "registration", "email": "registration@example.com"}) {
								t.Errorf("wrong user data: %#v", payload.UserData)
							}
							assertIdentity()
							if operation == "register" {
								if payload.OTP != "123456" || msg.HeaderType != core.NHP_REG {
									t.Error("wrong registration OTP or packet type")
								}
								code := common.ErrSuccess.ErrorCode()
								if outcome == "rejected" {
									code = common.ErrInvalidInput.ErrorCode()
								}
								body, _ := json.Marshal(common.ServerRegisterAckMsg{ErrCode: code})
								msg.ResponseMsgCh <- &core.PacketParserData{HeaderType: core.NHP_RAK, BodyMessage: body}
							} else if msg.HeaderType != core.NHP_OTP {
								t.Error("wrong OTP packet type")
							}
						case <-time.After(5 * time.Second):
							t.Fatal("request was not sent")
						}
					}
					select {
					case result := <-done:
						if (result.ErrCode == common.ErrSuccess.ErrorCode()) != (outcome == "success") {
							t.Errorf("unexpected result: code=%s message=%s", result.ErrCode, result.ErrMsg)
						}
					case <-time.After(5 * time.Second):
						t.Fatal("request did not finish")
					}
					assertIdentity()
					if !a.IsRunning() || a.PrivateKeyBase64() != privateKey {
						t.Error("agent was stopped or re-keyed")
					}
				})
			}
		}
	}
}
