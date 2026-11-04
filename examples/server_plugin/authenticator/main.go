package main

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/netip"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/OpenNHP/opennhp/nhp/common"
	nhplog "github.com/OpenNHP/opennhp/nhp/log"
	"github.com/OpenNHP/opennhp/nhp/plugins"
	"github.com/OpenNHP/opennhp/nhp/utils"
	"github.com/gin-gonic/gin"

	toml "github.com/pelletier/go-toml/v2"
)

type config struct {
	// Plugin specific config
	OTPSecretKey string `toml:"OTPSecretKey"` // Fixed OTP secret key for manual OTP input
	OTPPeriod    int    `toml:"OTPPeriod"`    // OTP validity period in seconds (default 30)
	OTPDigits    int    `toml:"OTPDigits"`    // Number of OTP digits (default 6)
	QRCodeExpiry int    `toml:"QRCodeExpiry"` // QR code expiry in seconds (default 300)
	// OTPAllowReuse lets the same OTP code be accepted more than once while
	// it is valid. Off by default: a code observed in transit cannot be
	// replayed. All users share OTPSecretKey, so a deployment where several
	// people log in within the same 30-second window needs it on.
	OTPAllowReuse bool `toml:"OTPAllowReuse"`
	// ShowOTPSecret renders OTPSecretKey into the public login page so a
	// visitor can enrol it in an authenticator app. Anyone who can load the
	// page can then generate valid codes, so OTP stops being a secret and
	// only the rate limit is left: enable it for public demos only.
	ShowOTPSecret bool `toml:"ShowOTPSecret"`
	// TrustedProxies lists reverse proxies (addresses or CIDR prefixes) in
	// front of nhp-server. OTP attempts arriving from one of them are limited
	// by the X-Forwarded-For client address instead of the proxy's own.
	TrustedProxies []string `toml:"TrustedProxies"`

	trustedProxies []netip.Prefix
}

var (
	// Plugin Settings
	log           *nhplog.Logger
	pluginDirPath string
	hostname      string
	localIp       string
	localMac      string
)

var (
	name    = "authenticator"
	version = "1.0.0"

	baseConfigWatch io.Closer
	resConfigWatch  io.Closer

	baseConf         *config
	resourceMapMutex sync.Mutex
	resourceMap      common.ResourceGroupMap
)

var (
	errLoadConfig = fmt.Errorf("config load error")
)

func Version() string {
	return fmt.Sprintf("%s v%s", name, version)
}

func Init(in *plugins.PluginParamsIn) error {
	if in.PluginDirPath != nil {
		pluginDirPath = *in.PluginDirPath
	}
	if in.Log != nil {
		log = in.Log
	}
	if in.Hostname != nil {
		hostname = *in.Hostname
	}
	if in.LocalIp != nil {
		localIp = *in.LocalIp
	}
	if in.LocalMac != nil {
		localMac = *in.LocalMac
	}

	// Initialize QR Auth Service
	InitQRAuthService()

	// load config
	fileNameBase := filepath.Join(pluginDirPath, "etc", "config.toml")
	if err := updateConfig(fileNameBase); err != nil {
		_ = err
	}

	baseConfigWatch = utils.WatchFile(fileNameBase, func() {
		log.Info("base config: %s has been updated", fileNameBase)
		updateConfig(fileNameBase)
	})

	fileNameRes := filepath.Join(pluginDirPath, "etc", "resource.toml")
	if err := updateResource(fileNameRes); err != nil {
		_ = err
	}
	resConfigWatch = utils.WatchFile(fileNameRes, func() {
		log.Info("resource config: %s has been updated", fileNameRes)
		updateResource(fileNameRes)
	})

	return nil
}

