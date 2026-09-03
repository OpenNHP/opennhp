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

// Relay probe loop budget. The constants here are read by both
// fetchServerListFromRelay (which runs the per-probe attempt loop) and
// run() in demoapp_main.go (which builds the parent ctx that bounds it).
//
// relayProbeTotal MUST be at least maxAttempts * (probeTimeout + backoff);
// a shorter budget only lets ~1-2 attempts complete before ctx.Done()
// fires and the loop always exits without ever satisfying
// maxAttempts. autoFillConfig keeps the placeholder key and the SPA is
// handed a REPLACE_… server key it cannot knock against. The +5s safety
// margin absorbs a probe that overruns its 3s budget so the final
// attempt is not truncated.
const (
	probeTimeout    = 3 * time.Second
	backoff         = 1 * time.Second
	maxAttempts     = 30 // ~30s total: long enough for relay to register server
	relayProbeTotal = maxAttempts*(probeTimeout+backoff) + 5*time.Second
)

// fetchServerListFromRelay hits GET <relayURL>/servers and returns the full
// server list the relay reports. Retries with a short backoff so a demoapp
// that boots before the relay has registered any nhp-server eventually sees
// a non-empty list (or times out via ctx).
func fetchServerListFromRelay(ctx context.Context, relayURL string) ([]relayServerInfo, error) {
	if relayURL == "" {
		return nil, errors.New("RelayUrl is empty")
	}
	url := strings.TrimRight(relayURL, "/") + "/servers"

	for attempt := 1; attempt <= maxAttempts; attempt++ {
		servers, err := probeRelayListOnce(ctx, url, probeTimeout)
		if err == nil {
			return servers, nil
		}
		// Log every few attempts so an operator tailing docker logs sees
		// the relay-bringing-up dance.
		if attempt == 1 || attempt%5 == 0 {
			fmt.Fprintf(os.Stderr, "[demoapp] relay /servers attempt %d/%d: %v\n", attempt, maxAttempts, err)
		}
		select {
		case <-ctx.Done():
			return nil, fmt.Errorf("relay /servers never returned a server after %d attempts: %w", attempt, ctx.Err())
		case <-time.After(backoff):
		}
	}
	return nil, fmt.Errorf("relay /servers never returned a server after %d attempts", maxAttempts)
}

// probeRelayListOnce is the single-shot helper behind fetchServerListFromRelay.
func probeRelayListOnce(ctx context.Context, url string, timeout time.Duration) ([]relayServerInfo, error) {
	probeCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()
	req, err := http.NewRequestWithContext(probeCtx, http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("build request: %w", err)
	}
	r, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer r.Body.Close()
	if r.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("HTTP %d", r.StatusCode)
	}
	var servers []relayServerInfo
	if err := json.NewDecoder(r.Body).Decode(&servers); err != nil {
		return nil, fmt.Errorf("decode: %w", err)
	}
	if len(servers) == 0 || servers[0].PublicKeyBase64 == "" {
		return nil, fmt.Errorf("relay has no servers yet")
	}
	return servers, nil
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

