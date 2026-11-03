package server

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strconv"
	"testing"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/OpenNHP/opennhp/nhp/common"
	"github.com/OpenNHP/opennhp/nhp/common/storageauth"
	"github.com/OpenNHP/opennhp/nhp/core"
)

// storageFixture is an HttpServer whose UdpServer has a device key and one
// nhp-db listed in db.toml, plus that nhp-db's device for signing.
type storageFixture struct {
	hs     *HttpServer
	db     *core.Device
	scheme int
}

func newStorageFixture(t *testing.T, scheme int, hc *HttpConfig) *storageFixture {
	t.Helper()
	gin.SetMode(gin.TestMode)
	originalExeDir := ExeDirPath
	ExeDirPath = t.TempDir()
	t.Cleanup(func() { ExeDirPath = originalExeDir })

	server := core.NewDevice(core.NHP_SERVER, core.NewECDH(core.ECC_CURVE25519).PrivateKey(), nil)
	db := core.NewDevice(core.NHP_DB, core.NewECDH(core.ECC_CURVE25519).PrivateKey(), nil)
	if server == nil || db == nil {
		t.Fatal("NewDevice failed")
	}
	dbPub := db.PublicKeyBase64()
	if scheme == common.CIPHER_SCHEME_GMSM {
		dbPub = db.PublicKeyExBase64()
	}
	us := &UdpServer{
		device:    server,
		dbPeerMap: map[string]*core.UdpPeer{dbPub: {PubKeyBase64: dbPub}},
	}
	hs := &HttpServer{ginEngine: gin.New(), udpServer: us, config: hc}
	hs.initStorageRouter()
	return &storageFixture{hs: hs, db: db, scheme: scheme}
}

// sign signs req the way nhp-db's signStorageRequest does.
func (f *storageFixture) sign(t *testing.T, req *http.Request, ts time.Time) {
	t.Helper()
	serverPub := f.hs.udpServer.device.GetEcdhByCipherScheme(f.scheme).PublicKey()
	selfPub := f.db.PublicKeyBase64()
	if f.scheme == common.CIPHER_SCHEME_GMSM {
		selfPub = f.db.PublicKeyExBase64()
	}
	secret := f.db.GetEcdhByCipherScheme(f.scheme).SharedSecret(serverPub)
	var body []byte
	if req.Body != nil {
		var err error
		if body, err = io.ReadAll(req.Body); err != nil {
			t.Fatal(err)
		}
		req.Body = io.NopCloser(bytes.NewReader(body))
	}
	sr, err := storageauth.NewRequest(req.Method, req.URL.Path, body, selfPub, base64.StdEncoding.EncodeToString(serverPub))
	if err != nil {
		t.Fatal(err)
	}
	// Sign the Content-Length the request declares, even when a test makes
	// it disagree with the body.
	sr.ContentLength = max(req.ContentLength, 0)
	sr.Timestamp = ts.Unix()
	sr.SetHeaders(req.Header, storageauth.Sign(secret, sr))
}

func (f *storageFixture) serve(req *http.Request) *httptest.ResponseRecorder {
	w := httptest.NewRecorder()
	f.hs.ginEngine.ServeHTTP(w, req)
	return w
}

func uploadRequest(t *testing.T, filename string, content []byte) *http.Request {
	t.Helper()
	body := &bytes.Buffer{}
	mw := multipart.NewWriter(body)
	part, err := mw.CreateFormFile("file", filename)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := part.Write(content); err != nil {
		t.Fatal(err)
	}
	if err := mw.Close(); err != nil {
		t.Fatal(err)
	}
	req := httptest.NewRequest(http.MethodPost, "/storage/upload", bytes.NewReader(body.Bytes()))
	req.ContentLength = int64(body.Len())
	req.Header.Set("Content-Type", mw.FormDataContentType())
	return req
}

func storedObjects(t *testing.T) []string {
	t.Helper()
	entries, err := os.ReadDir(filepath.Join(ExeDirPath, uploadDir))
	if os.IsNotExist(err) {
		return nil
	}
	if err != nil {
		t.Fatal(err)
	}
	var names []string
	for _, e := range entries {
		names = append(names, e.Name())
	}
	return names
}

