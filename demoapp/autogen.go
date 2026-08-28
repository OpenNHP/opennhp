package main

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// autogenFile is the sidecar config the demoapp writes next to the
// operator-authored etc/config.toml. Keeping the operator's file pristine
// means its inline comments survive across rebuilds — a fresh container
// only sees the auto-generated layer fill in placeholders.
//
// Values stored here:
//   - KeyEnvelopeKey: AES-256 master key (base64, 32 bytes)
//   - SessionKey:     session cookie signing key (>= 16 bytes recommended)
//   - ServerPublicKey: server pubkey fetched from the relay's /servers
//
// On first boot the demoapp fills placeholders. On subsequent boots it
// loads the sidecar as a base layer and the operator's config.toml as an
// override — explicit operator edits always win.
const autogenFile = ".demoapp-autogen.json"

// autogenValues is the JSON shape persisted to autogenFile. Versioned so
// future schema changes can migrate.
type autogenValues struct {
	Version         int    `json:"version"`
	KeyEnvelopeKey  string `json:"keyEnvelopeKey"`
	SessionKey      string `json:"sessionKey"`
	ServerPublicKey string `json:"serverPublicKey,omitempty"`
	// CipherSchemeNum is the relay's integer cipher scheme (0/1) at the
	// time ServerPublicKey was fetched. Persisted so a restart that
	// re-reads the sidecar before probing /servers knows which scheme
	// the persisted key belongs to. 0=Curve25519, 1=GMSM.
	CipherSchemeNum int `json:"cipherSchemeNum,omitempty"`
}

// isPlaceholder returns true for the canonical "operator must replace
// this" markers we ship in the example configs. We treat them as "auto-
// generate me" rather than as literal values so a fresh `docker compose up`
// works without an editor step.
func isPlaceholder(v string) bool {
	v = strings.TrimSpace(v)
	if v == "" {
		return true
	}
	upper := strings.ToUpper(v)
	if strings.HasPrefix(upper, "REPLACE_") || strings.HasPrefix(upper, "CHANGE_ME") {
		return true
	}
	return false
}

// generateRandomKey returns n random bytes encoded as base64.
func generateRandomKey(n int) (string, error) {
	b := make([]byte, n)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return base64.StdEncoding.EncodeToString(b), nil
}

// loadAutogen reads the sidecar (if present) and returns its values.
// Missing files are not errors — that's the first-boot case the auto-
// generator handles.
func loadAutogen(etcDir string) (*autogenValues, error) {
	path := filepath.Join(etcDir, autogenFile)
	data, err := os.ReadFile(path)
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			return &autogenValues{Version: 1}, nil
		}
		return nil, fmt.Errorf("read autogen file: %w", err)
	}
	out := &autogenValues{}
	if err := json.Unmarshal(data, out); err != nil {
		return nil, fmt.Errorf("parse autogen file: %w", err)
	}
	return out, nil
}

// saveAutogen persists the generated values so a container restart keeps
// them. The file is overwritten atomically (write to .tmp + rename) so a
// crash during write doesn't leave a half-written key.
func saveAutogen(etcDir string, v *autogenValues) error {
	if err := os.MkdirAll(etcDir, 0o755); err != nil {
		return fmt.Errorf("mkdir etc: %w", err)
	}
	path := filepath.Join(etcDir, autogenFile)
	tmp := path + ".tmp"
	data, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal autogen: %w", err)
	}
	if err := os.WriteFile(tmp, data, 0o600); err != nil {
		return fmt.Errorf("write autogen tmp: %w", err)
	}
	return os.Rename(tmp, path)
}

// relayServerInfo mirrors endpoints/relay/relay.go serverInfo so we don't
// pull the relay module into the demoapp module. Only the fields we need.
//
// CipherScheme is the relay's int scheme (0 = Curve25519, 1 = GMSM).
// Demoapp auto-detects this on first boot when CipherScheme is still at
// its default — see autoFillConfig.
type relayServerInfo struct {
	ID              string `json:"id"`
	Name            string `json:"name,omitempty"`
	PublicKeyBase64 string `json:"publicKeyBase64"`
	CipherScheme    int    `json:"cipherScheme"`
}

