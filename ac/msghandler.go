package ac

import (
	"context"
	"encoding/json"
	"fmt"
	"net"
	"strconv"
	"strings"
	"time"

	"github.com/OpenNHP/opennhp/common"
	"github.com/OpenNHP/opennhp/log"
	"github.com/OpenNHP/opennhp/nhp"
	"github.com/OpenNHP/opennhp/utils"
)

// IP pass mode
const (
	PASS_KNOCK_IP = iota
	PASS_PRE_ACCESS_IP
)

func (d *UdpDoor) HandleACOperations(ppd *nhp.PacketParserData) (err error) {
	defer d.wg.Done()
	d.wg.Add(1)

	acId := d.config.ACId
	dopMsg := &common.ServerACOpsMsg{}
	artMsg := &common.ACOpsResultMsg{}
	transactionId := ppd.SenderId

	// process ac operation
	func() {
		err = json.Unmarshal(ppd.BodyMessage, dopMsg)
		if err != nil {
			log.Error("ac(%s#%d)[HandleACOperations] failed to parse %s message: %v", acId, transactionId, nhp.HeaderTypeToString(ppd.HeaderType), err)
			artMsg.ErrCode = common.ErrJsonParseFailed.ErrorCode()
			artMsg.ErrMsg = err.Error()
			return
		}

		srcAddrs := dopMsg.SourceAddrs
		dstAddrs := dopMsg.DestinationAddrs
		openTimeSec := int(dopMsg.OpenTime)
		tempOpenTimeSec := TempPortOpenTime
		// 1 sec timeout means exit defaultset access, so exit tempset too
		if openTimeSec == 1 {
			tempOpenTimeSec = 1
		}

		// check empty src address
		if len(srcAddrs) == 0 || len(dstAddrs) == 0 {
			log.Error("ac(%s#%d)[HandleACOperations] no source or destination address specified", acId, transactionId)
			err = common.ErrACEmptyPassAddress
			artMsg.ErrCode = common.ErrACEmptyPassAddress.ErrorCode()
			artMsg.ErrMsg = err.Error()
			return
		}

		// ac ipset operations
		if d.ipset == nil {
			log.Error("ac(%s#%d)[HandleACOperations] ipset is nil", acId, transactionId)
			err = common.ErrACIPSetNotFound
			artMsg.ErrCode = common.ErrACIPSetNotFound.ErrorCode()
			artMsg.ErrMsg = err.Error()
			return
		}

		// use ac default ip to override empty destination ip
		if len(d.config.DefaultIp) > 0 {
			for _, addr := range dstAddrs {
				if len(addr.Ip) == 0 {
					addr.Ip = d.config.DefaultIp
				}
			}
		}

		switch d.IpPassMode() {
		// pass the knock ip immediately
		case PASS_KNOCK_IP:
			for _, srcAddr := range srcAddrs {
				var ipType utils.IPTYPE
				var ipNet *net.IPNet
				if strings.Contains(srcAddr.Ip, ":") {
					ipType = utils.IPV6
					_, ipNet, _ = net.ParseCIDR(srcAddr.Ip + "/121")
				} else {
					ipType = utils.IPV4
					_, ipNet, _ = net.ParseCIDR(srcAddr.Ip + "/25")
				}
				log.Debug("src ip is %s, net range is %s", srcAddr, ipNet.String())

				for _, dstAddr := range dstAddrs {
					// for tcp
					if len(dstAddr.Protocol) == 0 || dstAddr.Protocol == "tcp" || dstAddr.Protocol == "any" {
						ipHashStr := fmt.Sprintf("%s,%d,%s", srcAddr.Ip, dstAddr.Port, dstAddr.Ip)
						if dstAddr.Port == 0 {
							ipHashStr = fmt.Sprintf("%s,1-65535,%s", srcAddr.Ip, dstAddr.Ip)
						}

						_, err = d.ipset.Add(ipType, 1, openTimeSec, ipHashStr)
						if err != nil {
							log.Error("ac(%s#%d)[HandleACOperations] add ipset %s error: %v", acId, transactionId, ipHashStr, err)
							err = common.ErrACIPSetOperationFailed
							artMsg.ErrCode = common.ErrACIPSetOperationFailed.ErrorCode()
							artMsg.ErrMsg = err.Error()
							return
						}
					}

					// for udp
					if len(dstAddr.Protocol) == 0 || dstAddr.Protocol == "udp" || dstAddr.Protocol == "any" {
						ipHashStr := fmt.Sprintf("%s,udp:%d,%s", srcAddr.Ip, dstAddr.Port, dstAddr.Ip)
						if dstAddr.Port == 0 {
							ipHashStr = fmt.Sprintf("%s,udp:1-65535,%s", srcAddr.Ip, dstAddr.Ip)
						}

						_, err = d.ipset.Add(ipType, 1, openTimeSec, ipHashStr)
						if err != nil {
							log.Error("ac(%s#%d)[HandleACOperations] add ipset %s error: %v", acId, transactionId, ipHashStr, err)
							err = common.ErrACIPSetOperationFailed
							artMsg.ErrCode = common.ErrACIPSetOperationFailed.ErrorCode()
							artMsg.ErrMsg = err.Error()
							return
						}
					}

					// for icmp ping
					if dstAddr.Port == 0 && (len(dstAddr.Protocol) == 0 || dstAddr.Protocol == "any") {
						for _, dstAddr := range dstAddrs {
							ipHashStr := fmt.Sprintf("%s,icmp:8/0,%s", srcAddr.Ip, dstAddr.Ip)

							_, err = d.ipset.Add(ipType, 1, openTimeSec, ipHashStr)
							if err != nil {
								log.Error("ac(%s#%d)[HandleACOperations] add ipset %s error: %v", acId, transactionId, ipHashStr, err)
								err = common.ErrACIPSetOperationFailed
								artMsg.ErrCode = common.ErrACIPSetOperationFailed.ErrorCode()
								artMsg.ErrMsg = err.Error()
								return
							}
						}
					}

					// add tempset
					if ipNet != nil {
						netStr := ipNet.String()
						if len(dstAddr.Protocol) == 0 || dstAddr.Protocol == "tcp" || dstAddr.Protocol == "any" {
							netHashStr := fmt.Sprintf("%s,%d", netStr, dstAddr.Port)
							if dstAddr.Port == 0 {
								netHashStr = fmt.Sprintf("%s,1-65535", netStr)
							}
							_, err = d.ipset.Add(ipType, 4, tempOpenTimeSec, netHashStr)
						}

						if len(dstAddr.Protocol) == 0 || dstAddr.Protocol == "udp" || dstAddr.Protocol == "any" {
							netHashStr := fmt.Sprintf("%s,udp:%d", netStr, dstAddr.Port)
							if dstAddr.Port == 0 {
								netHashStr = fmt.Sprintf("%s,udp:1-65535", netStr)
							}
							_, err = d.ipset.Add(ipType, 4, tempOpenTimeSec, netHashStr)
						}

						if dstAddr.Port == 0 && (len(dstAddr.Protocol) == 0 || dstAddr.Protocol == "any") {
							netHashStr := fmt.Sprintf("%s,icmp:8/0", netStr)
							_, err = d.ipset.Add(ipType, 4, tempOpenTimeSec, netHashStr)
						}
					}
				}
			}

			// return temporary listened port(s) and nhp access token, then pass the real ip when agent sends access message
		case PASS_PRE_ACCESS_IP:
			fallthrough
		default:
			// door open a temporary tcp or udp port for access
			dstIp := net.ParseIP(dstAddrs[0].Ip)
			if dstIp == nil {
				log.Error("ac(%s#%d)[HandleACOperations] destination IP %s is invalid", acId, transactionId, dstAddrs[0].Ip)
				err = common.ErrInvalidIpAddress
				artMsg.ErrCode = common.ErrInvalidIpAddress.ErrorCode()
				artMsg.ErrMsg = err.Error()
				return
			}

			var ipType utils.IPTYPE
			var netStr string
			var netStr1 string
			var pickedPort int
			var tcpListener *net.TCPListener
			var udpListener *net.UDPConn

			if strings.Contains(dstAddrs[0].Ip, ":") {
				ipType = utils.IPV6
				netStr = "0:0:0:0:0:0:0:0/0"
			} else {
				// since ipset does not allow full ip range 0.0.0.0/0, we use two ip ranges
				ipType = utils.IPV4
				netStr = "0.0.0.0/1"
				netStr1 = "128.0.0.0/1"
			}

			// openning temp tcp access
			tcpListener, err = net.ListenTCP("tcp", &net.TCPAddr{
				IP:   dstIp,
				Port: 0, // ephemeral port
			})

			if err != nil {
				log.Error("ac(%s#%d)[HandleACOperations] temporary tcp listening error: %v", acId, transactionId, err)
				err = common.ErrACTempPortListenFailed
				artMsg.ErrCode = common.ErrACTempPortListenFailed.ErrorCode()
				artMsg.ErrMsg = err.Error()
				return
			}

			// retrieve local port
			tladdr := tcpListener.Addr()
			tlocalAddr, locErr := net.ResolveTCPAddr(tladdr.Network(), tladdr.String())
			if locErr != nil {
				log.Error("ac(%s#%d)[HandleACOperations] resolve local TCPAddr error: %v", acId, transactionId, locErr)
				err = common.ErrACResolveTempPortFailed
				artMsg.ErrCode = common.ErrACResolveTempPortFailed.ErrorCode()
				artMsg.ErrMsg = err.Error()
				return
			}

			log.Debug("open temporary tcp port %s", tlocalAddr.String())
			portHashStr := fmt.Sprintf("%s,%d", netStr, tlocalAddr.Port)
			_, err = d.ipset.Add(ipType, 4, tempOpenTimeSec, portHashStr)
			if err != nil {
				log.Error("ac(%s#%d)[HandleACOperations] add ipset %s error: %v", acId, transactionId, portHashStr, err)
				err = common.ErrACIPSetOperationFailed
				artMsg.ErrCode = common.ErrACIPSetOperationFailed.ErrorCode()
				artMsg.ErrMsg = err.Error()
				return
			}
			portHashStr = fmt.Sprintf("%s,%d", netStr1, tlocalAddr.Port)
			_, err = d.ipset.Add(ipType, 4, tempOpenTimeSec, portHashStr)
			if err != nil {
				log.Error("ac(%s#%d)[HandleACOperations] add ipset %s error: %v", acId, transactionId, portHashStr, err)
				err = common.ErrACIPSetOperationFailed
				artMsg.ErrCode = common.ErrACIPSetOperationFailed.ErrorCode()
				artMsg.ErrMsg = err.Error()
				return
			}

			pickedPort = tlocalAddr.Port
			log.Info("ac(%s#%d)[HandleACOperations] open temporary tcp port on %s", acId, transactionId, tladdr.String())

			// for temp udp access
			udpListener, err = net.ListenUDP("udp", &net.UDPAddr{
				IP:   dstIp,
				Port: pickedPort, // ephemeral port(0) or continue with previously picked tcp port
			})
			if err != nil {
				log.Error("ac(%s#%d)[HandleACOperations] temporary udp listening error: %v", acId, transactionId, err)
				err = common.ErrACTempPortListenFailed
				artMsg.ErrCode = common.ErrACTempPortListenFailed.ErrorCode()
				artMsg.ErrMsg = err.Error()
				return
			}

			// retrieve local port
			uladdr := udpListener.LocalAddr()
			_, locErr = net.ResolveUDPAddr(uladdr.Network(), uladdr.String())
			if locErr != nil {
				log.Error("ac(%s#%d)[HandleACOperations] resolve local UDPAddr error: %v", acId, transactionId, locErr)
				err = common.ErrACResolveTempPortFailed
				artMsg.ErrCode = common.ErrACResolveTempPortFailed.ErrorCode()
				artMsg.ErrMsg = err.Error()
				return
			}

			log.Debug("open temporary udp port %s", tlocalAddr.String())
			pickedPort = tlocalAddr.Port
			portHashStr = fmt.Sprintf("%s,udp:%d", netStr, tlocalAddr.Port)
			_, err = d.ipset.Add(ipType, 4, tempOpenTimeSec, portHashStr)
			if err != nil {
				log.Error("ac(%s#%d)[HandleACOperations] add ipset %s error: %v", acId, transactionId, portHashStr, err)
				err = common.ErrACIPSetOperationFailed
				artMsg.ErrCode = common.ErrACIPSetOperationFailed.ErrorCode()
				artMsg.ErrMsg = err.Error()
				return
			}
			portHashStr = fmt.Sprintf("%s,udp:%d", netStr1, tlocalAddr.Port)
			_, err = d.ipset.Add(ipType, 4, tempOpenTimeSec, portHashStr)
			if err != nil {
				log.Error("ac(%s#%d)[HandleACOperations] add ipset %s error: %v", acId, transactionId, portHashStr, err)
				err = common.ErrACIPSetOperationFailed
				artMsg.ErrCode = common.ErrACIPSetOperationFailed.ErrorCode()
				artMsg.ErrMsg = err.Error()
				return
			}

			log.Info("ac(%s#%d)[HandleACOperations] open temporary udp port on %s", acId, transactionId, tladdr.String())

			agentUser := &AgentUser{
				UserId:         dopMsg.UserId,
				DeviceId:       dopMsg.DeviceId,
				OrganizationId: dopMsg.OrganizationId,
			}

			artMsg.PreAccessAction = &common.PreAccessInfo{
				AccessPort: strconv.Itoa(pickedPort),
				ACPubKey:   d.device.PublicKeyExBase64(),
				ACToken:    d.GenerateAccessToken(agentUser),
			}

			if tcpListener != nil {
				d.wg.Add(1)
				go d.tcpTempAccessHandler(transactionId, tcpListener, agentUser, tempOpenTimeSec, dopMsg)
			}

			if udpListener != nil {
				d.wg.Add(1)
				go d.udpTempAccessHandler(transactionId, udpListener, agentUser, tempOpenTimeSec, dopMsg)
			}
		}

		log.Info("ac(%s#%d)[HandleACOperations] succeed", acId, transactionId)
		artMsg.ErrCode = common.ErrSuccess.ErrorCode()
		artMsg.OpenTime = dopMsg.OpenTime
	}()

	// send ac result
	artBytes, _ := json.Marshal(artMsg)

	md := &nhp.MsgData{
		HeaderType:     nhp.NHP_ART,
		TransactionId:  transactionId,
		Compress:       true,
		PrevParserData: ppd,
		Message:        artBytes,
	}

	// forward to a specific transaction
	transaction := ppd.ConnData.FindRemoteTransaction(transactionId)
	if transaction == nil {
		log.Error("ac(%s#%d)[HandleACOperations] transaction is not available", acId, transactionId)
		err = common.ErrTransactionIdNotFound
		return err
	}

	transaction.NextMsgCh <- md

	return nil
}