// An nhp-db listed in db.toml can upload with either scheme; the object is
// then downloadable without credentials.
func TestStorageSignedUploadAndAnonymousDownload(t *testing.T) {
	for name, scheme := range map[string]int{"curve": common.CIPHER_SCHEME_CURVE, "gmsm": common.CIPHER_SCHEME_GMSM} {
		t.Run(name, func(t *testing.T) {
			f := newStorageFixture(t, scheme, nil)
			req := uploadRequest(t, "data.ztdo", []byte("ztdo payload "+name))
			f.sign(t, req, time.Now())
			w := f.serve(req)
			if w.Code != http.StatusOK {
				t.Fatalf("signed upload: got %d %s", w.Code, w.Body.String())
			}
			var resp struct {
				FileURI string `json:"file_uri"`
				UUID    string `json:"uuid"`
			}
			if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
				t.Fatal(err)
			}

			dl := f.serve(httptest.NewRequest(http.MethodGet, "/"+resp.FileURI, nil))
			if dl.Code != http.StatusOK || dl.Body.String() != "ztdo payload "+name {
				t.Fatalf("anonymous download: got %d %q", dl.Code, dl.Body.String())
			}

			md := httptest.NewRequest(http.MethodGet, "/storage/metadata/"+resp.UUID, nil)
			if w := f.serve(md); w.Code != http.StatusUnauthorized {
				t.Fatalf("unsigned metadata: got %d", w.Code)
			}
			md = httptest.NewRequest(http.MethodGet, "/storage/metadata/"+resp.UUID, nil)
			f.sign(t, md, time.Now())
			if w := f.serve(md); w.Code != http.StatusOK {
				t.Fatalf("signed metadata: got %d %s", w.Code, w.Body.String())
			}
		})
	}
}

// Requests that are unsigned, signed by an unknown or expired nhp-db,
// replayed, stale or tampered with are refused before anything is stored.
func TestStorageUploadRejectsUnauthenticated(t *testing.T) {
	f := newStorageFixture(t, common.CIPHER_SCHEME_CURVE, nil)
	content := []byte("payload")

	cases := map[string]func() *http.Request{
		"unsigned": func() *http.Request { return uploadRequest(t, "a.ztdo", content) },
		"stale timestamp": func() *http.Request {
			r := uploadRequest(t, "a.ztdo", content)
			f.sign(t, r, time.Now().Add(-2*storageauth.MaxClockSkew))
			return r
		},
		"future timestamp": func() *http.Request {
			r := uploadRequest(t, "a.ztdo", content)
			f.sign(t, r, time.Now().Add(2*storageauth.MaxClockSkew))
			return r
		},
		"tampered content length": func() *http.Request {
			r := uploadRequest(t, "a.ztdo", content)
			f.sign(t, r, time.Now())
			bigger := uploadRequest(t, "a.ztdo", append(content, 'x'))
			bigger.Header = r.Header.Clone()
			return bigger
		},
		"bad signature": func() *http.Request {
			r := uploadRequest(t, "a.ztdo", content)
			f.sign(t, r, time.Now())
			r.Header.Set(storageauth.HeaderSignature, "AAAA")
			return r
		},
		"unknown nhp-db": func() *http.Request {
			r := uploadRequest(t, "a.ztdo", content)
			other := &storageFixture{hs: f.hs, db: core.NewDevice(core.NHP_DB, core.NewECDH(core.ECC_CURVE25519).PrivateKey(), nil), scheme: f.scheme}
			other.sign(t, r, time.Now())
			return r
		},
	}
	for name, build := range cases {
		if w := f.serve(build()); w.Code != http.StatusUnauthorized {
			t.Errorf("%s: got %d %s, want 401", name, w.Code, w.Body.String())
		}
	}
	if objs := storedObjects(t); len(objs) != 0 {
		t.Fatalf("rejected uploads left %v", objs)
	}

	// A replayed request: the same headers a second time.
	r := uploadRequest(t, "a.ztdo", content)
	f.sign(t, r, time.Now())
	replay := uploadRequest(t, "a.ztdo", content)
	replay.Header = r.Header.Clone()
	if w := f.serve(r); w.Code != http.StatusOK {
		t.Fatalf("first request: got %d %s", w.Code, w.Body.String())
	}
	if w := f.serve(replay); w.Code != http.StatusUnauthorized {
		t.Fatalf("replay: got %d, want 401", w.Code)
	}
}

