package test

import (
	"bytes"
	"testing"

	"github.com/OpenNHP/opennhp/nhp/utils"
)

func TestPKCS7Pad(t *testing.T) {
	tests := []struct {
		name      string
		data      []byte
		blockSize int
		expected  []byte
	}{
		{
			name:      "empty data",
			data:      []byte{},
			blockSize: 16,
			expected:  bytes.Repeat([]byte{16}, 16),
		},
		{
			name:      "one byte",
			data:      []byte{0x01},
			blockSize: 16,
			expected:  append([]byte{0x01}, bytes.Repeat([]byte{15}, 15)...),
		},
		{
			name:      "block size minus one",
			data:      bytes.Repeat([]byte{0xAA}, 15),
			blockSize: 16,
			expected:  append(bytes.Repeat([]byte{0xAA}, 15), []byte{1}...),
		},
		{
			name:      "exact block size",
			data:      bytes.Repeat([]byte{0xBB}, 16),
			blockSize: 16,
			expected:  append(bytes.Repeat([]byte{0xBB}, 16), bytes.Repeat([]byte{16}, 16)...),
		},
		{
			name:      "block size 8",
			data:      []byte{1, 2, 3, 4, 5},
			blockSize: 8,
			expected:  []byte{1, 2, 3, 4, 5, 3, 3, 3},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := utils.PKCS7Pad(tt.data, tt.blockSize)
			if !bytes.Equal(result, tt.expected) {
				t.Errorf("PKCS7Pad() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestPKCS7Unpad(t *testing.T) {
	tests := []struct {
		name      string
		data      []byte
		blockSize int
		expected  []byte
	}{
		{
			name:      "full block padding",
			data:      bytes.Repeat([]byte{16}, 16),
			blockSize: 16,
			expected:  []byte{},
		},
		{
			name:      "one byte padding",
			data:      append(bytes.Repeat([]byte{0xAA}, 15), []byte{1}...),
			blockSize: 16,
			expected:  bytes.Repeat([]byte{0xAA}, 15),
		},
		{
			name:      "standard padding",
			data:      []byte{1, 2, 3, 4, 5, 3, 3, 3},
			blockSize: 8,
			expected:  []byte{1, 2, 3, 4, 5},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := utils.PKCS7Unpad(tt.data, tt.blockSize)
			if !bytes.Equal(result, tt.expected) {
				t.Errorf("PKCS7Unpad() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestPKCS7UnpadInvalid(t *testing.T) {
	tests := []struct {
		name      string
		data      []byte
		blockSize int
	}{
		{
			name:      "empty data",
			data:      []byte{},
			blockSize: 16,
		},
		{
			name:      "padding byte larger than block size",
			data:      []byte{1, 2, 3, 4, 5, 6, 7, 17},
			blockSize: 8,
		},
		{
			name:      "zero padding",
			data:      []byte{1, 2, 3, 4, 5, 6, 7, 0},
			blockSize: 8,
		},
		{
			name:      "inconsistent padding bytes",
			data:      []byte{1, 2, 3, 4, 5, 3, 2, 3},
			blockSize: 8,
		},
		{
			name:      "padding larger than data length",
			data:      []byte{1, 2, 5},
			blockSize: 8,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := utils.PKCS7Unpad(tt.data, tt.blockSize)
			if result != nil {
				t.Errorf("PKCS7Unpad() = %v, want nil for invalid padding", result)
			}
		})
	}
}

func TestPKCS7RoundTrip(t *testing.T) {
	testCases := [][]byte{
		{},
		{1},
		{1, 2, 3, 4, 5},
		bytes.Repeat([]byte{0xAB}, 15),
		bytes.Repeat([]byte{0xCD}, 16),
		bytes.Repeat([]byte{0xEF}, 17),
		bytes.Repeat([]byte{0x12}, 100),
	}

	for _, original := range testCases {
		padded := utils.PKCS7Pad(original, 16)
		if len(padded)%16 != 0 {
			t.Errorf("padded length %d is not multiple of 16", len(padded))
		}

		unpadded := utils.PKCS7Unpad(padded, 16)
		if !bytes.Equal(unpadded, original) {
			t.Errorf("round trip failed: original=%v, got=%v", original, unpadded)
		}
	}
}
