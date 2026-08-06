package state

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadMissing(t *testing.T) {
	_, ok, err := Load(t.TempDir())
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if ok {
		t.Error("esperaba ok=false cuando no hay estado")
	}
}

func TestLoadPresent(t *testing.T) {
	dir := t.TempDir()
	body := `
schemaVersion: 1
subproject: demo
change: add-auth
status: active
phase: apply
knowledgeLevel: K2
`
	if err := os.WriteFile(filepath.Join(dir, FileName), []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}

	state, ok, err := Load(dir)
	if err != nil || !ok {
		t.Fatalf("Load ok=%v err=%v", ok, err)
	}
	if state.Subproject != "demo" || state.Change != "add-auth" || state.Phase != "apply" {
		t.Fatalf("state = %+v", state)
	}
}

func TestLoadRejectsInvalidState(t *testing.T) {
	dir := t.TempDir()
	body := `schemaVersion: 1
subproject: ../bad
change: add-auth
status: active
phase: apply
knowledgeLevel: K2
`
	if err := os.WriteFile(filepath.Join(dir, FileName), []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
	if _, _, err := Load(dir); err == nil {
		t.Fatal("Load debería rechazar identidad inválida")
	}
}

func TestListChanges(t *testing.T) {
	root := t.TempDir()
	for _, name := range []string{"zeta", "archive", "alpha"} {
		if err := os.Mkdir(filepath.Join(root, name), 0o755); err != nil {
			t.Fatal(err)
		}
	}
	changes, err := ListChanges(root)
	if err != nil {
		t.Fatal(err)
	}
	if len(changes) != 2 || changes[0] != "alpha" || changes[1] != "zeta" {
		t.Fatalf("changes = %v", changes)
	}
}
