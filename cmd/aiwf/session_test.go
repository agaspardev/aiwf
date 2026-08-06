package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/agaspardev/aiwf/internal/workspace"
)

func TestParseRunArgsSubprojectOptional(t *testing.T) {
	sub, dryRun, skipPerms, err := parseRunArgs([]string{"--dry-run"})
	if err != nil {
		t.Fatalf("parseRunArgs sin subproject: %v", err)
	}
	if sub != "" || !dryRun || skipPerms {
		t.Fatalf("sub=%q dryRun=%v skipPerms=%v", sub, dryRun, skipPerms)
	}
}

func TestParseRunArgsWithSubproject(t *testing.T) {
	sub, dryRun, skipPerms, err := parseRunArgs([]string{"aiwf-core", "--dry-run", "--skip-perms"})
	if err != nil {
		t.Fatalf("parseRunArgs: %v", err)
	}
	if sub != "aiwf-core" || !dryRun || !skipPerms {
		t.Fatalf("sub=%q dryRun=%v skipPerms=%v", sub, dryRun, skipPerms)
	}
}

func TestParseRunArgsRejectsTwoPositionals(t *testing.T) {
	if _, _, _, err := parseRunArgs([]string{"a", "b"}); err == nil {
		t.Fatal("expected error for two positionals")
	}
}

func TestResolveSessionSubprojectExisting(t *testing.T) {
	root := t.TempDir()
	mkdir(t, filepath.Join(root, ".ai-workflow", "projects", "aiwf-core"))
	touch(t, filepath.Join(root, ".ai-workflow", "projects", "aiwf-core", "project.yaml"))

	got, err := resolveSessionSubproject(root, "aiwf-core")
	if err != nil {
		t.Fatalf("resolveSessionSubproject: %v", err)
	}
	if got != "aiwf-core" {
		t.Fatalf("got %q", got)
	}
}

func TestResolveSessionSubprojectCreatesAfterConfirm(t *testing.T) {
	root := t.TempDir()
	orig := confirmCreate
	confirmCreate = func() bool { return true }
	defer func() { confirmCreate = orig }()

	got, err := resolveSessionSubproject(root, "nuevo-proj")
	if err != nil {
		t.Fatalf("resolveSessionSubproject: %v", err)
	}
	if got != "nuevo-proj" {
		t.Fatalf("got %q", got)
	}
	if _, err := os.Stat(filepath.Join(root, ".ai-workflow", "projects", "nuevo-proj", "project.yaml")); err != nil {
		t.Fatalf("project manifest no creado: %v", err)
	}
}

func TestResolveSessionSubprojectDeclinesCreate(t *testing.T) {
	root := t.TempDir()
	orig := confirmCreate
	confirmCreate = func() bool { return false }
	defer func() { confirmCreate = orig }()

	if _, err := resolveSessionSubproject(root, "typo-proj"); err == nil {
		t.Fatal("expected error when user declines creation")
	}
	if _, err := os.Stat(filepath.Join(root, ".ai-workflow", "projects", "typo-proj")); err == nil {
		t.Fatal("subproject no debería existir tras declinar")
	}
}

func TestResolveSessionSubprojectNoTTY(t *testing.T) {
	root := t.TempDir()
	orig := isInteractive
	isInteractive = func() bool { return false }
	defer func() { isInteractive = orig }()

	_, err := resolveSessionSubproject(root, "no-existe")
	if err == nil || !strings.Contains(err.Error(), "crealo antes") {
		t.Fatalf("sin TTY debe fallar con hint: %v", err)
	}
}

func TestResolveDefaultSubprojectCreatesBaseAndPersists(t *testing.T) {
	root := t.TempDir()
	orig := isInteractive
	isInteractive = func() bool { return false }
	defer func() { isInteractive = orig }()

	got, err := resolveSessionSubproject(root, "")
	if err != nil {
		t.Fatalf("resolveSessionSubproject: %v", err)
	}
	if got != "base" {
		t.Fatalf("got %q, want base", got)
	}
	if _, err := os.Stat(filepath.Join(root, ".ai-workflow", "projects", "base", "project.yaml")); err != nil {
		t.Fatalf("base no creado: %v", err)
	}
	m, ok, err := loadWorkspaceManifest(root)
	if err != nil || !ok || m.DefaultSubproject != "base" {
		t.Fatalf("default no persistido: ok=%v manifest=%+v err=%v", ok, m, err)
	}
}

func TestResolveDefaultSubprojectUsesPersistedDefault(t *testing.T) {
	root := t.TempDir()
	mkdir(t, filepath.Join(root, ".ai-workflow", "projects", "aiwf-core"))
	touch(t, filepath.Join(root, ".ai-workflow", "projects", "aiwf-core", "project.yaml"))
	manifest := &workspace.WorkspaceManifest{
		SchemaVersion:     1,
		LayoutVersion:     1,
		RepositoryID:      "aiwf",
		DefaultSubproject: "aiwf-core",
	}
	if err := saveWorkspaceManifest(root, manifest); err != nil {
		t.Fatal(err)
	}
	orig := isInteractive
	isInteractive = func() bool { return true }
	defer func() { isInteractive = orig }()

	got, err := resolveSessionSubproject(root, "")
	if err != nil {
		t.Fatalf("resolveSessionSubproject: %v", err)
	}
	if got != "aiwf-core" {
		t.Fatalf("got %q, want persisted default aiwf-core", got)
	}
}

func touch(t *testing.T, path string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(""), 0o644); err != nil {
		t.Fatal(err)
	}
}
