package main

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/pquerna/otp"
	"github.com/pquerna/otp/totp"
)

// QRCodeSession represents a QR code login session
type QRCodeSession struct {
	SessionID     string    `json:"sessionId"`
	Token         string    `json:"token"`
	AuthServiceId string    `json:"aspId"`
	ResourceId    string    `json:"resId"`
	SrcIp         string    `json:"srcIp"`
	Status        int       `json:"status"` // 0: pending, 1: scanned, 2: confirmed, 3: expired, 4: failed
	CreatedAt     time.Time `json:"createdAt"`
	ExpiresAt     time.Time `json:"expiresAt"`
	ConfirmedAt   time.Time `json:"confirmedAt,omitempty"`
	MobileIP      string    `json:"mobileIp,omitempty"`
	MobileDevice  string    `json:"mobileDevice,omitempty"`
	ErrMsg        string    `json:"errMsg,omitempty"`

	otpFailures int  // wrong codes entered for this session
	otpPending  int  // codes being checked right now (see VerifySession)
	knocked     bool // qrvalid already ran for this session
}

// qrMaxOTPFailures is how many wrong codes a session accepts before it is
// dead. Codes are checked against the shared OTPSecretKey under otpLimiter;
// this cap is on top of that, so one session cannot be used for unlimited
// guesses. It does not imply a per-session secret, and none should be added:
// whoever displays the QR code would hold it.
const qrMaxOTPFailures = 5

// QRCodeStatus constants
const (
	QRStatusPending   = 0
	QRStatusScanned   = 1
	QRStatusConfirmed = 2
	QRStatusExpired   = 3
	QRStatusFailed    = 4
)

// QRAuthService manages QR code authentication sessions
type QRAuthService struct {
	sessions   map[string]*QRCodeSession
	mutex      sync.RWMutex
	encryptKey []byte
	hmacKey    []byte
	expiry     time.Duration
}

// QRCodeData structure (encoded in QR)
type QRCodeData struct {
	SessionID string `json:"sid"`
	Token     string `json:"tok"`
	Timestamp int64  `json:"ts"`
	AspId     string `json:"asp"`
	ResId     string `json:"res"`
	Server    string `json:"srv"`
	Signature string `json:"sig"`
}

// QRVerifyRequest is the request from mobile device
type QRVerifyRequest struct {
	SessionID     string `json:"sessionId"`
	Token         string `json:"token"`
	OTPCode       string `json:"otpCode"`
	DeviceInfo    string `json:"deviceInfo"`
	EncryptedData string `json:"encryptedData"`
}

// QRGenerateResponse is the response when generating QR code
type QRGenerateResponse struct {
	Success   bool   `json:"success"`
	SessionID string `json:"sessionId,omitempty"`
	QRData    string `json:"qrData,omitempty"`
	ServerUrl string `json:"serverUrl,omitempty"`
	AspId     string `json:"aspId,omitempty"`
	ResId     string `json:"resId,omitempty"`
	ExpiresAt int64  `json:"expiresAt,omitempty"`
	ErrMsg    string `json:"errMsg,omitempty"`
}

// QRStatusResponse is the response for status check
type QRStatusResponse struct {
	Success     bool   `json:"success"`
	Status      int    `json:"status"`
	StatusText  string `json:"statusText"`
	RedirectUrl string `json:"redirectUrl,omitempty"`
	ErrMsg      string `json:"errMsg,omitempty"`
}

// QRVerifyResponse is the response for mobile verification
type QRVerifyResponse struct {
	Success bool   `json:"success"`
	ErrMsg  string `json:"errMsg,omitempty"`
}

var qrAuthService *QRAuthService
var qrAuthOnce sync.Once

