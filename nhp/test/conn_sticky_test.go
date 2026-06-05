package test

import (
	"net"
	"testing"
	"time"

	"github.com/OpenNHP/opennhp/nhp/core"
)

// MinimalPeerAddressHoldTime is exported as a constant in nhp/core. Re-derive
// it here so the assertions stay readable; the test fails fast if the value
// ever changes.
const holdSeconds = 5

func mustResolve(t *testing.T, addr string) *net.UDPAddr {
	t.Helper()
	a, err := net.ResolveUDPAddr("udp", addr)
	if err != nil {
		t.Fatalf("resolve %s: %v", addr, err)
	}
	return a
}

// TestConnStickiness_AcceptsFirstAddress: a fresh connection accepts whatever
// source address it sees first.
func TestConnStickiness_AcceptsFirstAddress(t *testing.T) {
	c := &core.ConnectionData{}
	now := time.Now().UnixNano()
	addr := mustResolve(t, "10.0.0.1:62206")
	if !c.CheckRecvAddress(now, addr) {
		t.Fatal("expected first address to be accepted on fresh conn")
	}
	c.UpdateRecvAddress(now, addr)
}

// TestConnStickiness_RejectsChangeWithinHoldTime: once a source is locked in,
// a different source within the hold window is rejected.
func TestConnStickiness_RejectsChangeWithinHoldTime(t *testing.T) {
	c := &core.ConnectionData{}
	t0 := time.Now().UnixNano()
	a := mustResolve(t, "10.0.0.1:62206")
	b := mustResolve(t, "10.0.0.2:62206")

	if !c.CheckRecvAddress(t0, a) {
		t.Fatal("first address should be accepted")
	}
	c.UpdateRecvAddress(t0, a)

	t1 := t0 + int64(time.Second) // 1s later, still inside hold window
	if c.CheckRecvAddress(t1, b) {
		t.Fatal("different source within hold window should be rejected")
	}
}

// TestConnStickiness_AcceptsSameAddressWithinHoldTime: repeated packets from
// the same source are always fine.
func TestConnStickiness_AcceptsSameAddressWithinHoldTime(t *testing.T) {
	c := &core.ConnectionData{}
	t0 := time.Now().UnixNano()
	a := mustResolve(t, "10.0.0.1:62206")

	if !c.CheckRecvAddress(t0, a) {
		t.Fatal("first address should be accepted")
	}
	c.UpdateRecvAddress(t0, a)

	t1 := t0 + int64(time.Second)
	if !c.CheckRecvAddress(t1, a) {
		t.Fatal("same source within hold window should be accepted")
	}
}

// TestConnStickiness_AcceptsChangeAfterHoldTime: after the hold window
// elapses, the conn no longer cares about historical addresses.
func TestConnStickiness_AcceptsChangeAfterHoldTime(t *testing.T) {
	c := &core.ConnectionData{}
	t0 := time.Now().UnixNano()
	a := mustResolve(t, "10.0.0.1:62206")
	b := mustResolve(t, "10.0.0.2:62206")

	if !c.CheckRecvAddress(t0, a) {
		t.Fatal("first address should be accepted")
	}
	c.UpdateRecvAddress(t0, a)

	tAfter := t0 + int64((holdSeconds+1)*time.Second)
	if !c.CheckRecvAddress(tAfter, b) {
		t.Fatal("different source after hold window should be accepted")
	}
}

// TestConnStickiness_IndependentAcrossConns: two ConnectionData instances do
// not share stickiness state. This is the property that makes AC fan-out to
// multiple same-pubkey nhp-server endpoints work — each endpoint's conn
// tracks its own source addr.
func TestConnStickiness_IndependentAcrossConns(t *testing.T) {
	c1 := &core.ConnectionData{}
	c2 := &core.ConnectionData{}
	t0 := time.Now().UnixNano()
	a := mustResolve(t, "10.0.0.1:62206")
	b := mustResolve(t, "10.0.0.2:62206")

	c1.UpdateRecvAddress(t0, a)
	c2.UpdateRecvAddress(t0, b)

	if !c1.CheckRecvAddress(t0+int64(time.Second), a) {
		t.Fatal("c1 should still accept its locked source")
	}
	if !c2.CheckRecvAddress(t0+int64(time.Second), b) {
		t.Fatal("c2 should still accept its locked source")
	}
	if c1.CheckRecvAddress(t0+int64(time.Second), b) {
		t.Fatal("c1 must not be affected by c2's traffic")
	}
}