// The signature covers the body: an on-path attacker who keeps the signed
// headers but swaps the content (same length) is refused, and nothing is
// stored.
func TestStorageUploadRejectsSwappedBody(t *testing.T) {
	f := newStorageFixture(t, common.CIPHER_SCHEME_CURVE, nil)
	signed := uploadRequest(t, "a.ztdo", []byte("genuine ztdo"))
	f.sign(t, signed, time.Now())

	swapped := uploadRequest(t, "a.ztdo", []byte("evil!! ztdo!"))
	if swapped.ContentLength != signed.ContentLength {
		t.Fatal("test bodies must have the same length")
	}
	swapped.Header = signed.Header.Clone()
	if w := f.serve(swapped); w.Code != http.StatusBadRequest {
		t.Fatalf("swapped body: got %d %s, want 400", w.Code, w.Body.String())
	}
	if objs := storedObjects(t); len(objs) != 0 {
		t.Fatalf("swapped body left %v", objs)
	}

	// A bodyless request must declare the empty digest.
	md := httptest.NewRequest(http.MethodGet, "/storage/metadata/x", nil)
	f.sign(t, md, time.Now())
	md.Header.Set(storageauth.HeaderBodyHash, storageauth.BodyHash([]byte("x")))
	if w := f.serve(md); w.Code != http.StatusUnauthorized {
		t.Fatalf("bodyless request with a non-empty digest: got %d, want 401", w.Code)
	}
}

func TestStorageUploadRejectsExpiredDB(t *testing.T) {
	f := newStorageFixture(t, common.CIPHER_SCHEME_CURVE, nil)
	for _, p := range f.hs.udpServer.dbPeerMap {
		p.ExpireTime = time.Now().Add(-time.Hour).Unix()
	}
	r := uploadRequest(t, "a.ztdo", []byte("x"))
	f.sign(t, r, time.Now())
	if w := f.serve(r); w.Code != http.StatusUnauthorized {
		t.Fatalf("expired nhp-db: got %d, want 401", w.Code)
	}
}

// Size, free-space and concurrency limits refuse an upload without leaving
// anything behind.
func TestStorageUploadLimits(t *testing.T) {
	t.Run("too large", func(t *testing.T) {
		f := newStorageFixture(t, common.CIPHER_SCHEME_CURVE, &HttpConfig{StorageMaxUploadMB: 1})
		r := uploadRequest(t, "big.ztdo", bytes.Repeat([]byte("x"), 2<<20))
		f.sign(t, r, time.Now())
		if w := f.serve(r); w.Code != http.StatusRequestEntityTooLarge {
			t.Fatalf("got %d, want 413", w.Code)
		}
		if objs := storedObjects(t); len(objs) != 0 {
			t.Fatalf("left %v", objs)
		}
	})
	t.Run("insufficient free space", func(t *testing.T) {
		if _, ok := diskFreeBytes(os.TempDir()); !ok {
			t.Skip("free space is not available on this platform")
		}
		f := newStorageFixture(t, common.CIPHER_SCHEME_CURVE, &HttpConfig{StorageMinFreeMB: 1 << 40})
		r := uploadRequest(t, "a.ztdo", []byte("x"))
		f.sign(t, r, time.Now())
		if w := f.serve(r); w.Code != http.StatusInsufficientStorage {
			t.Fatalf("got %d, want 507", w.Code)
		}
	})
	t.Run("free space reserved by uploads in progress", func(t *testing.T) {
		free, ok := diskFreeBytes(os.TempDir())
		if !ok {
			t.Skip("free space is not available on this platform")
		}
		f := newStorageFixture(t, common.CIPHER_SCHEME_CURVE, &HttpConfig{StorageMinFreeMB: 1})
		// Another upload in progress has reserved all but a little of it.
		f.hs.storage.reservedBytes.Store(free)
		r := uploadRequest(t, "a.ztdo", []byte("x"))
		f.sign(t, r, time.Now())
		if w := f.serve(r); w.Code != http.StatusInsufficientStorage {
			t.Fatalf("got %d, want 507", w.Code)
		}
		if got := f.hs.storage.reservedBytes.Load(); got != free {
			t.Fatalf("reservation not released: %d, want %d", got, free)
		}
	})
	t.Run("too many concurrent uploads", func(t *testing.T) {
		f := newStorageFixture(t, common.CIPHER_SCHEME_CURVE, &HttpConfig{StorageMaxConcurrentUploads: 1})
		f.hs.storage.uploadSlots <- struct{}{} // one upload in progress
		r := uploadRequest(t, "a.ztdo", []byte("x"))
		f.sign(t, r, time.Now())
		if w := f.serve(r); w.Code != http.StatusTooManyRequests {
			t.Fatalf("got %d, want 429", w.Code)
		}
	})
	t.Run("invalid file name", func(t *testing.T) {
		f := newStorageFixture(t, common.CIPHER_SCHEME_CURVE, nil)
		r := uploadRequest(t, `a\b.ztdo`, []byte("x"))
		f.sign(t, r, time.Now())
		if w := f.serve(r); w.Code != http.StatusBadRequest {
			t.Fatalf("got %d, want 400", w.Code)
		}
	})
	t.Run("body overruns the cap while copying", func(t *testing.T) {
		f := newStorageFixture(t, common.CIPHER_SCHEME_CURVE, nil)
		f.hs.storage.maxUploadBytes = 64 << 10
		r := uploadRequest(t, "a.ztdo", bytes.Repeat([]byte("x"), 900<<10))
		// Declare less than is sent, so the body passes the up-front size
		// check and overruns the cap while the file is being written.
		r.ContentLength = 32 << 10
		f.sign(t, r, time.Now())
		if w := f.serve(r); w.Code != http.StatusRequestEntityTooLarge {
			t.Fatalf("got %d, want 413", w.Code)
		}
		if objs := storedObjects(t); len(objs) != 0 {
			t.Fatalf("failed upload left %v", objs)
		}
	})
}

