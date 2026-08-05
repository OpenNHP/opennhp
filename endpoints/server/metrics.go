package server

import (
	"time"

	"github.com/OpenNHP/opennhp/nhp/metrics"
)

// serverMetrics holds the collectors the nhp-server updates. Collection is
// always on (the increments are cheap atomics); only the HTTP endpoint that
// exposes them is opt-in. Every method is safe to call on a nil receiver so
// code paths that build a partial UdpServer (tests, early startup) don't have
// to guard each call site.
type serverMetrics struct {
	registry *metrics.Registry

	messagesReceived *metrics.CounterVec // by protocol message type
	knockAuth        *metrics.CounterVec // result=ok|denied
	acOperations     *metrics.CounterVec // result=ok|error
	acOpDuration     *metrics.Histogram  // server->AC round trip, seconds
	blockedAddrs     *metrics.Counter    // sources blocked past the threat threshold
}

func newServerMetrics(s *UdpServer, startTime time.Time) *serverMetrics {
	reg := metrics.NewRegistry()

	reg.NewGauge("nhp_server_start_time_seconds",
		"Unix time the nhp-server process started; subtract from time() for uptime.").
		Set(startTime.Unix())

	reg.NewGaugeFunc("nhp_server_active_connections",
		"Tracked remote UDP connections right now.",
		func() float64 {
			s.remoteConnectionMapMutex.Lock()
			n := len(s.remoteConnectionMap)
			s.remoteConnectionMapMutex.Unlock()
			return float64(n)
		})

	reg.NewGaugeFunc("nhp_server_overloaded",
		"1 when the server is in cookie-gated overload mode, else 0.",
		func() float64 {
			if s.device != nil && s.device.IsOverload() {
				return 1
			}
			return 0
		})

	sm := &serverMetrics{
		registry: reg,
		messagesReceived: reg.NewCounter("nhp_server_messages_received_total",
			"Decrypted protocol messages received, by message type.", "type"),
		knockAuth: reg.NewCounter("nhp_server_knock_auth_total",
			"Knock authentication outcomes.", "result"),
		acOperations: reg.NewCounter("nhp_server_ac_operations_total",
			"Server-to-AC operations, by result.", "result"),
		acOpDuration: reg.NewHistogram("nhp_server_ac_operation_duration_seconds",
			"Latency of server-to-AC operations, in seconds.", nil).With(),
		blockedAddrs: reg.NewCounter("nhp_server_blocked_source_addresses_total",
			"Source addresses blocked after exceeding the threat threshold.").With(),
	}

	// Pre-create the closed-set label series so they export an explicit 0
	// from the first scrape, rather than appearing only after the first
	// event — this keeps dashboards and rate() alerts from seeing a missing
	// series (which reads as a gap, not a zero). Only the fully-enumerable
	// label sets are seeded; message "type" is open-ended and left lazy.
	sm.knockAuth.With("ok")
	sm.knockAuth.With("denied")
	sm.acOperations.With("ok")
	sm.acOperations.With("error")

	return sm
}

func (m *serverMetrics) recordMessageReceived(msgType string) {
	if m == nil {
		return
	}
	m.messagesReceived.With(msgType).Inc()
}

func (m *serverMetrics) recordKnockAuth(ok bool) {
	if m == nil {
		return
	}
	result := "denied"
	if ok {
		result = "ok"
	}
	m.knockAuth.With(result).Inc()
}

func (m *serverMetrics) recordACOperation(ok bool, seconds float64) {
	if m == nil {
		return
	}
	result := "error"
	if ok {
		result = "ok"
	}
	m.acOperations.With(result).Inc()
	m.acOpDuration.Observe(seconds)
}

func (m *serverMetrics) recordBlockedAddr() {
	if m == nil {
		return
	}
	m.blockedAddrs.Inc()
}
