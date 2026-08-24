// Package main — demo-app-plugin is the nhp-server plugin that bridges
// the standalone Demo App to the NHP protocol.
//
// On the inbound side (server -> demo app):
//
//	RequestOTP   -> generates a one-time password via helper, then POSTs
//	               it to the Demo App's /internal/nhp/otp-deliver for
//	               email delivery.
//	RegisterAgent -> validates the OTP, registers the public key via
//	               helper.RegisterKeyFunc, then POSTs to the Demo App's
//	               /internal/nhp/mark-registered so the user row's
//	               nhp_registered_at timestamp is updated.
//	ListService  -> returns the static resource list from the plugin
//	               config (mirrors resource.toml).
//	AuthWithNHP  -> delegates to helper.AuthWithNhpCallbackFunc to
//	               trigger the AC operation.
//	AuthWithHttp -> not supported by the demo (we don't render HTML
//	               login pages from the plugin); returns an error.
//
// On the outbound side, the Demo App's /internal API authenticates
// callers with a shared X-Plugin-Api-Key header — see DemoAppBaseUrl
// and InternalApiKey in etc/config.toml.
package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"sync"
	"time"

	toml "github.com/pelletier/go-toml/v2"

	"github.com/gin-gonic/gin"

	"github.com/OpenNHP/opennhp/nhp/common"
	nhplog "github.com/OpenNHP/opennhp/nhp/log"
	"github.com/OpenNHP/opennhp/nhp/plugins"
)

// ── Plugin metadata ───────────────────────────────────────────────────

var (
	name    = "demo-app"
	version = "0.1.0"
)

// Version / Signature / ExportedData / Init / Close are the nhp-server
// plugin-loader entry points.
func Version() string { return fmt.Sprintf("%s v%s", name, version) }

func Signature() string { return name + "/" + version }

func ExportedData() *plugins.PluginParamsOut { return &plugins.PluginParamsOut{} }

// Plugin config.
type config struct {
	DemoAppBaseUrl string `toml:"DemoAppBaseUrl"`
	InternalApiKey string `toml:"InternalApiKey"`
	AuthServiceId  string `toml:"AuthServiceId"`

	// Resources is intentionally simple: each entry is a flat map keyed
	// by resId with the fields the js-agent dashboard needs to render a
	// card and call knockResource().
	Resources map[string]resourceInfo `toml:"Resources"`
}

type resourceInfo struct {
	ResId      string `toml:"ResId"`
	ACId       string `toml:"ACId"`
	Hostname   string `toml:"Hostname"`
	OpenTime   uint32 `toml:"OpenTime"`
	SkipAuth   bool   `toml:"SkipAuth"`
	RedirectUrl string `toml:"RedirectUrl"`
}

var (
	log        *nhplog.Logger
	pluginDir  string
	baseConf   *config
	configLock sync.RWMutex
)

// Init wires the plugin's runtime state from the params nhp-server
// passes in. We also kick off a config file watcher so changes to
// etc/config.toml take effect without a server restart.
func Init(in *plugins.PluginParamsIn) error {
	if in.PluginDirPath != nil {
		pluginDir = *in.PluginDirPath
	}
	if in.Log != nil {
		log = in.Log
	}

	if pluginDir == "" {
		// Fallback to executable directory when nhp-server didn't pass
		// one in (e.g., local testing). The plugin loader guarantees a
		// value under normal operation.
		exe, _ := os.Executable()
		pluginDir = filepath.Dir(exe)
	}

	confPath := filepath.Join(pluginDir, "etc", "config.toml")
	if err := loadConfig(confPath); err != nil {
		log.Error("demo-app-plugin: load config failed: %v", err)
		return err
	}

	// The nhp-server already watches etc/ for hot-reload, but we
	// re-read on signal in case it's running standalone. The watcher
	// implementation lives in nhp/utils but exposing it here would mean
	// linking against utils; the server-side watcher handles reload.
	return nil
}

func loadConfig(path string) error {
	content, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("read %s: %w", path, err)
	}
	var c config
	if err := toml.Unmarshal(content, &c); err != nil {
		return fmt.Errorf("parse %s: %w", path, err)
	}
	configLock.Lock()
	baseConf = &c
	configLock.Unlock()
	return nil
}

func Close() error { return nil }

// ── PluginHandler implementations ─────────────────────────────────────

