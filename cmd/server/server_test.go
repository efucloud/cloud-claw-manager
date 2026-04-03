package server

import (
	"os"
	"path/filepath"
	"testing"
)

func TestResolveRunNamespaceFromServiceAccountFile(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	namespaceFile := filepath.Join(dir, "namespace")
	if err := os.WriteFile(namespaceFile, []byte("openclaw-system\n"), 0o600); err != nil {
		t.Fatalf("write namespace file failed: %v", err)
	}

	ns := resolveRunNamespaceWithNamespaceFile(namespaceFile)
	if ns != "openclaw-system" {
		t.Fatalf("expected namespace openclaw-system, got %q", ns)
	}
}

func TestResolveRunNamespaceDefaultOpenClaw(t *testing.T) {
	t.Parallel()
	got := resolveRunNamespaceWithNamespaceFile(filepath.Join(t.TempDir(), "missing"))
	if got != "openclaw" {
		t.Fatalf("expected openclaw, got %q", got)
	}
}
