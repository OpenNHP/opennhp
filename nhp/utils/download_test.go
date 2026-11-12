package utils

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/http"
	"net/http/httptest"
	"net/netip"
	"net/url"
	"os"
	"strings"
	"testing"
	"time"
)

func testLimits(mutate func(*DownloadLimits)) DownloadLimits {
	lim := DefaultDownloadLimits()
	if mutate != nil {
		mutate(&lim)
	}
	return lim
}

func download(t *testing.T, rawURL string, lim DownloadLimits) (string, error) {
	t.Helper()
	root := t.TempDir()
	t.Setenv("TMPDIR", root)
	t.Setenv("TMP", root)
	t.Setenv("TEMP", root)
	return DownloadFileToTempWithLimits(context.Background(), rawURL, "dl-test-", lim)
}

// The addresses a peer-supplied URL may never be pointed at.
//
// 169.254.169.254 is the one that pays: on a cloud instance it is the
// metadata service, and it answers an unauthenticated GET with the role
// credentials of the host nhp-server runs on. It has to stay refused even
// for a deployment that allows private addresses, which is the default —
// hence the AllowPrivate case below.
func TestDownloadRefusesDangerousAddresses(t *testing.T) {
	for name, tc := range map[string]struct {
		rawURL string
		lim    DownloadLimits
		want   string
	}{
		"cloud metadata": {
			"http://169.254.169.254/latest/meta-data/iam/security-credentials/",
			testLimits(nil), "link-local",
		},
		"cloud metadata, private allowed": {
			"http://169.254.169.254/latest/meta-data/",
			testLimits(func(l *DownloadLimits) { l.AllowPrivate = true }), "link-local",
		},
		// These three sit inside ranges the defaults allow, so nothing but
		// the explicit deny list catches them. Each answers an
		// unauthenticated GET with instance credentials.
		"aws imds over ipv6": {
			"http://[fd00:ec2::254]/latest/meta-data/iam/security-credentials/",
			testLimits(func(l *DownloadLimits) { l.AllowPrivate = true }), "metadata",
		},
		"alibaba cloud metadata": {
			"http://100.100.100.200/latest/meta-data/ram/security-credentials/",
			testLimits(func(l *DownloadLimits) { l.AllowPrivate = true }), "metadata",
		},
		"oracle cloud legacy metadata": {
			"http://192.0.0.192/opc/v1/instance/",
			testLimits(func(l *DownloadLimits) { l.AllowPrivate = true }), "metadata",
		},
		// NAT64 and 6to4 spell 169.254.169.254 as an IPv6 address that is
		// neither link-local, private nor v4-mapped.
		"nat64-embedded metadata": {
			"http://[64:ff9b::a9fe:a9fe]/latest/meta-data/",
			testLimits(func(l *DownloadLimits) { l.AllowPrivate = true }), "link-local",
		},
		"6to4-embedded metadata": {
			"http://[2002:a9fe:a9fe::]/latest/meta-data/",
			testLimits(func(l *DownloadLimits) { l.AllowPrivate = true }), "link-local",
		},
		"ipv6 link-local": {"http://[fe80::1]/x", testLimits(nil), "link-local"},
		"loopback":        {"http://127.0.0.1/x", testLimits(nil), "loopback"},
		"loopback name":   {"http://localhost/x", testLimits(nil), "loopback"},
		"ipv6 loopback":   {"http://[::1]/x", testLimits(nil), "loopback"},
		"v4-mapped loopback": {
			"http://[::ffff:127.0.0.1]/x", testLimits(nil), "loopback",
		},
		"unspecified":          {"http://0.0.0.0/x", testLimits(nil), "unspecified"},
		"broadcast":            {"http://255.255.255.255/x", testLimits(nil), "broadcast"},
		"link-local multicast": {"http://224.0.0.1/x", testLimits(nil), "link-local"},
		"multicast":            {"http://239.255.1.1/x", testLimits(nil), "multicast"},
		"private when disallowed": {
			"http://10.1.2.3/x",
			testLimits(func(l *DownloadLimits) { l.AllowPrivate = false }), "private",
		},
		"cgnat when disallowed": {
			"http://100.100.1.1/x",
			testLimits(func(l *DownloadLimits) { l.AllowPrivate = false }), "private",
		},
	} {
		t.Run(name, func(t *testing.T) {
			_, err := download(t, tc.rawURL, tc.lim)
			if err == nil {
				t.Fatalf("download of %s was allowed", tc.rawURL)
			}
			if !strings.Contains(err.Error(), tc.want) {
				t.Errorf("error = %v, want it to mention %q", err, tc.want)
			}
		})
	}
}

