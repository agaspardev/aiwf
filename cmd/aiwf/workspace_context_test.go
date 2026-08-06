package main

import (
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/agaspardev/aiwf/internal/workspace"
)

func TestParseScopedArgs(t *testing.T) {
	scope, rest, err := parseScopedArgs([]string{"full", "--subproject", "aiwf-core", "--change", "f0-layout", "--synthesize"})
	if err != nil {
		t.Fatalf("parseScopedArgs: %v", err)
	}
	if scope.Subproject != "aiwf-core" || scope.Change != "f0-layout" {
		t.Fatalf("scope = %+v", scope)
	}
	if len(rest) != 2 || rest[0] != "full" || rest[1] != "--synthesize" {
		t.Fatalf("rest = %v", rest)
	}
}

func TestParseScopedArgsRejectsMissingValue(t *testing.T) {
	if _, _, err := parseScopedArgs([]string{"--change"}); err == nil {
		t.Fatal("expected missing flag value error")
	}
}

func TestResolveScope(t *testing.T) {
	root := t.TempDir()
	mkdir(t, filepath.Join(root, ".ai-workflow", "projects", "aiwf-core", "changes", "change-one"))

	scope, err := resolveScope(root, scopeArgs{}, true)
	if err != nil {
		t.Fatalf("resolveScope: %v", err)
	}
	if scope.Subproject != "aiwf-core" || scope.Change != "change-one" {
		t.Fatalf("scope = %+v", scope)
	}
	if scope.Paths.Change == "" {
		t.Fatal("change path is empty")
	}
}

func TestResolveScopeRejectsAmbiguousChange(t *testing.T) {
	root := t.TempDir()
	for _, change := range []string{"change-one", "change-two"} {
		mkdir(t, filepath.Join(root, ".ai-workflow", "projects", "aiwf-core", "changes", change))
	}
	_, err := resolveScope(root, scopeArgs{Subproject: "aiwf-core"}, true)
	if !errors.Is(err, workspace.ErrIdentityAmbiguous) {
		t.Fatalf("error = %v, want ErrIdentityAmbiguous", err)
	}
}

func TestResolveScopeUsesSessionValues(t *testing.T) {
	root := t.TempDir()
	mkdir(t, filepath.Join(root, ".ai-workflow", "projects", "aiwf-core", "changes", "change-one"))
	t.Setenv("AIWF_SUBPROJECT", "aiwf-core")
	t.Setenv("AIWF_CHANGE", "change-one")

	scope, err := resolveScope(root, scopeArgs{}, true)
	if err != nil {
		t.Fatalf("resolveScope: %v", err)
	}
	if scope.Subproject != "aiwf-core" || scope.Change != "change-one" {
		t.Fatalf("scope = %+v", scope)
	}
}

func mkdir(t *testing.T, path string) {
	t.Helper()
	if err := os.MkdirAll(path, 0o755); err != nil {
		t.Fatal(err)
	}
}
