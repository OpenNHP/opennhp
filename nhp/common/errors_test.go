package common

import "testing"

func TestErrorFromResponseKnownCodeUsesCanonicalError(t *testing.T) {
	got := ErrorFromResponse(ErrResourceNotFound.ErrorCode(), "untrusted wire text")
	if got != ErrResourceNotFound {
		t.Fatalf("ErrorFromResponse() = %p, want canonical error %p", got, ErrResourceNotFound)
	}
}

func TestErrorFromResponseUnknownCodePreservesResponse(t *testing.T) {
	const (
		code    = "52142"
		message = "registered agent owner is missing"
	)

	got := ErrorFromResponse(code, message)
	if got == nil {
		t.Fatal("ErrorFromResponse() returned nil")
	}
	if got.ErrorCode() != code {
		t.Fatalf("ErrorCode() = %q, want %q", got.ErrorCode(), code)
	}
	if got.Error() != message {
		t.Fatalf("Error() = %q, want %q", got.Error(), message)
	}
}

func TestErrorFromResponseUnknownCodeWithoutMessageIsUseful(t *testing.T) {
	const code = "59999"

	got := ErrorFromResponse(code, "")
	if got == nil {
		t.Fatal("ErrorFromResponse() returned nil")
	}
	if got.ErrorCode() != code {
		t.Fatalf("ErrorCode() = %q, want %q", got.ErrorCode(), code)
	}
	if got.Error() == "" {
		t.Fatal("Error() is empty")
	}
}
