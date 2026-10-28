package core

import (
	"errors"
	"net"
	"strings"
	"testing"
)

// These tests pin the contract fixed in OpenNHP/opennhp#1681: a DNS failure
// for a Host-only peer must be reported as a resolution problem that names
// the host, never as "IP cannot be parsed", while a literal Ip keeps its
// existing behavior.

// unresolvableHost uses the reserved .invalid TLD (RFC 2606), which is
// guaranteed never to resolve, so the lookup fails deterministically with
// or without network access.
const unresolvableHost = "nonexistent-peer.invalid"

func TestSendAddrErr_LiteralIpUnchanged(t *testing.T) {
	p := &UdpPeer{Ip: "203.0.113.7", Port: 62206}

	addr, err := p.SendAddrErr()
	if err != nil {
		t.Fatalf("literal Ip must not error, got: %v", err)
	}
	udp, ok := addr.(*net.UDPAddr)
	if !ok || udp.IP.String() != "203.0.113.7" || udp.Port != 62206 {
		t.Fatalf("unexpected addr %v", addr)
	}
	if got := p.SendAddr(); got == nil || got.String() != addr.String() {
		t.Fatalf("SendAddr must agree with SendAddrErr, got %v", got)
	}
}

func TestSendAddrErr_InvalidLiteralIp(t *testing.T) {
	p := &UdpPeer{Ip: "not-an-ip", Port: 62206}

	addr, err := p.SendAddrErr()
	if addr != nil {
		t.Fatalf("expected nil addr, got %v", addr)
	}
	if !errors.Is(err, ErrPeerInvalidIp) {
		t.Fatalf("expected ErrPeerInvalidIp, got: %v", err)
	}
	if errors.Is(err, ErrPeerHostResolve) {
		t.Fatalf("an unparseable literal Ip must not be reported as a DNS failure: %v", err)
	}
	if p.SendAddr() != nil {
		t.Fatal("SendAddr must stay nil for an unparseable Ip")
	}
}

func TestSendAddrErr_HostResolutionFailure(t *testing.T) {
	p := &UdpPeer{Hostname: unresolvableHost, Port: 62206}

	addr, err := p.SendAddrErr()
	if addr != nil {
		t.Fatalf("expected nil addr for unresolvable host, got %v", addr)
	}
	if !errors.Is(err, ErrPeerHostResolve) {
		t.Fatalf("expected ErrPeerHostResolve, got: %v", err)
	}
	if errors.Is(err, ErrPeerInvalidIp) {
		t.Fatalf("a DNS failure must not be reported as an IP-parsing failure: %v", err)
	}
	if !strings.Contains(err.Error(), unresolvableHost) {
		t.Fatalf("error must name the host so the user can act on it, got: %v", err)
	}
	if strings.Contains(strings.ToLower(err.Error()), "cannot be parsed") {
		t.Fatalf("DNS failure message must not mention IP parsing: %v", err)
	}
	// Interface-compatible path keeps returning nil.
	if p.SendAddr() != nil {
		t.Fatal("SendAddr must stay nil when the host cannot be resolved")
	}
}

func TestSendAddrErr_HostFailureFallsBackToStaticIp(t *testing.T) {
	// A Host that fails to resolve but has a valid static Ip configured
	// must keep working via the fallback, exactly as before.
	p := &UdpPeer{Hostname: unresolvableHost, Ip: "203.0.113.9", Port: 62206}

	addr, err := p.SendAddrErr()
	if err != nil {
		t.Fatalf("static Ip fallback must not error, got: %v", err)
	}
	if udp, ok := addr.(*net.UDPAddr); !ok || udp.IP.String() != "203.0.113.9" {
		t.Fatalf("expected fallback to static Ip, got %v", addr)
	}
}
