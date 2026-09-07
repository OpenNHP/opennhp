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
	// DefaultPort is used when Config.ListenPort is 0. Each daemon passes its
	// own so a host running several daemons does not collide on one port.
	DefaultPort int
	// OnListening / OnServeError / OnRenderError are optional log hooks so the
	// metrics package stays free of a logging dependency.
	OnListening   func(addr string)
	OnServeError  func(err error)
	OnRenderError func(err error)
}

// Endpoint is a running metrics HTTP listener serving /metrics and a minimal
// /healthz on its own socket.
type Endpoint struct {
	http *http.Server
}

// StartEndpoint binds and serves the metrics endpoint. The bind happens
// before returning, so a bad address or a port already in use is reported to
// the caller rather than failing silently in the serve goroutine.
func StartEndpoint(cfg Config, opts EndpointOptions) (*Endpoint, error) {
	ip := cfg.ListenIp
	if ip == "" {
		ip = DefaultListenIP
	}
	port := cfg.ListenPort
	if port == 0 {
		port = opts.DefaultPort
	}
	addr := net.JoinHostPort(ip, strconv.Itoa(port))

	ln, err := net.Listen("tcp", addr)
	if err != nil {
		return nil, fmt.Errorf("listen %s: %w", addr, err)
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/metrics", func(w http.ResponseWriter, r *http.Request) {
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
		// Deliberately minimal: status + uptime only. No version/commit, so
		// an unauthenticated probe cannot fingerprint the exact build if the
		// endpoint is ever bound off-loopback — this is a network-hiding
		// product and a liveness check does not need to leak that.
		var up int64
		if opts.Uptime != nil {
			up = int64(opts.Uptime().Seconds())
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(map[string]any{"status": "ok", "uptime_s": up})
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
	return &Endpoint{http: srv}, nil
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
