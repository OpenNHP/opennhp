package core

import (
	"net"
	"runtime"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

func TestConnectionDataCloseConcurrentIsIdempotent(t *testing.T) {
	conn := newConnectionDataCloseRaceHarness()

	const closers = 32
	var wg sync.WaitGroup
	for i := 0; i < closers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			conn.Close()
		}()
	}
	wg.Wait()

	if !conn.IsClosed() {
		t.Fatal("Close did not mark the connection closed")
	}
	assertConnectionChannelClosed(t, "StopSignal", conn.StopSignal)
	assertConnectionChannelClosed(t, "SendQueue", conn.SendQueue)
	assertConnectionChannelClosed(t, "RecvQueue", conn.RecvQueue)
	assertConnectionChannelClosed(t, "BlockSignal", conn.BlockSignal)
	assertConnectionChannelClosed(t, "SetTimeoutSignal", conn.SetTimeoutSignal)
}

func TestConnectionDataPacketSendsRaceClose(t *testing.T) {
	tests := []struct {
		name string
		send func(*ConnectionData)
	}{
		{
			name: "outbound",
			send: func(conn *ConnectionData) {
				conn.ForwardOutboundPacket(&Packet{Content: []byte{1}})
			},
		},
		{
			name: "inbound",
			send: func(conn *ConnectionData) {
				conn.ForwardInboundPacket(&Packet{Content: []byte{1}})
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			const (
				iterations = 50
				senders    = 64
			)
			for iteration := 0; iteration < iterations; iteration++ {
				conn := newConnectionDataCloseRaceHarness()
				var entered atomic.Int64
				var sendersWG sync.WaitGroup
				panicCh := make(chan any, senders)
				for i := 0; i < senders; i++ {
					sendersWG.Add(1)
					go func() {
						defer sendersWG.Done()
						defer captureConnectionPanic(panicCh)
						entered.Add(1)
						test.send(conn)
					}()
				}

				for entered.Load() != senders {
					runtime.Gosched()
				}
				conn.Close()
				sendersWG.Wait()

				select {
				case recovered := <-panicCh:
					t.Fatalf("iteration %d: channel send raced Close and panicked: %v", iteration, recovered)
				default:
				}
			}
		})
	}
}

func TestConnectionDataSignalsRaceClose(t *testing.T) {
	const iterations = 100

	for iteration := 0; iteration < iterations; iteration++ {
		conn := newConnectionDataCloseRaceHarness()
		panicCh := make(chan any, 2)
		var wg sync.WaitGroup
		wg.Add(2)
		go func() {
			defer wg.Done()
			defer captureConnectionPanic(panicCh)
			conn.SetTimeout(iteration + 1)
		}()
		go func() {
			defer wg.Done()
			defer captureConnectionPanic(panicCh)
			for i := 0; i < 64; i++ {
				conn.SendBlockSignal()
				runtime.Gosched()
			}
		}()

		runtime.Gosched()
		conn.Close()
		wg.Wait()

		select {
		case recovered := <-panicCh:
			t.Fatalf("iteration %d: signal send raced Close and panicked: %v", iteration, recovered)
		default:
		}
	}
}

func TestConnectionDataCloseUnblocksPendingSends(t *testing.T) {
	conn := newConnectionDataCloseRaceHarness()

	done := make(chan struct{})
	go func() {
		conn.ForwardOutboundPacket(&Packet{})
		conn.ForwardInboundPacket(&Packet{})
		conn.SetTimeout(250)
		close(done)
	}()

	runtime.Gosched()
	conn.Close()
	select {
	case <-done:
	case <-time.After(5 * time.Second):
		t.Fatal("pending channel send remained blocked after Close")
	}
}

func newConnectionDataCloseRaceHarness() *ConnectionData {
	return &ConnectionData{
		Device:               &Device{},
		LocalAddr:            &net.UDPAddr{IP: net.ParseIP("127.0.0.1"), Port: 2000},
		RemoteAddr:           &net.UDPAddr{IP: net.ParseIP("127.0.0.1"), Port: 2001},
		CookieStore:          &CookieStore{},
		SendQueue:            make(chan *Packet),
		RecvQueue:            make(chan *Packet),
		BlockSignal:          make(chan struct{}),
		SetTimeoutSignal:     make(chan struct{}),
		StopSignal:           make(chan struct{}),
		RemoteTransactionMap: make(map[uint64]*RemoteTransaction),
	}
}

func assertConnectionChannelClosed[T any](t *testing.T, name string, ch <-chan T) {
	t.Helper()
	if ch == nil {
		t.Fatalf("%s is nil after Close", name)
	}
	select {
	case _, ok := <-ch:
		if ok {
			t.Fatalf("%s remained open after Close", name)
		}
	default:
		t.Fatalf("%s remained open after Close", name)
	}
}

func captureConnectionPanic(ch chan<- any) {
	if recovered := recover(); recovered != nil {
		ch <- recovered
	}
}
