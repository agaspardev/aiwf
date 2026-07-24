package overlay

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func manifestPath(root string) string { return filepath.Join(root, ".aiwf", "manifest.json") }

func TestApplyOwnedWritesFile(t *testing.T) {
	root := t.TempDir()
	o := New(root, manifestPath(root), []Entry{
		{Path: "skills/aiwf-doctor/SKILL.md", Type: Owned, Payload: []byte("skill propia")},
	})
	if err := o.Apply(); err != nil {
		t.Fatalf("Apply: %v", err)
	}
	got, err := os.ReadFile(filepath.Join(root, "skills/aiwf-doctor/SKILL.md"))
	if err != nil {
		t.Fatalf("no se escribió el archivo owned: %v", err)
	}
	if string(got) != "skill propia" {
		t.Errorf("contenido = %q", got)
	}
}

// TestReconcileSurvivesGentleAIUpdate es la garantía central del spec: tras un update
// de gentle-ai que reescribe un archivo compartido, la reconciliación restaura la
// customización SIN perder lo de gentle-ai.
func TestReconcileSurvivesGentleAIUpdate(t *testing.T) {
	root := t.TempDir()
	claude := filepath.Join(root, "CLAUDE.md")
	settings := filepath.Join(root, "settings.json")

	// gentle-ai instala su base.
	must(t, os.WriteFile(claude, []byte("# CLAUDE de gentle-ai\nreglas base\n"), 0o644))
	must(t, os.WriteFile(settings, []byte(`{"model":"base","permissions":{"allow":["g"]}}`), 0o644))

	o := New(root, manifestPath(root), []Entry{
		{Path: "CLAUDE.md", Type: MarkerBlock, Payload: []byte("@CLAUDE.aiwf.md")},
		{Path: "settings.json", Type: JSONMerge, Payload: []byte(`{"model":"aiwf-router"}`)},
	})

	// Primera aplicación.
	must(t, o.Apply())
	assertContains(t, claude, "reglas base")
	assertContains(t, claude, "@CLAUDE.aiwf.md")
	assertContains(t, settings, "aiwf-router")
	assertContains(t, settings, `"allow"`) // no se perdió lo de gentle-ai

	// gentle-ai se ACTUALIZA y reescribe ambos archivos (borra lo nuestro).
	must(t, os.WriteFile(claude, []byte("# CLAUDE de gentle-ai v2\nreglas base v2\n"), 0o644))
	must(t, os.WriteFile(settings, []byte(`{"model":"base2","permissions":{"allow":["g2"]}}`), 0o644))

	// Reconciliación: re-aplica el overlay.
	must(t, o.Reconcile())

	// La customización volvió y lo nuevo de gentle-ai sigue intacto.
	assertContains(t, claude, "reglas base v2")
	assertContains(t, claude, "@CLAUDE.aiwf.md")
	assertContains(t, settings, "aiwf-router")
	assertContains(t, settings, "g2")

	// Idempotencia: reconciliar de nuevo no cambia el resultado.
	before, _ := os.ReadFile(claude)
	must(t, o.Reconcile())
	after, _ := os.ReadFile(claude)
	if string(before) != string(after) {
		t.Error("Reconcile no es idempotente sobre CLAUDE.md")
	}
}

func TestManifestRecordsEntries(t *testing.T) {
	root := t.TempDir()
	mp := manifestPath(root)
	o := New(root, mp, []Entry{
		{Path: "a.md", Type: Owned, Payload: []byte("x")},
		{Path: "CLAUDE.md", Type: MarkerBlock, Payload: []byte("y")},
	})
	must(t, o.Apply())

	m, err := LoadManifest(mp)
	if err != nil {
		t.Fatalf("LoadManifest: %v", err)
	}
	if len(m.Records) != 2 {
		t.Fatalf("manifiesto con %d registros, want 2", len(m.Records))
	}
	// Re-aplicar no duplica registros (upsert por path).
	must(t, o.Apply())
	m2, _ := LoadManifest(mp)
	if len(m2.Records) != 2 {
		t.Errorf("tras re-apply hay %d registros, want 2 (upsert)", len(m2.Records))
	}
}

func TestJSONMergeRastreaAddedKeys(t *testing.T) {
	root := t.TempDir()
	mp := manifestPath(root)
	must(t, os.WriteFile(filepath.Join(root, "settings.json"), []byte(`{"a":1,"b":{"c":2}}`), 0o644))
	o := New(root, mp, []Entry{
		{Path: "settings.json", Type: JSONMerge, Payload: []byte(`{"b":{"d":3},"e":4}`)},
	})
	must(t, o.Apply())

	m, err := LoadManifest(mp)
	must(t, err)
	if len(m.Records) != 1 {
		t.Fatalf("manifiesto: %d records", len(m.Records))
	}
	rec := m.Records[0]
	if rec.Type != JSONMerge {
		t.Fatalf("tipo = %d, want JSONMerge", rec.Type)
	}
	// "e" es nueva, "b" ya existe → solo "e" debe estar en AddedKeys.
	if len(rec.AddedKeys) != 1 || rec.AddedKeys[0] != "e" {
		t.Errorf("AddedKeys = %v, want [e] (b ya existía en base)", rec.AddedKeys)
	}
}

func TestComputeAddedKeys(t *testing.T) {
	tests := []struct {
		name    string
		base    string
		overlay string
		want    []string
	}{
		{"nueva key", `{"a":1}`, `{"b":2}`, []string{"b"}},
		{"key existente", `{"a":1}`, `{"a":2}`, nil},
		{"mixto", `{"a":1}`, `{"a":2,"b":3}`, []string{"b"}},
		{"base vacío", ``, `{"a":1}`, []string{"a"}},
		{"overlay vacío", `{}`, `{}`, nil},
		{"base inválido ok", `no-json`, `{"a":1}`, []string{"a"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := computeAddedKeys([]byte(tt.base), []byte(tt.overlay))
			if len(got) != len(tt.want) {
				t.Fatalf("computeAddedKeys = %v, want %v", got, tt.want)
			}
			for i := range got {
				if got[i] != tt.want[i] {
					t.Fatalf("computeAddedKeys = %v, want %v", got, tt.want)
				}
			}
		})
	}
}

// helpers

func must(t *testing.T, err error) {
	t.Helper()
	if err != nil {
		t.Fatal(err)
	}
}

func assertContains(t *testing.T, path, sub string) {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("leyendo %s: %v", path, err)
	}
	if !strings.Contains(string(data), sub) {
		t.Errorf("%s no contiene %q\n---\n%s", filepath.Base(path), sub, data)
	}
}
