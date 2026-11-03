package db

import (
	"encoding/base64"
	"encoding/json"
	"io"
	"net"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/OpenNHP/opennhp/nhp/common"
	"github.com/OpenNHP/opennhp/nhp/common/storageauth"
	"github.com/OpenNHP/opennhp/nhp/core"
)

// fakeStorageServer stands in for nhp-server's /storage API: it verifies
// the nhp-db signature and body digest the way nhp-server does and answers
// with the status the test asks for.
type fakeStorageServer struct {
	t       *testing.T
	server  *core.Device
	scheme  int
	status  int // 0: verify and answer 200 or 401
	checked bool
}

func (f *fakeStorageServer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path == "/" { // the http/https probe
		w.WriteHeader(http.StatusOK)
		return
	}
	if f.status != 0 {
		w.WriteHeader(f.status)
		return
	}
	body, err := io.ReadAll(r.Body)
	if err != nil {
		f.t.Errorf("read body: %v", err)
		return
	}
	pubB64 := r.Header.Get(storageauth.HeaderPublicKey)
	pub, _ := base64.StdEncoding.DecodeString(pubB64)
	ts, _ := strconv.ParseInt(r.Header.Get(storageauth.HeaderTimestamp), 10, 64)
	ecdh := f.server.GetEcdhByCipherScheme(f.scheme)
	req := storageauth.Request{
		Method:                r.Method,
		Path:                  r.URL.Path,
		ContentLength:         r.ContentLength,
		PublicKeyBase64:       pubB64,
		ServerPublicKeyBase64: base64.StdEncoding.EncodeToString(ecdh.PublicKey()),
		Timestamp:             ts,
		Nonce:                 r.Header.Get(storageauth.HeaderNonce),
		BodySHA256:            r.Header.Get(storageauth.HeaderBodyHash),
	}
	f.checked = true
	if req.BodySHA256 != storageauth.BodyHash(body) ||
		!storageauth.Verify(ecdh.SharedSecret(pub), req, r.Header.Get(storageauth.HeaderSignature)) {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}
	_ = json.NewEncoder(w).Encode(ServerResponse{FileURI: "storage/download/id/data.ztdo"})
}

// newUploadDevice returns an nhp-db device whose single nhp-server peer is
// the given HTTP test server, with that server's key of the given scheme.
func newUploadDevice(t *testing.T, scheme int, handler *fakeStorageServer) (*UdpDevice, string) {
	t.Helper()
	server := core.NewDevice(core.NHP_SERVER, core.NewECDH(core.ECC_CURVE25519).PrivateKey(), nil)
	db := core.NewDevice(core.NHP_DB, core.NewECDH(core.ECC_CURVE25519).PrivateKey(), nil)
	if server == nil || db == nil {
		t.Fatal("NewDevice failed")
	}
	handler.t, handler.server, handler.scheme = t, server, scheme

	ts := httptest.NewServer(handler)
	t.Cleanup(ts.Close)
	host, portStr, _ := net.SplitHostPort(strings.TrimPrefix(ts.URL, "http://"))
	port, _ := strconv.Atoi(portStr)

	serverPub := server.PublicKeyBase64()
	if scheme == common.CIPHER_SCHEME_GMSM {
		serverPub = server.PublicKeyExBase64()
	}
	peer := &core.UdpPeer{PubKeyBase64: serverPub, Ip: host, Port: port}
	return &UdpDevice{device: db, serverPeerMap: map[string]*core.UdpPeer{serverPub: peer}}, ts.URL + "/"
}

func writeUploadFile(t *testing.T) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), "data.ztdo")
	if err := os.WriteFile(path, []byte("ztdo bytes"), 0o600); err != nil {
		t.Fatal(err)
	}
	return path
}

// nhp-db signs its upload so that nhp-server can verify it, with the
// scheme of the server key in server.toml.
func TestUploadFileToNHPServerSigns(t *testing.T) {
	for name, scheme := range map[string]int{"curve": common.CIPHER_SCHEME_CURVE, "gmsm": common.CIPHER_SCHEME_GMSM} {
		t.Run(name, func(t *testing.T) {
			handler := &fakeStorageServer{}
			a, base := newUploadDevice(t, scheme, handler)
			url, err := a.UploadFileToNHPServer(writeUploadFile(t))
			if err != nil {
				t.Fatalf("upload: %v", err)
			}
			if !handler.checked {
				t.Fatal("the upload never reached the storage endpoint")
			}
			if url != base+"storage/download/id/data.ztdo" {
				t.Fatalf("download URL %q", url)
			}
		})
	}
}

