package agent

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/OpenNHP/opennhp/nhp/common"
	"github.com/OpenNHP/opennhp/nhp/core"
	"github.com/OpenNHP/opennhp/nhp/log"
)

// RegistrationResult is the OTP and device-registration outcome shared by the
// iOS and Android SDK wrappers. PrivateKey is filled for the native layer.
// Wrappers must not pass it through to untrusted callers.
type RegistrationResult struct {
	ErrCode       string `json:"errCode"`
	ErrMsg        string `json:"errMsg,omitempty"`
	PrivateKey    string `json:"privateKey,omitempty"`
	PublicKey     string `json:"publicKey,omitempty"`
	AuthServiceId string `json:"aspId,omitempty"`
	ExpiresAt     *int64 `json:"expiresAt,omitempty"`
}

// JSON encodes the result. A marshal failure becomes a fixed error object.
func (r RegistrationResult) JSON() string {
	bytes, err := json.Marshal(r)
	if err != nil {
		return `{"errCode":"1","errMsg":"failed to encode registration result"}`
	}
	return string(bytes)
}

func registrationErrorCode(err error) string {
	type coder interface{ ErrorCode() string }
	if coded, ok := err.(coder); ok {
		if code := coded.ErrorCode(); code != "" {
			return code
		}
	}
	return "1"
}

func cipherSchemeFor(cipherType int) int {
	if core.EccTypeEnum(cipherType) == core.ECC_SM2 {
		return common.CIPHER_SCHEME_GMSM
	}
	return common.CIPHER_SCHEME_CURVE
}

func configTomlExists(workingDir string) bool {
	_, err := os.Stat(filepath.Join(workingDir, "etc", "config.toml"))
	return err == nil
}

// StartAllowingMissingConfig starts an agent that may have no config.toml.
// A missing file makes Start replace the requested log level with silent, so
// the requested level is applied again in that case.
func StartAllowingMissingConfig(workingDir string, logLevel int) (*UdpAgent, error) {
	if workingDir == "" {
		return nil, fmt.Errorf("working directory is required")
	}
	missingConfig := !configTomlExists(workingDir)
	instance := &UdpAgent{}
	instance.SetAllowMissingConfig(true)
	if err := instance.Start(workingDir, logLevel); err != nil {
		return nil, err
	}
	if missingConfig {
		log.SetLevel(logLevel)
	}
	return instance, nil
}

func agentCipherScheme(a *UdpAgent) int {
	if a == nil || a.config == nil {
		return common.CIPHER_SCHEME_CURVE
	}
	return a.config.DefaultCipherScheme
}

func invalidRegistration(msg string) RegistrationResult {
	return RegistrationResult{
		ErrCode: common.ErrInvalidInput.ErrorCode(),
		ErrMsg:  msg,
	}
}

// InstallRegistrationKey installs privateKey on an idle agent. An empty
// privateKey generates a new key for cipherType (0 Curve25519, 1 SM2).
// When the key already matches, a different cipherType is still applied.
func InstallRegistrationKey(a *UdpAgent, privateKey string, cipherType int) error {
	if a == nil {
		return fmt.Errorf("agent is not started")
	}
	if privateKey != "" && a.PrivateKeyBase64() == privateKey && agentCipherScheme(a) == cipherSchemeFor(cipherType) {
		return nil
	}
	var priv []byte
	if privateKey == "" {
		ecc := core.ECC_CURVE25519
		if core.EccTypeEnum(cipherType) == core.ECC_SM2 {
			ecc = core.ECC_SM2
		}
		priv = core.NewECDH(ecc).PrivateKey()
	} else {
		decoded, err := base64.StdEncoding.DecodeString(privateKey)
		if err != nil {
			return fmt.Errorf("private key is not valid base64")
		}
		priv = decoded
	}
	return a.ReinitWithKey(priv, cipherSchemeFor(cipherType))
}

func registrationUserData(email string, userData string) (map[string]any, error) {
	data := map[string]any{}
	if userData != "" {
		if err := json.Unmarshal([]byte(userData), &data); err != nil {
			return nil, fmt.Errorf("userData is not valid JSON")
		}
	}
	// JSON null is an empty optional user-data object.
	if data == nil {
		data = map[string]any{}
	}
	if email != "" {
		data["email"] = email
	}
	return data, nil
}

func registrationTarget(a *UdpAgent, aspId string, resId string, cluster string) (*KnockTarget, error) {
	resource := &KnockResource{
		AuthServiceId: aspId,
		ResourceId:    resId,
		Cluster:       cluster,
	}
	sc, err := a.FindServerClusterFromResource(resource)
	if err != nil {
		return nil, err
	}
	if sc == nil {
		return nil, common.ErrKnockServerNotFound
	}
	return &KnockTarget{
		KnockResource: *resource,
		ServerPeer:    sc.RepresentativePeer(),
		ServerCluster: sc,
	}, nil
}

