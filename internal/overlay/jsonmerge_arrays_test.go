package overlay

import (
	"encoding/json"
	"testing"
)

func TestMergeJSONUnionsArrays(t *testing.T) {
	// gentle-ai trae permissions.allow con "g"; el overlay agrega los suyos.
	base := []byte(`{"permissions":{"allow":["g","shared"]}}`)
	overlay := []byte(`{"permissions":{"allow":["shared","aiwf1","aiwf2"]}}`)

	out, err := MergeJSON(base, overlay)
	if err != nil {
		t.Fatalf("MergeJSON: %v", err)
	}
	var m map[string]any
	if err := json.Unmarshal(out, &m); err != nil {
		t.Fatalf("salida inválida: %v", err)
	}
	allow := m["permissions"].(map[string]any)["allow"].([]any)

	got := make([]string, len(allow))
	for i, v := range allow {
		got[i] = v.(string)
	}
	want := []string{"g", "shared", "aiwf1", "aiwf2"} // base primero, dedup "shared", luego nuevos
	if len(got) != len(want) {
		t.Fatalf("allow = %v, want %v", got, want)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Errorf("allow[%d] = %q, want %q (full=%v)", i, got[i], want[i], got)
		}
	}
}

func TestUnionArraysDedups(t *testing.T) {
	got := unionArrays([]any{"a", "b"}, []any{"b", "c", "a"})
	if len(got) != 3 {
		t.Fatalf("union = %v, want 3 elementos", got)
	}
}