func updateConfig(file string) (err error) {
	utils.CatchPanicThenRun(func() {
		err = errLoadConfig
	})

	content, err := os.ReadFile(file)
	if err != nil {
		log.Error("failed to read config file %s: %v", file, err)
		return err
	}

	conf := &config{}
	if err = toml.Unmarshal(content, conf); err != nil {
		log.Error("failed to unmarshal config: %v", err)
		return err
	}

	// Set defaults
	if conf.OTPPeriod == 0 {
		conf.OTPPeriod = 30
	}
	if conf.OTPPeriod < 15 || conf.OTPPeriod > 300 {
		log.Error("OTPPeriod=%d is outside 15..300 seconds, using 30", conf.OTPPeriod)
		conf.OTPPeriod = 30
	}
	if conf.OTPDigits == 0 {
		conf.OTPDigits = 6
	}
	if conf.OTPDigits != 6 {
		// The login page takes exactly six digits.
		log.Error("OTPDigits=%d is not supported, using 6", conf.OTPDigits)
		conf.OTPDigits = 6
	}
	if conf.trustedProxies, err = parseTrustedProxies(conf.TrustedProxies); err != nil {
		// Fail closed: limit by the TCP peer, never trust the header.
		log.Error("%v; ignoring TrustedProxies", err)
		conf.trustedProxies, err = nil, nil
	}
	if conf.QRCodeExpiry == 0 {
		conf.QRCodeExpiry = 300
	}

	baseConf = conf
	log.Info("Authenticator plugin config loaded: OTPPeriod=%d, OTPDigits=%d, QRCodeExpiry=%d, OTPAllowReuse=%v, ShowOTPSecret=%v, TrustedProxies=%d",
		conf.OTPPeriod, conf.OTPDigits, conf.QRCodeExpiry, conf.OTPAllowReuse, conf.ShowOTPSecret, len(conf.trustedProxies))

	return nil
}

func updateResource(file string) (err error) {
	utils.CatchPanicThenRun(func() {
		err = errLoadConfig
	})

	content, err := os.ReadFile(file)
	if err != nil {
		log.Error("failed to read resource file %s: %v", file, err)
		return err
	}

	resourceMapMutex.Lock()
	defer resourceMapMutex.Unlock()

	resourceMap = make(common.ResourceGroupMap)
	if err = toml.Unmarshal(content, &resourceMap); err != nil {
		log.Error("failed to unmarshal resource: %v", err)
		return err
	}

	// Set resource ID and auth service ID
	for resId, res := range resourceMap {
		res.AuthServiceId = name
		res.ResourceId = resId
	}

	return nil
}

func findResource(resId string) *common.ResourceData {
	resourceMapMutex.Lock()
	defer resourceMapMutex.Unlock()
	res := resourceMap[resId]
	return res
}

func Close() error {
	if baseConfigWatch != nil {
		baseConfigWatch.Close()
	}
	if resConfigWatch != nil {
		resConfigWatch.Close()
	}
	return nil
}

