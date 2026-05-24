package core

import "strconv"

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

func newError(number int, msg string) *Error {
	e := &Error{
		num: number,
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

// NHP error numbers. Keep these in sync with the NhpError enum
// in nhp/core/main/nhpdevicedef.h; errors_test.go verifies parity.
const (
	errNHPSuccess = 0
)

const (
	errNHPDeviceNotInitialized = 30000 + iota
	errNHPDeviceAlreadyCreated
	errNHPCipherNotSupported
	errNHPOperationNotApplicable
	errNHPCreateDeviceFailed
	errNHPCloseDeviceFailed
	errNHPSDKRuntimePanic
)

const (
	errNHPWrongCipherScheme = 31000 + iota
	errNHPEmptyPeerPublicKey
	errNHPEphermalECDHPeerFailed
	errNHPDeviceECDHPeerFailed
	errNHPIdentityTooLong
	errNHPDataCompressionFailed
	errNHPPacketSizeExceedsBuffer
)

const (
	errNHPCloseConnection = 32001 + iota
	errNHPIncorrectPacketSize
	errNHPMessageTypeNotMatchDevice
	errNHPServerOverload
	errNHPHMACCheckFailed
	errNHPServerHMACCheckFailed
	errNHPDeviceECDHEphermalFailed
	errNHPPeerIdentityVerificationFailed
	errNHPAEADDecryptionFailed
	errNHPDataDecompressionFailed
	errNHPDeviceECDHObtainedPeerFailed
	errNHPServerRejectWithCookie
	errNHPReplayPacketReceived
	errNHPFloodPacketReceived
	errNHPStalePacketReceived
)

// device sdk errors
var (
	ErrSuccess = newError(errNHPSuccess, "")

	// device
	ErrCipherNotSupported = newError(errNHPCipherNotSupported, "cipher scheme not supported")
	ErrNotApplicable      = newError(errNHPOperationNotApplicable, "operation not applicable")
	ErrCreateDeviceFailed = newError(errNHPCreateDeviceFailed, "failed to create nhp device")
	ErrCloseDeviceFailed  = newError(errNHPCloseDeviceFailed, "attempt to close a non-initialized nhp device")
	ErrRuntimePanic       = newError(errNHPSDKRuntimePanic, "runtime panic encountered")

	// initiator and encryption
	ErrWrongCipherScheme       = newError(errNHPWrongCipherScheme, "a wrong cipher scheme is used")
	ErrEmptyPeerPublicKey      = newError(errNHPEmptyPeerPublicKey, "remote peer public key is not set")
	ErrEphermalECDHPeerFailed  = newError(errNHPEphermalECDHPeerFailed, "ephermal ECDH failed with peer")
	ErrDeviceECDHPeerFailed    = newError(errNHPDeviceECDHPeerFailed, "device ECDH failed with peer")
	ErrIdentityTooLong         = newError(errNHPIdentityTooLong, "identity exceeds max length")
	ErrDataCompressionFailed   = newError(errNHPDataCompressionFailed, "data compression failed")
	ErrPacketSizeExceedsBuffer = newError(errNHPPacketSizeExceedsBuffer, "packet size longer than send buffer")

	// responder and decryption
	ErrCloseConnection                = newError(errNHPCloseConnection, "disengage nhp access immediately")
	ErrIncorrectPacketSize            = newError(errNHPIncorrectPacketSize, "incorrect packet size")
	ErrMessageTypeNotMatchDevice      = newError(errNHPMessageTypeNotMatchDevice, "message type does not match device")
	ErrServerOverload                 = newError(errNHPServerOverload, "the packet is dropped due to server overload")
	ErrHMACCheckFailed                = newError(errNHPHMACCheckFailed, "HMAC validation failed")
	ErrServerHMACCheckFailed          = newError(errNHPServerHMACCheckFailed, "server HMAC validation failed")
	ErrDeviceECDHEphermalFailed       = newError(errNHPDeviceECDHEphermalFailed, "device ECDH failed with ephermal")
	ErrPeerIdentityVerificationFailed = newError(errNHPPeerIdentityVerificationFailed, "failed to verify peer's identity with apk")
	ErrAEADDecryptionFailed           = newError(errNHPAEADDecryptionFailed, "aead decryption failed")
	ErrDataDecompressionFailed        = newError(errNHPDataDecompressionFailed, "data decompression failed")
	ErrDeviceECDHObtainedPeerFailed   = newError(errNHPDeviceECDHObtainedPeerFailed, "device ECDH failed with obtained peer")
	ErrServerRejectWithCookie         = newError(errNHPServerRejectWithCookie, "server overload, stop processing packet and return cookie")
	ErrReplayPacketReceived           = newError(errNHPReplayPacketReceived, "received replay packet, drop")
	ErrFloodPacketReceived            = newError(errNHPFloodPacketReceived, "received flood packet, drop")
	ErrStalePacketReceived            = newError(errNHPStalePacketReceived, "received stale packet, drop")
)
