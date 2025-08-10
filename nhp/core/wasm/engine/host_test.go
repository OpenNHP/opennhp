package engine

import (
	"testing"
)

func TestGetEvidence(t *testing.T) {
	evidence, err := GetEvidence()
	if err != nil {
		t.Errorf("GetEvidence() error = %v", err)
		return
	}

	t.Logf("GetEvidence() = %v", evidence)
}
