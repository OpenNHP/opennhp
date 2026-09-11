package utils

import (
	"crypto/rand"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"runtime/debug"
	"strings"
	"time"

	toml "github.com/pelletier/go-toml/v2"

	"github.com/OpenNHP/opennhp/nhp/log"
)

// GetRandomUint32 returns a uniformly random non-zero uint32 from the system
// CSPRNG. Zero is excluded so callers can use the result directly as an XOR
// mask without a degenerate all-zero preamble.
func GetRandomUint32() uint32 {
	var b [4]byte
	for {
		if _, err := rand.Read(b[:]); err != nil {
			panic(fmt.Sprintf("utils.GetRandomUint32: crypto/rand failed: %v", err))
		}
		if value := binary.BigEndian.Uint32(b[:]); value != 0 {
			return value
		}
	}
}

func CatchPanic() {
	if x := recover(); x != nil {
		for _, line := range append([]string{fmt.Sprint(x)}, strings.Split(string(debug.Stack()), "\n")...) {
			if len(strings.TrimSpace(line)) > 0 {
				log.Error("%s", line)
			}
		}
	}
}

func CatchPanicThenRun(catchFun func()) {
	if x := recover(); x != nil {
		for _, line := range append([]string{fmt.Sprint(x)}, strings.Split(string(debug.Stack()), "\n")...) {
			if len(strings.TrimSpace(line)) > 0 {
				log.Error("%s", line)
			}
		}
		if catchFun != nil {
			catchFun()
		}
	}
}

// Here's how to get the current date string in the format yyyyMMdd (like 20250210) in various programming languages:
func GetCurrentDate() (date string) {
	now := time.Now()
	date = now.Format("20060102")
	return date
}

func DownloadFileToTemp(fileUrl string, pattern string) (string, error) {
	tempDir, err := os.MkdirTemp("", pattern)
	if err != nil {
		return "", err
	}

	fileName := filepath.Base(fileUrl)
	tempFilePath := filepath.Join(tempDir, fileName)

	outFile, err := os.Create(tempFilePath)
	if err != nil {
		return "", err
	}
	defer outFile.Close()

	resp, err := http.Get(fileUrl) //nolint:gosec // G107: URL comes from trusted configuration
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("failed to download file (%s): status code %s", fileUrl, resp.Status)
	}

	_, err = io.Copy(outFile, resp.Body)
	if err != nil {
		return "", err
	}

	return tempFilePath, nil
}

func GenerateTempFilePath(pattern string) (string, error) {
	file, err := os.CreateTemp("", pattern)
	if err != nil {
		return "", err
	}

	tempPath := file.Name()

	if err := file.Close(); err != nil {
		return "", err
	}

	return tempPath, nil
}

func SaveStructAsJsonFile(filePath string, data any) error {
	if data == nil {
		return fmt.Errorf("data cannot be nil")
	}
	if filePath == "" {
		return fmt.Errorf("file path cannot be empty")
	}

	jsonData, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal data to JSON: %w", err)
	}

	err = os.WriteFile(filePath, jsonData, 0644) //nolint:gosec // G306: Generic utility - callers determine sensitivity
	if err != nil {
		return fmt.Errorf("failed to write JSON to file: %w", err)
	}

	return nil
}

func LoadJsonFileAsStruct(filePath string) (any, error) {
	if filePath == "" {
		return nil, fmt.Errorf("file path cannot be empty")
	}

	jsonData, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %w", err)
	}

	var data map[string]any

	if err := json.Unmarshal(jsonData, &data); err != nil {
		return nil, fmt.Errorf("failed to unmarshal JSON %s to struct: %w", string(jsonData), err)
	}

	return data, nil
}

