package ac

import (
	"context"
	"encoding/json"
	"fmt"
	"net"
	"strconv"
	"time"

	"github.com/OpenNHP/opennhp/nhp/common"
	"github.com/OpenNHP/opennhp/nhp/core"
	"github.com/OpenNHP/opennhp/nhp/log"
	"github.com/OpenNHP/opennhp/nhp/utils"
	"github.com/OpenNHP/opennhp/nhp/utils/ebpf"
)

// IP pass mode
const (
	PASS_KNOCK_IP = iota
	PASS_KNOCKIP_WITH_RANGE
	PASS_PRE_ACCESS_IP
)

func (a *UdpAC) HandleUdpACOperations(ppd *core.PacketParserData) (err error) {
	defer a.wg.Done()

	acId := a.config.ACId
	dopMsg := &common.ServerACOpsMsg{}
	artMsg := &common.ACOpsResultMsg{}
	transactionId := ppd.SenderTrxId

	err = json.Unmarshal(ppd.BodyMessage, dopMsg)
	if err != nil {
		log.Error("ac(%s#%d)[HandleUdpACOperations] failed to parse %s message: %v", acId, transactionId, core.HeaderTypeToString(ppd.HeaderType), err)
		artMsg.ErrCode = common.ErrJsonParseFailed.ErrorCode()
		artMsg.ErrMsg = err.Error()
		return
	}

	srcAddrs := dopMsg.SourceAddrs
	dstAddrs := dopMsg.DestinationAddrs
	openTimeSec := int(dopMsg.OpenTime)
	agentUser := &common.AgentUser{
		UserId:         dopMsg.UserId,
		DeviceId:       dopMsg.DeviceId,
		OrganizationId: dopMsg.OrganizationId,
		AuthServiceId:  dopMsg.AuthServiceId,
	}
	artMsg, err = a.HandleAccessControl(agentUser, srcAddrs, dstAddrs, openTimeSec, artMsg)
	if err != nil {
		log.Error("ac(%s#%d)[HandleUdpACOperations] HandleAccessControl failed, err: %v", acId, transactionId, err)
	}

	// generate ac token and save user and access information
	entry := &AccessEntry{
		User:     agentUser,
		SrcAddrs: srcAddrs,
		DstAddrs: dstAddrs,
		OpenTime: openTimeSec,
	}
	artMsg.ACToken = a.GenerateAccessToken(entry)
	//log.Info("generate knock token: %s", artMsg.ACToken)

	// send ac result
	artBytes, _ := json.Marshal(artMsg)
	md := &core.MsgData{
		HeaderType:     core.NHP_ART,
		TransactionId:  transactionId,
		Compress:       true,
		PrevParserData: ppd,
		Message:        artBytes,
	}
	//log.Info("ART result: %s", string(artBytes))

	// forward to a specific transaction
	transaction := ppd.ConnData.FindRemoteTransaction(transactionId)
	if transaction == nil {
		log.Error("ac(%s#%d)[HandleUdpACOperations] transaction is not available", acId, transactionId)
		err = common.ErrTransactionIdNotFound
		return err
	}

	transaction.NextMsgCh <- md

	return err
}