// fillServerPubKey sets the public key matching the server's
// RelayRegisteredScheme. The cross-scheme key is left untouched — it cannot
// be derived from the relay's registered key (that requires the server's
// private key), so cross-scheme knock stays unavailable until the operator
// fills the other key by hand. This is a documented limitation.
func fillServerPubKey(se *ServerEntry, pubKeyBase64 string) {
	if se.RelayRegisteredScheme == CipherSchemeGMSM {
		se.Sm2PublicKey = pubKeyBase64
	} else {
		se.Curve25519PublicKey = pubKeyBase64
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

	// Server public keys + scheme: auto-fill from the relay's /servers when
	// the configured servers carry placeholder keys. This is a local-dev
	// convenience for someone running `demoapp` outside docker; the shipped
	// docker configs commit real keys and skip the probe.
	//
	// Single-server case (the promoted-legacy config or one [[Servers]]
	// entry with a placeholder key): probe /servers, fill the entry's
	// relay-registered-scheme public key + set RelayRegisteredScheme. The
	// sidecar caches that one key across restarts.
	//
	// Multi-server case: the operator is expected to hand-fill real keys
	// in [[Servers]] (the relay's /servers exposes only the
	// relay-registered key per server, NOT the cross-scheme key, so we
	// cannot reliably auto-derive a full multi-server catalog). We still
	// probe and fill any placeholder entry we can match by Name.
	defaultSrv := cfg.DefaultServer()
	needProbe := defaultSrv == nil ||
		isPlaceholder(defaultSrv.publicKeyFor(defaultSrv.RelayRegisteredScheme))
	if needProbe {
		servers, err := func() ([]relayServerInfo, error) {
			if values.ServerPublicKey != "" && len(cfg.Servers) <= 1 {
				// Sidecar has the single server's key from a previous boot.
				return []relayServerInfo{{
					PublicKeyBase64: values.ServerPublicKey,
					CipherScheme:    values.CipherSchemeNum,
				}}, nil
			}
			if cfg.RelayURL == "" {
				return nil, errors.New("RelayUrl is empty; cannot probe relay")
			}
			return fetchServerListFromRelay(relayProbeCtx, cfg.RelayURL)
		}()
		if err != nil {
			// Non-fatal: log and continue with the placeholder.
			fmt.Fprintf(os.Stderr, "[demoapp] could not auto-fill from relay: %v\n", err)
		} else if len(servers) > 0 {
			relayed := relaySchemeToString(servers[0].CipherScheme)
			if relayed == "" {
				fmt.Fprintf(os.Stderr, "[demoapp] relay reported unknown cipher scheme %d; not auto-filling\n", servers[0].CipherScheme)
			} else if len(cfg.Servers) == 1 {
				se := &cfg.Servers[0]
				// The configured scheme and the relay's scheme must agree
				// on which public key to use — the relay's routing key is
				// the key the nhp-server registered with the relay under.
				//
				// LoadConfig's legacy single-server promotion writes
				// CipherScheme into RelayRegisteredScheme, so by the time
				// we get here RelayRegisteredScheme is almost never a
				// literal "REPLACE_*" placeholder — the previous version
				// of this branch therefore never fired, and the key ended
				// up persisted into the wrong slot (a GMSM key into
				// Curve25519PublicKey when CipherScheme="curve25519" but
				// the relay registered GMSM). Adopt the relay's scheme
				// whenever the matching key slot is still a placeholder:
				// the operator never filled this key by hand, so the
				// relay is authoritative.
				if relayed != se.RelayRegisteredScheme {
					matchKey := se.publicKeyFor(se.RelayRegisteredScheme)
					if isPlaceholder(matchKey) {
						fmt.Fprintf(os.Stderr,
							"[demoapp] auto-fill: adopting relay scheme %q (was %q) because the configured public key is still a placeholder\n",
							relayed, se.RelayRegisteredScheme)
						se.RelayRegisteredScheme = relayed
					} else {
						// Persist any keys this call already generated
						// before returning the validation error. Without
						// this, a first boot against a relay whose
						// scheme disagrees with a non-placeholder
						// configured pubkey would mint a fresh KeyEnvelopeKey
						// + SessionKey, error out, and on the next boot
						// (after the operator fixes the config) generate
						// DIFFERENT keys — silently rotating the AES-256
						// master that wraps their NHP private key at rest
						// (review follow-up to #2).
						if changed {
							if err := saveAutogen(etcDir, values); err != nil {
								return false, fmt.Errorf("persist autogen before validation error: %w (validation: server %q declares %q but relay %s reports %q)", err, se.Name, se.RelayRegisteredScheme, cfg.RelayURL, relayed)
							}
						}
						return false, fmt.Errorf(
							"server %q declares %q but relay %s reports %q; the relay-registered scheme must match the configured public key. "+
								"Update RelayRegisteredScheme (or CipherScheme, for legacy single-server configs) to %q, or replace the placeholder marker on the configured key",
							se.Name, se.RelayRegisteredScheme, cfg.RelayURL, relayed, relayed)
					}
				}
				if pub := se.publicKeyFor(se.RelayRegisteredScheme); isPlaceholder(pub) {
					fillServerPubKey(se, servers[0].PublicKeyBase64)
					values.ServerPublicKey = servers[0].PublicKeyBase64
					values.CipherSchemeNum = servers[0].CipherScheme
					changed = true
				}
			}
			// Adopt the relay's scheme as the global default when the
			// operator left CipherScheme empty.
			if isPlaceholder(string(cfg.CipherScheme)) && relayed != "" {
				cfg.CipherScheme = relayed
				changed = true
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