// CORS middleware
func corsMiddleware(ctx *gin.Context) {
	ctx.Header("Access-Control-Allow-Origin", "*")
	ctx.Header("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
	ctx.Header("Access-Control-Allow-Headers", "Content-Type, Authorization")
}

// AuthWithHttp handles HTTP authentication requests
func AuthWithHttp(ctx *gin.Context, req *common.HttpKnockRequest, helper *plugins.HttpServerPluginHelper) (ackMsg *common.ServerKnockAckMsg, err error) {
	if helper == nil {
		return nil, fmt.Errorf("AuthWithHttp: helper is null")
	}

	resId := ctx.Query("resid")
	action := ctx.Query("action")
	if len(resId) > 0 && strings.Contains(resId, "|") {
		params := strings.Split(resId, "|")
		resId = params[0]
		if len(params) > 1 {
			action = params[1]
		}
	}

	res := findResource(resId)
	if res == nil || len(res.Resources) == 0 {
		ackMsg = nil
		err = common.ErrResourceNotFound
		log.Error("resource error: %v", err)
		ctx.String(http.StatusOK, "{\"errMsg\": \"resource error: %v\"}", err)
		return
	}

	corsMiddleware(ctx)

	switch {
	case strings.EqualFold(action, "otpvalid"):
		// OTP code validation (manual input)
		ackMsg, err = authOTPCode(ctx, req, res, helper)

	case strings.EqualFold(action, "qrvalid"):
		// QR code authentication (after mobile scan)
		ackMsg, err = authQRCode(ctx, req, res, helper)

	case strings.EqualFold(action, "login"):
		// Show login page
		ackMsg, err = showLoginPage(ctx, req, res, helper)

	case strings.EqualFold(action, "generate"):
		// Generate QR code
		HandleQRGenerate(ctx, resId)
		ackMsg, err = nil, nil

	case strings.EqualFold(action, "verify"):
		// Mobile device QR verification
		HandleQRVerify(ctx)
		ackMsg, err = nil, nil

	case strings.EqualFold(action, "status"):
		// Check QR code status (polling from web browser)
		HandleQRStatus(ctx)
		ackMsg, err = nil, nil

	case strings.EqualFold(action, "scan"):
		// Mobile device scan notification
		HandleQRScan(ctx)
		ackMsg, err = nil, nil

	default:
		ackMsg = nil
		err = fmt.Errorf("action invalid")
	}
	return
}

// showLoginPage displays the authenticator login page
func showLoginPage(ctx *gin.Context, req *common.HttpKnockRequest, res *common.ResourceData, helper *plugins.HttpServerPluginHelper) (*common.ServerKnockAckMsg, error) {
	_ = helper

	title := "OTP Authentication"
	if res.ExInfo != nil {
		if t, ok := res.ExInfo["Title"].(string); ok {
			title = t
		}
	}

	// The shared OTP seed is only published when the operator opted in.
	otpSecretKey := ""
	if conf := baseConf; conf != nil && conf.ShowOTPSecret {
		otpSecretKey = conf.OTPSecretKey
	}

	ctx.HTML(http.StatusOK, "authenticator/authenticator_login.html", gin.H{
		"title":        title,
		"nhpServer":    hostname,
		"aspId":        name, // Use plugin name as aspId
		"resId":        res.ResourceId,
		"otpSecretKey": otpSecretKey,
	})

	return nil, nil
}

// authOTPCode validates OTP code entered manually
func authOTPCode(ctx *gin.Context, req *common.HttpKnockRequest, res *common.ResourceData, helper *plugins.HttpServerPluginHelper) (*common.ServerKnockAckMsg, error) {
	otpCode := ctx.Query("otpCode")

	if otpCode == "" {
		log.Info("OTP auth failed: missing otpCode")
		ctx.JSON(http.StatusOK, gin.H{
			"success": false,
			"errMsg":  "missing otpCode",
		})
		return nil, fmt.Errorf("missing otpCode")
	}

	// Verify OTP code using configured secret key
	conf := baseConf
	if conf == nil || conf.OTPSecretKey == "" {
		log.Error("OTP auth failed: OTPSecretKey not configured")
		ctx.JSON(http.StatusOK, gin.H{
			"success": false,
			"errMsg":  "OTP secret key not configured",
		})
		return nil, fmt.Errorf("OTPSecretKey not configured")
	}

	clientIP := otpClientIP(ctx, conf.trustedProxies)
	switch err := otpCheck(conf, clientIP, otpCode); {
	case errors.Is(err, errNoClientAddr):
		log.Error("OTP auth refused: %v (RemoteAddr %q)", err, ctx.Request.RemoteAddr)
		ctx.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"errMsg":  "Cannot determine the client address",
		})
		return nil, err
	case errors.Is(err, errOTPRateLimited):
		log.Warning("OTP auth refused for %s: too many failed attempts", clientIP)
		ctx.JSON(http.StatusTooManyRequests, gin.H{
			"success": false,
			"errMsg":  "Too many failed attempts, try again later",
		})
		return nil, err
	case errors.Is(err, errOTPInvalid):
		log.Info("OTP verification failed for %s: invalid OTP code", clientIP)
		ctx.JSON(http.StatusOK, gin.H{
			"success": false,
			"errMsg":  "Invalid OTP code",
		})
		return nil, err
	case errors.Is(err, errOTPUsed):
		log.Warning("OTP verification failed for %s: code already used", clientIP)
		ctx.JSON(http.StatusOK, gin.H{
			"success": false,
			"errMsg":  "OTP code already used, wait for the next code",
		})
		return nil, err
	case err != nil:
		// Fail closed: an error this switch does not know about must never
		// fall through to the knock.
		log.Error("OTP auth refused for %s: %v", clientIP, err)
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"errMsg":  "OTP verification failed",
		})
		return nil, err
	}

	log.Info("OTP authentication successful using configured secret key")

	// Open access for the address that authenticated, not for whatever
	// X-Forwarded-For claims: nhp-server fills req.SrcIp from gin's
	// ClientIP, which trusts that header from any peer.
	req.SrcIp = otpClientAddr(ctx, conf.trustedProxies)

	// Proceed with knock operation
	ackMsg, err := helper.AuthWithHttpCallbackFunc(req, res)
	if ackMsg == nil || ackMsg.ErrCode != common.ErrSuccess.ErrorCode() {
		log.Error("OTP knock failed. ackMsg is nil")
		ackMsg = &common.ServerKnockAckMsg{}
		ackMsg.ErrCode = common.ErrServerACOpsFailed.ErrorCode()
		if err != nil {
			ackMsg.ErrMsg = err.Error()
		}
	} else {
		log.Info("OTP knock succeeded.")
		ackMsg.ErrMsg = ""
		if len(res.RedirectUrl) > 0 {
			ackMsg.RedirectUrl = res.RedirectUrl
		}
	}
	ctx.JSON(http.StatusOK, ackMsg)
	return ackMsg, nil
}

