package utils

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/netip"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"strings"
	"syscall"
	"time"
)

// Bounds for a download whose URL arrived in a peer message.
//
// Two different shapes of file come down this path and they do not deserve
// the same allowance. A smart-policy WASM module is code the process is
// about to load into wazero: megabytes, seconds. A ZTDO data object is the
// payload itself: it can legitimately be large and slow.
const (
	// DefaultDownloadMaxBytes bounds a data-object download.
	DefaultDownloadMaxBytes int64 = 1 << 30 // 1 GiB
	// DefaultDownloadTimeout bounds one data-object download end to end.
	DefaultDownloadTimeout = 10 * time.Minute

	// WasmPolicyMaxBytes bounds a smart-policy module.
	WasmPolicyMaxBytes int64 = 16 << 20 // 16 MiB
	// WasmPolicyTimeout bounds one smart-policy fetch end to end.
	WasmPolicyTimeout = 30 * time.Second

	// defaultMaxRedirects is how many hops a download may be bounced
	// through. Each one is re-validated, so this only bounds the work.
	defaultMaxRedirects = 3

	dialTimeout           = 10 * time.Second
	tlsHandshakeTimeout   = 10 * time.Second
	responseHeaderTimeout = 30 * time.Second
)

// ErrDownloadTooLarge is returned when a response exceeds MaxBytes.
var ErrDownloadTooLarge = errors.New("download exceeds the allowed size")

// DownloadLimits bounds one download of a peer-supplied URL.
//
// The URLs these limits apply to are not configuration. `SmartPolicy.Policy`
// and `DAGMsg.AccessUrl` arrive in DRG/DSA/DAG messages, so the address they
// name is chosen by whoever sent the message — a data broker, or anything
// that can impersonate one. The process that fetches them is nhp-server or
// nhp-agent, which sit inside the network the product exists to protect, so
// an unchecked fetch is a request generator pointed at the inside: CWE-918
// on the way out, CWE-400 on the way back.
type DownloadLimits struct {
	// MaxBytes caps the response body. 0 means no cap and should only be
	// used where the input is genuinely trusted.
	MaxBytes int64
	// Timeout bounds the whole request, headers and body together. The
	// zero value means no overall bound; the per-phase transport timeouts
	// still apply, so a peer cannot hold a goroutine open by stalling
	// before it answers.
	Timeout time.Duration
	// MaxRedirects bounds redirect hops. Every hop is validated again, so
	// a redirect cannot be used to reach an address the first URL could
	// not.
	MaxRedirects int

	// AllowLoopback permits 127.0.0.0/8 and ::1. Off by default: a
	// peer-named URL has no business reaching a service that bound to
	// loopback precisely because it did not want callers.
	AllowLoopback bool
	// AllowPrivate permits RFC1918, CGNAT (100.64/10) and IPv6 ULA. It
	// does not reach the metadata endpoints that live inside those ranges
	// (AWS's fd00:ec2::254, Alibaba's 100.100.100.200) — see
	// deniedMetadataPrefixes, which is checked first and has no opt-out.
	// ON by default, because the shipped topology puts them there: an
	// nhp-db uploads the policy to the nhp-server's own /storage endpoint
	// and registers that URL, and in an on-prem or docker deployment the
	// nhp-server's address is private. Turning this off is the tighter
	// posture for a deployment whose policy and data hosts are public.
	AllowPrivate bool
	// RequireHTTPS rejects plaintext http. Off by default because the
	// nhp-db's own storage upload falls back to http (endpoints/db/utils.go),
	// so requiring TLS here would refuse URLs the product itself mints.
	RequireHTTPS bool
	// AllowedHosts, when non-empty, is an exact host allowlist (lower-case
	// host name or IP literal, no port). It is checked on every hop.
	AllowedHosts []string
}

// DefaultDownloadLimits bounds a data-object download.
func DefaultDownloadLimits() DownloadLimits {
	return DownloadLimits{
		MaxBytes:     DefaultDownloadMaxBytes,
		Timeout:      DefaultDownloadTimeout,
		MaxRedirects: defaultMaxRedirects,
		AllowPrivate: true,
	}
}

// WasmPolicyDownloadLimits bounds a smart-policy module download.
func WasmPolicyDownloadLimits() DownloadLimits {
	return DownloadLimits{
		MaxBytes:     WasmPolicyMaxBytes,
		Timeout:      WasmPolicyTimeout,
		MaxRedirects: defaultMaxRedirects,
		AllowPrivate: true,
	}
}