// relayServerSnapshot bundles the per-server fields demoapp needs from
// /servers: the public key AND the cipher scheme. Together they let us
// detect a Curve25519 vs GMSM relay and pick the matching scheme.
type relayServerSnapshot struct {
	PublicKeyBase64 string
	CipherScheme    int
}

// fetchServerSnapshotFromRelay hits GET <relayURL>/servers on the local
// relay and returns the first server's publicKeyBase64 + cipherScheme.
// The endpoint is intentionally unauthenticated (see relay.go:handleServers).
//
// Retries with a short backoff: in docker compose the relay container may
// not have finished registering the nhp-server yet when demoapp first
// starts, so a single 3s probe often catches "0 servers" before the
// relay has anything to report. We retry until either we get a non-empty
// list or ctx (caller's deadline) elapses.
func fetchServerSnapshotFromRelay(ctx context.Context, relayURL string) (relayServerSnapshot, error) {
	if relayURL == "" {
		return relayServerSnapshot{}, errors.New("RelayUrl is empty")
	}
	url := strings.TrimRight(relayURL, "/") + "/servers"

	const probeTimeout = 3 * time.Second
	const backoff = 1 * time.Second
	const maxAttempts = 30 // ~30s total: long enough for relay to register server
	for attempt := 1; attempt <= maxAttempts; attempt++ {
		snap, err := probeRelayOnce(ctx, url, probeTimeout)
		if err == nil {
			return snap, nil
		}
		// Log every few attempts so an operator tailing docker logs sees
		// the relay-bringing-up dance.
		if attempt == 1 || attempt%5 == 0 {
			fmt.Fprintf(os.Stderr, "[demoapp] relay /servers attempt %d/%d: %v\n", attempt, maxAttempts, err)
		}
		select {
		case <-ctx.Done():
			return relayServerSnapshot{}, fmt.Errorf("relay /servers never returned a server after %d attempts: %w", attempt, ctx.Err())
		case <-time.After(backoff):
		}
	}
	return relayServerSnapshot{}, fmt.Errorf("relay /servers never returned a server after %d attempts", maxAttempts)
}

// probeRelayOnce is the single-shot helper behind fetchServerSnapshotFromRelay.
func probeRelayOnce(ctx context.Context, url string, timeout time.Duration) (relayServerSnapshot, error) {
	probeCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()
	req, err := http.NewRequestWithContext(probeCtx, http.MethodGet, url, nil)
	if err != nil {
		return relayServerSnapshot{}, fmt.Errorf("build request: %w", err)
	}
	r, err := http.DefaultClient.Do(req)
	if err != nil {
		return relayServerSnapshot{}, err
	}
	defer r.Body.Close()
	if r.StatusCode != http.StatusOK {
		return relayServerSnapshot{}, fmt.Errorf("HTTP %d", r.StatusCode)
	}
	var servers []relayServerInfo
	if err := json.NewDecoder(r.Body).Decode(&servers); err != nil {
		return relayServerSnapshot{}, fmt.Errorf("decode: %w", err)
	}
	if len(servers) == 0 || servers[0].PublicKeyBase64 == "" {
		return relayServerSnapshot{}, fmt.Errorf("relay has no servers yet")
	}
	return relayServerSnapshot{
		PublicKeyBase64: servers[0].PublicKeyBase64,
		CipherScheme:    servers[0].CipherScheme,
	}, nil
}

// relaySchemeToString maps the relay's int scheme to our CipherScheme
// string. Returns "" for unknown values so the caller can leave the
// operator-configured value untouched.
func relaySchemeToString(n int) CipherScheme {
	switch n {
	case 0:
		return CipherSchemeCurve25519
	case 1:
		return CipherSchemeGMSM
	default:
		return ""
	}
}