// InitQRAuthService initializes the QR auth service
func InitQRAuthService() {
	qrAuthOnce.Do(func() {
		encKey := make([]byte, 32)
		hmacKey := make([]byte, 32)
		rand.Read(encKey)
		rand.Read(hmacKey)

		qrAuthService = &QRAuthService{
			sessions:   make(map[string]*QRCodeSession),
			encryptKey: encKey,
			hmacKey:    hmacKey,
			expiry:     5 * time.Minute,
		}

		go qrAuthService.cleanupExpiredSessions()
	})
}

// GetQRAuthService returns the singleton QR auth service
func GetQRAuthService() *QRAuthService {
	if qrAuthService == nil {
		InitQRAuthService()
	}
	return qrAuthService
}

// GenerateSession creates a new QR code session
//
// The session has no OTP secret of its own: whoever displays the QR code
// would hold it and could confirm the session themselves. The phone proves
// itself with a code from the configured OTPSecretKey instead.
func (s *QRAuthService) GenerateSession(aspId, resId, srcIp, serverUrl string) (*QRCodeSession, *QRCodeData, error) {
	// Generate session ID
	sessionBytes := make([]byte, 16)
	rand.Read(sessionBytes)
	sessionID := hex.EncodeToString(sessionBytes)

	// Generate random token
	tokenBytes := make([]byte, 32)
	rand.Read(tokenBytes)
	token := base64.URLEncoding.EncodeToString(tokenBytes)

	now := time.Now()
	session := &QRCodeSession{
		SessionID:     sessionID,
		Token:         token,
		AuthServiceId: aspId,
		ResourceId:    resId,
		SrcIp:         srcIp,
		Status:        QRStatusPending,
		CreatedAt:     now,
		ExpiresAt:     now.Add(s.expiry),
	}

	// Store session
	s.mutex.Lock()
	s.sessions[sessionID] = session
	s.mutex.Unlock()

	// Create QR code data
	qrData := &QRCodeData{
		SessionID: sessionID,
		Token:     token,
		Timestamp: now.Unix(),
		AspId:     aspId,
		ResId:     resId,
		Server:    serverUrl,
	}

	// Sign the QR data
	qrData.Signature = s.signQRData(qrData)

	log.Info("Authenticator session created: sessionId=%s, aspId=%s, resId=%s", sessionID, aspId, resId)
	return session, qrData, nil
}

// signQRData creates HMAC signature for QR data
func (s *QRAuthService) signQRData(data *QRCodeData) string {
	h := hmac.New(sha256.New, s.hmacKey)
	signData := fmt.Sprintf("%s|%s|%d|%s|%s|%s", data.SessionID, data.Token, data.Timestamp, data.AspId, data.ResId, data.Server)
	h.Write([]byte(signData))
	return base64.URLEncoding.EncodeToString(h.Sum(nil))
}

// EncryptQRData encrypts the QR code data
func (s *QRAuthService) EncryptQRData(data *QRCodeData) (string, error) {
	jsonData, err := json.Marshal(data)
	if err != nil {
		return "", err
	}

	block, err := aes.NewCipher(s.encryptKey)
	if err != nil {
		return "", err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}

	nonce := make([]byte, gcm.NonceSize())
	if _, err := rand.Read(nonce); err != nil {
		return "", err
	}

	ciphertext := gcm.Seal(nonce, nonce, jsonData, nil)
	return base64.URLEncoding.EncodeToString(ciphertext), nil
}

// DecryptQRData decrypts the QR code data
func (s *QRAuthService) DecryptQRData(encrypted string) (*QRCodeData, error) {
	ciphertext, err := base64.URLEncoding.DecodeString(encrypted)
	if err != nil {
		return nil, err
	}

	block, err := aes.NewCipher(s.encryptKey)
	if err != nil {
		return nil, err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	nonceSize := gcm.NonceSize()
	if len(ciphertext) < nonceSize {
		return nil, fmt.Errorf("ciphertext too short")
	}

	nonce, ciphertext := ciphertext[:nonceSize], ciphertext[nonceSize:]
	plaintext, err := gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return nil, err
	}

	var data QRCodeData
	if err := json.Unmarshal(plaintext, &data); err != nil {
		return nil, err
	}

	return &data, nil
}