func TestStorageCanBeDisabled(t *testing.T) {
	off := false
	f := newStorageFixture(t, common.CIPHER_SCHEME_CURVE, &HttpConfig{EnableStorage: &off})
	r := uploadRequest(t, "a.ztdo", []byte("x"))
	f.sign(t, r, time.Now())
	if w := f.serve(r); w.Code != http.StatusNotFound {
		t.Fatalf("disabled storage: got %d, want 404", w.Code)
	}
}

// The signed routes are rate limited per source IP, keyed by the TCP peer
// address rather than X-Forwarded-For; anonymous downloads are not limited.
func TestStorageRateLimit(t *testing.T) {
	f := newStorageFixture(t, common.CIPHER_SCHEME_CURVE, nil)
	limited := false
	for i := 0; i < storageRateBurst+5; i++ {
		req := httptest.NewRequest(http.MethodGet, "/storage/metadata/x", nil)
		req.Header.Set("X-Forwarded-For", "10.0.0."+strconv.Itoa(i))
		if w := f.serve(req); w.Code == http.StatusTooManyRequests {
			limited = true
			break
		}
	}
	if !limited {
		t.Fatal("rotating X-Forwarded-For evaded the per-IP rate limit")
	}

	for i := 0; i < 3*storageRateBurst; i++ {
		req := httptest.NewRequest(http.MethodGet, "/storage/download/x/y", nil)
		if w := f.serve(req); w.Code == http.StatusTooManyRequests {
			t.Fatalf("anonymous download %d was rate limited", i)
		}
	}
}

func TestNonceCacheBounded(t *testing.T) {
	nc := newNonceCache(2)
	now := time.Now()
	if nc.add("a", now) != nonceAccepted || nc.add("b", now) != nonceAccepted {
		t.Fatal("fresh nonces refused")
	}
	if nc.add("a", now) != nonceReplayed {
		t.Fatal("replayed nonce accepted")
	}
	if nc.add("c", now) != nonceCacheFull {
		t.Fatal("full cache accepted a nonce while every entry is live")
	}
	later := now.Add(storageNonceTTL + time.Second)
	if nc.add("c", later) != nonceAccepted {
		t.Fatal("expired entries were not reclaimed")
	}
	if nc.add("a", later) != nonceAccepted {
		t.Fatal("an expired nonce was still remembered")
	}
}

// A valid request refused only because the nonce cache is full gets 503,
// not a 401 that would send the operator looking at keys and clocks.
func TestStorageFullNonceCacheIs503(t *testing.T) {
	f := newStorageFixture(t, common.CIPHER_SCHEME_CURVE, nil)
	f.hs.storage.nonces = newNonceCache(0)
	md := httptest.NewRequest(http.MethodGet, "/storage/metadata/x", nil)
	f.sign(t, md, time.Now())
	if w := f.serve(md); w.Code != http.StatusServiceUnavailable {
		t.Fatalf("got %d, want 503", w.Code)
	}
}
