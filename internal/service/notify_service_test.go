package service

import (
	"strings"
	"testing"
)

func TestBuildAuditDetail(t *testing.T) {
	t.Parallel()
	raw := BuildAuditDetail(map[string]any{"name": "botA", "enabled": true})
	if !strings.Contains(raw, "botA") {
		t.Fatalf("expected detail contains name")
	}
}
