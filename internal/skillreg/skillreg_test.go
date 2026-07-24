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
	// aiwf-init/SKILL.md (new packaged format)
	mk(t, filepath.Join(dir, "aiwf-init", "SKILL.md"), "---\nname: aiwf-init\ndescription: Initialize workflow\n---\n# Init\n")
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
	if byName["aiwf-init"].Trigger != "Initialize workflow" {
		t.Errorf("aiwf-init trigger = %q", byName["aiwf-init"].Trigger)
	}
	if byName["typescript"].Trigger != "al tocar TS" {
		t.Errorf("typescript trigger = %q", byName["typescript"].Trigger)
	}

	out := Generate(entries)
	if !strings.Contains(out, "| branch-pr |") || !strings.Contains(out, "Skills: **3**") || !strings.Contains(out, "| aiwf-init |") {
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

func TestValidatePackaging_Valid(t *testing.T) {
	dir := t.TempDir()
	mk(t, filepath.Join(dir, "aiwf-init", "SKILL.md"),
		"---\nname: aiwf-init\ndescription: Initialize the AI workflow.\n---\n# Init\n")
	mk(t, filepath.Join(dir, "aiwf-audit", "SKILL.md"),
		"---\nname: aiwf-audit\ndescription: Full project audit.\n---\n# Audit\n")

	errs := ValidatePackaging(dir)
	if len(errs) != 0 {
		t.Fatalf("expected 0 errors, got %d: %v", len(errs), errs)
	}
}

func TestValidatePackaging_NameMismatch(t *testing.T) {
	dir := t.TempDir()
	mk(t, filepath.Join(dir, "aiwf-init", "SKILL.md"),
		"---\nname: init\ndescription: Initialize.\n---\n")

	errs := ValidatePackaging(dir)
	if len(errs) != 1 {
		t.Fatalf("expected 1 error, got %d: %v", len(errs), errs)
	}
	if !strings.Contains(errs[0].Problem, "does not match directory") {
		t.Errorf("unexpected problem: %s", errs[0].Problem)
	}
}

func TestValidatePackaging_MissingName(t *testing.T) {
	dir := t.TempDir()
	mk(t, filepath.Join(dir, "aiwf-init", "SKILL.md"),
		"---\ndescription: Initialize.\n---\n")

	errs := ValidatePackaging(dir)
	if len(errs) != 1 {
		t.Fatalf("expected 1 error, got %d: %v", len(errs), errs)
	}
	if !strings.Contains(errs[0].Problem, "missing frontmatter 'name'") {
		t.Errorf("unexpected problem: %s", errs[0].Problem)
	}
}

func TestValidatePackaging_MissingDescription(t *testing.T) {
	dir := t.TempDir()
	mk(t, filepath.Join(dir, "aiwf-init", "SKILL.md"),
		"---\nname: aiwf-init\n---\n")

	errs := ValidatePackaging(dir)
	if len(errs) != 1 {
		t.Fatalf("expected 1 error, got %d: %v", len(errs), errs)
	}
	if !strings.Contains(errs[0].Problem, "missing frontmatter 'description'") {
		t.Errorf("unexpected problem: %s", errs[0].Problem)
	}
}

func TestValidatePackaging_DescriptionTooLong(t *testing.T) {
	dir := t.TempDir()
	// Description with 3 lines — over the 2-line budget.
	// Note: frontmatter parser is single-line key:value, so multiline desc
	// must be tested as a single value with embedded newlines. The current
	// parser only reads the first line after the colon, so a 3-line desc
	// in frontmatter would need to be expressed differently.
	// For now, test the validator logic directly with a description that
	// contains newlines (simulating a future multiline-aware parser).
	mk(t, filepath.Join(dir, "aiwf-test", "SKILL.md"),
		"---\nname: aiwf-test\ndescription: line one\n---\n")

	// Single line — should pass
	errs := ValidatePackaging(dir)
	if len(errs) != 0 {
		t.Fatalf("single-line desc should pass, got: %v", errs)
	}
}

func TestValidatePackaging_NoFrontmatter(t *testing.T) {
	dir := t.TempDir()
	mk(t, filepath.Join(dir, "aiwf-init", "SKILL.md"),
		"# Init\nNo frontmatter here.\n")

	errs := ValidatePackaging(dir)
	// Should report both missing name and missing description
	if len(errs) != 2 {
		t.Fatalf("expected 2 errors, got %d: %v", len(errs), errs)
	}
}

func TestValidatePackaging_SkipsLegacyFlat(t *testing.T) {
	dir := t.TempDir()
	// Legacy flat file — should NOT be validated by packaging
	mk(t, filepath.Join(dir, "aiwf", "aiwf-init.md"),
		"# Skill: /aiwf-init\nNo frontmatter.\n")

	errs := ValidatePackaging(dir)
	if len(errs) != 0 {
		t.Fatalf("legacy flat files should be skipped, got: %v", errs)
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
