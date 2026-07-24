package skillreg

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestScanAndGenerate(t *testing.T) {
	dir := t.TempDir()
	// SKILL.md con frontmatter name+description
	mk(t, filepath.Join(dir, "branch-pr", "SKILL.md"), "---\nname: branch-pr\ndescription: crea PRs\n---\n# Branch PR\n")
	// aiwf/*.md sin frontmatter → nombre por archivo, trigger por H1
	mk(t, filepath.Join(dir, "aiwf", "aiwf-doctor.md"), "# Doctor\n\nVerifica entorno\n")
	// rules/*.md
	mk(t, filepath.Join(dir, "rules", "typescript.md"), "---\nname: typescript\ntrigger: al tocar TS\n---\n")

	entries, err := Scan(dir)
	if err != nil {
		t.Fatal(err)
	}
	if len(entries) != 3 {
		t.Fatalf("scan = %d entries, want 3: %+v", len(entries), entries)
	}
	byName := map[string]Entry{}
	for _, e := range entries {
		byName[e.Name] = e
	}
	if byName["branch-pr"].Trigger != "crea PRs" {
		t.Errorf("branch-pr trigger = %q", byName["branch-pr"].Trigger)
	}
	if byName["aiwf-doctor"].Trigger != "Doctor" {
		t.Errorf("aiwf-doctor trigger (H1 fallback) = %q", byName["aiwf-doctor"].Trigger)
	}
	if byName["typescript"].Trigger != "al tocar TS" {
		t.Errorf("typescript trigger = %q", byName["typescript"].Trigger)
	}

	out := Generate(entries)
	if !strings.Contains(out, "| branch-pr |") || !strings.Contains(out, "Skills: **3**") {
		t.Errorf("registry mal generado:\n%s", out)
	}
}

func TestLintDetectsDupAndMissingTrigger(t *testing.T) {
	entries := []Entry{
		{Name: "a", Trigger: "t"},
		{Name: "a", Trigger: "t"},
		{Name: "b", Trigger: ""},
	}
	problems := Lint(entries)
	if len(problems) != 2 {
		t.Fatalf("problems = %v, want 2 (dup + sin trigger)", problems)
	}
}

func mk(t *testing.T, path, content string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
}
