package ztdo

import (
	"bytes"
	"testing"
)

func TestEncodeMetadataLength(t *testing.T) {
	testCases := []struct {
		name            string
		length          int
		continuation    bool
		expectedEncoded [2]byte
		expectedError   bool
	}{
		{
			name:            "length 0",
			length:          0,
			continuation:    false,
			expectedEncoded: [2]byte{0x00, 0x00},
			expectedError:   false,
		},
		{
			name:            "Length 15036 (0x3ABC), with continuation",
			length:          15036,
			continuation:    true,
			expectedEncoded: [2]byte{0xBC, 0xBA},
			expectedError:   false,
		},
		{
			name:            "Length 32767 (max 15-bit), no continuation",
			length:          32767,
			continuation:    false,
			expectedEncoded: [2]byte{0xFF, 0x7F},
			expectedError:   false,
		},
		// Error cases
		{
			name:            "Length too large (32768)",
			length:          32768,
			continuation:    false,
			expectedEncoded: [2]byte{},
			expectedError:   true,
		},
		{
			name:            "Negative length (-1)",
			length:          -1,
			continuation:    false,
			expectedEncoded: [2]byte{},
			expectedError:   true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			encoded, err := encodeMetadataLength(tc.length, tc.continuation)
			if tc.expectedError {
				if err == nil {
					t.Errorf("Expected error, but got none")
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				}
				if !bytes.Equal(encoded[:], tc.expectedEncoded[:]) {
					t.Errorf("Expected encoded length %v, got %v", tc.expectedEncoded, encoded)
				}
			}
		})
	}
}
