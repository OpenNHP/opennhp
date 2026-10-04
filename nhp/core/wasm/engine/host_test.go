package engine

import (
	"testing"
)

func TestGetEvidence(t *testing.T) {
	evidence, scheme, err := GetEvidence()
	if err != nil {
		t.Errorf("GetEvidence() error = %v", err)
		return
	}

	t.Logf("GetEvidence() = %v scheme=%s", evidence, scheme)
}
