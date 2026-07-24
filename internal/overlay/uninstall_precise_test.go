package overlay

import (
	"encoding/json"
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

// readJSON lee y parsea un archivo JSON como mapa para aserciones estructurales
// (evita depender del formato/orden del texto).
func readJSON(t *testing.T, path string) map[string]any {
	t.Helper()
	data, err := os.ReadFile(path)
	must(t, err)
	m := map[string]any{}
	must(t, json.Unmarshal(data, &m))
	return m
}

// TestUninstallRestauraArraysDeClavePreexistente cubre el Hallazgo 2: los ítems que
// aiwf UNE dentro de una clave que ya existía (permissions.allow/deny) deben quitarse
// en el uninstall, dejando exactamente lo que tenía gentle-ai.
func TestUninstallRestauraArraysDeClavePreexistente(t *testing.T) {
	root := t.TempDir()
	mp := manifestPath(root)
	settings := filepath.Join(root, "settings.json")

	// gentle-ai ya administra permissions.
	must(t, os.WriteFile(settings, []byte(
		`{"model":"base","permissions":{"allow":["g"],"deny":["gd"]}}`), 0o644))

	o := New(root, mp, []Entry{{
		Path: "settings.json", Type: JSONMerge,
		Payload: []byte(`{"permissions":{"allow":["aiwfA","aiwfB"],"deny":["aiwfD"]},"env":{"AI_WORKFLOW_ACTIVE":"1"}}`),
	}})
	must(t, o.Apply())

	// Precondición: se unieron los arrays y se agregó env.
	after := readJSON(t, settings)
	perm := after["permissions"].(map[string]any)
	if len(perm["allow"].([]any)) != 3 {
		t.Fatalf("allow tras apply = %v, want 3 items", perm["allow"])
	}

	// Uninstall.
	skipped, err := Uninstall(root, mp)
	must(t, err)
	if len(skipped) != 0 {
		t.Errorf("no debería saltar nada, got %v", skipped)
	}

	// El settings.json debe volver EXACTAMENTE a la base de gentle-ai.
	got := readJSON(t, settings)
	want := map[string]any{
		"model": "base",
		"permissions": map[string]any{
			"allow": []any{"g"},
			"deny":  []any{"gd"},
		},
	}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("tras uninstall = %#v\nwant %#v", got, want)
	}
}

// TestUninstallDespuesDeReconcile cubre el Hallazgo 1: tras uno o más reconcile, el
// uninstall SIGUE revirtiendo lo de aiwf (env nueva + arrays unidos), porque el
// snapshot base y AddedKeys se congelan en el primer Apply.
func TestUninstallDespuesDeReconcile(t *testing.T) {
	root := t.TempDir()
	mp := manifestPath(root)
	settings := filepath.Join(root, "settings.json")

	must(t, os.WriteFile(settings, []byte(`{"permissions":{"allow":["g"]}}`), 0o644))

	o := New(root, mp, []Entry{{
		Path: "settings.json", Type: JSONMerge,
		Payload: []byte(`{"permissions":{"allow":["aiwfA"]},"env":{"AI_WORKFLOW_ACTIVE":"1"}}`),
	}})
	must(t, o.Apply())

	// gentle-ai se actualiza y reescribe el archivo; aiwf reconcilia (dos veces).
	must(t, os.WriteFile(settings, []byte(`{"permissions":{"allow":["g2"]}}`), 0o644))
	must(t, o.Reconcile())
	must(t, o.Reconcile())

	// El snapshot base debe seguir siendo el del PRIMER apply, no el archivo ya tocado.
	m, err := LoadManifest(mp)
	must(t, err)
	rec, ok := m.Find("settings.json")
	if !ok || len(rec.BaseSnapshot) == 0 {
		t.Fatalf("BaseSnapshot no se congeló: %+v", rec)
	}

	// Uninstall tras reconcile: env debe irse y allow volver a lo de gentle-ai (g2).
	skipped, err := Uninstall(root, mp)
	must(t, err)
	if len(skipped) != 0 {
		t.Errorf("no debería saltar nada tras reconcile, got %v", skipped)
	}
	got := readJSON(t, settings)
	want := map[string]any{
		"permissions": map[string]any{"allow": []any{"g2"}},
	}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("tras uninstall post-reconcile = %#v\nwant %#v", got, want)
	}
}

// TestUninstallRestauraEscalarSobrescrito verifica que un escalar preexistente que
// aiwf sobrescribió se restaura a su valor base (no queda el valor de aiwf).
func TestUninstallRestauraEscalarSobrescrito(t *testing.T) {
	root := t.TempDir()
	mp := manifestPath(root)
	settings := filepath.Join(root, "settings.json")

	must(t, os.WriteFile(settings, []byte(`{"model":"base"}`), 0o644))
	o := New(root, mp, []Entry{{
		Path: "settings.json", Type: JSONMerge, Payload: []byte(`{"model":"aiwf-router"}`),
	}})
	must(t, o.Apply())
	assertContains(t, settings, "aiwf-router")

	_, err := Uninstall(root, mp)
	must(t, err)

	got := readJSON(t, settings)
	if got["model"] != "base" {
		t.Errorf("model tras uninstall = %v, want \"base\" (restaurado)", got["model"])
	}
}

// TestSubtractAdded valida la resta de arrays de forma aislada.
func TestSubtractAdded(t *testing.T) {
	current := []any{"g", "aiwfA", "aiwfB"}
	overlay := []any{"aiwfA", "aiwfB", "g"} // "g" ya estaba en base → no se quita
	base := []any{"g"}
	got := subtractAdded(current, overlay, base)
	want := []any{"g"}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("subtractAdded = %v, want %v", got, want)
	}
}
