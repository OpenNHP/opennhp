package agent

import (
	"encoding/base64"
	"encoding/json"

	"github.com/OpenNHP/opennhp/nhp/common"
	"github.com/OpenNHP/opennhp/nhp/core"
	"github.com/OpenNHP/opennhp/nhp/log"
)

func (a *UdpAgent) HandleCookieMessage(ppd *core.PacketParserData) bool {
	defer a.wg.Done()
	a.wg.Add(1)

	// redirect cookie response message to original knock request
	cokMsg := &common.ServerCookieMsg{}
	err := json.Unmarshal(ppd.BodyMessage, cokMsg)

	if err != nil {
		log.Error("agent[HandleCookieMessage] failed to parse %s message: %v", core.HeaderTypeToString(ppd.HeaderType), err)
		return false
	}

	// update cookie
	cokBytes, _ := base64.StdEncoding.DecodeString(cokMsg.Cookie)
	copy(ppd.ConnData.CookieStore.CurrCookie[:], cokBytes)

	transactionId := cokMsg.TransactionId // note this transaction id is in message structure, not the one in packet
	transaction := a.device.FindLocalTransaction(transactionId)
	if transaction == nil {
		log.Error("agent(#%d)[HandleCookieMessage] transaction is not available", transactionId)
		return false
	}

	log.Info("agent(#%d)[HandleCookieMessage] redirect cookie message to original knock transaction", transactionId)
	transaction.ExternalMsgCh <- ppd
	return true
}
