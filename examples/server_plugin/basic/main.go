package main

import (
	"fmt"
	"io"
	"net/http"
	"net/smtp"
	"net/url"
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
	ExampleUsername string
	ExamplePassword string
	// SMTP settings for OTP email delivery.
	SMTPHost     string `toml:"smtp_host"`
	SMTPPort     int    `toml:"smtp_port"`
	SMTPUsername string `toml:"smtp_username"`
	SMTPPassword string `toml:"smtp_password"`
	SMTPFrom     string `toml:"smtp_from"`
	SMTPSubject  string `toml:"smtp_subject"`
}

var (
	// Example Plugin Settings
	log           *nhplog.Logger
	pluginDirPath string
	hostname      string
	localIp       string
	localMac      string
)

var (
	name    = "example"
	version = "0.1.1"

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

	// load config
	fileNameBase := (filepath.Join(pluginDirPath, "etc", "config.toml"))
	if err := updateConfig(fileNameBase); err != nil {
		// ignore error
		_ = err
	}

	baseConfigWatch = utils.WatchFile(fileNameBase, func() {
		log.Info("base config: %s has been updated", fileNameBase)
		updateConfig(fileNameBase)
	})

	fileNameRes := filepath.Join(pluginDirPath, "etc", "resource.toml")
	if err := updateResource(fileNameRes); err != nil {
		// ignore error
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
		log.Error("failed to read base config: %v", err)
	}

	var conf config
	if err := toml.Unmarshal(content, &conf); err != nil {
		log.Error("failed to unmarshal base config: %v", err)
	}

	baseConf = &conf
	return err
}

func updateResource(file string) (err error) {
	utils.CatchPanicThenRun(func() {
		err = errLoadConfig
	})

	content, err := os.ReadFile(file)
	if err != nil {
		log.Error("failed to read resource config: %v", err)
	}

	resourceMapMutex.Lock()
	defer resourceMapMutex.Unlock()

	resourceMap = make(common.ResourceGroupMap)
	if err := toml.Unmarshal(content, &resourceMap); err != nil {
		log.Error("failed to unmarshal resource config: %v", err)
	}

	// res is pointer so we can update its fields
	for resId, res := range resourceMap {
		res.AuthServiceId = name
		res.ResourceId = resId
	}

	return err
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

func findResource(resId string) *common.ResourceData {
	resourceMapMutex.Lock()
	defer resourceMapMutex.Unlock()

	res, found := resourceMap[resId]
	if found {
		return res
	}
	return nil
}

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
	case strings.EqualFold(action, "valid"):
		ackMsg, err = authRegular(ctx, req, res, helper)

	case strings.EqualFold(action, "login"):
		ackMsg, err = authAndShowLogin(ctx, req, res, helper)

	default:
		ackMsg = nil
		err = fmt.Errorf("action invalid")
	}
	return
}

func authAndShowLogin(ctx *gin.Context, req *common.HttpKnockRequest, res *common.ResourceData, helper *plugins.HttpServerPluginHelper) (*common.ServerKnockAckMsg, error) {
	_ = helper

	if res.ExInfo == nil {
		log.Error("extra login info not available")
		ctx.String(http.StatusOK, "{\"errMsg\": \"extra login info not available\"}")
		return nil, fmt.Errorf("extra login info not available")
	}

	ctx.HTML(http.StatusOK, "example/example_login.html", gin.H{
		"title":     res.ExInfo["Title"].(string),
		"nhpServer": hostname,
		"aspId":     req.AuthServiceId,
		"resId":     res.ResourceId,
	})

	return nil, nil
}