// RequestOTP generates an OTP via the keystore helper and forwards it
// to the Demo App for email delivery. The Demo App is responsible for
// actually sending the email.
//
// The recipient email is taken from req.Msg.UserData["email"]. If the
// caller didn't supply one we fall back to looking the user up by
// userId on the Demo App side via /api/internal/users/{userId}/email.
// For the password-registration flow the browser sets the email, so
// the lookup is mostly used by OIDC re-keying.
func RequestOTP(req *common.NhpOTPRequest, helper *plugins.NhpServerPluginHelper) error {
	if helper == nil || helper.GenerateOTPFunc == nil {
		return fmt.Errorf("demo-app-plugin: keystore helper not available")
	}

	ttl := helper.OTPTTLSeconds
	if ttl <= 0 {
		ttl = 300
	}
	otpCode, err := helper.GenerateOTPFunc(req.Msg.UserId, req.Msg.DeviceId, req.PublicKey, ttl)
	if err != nil {
		return fmt.Errorf("generate otp: %w", err)
	}

	email, _ := req.Msg.UserData["email"].(string)
	if email == "" {
		// Fallback: ask the Demo App for the user's email.
		var err error
		email, err = lookupEmail(req.Msg.UserId)
		if err != nil || email == "" {
			return fmt.Errorf("no email supplied and lookup failed for user %q: %v", req.Msg.UserId, err)
		}
	}

	go postToDemoApp("/internal/nhp/otp-deliver", map[string]any{
		"userId":  req.Msg.UserId,
		"otpCode": otpCode,
		"email":   email,
	})
	return nil
}

// RegisterAgent validates the OTP, persists the public key, then asks
// the Demo App to mark the corresponding user row as NHP-registered.
//
// We return early with an error ACK on failure but still emit a best-
// effort callback to the Demo App so the dashboard can display the
// reason (instead of being stuck on "registration complete").
func RegisterAgent(req *common.NhpRegisterRequest, helper *plugins.NhpServerPluginHelper) (*common.ServerRegisterAckMsg, error) {
	ack := req.Ack
	if ack == nil {
		ack = &common.ServerRegisterAckMsg{}
	}

	if helper == nil || helper.ValidateOTPFunc == nil || helper.RegisterKeyFunc == nil {
		err := fmt.Errorf("demo-app-plugin: keystore helper not available")
		ack.ErrCode = common.ErrAgentKeyStoreError.ErrorCode()
		ack.ErrMsg = err.Error()
		return ack, err
	}

	if err := helper.ValidateOTPFunc(req.Msg.UserId, req.Msg.DeviceId, req.Msg.OTP, req.PublicKey); err != nil {
		log.Error("RegisterAgent: otp validation failed for user=%s: %v", req.Msg.UserId, err)
		ack.ErrCode = common.ErrorToErrorCode(err)
		ack.ErrMsg = common.ErrorToString(err)
		return ack, err
	}

	if err := helper.RegisterKeyFunc(req.Msg.UserId, req.Msg.DeviceId, req.PublicKey, req.CipherScheme); err != nil {
		log.Error("RegisterAgent: register key failed for user=%s: %v", req.Msg.UserId, err)
		ack.ErrCode = common.ErrorToErrorCode(err)
		ack.ErrMsg = common.ErrorToString(err)
		return ack, err
	}

	if helper.GetAgentKeyExpiryFunc != nil {
		if _, exp, err := helper.GetAgentKeyExpiryFunc(req.Msg.UserId, req.Msg.DeviceId); err == nil {
			ack.ExpiresAt = exp
		} else {
			log.Warning("RegisterAgent: read expiry failed: %v", err)
		}
	}

	// Tell the Demo App that registration succeeded. The callback is
	// best-effort — failure doesn't roll back the key registration, but
	// it does mean the user row's nhp_registered_at flag stays NULL
	// until they re-register.
	go postToDemoApp("/internal/nhp/mark-registered", map[string]any{
		"userId":    req.Msg.UserId,
		"publicKey": req.PublicKey,
	})

	ack.ErrCode = common.ErrSuccess.ErrorCode()
	ack.AuthServiceId = req.Msg.AuthServiceId
	return ack, nil
}

// ListService returns the static resource catalog read from the plugin
// config. The dashboard renders one card per entry.
func ListService(req *common.NhpListRequest, helper *plugins.NhpServerPluginHelper) (*common.ServerListResultMsg, error) {
	ack := req.Ack
	if ack == nil {
		ack = &common.ServerListResultMsg{}
	}
	configLock.RLock()
	resources := baseConf.Resources
	configLock.RUnlock()

	ack.ListResults = make(map[string]any, len(resources))
	for resId, r := range resources {
		ack.ListResults[resId] = map[string]any{
			"acId":     r.ACId,
			"hostname": r.Hostname,
			"openTime": r.OpenTime,
			"skipAuth": r.SkipAuth,
		}
	}
	ack.ErrCode = common.ErrSuccess.ErrorCode()
	return ack, nil
}

