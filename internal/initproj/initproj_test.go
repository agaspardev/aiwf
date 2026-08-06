package initproj

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"gopkg.in/yaml.v3"

	"github.com/agaspardev/aiwf/internal/workspace"
)

func gitInit(t *testing.T) string {
	t.Helper()
	root := t.TempDir()
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git no disponible")
	}
	if err := exec.Command("git", "-C", root, "init").Run(); err != nil {
		t.Fatalf("git init: %v", err)
	}
	return root
}

func TestInitCreatesMinimalWorkspace(t *testing.T) {
	root := gitInit(t)
	report, err := Init(root, "demo", false)
	if err != nil {
		t.Fatalf("Init: %v", err)
	}

	manifestPath := filepath.Join(root, ".ai-workflow", "config", "workspace.yaml")
	data, err := os.ReadFile(manifestPath)
	if err != nil {
		t.Fatalf("workspace manifest no creado: %v", err)
	}
	var manifest workspace.WorkspaceManifest
	if err := yaml.Unmarshal(data, &manifest); err != nil {
		t.Fatalf("workspace manifest inválido: %v", err)
	}
	if err := manifest.Validate(); err != nil {
		t.Fatalf("workspace manifest no cumple contrato: %v", err)
	}
	if manifest.RepositoryID != "demo" {
		t.Fatalf("repositoryId = %q, want demo", manifest.RepositoryID)
	}
	if len(report.Created) != 1 || report.Created[0] != ".ai-workflow/config/workspace.yaml" {
		t.Fatalf("created = %v, want solo workspace manifest", report.Created)
	}

	for _, forbidden := range []string{
		".claude/knowledge",
		".claude/CLAUDE.md",
		".ai-workflow/openspec",
		".ai-workflow/state",
		".ai-workflow/evidence",
		".ai-workflow/notes",
		"sonar-project.properties",
	} {
		if _, err := os.Stat(filepath.Join(root, filepath.FromSlash(forbidden))); !os.IsNotExist(err) {
			t.Errorf("init creó ruta legacy %s", forbidden)
		}
	}

	assertNoEmptyDirs(t, filepath.Join(root, ".ai-workflow"))
}

func TestInitIdempotent(t *testing.T) {
	root := gitInit(t)
	if _, err := Init(root, "demo", false); err != nil {
		t.Fatal(err)
	}
	report, err := Init(root, "demo", false)
	if err != nil {
		t.Fatal(err)
	}
	if len(report.Created) != 0 || len(report.Skipped) != 1 {
		t.Fatalf("second init created=%v skipped=%v", report.Created, report.Skipped)
	}

	exclude, err := os.ReadFile(filepath.Join(root, ".git", "info", "exclude"))
	if err != nil {
		t.Fatal(err)
	}
	if count := strings.Count(string(exclude), ".ai-workflow/"); count != 1 {
		t.Fatalf("patrón .ai-workflow/ aparece %d veces, want 1", count)
	}
}

func TestInitProjectCreatesOnlyManifest(t *testing.T) {
	root := gitInit(t)
	report, err := InitProject(root, "aiwf-core", false)
	if err != nil {
		t.Fatalf("InitProject: %v", err)
	}
	path := filepath.Join(root, ".ai-workflow", "projects", "aiwf-core", "project.yaml")
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	var manifest workspace.ProjectManifest
	if err := yaml.Unmarshal(data, &manifest); err != nil {
		t.Fatal(err)
	}
	if err := manifest.Validate(); err != nil {
		t.Fatalf("invalid project manifest: %v", err)
	}
	if manifest.ArtifactStore != "hybrid" {
		t.Fatalf("artifactStore = %q, want hybrid", manifest.ArtifactStore)
	}
	if len(report.Created) != 1 {
		t.Fatalf("created = %v", report.Created)
	}
	assertNoEmptyDirs(t, filepath.Join(root, ".ai-workflow"))
}

func TestInitProjectIsIdempotent(t *testing.T) {
	root := gitInit(t)
	if _, err := InitProject(root, "aiwf-core", false); err != nil {
		t.Fatal(err)
	}
	report, err := InitProject(root, "aiwf-core", false)
	if err != nil {
		t.Fatal(err)
	}
	if len(report.Created) != 0 || len(report.Skipped) != 1 {
		t.Fatalf("created=%v skipped=%v", report.Created, report.Skipped)
	}
}

func TestInitRejectsInvalidRepositoryID(t *testing.T) {
	root := gitInit(t)
	if _, err := Init(root, "../demo", false); err == nil {
		t.Fatal("Init debería rechazar repositoryId inseguro")
	}
	if _, err := os.Stat(filepath.Join(root, ".ai-workflow")); !os.IsNotExist(err) {
		t.Fatal("Init inválido no debe escribir .ai-workflow")
	}
}

func assertNoEmptyDirs(t *testing.T, root string) {
	t.Helper()
	err := filepath.WalkDir(root, func(path string, entry os.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if !entry.IsDir() {
			return nil
		}
		children, err := os.ReadDir(path)
		if err != nil {
			return err
		}
		if len(children) == 0 {
			t.Errorf("directorio vacío creado: %s", path)
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
}

func TestInitReportsGitRepoTrue(t *testing.T) {
	root := t.TempDir()
	if err := os.MkdirAll(filepath.Join(root, ".git"), 0o755); err != nil {
		t.Fatal(err)
	}
	rep, err := Init(root, "demo", false)
	if err != nil {
		t.Fatalf("Init: %v", err)
	}
	if !rep.GitRepo {
		t.Error("esperaba GitRepo=true con .git presente")
	}
	if rep.AlreadyInit {
		t.Error("primera init no debería ser AlreadyInit")
	}
}

func TestInitReportsGitRepoFalseWithWarning(t *testing.T) {
	root := t.TempDir() // sin .git
	rep, err := Init(root, "demo", false)
	if err != nil {
		t.Fatalf("Init: %v", err)
	}
	if rep.GitRepo {
		t.Error("esperaba GitRepo=false sin .git")
	}
	if len(rep.Warnings) == 0 {
		t.Error("esperaba warning cuando falta .git")
	}
}

func TestInitSecondRunReportsAlreadyInit(t *testing.T) {
	root := t.TempDir()
	if _, err := Init(root, "demo", false); err != nil {
		t.Fatal(err)
	}
	rep, err := Init(root, "demo", false)
	if err != nil {
		t.Fatalf("Init 2: %v", err)
	}
	if !rep.AlreadyInit {
		t.Error("segunda init sobre workspace existente debería ser AlreadyInit")
	}
}

func TestInitCreatesNoEmptyDirs(t *testing.T) {
	root := t.TempDir()
	if _, err := Init(root, "demo", false); err != nil {
		t.Fatal(err)
	}
	base := filepath.Join(root, ".ai-workflow")
	err := filepath.WalkDir(base, func(p string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			entries, _ := os.ReadDir(p)
			if len(entries) == 0 {
				t.Errorf("directorio vacío creado por init: %s", p)
			}
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
}
