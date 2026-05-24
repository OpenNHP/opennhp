package core

import (
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
)

func TestNHPErrorCodesMatchCHeader(t *testing.T) {
	headerCodes := parseNHPErrorCodes(t)

	tests := []struct {
		name   string
		goCode int
	}{
		{name: "ERR_NHP_SUCCESS", goCode: errNHPSuccess},
		{name: "ERR_NHP_DEVICE_NOT_INITIALIZED", goCode: errNHPDeviceNotInitialized},
		{name: "ERR_NHP_DEVICE_ALREADY_CREATED", goCode: errNHPDeviceAlreadyCreated},
		{name: "ERR_NHP_CIPHER_NOT_SUPPORTED", goCode: errNHPCipherNotSupported},
		{name: "ERR_NHP_OPERATION_NOT_APPLICABLE", goCode: errNHPOperationNotApplicable},
		{name: "ERR_NHP_CREATE_DEVICE_FAILED", goCode: errNHPCreateDeviceFailed},
		{name: "ERR_NHP_CLOSE_DEVICE_FAILED", goCode: errNHPCloseDeviceFailed},
		{name: "ERR_NHP_SDK_RUNTIME_PANIC", goCode: errNHPSDKRuntimePanic},
		{name: "ERR_NHP_WRONG_CIPHER_SCHEME", goCode: errNHPWrongCipherScheme},
		{name: "ERR_NHP_EMPTY_PEER_PUBLIC_KEY", goCode: errNHPEmptyPeerPublicKey},
		{name: "ERR_NHP_EPHERMAL_ECDH_PEER_FAILED", goCode: errNHPEphermalECDHPeerFailed},
		{name: "ERR_NHP_DEVICE_ECDH_PEER_FAILED", goCode: errNHPDeviceECDHPeerFailed},
		{name: "ERR_NHP_IDENTITY_TOO_LONG", goCode: errNHPIdentityTooLong},
		{name: "ERR_NHP_DATA_COMPRESSION_FAILED", goCode: errNHPDataCompressionFailed},
		{name: "ERR_NHP_PACKET_SIZE_EXCEEDS_BUFFER", goCode: errNHPPacketSizeExceedsBuffer},
		{name: "ERR_NHP_CLOSE_CONNECTION", goCode: errNHPCloseConnection},
		{name: "ERR_NHP_INCORRECT_PACKET_SIZE", goCode: errNHPIncorrectPacketSize},
		{name: "ERR_NHP_MESSAGE_TYPE_NOT_MATCH_DEVICE", goCode: errNHPMessageTypeNotMatchDevice},
		{name: "ERR_NHP_SERVER_OVERLOAD", goCode: errNHPServerOverload},
		{name: "ERR_NHP_HMAC_CHECK_FAILED", goCode: errNHPHMACCheckFailed},
		{name: "ERR_NHP_SERVER_HMAC_CHECK_FAILED", goCode: errNHPServerHMACCheckFailed},
		{name: "ERR_NHP_DEVICE_ECDH_EPHERMAL_FAILED", goCode: errNHPDeviceECDHEphermalFailed},
		{name: "ERR_NHP_PEER_IDENTITY_VERIFICATION_FAILED", goCode: errNHPPeerIdentityVerificationFailed},
		{name: "ERR_NHP_AEAD_DECRYPTION_FAILED", goCode: errNHPAEADDecryptionFailed},
		{name: "ERR_NHP_DATA_DECOMPRESSION_FAILED", goCode: errNHPDataDecompressionFailed},
		{name: "ERR_NHP_DEVICE_ECDH_OBTAINED_PEER_FAILED", goCode: errNHPDeviceECDHObtainedPeerFailed},
		{name: "ERR_NHP_SERVER_REJECT_WITH_COOKIE", goCode: errNHPServerRejectWithCookie},
		{name: "ERR_NHP_REPLAY_PACKET_RECEIVED", goCode: errNHPReplayPacketReceived},
		{name: "ERR_NHP_FLOOD_PACKET_RECEIVED", goCode: errNHPFloodPacketReceived},
		{name: "ERR_NHP_STALE_PACKET_RECEIVED", goCode: errNHPStalePacketReceived},
	}

	for _, tt := range tests {
		headerCode, ok := headerCodes[tt.name]
		if !ok {
			t.Fatalf("%s missing from nhpdevicedef.h NhpError enum", tt.name)
		}
		if tt.goCode != headerCode {
			t.Fatalf("%s = %d, want header value %d", tt.name, tt.goCode, headerCode)
		}
	}
}

func parseNHPErrorCodes(t *testing.T) map[string]int {
	t.Helper()

	header, err := os.ReadFile(filepath.Join("main", "nhpdevicedef.h"))
	if err != nil {
		t.Fatalf("read nhpdevicedef.h: %v", err)
	}

	const start = "typedef enum _NhpError {"
	text := string(header)
	startIdx := strings.Index(text, start)
	if startIdx < 0 {
		t.Fatalf("NhpError enum start not found in nhpdevicedef.h")
	}
	enumBody := text[startIdx+len(start):]
	endIdx := strings.Index(enumBody, "} NhpError;")
	if endIdx < 0 {
		t.Fatalf("NhpError enum end not found in nhpdevicedef.h")
	}
	enumBody = enumBody[:endIdx]

	codes := make(map[string]int)
	value := -1
	for _, rawLine := range strings.Split(enumBody, "\n") {
		line, _, _ := strings.Cut(rawLine, "//")
		line = strings.TrimSpace(line)
		line = strings.TrimSpace(strings.TrimSuffix(line, ","))
		if line == "" {
			continue
		}

		name := line
		if before, after, ok := strings.Cut(line, "="); ok {
			name = strings.TrimSpace(before)
			value, err = strconv.Atoi(strings.TrimSpace(after))
			if err != nil {
				t.Fatalf("parse %s value %q: %v", name, after, err)
			}
		} else {
			value++
		}
		codes[name] = value
	}
	return codes
}
