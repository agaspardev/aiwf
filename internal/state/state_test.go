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
	full := filepath.Join(dir, filepath.FromSlash(RelPath))
	if err := os.MkdirAll(filepath.Dir(full), 0o755); err != nil {
		t.Fatal(err)
	}
	body := `{
		"project":"demo","phase":"apply","nextAction":"seguir con L1",
		"activeTasks":[1,2],"completedTasks":[1,2,3],"blockedTasks":[]
	}`
	if err := os.WriteFile(full, []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}

	s, ok, err := Load(dir)
	if err != nil || !ok {
		t.Fatalf("Load ok=%v err=%v", ok, err)
	}
	if s.Project != "demo" || s.Phase != "apply" {
		t.Errorf("campos = %+v", s)
	}
	if len(s.ActiveTasks) != 2 || len(s.CompletedTasks) != 3 || len(s.BlockedTasks) != 0 {
		t.Errorf("conteos = %d/%d/%d, want 2/3/0", len(s.ActiveTasks), len(s.CompletedTasks), len(s.BlockedTasks))
	}
	if s.NextAction != "seguir con L1" {
		t.Errorf("nextAction = %q", s.NextAction)
	}
}