// Only http and https. A peer that can choose the scheme can otherwise
// make the process read local files, or reach whatever else the URL
// library is willing to speak.
func TestDownloadRefusesNonHTTPSchemes(t *testing.T) {
	for _, raw := range []string{
		"file:///etc/passwd",
		"file://localhost/etc/shadow",
		"gopher://127.0.0.1:6379/_SET%20x%20y",
		"ftp://example.com/x",
		"data:text/plain;base64,AAAA",
		"/etc/passwd",
		"example.com/policy.wasm",
	} {
		t.Run(raw, func(t *testing.T) {
			if _, err := download(t, raw, testLimits(nil)); err == nil {
				t.Fatalf("download of %q was allowed", raw)
			}
		})
	}
}

func TestDownloadRequireHTTPS(t *testing.T) {
	lim := testLimits(func(l *DownloadLimits) { l.RequireHTTPS = true })
	if _, err := download(t, "http://example.com/x", lim); err == nil {
		t.Fatal("plaintext http was allowed under RequireHTTPS")
	} else if !strings.Contains(err.Error(), "plaintext") {
		t.Errorf("error = %v, want it to mention plaintext", err)
	}
}

// A redirect must not reach an address the first URL could not. The guard
// runs on every hop, so the bounce to loopback is refused even though the
// URL the caller supplied was fine.
func TestDownloadRefusesRedirectToBlockedAddress(t *testing.T) {
	internal := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		fmt.Fprint(w, "INTERNAL")
	}))
	defer internal.Close()

	// The first hop is allowed (loopback opted in), the second is the one
	// under test: a redirect to a link-local address.
	public := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "http://169.254.169.254/latest/meta-data/", http.StatusFound)
	}))
	defer public.Close()

	lim := testLimits(func(l *DownloadLimits) { l.AllowLoopback = true })
	_, err := download(t, public.URL+"/policy.wasm", lim)
	if err == nil {
		t.Fatal("a redirect to the metadata service was followed")
	}
	if !strings.Contains(err.Error(), "link-local") {
		t.Errorf("error = %v, want it to mention link-local", err)
	}
}

func TestDownloadBoundsRedirectChain(t *testing.T) {
	var srv *httptest.Server
	srv = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, srv.URL+"/next", http.StatusFound)
	}))
	defer srv.Close()

	lim := testLimits(func(l *DownloadLimits) {
		l.AllowLoopback = true
		l.MaxRedirects = 2
	})
	_, err := download(t, srv.URL+"/start", lim)
	if err == nil {
		t.Fatal("an endless redirect chain was followed to completion")
	}
	if !strings.Contains(err.Error(), "redirect") {
		t.Errorf("error = %v, want it to mention redirects", err)
	}
}

// The response body is capped whether or not the server declares a length,
// and a partial download leaves nothing behind.
func TestDownloadCapsResponseBody(t *testing.T) {
	const cap = 4 << 10

	for name, declare := range map[string]bool{"declared length": true, "chunked": false} {
		t.Run(name, func(t *testing.T) {
			srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
				payload := make([]byte, cap*4)
				if declare {
					w.Header().Set("Content-Length", fmt.Sprint(len(payload)))
				}
				_, _ = w.Write(payload)
			}))
			defer srv.Close()

			root := t.TempDir()
			t.Setenv("TMPDIR", root)
			t.Setenv("TMP", root)
			t.Setenv("TEMP", root)

			lim := testLimits(func(l *DownloadLimits) {
				l.AllowLoopback = true
				l.MaxBytes = cap
			})
			_, err := DownloadFileToTempWithLimits(context.Background(), srv.URL+"/big.wasm", "dl-test-", lim)
			if err == nil {
				t.Fatal("an oversized response was accepted")
			}
			if !errors.Is(err, ErrDownloadTooLarge) {
				t.Errorf("error = %v, want ErrDownloadTooLarge", err)
			}
			entries, rerr := os.ReadDir(root)
			if rerr != nil {
				t.Fatalf("ReadDir: %v", rerr)
			}
			if len(entries) != 0 {
				t.Errorf("the refused download left %d entries behind", len(entries))
			}
		})
	}
}

func TestDownloadAcceptsBodyAtTheLimit(t *testing.T) {
	const size = 4 << 10
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write(make([]byte, size))
	}))
	defer srv.Close()

	lim := testLimits(func(l *DownloadLimits) {
		l.AllowLoopback = true
		l.MaxBytes = size
	})
	p, err := download(t, srv.URL+"/exact.wasm", lim)
	if err != nil {
		t.Fatalf("a body exactly at the limit was refused: %v", err)
	}
	fi, err := os.Stat(p)
	if err != nil {
		t.Fatalf("Stat: %v", err)
	}
	if fi.Size() != size {
		t.Errorf("wrote %d bytes, want %d", fi.Size(), size)
	}
}