func authRegular(ctx *gin.Context, req *common.HttpKnockRequest, res *common.ResourceData, helper *plugins.HttpServerPluginHelper) (*common.ServerKnockAckMsg, error) {
	var err error
	username, _ := url.QueryUnescape(ctx.Query("username"))
	password, _ := url.QueryUnescape(ctx.Query("password"))

	if !strings.EqualFold(username, baseConf.ExampleUsername) || !strings.EqualFold(password, baseConf.ExamplePassword) {
		log.Info("Authenticating user: %s failed!", username)
		return nil, fmt.Errorf("user or password is incorrect")
	}

	// interact with udp server for ac operation
	ackMsg, err := helper.AuthWithHttpCallbackFunc(req, res)
	if ackMsg == nil || ackMsg.ErrCode != common.ErrSuccess.ErrorCode() {
		log.Error("knock failed. ackMsg is nil")
		ackMsg = &common.ServerKnockAckMsg{}
		ackMsg.ErrCode = common.ErrServerACOpsFailed.ErrorCode()
		ackMsg.ErrMsg = err.Error()
	} else {
		log.Info("knock succeeded.")
		ackMsg.ErrMsg = ""
		// assign the redirect url to the ackMsg
		if len(res.RedirectUrl) == 0 {
			log.Error("RedirectUrl is not provided.")
		} else {
			ackMsg.RedirectUrl = res.RedirectUrl
		}

		// set cookies
		// note that a dot in domain prefix used to make a difference, but now it doesn't (RFC6265).
		// The cookie will be sent to any subdomain of the specified domain, with or without the leading dot.
		singleHost := len(ackMsg.ACTokens) == 1
		for resName, token := range ackMsg.ACTokens {
			if singleHost {
				ctx.SetCookie(
					"nhp-token",            // Name
					url.QueryEscape(token), // Value
					int(res.OpenTime),      // MaxAge - use the knock interval time
					"/",                    // Path
					res.CookieDomain,       // Domain
					true,                   // Secure - if true, this cookie will only be sent on https, not http
					true)                   // HttpOnly - if true, this cookie will only be sent on http(s)
			} else {
				domain := strings.Split(ackMsg.ResourceHost[resName], ":")[0]
				ctx.SetCookie(
					"nhp-token"+"/"+resName, // Name
					url.QueryEscape(token),  // Value
					int(res.OpenTime),       // MaxAge - use the knock interval time
					"/",                     // Path
					domain,                  // Domain
					true,                    // Secure - if true, this cookie will only be sent on https, not http
					true)                    // HttpOnly - if true, this cookie will only be sent on http(s)
			}
			log.Info("ctx.SetCookie.")
		}
	}
	ctx.JSON(http.StatusOK, ackMsg)
	return ackMsg, nil
}

func AuthWithNHP(req *common.NhpAuthRequest, helper *plugins.NhpServerPluginHelper) (ackMsg *common.ServerKnockAckMsg, err error) {
	ackMsg = req.Ack
	if helper == nil {
		return ackMsg, fmt.Errorf("AuthWithNHP: helper is null")
	}

	var found bool
	var res *common.ResourceData
	resourceMapMutex.Lock()
	res, found = resourceMap[req.Msg.ResourceId]
	resourceMapMutex.Unlock()

	if !found || len(res.Resources) == 0 {
		err = common.ErrResourceNotFound
		ackMsg.ErrCode = common.ErrResourceNotFound.ErrorCode()
		ackMsg.ErrMsg = err.Error()
		return
	}

	// there is no backend auth in this plugin, fail the request if SkipAuth is false
	if !res.SkipAuth {
		err = common.ErrBackendAuthRequired
		ackMsg.ErrCode = common.ErrBackendAuthRequired.ErrorCode()
		ackMsg.ErrMsg = err.Error()
		return
	}

	// skip backend auth and continue with AC operations
	log.Info("agent user [%s]: skip auth", req.Msg.UserId)
	ackMsg.OpenTime = res.OpenTime

	// PART III: request ac operation for each resource and block for response
	ackMsg, err = helper.AuthWithNhpCallbackFunc(req, res)

	return ackMsg, err
}

// ── Plugin metadata exports ──────────────────────────────────────────────

func Signature() string {
	return name + "/" + version
}

func ExportedData() *plugins.PluginParamsOut {
	return &plugins.PluginParamsOut{}
}

// ── OTP and registration ─────────────────────────────────────────────────

// RequestOTP generates a one-time password, sends it via email, and stores
// it via the server's key store.
func RequestOTP(req *common.NhpOTPRequest, helper *plugins.NhpServerPluginHelper) error {
	if helper == nil || helper.GenerateOTPFunc == nil {
		return fmt.Errorf("RequestOTP: keystore helper not available")
	}

	// Use server-configured OTP TTL (default 300s = 5 min).
	ttl := helper.OTPTTLSeconds
	if ttl <= 0 {
		ttl = 300
	}
	otpCode, err := helper.GenerateOTPFunc(req.Msg.UserId, req.Msg.DeviceId, ttl)
	if err != nil {
		log.Error("RequestOTP: generate otp failed: %v", err)
		return err
	}

	// Send OTP via email.
	to := req.Msg.UserData["email"]
	emailAddr, ok := to.(string)
	if !ok || emailAddr == "" {
		emailAddr = req.Msg.UserId // fallback: use userId as email
		log.Warning("RequestOTP: no email in UserData, using userId as email recipient")
	}

	if err := sendOTPEmail(emailAddr, otpCode); err != nil {
		log.Error("RequestOTP: send email failed: %v", err)
		return err
	}

	log.Info("RequestOTP: otp sent to %s for user=%s device=%s", emailAddr, req.Msg.UserId, req.Msg.DeviceId)
	return nil
}