// GetSession returns a copy of the session, taken under the lock, or nil.
// A copy because other handlers write the session concurrently.
func (s *QRAuthService) GetSession(sessionID string) *QRCodeSession {
	s.mutex.Lock() // may mark the session expired
	defer s.mutex.Unlock()

	session, exists := s.sessions[sessionID]
	if !exists {
		return nil
	}

	if time.Now().After(session.ExpiresAt) {
		session.Status = QRStatusExpired
	}

	snapshot := *session
	return &snapshot
}

// ValidateConfiguredOTP validates a 6-digit SHA1 TOTP code against the
// configured secret key with the configured period, accepting
// otpValidationSkew steps either side of the current one.
func (s *QRAuthService) ValidateConfiguredOTP(secretKey, otpCode string, periodSeconds int) bool {
	ok, err := totp.ValidateCustom(otpCode, secretKey, time.Now().UTC(), totp.ValidateOpts{
		Period:    uint(periodSeconds),
		Skew:      otpValidationSkew,
		Digits:    otp.DigitsSix,
		Algorithm: otp.AlgorithmSHA1,
	})
	return ok && err == nil
}

// ConsumeConfirmedSession authorizes the one knock a QR session is good
// for: the session must exist, be unexpired, have had its OTP verified by
// VerifySession, belong to resId, be requested from the client that
// generated it (same otpLimitKey as SrcIp), and not have been used before.
// The session ID is in the QR payload, so without the client check anyone
// who sees the QR code could race the browser for the knock.
func (s *QRAuthService) ConsumeConfirmedSession(sessionID, resId, requesterKey string) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	session, exists := s.sessions[sessionID]
	if !exists {
		return fmt.Errorf("session not found")
	}
	if time.Now().After(session.ExpiresAt) {
		session.Status = QRStatusExpired
		return fmt.Errorf("session expired")
	}
	if session.Status != QRStatusConfirmed {
		return fmt.Errorf("session not confirmed")
	}
	if session.ResourceId != resId {
		return fmt.Errorf("session is for a different resource")
	}
	if otpLimitKey(session.SrcIp) != requesterKey {
		// Also what a dual-stack browser sees if it switched address family
		// between generate and qrvalid; a new QR code fixes both.
		return fmt.Errorf("this QR code was generated from a different network address; refresh the QR code and try again")
	}
	if session.knocked {
		return fmt.Errorf("session already used")
	}
	session.knocked = true
	return nil
}

// VerifySession confirms a session. checkOTP verifies the phone's code
// (checkConfiguredOTP in production) and runs only once the session and its
// token are known to be good; an errOTPInvalid from it counts against the
// session's qrMaxOTPFailures.
//
// checkOTP runs without the service lock, so status polls and generate do
// not wait on it. The attempt is reserved against the session's cap before
// the lock is released, so parallel verifies cannot exceed the cap, and the
// session is re-checked before it is confirmed.
func (s *QRAuthService) VerifySession(sessionID, token, mobileIP, deviceInfo string, checkOTP func() error) error {
	s.mutex.Lock()
	session, err := s.verifiableSessionLocked(sessionID, token)
	if err == nil && session.otpFailures+session.otpPending >= qrMaxOTPFailures {
		err = fmt.Errorf("too many failed attempts, generate a new QR code")
	}
	if err != nil {
		s.mutex.Unlock()
		return err
	}
	session.otpPending++
	s.mutex.Unlock()

	otpErr := checkOTP()

	s.mutex.Lock()
	defer s.mutex.Unlock()
	session.otpPending--
	if otpErr != nil {
		if errors.Is(otpErr, errOTPInvalid) {
			session.otpFailures++
			session.Status = QRStatusFailed
			session.ErrMsg = otpErr.Error()
		}
		return otpErr
	}
	// The session may have expired, been confirmed or been cleaned up while
	// the code was checked.
	if _, err := s.verifiableSessionLocked(sessionID, token); err != nil {
		return err
	}

	session.Status = QRStatusConfirmed
	session.ConfirmedAt = time.Now()
	session.MobileIP = mobileIP
	session.MobileDevice = deviceInfo

	return nil
}

