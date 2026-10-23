package agent

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net"
	"os"
	"path/filepath"
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

func TestRegistrationUserData(t *testing.T) {
	for _, input := range []string{"", "null", "{}", `{"email":"ignored","role":"member"}`} {
		t.Run(input, func(t *testing.T) {
			data, err := registrationUserData("user@example.com", input)
			if err != nil {
				t.Fatal(err)
			}
			if data["email"] != "user@example.com" {
				t.Fatalf("email missing: %#v", data)
			}
			if input == `{"email":"ignored","role":"member"}` && data["role"] != "member" {
				t.Fatal("extra data was lost")
			}
		})
	}
	for _, input := range []string{"[", "[]", `"text"`, "42"} {
		if _, err := registrationUserData("user@example.com", input); err == nil {
			t.Errorf("accepted non-object data %q", input)
		}
	}
}

// Exercise the public API through temporary-agent shutdown, then read the
// socket. Both cipher schemes must produce a real packet before success.
func TestRegistrationOtpTemporaryAgentTransmits(t *testing.T) {
	for _, cipherType := range []int{0, 1} {
		t.Run(fmt.Sprint(cipherType), func(t *testing.T) {
			receiver, err := net.ListenUDP("udp4", &net.UDPAddr{IP: net.IPv4(127, 0, 0, 1)})
			if err != nil {
				t.Fatal(err)
			}
			defer receiver.Close()
			key := make([]byte, core.PrivateKeySize)
			key[0] = 2
			server := core.NewDevice(core.NHP_SERVER, key, nil)
			publicKey := server.PublicKeyBase64()
			if cipherType == 1 {
				publicKey = server.PublicKeyExBase64()
			}
			dir := t.TempDir()
			etc := filepath.Join(dir, "etc")
			if err := os.MkdirAll(etc, 0700); err != nil {
				t.Fatal(err)
			}
			// Keep startup logging configuration fixed; changing its level after
			// Start has a separate pre-existing race when config.toml is absent.
			if err := os.WriteFile(filepath.Join(etc, "config.toml"), []byte("LogLevel = 0\n"), 0600); err != nil {
				t.Fatal(err)
			}
			config := fmt.Sprintf("[[Servers]]\nName = %q\nPubKeyBase64 = %q\n[[Servers.Instances]]\nIp = %q\nPort = %d\n", "test", publicKey, "127.0.0.1", receiver.LocalAddr().(*net.UDPAddr).Port)
			if err := os.WriteFile(filepath.Join(etc, "server.toml"), []byte(config), 0600); err != nil {
				t.Fatal(err)
			}
			result := RequestRegistrationOtp(nil, dir, log.LogLevelSilent, "", cipherType, "user@example.com", "", "org", "null", "asp", "res", "test")
			if result.ErrCode != common.ErrSuccess.ErrorCode() {
				t.Fatalf("OTP failed: %s %s", result.ErrCode, result.ErrMsg)
			}
			if result.PrivateKey == "" || result.PublicKey == "" {
				t.Fatal("generated key was not returned")
			}
			if err := receiver.SetReadDeadline(time.Now().Add(time.Second)); err != nil {
				t.Fatal(err)
			}
			var buf [core.PacketBufferSize]byte
			n, _, err := receiver.ReadFromUDP(buf[:])
			if err != nil {
				t.Fatalf("SDK reported success but no packet arrived: %v", err)
			}
			packet := &core.Packet{Buf: &buf, Content: buf[:n]}
			kind, _, err := server.RecvPrecheck(packet)
			if err != nil || kind != core.NHP_OTP {
				t.Fatalf("invalid OTP packet: type=%d error=%v", kind, err)
			}

			// An invalid peer key must report encryption failure, not queued success.
			config = fmt.Sprintf("[[Servers]]\nName = %q\nPubKeyBase64 = %q\n[[Servers.Instances]]\nIp = %q\nPort = %d\n", "test", "AQ==", "127.0.0.1", receiver.LocalAddr().(*net.UDPAddr).Port)
			// Use a new directory so no previous file watcher can observe this write.
			badDir := t.TempDir()
			if err := os.MkdirAll(filepath.Join(badDir, "etc"), 0700); err != nil {
				t.Fatal(err)
			}
			if err := os.WriteFile(filepath.Join(badDir, "etc", "config.toml"), []byte("LogLevel = 0\n"), 0600); err != nil {
				t.Fatal(err)
			}
			if err := os.WriteFile(filepath.Join(badDir, "etc", "server.toml"), []byte(config), 0600); err != nil {
				t.Fatal(err)
			}
			failed := RequestRegistrationOtp(nil, badDir, log.LogLevelSilent, "", cipherType, "user@example.com", "", "org", "null", "asp", "res", "test")
			if failed.ErrCode == common.ErrSuccess.ErrorCode() {
				t.Fatal("invalid peer key reported success")
			}
		})
	}
}