// RegisterAgent validates the OTP and registers the agent's public key.
func RegisterAgent(req *common.NhpRegisterRequest, helper *plugins.NhpServerPluginHelper) (*common.ServerRegisterAckMsg, error) {
	ack := req.Ack
	if ack == nil {
		ack = &common.ServerRegisterAckMsg{}
	}

	if helper == nil || helper.ValidateOTPFunc == nil || helper.RegisterKeyFunc == nil {
		err := fmt.Errorf("RegisterAgent: keystore helper not available")
		ack.ErrCode = common.ErrAgentKeyStoreError.ErrorCode()
		ack.ErrMsg = err.Error()
		return ack, err
	}

	// Step 1: validate OTP.
	if err := helper.ValidateOTPFunc(req.Msg.UserId, req.Msg.DeviceId, req.Msg.OTP); err != nil {
		log.Error("RegisterAgent: otp validation failed for user=%s: %v", req.Msg.UserId, err)
		ack.ErrCode = common.ErrorToErrorCode(err)
		ack.ErrMsg = common.ErrorToString(err)
		return ack, err
	}

	// Step 2: register the agent's public key.
	if err := helper.RegisterKeyFunc(req.Msg.UserId, req.Msg.DeviceId, req.PublicKey); err != nil {
		log.Error("RegisterAgent: register key failed for user=%s: %v", req.Msg.UserId, err)
		ack.ErrCode = common.ErrorToErrorCode(err)
		ack.ErrMsg = common.ErrorToString(err)
		return ack, err
	}

	ack.ErrCode = common.ErrSuccess.ErrorCode()
	ack.AuthServiceId = req.Msg.AuthServiceId
	log.Info("RegisterAgent: registered user=%s device=%s", req.Msg.UserId, req.Msg.DeviceId)
	return ack, nil
}

// ListService returns the list of available services for the agent.
func ListService(req *common.NhpListRequest, helper *plugins.NhpServerPluginHelper) (*common.ServerListResultMsg, error) {
	ack := req.Ack
	if ack == nil {
		ack = &common.ServerListResultMsg{}
	}

	resourceMapMutex.Lock()
	defer resourceMapMutex.Unlock()

	if ack.ListResults == nil {
		ack.ListResults = make(map[string]any)
	}
	for resId := range resourceMap {
		ack.ListResults[resId] = nil
	}

	ack.ErrCode = common.ErrSuccess.ErrorCode()
	return ack, nil
}

// ── Email helper ─────────────────────────────────────────────────────────

func sendOTPEmail(to, code string) error {
	if baseConf == nil || baseConf.SMTPHost == "" {
		// No SMTP configured — log the OTP for development/demo use.
		log.Info("OTP CODE for %s: %s (SMTP not configured, printed to log)", to, code)
		return nil
	}

	subject := baseConf.SMTPSubject
	if subject == "" {
		subject = "Your OpenNHP Verification Code"
	}

	body := fmt.Sprintf("Subject: %s\r\n", subject)
	body += "MIME-Version: 1.0\r\n"
	body += "Content-Type: text/plain; charset=\"UTF-8\"\r\n"
	body += "\r\n"
	body += fmt.Sprintf("Your verification code is: %s\r\n", code)
	body += fmt.Sprintf("This code will expire in 5 minutes.\r\n")

	addr := fmt.Sprintf("%s:%d", baseConf.SMTPHost, baseConf.SMTPPort)

	var auth smtp.Auth
	if baseConf.SMTPUsername != "" {
		auth = smtp.PlainAuth("", baseConf.SMTPUsername, baseConf.SMTPPassword, baseConf.SMTPHost)
	}

	from := baseConf.SMTPFrom
	if from == "" {
		from = "noreply@opennhp.org"
	}

	return smtp.SendMail(addr, auth, from, []string{to}, []byte(body))
}

func corsMiddleware(ctx *gin.Context) {
	originResource := ctx.Request.Header.Get("Origin")

	if originResource != "" {
		// HTTP headers for CORS
		ctx.Writer.Header().Set("Access-Control-Allow-Origin", originResource) // allow cross-origin resource sharing
	}

	ctx.Next()
}

func main() {

}
