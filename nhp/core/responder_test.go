package core

import (
	"testing"
	"time"
)

func TestShouldCheckRecvAttack(t *testing.T) {
	tests := []struct {
		name                          string
		deviceType, peerType, msgType int
		want                          bool
	}{
		{"AOP uses replay gate", NHP_AC, NHP_SERVER, NHP_AOP, true},
		{"ART remains exempt", NHP_SERVER, NHP_AC, NHP_ART, false},
		{"default", NHP_SERVER, NHP_AGENT, NHP_KNK, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := shouldCheckRecvAttack(tt.deviceType, tt.peerType, tt.msgType); got != tt.want {
				t.Fatalf("shouldCheckRecvAttack() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestShouldCheckFlood(t *testing.T) {
	tests := []struct {
		name                          string
		deviceType, peerType, msgType int
		want                          bool
	}{
		{"AOP burst exempt", NHP_AC, NHP_SERVER, NHP_AOP, false},
		{"ART remains exempt", NHP_SERVER, NHP_AC, NHP_ART, false},
		{"default", NHP_SERVER, NHP_AGENT, NHP_KNK, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := shouldCheckFlood(tt.deviceType, tt.peerType, tt.msgType); got != tt.want {
				t.Fatalf("shouldCheckFlood() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestShouldEscalateReplay(t *testing.T) {
	tests := []struct {
		name                          string
		deviceType, peerType, msgType int
		want                          bool
	}{
		{"AOP is drop-only", NHP_AC, NHP_SERVER, NHP_AOP, false},
		{"ART unchanged", NHP_SERVER, NHP_AC, NHP_ART, true},
		{"AOP wrong scope", NHP_SERVER, NHP_SERVER, NHP_AOP, true},
		{"default", NHP_SERVER, NHP_AGENT, NHP_KNK, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := shouldEscalateReplay(tt.deviceType, tt.peerType, tt.msgType); got != tt.want {
				t.Fatalf("shouldEscalateReplay() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestShouldEscalateStale(t *testing.T) {
	tests := []struct {
		name                          string
		deviceType, peerType, msgType int
		want                          bool
	}{
		{"AOP is drop-only", NHP_AC, NHP_SERVER, NHP_AOP, false},
		{"ART unchanged", NHP_SERVER, NHP_AC, NHP_ART, true},
		{"AOP wrong scope", NHP_SERVER, NHP_SERVER, NHP_AOP, true},
		{"default", NHP_SERVER, NHP_AGENT, NHP_KNK, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := shouldEscalateStale(tt.deviceType, tt.peerType, tt.msgType); got != tt.want {
				t.Fatalf("shouldEscalateStale() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestRecvStalenessFloor(t *testing.T) {
	aop := recvStalenessFloor(NHP_AC, NHP_SERVER, NHP_AOP)
	if want := AOPRecvStalenessFloorSeconds * int64(time.Second); aop != want {
		t.Fatalf("AOP floor = %d, want %d", aop, want)
	}
	defaultFloor := recvStalenessFloor(NHP_SERVER, NHP_AGENT, NHP_KNK)
	if want := DefaultRecvStalenessFloorSeconds * int64(time.Second); defaultFloor != want {
		t.Fatalf("default floor = %d, want %d", defaultFloor, want)
	}
	if aop >= defaultFloor {
		t.Fatalf("AOP floor (%d) must be tighter than default (%d)", aop, defaultFloor)
	}
}
