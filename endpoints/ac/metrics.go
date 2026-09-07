package ac

import (
	"time"

	"github.com/OpenNHP/opennhp/nhp/metrics"
)

// acMetrics holds the collectors nhp-ac updates. Collection is always on
// (cheap atomics); only the HTTP endpoint that exposes them is opt-in.
// Every method is safe on a nil receiver so partially-built ACs (tests,
// early startup) need no per-call guard.
type acMetrics struct {
	registry *metrics.Registry

	messagesReceived *metrics.CounterVec // type=NHP-AOP|...
	acOperations     *metrics.CounterVec // result=ok|error
	acOpDuration     *metrics.Histogram  // server->AC operation handling, seconds
	packetsDropped   *metrics.CounterVec // stage=parse|validate|decrypt|queue_full
}

func newACMetrics(a *UdpAC, startTime time.Time) *acMetrics {
	reg := metrics.NewRegistry()

	reg.NewGauge("nhp_ac_start_time_seconds",
		"Unix time the nhp-ac process started; subtract from time() for uptime.").
		Set(startTime.Unix())

	reg.NewGaugeFunc("nhp_ac_active_connections",
		"Tracked server UDP connections right now.",
		func() float64 {
			a.remoteConnectionMutex.Lock()
			n := len(a.remoteConnectionMap)
			a.remoteConnectionMutex.Unlock()
			return float64(n)
		})

	m := &acMetrics{
		registry: reg,
		messagesReceived: reg.NewCounter("nhp_ac_messages_received_total",
			"Decrypted protocol messages received, by message type.", "type"),
		acOperations: reg.NewCounter("nhp_ac_operations_total",
			"Access-control operations handled for the server, by result.", "result"),
		acOpDuration: reg.NewHistogram("nhp_ac_operation_duration_seconds",
			"Time to apply one server access-control operation, in seconds.", nil).With(),
		packetsDropped: reg.NewCounter("nhp_ac_packets_dropped_total",
			"Inbound packets discarded before becoming a decrypted message, by stage.", "stage"),
	}

	// Pre-create the closed-set label series so a fresh scrape shows an
	// explicit 0 rather than a missing series.
	m.acOperations.With("ok")
	m.acOperations.With("error")
	for _, s := range []string{"parse", "validate", "decrypt", "queue_full"} {
		m.packetsDropped.With(s)
	}

	return m
}

func (m *acMetrics) recordMessageReceived(msgType string) {
	if m == nil {
		return
	}
	m.messagesReceived.With(msgType).Inc()
}

func (m *acMetrics) recordACOperation(ok bool, seconds float64) {
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

func (m *acMetrics) recordDroppedPacket(stage string) {
	if m == nil {
		return
	}
	m.packetsDropped.With(stage).Inc()
}