// authQRCode handles authentication after QR code scan
func authQRCode(ctx *gin.Context, req *common.HttpKnockRequest, res *common.ResourceData, helper *plugins.HttpServerPluginHelper) (*common.ServerKnockAckMsg, error) {
	sessionId := ctx.Query("sessionId")
	if sessionId == "" {
		log.Info("QR auth failed: missing sessionId")
		return nil, fmt.Errorf("missing sessionId for QR authentication")
	}

	// Knock only for a session confirmed by HandleQRVerify, which checks the
	// configured OTP through the same limiter as otpvalid, for the resource
	// it was generated for, from the client that generated it, and only once.
	requester := otpClientIP(ctx, requestTrustedProxies())
	if requester == "" {
		log.Error("QR auth refused: %v (RemoteAddr %q)", errNoClientAddr, ctx.Request.RemoteAddr)
		ctx.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"errMsg":  "Cannot determine the client address",
		})
		return nil, errNoClientAddr
	}
	if err := GetQRAuthService().ConsumeConfirmedSession(sessionId, res.ResourceId, requester); err != nil {
		log.Warning("QR auth refused for sessionId %s: %v", sessionId, err)
		ctx.JSON(http.StatusOK, gin.H{
			"success": false,
			"errMsg":  err.Error(),
		})
		return nil, err
	}

	log.Info("QR authentication for sessionId: %s", sessionId)

	// Open access for the address the session is bound to; see authOTPCode.
	req.SrcIp = otpClientAddr(ctx, requestTrustedProxies())

	// Proceed with knock operation
	ackMsg, err := helper.AuthWithHttpCallbackFunc(req, res)
	if ackMsg == nil || ackMsg.ErrCode != common.ErrSuccess.ErrorCode() {
		log.Error("QR knock failed. ackMsg is nil")
		ackMsg = &common.ServerKnockAckMsg{}
		ackMsg.ErrCode = common.ErrServerACOpsFailed.ErrorCode()
		if err != nil {
			ackMsg.ErrMsg = err.Error()
		}
	} else {
		log.Info("QR knock succeeded.")
		ackMsg.ErrMsg = ""
		if len(res.RedirectUrl) > 0 {
			ackMsg.RedirectUrl = res.RedirectUrl
		}
	}
	ctx.JSON(http.StatusOK, ackMsg)
	return ackMsg, nil
}
