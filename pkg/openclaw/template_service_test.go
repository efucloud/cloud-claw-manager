package openclaw

import (
	"encoding/json"
	"testing"

	"k8s.io/apimachinery/pkg/runtime"
)

func TestInjectOpenClawOwnershipLabels_WithTemplateRef(t *testing.T) {
	in := []runtime.RawExtension{
		{
			Raw: []byte(`{"apiVersion":"v1","kind":"ConfigMap","metadata":{"name":"demo","labels":{"keep":"yes"}}}`),
		},
	}

	out, err := InjectOpenClawOwnershipLabels(in, "user-a", "openclaw-abc123", "tpl-demo")
	if err != nil {
		t.Fatalf("InjectOpenClawOwnershipLabels returned error: %v", err)
	}
	if len(out) != 1 {
		t.Fatalf("expected 1 resource, got %d", len(out))
	}

	obj := map[string]interface{}{}
	if err = json.Unmarshal(out[0].Raw, &obj); err != nil {
		t.Fatalf("unmarshal output failed: %v", err)
	}
	metadata, ok := obj["metadata"].(map[string]interface{})
	if !ok {
		t.Fatalf("metadata missing or invalid type")
	}
	labels, ok := metadata["labels"].(map[string]interface{})
	if !ok {
		t.Fatalf("labels missing or invalid type")
	}
	if got := labels["keep"]; got != "yes" {
		t.Fatalf("expected keep=yes, got %v", got)
	}
	if got := labels[openClawResourceLabelOwner]; got != "user-a" {
		t.Fatalf("expected owner label user-a, got %v", got)
	}
	if got := labels[openClawResourceLabelInstance]; got != "openclaw-abc123" {
		t.Fatalf("expected instance label openclaw-abc123, got %v", got)
	}
	if got := labels[OpenClawInstanceLabelTemplateRef]; got != "tpl-demo" {
		t.Fatalf("expected template-ref label tpl-demo, got %v", got)
	}
}

func TestInjectOpenClawOwnershipLabels_WithoutTemplateRef(t *testing.T) {
	in := []runtime.RawExtension{
		{
			Raw: []byte(`{"apiVersion":"v1","kind":"ConfigMap","metadata":{"name":"demo"}}`),
		},
	}

	out, err := InjectOpenClawOwnershipLabels(in, "user-a", "openclaw-abc123", "")
	if err != nil {
		t.Fatalf("InjectOpenClawOwnershipLabels returned error: %v", err)
	}
	if len(out) != 1 {
		t.Fatalf("expected 1 resource, got %d", len(out))
	}

	obj := map[string]interface{}{}
	if err = json.Unmarshal(out[0].Raw, &obj); err != nil {
		t.Fatalf("unmarshal output failed: %v", err)
	}
	metadata, ok := obj["metadata"].(map[string]interface{})
	if !ok {
		t.Fatalf("metadata missing or invalid type")
	}
	labels, ok := metadata["labels"].(map[string]interface{})
	if !ok {
		t.Fatalf("labels missing or invalid type")
	}
	if _, exists := labels[OpenClawInstanceLabelTemplateRef]; exists {
		t.Fatalf("template-ref label should not exist when templateRef is empty")
	}
}