// Metadata endpoints that must never be reachable, whatever else a
// deployment allows.
//
// 169.254.169.254 is the famous one and link-local already covers it, but
// it is not the only one, and the others sit inside ranges a deployment has
// good reason to allow:
//
//	fd00:ec2::254     AWS IMDS over IPv6 (Nitro, HttpProtocolIpv6=enabled).
//	                  Inside ULA fc00::/7, so IsPrivate() reports true.
//	100.100.100.200   Alibaba Cloud metadata. Inside CGNAT 100.64/10.
//	192.0.0.0/24      IETF protocol assignments, which is where Oracle
//	                  Cloud's legacy 192.0.0.192 endpoint lives. No host a
//	                  policy or data object could legitimately be served
//	                  from is in this block.
//
// Each of these answers an unauthenticated GET with instance credentials,
// so they are checked ahead of the AllowLoopback/AllowPrivate branches
// rather than inside them.
var deniedMetadataPrefixes = []netip.Prefix{
	netip.MustParsePrefix("fd00:ec2::254/128"),
	netip.MustParsePrefix("100.100.100.200/32"),
	netip.MustParsePrefix("192.0.0.0/24"),
}

// NAT64 (RFC 6052 well-known prefix) and 6to4 embed an IPv4 address in an
// IPv6 one. On a network that runs either, 64:ff9b::a9fe:a9fe and
// 2002:a9fe:a9fe:: both reach 169.254.169.254 — and neither is link-local,
// private or v4-mapped, so nothing else here would catch them.
//
// Only these two fixed prefixes are handled. An operator-chosen NAT64
// prefix (RFC 6052 allows /32 through /96) cannot be enumerated from here;
// a deployment on one should set AllowPrivate=false and an AllowedHosts
// list.
var (
	nat64WellKnownPrefix = netip.MustParsePrefix("64:ff9b::/96")
	sixToFourPrefix      = netip.MustParsePrefix("2002::/16")
)

// embeddedIPv4 extracts the IPv4 address carried inside a NAT64 or 6to4
// address, so it can be judged as the IPv4 address it actually reaches.
func embeddedIPv4(ip netip.Addr) (netip.Addr, bool) {
	if !ip.Is6() || ip.Is4In6() {
		return netip.Addr{}, false
	}
	b := ip.As16()
	switch {
	case nat64WellKnownPrefix.Contains(ip):
		return netip.AddrFrom4([4]byte{b[12], b[13], b[14], b[15]}), true
	case sixToFourPrefix.Contains(ip):
		return netip.AddrFrom4([4]byte{b[2], b[3], b[4], b[5]}), true
	}
	return netip.Addr{}, false
}

// checkDownloadAddr decides whether one resolved address may be dialed.
//
// The deny list and the link-local branch are refused whatever the limits
// say; everything below them is a deployment choice. That split is the
// point: a deployment that allows private space (the default, because the
// shipped topology puts the storage endpoint there) must not thereby get
// its instance credentials read out by a peer-supplied URL.
func checkDownloadAddr(addr netip.Addr, lim DownloadLimits) error {
	if !addr.IsValid() {
		return errors.New("refusing to dial an invalid address")
	}
	ip := addr.Unmap() // ::ffff:127.0.0.1 must be judged as 127.0.0.1

	for _, p := range deniedMetadataPrefixes {
		if p.Contains(ip) {
			return fmt.Errorf("refusing to dial %s: a cloud instance metadata endpoint", ip)
		}
	}

	// Judge what the address actually reaches, not how it is spelled.
	if v4, ok := embeddedIPv4(ip); ok {
		if err := checkDownloadAddr(v4, lim); err != nil {
			return fmt.Errorf("refusing to dial %s: it embeds %s: %w", ip, v4, err)
		}
	}

	switch {
	case ip.IsUnspecified():
		return fmt.Errorf("refusing to dial the unspecified address %s", ip)
	case ip.IsLinkLocalUnicast(), ip.IsLinkLocalMulticast():
		return fmt.Errorf("refusing to dial link-local address %s", ip)
	case ip.IsMulticast(), ip.IsInterfaceLocalMulticast():
		return fmt.Errorf("refusing to dial multicast address %s", ip)
	case ip.Is4() && ip == netip.AddrFrom4([4]byte{255, 255, 255, 255}):
		return fmt.Errorf("refusing to dial the broadcast address %s", ip)
	case ip.IsLoopback():
		if !lim.AllowLoopback {
			return fmt.Errorf("refusing to dial loopback address %s", ip)
		}
	case ip.IsPrivate(), isCGNAT(ip):
		if !lim.AllowPrivate {
			return fmt.Errorf("refusing to dial private address %s", ip)
		}
	}
	return nil
}