// verifiableSessionLocked returns the session if it exists, is unexpired,
// is not yet confirmed and token matches. s.mutex must be held.
func (s *QRAuthService) verifiableSessionLocked(sessionID, token string) (*QRCodeSession, error) {
	session, exists := s.sessions[sessionID]
	if !exists {
		return nil, fmt.Errorf("session not found")
	}

	if time.Now().After(session.ExpiresAt) {
		session.Status = QRStatusExpired
		return nil, fmt.Errorf("session expired")
	}

	if session.Status == QRStatusConfirmed {
		return nil, fmt.Errorf("session already confirmed")
	}

	if !hmac.Equal([]byte(session.Token), []byte(token)) {
		session.Status = QRStatusFailed
		session.ErrMsg = "invalid token"
		return nil, fmt.Errorf("invalid token")
	}
	return session, nil
}

// MarkSessionScanned records that a phone scanned the QR code. The request
// is unauthenticated, so it only moves a pending session forward and never
// overwrites a confirmed, failed or expired one.
func (s *QRAuthService) MarkSessionScanned(sessionID string) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	session, exists := s.sessions[sessionID]
	if !exists {
		return fmt.Errorf("session not found")
	}
	if time.Now().After(session.ExpiresAt) {
		session.Status = QRStatusExpired
		return fmt.Errorf("session expired")
	}
	if session.Status == QRStatusPending {
		session.Status = QRStatusScanned
	}
	return nil
}

// cleanupExpiredSessions removes expired sessions
func (s *QRAuthService) cleanupExpiredSessions() {
	ticker := time.NewTicker(1 * time.Minute)
	defer ticker.Stop()

	for range ticker.C {
		s.mutex.Lock()
		now := time.Now()
		for id, session := range s.sessions {
			if now.After(session.ExpiresAt.Add(5 * time.Minute)) {
				delete(s.sessions, id)
			}
		}
		s.mutex.Unlock()
	}
}

// HandleQRGenerate handles QR code generation
func HandleQRGenerate(ctx *gin.Context, resId string) {
	service := GetQRAuthService()

	aspId := "authenticator"
	srcIp := otpClientAddr(ctx, requestTrustedProxies())

	// Build server URL
	scheme := "https"
	if ctx.Request.TLS == nil {
		scheme = "http"
	}
	serverUrl := fmt.Sprintf("%s://%s", scheme, ctx.Request.Host)

	session, qrData, err := service.GenerateSession(aspId, resId, srcIp, serverUrl)
	if err != nil {
		ctx.JSON(http.StatusOK, QRGenerateResponse{
			Success: false,
			ErrMsg:  err.Error(),
		})
		return
	}

	// Encrypt QR data
	encryptedData, err := service.EncryptQRData(qrData)
	if err != nil {
		ctx.JSON(http.StatusOK, QRGenerateResponse{
			Success: false,
			ErrMsg:  "Failed to encrypt QR data",
		})
		return
	}

	response := QRGenerateResponse{
		Success:   true,
		SessionID: session.SessionID,
		QRData:    encryptedData,
		ServerUrl: serverUrl,
		AspId:     aspId,
		ResId:     resId,
		ExpiresAt: session.ExpiresAt.Unix(),
	}

	ctx.JSON(http.StatusOK, response)
}

