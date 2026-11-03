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
	"strings"
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

// The reservation shrinks as the file is written, since the filesystem's
// free space already accounts for written bytes; whatever is left (the
// multipart overhead, or a failed upload) is given back on release.
func TestUploadReservationShrinksAsWritten(t *testing.T) {
	st := &storageState{}
	other, _ := st.reserve(100) // another upload in progress
	r, total := st.reserve(1000)
	if total != 1100 {
		t.Fatalf("total reserved %d, want 1100", total)
	}
	if _, err := r.Write(make([]byte, 400)); err != nil {
		t.Fatal(err)
	}
	if got := st.reservedBytes.Load(); got != 700 {
		t.Fatalf("after writing 400 bytes: %d reserved, want 700", got)
	}
	// Writing past the declared size never releases more than was claimed.
	if _, err := r.Write(make([]byte, 5000)); err != nil {
		t.Fatal(err)
	}
	if got := st.reservedBytes.Load(); got != 100 {
		t.Fatalf("after overrunning: %d reserved, want the other upload's 100", got)
	}
	r.release()
	other.release()
	if got := st.reservedBytes.Load(); got != 0 {
		t.Fatalf("after release: %d reserved, want 0", got)
	}
}

// The replay cache belongs to the UdpServer, so rebuilding the HTTP server
// (an http.toml change) does not let a recent signed request be replayed.
func TestStorageNonceCacheSurvivesHttpRestart(t *testing.T) {
	f := newStorageFixture(t, common.CIPHER_SCHEME_CURVE, nil)
	md := httptest.NewRequest(http.MethodGet, "/storage/metadata/x", nil)
	f.sign(t, md, time.Now())
	replay := md.Clone(md.Context())
	if w := f.serve(md); w.Code == http.StatusUnauthorized {
		t.Fatalf("first request: got %d", w.Code)
	}

	restarted := &HttpServer{ginEngine: gin.New(), udpServer: f.hs.udpServer}
	restarted.initStorageRouter()
	w := httptest.NewRecorder()
	restarted.ginEngine.ServeHTTP(w, replay)
	if w.Code != http.StatusUnauthorized {
		t.Fatalf("replay after an HTTP restart: got %d, want 401", w.Code)
	}
}

// Malformed authentication headers are refused, whatever is wrong.
func TestStorageRejectsMalformedAuthHeaders(t *testing.T) {
	f := newStorageFixture(t, common.CIPHER_SCHEME_CURVE, nil)
	for name, mutate := range map[string]func(h http.Header){
		"no signature":    func(h http.Header) { h.Del(storageauth.HeaderSignature) },
		"bad timestamp":   func(h http.Header) { h.Set(storageauth.HeaderTimestamp, "yesterday") },
		"bad nonce":       func(h http.Header) { h.Set(storageauth.HeaderNonce, "short") },
		"bad body digest": func(h http.Header) { h.Set(storageauth.HeaderBodyHash, "zz") },
		"upper-case digest": func(h http.Header) {
			h.Set(storageauth.HeaderBodyHash, strings.ToUpper(h.Get(storageauth.HeaderBodyHash)))
		},
		"key not base64": func(h http.Header) { h.Set(storageauth.HeaderPublicKey, "!!!") },
		"key of wrong size": func(h http.Header) {
			h.Set(storageauth.HeaderPublicKey, base64.StdEncoding.EncodeToString([]byte("short key")))
		},
		"signature mismatch": func(h http.Header) {
			h.Set(storageauth.HeaderSignature, base64.StdEncoding.EncodeToString(make([]byte, 32)))
		},
	} {
		md := httptest.NewRequest(http.MethodGet, "/storage/metadata/x", nil)
		f.sign(t, md, time.Now())
		mutate(md.Header)
		if w := f.serve(md); w.Code != http.StatusUnauthorized {
			t.Errorf("%s: got %d, want 401", name, w.Code)
		}
	}
}

