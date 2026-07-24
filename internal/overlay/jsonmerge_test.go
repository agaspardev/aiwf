package overlay

import (
	"encoding/json"
	"testing"
)

func TestDeepMergePreservesBaseAndOverrides(t *testing.T) {
	base := map[string]any{
		"a": float64(1),
		"nested": map[string]any{
			"keep":     "base",
			"override": "base",
		},
	}
	overlay := map[string]any{
		"b": float64(2),
		"nested": map[string]any{
			"override": "aiwf",
			"new":      true,
		},
	}
	got := DeepMerge(base, overlay)

	if got["a"] != float64(1) || got["b"] != float64(2) {
		t.Errorf("escalares mal fusionados: %+v", got)
	}
	nested := got["nested"].(map[string]any)
	if nested["keep"] != "base" {
		t.Error("se perdió una clave base en el nested")
	}
	if nested["override"] != "aiwf" {
		t.Error("overlay debería ganar en conflicto")
	}
	if nested["new"] != true {
		t.Error("falta la clave nueva del overlay")
	}
}

func TestMergeJSONWithBase(t *testing.T) {
	base := []byte(`{"permissions":{"allow":["gentleTool"]},"model":"base"}`)
	overlay := []byte(`{"permissions":{"deny":["danger"]},"model":"aiwf"}`)
	out, err := MergeJSON(base, overlay)
	if err != nil {
		t.Fatalf("MergeJSON error: %v", err)
	}
	var m map[string]any
	if err := json.Unmarshal(out, &m); err != nil {
		t.Fatalf("salida no es JSON válido: %v", err)
	}
	perms := m["permissions"].(map[string]any)
	if perms["allow"] == nil {
		t.Error("se perdió permissions.allow de la base (gentle-ai)")
	}
	if perms["deny"] == nil {
		t.Error("falta permissions.deny del overlay")
	}
	if m["model"] != "aiwf" {
		t.Error("overlay debería ganar en 'model'")
	}
}

func TestMergeJSONEmptyBase(t *testing.T) {
	out, err := MergeJSON(nil, []byte(`{"x":1}`))
	if err != nil {
		t.Fatalf("MergeJSON con base vacía: %v", err)
	}
	var m map[string]any
	json.Unmarshal(out, &m)
	if m["x"] != float64(1) {
		t.Error("no se aplicó el overlay sobre base vacía")
	}
}

func TestMergeJSONInvalidBaseErrors(t *testing.T) {
	if _, err := MergeJSON([]byte(`no soy json`), []byte(`{"x":1}`)); err == nil {
		t.Error("esperaba error con base JSON inválida (no clobberear archivo de gentle-ai)")
	}
}