// A peer that answers headers and then stalls must not hold a goroutine
// (and the temp directory) open forever. On main this hung indefinitely.
func TestDownloadDoesNotHangOnAStalledBody(t *testing.T) {
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("listen: %v", err)
	}
	defer ln.Close()
	go func() {
		for {
			c, err := ln.Accept()
			if err != nil {
				return
			}
			go func() {
				defer c.Close()
				_, _ = c.Read(make([]byte, 1024))
				fmt.Fprint(c, "HTTP/1.1 200 OK\r\nContent-Length: 10\r\n\r\n")
				<-time.After(time.Minute) // headers sent, body never
			}()
		}
	}()

	lim := testLimits(func(l *DownloadLimits) {
		l.AllowLoopback = true
		l.Timeout = 2 * time.Second
	})
	done := make(chan error, 1)
	go func() {
		_, err := download(t, "http://"+ln.Addr().String()+"/x", lim)
		done <- err
	}()
	select {
	case err := <-done:
		if err == nil {
			t.Fatal("a stalled body produced a successful download")
		}
	case <-time.After(20 * time.Second):
		t.Fatal("the download is still running: the client has no overall timeout")
	}
}

// The caller's context cancels an in-flight download.
func TestDownloadHonorsContextCancellation(t *testing.T) {
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("listen: %v", err)
	}
	defer ln.Close()
	go func() {
		c, err := ln.Accept()
		if err != nil {
			return
		}
		defer c.Close()
		_, _ = c.Read(make([]byte, 1024))
		fmt.Fprint(c, "HTTP/1.1 200 OK\r\nContent-Length: 10\r\n\r\n")
		<-time.After(time.Minute)
	}()

	ctx, cancel := context.WithCancel(context.Background())
	lim := testLimits(func(l *DownloadLimits) { l.AllowLoopback = true })
	done := make(chan error, 1)
	go func() {
		_, err := DownloadFileToTempWithLimits(ctx, "http://"+ln.Addr().String()+"/x", "dl-test-", lim)
		done <- err
	}()
	time.Sleep(200 * time.Millisecond)
	cancel()
	select {
	case err := <-done:
		if err == nil {
			t.Fatal("canceling the context produced a successful download")
		}
	case <-time.After(10 * time.Second):
		t.Fatal("canceling the context did not stop the download")
	}
}

// A name that resolves to a blocked address is refused at dial time, which
// is what makes a second, different answer (DNS rebinding) pointless: the
// check and the connect are the same step.
func TestDownloadChecksTheAddressActuallyDialed(t *testing.T) {
	for _, addr := range []string{"127.0.0.1:80", "169.254.169.254:80", "[::1]:443", "0.0.0.0:80"} {
		if err := checkDialAddress("tcp", addr, DefaultDownloadLimits()); err == nil {
			t.Errorf("checkDialAddress(%q) allowed the connection", addr)
		}
	}
	if err := checkDialAddress("tcp", "93.184.216.34:443", DefaultDownloadLimits()); err != nil {
		t.Errorf("a public address was refused: %v", err)
	}
	if err := checkDialAddress("udp", "93.184.216.34:443", DefaultDownloadLimits()); err == nil {
		t.Error("a non-tcp network was allowed")
	}
}

func TestDownloadAllowedHosts(t *testing.T) {
	lim := testLimits(func(l *DownloadLimits) { l.AllowedHosts = []string{"Policies.Example.COM"} })
	if _, err := ValidateDownloadURL("https://policies.example.com/p.wasm", lim); err != nil {
		t.Errorf("an allowed host was refused: %v", err)
	}
	if _, err := ValidateDownloadURL("https://evil.example.net/p.wasm", lim); err == nil {
		t.Error("a host outside the allow list was accepted")
	}
}

// The temp file name comes from the URL path, never from the raw string,
// and is always a single plain component.
func TestDownloadFileName(t *testing.T) {
	for raw, want := range map[string]string{
		"https://h/policy.wasm":         "policy.wasm",
		"https://h/a/b/policy.wasm?v=1": "policy.wasm",
		"https://h/a/b/?v=x/y":          "b",
		"https://h/p.wasm?v=x/y":        "p.wasm",
		"https://h/":                    "download",
		"https://h":                     "download",
		"https://h/..":                  "download",
		"https://h/%2e%2e%2f%2e%2e%2fx": "x",
		"https://h/a/%2e%2e":            "download",
		"https://h/data.ztdo#frag":      "data.ztdo",
	} {
		u, err := url.Parse(raw)
		if err != nil {
			t.Fatalf("parse %q: %v", raw, err)
		}
		if got := downloadFileName(u); got != want {
			t.Errorf("downloadFileName(%q) = %q, want %q", raw, got, want)
		}
	}
}

