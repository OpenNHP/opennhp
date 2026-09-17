package core

import (
	"errors"
	"runtime"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	common "github.com/OpenNHP/opennhp/nhp/common"
)

func TestRemoteTransactionSendMessageAfterDone(t *testing.T) {
	tx := &RemoteTransaction{
		NextMsgCh: make(chan *MsgData),
		done:      make(chan struct{}),
	}
	close(tx.done)

	if err := tx.SendMessage(&MsgData{}); !errors.Is(err, common.ErrTransactionClosed) {
		t.Fatalf("SendMessage after done = %v, want ErrTransactionClosed", err)
	}
}

func TestRemoteTransactionSendRacesExit(t *testing.T) {
	const (
		iterations = 50
		senders    = 64
	)

	for iteration := 0; iteration < iterations; iteration++ {
		tx := &RemoteTransaction{
			NextMsgCh: make(chan *MsgData),
			done:      make(chan struct{}),
		}

		var delivered atomic.Int64
		var closed atomic.Int64
		var drainer sync.WaitGroup
		drainer.Add(1)
		go func() {
			defer drainer.Done()
			for {
				select {
				case <-tx.NextMsgCh:
					delivered.Add(1)
				case <-tx.done:
					return
				}
			}
		}()

		panicCh := make(chan any, senders)
		var entered atomic.Int64
		var sendersWG sync.WaitGroup
		for i := 0; i < senders; i++ {
			sendersWG.Add(1)
			go func() {
				defer sendersWG.Done()
				defer captureTransactionPanic(panicCh)
				entered.Add(1)
				if err := tx.SendMessage(&MsgData{}); errors.Is(err, common.ErrTransactionClosed) {
					closed.Add(1)
				} else if err != nil {
					panicCh <- err
				}
			}()
		}

		for entered.Load() != senders {
			runtime.Gosched()
		}
		close(tx.done)
		sendersWG.Wait()
		drainer.Wait()

		select {
		case recovered := <-panicCh:
			t.Fatalf("iteration %d: send raced exit and panicked: %v", iteration, recovered)
		default:
		}
		if got := delivered.Load() + closed.Load(); got != senders {
			t.Fatalf("iteration %d: delivered + closed = %d, want %d", iteration, got, senders)
		}
	}
}

func TestRemoteTransactionSendRacesExitWithoutReceiver(t *testing.T) {
	const senders = 64

	tx := &RemoteTransaction{
		NextMsgCh: make(chan *MsgData),
		done:      make(chan struct{}),
	}
	var entered atomic.Int64
	var sendersWG sync.WaitGroup
	for i := 0; i < senders; i++ {
		sendersWG.Add(1)
		go func() {
			defer sendersWG.Done()
			entered.Add(1)
			if err := tx.SendMessage(&MsgData{}); !errors.Is(err, common.ErrTransactionClosed) {
				t.Errorf("SendMessage = %v, want ErrTransactionClosed", err)
			}
		}()
	}

	for entered.Load() != senders {
		runtime.Gosched()
	}
	close(tx.done)

	done := make(chan struct{})
	go func() {
		sendersWG.Wait()
		close(done)
	}()
	select {
	case <-done:
	case <-time.After(5 * time.Second):
		t.Fatal("senders remained blocked after transaction exit")
	}
}

func TestLocalTransactionSendsRaceExit(t *testing.T) {
	const (
		iterations = 30
		senders    = 32
	)

	for iteration := 0; iteration < iterations; iteration++ {
		tx := &LocalTransaction{
			NextPacketCh:  make(chan *Packet),
			ExternalMsgCh: make(chan *PacketParserData),
			done:          make(chan struct{}),
		}

		var drainers sync.WaitGroup
		drainers.Add(2)
		go drainUntilTransactionDone(tx.NextPacketCh, tx.done, &drainers)
		go drainUntilTransactionDone(tx.ExternalMsgCh, tx.done, &drainers)

		panicCh := make(chan any, senders*2)
		var sendersWG sync.WaitGroup
		for i := 0; i < senders; i++ {
			sendersWG.Add(2)
			go func() {
				defer sendersWG.Done()
				defer captureTransactionPanic(panicCh)
				_ = tx.SendPacket(&Packet{})
			}()
			go func() {
				defer sendersWG.Done()
				defer captureTransactionPanic(panicCh)
				_ = tx.SendExternalMsg(&PacketParserData{})
			}()
		}

		runtime.Gosched()
		close(tx.done)
		sendersWG.Wait()
		drainers.Wait()

		select {
		case recovered := <-panicCh:
			t.Fatalf("iteration %d: local send raced exit and panicked: %v", iteration, recovered)
		default:
		}
	}
}

func TestRemoteTransactionCleanupOrdering(t *testing.T) {
	conn := &ConnectionData{RemoteTransactionMap: make(map[uint64]*RemoteTransaction)}
	tx := &RemoteTransaction{
		transactionId: 42,
		NextMsgCh:     make(chan *MsgData),
		done:          make(chan struct{}),
	}
	conn.RemoteTransactionMap[tx.transactionId] = tx

	captured := conn.FindRemoteTransaction(tx.transactionId)
	conn.RemoteTransactionMutex.Lock()
	delete(conn.RemoteTransactionMap, tx.transactionId)
	close(tx.done)
	conn.RemoteTransactionMutex.Unlock()

	if got := conn.FindRemoteTransaction(tx.transactionId); got != nil {
		t.Fatalf("FindRemoteTransaction after cleanup = %p, want nil", got)
	}
	if err := captured.SendMessage(&MsgData{}); !errors.Is(err, common.ErrTransactionClosed) {
		t.Fatalf("stale transaction SendMessage = %v, want ErrTransactionClosed", err)
	}
}

func captureTransactionPanic(ch chan<- any) {
	if recovered := recover(); recovered != nil {
		ch <- recovered
	}
}

func drainUntilTransactionDone[T any](ch <-chan T, done <-chan struct{}, wg *sync.WaitGroup) {
	defer wg.Done()
	for {
		select {
		case <-ch:
		case <-done:
			return
		}
	}
}