// HandleQRStatus handles status check requests
func HandleQRStatus(ctx *gin.Context) {
	sessionId := ctx.Query("sessionId")
	if sessionId == "" {
		ctx.JSON(http.StatusOK, QRStatusResponse{
			Success: false,
			ErrMsg:  "missing sessionId",
		})
		return
	}

	service := GetQRAuthService()
	session := service.GetSession(sessionId)
	if session == nil {
		ctx.JSON(http.StatusOK, QRStatusResponse{
			Success: false,
			ErrMsg:  "session not found",
		})
		return
	}

	statusTexts := map[int]string{
		QRStatusPending:   "pending",
		QRStatusScanned:   "scanned",
		QRStatusConfirmed: "confirmed",
		QRStatusExpired:   "expired",
		QRStatusFailed:    "failed",
	}

	response := QRStatusResponse{
		Success:    true,
		Status:     session.Status,
		StatusText: statusTexts[session.Status],
	}

	if session.Status == QRStatusConfirmed {
		res := findResource(session.ResourceId)
		if res != nil && len(res.RedirectUrl) > 0 {
			response.RedirectUrl = res.RedirectUrl
		}
	}

	ctx.JSON(http.StatusOK, response)
}

// HandleQRScan handles scan notification from mobile
func HandleQRScan(ctx *gin.Context) {
	sessionId := ctx.Query("sessionId")
	if sessionId == "" {
		ctx.JSON(http.StatusOK, gin.H{
			"success": false,
			"errMsg":  "missing sessionId",
		})
		return
	}

	service := GetQRAuthService()
	err := service.MarkSessionScanned(sessionId)
	if err != nil {
		ctx.JSON(http.StatusOK, gin.H{
			"success": false,
			"errMsg":  err.Error(),
		})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"success": true,
	})
}

// HandleQRVerify handles verification from mobile device
func HandleQRVerify(ctx *gin.Context) {
	var req QRVerifyRequest

	// Support both GET and POST
	if ctx.Request.Method == "GET" {
		req.EncryptedData = ctx.Query("encryptedData")
		req.OTPCode = ctx.Query("otpCode")
		req.DeviceInfo = ctx.Query("deviceInfo")
	} else {
		if err := ctx.ShouldBindJSON(&req); err != nil {
			ctx.JSON(http.StatusOK, QRVerifyResponse{
				Success: false,
				ErrMsg:  "invalid request",
			})
			return
		}
	}

	if req.EncryptedData == "" || req.OTPCode == "" {
		ctx.JSON(http.StatusOK, QRVerifyResponse{
			Success: false,
			ErrMsg:  "missing required fields",
		})
		return
	}

	service := GetQRAuthService()

	// Decrypt QR data
	qrData, err := service.DecryptQRData(req.EncryptedData)
	if err != nil {
		ctx.JSON(http.StatusOK, QRVerifyResponse{
			Success: false,
			ErrMsg:  "invalid QR data",
		})
		return
	}

	conf := baseConf
	if conf == nil || conf.OTPSecretKey == "" {
		ctx.JSON(http.StatusOK, QRVerifyResponse{
			Success: false,
			ErrMsg:  "OTP secret key not configured",
		})
		return
	}

	// The phone's code is checked against the configured secret under the
	// same limiter as otpvalid, keyed by the phone's address.
	clientIP := otpClientIP(ctx, conf.trustedProxies)
	err = service.VerifySession(qrData.SessionID, qrData.Token, otpClientAddr(ctx, conf.trustedProxies), req.DeviceInfo, func() error {
		return checkConfiguredOTP(conf, clientIP, req.OTPCode)
	})
	if err != nil {
		status := http.StatusOK
		if errors.Is(err, errOTPRateLimited) {
			status = http.StatusTooManyRequests
		}
		ctx.JSON(status, QRVerifyResponse{
			Success: false,
			ErrMsg:  err.Error(),
		})
		return
	}

	ctx.JSON(http.StatusOK, QRVerifyResponse{
		Success: true,
	})
}
