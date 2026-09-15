package main

import (
	"testing"
	"time"
)

func TestIPLimiterAllow(t *testing.T) {
	l := newIPLimiter(3, time.Minute)
	for i := 1; i <= 3; i++ {
		if !l.Allow("1.2.3.4") {
			t.Fatalf("request %d within limit rejected", i)
		}
	}
	if l.Allow("1.2.3.4") {
		t.Fatal("4th request over the limit was allowed")
	}
	// A different IP has its own bucket.
	if !l.Allow("5.6.7.8") {
		t.Fatal("different IP rejected by another IP's bucket")
	}
}

func TestIPLimiterWindowReset(t *testing.T) {
	// Tiny window so the test can observe a reset without real sleeping.
	l := newIPLimiter(1, 20*time.Millisecond)
	if !l.Allow("9.9.9.9") {
		t.Fatal("first request rejected")
	}
	if l.Allow("9.9.9.9") {
		t.Fatal("second request within window allowed")
	}
	time.Sleep(25 * time.Millisecond)
	if !l.Allow("9.9.9.9") {
		t.Fatal("request after window expiry rejected")
	}
}

func TestItoa(t *testing.T) {
	cases := map[int]string{0: "0", 1: "1", 9: "9", 10: "10", 600: "600"}
	for in, want := range cases {
		if got := itoa(in); got != want {
			t.Errorf("itoa(%d) = %q, want %q", in, got, want)
		}
	}
}