// AuthWithNHP doesn't do any user-facing authentication itself; the
// browser already proved possession of the private key via Noise. We
// just delegate to helper.AuthWithNhpCallbackFunc, which triggers the
// AC operation against the resource and blocks until NHP_ART lands.
func AuthWithNHP(req *common.NhpAuthRequest, helper *plugins.NhpServerPluginHelper) (*common.ServerKnockAckMsg, error) {
	if helper == nil {
		return req.Ack, fmt.Errorf("demo-app-plugin: helper is null")
	}

	configLock.RLock()
	r, ok := baseConf.Resources[req.Msg.ResourceId]
	configLock.RUnlock()
	if !ok {
		req.Ack.ErrCode = common.ErrResourceNotFound.ErrorCode()
		req.Ack.ErrMsg = "resource not found"
		return req.Ack, common.ErrResourceNotFound
	}

	// Mimic the basic plugin's behavior: respect SkipAuth so demos can
	// skip the (nonexistent) backend auth step.
	if !r.SkipAuth {
		req.Ack.ErrCode = common.ErrBackendAuthRequired.ErrorCode()
		req.Ack.ErrMsg = "backend auth required"
		return req.Ack, common.ErrBackendAuthRequired
	}

	// Build a ResourceData for the callback. The exact fields the
	// callback consults are OpenTime and Resources; we populate them
	// from the plugin config.
	resData := &common.ResourceData{
		ResourceGroup: common.ResourceGroup{
			AuthServiceId: baseConf.AuthServiceId,
			ResourceId:    r.ResId,
			OpenTime:      r.OpenTime,
			Resources: map[string]*common.ResourceInfo{
				r.ResId: {
					ACId:     r.ACId,
					Hostname: r.Hostname,
				},
			},
		},
		RedirectUrl: r.RedirectUrl,
	}
	return helper.AuthWithNhpCallbackFunc(req, resData)
}

// AuthWithHttp is a no-op for the demo. The browser-side flow uses the
// NHP protocol; HTTP-auth would defeat the network-hiding purpose.
func AuthWithHttp(ctx *gin.Context, req *common.HttpKnockRequest, helper *plugins.HttpServerPluginHelper) (*common.ServerKnockAckMsg, error) {
	return nil, fmt.Errorf("demo-app-plugin: HTTP auth is not supported; use the browser NHP flow")
}

// ── Internal API helpers ──────────────────────────────────────────────

// postToDemoApp is fire-and-forget. The plugin goroutine logs errors
// but doesn't propagate them — nhp-server has already returned a 200
// to the agent by the time this runs.
func postToDemoApp(path string, body map[string]any) {
	configLock.RLock()
	baseURL := baseConf.DemoAppBaseUrl
	key := baseConf.InternalApiKey
	configLock.RUnlock()

	if baseURL == "" || key == "" {
		log.Error("demo-app-plugin: missing DemoAppBaseUrl/InternalApiKey")
		return
	}

	payload, err := json.Marshal(body)
	if err != nil {
		log.Error("demo-app-plugin: marshal: %v", err)
		return
	}
	client := &http.Client{Timeout: 5 * time.Second}
	req, err := http.NewRequest("POST", baseURL+path, bytes.NewReader(payload))
	if err != nil {
		log.Error("demo-app-plugin: new request: %v", err)
		return
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Plugin-Api-Key", key)

	resp, err := client.Do(req)
	if err != nil {
		log.Error("demo-app-plugin: POST %s: %v", path, err)
		return
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 300 {
		buf, _ := io.ReadAll(resp.Body)
		log.Error("demo-app-plugin: POST %s -> %d: %s", path, resp.StatusCode, string(buf))
	}
}

func lookupEmail(userId string) (string, error) {
	configLock.RLock()
	baseURL := baseConf.DemoAppBaseUrl
	key := baseConf.InternalApiKey
	configLock.RUnlock()

	if baseURL == "" || key == "" {
		return "", fmt.Errorf("plugin not configured")
	}
	client := &http.Client{Timeout: 3 * time.Second}
	url := fmt.Sprintf("%s/internal/nhp/lookup-email?userId=%s", baseURL, userId)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return "", err
	}
	req.Header.Set("X-Plugin-Api-Key", key)
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return "", fmt.Errorf("lookup-email status %d", resp.StatusCode)
	}
	var out struct {
		Email string `json:"email"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		return "", err
	}
	return out.Email, nil
}

func main() {
	// Loaded as a Go plugin; main() is unused but required.
}
