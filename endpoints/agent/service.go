package agent

import (
	"bytes"
	"context"
	"crypto/ed25519"
	"crypto/sha256"
	"crypto/tls"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	toml "github.com/pelletier/go-toml/v2"

	"github.com/OpenNHP/opennhp/nhp/core/verifier"
	wasmEngine "github.com/OpenNHP/opennhp/nhp/core/wasm/engine"
	"github.com/OpenNHP/opennhp/nhp/log"
	utils "github.com/OpenNHP/opennhp/nhp/utils"
)

var (
	routes = struct {
		sync.RWMutex
		m map[string]map[string]gin.HandlerFunc // method -> path -> handler
	}{m: make(map[string]map[string]gin.HandlerFunc)}

	serviceApiPrefix = "/api/v1"
)

func (a *UdpAgent) CheckAgentSafeOrNot(targetPaths ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		path := c.Request.URL.Path
		for _, target := range targetPaths {
			if strings.HasPrefix(path, target) {
				if !a.safeTee.Load() {
					c.JSON(http.StatusForbidden, gin.H{"error": "TEE in which is Agent is not safe"})
					c.Abort()
					return
				}
			}
		}
	}
}

// PrepareDHPWebConsole loads the console configuration and bearer token,
// loads the TLS certificate and binds the listener, without serving yet.
// Call it before Start so a bad console.toml, a missing certificate or a
// port already in use stops the agent before it touches the network.
func (a *UdpAgent) PrepareDHPWebConsole() error {
	etcDir := filepath.Join(ExeDirPath, "etc")

	conf, err := loadConsoleConfig(etcDir)
	if err != nil {
		return fmt.Errorf("DHP console: %w", err)
	}
	token, notices, err := loadOrCreateConsoleToken(etcDir)
	if err != nil {
		return fmt.Errorf("DHP console: %w", err)
	}
	cert, err := tls.LoadX509KeyPair(
		filepath.Join(etcDir, "certs", "server.crt"),
		filepath.Join(etcDir, "certs", "server.key"),
	)
	if err != nil {
		return fmt.Errorf("DHP console: load TLS certificate: %w", err)
	}
	ln, err := net.Listen("tcp", conf.ListenAddr)
	if err != nil {
		return fmt.Errorf("DHP console: listen on %s: %w", conf.ListenAddr, err)
	}

	a.consoleConf = conf
	a.consoleNotices = notices
	a.consoleListener = ln
	a.consoleServer = &http.Server{
		Handler:           a.newConsoleRouter(token),
		ReadHeaderTimeout: 10 * time.Second,
		// Reap idle keep-alive connections. ReadTimeout/WriteTimeout are
		// deliberately unset: they would cut off slow TA uploads and
		// long-running TA calls.
		IdleTimeout: 2 * time.Minute,
		TLSConfig: &tls.Config{
			MinVersion:   tls.VersionTLS12,
			Certificates: []tls.Certificate{cert},
		},
	}
	return nil
}

// ServeDHPWebConsole starts serving on the listener bound by
// PrepareDHPWebConsole. Call it once the agent has started, so no request
// is handled against a half-initialized agent.
func (a *UdpAgent) ServeDHPWebConsole() error {
	if a.consoleServer == nil || a.consoleListener == nil {
		return errors.New("DHP console: PrepareDHPWebConsole has not been called")
	}

	id := time.Now().Format("2006-01-02 15:04:05")
	log.Info("==================================================")
	log.Info("===  DHP Web Console (%s) started on %s  ===", id, a.consoleListener.Addr())
	log.Info("==================================================")
	for _, n := range a.consoleNotices {
		n.log()
	}
	if len(a.consoleConf.trustedTAKeys) == 0 {
		log.Warning("DHP console: no TrustedTAPublicKeys configured, trusted application registration is disabled")
	}

	srv, ln := a.consoleServer, a.consoleListener
	go func() {
		if err := srv.ServeTLS(ln, "", ""); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Error("DHP console stopped: %v", err)
		}
	}()
	return nil
}

// CloseDHPWebConsole stops the console: it stops accepting connections and
// waits up to consoleShutdownTimeout for in-flight requests to finish
// before closing their connections. It does not by itself guarantee that
// no handler is still running afterwards — StopDHP adds that guarantee for
// the operations that restart the agent. Use it directly only when the
// agent has not been started (startup failure after PrepareDHPWebConsole).
func (a *UdpAgent) CloseDHPWebConsole() {
	if a.consoleServer != nil {
		ctx, cancel := context.WithTimeout(context.Background(), consoleShutdownTimeout)
		defer cancel()
		if err := a.consoleServer.Shutdown(ctx); err != nil {
			_ = a.consoleServer.Close()
		}
	}
	if a.consoleListener != nil {
		_ = a.consoleListener.Close()
	}
}

