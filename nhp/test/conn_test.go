package test

import (
	"fmt"
	"net"
	"sync"
	"testing"
)

func TestUdpConnection(t *testing.T) {
	listenConn, err := net.ListenUDP("udp", &net.UDPAddr{
		//IP:   net.ParseIP("127.0.0.1"),
		IP:   net.IPv4zero,
		Port: 9999,
	})
	if err != nil {
		fmt.Printf("listen error %v\n", err)
		return
	}
	laddr := listenConn.LocalAddr()
	localAddr, err := net.ResolveUDPAddr(laddr.Network(), laddr.String())
	if err != nil {
		fmt.Printf("resolve local UDPAddr error %v\n", err)
		return
	}
	fmt.Printf("listening on address: %s\n", localAddr.String())

	var wait sync.WaitGroup
	wait.Add(2)
	go func() {
		defer wait.Done()

		pkt := make([]byte, 4)
		// udp recv, blocking until packet arrives or conn.Close()
		n, remoteAddr, err := listenConn.ReadFromUDP(pkt[:])
		if err != nil {
			fmt.Printf("Read %d bytes from UDP %s error: (%d) %v\n", n, remoteAddr.String(), err, err)
			return
		}
		fmt.Printf("Read %d bytes from UDP %s: %v\n", n, remoteAddr.String(), pkt[:])
	}()

	// close connection
	/* go func() {
		defer wait.Done()
		listenConn.Close()
	}() */

	// send a larger packet
	/* go func() {
		defer wait.Done()

		remoteAddr := &net.UDPAddr{
			IP:   net.ParseIP("127.0.0.1"),
			Port: 9999,
		}
		cConn, err := net.DialUDP("udp", nil, remoteAddr)
		if err != nil {
			fmt.Printf("could not connect to remote addr %s", remoteAddr.String())
			return
		}
		pkt := []byte{1, 2, 3, 4, 5, 6}
		n, err := cConn.Write(pkt[:])
		if err != nil {
			fmt.Printf("Write %d bytes to UDP %s error: %v\n", n, remoteAddr.String(), err)
			return
		}
		fmt.Printf("Write %d bytes to UDP %s: %v\n", n, remoteAddr.String(), pkt[:])
	}() */

	// send 0 byte packet
	go func() {
		defer wait.Done()

		remoteAddr := &net.UDPAddr{
			IP:   net.ParseIP("127.0.0.1"),
			Port: 9999,
		}
		cConn, err := net.DialUDP("udp", nil, remoteAddr)
		if err != nil {
			fmt.Printf("could not connect to remote addr %s", remoteAddr.String())
			return
		}
		pkt := []byte{1}
		n, err := cConn.Write(pkt[:])
		if err != nil {
			fmt.Printf("Write %d bytes to UDP %s error: %v\n", n, remoteAddr.String(), err)
			return
		}
		fmt.Printf("Write %d bytes to UDP %s: %v\n", n, remoteAddr.String(), pkt[:])
	}()

	wait.Wait()
}
