package core

/*
#include "main/nhpdevicedef.h"
*/
import "C"
import (
	"strconv"
)

var errorMap map[int]*Error = make(map[int]*Error)

type Error struct {
	num         int
	msg         string
	extraErr    error
	hasExtraErr bool
}

func (e *Error) SetExtraError(err error) {
	e.extraErr = err
	if err != nil {
		e.hasExtraErr = true
	}
}

// implment NhpError interface
func (e *Error) Error() string {
	if e.hasExtraErr {
		e.hasExtraErr = false
		defer e.SetExtraError(nil)
		return e.msg + ": " + e.extraErr.Error()
	}
	return e.msg
}

func (e *Error) ErrorCode() string {
	return strconv.Itoa(e.num)
}

func (e *Error) ErrorNumber() int {
	return e.num
}

func newError(number C.int, msg string) *Error {
	e := &Error{
		num: int(number),
		msg: msg,
	}
	errorMap[e.num] = e
	return e
}

func ErrorToErrorNumber(err error) int {
	e, ok := err.(*Error)
	if ok {
		return e.ErrorNumber()
	}
	return -1
}

func ErrorToString(err error) string {
	e, ok := err.(*Error)
	if ok {
		return e.Error()
	}
	return ""
}

func ErrorCodeToError(number int) *Error {
	e, found := errorMap[number]
	if found {
		return e
	}
	return nil // should not happen
}

// device sdk errors
var (
	ErrSuccess = newError(C.ERR_NHP_SUCCESS, "")

	// device
	ErrCipherNotSupported = newError(C.ERR_NHP_CIPHER_NOT_SUPPORTED, "cipher scheme not supported")
	ErrNotApplicable      = newError(C.ERR_NHP_OPERATION_NOT_APPLICABLE, "operation not applicable")
	ErrCreateDeviceFailed = newError(C.ERR_NHP_CREATE_DEVICE_FAILED, "failed to create nhp device")
	ErrCloseDeviceFailed  = newError(C.ERR_NHP_CLOSE_DEVICE_FAILED, "attempt to close a non-initialized nhp device")
	ErrRuntimePanic       = newError(C.ERR_NHP_SDK_RUNTIME_PANIC, "runtime panic encountered")

	// initiator and encryption
	ErrWrongCipherScheme       = newError(C.ERR_NHP_WRONG_CIPHER_SCHEME, "a wrong cipher scheme is used")
	ErrEmptyPeerPublicKey      = newError(C.ERR_NHP_EMPTY_PEER_PUBLIC_KEY, "remote peer public key is not set")
	ErrEphermalECDHPeerFailed  = newError(C.ERR_NHP_EPHERMAL_ECDH_PEER_FAILED, "ephermal ECDH failed with peer")
	ErrDeviceECDHPeerFailed    = newError(C.ERR_NHP_DEVICE_ECDH_PEER_FAILED, "device ECDH failed with peer")
	ErrIdentityTooLong         = newError(C.ERR_NHP_IDENTITY_TOO_LONG, "identity exceeds max length")
	ErrDataCompressionFailed   = newError(C.ERR_NHP_DATA_COMPRESSION_FAILED, "data compression failed")
	ErrPacketSizeExceedsBuffer = newError(C.ERR_NHP_PACKET_SIZE_EXCEEDS_BUFFER, "packet size longer than send buffer")

	// responder and decryption
	ErrCloseConnection                = newError(C.ERR_NHP_CLOSE_CONNECTION, "disengage nhp access immediately")
	ErrIncorrectPacketSize            = newError(C.ERR_NHP_INCORRECT_PACKET_SIZE, "incorrect packet size")
	ErrMessageTypeNotMatchDevice      = newError(C.ERR_NHP_MESSAGE_TYPE_NOT_MATCH_DEVICE, "message type does not match device")
	ErrServerOverload                 = newError(C.ERR_NHP_SERVER_OVERLOAD, "the packet is dropped due to server overload")
	ErrHMACCheckFailed                = newError(C.ERR_NHP_HMAC_CHECK_FAILED, "HMAC validation failed")
	ErrServerHMACCheckFailed          = newError(C.ERR_NHP_SERVER_HMAC_CHECK_FAILED, "server HMAC validation failed")
	ErrDeviceECDHEphermalFailed       = newError(C.ERR_NHP_DEVICE_ECDH_EPHERMAL_FAILED, "device ECDH failed with ephermal")
	ErrPeerIdentityVerificationFailed = newError(C.ERR_NHP_PEER_IDENTITY_VERIFICATION_FAILED, "failed to verify peer's identity with apk")
	ErrAEADDecryptionFailed           = newError(C.ERR_NHP_AEAD_DECRYPTION_FAILED, "aead decryption failed")
	ErrDataDecompressionFailed        = newError(C.ERR_NHP_DATA_DECOMPRESSION_FAILED, "data decompression failed")
	ErrDeviceECDHObtainedPeerFailed   = newError(C.ERR_NHP_DEVICE_ECDH_OBTAINED_PEER_FAILED, "device ECDH failed with obtained peer")
	ErrServerRejectWithCookie         = newError(C.ERR_NHP_SERVER_REJECT_WITH_COOKIE, "server overload, stop processing packet and return cookie")
	ErrReplayPacketReceived           = newError(C.ERR_NHP_REPLAY_PACKET_RECEIVED, "received replay packet, drop")
	ErrFloodPacketReceived            = newError(C.ERR_NHP_FLOOD_PACKET_RECEIVED, "received flood packet, drop")
	ErrStalePacketReceived            = newError(C.ERR_NHP_STALE_PACKET_RECEIVED, "received stale packet, drop")
)