// A refused upload reports what the operator has to check.
func TestUploadFileToNHPServerErrors(t *testing.T) {
	for name, tc := range map[string]struct {
		status int
		want   string
	}{
		"unauthorized": {http.StatusUnauthorized, "db.toml must list this nhp-db's public key"},
		"nonce cache":  {http.StatusServiceUnavailable, "retry later"},
		"other":        {http.StatusInternalServerError, "unexpected status code: 500"},
	} {
		t.Run(name, func(t *testing.T) {
			a, _ := newUploadDevice(t, common.CIPHER_SCHEME_CURVE, &fakeStorageServer{status: tc.status})
			_, err := a.UploadFileToNHPServer(writeUploadFile(t))
			if err == nil || !strings.Contains(err.Error(), tc.want) {
				t.Fatalf("got %v, want an error containing %q", err, tc.want)
			}
		})
	}
	t.Run("unauthorized names the key", func(t *testing.T) {
		a, _ := newUploadDevice(t, common.CIPHER_SCHEME_GMSM, &fakeStorageServer{status: http.StatusUnauthorized})
		_, err := a.UploadFileToNHPServer(writeUploadFile(t))
		if err == nil || !strings.Contains(err.Error(), a.device.PublicKeyExBase64()) {
			t.Fatalf("401 error does not name the SM2 key it signed with: %v", err)
		}
	})
}

func TestSignStorageRequestNeedsUsableServerKey(t *testing.T) {
	db := core.NewDevice(core.NHP_DB, core.NewECDH(core.ECC_CURVE25519).PrivateKey(), nil)
	req, _ := http.NewRequest(http.MethodPost, "http://127.0.0.1/storage/upload", nil)

	a := &UdpDevice{device: db}
	if _, err := a.signStorageRequest(req, nil, nil); err == nil {
		t.Fatal("signed without an nhp-server peer")
	}
	if _, err := a.UploadFileToNHPServer(writeUploadFile(t)); err == nil {
		t.Fatal("uploaded without an nhp-server peer")
	}
	bad := &core.UdpPeer{PubKeyBase64: base64.StdEncoding.EncodeToString([]byte("not a key"))}
	if _, err := a.signStorageRequest(req, nil, bad); err == nil {
		t.Fatal("signed for a server key of the wrong size")
	}
}

// The signature must expire with the clock skew window, so check the
// timestamp nhp-db puts on the request is current.
func TestSignStorageRequestTimestamp(t *testing.T) {
	handler := &fakeStorageServer{}
	a, base := newUploadDevice(t, common.CIPHER_SCHEME_CURVE, handler)
	req, _ := http.NewRequest(http.MethodPost, base+"storage/upload", strings.NewReader("x"))
	if _, err := a.signStorageRequest(req, []byte("x"), a.GetServerPeer()); err != nil {
		t.Fatal(err)
	}
	ts, err := strconv.ParseInt(req.Header.Get(storageauth.HeaderTimestamp), 10, 64)
	if err != nil || time.Since(time.Unix(ts, 0)).Abs() > time.Minute {
		t.Fatalf("timestamp %q is not current", req.Header.Get(storageauth.HeaderTimestamp))
	}
}

// With several nhp-servers in server.toml, the upload is signed for the
// server it is sent to. GetServerPeer ranges over a map, so picking the
// peer more than once used to pair one server's address with another's key.
func TestUploadFileToNHPServerSignsForTheServerItContacts(t *testing.T) {
	handler := &fakeStorageServer{}
	a, _ := newUploadDevice(t, common.CIPHER_SCHEME_CURVE, handler)
	// Add servers that are listed but unreachable and have other keys.
	for i := 0; i < 4; i++ {
		other := core.NewDevice(core.NHP_SERVER, core.NewECDH(core.ECC_CURVE25519).PrivateKey(), nil)
		pub := other.PublicKeyBase64()
		a.serverPeerMap[pub] = &core.UdpPeer{PubKeyBase64: pub, Ip: "127.0.0.1", Port: 1}
	}
	// Every upload that reaches the test server must carry a valid
	// signature for it; uploads that pick an unreachable server fail to
	// connect, which is not what this test is about.
	for i := 0; i < 40; i++ {
		handler.checked = false
		_, err := a.UploadFileToNHPServer(writeUploadFile(t))
		if handler.checked && err != nil {
			t.Fatalf("upload to the test server was refused: %v", err)
		}
	}
}