// autoFillConfig fills any placeholder values on cfg with auto-generated
// content. It loads the sidecar as a base layer (so values persist across
// restarts), generates fresh values only for fields still placeholders,
// and writes the sidecar back if anything changed.
//
// The boolean return is true when the sidecar was rewritten — used by
// main to log "first-boot auto-configured X, Y, Z".
func autoFillConfig(cfg *Config, etcDir string, relayProbeCtx context.Context) (bool, error) {
	values, err := loadAutogen(etcDir)
	if err != nil {
		return false, err
	}
	values.Version = 1
	changed := false

	// KeyEnvelopeKey: AES-256 master key. Generate when placeholder;
	// existing valid 32-byte value from the sidecar (or operator config)
	// is preserved.
	if isPlaceholder(cfg.KeyEnvelopeKey) {
		if values.KeyEnvelopeKey != "" {
			// Sidecar still wins over a placeholder in the operator file.
			cfg.KeyEnvelopeKey = values.KeyEnvelopeKey
		} else {
			k, err := generateRandomKey(32)
			if err != nil {
				return false, fmt.Errorf("generate key envelope: %w", err)
			}
			cfg.KeyEnvelopeKey = k
			values.KeyEnvelopeKey = k
			changed = true
		}
	}

	// SessionKey: 32 random bytes is enough for gin-contrib/sessions/cookie.
	// Treat the placeholder marker (or any short string) as "auto-generate"
	// so a freshly-extracted config.toml Just Works.
	if isPlaceholder(cfg.SessionKey) || len(cfg.SessionKey) < 16 {
		if values.SessionKey != "" {
			cfg.SessionKey = values.SessionKey
		} else {
			k, err := generateRandomKey(32)
			if err != nil {
				return false, fmt.Errorf("generate session key: %w", err)
			}
			cfg.SessionKey = k
			values.SessionKey = k
			changed = true
		}
	}

	// ServerPublicKey + CipherScheme: fetch from the relay's /servers only
	// when ServerPublicKey is the placeholder. We do NOT probe when the
	// operator (or a shipped docker config) has already provided a real
	// key — the nhp-server public key is committed in
	// docker/nhp-relay/etc/config.toml alongside this repo and can be
	// mirrored into docker/demoapp/etc/config.toml without a runtime
	// network call. The /servers auto-fetch is a local-dev convenience
	// for someone running `demoapp` outside docker, not a docker boot
	// path.
	//
	// CipherScheme detection piggybacks on the same probe: if we DO end
	// up probing (because ServerPublicKey is unset), adopt the relay's
	// scheme. Otherwise leave the operator's explicit choice alone.
	if isPlaceholder(cfg.ServerPublicKey) {
		snap, err := func() (relayServerSnapshot, error) {
			if values.ServerPublicKey != "" {
				// Sidecar has the key already from a previous boot;
				// reuse the persisted cipher scheme if present.
				return relayServerSnapshot{
					PublicKeyBase64: values.ServerPublicKey,
					CipherScheme:    values.CipherSchemeNum,
				}, nil
			}
			if cfg.RelayURL == "" {
				return relayServerSnapshot{}, errors.New("RelayUrl is empty; cannot probe relay")
			}
			return fetchServerSnapshotFromRelay(relayProbeCtx, cfg.RelayURL)
		}()
		if err != nil {
			// Non-fatal: log and continue with the placeholder.
			// The operator can edit config.toml later.
			fmt.Fprintf(os.Stderr, "[demoapp] could not auto-fill from relay: %v\n", err)
		} else {
			if snap.PublicKeyBase64 != "" {
				cfg.ServerPublicKey = snap.PublicKeyBase64
				values.ServerPublicKey = snap.PublicKeyBase64
				changed = true
			}
			// Adopt the relay's scheme only when the operator left it
			// empty in config.toml. AllowPlaceholder covers both "" and
			// the canonical REPLACE_/CHANGE_ME markers so an operator
			// who copies the auto-fill template still benefits.
			if isPlaceholder(string(cfg.CipherScheme)) {
				if s := relaySchemeToString(snap.CipherScheme); s != "" {
					cfg.CipherScheme = s
					values.CipherSchemeNum = snap.CipherScheme
					changed = true
				}
			}
		}
	}

	if changed {
		if err := saveAutogen(etcDir, values); err != nil {
			return true, fmt.Errorf("persist autogen: %w", err)
		}
	}
	return changed, nil
}

// EtcDirOf returns the directory holding the config file, so autoFillConfig
// can persist its sidecar next to it. We pass the file path (not just the
// dir) so the loader stays free of path-derivation concerns.
func EtcDirOf(configPath string) string {
	return filepath.Dir(configPath)
}