// UpdateTomlConfig sets a top-level string key in a TOML file.
//
// If a `key = "…"` assignment exists (empty value included) it is replaced
// in place. If the key is ABSENT and the file has no `[table]` headers, the
// assignment is appended so a bootstrap/partial config self-heals rather
// than the update silently doing nothing. If the key is absent but the file
// DOES have table headers, appending a bare key at EOF would land it inside
// the last table, so this returns an error instead — the caller must add the
// root-table line by hand.
func UpdateTomlConfig(filePath string, key string, value any) error {
	content, err := os.ReadFile(filePath)
	if err != nil {
		return err
	}

	var newContent string

	switch value := value.(type) {
	case string:
		replacement := fmt.Sprintf("%s = \"%s\"", key, value)
		// Match ONLY a single-line string assignment (`key = "…"`, empty value
		// included), never `key = [` of a multi-line array or a bare number,
		// so a general caller can't mangle those.
		// [ \t\r]*$: (?m) makes $ match immediately before \n, not consume
		// it, and — unlike the \s*$ this replaced — \r is not implied, so a
		// CRLF file needs it spelled out or the regex stops matching a line
		// it used to match on Unix-checked-out content.
		lineRe := regexp.MustCompile(`(?m)^[ \t]*` + regexp.QuoteMeta(key) + `[ \t]*=[ \t]*"[^"]*"[ \t\r]*$`)
		switch {
		case lineRe.MatchString(string(content)):
			// ReplaceAllLiteralString, not ReplaceAllString: the replacement
			// is a verbatim value, and a sealed key blob ("v1$argon2id$...")
			// contains '$' sequences that ReplaceAllString would interpret as
			// capture-group references and mangle.
			newContent = lineRe.ReplaceAllLiteralString(string(content), replacement)
		default:
			// The regex found no single-line "key = \"...\"" assignment. That
			// can mean the key is genuinely absent — or it can mean the
			// existing value just doesn't look like one (a trailing comment,
			// a single-quoted literal, CRLF the regex still doesn't cover,
			// ...). Telling those apart by parsing the file, rather than by
			// the regex missing, matters: append-on-no-match used to fire
			// either way, so a value the regex could not see produced a
			// second "key = ..." line next to the first one, a file go-toml
			// then refuses to parse at all.
			var doc map[string]any
			if uerr := toml.Unmarshal(content, &doc); uerr != nil {
				return fmt.Errorf("key %q (string) not found via pattern match in %s, and the file does not parse as TOML to check for real: %w", key, filePath, uerr)
			}
			if _, exists := doc[key]; exists {
				return fmt.Errorf("key %q already exists in %s in a form UpdateTomlConfig does not rewrite (not a single-line \"...\" assignment); edit it by hand", key, filePath)
			}
			if regexp.MustCompile(`(?m)^\s*\[`).MatchString(string(content)) {
				return fmt.Errorf("key %q (string) not found in %s and the file has [table] sections; add the line under the root table by hand", key, filePath)
			}
			// Key is confirmed absent at the root, and there are no [table]
			// sections it could be silently appended after: safe to append.
			base := string(content)
			if len(base) > 0 && !strings.HasSuffix(base, "\n") {
				base += "\n"
			}
			newContent = base + replacement + "\n"
		}
	default:
		return fmt.Errorf("unsupported type: %T", value)
	}

	return atomicWriteFile(filePath, []byte(newContent))
}

// atomicWriteFile replaces filePath's contents via a same-dir temp file +
// rename, so a crash mid-write cannot leave a truncated file — RotateAgentKey
// now persists a re-SEALED key through this path, and that key exists nowhere
// else. The existing file's permission bits are preserved (a new file gets
// 0600, the CreateTemp default), so a rotation never widens a mode-0600
// config to world-readable; ownership is NOT preserved (Rename does not
// chown), and a rename onto a path that was a symlink replaces the symlink
// with a regular file rather than writing through it. Falls back to a
// non-atomic in-place write when the directory is not writable (a hardened
// root-owned etc/) — that fallback is logged, since it silently drops the
// crash-safety this function otherwise provides, on the one file that may
// hold the only copy of a freshly re-sealed key.
func atomicWriteFile(filePath string, data []byte) error {
	mode := os.FileMode(0o600)
	if fi, statErr := os.Stat(filePath); statErr == nil {
		mode = fi.Mode().Perm() // keep whatever the operator set
	}

	dir := filepath.Dir(filePath)
	tmp, err := os.CreateTemp(dir, ".toml-*")
	if err != nil {
		if os.IsPermission(err) {
			// Can't create a sibling temp file (root-owned etc/). Fall back to
			// a plain in-place write, which only needs +w on the file itself;
			// WriteFile's perm arg is ignored for an existing file, so the
			// mode is preserved here too.
			log.Warning("atomicWriteFile: %s directory is not writable; falling back to a "+
				"non-atomic in-place write of %s (a crash mid-write can now truncate it)", dir, filePath)
			return os.WriteFile(filePath, data, mode)
		}
		return err
	}
	tmpName := tmp.Name()
	defer os.Remove(tmpName) // no-op after a successful rename

	if _, err = tmp.Write(data); err != nil {
		tmp.Close()
		return err
	}
	if err = tmp.Sync(); err != nil {
		tmp.Close()
		return err
	}
	if err = tmp.Close(); err != nil {
		return err
	}
	if err = os.Chmod(tmpName, mode); err != nil {
		return err
	}
	if err = os.Rename(tmpName, filePath); err != nil {
		return err
	}
	// fsync the directory so the rename itself survives a crash, not just the
	// temp file's contents. Best-effort: some filesystems disallow it.
	if d, dErr := os.Open(dir); dErr == nil {
		_ = d.Sync()
		_ = d.Close()
	}
	return nil
}