// prepareRegistration selects an agent that can send the OTP or register call.
// A running agent is used only when its key and cipher scheme already match.
// ReinitWithKey is not called on that agent. Any other case starts a
// short-lived agent, which the caller must Stop, and which must not replace
// the caller's saved instance.
func prepareRegistration(current *UdpAgent, workingDir string, logLevel int, privateKey string, cipherType int, email string, aspId string, resId string, cluster string, generateKey bool) (RegistrationResult, *UdpAgent, *KnockTarget, bool) {
	if email == "" || aspId == "" || resId == "" || cluster == "" {
		return invalidRegistration("email, aspId, resId, and cluster are required"), nil, nil, false
	}

	scheme := cipherSchemeFor(cipherType)
	var instance *UdpAgent
	owned := false
	if current != nil && current.IsRunning() {
		switch {
		case privateKey == "" && generateKey:
			return invalidRegistration("close the running agent before generating a registration key"), nil, nil, false
		case privateKey == "" && current.PrivateKeyBase64() == "":
			return invalidRegistration("privateKey is required"), nil, nil, false
		case privateKey != "" && current.PrivateKeyBase64() != privateKey:
			return invalidRegistration("close the running agent before registering a different key"), nil, nil, false
		case agentCipherScheme(current) != scheme:
			return invalidRegistration("close the running agent before changing the cipher scheme"), nil, nil, false
		default:
			instance = current
		}
	} else if privateKey == "" && !generateKey {
		return invalidRegistration("privateKey is required"), nil, nil, false
	} else {
		started, err := StartAllowingMissingConfig(workingDir, logLevel)
		if err != nil {
			return RegistrationResult{
				ErrCode: registrationErrorCode(err),
				ErrMsg:  err.Error(),
			}, nil, nil, false
		}
		instance = started
		owned = true
		if err := InstallRegistrationKey(instance, privateKey, cipherType); err != nil {
			instance.Stop()
			return RegistrationResult{
				ErrCode: registrationErrorCode(err),
				ErrMsg:  err.Error(),
			}, nil, nil, false
		}
	}

	result := RegistrationResult{
		PrivateKey: instance.PrivateKeyBase64(),
		PublicKey:  instance.PublicKeyBase64ByCipherScheme(),
	}
	target, err := registrationTarget(instance, aspId, resId, cluster)
	if err != nil {
		if owned {
			instance.Stop()
		}
		result.ErrCode = registrationErrorCode(err)
		result.ErrMsg = err.Error()
		return result, nil, nil, false
	}
	return result, instance, target, owned
}

func finishOwnedAgent(instance *UdpAgent, owned bool) {
	if owned && instance != nil {
		instance.Stop()
	}
}

// RequestRegistrationOtp sends NHP-OTP for this agent key. An empty privateKey
// generates a key only when no agent is running. A running agent is never re-keyed
// and its knock identity is unchanged.
func RequestRegistrationOtp(current *UdpAgent, workingDir string, logLevel int, privateKey string, cipherType int, email string, deviceId string, orgId string, userData string, aspId string, resId string, cluster string) RegistrationResult {
	data, err := registrationUserData(email, userData)
	if err != nil {
		return invalidRegistration(err.Error())
	}
	result, instance, target, owned := prepareRegistration(
		current, workingDir, logLevel, privateKey, cipherType, email, aspId, resId, cluster, true,
	)
	defer finishOwnedAgent(instance, owned)
	if target == nil {
		return result
	}
	if err := instance.sendOtp(target, &common.AgentOTPMsg{
		UserId: email, DeviceId: deviceId, OrganizationId: orgId,
		AuthServiceId: target.AuthServiceId, PublicKey: result.PublicKey, UserData: data,
	}, owned); err != nil {
		result.ErrCode = registrationErrorCode(err)
		result.ErrMsg = err.Error()
		return result
	}
	result.PrivateKey = instance.PrivateKeyBase64()
	result.PublicKey = instance.PublicKeyBase64ByCipherScheme()
	result.ErrCode = common.ErrSuccess.ErrorCode()
	return result
}

// RegisterDevice submits the one-time password for the key from
// RequestRegistrationOtp. privateKey must be that key. An empty privateKey is
// accepted only when current is running and already holds one. A running agent
// is never re-keyed and its knock identity is unchanged.
func RegisterDevice(current *UdpAgent, workingDir string, logLevel int, privateKey string, cipherType int, email string, deviceId string, orgId string, userData string, aspId string, resId string, cluster string, otp string) RegistrationResult {
	if otp == "" {
		return invalidRegistration("otp is required")
	}
	data, err := registrationUserData(email, userData)
	if err != nil {
		return invalidRegistration(err.Error())
	}
	result, instance, target, owned := prepareRegistration(
		current, workingDir, logLevel, privateKey, cipherType, email, aspId, resId, cluster, false,
	)
	defer finishOwnedAgent(instance, owned)
	if target == nil {
		return result
	}
	ack, err := instance.registerPublicKey(target, &common.AgentRegisterMsg{
		UserId: email, DeviceId: deviceId, OrganizationId: orgId, OTP: otp,
		AuthServiceId: target.AuthServiceId, PublicKey: result.PublicKey, UserData: data,
	})
	if ack != nil {
		result.ErrCode = ack.ErrCode
		result.ErrMsg = ack.ErrMsg
		result.AuthServiceId = ack.AuthServiceId
		result.ExpiresAt = ack.ExpiresAt
	}
	if err != nil && result.ErrCode == "" {
		result.ErrCode = registrationErrorCode(err)
		result.ErrMsg = err.Error()
	}
	if result.ErrCode == "" {
		result.ErrCode = common.ErrSuccess.ErrorCode()
	}
	result.PrivateKey = instance.PrivateKeyBase64()
	result.PublicKey = instance.PublicKeyBase64ByCipherScheme()
	return result
}