func (d *UdpDoor) tcpTempAccessHandler(transactionId uint64, listener *net.TCPListener, au *AgentUser, timeoutSec int, dopMsg *common.ServerACOpsMsg) {
	defer d.wg.Done()
	defer d.DeleteAccessToken(au)
	defer listener.Close()

	acId := d.config.ACId
	// accept only the first incoming tcp connection
	startTime := time.Now()
	deadlineTime := startTime.Add(time.Duration(timeoutSec) * time.Second)
	localAddrStr := listener.Addr().String()
	err := listener.SetDeadline(deadlineTime)
	if err != nil {
		log.Error("ac(%s#%d)[HandleACOperations] temporary port on %s failed to set tcp listen timeout", acId, transactionId, localAddrStr)
		return
	}
	conn, err := listener.Accept()
	if err != nil {
		log.Error("ac(%s#%d)[HandleACOperations] temporary port on %s tcp listen timeout", acId, transactionId, localAddrStr)
		return
	}

	defer conn.Close()
	err = conn.SetDeadline(deadlineTime)
	if err != nil {
		log.Error("ac(%s#%d)[HandleACOperations] temporary port on %s failed to set tcp conn timeout", acId, transactionId, localAddrStr)
		return
	}

	remoteAddrStr := conn.RemoteAddr().String()
	pkt := d.device.AllocateUdpPacket()
	defer d.device.ReleaseUdpPacket(pkt)

	// monitor stop signals and quit connection earlier
	ctx, ctxCancel := context.WithDeadline(context.Background(), deadlineTime)
	defer ctxCancel()
	go d.tempConnTerminator(conn, ctx)

	// tcp recv common header first
	n, err := conn.Read(pkt.Buf[:nhp.HeaderCommonSize])
	if err != nil || n < nhp.HeaderCommonSize {
		log.Error("ac(%s#%d)[HandleACOperations] failed to receive tcp packet header from remote address %s (%v)", acId, transactionId, remoteAddrStr, err)
		return
	}

	pkt.Packet = pkt.Buf[:n]
	// check type and payload size
	msgType, msgSize := pkt.HeaderTypeAndSize()
	if msgType != nhp.NHP_ACC {
		log.Error("ac(%s#%d)[HandleACOperations] message type is not %s, close connection", acId, transactionId, nhp.HeaderTypeToString(nhp.NHP_ACC))
		return
	}

	var packetSize int
	flag := pkt.Flag()
	if flag&nhp.NHP_FLAG_EXTENDEDLENGTH == 0 {
		packetSize = nhp.HeaderSize + msgSize
	} else {
		packetSize = nhp.HeaderSizeEx + msgSize
	}
	remainingSize := packetSize - n
	n, err = conn.Read(pkt.Buf[n:packetSize])
	if err != nil || n < remainingSize {
		log.Error("ac(%s#%d)[HandleACOperations] failed to receive tcp message body from remote address %s (%v)", acId, transactionId, remoteAddrStr, err)
		return
	}

	pkt.Packet = pkt.Buf[:packetSize]
	log.Trace("receive tcp access packet (%s -> %s): %+v", remoteAddrStr, localAddrStr, pkt.Packet)
	log.Info("ac(%s#%d)[HandleACOperations] receive tcp access message (%s -> %s)", acId, transactionId, remoteAddrStr, localAddrStr)

	pd := &nhp.PacketData{
		BasePacket:     pkt,
		ConnData:       &nhp.ConnectionData{},
		InitTime:       time.Now().UnixNano(),
		HeaderType:     msgType,
		DecryptedMsgCh: make(chan *nhp.PacketParserData),
	}

	if !d.IsRunning() {
		log.Error("ac(%s#%d)[HandleACOperations] PacketData channel closed or being closed, skip decrypting", acId, transactionId)
		return
	}

	// start message decryption
	d.device.RecvPacketToMsg(pd)

	// waiting for message decryption
	accPpd := <-pd.DecryptedMsgCh
	close(pd.DecryptedMsgCh)

	if accPpd.Error != nil {
		log.Error("ac(%s#%d)[HandleACOperations] failed to decrypt tcp access message: %v", acId, transactionId, accPpd.Error)
		return
	}

	accMsg := &common.AgentAccessMsg{}
	err = json.Unmarshal(accPpd.BodyMessage, accMsg)
	if err != nil {
		log.Error("ac(%s#%d)[HandleACOperations] failed to parse %s message: %v", acId, transactionId, nhp.HeaderTypeToString(accPpd.HeaderType), err)
		return
	}

	remoteAgentUser := &AgentUser{
		UserId:         accMsg.UserId,
		DeviceId:       accMsg.DeviceId,
		OrganizationId: accMsg.OrganizationId,
	}

	if d.VerifyAccessToken(remoteAgentUser, accMsg.ACToken) {
		remoteAddr, _ := net.ResolveTCPAddr(conn.RemoteAddr().Network(), conn.RemoteAddr().String())
		srcAddrIp := remoteAddr.IP.String()
		dstAddrs := dopMsg.DestinationAddrs
		openTimeSec := int(dopMsg.OpenTime)
		var ipType utils.IPTYPE
		if strings.Contains(dstAddrs[0].Ip, ":") {
			ipType = utils.IPV6
		} else {
			ipType = utils.IPV4
		}

		for _, dstAddr := range dstAddrs {
			ipHashStr := fmt.Sprintf("%s,%d,%s", srcAddrIp, dstAddr.Port, dstAddr.Ip)
			if dstAddr.Port == 0 {
				ipHashStr = fmt.Sprintf("%s,1-65535,%s", srcAddrIp, dstAddr.Ip)
			}

			_, err = d.ipset.Add(ipType, 1, openTimeSec, ipHashStr)
			if err != nil {
				log.Error("ac(%s#%d)[HandleACOperations] add ipset %s error: %v", acId, transactionId, ipHashStr, err)
				return
			}
		}
	}
}

