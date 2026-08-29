package server

import (
	"context"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"strconv"
	"time"

	"github.com/OpenNHP/opennhp/nhp/log"
)

const (
	defaultMetricsListenIp   = "127.0.0.1"
	defaultMetricsListenPort = 9100
)

// metricsServer is the opt-in observability listener. It serves the Prometheus
// exposition on /metrics and a small JSON liveness probe on /healthz, on a
// separate socket from the public knock/HTTP surface so telemetry is never
// reachable off-host by default.
type metricsServer struct {
	us   *UdpServer
	http *http.Server
}

func newMetricsServer(us *UdpServer) *metricsServer {
	return &metricsServer{us: us}
}

func (ms *metricsServer) start() error {
	cfg := ms.us.config.Metrics

	ip := cfg.ListenIp
	if ip == "" {
		ip = defaultMetricsListenIp
	}
	port := cfg.ListenPort
	if port == 0 {
		port = defaultMetricsListenPort
	}
	addr := net.JoinHostPort(ip, strconv.Itoa(port))

	// Bind before returning so a bad address / port-in-use surfaces to the
	// caller instead of failing silently inside the serve goroutine.
	ln, err := net.Listen("tcp", addr)
	if err != nil {
		return fmt.Errorf("listen %s: %w", addr, err)
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/metrics", ms.handleMetrics)
	mux.HandleFunc("/healthz", ms.handleHealthz)

	ms.http = &http.Server{
		Addr:              addr,
		Handler:           mux,
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       5 * time.Second,
		WriteTimeout:      10 * time.Second,
	}

	log.Info("[Metrics] endpoint listening on http://%s (/metrics, /healthz)", addr)
	go func() {
		if serveErr := ms.http.Serve(ln); serveErr != nil && serveErr != http.ErrServerClosed {
			log.Error("[Metrics] endpoint stopped unexpectedly: %v", serveErr)
		}
	}()
	return nil
}

func (ms *metricsServer) stop() {
	if ms == nil || ms.http == nil {
		return
	}
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	_ = ms.http.Shutdown(ctx)
}

func (ms *metricsServer) handleMetrics(w http.ResponseWriter, r *http.Request) {
	reg := ms.us.metrics
	if reg == nil || reg.registry == nil {
		http.Error(w, "metrics not initialized", http.StatusServiceUnavailable)
		return
	}
	w.Header().Set("Content-Type", "text/plain; version=0.0.4; charset=utf-8")
	if err := reg.registry.WriteText(w); err != nil {
		log.Error("[Metrics] failed to render exposition: %v", err)
	}
}

func (ms *metricsServer) handleHealthz(w http.ResponseWriter, r *http.Request) {
	// Deliberately minimal: status + uptime only. Version/commit are
	// omitted so that if an operator ever binds this endpoint to a routable
	// address, an unauthenticated probe can't fingerprint the exact build —
	// this is a network-hiding product, and a liveness check doesn't need
	// to leak that.
	body := map[string]any{
		"status":   "ok",
		"uptime_s": int64(time.Since(ms.us.startTime).Seconds()),
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(body)
}
