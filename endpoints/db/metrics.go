package db

import (
	"sync/atomic"
	"time"

	"github.com/OpenNHP/opennhp/nhp/metrics"
)

// defaultDBMetricsPort is the metrics endpoint port when [Metrics] ListenPort
// is 0. Distinct from server (9100), ac (9101) and relay (9102).
const defaultDBMetricsPort = 9103

// dbMetrics holds the collectors nhp-db updates. Nil-receiver safe.
type dbMetrics struct {
	registry *metrics.Registry

	messagesReceived *metrics.CounterVec // type=DHP-DAK|...
	packetsDropped   *metrics.CounterVec // stage=parse|validate|decrypt|queue_full
}

func newDBMetrics(a *UdpDevice, startTime time.Time) *dbMetrics {
	reg := metrics.NewRegistry()

	reg.NewGauge("nhp_db_start_time_seconds",
		"Unix time the nhp-db process started; subtract from time() for uptime.").
		Set(startTime.Unix())

	reg.NewGaugeFunc("nhp_db_active_connections",
		"Tracked server UDP connections right now.",
		func() float64 {
			a.remoteConnectionMutex.Lock()
			n := len(a.remoteConnectionMap)
			a.remoteConnectionMutex.Unlock()
			return float64(n)
		})

	reg.NewGaugeFunc("nhp_db_received_bytes_total",
		"Total UDP payload bytes received.",
		func() float64 { return float64(atomic.LoadUint64(&a.stats.totalRecvBytes)) })
	reg.NewGaugeFunc("nhp_db_sent_bytes_total",
		"Total UDP payload bytes sent.",
		func() float64 { return float64(atomic.LoadUint64(&a.stats.totalSendBytes)) })

	m := &dbMetrics{
		registry: reg,
		messagesReceived: reg.NewCounter("nhp_db_messages_received_total",
			"Decrypted protocol messages received, by message type.", "type"),
		packetsDropped: reg.NewCounter("nhp_db_packets_dropped_total",
			"Inbound packets discarded before becoming a decrypted message, by stage.", "stage"),
	}
	for _, s := range []string{"parse", "validate", "decrypt", "queue_full"} {
		m.packetsDropped.With(s)
	}
	return m
}

func (m *dbMetrics) recordMessageReceived(msgType string) {
	if m == nil {
		return
	}
	m.messagesReceived.With(msgType).Inc()
}

func (m *dbMetrics) recordDroppedPacket(stage string) {
	if m == nil {
		return
	}
	m.packetsDropped.With(stage).Inc()
}