func (d *UdpDoor) udpTempAccessHandler(transactionId uint64, conn *net.UDPConn, au *AgentUser, timeoutSec int, dopMsg *common.ServerACOpsMsg) {
	defer d.wg.Done()
	defer d.DeleteAccessToken(au)
	defer conn.Close()

	acId := d.config.ACId
	// listen to accept and handle only one incoming connection
	startTime := time.Now()
	deadlineTime := startTime.Add(time.Duration(timeoutSec) * time.Second)
	localAddrStr := conn.LocalAddr().String()
	err := conn.SetDeadline(deadlineTime)
	if err != nil {
		log.Error("ac(%s#%d)[HandleACOperations] temporary port on %s failed to set udp conn timeout", acId, transactionId, localAddrStr)
		return
	}

	pkt := d.device.AllocateUdpPacket()
	defer d.device.ReleaseUdpPacket(pkt)

	// monitor stop signals and quit connection earlier
	ctx, ctxCancel := context.WithDeadline(context.Background(), deadlineTime)
	defer ctxCancel()
	go d.tempConnTerminator(conn, ctx)

	// udp recv, blocking until packet arrives or deadline reaches
	n, remoteAddr, err := conn.ReadFromUDP(pkt.Buf[:])
	if err != nil || n < nhp.HeaderCommonSize {
		log.Error("ac(%s#%d)[HandleACOperations] failed to receive udp packet (%v)", acId, transactionId, err)
		return
	}

	remoteAddrStr := remoteAddr.String()
	pkt.Packet = pkt.Buf[:n]

	// check type and payload size
	msgType, msgSize := pkt.HeaderTypeAndSize()
	if msgType != nhp.NHP_ACC {
		log.Error("ac(%s#%d)[HandleACOperations] message type is not %s, close connection", acId, transactionId, nhp.HeaderTypeToString(nhp.NHP_ACC))
		return
	}

	var packetSize int
	flag := pkt.Flag()
	if flag&nhp.NHP_FLAG_EXTENDEDLENGTH == 0 {
		packetSize = nhp.HeaderSize + msgSize
	} else {
		packetSize = nhp.HeaderSizeEx + msgSize
	}

	if n != packetSize {
		log.Error("ac(%s#%d)[HandleACOperations] udp packet size incorrect from remote address %s", acId, transactionId, remoteAddrStr)
		return
	}

	log.Trace("receive udp access packet (%s -> %s): %+v", remoteAddrStr, localAddrStr, pkt.Packet)
	log.Info("ac(%s#%d)[HandleACOperations] receive udp access message (%s -> %s)", acId, transactionId, remoteAddrStr, localAddrStr)

	pd := &nhp.PacketData{
		BasePacket:     pkt,
		ConnData:       &nhp.ConnectionData{},
		InitTime:       time.Now().UnixNano(),
		HeaderType:     msgType,
		DecryptedMsgCh: make(chan *nhp.PacketParserData),
	}

	if !d.IsRunning() {
		log.Error("ac(%s#%d)[HandleACOperations] PacketData channel closed or being closed, skip decrypting", acId, transactionId)
		return
	}

	// start packet decryption
	d.device.RecvPacketToMsg(pd)

	// waiting for packet decryption
	accPpd := <-pd.DecryptedMsgCh
	close(pd.DecryptedMsgCh)

	if accPpd.Error != nil {
		log.Error("ac(%s#%d)[HandleACOperations] failed to decrypt udp access message: %v", acId, transactionId, accPpd.Error)
		return
	}

	accMsg := &common.AgentAccessMsg{}
	err = json.Unmarshal(accPpd.BodyMessage, accMsg)
	if err != nil {
		log.Error("ac(%s#%d)[HandleACOperations] failed to parse %s message: %v", acId, transactionId, nhp.HeaderTypeToString(accPpd.HeaderType), err)
		return
	}

	remoteAgentUser := &AgentUser{
		UserId:         accMsg.UserId,
		DeviceId:       accMsg.DeviceId,
		OrganizationId: accMsg.OrganizationId,
	}

	if d.VerifyAccessToken(remoteAgentUser, accMsg.ACToken) {
		srcAddrIp := remoteAddr.IP.String()
		dstAddrs := dopMsg.DestinationAddrs
		openTimeSec := int(dopMsg.OpenTime)
		var ipType utils.IPTYPE
		if strings.Contains(dstAddrs[0].Ip, ":") {
			ipType = utils.IPV6
		} else {
			ipType = utils.IPV4
		}

		for _, dstAddr := range dstAddrs {
			if len(dstAddr.Protocol) == 0 || dstAddr.Protocol == "udp" || dstAddr.Protocol == "any" {
				ipHashStr := fmt.Sprintf("%s,udp:%d,%s", srcAddrIp, dstAddr.Port, dstAddr.Ip)
				if dstAddr.Port == 0 {
					ipHashStr = fmt.Sprintf("%s,udp:1-65535,%s", srcAddrIp, dstAddr.Ip)
				}

				_, err = d.ipset.Add(ipType, 1, openTimeSec, ipHashStr)
				if err != nil {
					log.Error("ac(%s#%d)[HandleACOperations] add ipset %s error: %v", acId, transactionId, ipHashStr, err)
					return
				}
			}

			// for ping
			if dstAddr.Port == 0 && (len(dstAddr.Protocol) == 0 || dstAddr.Protocol == "any") {
				ipHashStr := fmt.Sprintf("%s,icmp:8/0,%s", remoteAddr.IP.String(), dstAddr.Ip)
				d.ipset.Add(ipType, 1, openTimeSec, ipHashStr)
			}
		}
	}
}

func (d *UdpDoor) tempConnTerminator(conn net.Conn, ctx context.Context) {
	select {
	case <-d.signals.stop:
		conn.Close()
		return

	case <-ctx.Done():
		return
	}
}