// Authenticated requests that are not well-formed uploads are refused
// before anything is stored.
func TestStorageUploadRejectsMalformedBodies(t *testing.T) {
	f := newStorageFixture(t, common.CIPHER_SCHEME_CURVE, nil)

	// A signed upload with no body reaches the handler and gets 411. (A body
	// sent without Content-Length is refused earlier, with 401: its digest
	// cannot match the empty one a zero length implies.)
	noBody := httptest.NewRequest(http.MethodPost, "/storage/upload", nil)
	f.sign(t, noBody, time.Now())
	if w := f.serve(noBody); w.Code != http.StatusLengthRequired {
		t.Errorf("no body: got %d, want 411", w.Code)
	}

	notMultipart := httptest.NewRequest(http.MethodPost, "/storage/upload", strings.NewReader("plain body"))
	notMultipart.Header.Set("Content-Type", "text/plain")
	f.sign(t, notMultipart, time.Now())
	if w := f.serve(notMultipart); w.Code != http.StatusBadRequest {
		t.Errorf("not multipart: got %d, want 400", w.Code)
	}

	body := &bytes.Buffer{}
	mw := multipart.NewWriter(body)
	_ = mw.WriteField("other", "value")
	_ = mw.Close()
	noFile := httptest.NewRequest(http.MethodPost, "/storage/upload", bytes.NewReader(body.Bytes()))
	noFile.Header.Set("Content-Type", mw.FormDataContentType())
	f.sign(t, noFile, time.Now())
	if w := f.serve(noFile); w.Code != http.StatusBadRequest {
		t.Errorf("no file part: got %d, want 400", w.Code)
	}

	for _, name := range []string{"a:b.ztdo", "nul\x00.ztdo"} {
		r := uploadRequest(t, name, []byte("x"))
		f.sign(t, r, time.Now())
		if w := f.serve(r); w.Code != http.StatusBadRequest {
			t.Errorf("file name %q: got %d, want 400", name, w.Code)
		}
	}

	if objs := storedObjects(t); len(objs) != 0 {
		t.Fatalf("malformed uploads left %v", objs)
	}

	missing := httptest.NewRequest(http.MethodGet, "/storage/metadata/no-such-object", nil)
	f.sign(t, missing, time.Now())
	if w := f.serve(missing); w.Code != http.StatusNotFound {
		t.Errorf("metadata of a missing object: got %d, want 404", w.Code)
	}
}

// An upload of a file that is already stored returns the stored object and
// removes the duplicate copy.
func TestStorageUploadDeduplicates(t *testing.T) {
	f := newStorageFixture(t, common.CIPHER_SCHEME_CURVE, nil)
	var uuids []string
	for i := 0; i < 2; i++ {
		r := uploadRequest(t, "same.ztdo", []byte("identical content"))
		f.sign(t, r, time.Now())
		w := f.serve(r)
		if w.Code != http.StatusOK {
			t.Fatalf("upload %d: got %d %s", i, w.Code, w.Body.String())
		}
		var resp struct {
			UUID string `json:"uuid"`
		}
		_ = json.Unmarshal(w.Body.Bytes(), &resp)
		uuids = append(uuids, resp.UUID)
	}
	if uuids[0] != uuids[1] {
		t.Fatalf("duplicate upload got a new object: %v", uuids)
	}
	if objs := storedObjects(t); len(objs) != 1 {
		t.Fatalf("duplicate upload left %v", objs)
	}
}

func TestIsPlainStorageName(t *testing.T) {
	for name, want := range map[string]bool{
		"data.ztdo": true, "报告.ztdo": true,
		"": false, ".": false, "..": false, "a/b": false, `a\b`: false, "a:b": false, "a\x00b": false, "a\x7fb": false,
	} {
		if got := isPlainStorageName(name); got != want {
			t.Errorf("isPlainStorageName(%q) = %v, want %v", name, got, want)
		}
	}
}
