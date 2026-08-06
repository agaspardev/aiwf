package workspace

import (
	"path/filepath"
	"testing"
)

func TestNewPaths(t *testing.T) {
	root := filepath.Join("repo")
	paths, err := NewPaths(root, "aiwf-core", "add-auth")
	if err != nil {
		t.Fatalf("NewPaths: %v", err)
	}

	checks := map[string]string{
		"workflow":          paths.Workflow,
		"config":            paths.Config,
		"workspaceManifest": paths.WorkspaceManifest,
		"knowledgeShared":   paths.KnowledgeShared,
		"knowledgeProject":  paths.KnowledgeProject,
		"project":           paths.Project,
		"projectManifest":   paths.ProjectManifest,
		"specs":             paths.Specs,
		"changes":           paths.Changes,
		"change":            paths.Change,
		"evidence":          paths.Evidence,
		"reports":           paths.Reports,
		"handoffs":          paths.Handoffs,
		"notes":             paths.Notes,
		"scratch":           paths.Scratch,
		"migrations":        paths.Migrations,
	}
	want := map[string]string{
		"workflow":          filepath.Join(root, ".ai-workflow"),
		"config":            filepath.Join(root, ".ai-workflow", "config"),
		"workspaceManifest": filepath.Join(root, ".ai-workflow", "config", "workspace.yaml"),
		"knowledgeShared":   filepath.Join(root, ".ai-workflow", "knowledge", "shared"),
		"knowledgeProject":  filepath.Join(root, ".ai-workflow", "knowledge", "projects", "aiwf-core"),
		"project":           filepath.Join(root, ".ai-workflow", "projects", "aiwf-core"),
		"projectManifest":   filepath.Join(root, ".ai-workflow", "projects", "aiwf-core", "project.yaml"),
		"specs":             filepath.Join(root, ".ai-workflow", "projects", "aiwf-core", "specs"),
		"changes":           filepath.Join(root, ".ai-workflow", "projects", "aiwf-core", "changes"),
		"change":            filepath.Join(root, ".ai-workflow", "projects", "aiwf-core", "changes", "add-auth"),
		"evidence":          filepath.Join(root, ".ai-workflow", "projects", "aiwf-core", "changes", "add-auth", "evidence"),
		"reports":           filepath.Join(root, ".ai-workflow", "projects", "aiwf-core", "changes", "add-auth", "reports"),
		"handoffs":          filepath.Join(root, ".ai-workflow", "projects", "aiwf-core", "changes", "add-auth", "handoffs"),
		"notes":             filepath.Join(root, ".ai-workflow", "projects", "aiwf-core", "changes", "add-auth", "notes"),
		"scratch":           filepath.Join(root, ".ai-workflow", "projects", "aiwf-core", "changes", "add-auth", "scratch"),
		"migrations":        filepath.Join(root, ".ai-workflow", "migrations"),
	}

	for name, got := range checks {
		if got != want[name] {
			t.Errorf("%s = %q, want %q", name, got, want[name])
		}
	}
}

func TestNewProjectPathsWithoutChange(t *testing.T) {
	paths, err := NewPaths("repo", "aiwf-core", "")
	if err != nil {
		t.Fatalf("NewPaths: %v", err)
	}
	if paths.Change != "" || paths.Evidence != "" || paths.Reports != "" {
		t.Fatalf("change paths must be empty without change: %+v", paths)
	}
}

func TestNewPathsRejectsInvalidIdentity(t *testing.T) {
	tests := []struct {
		name       string
		subproject string
		change     string
	}{
		{name: "invalid subproject", subproject: "../core", change: "add-auth"},
		{name: "invalid change", subproject: "core", change: "Add Auth"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if _, err := NewPaths("repo", tt.subproject, tt.change); err == nil {
				t.Fatal("NewPaths should reject invalid identity")
			}
		})
	}
}