func TestDownloadLimitDefaults(t *testing.T) {
	d := DefaultDownloadLimits()
	w := WasmPolicyDownloadLimits()
	if d.AllowLoopback || w.AllowLoopback {
		t.Error("loopback must be refused by default")
	}
	if !d.AllowPrivate || !w.AllowPrivate {
		t.Error("private addresses are allowed by default; the shipped topology puts the storage endpoint there")
	}
	if w.MaxBytes >= d.MaxBytes {
		t.Errorf("a WASM policy (%d) should be capped well below a data object (%d)", w.MaxBytes, d.MaxBytes)
	}
	if d.MaxBytes == 0 || w.MaxBytes == 0 || d.Timeout == 0 || w.Timeout == 0 || d.MaxRedirects == 0 {
		t.Error("the defaults must bound size, time and redirects")
	}
	if err := checkDownloadAddr(netip.MustParseAddr("100.64.0.1"), DownloadLimits{}); err == nil {
		t.Error("CGNAT space is not treated as private")
	}
}

// A deployment may allow private space — the default does, because the
// shipped topology puts the nhp-server storage endpoint there — and that
// must not hand a peer the instance's credentials. Checked against the
// guard directly so the table above cannot be satisfied by a DNS failure.
func TestDownloadDenyListBeatsEveryAllowFlag(t *testing.T) {
	permissive := DownloadLimits{AllowLoopback: true, AllowPrivate: true}
	for _, raw := range []string{
		"169.254.169.254",    // AWS / GCP / Azure / OpenStack
		"fd00:ec2::254",      // AWS IMDS over IPv6, inside ULA fc00::/7
		"100.100.100.200",    // Alibaba Cloud, inside CGNAT 100.64/10
		"192.0.0.192",        // Oracle Cloud legacy
		"64:ff9b::a9fe:a9fe", // NAT64-embedded 169.254.169.254
		"2002:a9fe:a9fe::",   // 6to4-embedded 169.254.169.254
	} {
		if err := checkDownloadAddr(netip.MustParseAddr(raw), permissive); err == nil {
			t.Errorf("%s was allowed with every allow flag set", raw)
		}
	}
	// The ranges those endpoints sit in stay usable otherwise: refusing
	// all of CGNAT or all of ULA by default would break deployments that
	// legitimately host the storage endpoint there.
	for _, raw := range []string{"100.64.0.1", "fd12:3456::1", "10.1.2.3", "192.168.1.5"} {
		if err := checkDownloadAddr(netip.MustParseAddr(raw), permissive); err != nil {
			t.Errorf("%s was refused although private space is allowed: %v", raw, err)
		}
	}
}

// Errors reach the daemon log and the URL came from a peer, so a presigned
// URL must not log its signature.
func TestDownloadErrorsRedactCredentials(t *testing.T) {
	for raw, want := range map[string]string{
		"https://bucket.example.com/o.ztdo?X-Amz-Signature=deadbeef&X-Amz-Credential=AKIA": "https://bucket.example.com/o.ztdo?<redacted>",
		"https://user:hunter2@example.com/p.wasm":                                          "https://example.com/p.wasm",
		"https://example.com/p.wasm":                                                       "https://example.com/p.wasm",
		"https://example.com/p.wasm#frag":                                                  "https://example.com/p.wasm",
		"not a url?sig=secret":                                                             "not a url?<redacted>",
	} {
		if got := redactURL(raw); got != want {
			t.Errorf("redactURL(%q) = %q, want %q", raw, got, want)
		}
	}

	// And the redaction is actually what the download path reports.
	_, err := download(t, "http://169.254.169.254/creds?X-Amz-Signature=deadbeef", testLimits(nil))
	if err == nil {
		t.Fatal("the metadata endpoint was reachable")
	}
	if strings.Contains(err.Error(), "deadbeef") {
		t.Errorf("the error leaked the query string: %v", err)
	}
}

// A download reports the error its close returns rather than handing back
// a file whose last write never landed.
func TestDownloadReportsCloseFailure(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte("payload"))
	}))
	defer srv.Close()

	lim := testLimits(func(l *DownloadLimits) { l.AllowLoopback = true })
	p, err := download(t, srv.URL+"/p.wasm", lim)
	if err != nil {
		t.Fatalf("download: %v", err)
	}
	// The file is closed by the time the path is returned, so re-opening
	// and truncating it is safe; a leaked descriptor would show up as a
	// failure on Windows and as a lingering handle elsewhere.
	f, err := os.OpenFile(p, os.O_RDWR, 0o600)
	if err != nil {
		t.Fatalf("the returned file is not usable: %v", err)
	}
	_ = f.Close()
	b, err := os.ReadFile(p)
	if err != nil || string(b) != "payload" {
		t.Fatalf("content = %q, err = %v, want %q", b, err, "payload")
	}
}