// isCGNAT reports 100.64.0.0/10, which netip.Addr.IsPrivate does not cover
// and which carriers and several cloud providers use as internal space.
func isCGNAT(ip netip.Addr) bool {
	return ip.Is4() && netip.PrefixFrom(netip.AddrFrom4([4]byte{100, 64, 0, 0}), 10).Contains(ip)
}

// redactURL renders a URL for an error message with the parts that can
// carry a credential removed.
//
// These errors reach the daemon log, and the URL came from a peer: an
// AccessUrl may well be presigned, and a presigned URL carries its
// signature in the query string (AWS X-Amz-Signature, an Azure SAS token).
// Userinfo goes for the same reason.
func redactURL(raw string) string {
	raw = strings.TrimSpace(raw)
	u, err := url.Parse(raw)
	if err != nil || u.Scheme == "" {
		// Unparsable or scheme-less: keep only what cannot hold a secret.
		if i := strings.IndexAny(raw, "?#"); i >= 0 {
			return raw[:i] + "?<redacted>"
		}
		return raw
	}
	redacted := *u
	redacted.User = nil
	redacted.Fragment = ""
	if redacted.RawQuery != "" {
		redacted.RawQuery = ""
		return redacted.String() + "?<redacted>"
	}
	return redacted.String()
}

// ValidateDownloadURL parses and checks a URL against lim, without
// contacting anything. It is the fail-fast half; the authoritative check is
// checkDownloadAddr running on the address actually dialed.
func ValidateDownloadURL(rawURL string, lim DownloadLimits) (*url.URL, error) {
	u, err := url.Parse(strings.TrimSpace(rawURL))
	if err != nil {
		return nil, fmt.Errorf("refusing to download %q: %w", redactURL(rawURL), err)
	}

	switch strings.ToLower(u.Scheme) {
	case "https":
	case "http":
		if lim.RequireHTTPS {
			return nil, fmt.Errorf("refusing to download %q over plaintext http", redactURL(rawURL))
		}
	default:
		// file://, gopher://, ftp:// and friends. A peer that can name the
		// scheme can otherwise read local files through this function.
		return nil, fmt.Errorf("refusing to download over scheme %q: only http and https are allowed", u.Scheme)
	}

	host := u.Hostname()
	if host == "" {
		return nil, fmt.Errorf("refusing to download %q: no host", redactURL(rawURL))
	}
	if len(lim.AllowedHosts) > 0 && !hostAllowed(host, lim.AllowedHosts) {
		return nil, fmt.Errorf("refusing to download from %q: not in the allowed host list", host)
	}
	// A literal address can be judged now, which turns the common case into
	// a clear error instead of a dial failure. A name cannot: it is judged
	// at dial time, which is also what makes DNS rebinding pointless here.
	if ip, perr := netip.ParseAddr(host); perr == nil {
		if err := checkDownloadAddr(ip, lim); err != nil {
			return nil, err
		}
	}
	return u, nil
}

func hostAllowed(host string, allowed []string) bool {
	host = strings.ToLower(strings.TrimSuffix(host, "."))
	for _, a := range allowed {
		if host == strings.ToLower(strings.TrimSuffix(strings.TrimSpace(a), ".")) {
			return true
		}
	}
	return false
}

// newGuardedClient builds the client used for peer-supplied URLs.
//
// Control is the authoritative gate: it runs after DNS resolution, on the
// address the kernel is about to connect to, for every hop. A name that
// resolves to 169.254.169.254 — whether by accident, by an attacker's zone,
// or by a second lookup returning something different from the first
// (DNS rebinding) — is refused there, because there is no gap between the
// check and the connect.
//
// Proxy is nil on purpose. http.Transport's default is
// ProxyFromEnvironment, and a proxy would make Control see the proxy's
// address rather than the target's — the guard would pass every URL.
func newGuardedClient(lim DownloadLimits) *http.Client {
	dialer := &net.Dialer{
		Timeout:   dialTimeout,
		KeepAlive: 30 * time.Second,
		Control: func(network, address string, _ syscall.RawConn) error {
			return checkDialAddress(network, address, lim)
		},
	}
	maxRedirects := lim.MaxRedirects
	if maxRedirects < 0 {
		maxRedirects = 0
	}
	return &http.Client{
		Timeout: lim.Timeout,
		Transport: &http.Transport{
			Proxy:                 nil,
			DialContext:           dialer.DialContext,
			TLSHandshakeTimeout:   tlsHandshakeTimeout,
			ResponseHeaderTimeout: responseHeaderTimeout,
			ExpectContinueTimeout: time.Second,
			DisableKeepAlives:     true,
		},
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			if len(via) > maxRedirects {
				return fmt.Errorf("refusing to follow more than %d redirects", maxRedirects)
			}
			_, err := ValidateDownloadURL(req.URL.String(), lim)
			return err
		},
	}
}

