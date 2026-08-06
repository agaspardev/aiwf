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

func TestSummarizeWithStateYaml(t *testing.T) {
	changesDir := t.TempDir()
	ch := filepath.Join(changesDir, "add-auth")
	if err := os.MkdirAll(ch, 0o755); err != nil {
		t.Fatal(err)
	}
	body := "schemaVersion: 1\nsubproject: demo\nchange: add-auth\nstatus: active\nphase: apply\nknowledgeLevel: K2\n"
	if err := os.WriteFile(filepath.Join(ch, FileName), []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
	sums, err := Summarize(changesDir)
	if err != nil {
		t.Fatalf("Summarize: %v", err)
	}
	if len(sums) != 1 {
		t.Fatalf("esperaba 1 summary, got %d", len(sums))
	}
	s := sums[0]
	if s.Name != "add-auth" || s.Phase != "apply" || s.Status != "active" || s.Inferred {
		t.Fatalf("summary = %+v (esperaba phase=apply status=active inferred=false)", s)
	}
}

func TestSummarizeInfersPhaseFromArtifacts(t *testing.T) {
	changesDir := t.TempDir()
	cases := map[string]struct {
		files []string
		phase string
	}{
		"only-proposal": {[]string{"proposal.md"}, "propose"},
		"up-to-tasks":   {[]string{"exploration.md", "proposal.md", "design.md", "tasks.md"}, "tasks"},
		"archived":      {[]string{"proposal.md", "archive"}, "archive"},
		"empty":         {nil, "unknown"},
	}
	for name, tc := range cases {
		dir := filepath.Join(changesDir, name)
		if err := os.MkdirAll(dir, 0o755); err != nil {
			t.Fatal(err)
		}
		for _, f := range tc.files {
			p := filepath.Join(dir, f)
			if f == "archive" {
				if err := os.MkdirAll(p, 0o755); err != nil {
					t.Fatal(err)
				}
				continue
			}
			if err := os.WriteFile(p, []byte("x"), 0o644); err != nil {
				t.Fatal(err)
			}
		}
	}
	sums, err := Summarize(changesDir)
	if err != nil {
		t.Fatalf("Summarize: %v", err)
	}
	got := map[string]ChangeSummary{}
	for _, s := range sums {
		got[s.Name] = s
	}
	for name, tc := range cases {
		s := got[name]
		if s.Phase != tc.phase {
			t.Errorf("%s: phase=%q esperaba %q", name, s.Phase, tc.phase)
		}
		if !s.Inferred {
			t.Errorf("%s: esperaba Inferred=true", name)
		}
	}
}
