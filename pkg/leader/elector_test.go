package leader

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/efucloud/cloud-claw-manager/pkg/config"
)

func TestResolveLeaseNamespaceUsesRunNamespace(t *testing.T) {
	original := config.RunNamespace
	defer func() { config.RunNamespace = original }()
	config.RunNamespace = "run-ns"

	got := resolveLeaseNamespaceWithNamespaceFile("manual-ns", filepath.Join(t.TempDir(), "missing"))
	if got != "run-ns" {
		t.Fatalf("expected run-ns, got %q", got)
	}
}

func TestResolveLeaseNamespaceUsesServiceAccountFile(t *testing.T) {
	original := config.RunNamespace
	defer func() { config.RunNamespace = original }()
	config.RunNamespace = ""

	namespaceFile := filepath.Join(t.TempDir(), "namespace")
	if err := os.WriteFile(namespaceFile, []byte("openclaw-system\n"), 0o600); err != nil {
		t.Fatalf("write namespace file failed: %v", err)
	}

	got := resolveLeaseNamespaceWithNamespaceFile("manual-ns", namespaceFile)
	if got != "openclaw-system" {
		t.Fatalf("expected openclaw-system, got %q", got)
	}
	if config.RunNamespace != "openclaw-system" {
		t.Fatalf("expected config.RunNamespace to be openclaw-system, got %q", config.RunNamespace)
	}
}

func TestResolveLeaseNamespaceDefaultsToOpenClaw(t *testing.T) {
	original := config.RunNamespace
	defer func() { config.RunNamespace = original }()
	config.RunNamespace = ""

	got := resolveLeaseNamespaceWithNamespaceFile("manual-ns", filepath.Join(t.TempDir(), "missing"))
	if got != "openclaw" {
		t.Fatalf("expected openclaw, got %q", got)
	}
}
