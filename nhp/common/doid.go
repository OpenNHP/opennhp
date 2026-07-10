package common

import (
	"errors"
	"regexp"
)

// doIDPattern bounds identifiers that are concatenated into ztdo filenames.
// Production DoIds are UUIDs, but the slightly wider allowlist preserves room
// for compatible identifiers without permitting path separators, dots, null
// bytes, Unicode separators, whitespace, or shell metacharacters.
var doIDPattern = regexp.MustCompile(`^[a-zA-Z0-9_-]{1,64}$`)

// ErrInvalidDoID is a fixed sentinel so rejected wire input is not reflected
// into protocol error messages.
var ErrInvalidDoID = errors.New("invalid DoId")

// ValidateDoID validates a DoId before it is used as part of a filesystem path.
func ValidateDoID(doID string) error {
	if !doIDPattern.MatchString(doID) {
		return ErrInvalidDoID
	}
	return nil
}
