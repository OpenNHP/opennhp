package core

import "testing"

func TestSendMsgToPacketAfterQueueClosed(t *testing.T) {
	d := &Device{
		msgToPacketQueue: make(chan *MsgData, 1),
	}
	close(d.msgToPacketQueue)

	// A late send after shutdown must return without panicking.
	d.SendMsgToPacket(&MsgData{})
}

func TestRecvPacketToMsgAfterQueueClosed(t *testing.T) {
	d := &Device{
		packetToMsgQueue: make(chan *PacketData, 1),
	}
	close(d.packetToMsgQueue)

	// A late packet after shutdown must return without panicking.
	d.RecvPacketToMsg(&PacketData{})
}