// newConsoleRouter builds the console router. Every route, including the
// dynamic TA routes dispatched from NoRoute, sits behind the bearer-token
// middleware; the TEE posture check runs only after authentication.
func (a *UdpAgent) newConsoleRouter(token string) *gin.Engine {
	gin.SetMode(gin.ReleaseMode)
	router := gin.Default()
	// gin's path-fixing redirects run before any middleware, so an
	// unauthenticated "POST /api/v1/agent/restart/" would get a 307 to the
	// real route instead of a 401, letting callers map routes without a
	// token. Unauthenticated requests must always get 401.
	router.RedirectTrailingSlash = false
	router.RedirectFixedPath = false

	router.Use(consoleAuthMiddleware(token))
	router.Use(a.CheckAgentSafeOrNot(taApiPrefix))

	router.POST(fmt.Sprintf("%s/%s", serviceApiPrefix, "config/server"), a.configServer)
	router.GET(fmt.Sprintf("%s/%s", serviceApiPrefix, "config/server"), a.getServerConfig)

	router.GET(fmt.Sprintf("%s/%s", serviceApiPrefix, "key/agent"), a.getAgentPublicKey)
	router.GET(fmt.Sprintf("%s/%s", serviceApiPrefix, "key/tee"), a.getTeePublicKey)
	router.POST(fmt.Sprintf("%s/%s", serviceApiPrefix, "key/agent/rotate"), a.rotateAgentKey)
	router.POST(fmt.Sprintf("%s/%s", serviceApiPrefix, "key/tee/rotate"), a.rotateTeeKey)

	// State-changing: POST only, so a cross-site GET cannot trigger it.
	router.POST(fmt.Sprintf("%s/%s", serviceApiPrefix, "agent/restart"), a.restartAgent)

	router.GET(fmt.Sprintf("%s/%s", serviceApiPrefix, "status/agent"), a.getTeeStatus)

	router.GET(fmt.Sprintf("%s/%s", serviceApiPrefix, "attestation/tee"), a.getTeeAttestation)

	router.POST(fmt.Sprintf("%s/%s", taApiPrefix, "register"), a.registerTAService)

	// Dynamic route handler - this catches all requests and checks our dynamic routes
	router.NoRoute(func(c *gin.Context) {
		method := c.Request.Method
		path := c.Request.URL.Path

		routes.RLock()
		defer routes.RUnlock()

		// Check if we have a handler for this method and path
		if methodRoutes, ok := routes.m[method]; ok {
			if handler, ok := methodRoutes[path]; ok {
				handler(c)
				return
			}
		}

		// If no dynamic route found, return 404
		c.JSON(http.StatusNotFound, gin.H{"error": "Route not found"})
	})

	return router
}

// beginLifecycleOp serializes console operations that stop and restart the
// agent (restart, key rotation) with each other and with StopDHP. It
// returns false, after answering 503, once the agent is shutting down. On
// success the caller must defer endLifecycleOp.
func (a *UdpAgent) beginLifecycleOp(c *gin.Context) bool {
	a.consoleLifecycleMu.Lock()
	if a.consoleStopping {
		a.consoleLifecycleMu.Unlock()
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "agent is shutting down"})
		return false
	}
	return true
}

func (a *UdpAgent) endLifecycleOp() {
	a.consoleLifecycleMu.Unlock()
}

// StopDHP shuts down a DHP agent: it stops the console, waits for any
// in-flight restart or key rotation to finish (however long it takes —
// CloseDHPWebConsole's timeout only bounds connection draining), refuses
// new ones, and then stops the agent. Without this, a console-triggered
// RestartAgent could run Stop concurrently with the shutdown's Stop and
// close the same channels twice.
//
// Other in-flight console requests (e.g. a TA call) may still be running
// after the timeout; they do not stop or restart the agent and fail on
// their own once it is stopped.
func (a *UdpAgent) StopDHP() {
	a.CloseDHPWebConsole()

	a.consoleLifecycleMu.Lock()
	defer a.consoleLifecycleMu.Unlock()
	a.consoleStopping = true
	a.Stop()
}

// trustedTAKeys returns the operator keys allowed to sign trusted
// applications, or nil when the console config has not been loaded.
func (a *UdpAgent) trustedTAKeys() []ed25519.PublicKey {
	if a.consoleConf == nil {
		return nil
	}
	return a.consoleConf.trustedTAKeys
}

