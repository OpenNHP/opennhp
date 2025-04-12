package utils

import (
	"net"
	"strings"
)

func GetLocalOutboundAddress() net.IP {
	con, err := net.Dial("udp", "8.8.8.8:80")
	if err != nil {
		return nil
	}
	defer con.Close()

	addr := con.LocalAddr().(*net.UDPAddr)

	return addr.IP
}

func GetMacAddress(ip string) string {
	interfaces, err := net.Interfaces()
	if err != nil {
		return ""
	}

	for _, intf := range interfaces {
		if addrs, err := intf.Addrs(); err == nil {
			for _, addr := range addrs {
				if strings.Contains(addr.String(), ip) {
					mac := intf.HardwareAddr.String()
					return mac
				}
			}
		}
	}

	return ""
}
