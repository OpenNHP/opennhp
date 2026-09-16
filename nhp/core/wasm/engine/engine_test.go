package engine

import "testing"

func TestLoadFailureAndClose(t *testing.T) {
	e := NewEngine()
	e.Close()
	if err := e.LoadWasm([]byte("invalid wasm")); err == nil {
		t.Fatal("invalid module accepted")
	}
	e.Close()
	// The empty module is valid and lets us verify repeated-load cleanup.
	module := []byte{0, 97, 115, 109, 1, 0, 0, 0}
	if err := e.LoadWasm(module); err != nil {
		t.Fatal(err)
	}
	old := e.mod
	if err := e.LoadWasm(module); err != nil {
		t.Fatal(err)
	}
	if !old.IsClosed() {
		t.Fatal("previous module leaked")
	}
	e.Close()
	e.Close()
}