// verifyUploadedTA hashes the uploaded file straight from the multipart
// form and checks its signature, before anything is written into the TA
// directory, made executable or started. It returns the verified SHA-256
// digest, which is also the deduplication key for the TA.
func (a *UdpAgent) verifyUploadedTA(file *multipart.FileHeader, sigB64 string) ([]byte, error) {
	keys := a.trustedTAKeys()
	if len(keys) == 0 {
		// Defensive: registerTAService already refuses before parsing.
		return nil, errTANoTrustedKeys
	}

	f, err := file.Open()
	if err != nil {
		return nil, err
	}
	defer f.Close()

	h := sha256.New()
	if _, err := io.Copy(h, f); err != nil {
		return nil, err
	}
	digest := h.Sum(nil)

	if err := verifyTASignature(keys, digest, sigB64); err != nil {
		return nil, err
	}
	if a.consoleConf.isTARevoked(digest) {
		return nil, errTARevoked
	}
	return digest, nil
}

// isPlainPathElement reports whether name is a single, ordinary path
// element: non-empty, containing no path separator, and not "." or "..".
// (filepath.Base alone is not enough: it returns ".", ".." and "/"
// unchanged.)
func isPlainPathElement(name string) bool {
	return name != "" && name != "." && name != ".." &&
		!strings.ContainsRune(name, '/') && !strings.ContainsRune(name, filepath.Separator)
}

// lookupRegisteredTA resolves a dedup index entry to the stored binary. It
// only reports reuse when the entry is well formed and the stored file
// still exists and still hashes to the verified digest; anything else
// (missing, cleaned up, corrupt or tampered with) is treated as a fresh
// upload, whose index write then repairs the entry.
func lookupRegisteredTA(taDir, indexFile string, digest []byte) (uuid, fileName string, ok bool) {
	if _, err := os.Stat(indexFile); err != nil {
		return "", "", false
	}

	fileInfo, err := utils.LoadJsonFileAsStruct(indexFile)
	if err != nil {
		log.Warning("DHP console: ignoring unreadable trusted application index %s: %v", indexFile, err)
		return "", "", false
	}
	info, _ := fileInfo.(map[string]any)
	uuid, _ = info["uuid"].(string)
	fileName, _ = info["fileName"].(string)
	if !isPlainPathElement(uuid) || !isPlainPathElement(fileName) {
		log.Warning("DHP console: ignoring malformed trusted application index %s", indexFile)
		return "", "", false
	}

	// The stored binary is hashed here and executed later by
	// NewTrustApplication. Anyone who can write etc/ta between the two
	// could swap it; etc/ is trusted (it also holds the agent's keys), and
	// etc/ta is created owner-only to keep it that way.
	f, err := os.Open(filepath.Join(taDir, uuid, fileName))
	if err != nil {
		log.Warning("DHP console: trusted application index %s points to a missing binary: %v", indexFile, err)
		return "", "", false
	}
	defer f.Close()
	h := sha256.New()
	if _, err := io.Copy(h, f); err != nil || !bytes.Equal(h.Sum(nil), digest) {
		log.Warning("DHP console: stored binary for %s no longer matches its digest, re-registering", indexFile)
		return "", "", false
	}

	return uuid, fileName, true
}

