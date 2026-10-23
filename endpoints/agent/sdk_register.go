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

// EnsureRegistrationAgent returns the running agent, or starts one that can
// tolerate a missing config.toml. A stopped agent is replaced.
func EnsureRegistrationAgent(current *UdpAgent, workingDir string, logLevel int) (*UdpAgent, error) {
	if current != nil && current.IsRunning() {
		return current, nil
	}
	return StartAllowingMissingConfig(workingDir, logLevel)
}

// InstallRegistrationKey installs privateKey on the running agent. An empty
// privateKey generates a new key for cipherType (0 Curve25519, 1 SM2).
func InstallRegistrationKey(a *UdpAgent, privateKey string, cipherType int) error {
	if a == nil {
		return fmt.Errorf("agent is not started")
	}
	if privateKey != "" && a.PrivateKeyBase64() == privateKey {
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

func applyRegistrationUser(a *UdpAgent, email string, deviceId string, orgId string, userData string) error {
	data := map[string]any{}
	if userData != "" {
		if err := json.Unmarshal([]byte(userData), &data); err != nil {
			return fmt.Errorf("userData is not valid JSON")
		}
	}
	if email != "" {
		data["email"] = email
	}
	a.SetDeviceId(deviceId)
	a.SetKnockUser(email, orgId, data)
	return nil
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

func prepareRegistration(current *UdpAgent, workingDir string, logLevel int, privateKey string, cipherType int, email string, deviceId string, orgId string, userData string, aspId string, resId string, cluster string, generateKey bool) (RegistrationResult, *UdpAgent, *KnockTarget) {
	result := RegistrationResult{}
	if email == "" || aspId == "" || resId == "" || cluster == "" {
		result.ErrCode = common.ErrInvalidInput.ErrorCode()
		result.ErrMsg = "email, aspId, resId, and cluster are required"
		return result, current, nil
	}
	instance, err := EnsureRegistrationAgent(current, workingDir, logLevel)
	if err != nil {
		result.ErrCode = registrationErrorCode(err)
		result.ErrMsg = err.Error()
		return result, nil, nil
	}
	key := privateKey
	if key == "" && !generateKey {
		key = instance.PrivateKeyBase64()
	}
	if err := InstallRegistrationKey(instance, key, cipherType); err != nil {
		result.ErrCode = registrationErrorCode(err)
		result.ErrMsg = err.Error()
		return result, instance, nil
	}
	if err := applyRegistrationUser(instance, email, deviceId, orgId, userData); err != nil {
		result.ErrCode = common.ErrInvalidInput.ErrorCode()
		result.ErrMsg = err.Error()
		return result, instance, nil
	}
	result.PrivateKey = instance.PrivateKeyBase64()
	result.PublicKey = instance.PublicKeyBase64ByCipherScheme()
	target, err := registrationTarget(instance, aspId, resId, cluster)
	if err != nil {
		result.ErrCode = registrationErrorCode(err)
		result.ErrMsg = err.Error()
		return result, instance, nil
	}
	return result, instance, target
}

// RequestRegistrationOtp sends NHP-OTP for this agent key. An empty privateKey
// generates a key when the secure store does not already have one. The returned
// agent replaces the caller's saved instance, including nil when startup fails.
func RequestRegistrationOtp(current *UdpAgent, workingDir string, logLevel int, privateKey string, cipherType int, email string, deviceId string, orgId string, userData string, aspId string, resId string, cluster string) (RegistrationResult, *UdpAgent) {
	result, instance, target := prepareRegistration(
		current, workingDir, logLevel, privateKey, cipherType, email, deviceId, orgId, userData, aspId, resId, cluster, true,
	)
	if target == nil {
		return result, instance
	}
	if err := instance.RequestOtp(target); err != nil {
		result.ErrCode = registrationErrorCode(err)
		result.ErrMsg = err.Error()
		return result, instance
	}
	result.ErrCode = common.ErrSuccess.ErrorCode()
	return result, instance
}

// RegisterDevice submits the one-time password for the key installed by
// RequestRegistrationOtp. The returned agent is the instance the SDK wrapper
// should retain.
func RegisterDevice(current *UdpAgent, workingDir string, logLevel int, privateKey string, cipherType int, email string, deviceId string, orgId string, userData string, aspId string, resId string, cluster string, otp string) (RegistrationResult, *UdpAgent) {
	if otp == "" {
		return RegistrationResult{
			ErrCode: common.ErrInvalidInput.ErrorCode(),
			ErrMsg:  "otp is required",
		}, current
	}
	result, instance, target := prepareRegistration(
		current, workingDir, logLevel, privateKey, cipherType, email, deviceId, orgId, userData, aspId, resId, cluster, false,
	)
	if target == nil {
		return result, instance
	}
	ack, err := instance.RegisterPublicKey(otp, target)
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
	return result, instance
}