func checkDialAddress(network, address string, lim DownloadLimits) error {
	switch network {
	case "tcp", "tcp4", "tcp6":
	default:
		return fmt.Errorf("refusing to dial over %q", network)
	}
	host, _, err := net.SplitHostPort(address)
	if err != nil {
		return fmt.Errorf("refusing to dial %q: %w", address, err)
	}
	ip, err := netip.ParseAddr(host)
	if err != nil {
		return fmt.Errorf("refusing to dial %q: not an IP address", host)
	}
	return checkDownloadAddr(ip, lim)
}

// downloadFileName derives the temp file's name from the URL path.
//
// It is taken from the parsed path, not from the raw URL string: a query
// string in the raw form ("...?v=1/2") otherwise ends up in the name. The
// result is forced to a single plain component — the file lands in a
// freshly created private temp directory, and nothing downstream reads the
// name, so a URL that does not carry a usable one just gets "download".
func downloadFileName(u *url.URL) string {
	name := path.Base(u.Path)
	switch name {
	case "", ".", "..", "/":
		return "download"
	}
	if name != filepath.Base(name) || strings.ContainsAny(name, `/\`) {
		return "download"
	}
	if strings.ContainsRune(name, 0) || len(name) > 128 {
		return "download"
	}
	return name
}

// DownloadFileToTemp fetches a peer-supplied URL into a new private temp
// directory under the default data-object limits. It removes partial
// downloads on error; on success the caller owns the returned file and the
// directory holding it.
func DownloadFileToTemp(fileUrl string, pattern string) (tempFilePath string, err error) {
	return DownloadFileToTempWithLimits(context.Background(), fileUrl, pattern, DefaultDownloadLimits())
}

// DownloadFileToTempWithLimits is DownloadFileToTemp with explicit bounds
// and a caller context.
func DownloadFileToTempWithLimits(ctx context.Context, fileUrl string, pattern string, lim DownloadLimits) (tempFilePath string, err error) {
	// Validate before creating anything: a refused URL should leave no
	// directory behind.
	u, err := ValidateDownloadURL(fileUrl, lim)
	if err != nil {
		return "", err
	}

	tempDir, err := os.MkdirTemp("", pattern)
	if err != nil {
		return "", err
	}
	keepTempDir := false
	defer func() {
		if !keepTempDir {
			_ = os.RemoveAll(tempDir)
		}
	}()

	tempFilePath = filepath.Join(tempDir, downloadFileName(u))

	outFile, err := os.Create(tempFilePath)
	if err != nil {
		return "", err
	}
	// Closed explicitly on the success path below so a write error the
	// close reports (a delayed ENOSPC, an NFS flush) fails the download
	// instead of handing a truncated file to wazero or the ZTDO decoder.
	// This defer only covers the error returns.
	outFileClosed := false
	defer func() {
		if !outFileClosed {
			_ = outFile.Close()
		}
	}()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u.String(), nil)
	if err != nil {
		return "", err
	}

	resp, err := newGuardedClient(lim).Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("failed to download file (%s): status code %s", redactURL(fileUrl), resp.Status)
	}

	body := resp.Body
	if lim.MaxBytes > 0 {
		// A declared length over the cap is refused before a byte is read.
		if resp.ContentLength > lim.MaxBytes {
			return "", fmt.Errorf("%w: %s declares %d bytes, limit is %d", ErrDownloadTooLarge, redactURL(fileUrl), resp.ContentLength, lim.MaxBytes)
		}
		// Read one byte past the cap so an undeclared or lying length is
		// caught by what actually arrived, not by what was promised.
		body = struct {
			io.Reader
			io.Closer
		}{io.LimitReader(resp.Body, lim.MaxBytes+1), resp.Body}
	}

	written, err := io.Copy(outFile, body)
	if err != nil {
		return "", err
	}
	if lim.MaxBytes > 0 && written > lim.MaxBytes {
		return "", fmt.Errorf("%w: %s exceeds %d bytes", ErrDownloadTooLarge, redactURL(fileUrl), lim.MaxBytes)
	}

	if err := outFile.Close(); err != nil {
		return "", fmt.Errorf("failed to write the downloaded file: %w", err)
	}
	outFileClosed = true

	keepTempDir = true
	return tempFilePath, nil
}