func (a *UdpAgent) registerTAService(c *gin.Context) {
	// Refuse before reading the (up to maxTAUploadSize) body when no
	// signing keys are configured — the default — so the request cannot be
	// used to make the agent spool large uploads to disk for nothing.
	if len(a.trustedTAKeys()) == 0 {
		c.JSON(http.StatusForbidden, gin.H{"error": errTANoTrustedKeys.Error()})
		return
	}

	c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, maxTAUploadSize)

	// Parse explicitly: c.PostForm swallows parse errors, which would turn
	// an oversized upload into a misleading "missing parameters" 400.
	// Keep at most 8 MiB of the upload in memory; larger parts spill to
	// temporary files.
	if parseErr := c.Request.ParseMultipartForm(8 << 20); parseErr != nil {
		var tooLarge *http.MaxBytesError
		if errors.As(parseErr, &tooLarge) {
			c.JSON(http.StatusRequestEntityTooLarge, gin.H{"error": fmt.Sprintf("trusted application upload exceeds %d bytes", maxTAUploadSize)})
			return
		}
		c.JSON(http.StatusBadRequest, gin.H{"error": parseErr.Error()})
		return
	}

	description := c.PostForm("description")
	language := c.PostForm("language")
	entry := c.PostForm("entry")
	taName := c.PostForm("name")

	if taName == "" || description == "" || language == "" || entry == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid parameters, ta name, description, language, entry are required"})
		return
	}

	file, err := c.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	digest, verifyErr := a.verifyUploadedTA(file, c.PostForm("signature"))
	if verifyErr != nil {
		log.Warning("DHP console: rejected trusted application upload %q from %s: %v", file.Filename, c.Request.RemoteAddr, verifyErr)
		c.JSON(http.StatusForbidden, gin.H{"error": verifyErr.Error()})
		return
	}

	// Never let a client-supplied name escape the per-upload directory.
	fileName := filepath.Base(file.Filename)
	if !isPlainPathElement(fileName) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid file name"})
		return
	}

	taDir := filepath.Join(ExeDirPath, "etc", "ta")
	// Owner-only: only the agent itself needs to read or execute TAs.
	// MkdirAll leaves an existing directory's mode alone, so tighten an
	// etc/ta created by an older version (0755) explicitly.
	if mkdirErr := os.MkdirAll(taDir, 0o700); mkdirErr != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": mkdirErr.Error()})
		return
	}
	if chmodErr := os.Chmod(taDir, 0o700); chmodErr != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": chmodErr.Error()})
		return
	}

	// Deduplicate by the SHA-256 that was just verified, never by MD5: an
	// MD5 match could otherwise swap in a different (e.g. pre-signature,
	// unsigned) binary that was registered earlier. Index files are named
	// "sha256-<hex>", so legacy MD5-named entries are never matched. The
	// lookup happens before anything is written, so re-registering a known
	// TA costs no disk I/O.
	indexFile := filepath.Join(taDir, "sha256-"+hex.EncodeToString(digest))
	fileUuid, storedName, reuse := lookupRegisteredTA(taDir, indexFile, digest)
	if reuse {
		// Run the binary that is actually stored under the previous uuid.
		fileName = storedName
	} else {
		var uuidErr error
		fileUuid, uuidErr = utils.GenerateUUIDv4()
		if uuidErr != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": uuidErr.Error()})
			return
		}

		fullFilePath := filepath.Join(taDir, fileUuid, fileName)
		if saveErr := c.SaveUploadedFile(file, fullFilePath); saveErr != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": saveErr.Error()})
			return
		}

		if chmodErr := os.Chmod(fullFilePath, 0o700); chmodErr != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": chmodErr.Error()})
			return
		}
	}

	// save file information into the index file, no matter the file exists or not.
	if writeErr := utils.SaveStructAsJsonFile(indexFile, map[string]any{
		"fileName":    fileName,
		"name":        taName,
		"uuid":        fileUuid,
		"size":        file.Size,
		"description": description,
		"language":    language,
		"entry":       entry,
	}); writeErr != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": writeErr.Error()})
		return
	}

	ta, err := NewTrustApplication(fileUuid, language, filepath.Join(taDir, fileUuid, fileName))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	a.addTARoute(ta)

	c.JSON(http.StatusOK, ta.GetSupportedFunctions())
}

func (a *UdpAgent) addTARoute(ta *TrustedApplication) {
	routes.Lock()
	for _, function := range ta.Functions {
		if _, exists := routes.m[function.Method]; !exists {
			routes.m[function.Method] = make(map[string]gin.HandlerFunc)
		}

		if _, exists := routes.m[function.Method][function.Name]; !exists {
			routes.m[function.Method][function.Name] = a.callFunction
		}
	}
	routes.Unlock()
}

func (a *UdpAgent) callFunction(c *gin.Context) {
	path := c.Request.URL.Path

	var body map[string]any

	if err := c.BindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if _, exist := body["doId"]; !exist {
		c.JSON(http.StatusBadRequest, gin.H{"error": "doId is missing"})
		return
	}

	parts := strings.Split(path, "/")

	// url example: /api/v1/ta/<taId>/<function>
	function := parts[len(parts)-1]
	taId := parts[len(parts)-2]

	ccRes, err := a.StartConfidentialComputing(body["doId"].(string), taId, function, body)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, ccRes)
}

func (a *UdpAgent) getAgentPublicKey(c *gin.Context) {
	// GetAgentEcdh returns nil if the private key could not be resolved
	// (a sealed key with no usable passphrase). Start populates the cache
	// before this endpoint is served, so this should be unreachable — but
	// report it rather than dereferencing nil and relying on the handler
	// panic being recovered.
	ecdh := a.config.GetAgentEcdh()
	if ecdh == nil {
		log.Error("getAgentPublicKey: agent private key is unavailable")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "agent private key unavailable"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"publicKey": ecdh.PublicKeyBase64()})
}

