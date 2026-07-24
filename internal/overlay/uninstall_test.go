package overlay

import (
	"os"
	"path/filepath"
	"testing"
)

func TestUninstallRevertsOnlyAiwf(t *testing.T) {
	root := t.TempDir()
	claude := filepath.Join(root, "CLAUDE.md")
	owned := filepath.Join(root, "skills", "aiwf", "x.md")

	// gentle-ai base en el archivo compartido.
	must(t, os.WriteFile(claude, []byte("base gentle-ai\n"), 0o644))

	mp := manifestPath(root)
	o := New(root, mp, []Entry{
		{Path: "CLAUDE.md", Type: MarkerBlock, Payload: []byte("@CLAUDE.aiwf.md")},
		{Path: "skills/aiwf/x.md", Type: Owned, Payload: []byte("skill propia")},
	})
	must(t, o.Apply())

	// Precondición: ambos aplicados.
	assertContains(t, claude, "@CLAUDE.aiwf.md")
	if _, err := os.Stat(owned); err != nil {
		t.Fatalf("owned no se creó: %v", err)
	}

	// Uninstall.
	skipped, err := Uninstall(root, mp)
	must(t, err)
	if len(skipped) != 0 {
		t.Errorf("no debería haber JSON saltado: %v", skipped)
	}

	// El bloque aiwf se fue, la base de gentle-ai queda.
	assertContains(t, claude, "base gentle-ai")
	if HasBlock(readFileOrEmpty(claude)) {
		t.Error("el bloque aiwf no fue removido")
	}
	// El archivo owned fue borrado.
	if _, err := os.Stat(owned); !os.IsNotExist(err) {
		t.Error("el archivo owned no fue borrado")
	}
	// El manifiesto fue borrado.
	if _, err := os.Stat(mp); !os.IsNotExist(err) {
		t.Error("el manifiesto no fue borrado")
	}
}

func TestUninstallDeshaceJSONMerge(t *testing.T) {
	root := t.TempDir()
	mp := manifestPath(root)
	must(t, os.WriteFile(filepath.Join(root, "settings.json"), []byte(`{"a":1}`), 0o644))
	o := New(root, mp, []Entry{
		{Path: "settings.json", Type: JSONMerge, Payload: []byte(`{"b":2,"c":3}`)},
	})
	must(t, o.Apply())

	// Confirmar que se fusionó.
	assertContains(t, filepath.Join(root, "settings.json"), `"b": 2`)

	// Uninstall.
	skipped, err := Uninstall(root, mp)
	must(t, err)
	if len(skipped) != 0 {
		t.Errorf("no debería haber JSON saltado con AddedKeys, got %v", skipped)
	}

	// La key añadida debe haberse ido; la original sobrevive.
	got, err := os.ReadFile(filepath.Join(root, "settings.json"))
	must(t, err)
	if string(got) != "{\n  \"a\": 1\n}" {
		t.Errorf("contenido tras uninstall = %q, want {\"a\":1}", string(got))
	}
}

func TestUninstallSaltaJSONMergeSinAddedKeys(t *testing.T) {
	root := t.TempDir()
	mp := manifestPath(root)
	must(t, os.WriteFile(filepath.Join(root, "settings.json"), []byte(`{"a":1}`), 0o644))

	// Simular manifiesto antiguo: registrar JSONMerge sin AddedKeys.
	must(t, os.MkdirAll(filepath.Dir(mp), 0o755))
	must(t, os.WriteFile(mp, []byte(
		`{"records":[{"path":"settings.json","type":2,"checksum":"x","applied_at":"2026-01-01T00:00:00Z"}]}`,
	), 0o644))

	skipped, err := Uninstall(root, mp)
	must(t, err)
	if len(skipped) != 1 || skipped[0] != "settings.json" {
		t.Errorf("debería saltar JSONMerge sin AddedKeys, got %v", skipped)
	}
	// Contenido intacto.
	got, err := os.ReadFile(filepath.Join(root, "settings.json"))
	must(t, err)
	if string(got) != "{\"a\":1}" {
		t.Errorf("contenido = %q, want {\"a\":1}", string(got))
	}
}
