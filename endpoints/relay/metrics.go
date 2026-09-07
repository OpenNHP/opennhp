package relay

import (
	"sync/atomic"
	"time"

	"github.com/OpenNHP/opennhp/nhp/metrics"
)

// defaultRelayMetricsPort is the metrics endpoint port when [Metrics]
// ListenPort is 0. Distinct from server (9100) and ac (9101).
const defaultRelayMetricsPort = 9102

// relayMetrics holds the collectors nhp-relay updates. Nil-receiver safe.
type relayMetrics struct {
	registry *metrics.Registry

	messagesReceived *metrics.CounterVec // type=NHP-RKN|...
	packetsDropped   *metrics.CounterVec // stage=parse|validate|decrypt|queue_full
}

func newRelayMetrics(rs *RelayServer, startTime time.Time) *relayMetrics {
	reg := metrics.NewRegistry()

	reg.NewGauge("nhp_relay_start_time_seconds",
		"Unix time the nhp-relay process started; subtract from time() for uptime.").
		Set(startTime.Unix())

	reg.NewGaugeFunc("nhp_relay_upstream_servers",
		"Configured upstream nhp-server identities.",
		func() float64 { return float64(len(rs.servers)) })

	reg.NewCounterFunc("nhp_relay_received_bytes_total",
		"Total UDP payload bytes received on the relay data path.",
		func() float64 { return float64(atomic.LoadUint64(&rs.stats.totalRecvBytes)) })
	reg.NewCounterFunc("nhp_relay_sent_bytes_total",
		"Total UDP payload bytes sent on the relay data path.",
		func() float64 { return float64(atomic.LoadUint64(&rs.stats.totalSendBytes)) })

	m := &relayMetrics{
		registry: reg,
		messagesReceived: reg.NewCounter("nhp_relay_messages_received_total",
			"Decrypted protocol messages received, by message type.", "type"),
		packetsDropped: reg.NewCounter("nhp_relay_packets_dropped_total",
			"Inbound packets discarded before becoming a decrypted message, by stage (precheck, parse, validate, decrypt, queue_full).", "stage"),
	}
	for _, s := range []string{"precheck", "parse", "validate", "decrypt", "queue_full"} {
		m.packetsDropped.With(s)
	}
	return m
}

func (m *relayMetrics) recordMessageReceived(msgType string) {
	if m == nil {
		return
	}
	m.messagesReceived.With(msgType).Inc()
}

func (m *relayMetrics) recordDroppedPacket(stage string) {
	if m == nil {
		return
	}
	m.packetsDropped.With(stage).Inc()
}