func (a *UdpAC) HandleAccessControl(au *common.AgentUser, srcAddrs []*common.NetAddress, dstAddrs []*common.NetAddress, openTimeSec int, artMsgIn *common.ACOpsResultMsg) (artMsg *common.ACOpsResultMsg, err error) {
	if artMsgIn == nil {
		artMsg = &common.ACOpsResultMsg{}
	} else {
		artMsg = artMsgIn
	}
	// process ac operation
	tempOpenTimeSec := TempPortOpenTime
	// 1 sec timeout means exit defaultset access, so exit tempset too
	if openTimeSec == 1 {
		tempOpenTimeSec = 1
	}

	// check empty src address
	if len(srcAddrs) == 0 || len(dstAddrs) == 0 {
		log.Error("[HandleAccessControl] no source or destination address specified")
		err = common.ErrACEmptyPassAddress
		artMsg.ErrCode = common.ErrACEmptyPassAddress.ErrorCode()
		artMsg.ErrMsg = err.Error()
		return
	}

	// ac ipset operations
	if a.config.FilterMode == FilterMode_IPTABLES {
		if a.ipset == nil {
			log.Error("[HandleAccessControl] ipset is nil")
			err = common.ErrACIPSetNotFound
			artMsg.ErrCode = common.ErrACIPSetNotFound.ErrorCode()
			artMsg.ErrMsg = err.Error()
			return
		}
	}

	// use ac default ip to override empty destination ip
	if len(a.config.DefaultIp) > 0 {
		for _, addr := range dstAddrs {
			if len(addr.Ip) == 0 {
				addr.Ip = a.config.DefaultIp
			}
		}
	}

	ipPassMode := a.IpPassMode()
	switch ipPassMode {
	// pass the knock ip immediately
	case PASS_KNOCKIP_WITH_RANGE:
		fallthrough
	case PASS_KNOCK_IP:
		fallthrough
	default:
		for _, srcAddr := range srcAddrs {
			var ipNet *net.IPNet

			// Detect IP type using proper parsing instead of string matching
			ipType, ipErr := utils.DetectIPType(srcAddr.Ip)
			if ipErr != nil {
				log.Error("[HandleAccessControl] invalid source IP: %s, error: %v", srcAddr.Ip, ipErr)
				continue
			}

			// Use appropriate CIDR mask based on IP type and pass mode
			rangeMode := ipPassMode == PASS_KNOCKIP_WITH_RANGE
			cidrMask := utils.GetCIDRMask(ipType, rangeMode)
			_, ipNet, _ = net.ParseCIDR(srcAddr.Ip + cidrMask)
			log.Debug("src ip is %s, net range is %s", srcAddr, ipNet.String())

			for _, dstAddr := range dstAddrs {
				// for tcp
				if len(dstAddr.Protocol) == 0 || dstAddr.Protocol == "tcp" || dstAddr.Protocol == "any" {
					ipHashStr := fmt.Sprintf("%s,%d,%s", srcAddr.Ip, dstAddr.Port, dstAddr.Ip)
					if dstAddr.Port == 0 {
						ipHashStr = fmt.Sprintf("%s,1-65535,%s", srcAddr.Ip, dstAddr.Ip)
					}

					switch a.config.FilterMode {
					case FilterMode_IPTABLES:
						_, err = a.ipset.Add(ipType, 1, openTimeSec, ipHashStr)
						if err != nil {
							log.Error("[HandleAccessControl] add ipset %s error: %v", ipHashStr, err)
							err = common.ErrACIPSetOperationFailed
							artMsg.ErrCode = common.ErrACIPSetOperationFailed.ErrorCode()
							artMsg.ErrMsg = err.Error()
							return
						}
					//ebpf knock
					case FilterMode_EBPFXDP:
						if len(dstAddr.Protocol) == 0 || dstAddr.Protocol == "any" {
							ebpfHashStr := ebpf.EbpfRuleParams{
								SrcIP: srcAddr.Ip,
								DstIP: dstAddr.Ip,
							}
							err = ebpf.EbpfRuleAdd(2, ebpfHashStr, openTimeSec)
							if err != nil {
								log.Error("[EbpfRuleAdd] add ebpf src: %s dst: %s, error: %v", ebpfHashStr.SrcIP, ebpfHashStr.DstIP, err)
								return
							}
						}
						if dstAddr.Protocol == "tcp" {
							ebpfHashStr := ebpf.EbpfRuleParams{
								SrcIP:    srcAddr.Ip,
								DstIP:    dstAddr.Ip,
								DstPort:  dstAddr.Port,
								Protocol: dstAddr.Protocol,
							}
							err = ebpf.EbpfRuleAdd(1, ebpfHashStr, openTimeSec)
							if err != nil {
								log.Error("[EbpfRuleAdd] add ebpf tcp failed src: %s dst: %s, error: %v, protocol: %s, dstport: %d", ebpfHashStr.SrcIP, ebpfHashStr.DstIP, err, ebpfHashStr.Protocol, ebpfHashStr.DstPort)
								return
							}
						}
					default:
						log.Error("[HandleAccessControl] unsupported FilterMode: %d (expected 0=IPTABLES or 1=EBPFXDP)", a.config.FilterMode)
						return
					}
				}

				// for udp
				if len(dstAddr.Protocol) == 0 || dstAddr.Protocol == "udp" || dstAddr.Protocol == "any" {
					ipHashStr := fmt.Sprintf("%s,udp:%d,%s", srcAddr.Ip, dstAddr.Port, dstAddr.Ip)
					if dstAddr.Port == 0 {
						ipHashStr = fmt.Sprintf("%s,udp:1-65535,%s", srcAddr.Ip, dstAddr.Ip)
					}

					switch a.config.FilterMode {
					case FilterMode_IPTABLES:
						_, err = a.ipset.Add(ipType, 1, openTimeSec, ipHashStr)
						if err != nil {
							log.Error("[HandleAccessControl] add ipset %s error: %v", ipHashStr, err)
							err = common.ErrACIPSetOperationFailed
							artMsg.ErrCode = common.ErrACIPSetOperationFailed.ErrorCode()
							artMsg.ErrMsg = err.Error()
							return
						}
					case FilterMode_EBPFXDP:
						if len(dstAddr.Protocol) == 0 || dstAddr.Protocol == "any" {
							ebpfHashStr := ebpf.EbpfRuleParams{
								SrcIP: srcAddr.Ip,
								DstIP: dstAddr.Ip,
							}
							err = ebpf.EbpfRuleAdd(2, ebpfHashStr, openTimeSec)
							if err != nil {
								log.Error("[EbpfRuleAdd] add ebpf src: %s dst: %s, error: %v", ebpfHashStr.SrcIP, ebpfHashStr.DstIP, err)
								return
							}
						}
						if dstAddr.Protocol == "udp" {
							ebpfHashStr := ebpf.EbpfRuleParams{
								SrcIP:    srcAddr.Ip,
								DstIP:    dstAddr.Ip,
								DstPort:  dstAddr.Port,
								Protocol: dstAddr.Protocol,
							}
							err = ebpf.EbpfRuleAdd(1, ebpfHashStr, openTimeSec)

							if err != nil {
								log.Error("[EbpfRuleAdd] add ebpf udp failed src: %s dst: %s, error: %v, protocol: %s, dstport: %d", ebpfHashStr.SrcIP, ebpfHashStr.DstIP, err, ebpfHashStr.Protocol, ebpfHashStr.DstPort)
								return
							}
						}
					default:
						log.Error("[HandleAccessControl] unsupported FilterMode: %d (expected 0=IPTABLES or 1=EBPFXDP)", a.config.FilterMode)
						return
					}
				}

				// for icmp ping
				if dstAddr.Port == 0 && (len(dstAddr.Protocol) == 0 || dstAddr.Protocol == "any") {
					for _, dstAddr := range dstAddrs {
						ipHashStr := fmt.Sprintf("%s,icmp:8/0,%s", srcAddr.Ip, dstAddr.Ip)
						switch a.config.FilterMode {
						case FilterMode_IPTABLES:
							_, err = a.ipset.Add(ipType, 1, openTimeSec, ipHashStr)
							if err != nil {
								log.Error("[HandleAccessControl] add ipset %s error: %v", ipHashStr, err)
								err = common.ErrACIPSetOperationFailed
								artMsg.ErrCode = common.ErrACIPSetOperationFailed.ErrorCode()
								artMsg.ErrMsg = err.Error()
								return
							}
						case FilterMode_EBPFXDP:
							ebpfHashStr := ebpf.EbpfRuleParams{
								SrcIP: srcAddr.Ip,
								DstIP: dstAddr.Ip,
							}
							err = ebpf.EbpfRuleAdd(3, ebpfHashStr, openTimeSec)
							if err != nil {
								log.Error("[EbpfRuleAdd] add ebpf src: %s dst: %s, error: %v", ebpfHashStr.SrcIP, ebpfHashStr.DstIP, err)
								return
							}
						default:
							log.Error("[HandleAccessControl] unsupported FilterMode: %d (expected 0=IPTABLES or 1=EBPFXDP)", a.config.FilterMode)
							return
						}
					}
				}

				// add tempset for the adjacent 128 (25bit netmask ipv4, 121bit netmask ipv6) addresses derived from the target IP address
				if ipPassMode == PASS_KNOCKIP_WITH_RANGE && ipNet != nil {
					netStr := ipNet.String()
					switch a.config.FilterMode {
					case FilterMode_IPTABLES:
						if len(dstAddr.Protocol) == 0 || dstAddr.Protocol == "tcp" || dstAddr.Protocol == "any" {
							netHashStr := fmt.Sprintf("%s,%d", netStr, dstAddr.Port)
							if dstAddr.Port == 0 {
								netHashStr = fmt.Sprintf("%s,1-65535", netStr)
							}
							_, err = a.ipset.Add(ipType, 4, tempOpenTimeSec, netHashStr)
						}

						if len(dstAddr.Protocol) == 0 || dstAddr.Protocol == "udp" || dstAddr.Protocol == "any" {
							netHashStr := fmt.Sprintf("%s,udp:%d", netStr, dstAddr.Port)
							if dstAddr.Port == 0 {
								netHashStr = fmt.Sprintf("%s,udp:1-65535", netStr)
							}
							_, err = a.ipset.Add(ipType, 4, tempOpenTimeSec, netHashStr)
						}

						if dstAddr.Port == 0 && (len(dstAddr.Protocol) == 0 || dstAddr.Protocol == "any") {
							netHashStr := fmt.Sprintf("%s,icmp:8/0", netStr)
							_, err = a.ipset.Add(ipType, 4, tempOpenTimeSec, netHashStr)
						}

					case FilterMode_EBPFXDP:
						cidrIP, ipnet, cidrErr := net.ParseCIDR(netStr)
						if cidrErr != nil {
							log.Error("[HandleAccessControl] failed to parse CIDR %s: %v", netStr, cidrErr)
						}
						if len(dstAddr.Protocol) == 0 || dstAddr.Protocol == "tcp" || dstAddr.Protocol == "any" {
							for iterIP := cidrIP.Mask(ipnet.Mask); ipnet.Contains(iterIP); incrementIP(iterIP) {
								srcIpStr := iterIP.String()
								if dstAddr.Port != 0 {
									ebpfHashStr := ebpf.EbpfRuleParams{
										SrcIP:   srcIpStr,
										DstPort: dstAddr.Port,
									}
									addErr := ebpf.EbpfRuleAdd(4, ebpfHashStr, tempOpenTimeSec)
									if addErr != nil {
										log.Error("[EbpfRuleAdd] add ebpf for tcp dst port src: %s, dstport: %d, error: %v", ebpfHashStr.SrcIP, ebpfHashStr.DstPort, addErr)
									}

								} else {
									ebpfHashStr := ebpf.EbpfRuleParams{
										SrcIP:        srcIpStr,
										DstPortStart: 1,
										DstPortEnd:   65535,
									}
									addErr := ebpf.EbpfRuleAdd(5, ebpfHashStr, tempOpenTimeSec)
									if addErr != nil {
										log.Error("[EbpfRuleAdd] add ebpf src: %s dstportstart: %d, dstportend: %d, error: %v", ebpfHashStr.SrcIP, ebpfHashStr.DstPortStart, ebpfHashStr.DstPortEnd, addErr)
									}
								}
							}
						}
						if len(dstAddr.Protocol) == 0 || dstAddr.Protocol == "udp" || dstAddr.Protocol == "any" {
							for iterIP := cidrIP.Mask(ipnet.Mask); ipnet.Contains(iterIP); incrementIP(iterIP) {
								srcIpStr := iterIP.String()

								if dstAddr.Port != 0 {
									ebpfHashStr := ebpf.EbpfRuleParams{
										SrcIP:   srcIpStr,
										DstPort: dstAddr.Port,
									}
									addErr := ebpf.EbpfRuleAdd(4, ebpfHashStr, tempOpenTimeSec)
									if addErr != nil {
										log.Error("[EbpfRuleAdd] add ebpf for udp dst port src: %s, dstport: %d, error: %v", ebpfHashStr.SrcIP, ebpfHashStr.DstPort, addErr)
									}
								} else {
									ebpfHashStr := ebpf.EbpfRuleParams{
										SrcIP:        srcIpStr,
										DstPortStart: 1,
										DstPortEnd:   65535,
									}
									addErr := ebpf.EbpfRuleAdd(5, ebpfHashStr, tempOpenTimeSec)
									if addErr != nil {
										log.Error("[EbpfRuleAdd] add ebpf src: %s dstportstart: %d, dstportend: %d, error: %v", ebpfHashStr.SrcIP, ebpfHashStr.DstPortStart, ebpfHashStr.DstPortEnd, addErr)
									}
								}
							}
						}
						if dstAddr.Port == 0 && (len(dstAddr.Protocol) == 0 || dstAddr.Protocol == "any") {
							for iterIP := cidrIP.Mask(ipnet.Mask); ipnet.Contains(iterIP); incrementIP(iterIP) {
								srcIpStr := iterIP.String()
								ebpfHashStr := ebpf.EbpfRuleParams{
									SrcIP: srcIpStr,
									DstIP: dstAddr.Ip,
								}
								addErr := ebpf.EbpfRuleAdd(3, ebpfHashStr, openTimeSec)
								if addErr != nil {
									log.Error("[EbpfRuleAdd] add ebpf src: %s dst: %s, error: %v", ebpfHashStr.SrcIP, ebpfHashStr.DstIP, addErr)
								}
							}
						}
					default:
						log.Error("[HandleAccessControl] unsupported FilterMode: %d (expected 0=IPTABLES or 1=EBPFXDP)", a.config.FilterMode)
						return
					}
				}
			}
		}

		// return temporary listened port(s) and nhp access token, then pass the real ip when agent sends access message
	case PASS_PRE_ACCESS_IP:
		// ac open a temporary tcp or udp port for access
		dstIp := net.ParseIP(dstAddrs[0].Ip)
		if dstIp == nil {
			log.Error("[HandleAccessControl] destination IP %s is invalid", dstAddrs[0].Ip)
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

		// Detect IP type using proper parsing instead of string matching
		ipType, ipErr := utils.DetectIPType(dstAddrs[0].Ip)
		if ipErr != nil {
			log.Error("[HandleAccessControl] invalid destination IP for PASS_PRE_ACCESS_IP: %s", dstAddrs[0].Ip)
			err = common.ErrInvalidIpAddress
			artMsg.ErrCode = common.ErrInvalidIpAddress.ErrorCode()
			artMsg.ErrMsg = err.Error()
			return
		}
		if ipType == utils.IPV6 {
			netStr = "::/0" // Canonical IPv6 "any" notation
		} else {
			// since ipset does not allow full ip range 0.0.0.0/0, we use two ip ranges
			netStr = "0.0.0.0/1"
			netStr1 = "128.0.0.0/1"
		}

		// openning temp tcp access
		tcpListener, err = net.ListenTCP("tcp", &net.TCPAddr{
			IP:   dstIp,
			Port: 0, // ephemeral port
		})

		if err != nil {
			log.Error("[HandleAccessControl] temporary tcp listening error: %v", err)
			err = common.ErrACTempPortListenFailed
			artMsg.ErrCode = common.ErrACTempPortListenFailed.ErrorCode()
			artMsg.ErrMsg = err.Error()
			return
		}

		// retrieve local port
		tladdr := tcpListener.Addr()
		tlocalAddr, locErr := net.ResolveTCPAddr(tladdr.Network(), tladdr.String())
		if locErr != nil {
			log.Error("[HandleAccessControl] resolve local TCPAddr error: %v", locErr)
			err = common.ErrACResolveTempPortFailed
			artMsg.ErrCode = common.ErrACResolveTempPortFailed.ErrorCode()
			artMsg.ErrMsg = err.Error()
			return
		}

		log.Debug("open temporary tcp port %s", tlocalAddr.String())
		switch a.config.FilterMode {
		case FilterMode_IPTABLES:
			portHashStr := fmt.Sprintf("%s,%d", netStr, tlocalAddr.Port)
			_, err = a.ipset.Add(ipType, 4, tempOpenTimeSec, portHashStr)

			if err != nil {
				log.Error("[HandleAccessControl] add ipset %s error: %v", portHashStr, err)
				err = common.ErrACIPSetOperationFailed
				artMsg.ErrCode = common.ErrACIPSetOperationFailed.ErrorCode()
				artMsg.ErrMsg = err.Error()
				return
			}
			portHashStr = fmt.Sprintf("%s,%d", netStr1, tlocalAddr.Port)
			_, err = a.ipset.Add(ipType, 4, tempOpenTimeSec, portHashStr)
			if err != nil {
				log.Error("[HandleAccessControl] add ipset %s error: %v", portHashStr, err)
				err = common.ErrACIPSetOperationFailed
				artMsg.ErrCode = common.ErrACIPSetOperationFailed.ErrorCode()
				artMsg.ErrMsg = err.Error()
				return
			}
		case FilterMode_EBPFXDP:
			ebpfHashStr := ebpf.EbpfRuleParams{
				Protocol: "tcp",
				DstPort:  tlocalAddr.Port,
			}
			err = ebpf.EbpfRuleAdd(6, ebpfHashStr, tempOpenTimeSec)
			if err != nil {
				log.Error("[EbpfRuleAdd] add ebpf type 6 protocol: %s, dstport :%d, %v", ebpfHashStr.Protocol, ebpfHashStr.DstPort, err)
				return
			}
		default:
			log.Error("[HandleAccessControl] unsupported FilterMode: %d (expected 0=IPTABLES or 1=EBPFXDP)", a.config.FilterMode)
			return
		}

		pickedPort = tlocalAddr.Port
		log.Info("[HandleAccessControl] open temporary tcp port on %s", tladdr.String())

		// for temp udp access
		udpListener, err = net.ListenUDP("udp", &net.UDPAddr{
			IP:   dstIp,
			Port: pickedPort, // ephemeral port(0) or continue with previously picked tcp port
		})
		if err != nil {
			log.Error("[HandleAccessControl] temporary udp listening error: %v", err)
			err = common.ErrACTempPortListenFailed
			artMsg.ErrCode = common.ErrACTempPortListenFailed.ErrorCode()
			artMsg.ErrMsg = err.Error()
			return
		}

		// retrieve local port
		uladdr := udpListener.LocalAddr()
		_, locErr = net.ResolveUDPAddr(uladdr.Network(), uladdr.String())
		if locErr != nil {
			log.Error("[HandleAccessControl] resolve local UDPAddr error: %v", locErr)
			err = common.ErrACResolveTempPortFailed
			artMsg.ErrCode = common.ErrACResolveTempPortFailed.ErrorCode()
			artMsg.ErrMsg = err.Error()
			return
		}

		log.Debug("open temporary udp port %s", tlocalAddr.String())
		pickedPort = tlocalAddr.Port

		switch a.config.FilterMode {
		case FilterMode_IPTABLES:
			portHashStr := fmt.Sprintf("%s,udp:%d", netStr, tlocalAddr.Port)
			_, err = a.ipset.Add(ipType, 4, tempOpenTimeSec, portHashStr)
			if err != nil {
				log.Error("[HandleAccessControl] add ipset %s error: %v", portHashStr, err)
				err = common.ErrACIPSetOperationFailed
				artMsg.ErrCode = common.ErrACIPSetOperationFailed.ErrorCode()
				artMsg.ErrMsg = err.Error()
				return
			}
			portHashStr = fmt.Sprintf("%s,udp:%d", netStr1, tlocalAddr.Port)
			_, err = a.ipset.Add(ipType, 4, tempOpenTimeSec, portHashStr)
			if err != nil {
				log.Error("[HandleAccessControl] add ipset %s error: %v", portHashStr, err)
				err = common.ErrACIPSetOperationFailed
				artMsg.ErrCode = common.ErrACIPSetOperationFailed.ErrorCode()
				artMsg.ErrMsg = err.Error()
				return
			}
		case FilterMode_EBPFXDP:
			ebpfHashStr := ebpf.EbpfRuleParams{
				Protocol: "udp",
				DstPort:  tlocalAddr.Port,
			}
			err = ebpf.EbpfRuleAdd(6, ebpfHashStr, tempOpenTimeSec)
			if err != nil {
				log.Error("[EbpfRuleAdd] add ebpf type 6 protocol: %s, dstport :%d, %v", ebpfHashStr.Protocol, ebpfHashStr.DstPort, err)
				return
			}
		default:
			log.Error("[HandleAccessControl] unsupported FilterMode: %d (expected 0=IPTABLES or 1=EBPFXDP)", a.config.FilterMode)
			return
		}
		log.Info("[HandleAccessControl] open temporary udp port on %s", tladdr.String())

		tempEntry := &AccessEntry{
			User:     au,
			SrcAddrs: srcAddrs,
			DstAddrs: dstAddrs,
			OpenTime: tempOpenTimeSec,
		}
		artMsg.PreAccessAction = &common.PreAccessInfo{
			AccessPort:     strconv.Itoa(pickedPort),
			ACPubKey:       a.device.PublicKeyExBase64(),
			ACToken:        a.GenerateAccessToken(tempEntry),
			ACCipherScheme: a.config.DefaultCipherScheme,
		}

		if tcpListener != nil {
			a.wg.Add(1)
			go a.tcpTempAccessHandler(tcpListener, tempOpenTimeSec, dstAddrs, openTimeSec)
		}

		if udpListener != nil {
			a.wg.Add(1)
			go a.udpTempAccessHandler(udpListener, tempOpenTimeSec, dstAddrs, openTimeSec)
		}
	}

	log.Info("[HandleAccessControl] succeed")

	artMsg.ErrCode = common.ErrSuccess.ErrorCode()
	artMsg.OpenTime = uint32(openTimeSec)

	return
}

func (a *UdpAC) tcpTempAccessHandler(listener *net.TCPListener, timeoutSec int, dstAddrs []*common.NetAddress, openTimeSec int) {
	defer a.wg.Done()
	defer listener.Close()

	// accept only the first incoming tcp connection
	startTime := time.Now()
	deadlineTime := startTime.Add(time.Duration(timeoutSec) * time.Second)
	localAddrStr := listener.Addr().String()
	err := listener.SetDeadline(deadlineTime)
	if err != nil {
		log.Error("[tcpTempAccessHandler] temporary port on %s failed to set tcp listen timeout", localAddrStr)
		return
	}
	conn, err := listener.Accept()
	if err != nil {
		log.Error("[tcpTempAccessHandler] temporary port on %s tcp listen timeout", localAddrStr)
		return
	}

	defer conn.Close()
	err = conn.SetDeadline(deadlineTime)
	if err != nil {
		log.Error("[tcpTempAccessHandler] temporary port on %s failed to set tcp conn timeout", localAddrStr)
		return
	}

	remoteAddrStr := conn.RemoteAddr().String()
	pkt := a.device.AllocatePoolPacket()
	defer a.device.ReleasePoolPacket(pkt)

	// monitor stop signals and quit connection earlier
	ctx, ctxCancel := context.WithDeadline(context.Background(), deadlineTime)
	defer ctxCancel()
	go a.tempConnTerminator(conn, ctx)

	// tcp recv common header first
	n, err := conn.Read(pkt.Buf[:core.HeaderCommonSize])
	if err != nil || n < core.HeaderCommonSize {
		log.Error("[tcpTempAccessHandler] failed to receive tcp packet header from remote address %s (%v)", remoteAddrStr, err)
		return
	}

	pkt.Content = pkt.Buf[:n]
	// check type and payload size
	msgType, msgSize := pkt.HeaderTypeAndSize()
	if msgType != core.NHP_ACC {
		log.Error("[tcpTempAccessHandler] message type is not %s, close connection", core.HeaderTypeToString(core.NHP_ACC))
		return
	}

	packetSize := pkt.Header().Size() + msgSize
	remainingSize := packetSize - n
	n, err = conn.Read(pkt.Buf[n:packetSize])
	if err != nil || n < remainingSize {
		log.Error("[tcpTempAccessHandler] failed to receive tcp message body from remote address %s (%v)", remoteAddrStr, err)
		return
	}

	pkt.Content = pkt.Buf[:packetSize]
	//log.Trace("[tcpTempAccessHandler]receive tcp access packet (%s -> %s): %+v", remoteAddrStr, localAddrStr, pkt.Content)
	log.Info("[tcpTempAccessHandler] receive tcp access message (%s -> %s)", remoteAddrStr, localAddrStr)

	pd := &core.PacketData{
		BasePacket:     pkt,
		ConnData:       &core.ConnectionData{},
		InitTime:       time.Now().UnixNano(),
		DecryptedMsgCh: make(chan *core.PacketParserData),
	}

	if !a.IsRunning() {
		log.Error("[tcpTempAccessHandler] PacketData channel closed or being closed, skip decrypting")
		return
	}

	// start message decryption
	a.device.RecvPacketToMsg(pd)

	// waiting for message decryption
	accPpd := <-pd.DecryptedMsgCh
	close(pd.DecryptedMsgCh)

	if accPpd.Error != nil {
		log.Error("[tcpTempAccessHandler] failed to decrypt tcp access message: %v", accPpd.Error)
		return
	}

	accMsg := &common.AgentAccessMsg{}
	err = json.Unmarshal(accPpd.BodyMessage, accMsg)
	if err != nil {
		log.Error("[tcpTempAccessHandler] failed to parse %s message: %v", core.HeaderTypeToString(accPpd.HeaderType), err)
		return
	}

	if a.VerifyAccessToken(accMsg.ACToken) != nil {
		remoteAddr, _ := net.ResolveTCPAddr(conn.RemoteAddr().Network(), conn.RemoteAddr().String())
		srcAddrIp := remoteAddr.IP.String()

		// Detect IP type using proper parsing instead of string matching
		ipType, ipErr := utils.DetectIPType(dstAddrs[0].Ip)
		if ipErr != nil {
			log.Error("[tcpTempAccessHandler] invalid destination IP: %s", dstAddrs[0].Ip)
			return
		}

		for _, dstAddr := range dstAddrs {
			ipHashStr := fmt.Sprintf("%s,%d,%s", srcAddrIp, dstAddr.Port, dstAddr.Ip)
			if dstAddr.Port == 0 {
				ipHashStr = fmt.Sprintf("%s,1-65535,%s", srcAddrIp, dstAddr.Ip)
			}
			switch a.config.FilterMode {
			case FilterMode_IPTABLES:
				_, err = a.ipset.Add(ipType, 1, openTimeSec, ipHashStr)
				if err != nil {
					log.Error("[tcpTempAccessHandler] add ipset %s error: %v", ipHashStr, err)
					return
				}
			case FilterMode_EBPFXDP:
				ebpfHashStr := ebpf.EbpfRuleParams{
					SrcIP: srcAddrIp,
					DstIP: dstAddr.Ip,
				}
				err = ebpf.EbpfRuleAdd(2, ebpfHashStr, openTimeSec)
				if err != nil {
					log.Error("[EbpfRuleAdd] add ebpf src: %s dst: %s, error: %v", ebpfHashStr.SrcIP, ebpfHashStr.DstIP, err)
					return
				}
			default:
				log.Error("[HandleAccessControl] unsupported FilterMode: %d (expected 0=IPTABLES or 1=EBPFXDP)", a.config.FilterMode)
				return
			}
		}
	}
}

func (a *UdpAC) udpTempAccessHandler(conn *net.UDPConn, timeoutSec int, dstAddrs []*common.NetAddress, openTimeSec int) {
	defer a.wg.Done()
	defer conn.Close()
	// listen to accept and handle only one incoming connection
	startTime := time.Now()
	deadlineTime := startTime.Add(time.Duration(timeoutSec) * time.Second)
	localAddrStr := conn.LocalAddr().String()
	err := conn.SetDeadline(deadlineTime)
	if err != nil {
		log.Error("[udpTempAccessHandler] temporary port on %s failed to set udp conn timeout", localAddrStr)
		return
	}

	pkt := a.device.AllocatePoolPacket()
	defer a.device.ReleasePoolPacket(pkt)

	// monitor stop signals and quit connection earlier
	ctx, ctxCancel := context.WithDeadline(context.Background(), deadlineTime)
	defer ctxCancel()
	go a.tempConnTerminator(conn, ctx)

	// udp recv, blocking until packet arrives or deadline reaches
	n, remoteAddr, err := conn.ReadFromUDP(pkt.Buf[:])
	if err != nil || n < core.HeaderCommonSize {
		log.Error("[udpTempAccessHandler] failed to receive udp packet (%v)", err)
		return
	}

	remoteAddrStr := remoteAddr.String()
	pkt.Content = pkt.Buf[:n]

	// check type and payload size
	msgType, msgSize := pkt.HeaderTypeAndSize()
	if msgType != core.NHP_ACC {
		log.Error("[udpTempAccessHandler] message type is not %s, close connection", core.HeaderTypeToString(core.NHP_ACC))
		return
	}

	packetSize := pkt.Header().Size() + msgSize

	if n != packetSize {
		log.Error("[udpTempAccessHandler] udp packet size incorrect from remote address %s", remoteAddrStr)
		return
	}

	log.Trace("receive udp access packet (%s -> %s): %+v", remoteAddrStr, localAddrStr, pkt.Content)
	log.Info("[udpTempAccessHandler] receive udp access message (%s -> %s)", remoteAddrStr, localAddrStr)

	pd := &core.PacketData{
		BasePacket:     pkt,
		ConnData:       &core.ConnectionData{},
		InitTime:       time.Now().UnixNano(),
		DecryptedMsgCh: make(chan *core.PacketParserData),
	}

	if !a.IsRunning() {
		log.Error("[udpTempAccessHandler] PacketData channel closed or being closed, skip decrypting")
		return
	}

	// start packet decryption
	a.device.RecvPacketToMsg(pd)

	// waiting for packet decryption
	accPpd := <-pd.DecryptedMsgCh
	close(pd.DecryptedMsgCh)

	if accPpd.Error != nil {
		log.Error("[udpTempAccessHandler] failed to decrypt udp access message: %v", accPpd.Error)
		return
	}

	accMsg := &common.AgentAccessMsg{}
	err = json.Unmarshal(accPpd.BodyMessage, accMsg)
	if err != nil {
		log.Error("[udpTempAccessHandler] failed to parse %s message: %v", core.HeaderTypeToString(accPpd.HeaderType), err)
		return
	}

	if a.VerifyAccessToken(accMsg.ACToken) != nil {
		srcAddrIp := remoteAddr.IP.String()

		// Detect IP type using proper parsing instead of string matching
		ipType, ipErr := utils.DetectIPType(dstAddrs[0].Ip)
		if ipErr != nil {
			log.Error("[udpTempAccessHandler] invalid destination IP: %s", dstAddrs[0].Ip)
			return
		}

		for _, dstAddr := range dstAddrs {
			if len(dstAddr.Protocol) == 0 || dstAddr.Protocol == "udp" || dstAddr.Protocol == "any" {
				ipHashStr := fmt.Sprintf("%s,udp:%d,%s", srcAddrIp, dstAddr.Port, dstAddr.Ip)
				if dstAddr.Port == 0 {
					ipHashStr = fmt.Sprintf("%s,udp:1-65535,%s", srcAddrIp, dstAddr.Ip)
				}
				switch a.config.FilterMode {
				case FilterMode_IPTABLES:
					_, err = a.ipset.Add(ipType, 1, openTimeSec, ipHashStr)
					if err != nil {
						log.Error("[udpTempAccessHandler] add ipset %s error: %v", ipHashStr, err)
						return
					}
				case FilterMode_EBPFXDP:
					if len(dstAddr.Protocol) == 0 || dstAddr.Protocol == "any" {
						ebpfHashStr := ebpf.EbpfRuleParams{
							SrcIP: srcAddrIp,
							DstIP: dstAddr.Ip,
						}
						err = ebpf.EbpfRuleAdd(2, ebpfHashStr, openTimeSec)
						if err != nil {
							log.Error("[EbpfRuleAdd] add ebpf src: %s dst: %s, error: %v", ebpfHashStr.SrcIP, ebpfHashStr.DstIP, err)
							return
						}
					}
					if dstAddr.Protocol == "udp" {
						ebpfHashStr := ebpf.EbpfRuleParams{
							SrcIP:    srcAddrIp,
							DstIP:    dstAddr.Ip,
							DstPort:  dstAddr.Port,
							Protocol: dstAddr.Protocol,
						}
						err = ebpf.EbpfRuleAdd(1, ebpfHashStr, openTimeSec)

						if err != nil {
							log.Error("[EbpfRuleAdd] add ebpf udp failed src: %s dst: %s, error: %v, protocol: %s, dstport: %d", ebpfHashStr.SrcIP, ebpfHashStr.DstIP, err, ebpfHashStr.Protocol, ebpfHashStr.DstPort)
							return
						}
					}
				default:
					log.Error("[HandleAccessControl] unsupported FilterMode: %d (expected 0=IPTABLES or 1=EBPFXDP)", a.config.FilterMode)
					return
				}
			}
			// for ping
			if dstAddr.Port == 0 && (len(dstAddr.Protocol) == 0 || dstAddr.Protocol == "any") {

				switch a.config.FilterMode {
				case FilterMode_IPTABLES:
					ipHashStr := fmt.Sprintf("%s,icmp:8/0,%s", remoteAddr.IP.String(), dstAddr.Ip)
					a.ipset.Add(ipType, 1, openTimeSec, ipHashStr)
				case FilterMode_EBPFXDP:
					ebpfHashStr := ebpf.EbpfRuleParams{
						SrcIP: remoteAddr.IP.String(),
						DstIP: dstAddr.Ip,
					}
					err = ebpf.EbpfRuleAdd(3, ebpfHashStr, openTimeSec)
					if err != nil {
						log.Error("[EbpfRuleAdd] add ebpf icmp src: %s dst: %s,  error: %v, protocol: %d, dstport :%d, %v", ebpfHashStr.SrcIP, ebpfHashStr.DstIP, err)
						return
					}
				default:
					log.Error("[HandleAccessControl] unsupported FilterMode: %d (expected 0=IPTABLES or 1=EBPFXDP)", a.config.FilterMode)
					return
				}
			}
		}
	}
}

func (a *UdpAC) tempConnTerminator(conn net.Conn, ctx context.Context) {
	select {
	case <-a.signals.stop:
		conn.Close()
		return

	case <-ctx.Done():
		return
	}
}

func incrementIP(ip net.IP) {
	for j := len(ip) - 1; j >= 0; j-- {
		ip[j]++
		if ip[j] > 0 {
			break
		}
	}
}
