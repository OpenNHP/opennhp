package core

import (
	"testing"
	"testing/synctest"

	"github.com/OpenNHP/opennhp/nhp/log"
)

func TestDeviceStopWaitsForFinalTransactionLog(t *testing.T) {
	logger := log.NewLogger("", 0, "", "")
	defer log.SetGlobalLogger(log.NewLogger("", 3, "", ""))
	synctest.Test(t, func(t *testing.T) {
		entered, release := make(chan struct{}), make(chan struct{})
		logger.Debug = func(format string, args ...any) {
			if format == "Local transaction %d quit" {
				close(entered)
				<-release
			}
		}
		log.SetGlobalLogger(logger)
		d := NewDevice(NHP_AC, NewECDH(ECC_CURVE25519).PrivateKey(), nil)
		tx := newLocalTransaction(1, &ConnectionData{}, &MsgAssemblerData{device: d}, 10000)
		d.AddLocalTransaction(tx)
		stopped := make(chan struct{})
		go func() { d.Stop(); close(stopped) }()
		<-entered
		synctest.Wait()
		select {
		case <-stopped:
			t.Error("Device.Stop returned while a transaction still used the logger")
		default:
		}
		close(release)
		<-stopped
		synctest.Wait()
	})
}

func TestConnectionWaitsForFinalTransactionLog(t *testing.T) {
	logger := log.NewLogger("", 0, "", "")
	defer log.SetGlobalLogger(log.NewLogger("", 3, "", ""))
	synctest.Test(t, func(t *testing.T) {
		entered, release := make(chan struct{}), make(chan struct{})
		logger.Debug = func(format string, args ...any) {
			if format == "Remote transaction %d quit" {
				close(entered)
				<-release
			}
		}
		log.SetGlobalLogger(logger)
		d := NewDevice(NHP_SERVER, NewECDH(ECC_CURVE25519).PrivateKey(), nil)
		defer d.Stop()
		conn := &ConnectionData{StopSignal: make(chan struct{}), RemoteTransactionMap: make(map[uint64]*RemoteTransaction)}
		tx := newRemoteTransaction(1, conn, &PacketParserData{device: d}, 10000)
		conn.AddRemoteTransaction(tx)
		stopped := make(chan struct{})
		go func() { close(conn.StopSignal); conn.Wait(); close(stopped) }()
		<-entered
		synctest.Wait()
		select {
		case <-stopped:
			t.Error("connection wait returned while a transaction still used the logger")
		default:
		}
		close(release)
		<-stopped
		synctest.Wait()
	})
}
