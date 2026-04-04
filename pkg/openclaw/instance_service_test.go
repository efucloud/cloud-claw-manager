package openclaw

import (
	"encoding/json"
	"strings"
	"testing"

	"k8s.io/apimachinery/pkg/runtime"
)

func TestBuildPreviewEndpoint_NoDuplicateNamespaceInBaseDomain(t *testing.T) {
	t.Parallel()
	got := buildPreviewEndpoint("openclaw", "openclaw-fjmsh2", "openclaw.efucloud.com")
	want := "https://openclaw-fjmsh2-openclaw.openclaw.efucloud.com"
	if got != want {
		t.Fatalf("expected %q, got %q", want, got)
	}
}

func TestBuildPreviewEndpoint_NormalDomain(t *testing.T) {
	t.Parallel()
	got := buildPreviewEndpoint("openclaw", "openclaw-fjmsh2", "efucloud.com")
	want := "https://openclaw-fjmsh2-openclaw.efucloud.com"
	if got != want {
		t.Fatalf("expected %q, got %q", want, got)
	}
}

func TestBuildPreviewEndpoint_KeepNamespaceWhenBaseDomainNotPrefixed(t *testing.T) {
	t.Parallel()
	got := buildPreviewEndpoint("team-a", "openclaw-abc123", "openclaw.efucloud.com")
	want := "https://openclaw-abc123-team-a.openclaw.efucloud.com"
	if got != want {
		t.Fatalf("expected %q, got %q", want, got)
	}
}

func TestInjectIngressPreviewAnnotations(t *testing.T) {
	t.Parallel()
	in := []runtime.RawExtension{
		{
			Raw: []byte(`{"apiVersion":"networking.k8s.io/v1","kind":"Ingress","metadata":{"name":"demo","annotations":{"foo":"bar"}}}`),
		},
	}
	out, err := injectIngressPreviewAnnotations(in)
	if err != nil {
		t.Fatalf("injectIngressPreviewAnnotations returned error: %v", err)
	}
	if len(out) != 1 {
		t.Fatalf("expected 1 resource, got %d", len(out))
	}
	var obj map[string]interface{}
	if err := json.Unmarshal(out[0].Raw, &obj); err != nil {
		t.Fatalf("unmarshal output failed: %v", err)
	}
	metadata, ok := obj["metadata"].(map[string]interface{})
	if !ok {
		t.Fatalf("metadata missing or invalid")
	}
	annotations, ok := metadata["annotations"].(map[string]interface{})
	if !ok {
		t.Fatalf("annotations missing or invalid")
	}
	got := strings.TrimSpace(annotations[openClawIngressHideHeadersAnno].(string))
	if !strings.Contains(got, "X-Frame-Options") || !strings.Contains(got, "Content-Security-Policy") {
		t.Fatalf("proxy-hide-headers missing required headers, got %q", got)
	}
}

func TestMergeCSVAnnotationValue(t *testing.T) {
	t.Parallel()
	got := mergeCSVAnnotationValue("X-Frame-Options", []string{"X-Frame-Options", "Content-Security-Policy"})
	if got != "X-Frame-Options,Content-Security-Policy" {
		t.Fatalf("unexpected merge result: %q", got)
	}
}
