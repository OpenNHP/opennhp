package metrics

import (
	"context"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"strconv"
	"time"
)

// DefaultListenIP is the bind address used when Config.ListenIp is empty:
// loopback, so the observability surface is never reachable off-host unless
// the operator explicitly opts in.
const DefaultListenIP = "127.0.0.1"

// Config is the shared [Metrics] section shape for every NHP daemon. It is
// opt-in (Enabled defaults false) and local-by-default.
type Config struct {
	Enabled    bool   `json:"enabled" toml:"Enabled"`
	ListenIp   string `json:"listenIp" toml:"ListenIp"`
	ListenPort int    `json:"listenPort" toml:"ListenPort"`
}

// EndpointOptions configures a metrics HTTP endpoint.
type EndpointOptions struct {
	// Registry is rendered on GET /metrics in the Prometheus text format.
	Registry *Registry
	// Uptime is reported by GET /healthz. Optional; a nil func reports 0.
	Uptime func() time.Duration
	// IsRunning, when set, gates GET /healthz: it returns 503 (not 200) once
	// it reports false, so the endpoint is usable as a real container
	// liveness/readiness probe rather than staying green through shutdown.
	IsRunning func() bool
	// DefaultPort is used when Config.ListenPort is 0. Each daemon passes its
	// own so a host running several daemons does not collide on one port.
	DefaultPort int
	// OnListening / OnServeError / OnRenderError are optional log hooks so the
	// metrics package stays free of a logging dependency.
	OnListening   func(addr string)
	OnServeError  func(err error)
	OnRenderError func(err error)
	// OnInsecureBind, when set, is called with the resolved bind IP before
	// listening if that IP is not loopback. Nothing here stops an operator
	// from setting ListenIp to a routable address; for a port-hiding product
	// that turns /metrics and /healthz into an unauthenticated, self-
	// identifying TCP port, so the caller gets a chance to log it loudly.
	OnInsecureBind func(ip string)
}

// Endpoint is a running metrics HTTP listener serving /metrics and a minimal
// /healthz on its own socket.
type Endpoint struct {
	http *http.Server
	addr string
}

// Addr is the "host:port" the endpoint bound. Useful when Config.ListenPort
// was 0 (ephemeral). Empty on a nil Endpoint.
func (e *Endpoint) Addr() string {
	if e == nil {
		return ""
	}
	return e.addr
}

// StartEndpoint binds and serves the metrics endpoint. It returns (nil, nil)
// when cfg.Enabled is false, so callers do not each repeat that check. The
// bind happens before returning, so a bad address or a port already in use is
// reported to the caller rather than failing silently in the serve goroutine.
func StartEndpoint(cfg Config, opts EndpointOptions) (*Endpoint, error) {
	if !cfg.Enabled {
		return nil, nil
	}
	ip := cfg.ListenIp
	if ip == "" {
		ip = DefaultListenIP
	}
	port := cfg.ListenPort
	if port == 0 {
		port = opts.DefaultPort
	}
	if opts.OnInsecureBind != nil {
		if parsed := net.ParseIP(ip); parsed == nil || !parsed.IsLoopback() {
			opts.OnInsecureBind(ip)
		}
	}
	addr := net.JoinHostPort(ip, strconv.Itoa(port))

	ln, err := net.Listen("tcp", addr)
	if err != nil {
		return nil, fmt.Errorf("listen %s: %w", addr, err)
	}
	addr = ln.Addr().String() // resolves an ephemeral :0 to the real port

	mux := http.NewServeMux()
	mux.HandleFunc("/metrics", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet && r.Method != http.MethodHead {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		if opts.Registry == nil {
			http.Error(w, "metrics not initialized", http.StatusServiceUnavailable)
			return
		}
		w.Header().Set("Content-Type", "text/plain; version=0.0.4; charset=utf-8")
		if werr := opts.Registry.WriteText(w); werr != nil && opts.OnRenderError != nil {
			opts.OnRenderError(werr)
		}
	})
	mux.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet && r.Method != http.MethodHead {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		// Deliberately minimal: status + uptime only. No version/commit, so
		// an unauthenticated probe cannot fingerprint the exact build if the
		// endpoint is ever bound off-loopback — this is a network-hiding
		// product and a liveness check does not need to leak that.
		var up int64
		if opts.Uptime != nil {
			up = int64(opts.Uptime().Seconds())
		}
		status, code := "ok", http.StatusOK
		if opts.IsRunning != nil && !opts.IsRunning() {
			status, code = "stopping", http.StatusServiceUnavailable
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(code)
		_ = json.NewEncoder(w).Encode(map[string]any{"status": status, "uptime_s": up})
	})

	srv := &http.Server{
		Addr:              addr,
		Handler:           mux,
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       5 * time.Second,
		WriteTimeout:      10 * time.Second,
	}
	if opts.OnListening != nil {
		opts.OnListening(addr)
	}
	go func() {
		if serveErr := srv.Serve(ln); serveErr != nil && serveErr != http.ErrServerClosed && opts.OnServeError != nil {
			opts.OnServeError(serveErr)
		}
	}()
	return &Endpoint{http: srv, addr: addr}, nil
}

// Stop gracefully shuts the endpoint down. Safe on a nil Endpoint.
func (e *Endpoint) Stop() {
	if e == nil || e.http == nil {
		return
	}
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	_ = e.http.Shutdown(ctx)
}