func (a *UdpAgent) getTeePublicKey(c *gin.Context) {
	// GetTeeEcdh returns nil when TEEPrivateKeyBase64 is empty/unset — and
	// unlike the agent key that is genuinely reachable: updateDHPConfig
	// tolerates a missing dhp.toml, so a DHP agent that has never had
	// RotateTeeKey run hits this with no TEE key. Guard rather than deref nil.
	ecdh := a.config.GetTeeEcdh()
	if ecdh == nil {
		log.Error("getTeePublicKey: TEE private key is unavailable (run the DHP secret init)")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "TEE private key unavailable"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"publicKey": ecdh.PublicKeyBase64()})
}

func (a *UdpAgent) configServer(c *gin.Context) {
	var peers Peers

	if err := c.BindJSON(&peers); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// server.toml is the agent's trust anchor (nhp-server public keys and
	// addresses) and is hot-reloaded by the config watcher, so reject
	// anything the loader itself would reject instead of persisting it.
	if err := normalizeClusters(peers.Servers, log.Warning); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	for _, cfg := range peers.Servers {
		if _, err := buildCluster(cfg); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
	}

	// Encode fully before touching the file, so an encoding error cannot
	// leave server.toml truncated.
	var buf bytes.Buffer
	if err := toml.NewEncoder(&buf).Encode(peers); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Write in place (as before) and let the existing file watcher apply
	// it. A rename-based atomic replace is deliberately not used: on
	// kqueue platforms (macOS) it ends the watcher's loop, silently
	// disabling hot reload, and calling updateServerPeers from here would
	// race RestartAgent, which closes the watcher while it swaps the
	// device. What keeps the watcher from loading a half-written file is
	// its debounce: it reloads 100ms after the *last* event, and the whole
	// (small) file is written by a single WriteFile right after the
	// truncate. A truncated prefix can itself be valid TOML, so the
	// loader's validation alone would not catch it.
	fileName := filepath.Join(ExeDirPath, "etc", "server.toml")
	// The mode only applies if the file does not exist yet; an existing
	// server.toml keeps its permissions.
	if err := os.WriteFile(fileName, buf.Bytes(), 0o600); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	log.Info("DHP console: server peer config rewritten by %s (%d servers)", c.Request.RemoteAddr, len(peers.Servers))
	c.JSON(http.StatusOK, gin.H{"msg": "successfully configure nhp server info in agent"})
}

func (a *UdpAgent) getServerConfig(c *gin.Context) {
	var peers Peers

	fileName := filepath.Join(ExeDirPath, "etc", "server.toml")
	content, err := os.ReadFile(fileName)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	err = toml.Unmarshal(content, &peers)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, peers)
}

func (a *UdpAgent) restartAgent(c *gin.Context) {
	if !a.beginLifecycleOp(c) {
		return
	}
	defer a.endLifecycleOp()

	err := a.RestartAgent()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"msg": "successfully restart agent"})
}

func (a *UdpAgent) getTeeStatus(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"running":             a.IsRunning(),
		"attestationVerified": a.safeTee.Load(),
		"trustedByNHPServer":  a.trustedByNHPServer.Load(),
		"trustedByNHPDB":      a.trustedByNHPDB.Load(),
	})
}

func (a *UdpAgent) getTeeAttestation(c *gin.Context) {
	evidence, scheme, err := wasmEngine.GetEvidence()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	attestationVerifier, err := verifier.NewVerifier(evidence, scheme)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	measure := attestationVerifier.GetMeasure()
	sn := attestationVerifier.GetSerialNumber()

	c.JSON(http.StatusOK, gin.H{
		"measure": measure,
		"sn":      sn,
		"scheme":  string(scheme),
	})
}

func (a *UdpAgent) rotateAgentKey(c *gin.Context) {
	if !a.beginLifecycleOp(c) {
		return
	}
	defer a.endLifecycleOp()

	if err := a.RotateAgentKey(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if err := a.RestartAgent(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"msg": "successfully rotate agent key"})
}

func (a *UdpAgent) rotateTeeKey(c *gin.Context) {
	if !a.beginLifecycleOp(c) {
		return
	}
	defer a.endLifecycleOp()

	if err := a.RotateTeeKey(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if err := a.RestartAgent(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"msg": "successfully rotate TEE key"})
}
