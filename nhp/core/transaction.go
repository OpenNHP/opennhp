package core

import (
	"time"

	common "github.com/OpenNHP/opennhp/nhp/common"
	log "github.com/OpenNHP/opennhp/nhp/log"
)

type LocalTransaction struct {
	transactionId uint64
	connData      *ConnectionData
	mad           *MsgAssemblerData
	NextPacketCh  chan *Packet           // higher level entities should redirect packet to this channel
	ExternalMsgCh chan *PacketParserData // a channel to receive an external msg to complete the transaction
	timeout       int
}

type RemoteTransaction struct {
	transactionId uint64
	connData      *ConnectionData
	parserData    *PacketParserData
	NextMsgCh     chan *MsgData // higher level entities should redirect message to this channel
	timeout       int
}

func (d *Device) IsTransactionRequest(t int) bool {

	// NHP_KPL is handled separately
	log.Info("IsTransactionRequest: deviceType:%d", d.deviceType)
	switch d.deviceType {
	case NHP_AGENT:
		switch t {
		case NHP_REG, NHP_LST, NHP_KNK, DHP_KNK, NHP_RKN, NHP_EXT, NHP_DAR, NHP_DAV:
			return true
		}
	case NHP_SERVER:
		switch t {
		case NHP_REG, NHP_LST, NHP_KNK, DHP_KNK, NHP_RKN, NHP_EXT, NHP_AOL, NHP_AOP, NHP_DAK, NHP_DAG, NHP_DSA, NHP_DAR, NHP_DAV, NHP_DRG, NHP_DOL, NHP_DWR:
			return true
		}
	case NHP_AC:
		switch t {
		case NHP_AOL, NHP_AOP:
			return true
		}
	case NHP_DB:
		switch t {
		case NHP_DRG, NHP_DOL, NHP_DWR:
			return true
		}
	case NHP_RELAY:

		// no transaction request for relay
	}

	return false
}

func (d *Device) LocalTransactionTimeout() int {
	// NHP_KPL is handled separately
	switch d.deviceType {
	case NHP_AGENT:
		return AgentLocalTransactionResponseTimeoutMs
	case NHP_SERVER:
		return ServerLocalTransactionResponseTimeoutMs
	case NHP_AC:
		return ACLocalTransactionResponseTimeoutMs
	case NHP_DB:
		return DELocalTransactionResponseTimeoutMs
	case NHP_RELAY:
		// no transaction request for relay
	}

	return 0
}

func (d *Device) RemoteTransactionTimeout() int {
	return RemoteTransactionProcessTimeoutMs
}

func (d *Device) IsTransactionResponse(t int) bool {
	// NHP_KPL is handled separately
	switch d.deviceType {
	case NHP_AGENT:
		switch t {
		case NHP_RAK, NHP_LRT, NHP_ACK, NHP_DAG, NHP_DSA:
			// note NHP_COK is not handled as transaction for agent
			return true
		}
	case NHP_SERVER:
		switch t {
		case NHP_RAK, NHP_LRT, NHP_ACK, NHP_AAK, NHP_ART, NHP_DAK, NHP_DWA:
			// note NHP_COK is not handled as transaction for server
			return true
		}
	case NHP_AC:
		switch t {
		case NHP_AAK, NHP_ART:
			return true
		}
	case NHP_DB:
		switch t {
		case NHP_DAK, NHP_DBA:
			return true
		}
	case NHP_RELAY:
		// no transaction response for relay
	}

	return false
}

// LocalTransaction
func (d *Device) AddLocalTransaction(t *LocalTransaction) {
	d.localTransactionMutex.Lock()
	defer d.localTransactionMutex.Unlock()

	d.localTransactionMap[t.transactionId] = t

	d.wg.Add(1)
	go t.Run()
}

func (d *Device) FindLocalTransaction(id uint64) *LocalTransaction {
	d.localTransactionMutex.Lock()
	defer d.localTransactionMutex.Unlock()

	t, found := d.localTransactionMap[id]
	if found {
		return t
	}

	return nil
}

func (t *LocalTransaction) Run() {
	log.Debug("Local transaction %d start", t.transactionId)
	defer log.Debug("Local transaction %d quit", t.transactionId)

	device := t.mad.device
	var err error

	t.ExternalMsgCh = make(chan *PacketParserData)

	// clear up
	defer func() {
		t.mad.Destroy()
		close(t.NextPacketCh)
		close(t.ExternalMsgCh)

		device.localTransactionMutex.Lock()
		delete(device.localTransactionMap, t.transactionId)
		device.localTransactionMutex.Unlock()

		// if local transaction is expecting a response, return an error
		if err != nil && t.mad.ResponseMsgCh != nil {
			t.mad.ResponseMsgCh <- &PacketParserData{Error: err}
		}

		device.wg.Done()
	}()

	select {
	case pkt := <-t.NextPacketCh:
		pd := &PacketData{
			BasePacket:        pkt,
			PrevAssemblerData: t.mad,
			InitTime:          time.Now().UnixNano(),
		}

		device.RecvPacketToMsg(pd)
		return

	case ppd := <-t.ExternalMsgCh:
		// redirect it to mad.ResponseMsgCh and complete the transaction externally
		if t.mad.ResponseMsgCh != nil {
			t.mad.ResponseMsgCh <- ppd
		}
		return

	case <-t.connData.StopSignal:
		log.Warning("Local transaction %d stopped due to closed connection", t.transactionId)
		err = common.ErrTransactionFailedByClosedConnection
		return

	case <-device.signals.stop:
		// not needed in most case, just in case the device is closed first than connection by mistake
		log.Warning("Local transaction %d stopped due to closed device", t.transactionId)
		err = common.ErrTransactionFailedByClosedDevice
		return

	case <-time.After(time.Duration(t.timeout) * time.Millisecond):
		log.Warning("Local transaction %d stopped due to timeout", t.transactionId)
		err = common.ErrTransactionFailedByTimeout
		return
	}
}

// RemoteTransaction
func (c *ConnectionData) AddRemoteTransaction(t *RemoteTransaction) {
	c.RemoteTransactionMutex.Lock()
	defer c.RemoteTransactionMutex.Unlock()

	c.RemoteTransactionMap[t.transactionId] = t

	c.Add(1)
	go t.Run()
}

func (c *ConnectionData) FindRemoteTransaction(id uint64) *RemoteTransaction {
	if c.IsClosed() {
		log.Warning("connection is closed, all transactions are cleared")
		return nil
	}

	c.RemoteTransactionMutex.Lock()
	defer c.RemoteTransactionMutex.Unlock()

	t, found := c.RemoteTransactionMap[id]
	if found {
		return t
	}

	return nil
}

func (t *RemoteTransaction) Run() {
	log.Debug("Remote transaction %d start", t.transactionId)
	defer log.Debug("Remote transaction %d quit", t.transactionId)

	conn := t.connData

	defer func() {
		t.parserData.Destroy()
		close(t.NextMsgCh)

		conn.RemoteTransactionMutex.Lock()
		delete(conn.RemoteTransactionMap, t.transactionId)
		conn.RemoteTransactionMutex.Unlock()

		conn.Done()
	}()

	select {
	case md := <-t.NextMsgCh:
		md.PrevParserData = t.parserData
		conn.Device.SendMsgToPacket(md)
		return

	case <-conn.StopSignal:
		log.Warning("Remote transaction %d stopped due to closed connection", t.transactionId)
		return

	case <-time.After(time.Duration(t.timeout) * time.Millisecond):
		log.Warning("Remote transaction %d stopped due to timeout", t.transactionId)
		return
	}
}
