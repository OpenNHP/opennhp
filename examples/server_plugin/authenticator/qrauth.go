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
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/pquerna/otp/totp"
)

// QRCodeSession represents a QR code login session
type QRCodeSession struct {
	SessionID     string    `json:"sessionId"`
	Token         string    `json:"token"`
	OTPSecret     string    `json:"otpSecret"`
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
}

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
	OTPSecret string `json:"otpSecret,omitempty"`
	OTPUri    string `json:"otpUri,omitempty"` // otpauth:// URI for Google Authenticator
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
func (s *QRAuthService) GenerateSession(aspId, resId, srcIp, serverUrl string) (*QRCodeSession, *QRCodeData, string, error) {
	// Generate session ID
	sessionBytes := make([]byte, 16)
	rand.Read(sessionBytes)
	sessionID := hex.EncodeToString(sessionBytes)

	// Generate random token
	tokenBytes := make([]byte, 32)
	rand.Read(tokenBytes)
	token := base64.URLEncoding.EncodeToString(tokenBytes)

	// Generate TOTP secret
	key, err := totp.Generate(totp.GenerateOpts{
		Issuer:      "OpenNHP Authenticator",
		AccountName: sessionID[:8],
		Period:      30,
		SecretSize:  20,
	})
	if err != nil {
		return nil, nil, "", fmt.Errorf("failed to generate OTP secret: %v", err)
	}

	now := time.Now()
	session := &QRCodeSession{
		SessionID:     sessionID,
		Token:         token,
		OTPSecret:     key.Secret(),
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

	// Generate OTP URI for Google Authenticator
	otpUri := key.URL()

	log.Info("Authenticator session created: sessionId=%s, aspId=%s, resId=%s", sessionID, aspId, resId)
	return session, qrData, otpUri, nil
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

// GetSession retrieves a session by ID
func (s *QRAuthService) GetSession(sessionID string) *QRCodeSession {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	session, exists := s.sessions[sessionID]
	if !exists {
		return nil
	}

	if time.Now().After(session.ExpiresAt) {
		session.Status = QRStatusExpired
	}

	return session
}

// VerifyOTPOnly verifies OTP code without mobile scan
func (s *QRAuthService) VerifyOTPOnly(sessionID, otpCode string) error {
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

	// Verify OTP
	if !totp.Validate(otpCode, session.OTPSecret) {
		session.Status = QRStatusFailed
		session.ErrMsg = "invalid OTP code"
		return fmt.Errorf("invalid OTP code")
	}

	session.Status = QRStatusConfirmed
	session.ConfirmedAt = time.Now()
	return nil
}

// ValidateConfiguredOTP validates an OTP code against the configured secret key
func (s *QRAuthService) ValidateConfiguredOTP(secretKey, otpCode string) bool {
	return totp.Validate(otpCode, secretKey)
}

// VerifySession verifies a session with OTP
func (s *QRAuthService) VerifySession(sessionID, token, otpCode, mobileIP, deviceInfo string) error {
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

	if session.Status == QRStatusConfirmed {
		return fmt.Errorf("session already confirmed")
	}

	// Verify token
	if session.Token != token {
		session.Status = QRStatusFailed
		session.ErrMsg = "invalid token"
		return fmt.Errorf("invalid token")
	}

	// Verify OTP
	if !totp.Validate(otpCode, session.OTPSecret) {
		session.Status = QRStatusFailed
		session.ErrMsg = "invalid OTP code"
		return fmt.Errorf("invalid OTP code")
	}

	session.Status = QRStatusConfirmed
	session.ConfirmedAt = time.Now()
	session.MobileIP = mobileIP
	session.MobileDevice = deviceInfo

	return nil
}

// UpdateSessionStatus updates session status
func (s *QRAuthService) UpdateSessionStatus(sessionID string, status int) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	session, exists := s.sessions[sessionID]
	if !exists {
		return fmt.Errorf("session not found")
	}

	session.Status = status
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
	srcIp := ctx.ClientIP()

	// Build server URL
	scheme := "https"
	if ctx.Request.TLS == nil {
		scheme = "http"
	}
	serverUrl := fmt.Sprintf("%s://%s", scheme, ctx.Request.Host)

	session, qrData, otpUri, err := service.GenerateSession(aspId, resId, srcIp, serverUrl)
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
		OTPSecret: session.OTPSecret,
		OTPUri:    otpUri,
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
	err := service.UpdateSessionStatus(sessionId, QRStatusScanned)
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

	// Verify session
	err = service.VerifySession(qrData.SessionID, qrData.Token, req.OTPCode, ctx.ClientIP(), req.DeviceInfo)
	if err != nil {
		ctx.JSON(http.StatusOK, QRVerifyResponse{
			Success: false,
			ErrMsg:  err.Error(),
		})
		return
	}

	ctx.JSON(http.StatusOK, QRVerifyResponse{
		Success: true,
	})
}
