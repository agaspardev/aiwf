package initproj

import (
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
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

func TestInitCreatesStructure(t *testing.T) {
	root := gitInit(t)
	if _, err := Init(root, "demo", false); err != nil {
		t.Fatalf("Init: %v", err)
	}

	for _, d := range []string{
		".ai-workflow/state", ".ai-workflow/openspec/changes/archive",
		".ai-workflow/evidence/security", ".claude/knowledge",
	} {
		if _, err := os.Stat(filepath.Join(root, filepath.FromSlash(d))); err != nil {
			t.Errorf("falta dir %s", d)
		}
	}

	// Estado válido.
	data, err := os.ReadFile(filepath.Join(root, ".ai-workflow", "state", "workflow-state.json"))
	if err != nil {
		t.Fatalf("estado no creado: %v", err)
	}
	var st map[string]any
	if err := json.Unmarshal(data, &st); err != nil {
		t.Fatalf("estado inválido: %v", err)
	}
	if st["project"] != "demo" || st["phase"] != "INIT" {
		t.Errorf("estado = %v", st)
	}

	// Exclude con los patrones del workflow.
	ex, _ := os.ReadFile(filepath.Join(root, ".git", "info", "exclude"))
	for _, p := range []string{".ai-workflow/", ".claude/", "sonar-project.properties"} {
		if !strings.Contains(string(ex), p) {
			t.Errorf("exclude no contiene %q", p)
		}
	}

	// openspec + sonar props.
	if _, err := os.Stat(filepath.Join(root, ".ai-workflow", "openspec", "config.yaml")); err != nil {
		t.Error("falta openspec/config.yaml")
	}
	if _, err := os.Stat(filepath.Join(root, "sonar-project.properties")); err != nil {
		t.Error("falta sonar-project.properties")
	}
}

func TestInitIdempotent(t *testing.T) {
	root := gitInit(t)
	if _, err := Init(root, "demo", false); err != nil {
		t.Fatal(err)
	}
	rep, err := Init(root, "demo", false)
	if err != nil {
		t.Fatal(err)
	}
	// Segunda corrida: el estado ya existe → skipped.
	foundSkip := false
	for _, s := range rep.Skipped {
		if strings.Contains(s, "workflow-state.json") {
			foundSkip = true
		}
	}
	if !foundSkip {
		t.Error("estado debería reportarse como skipped en la 2da corrida")
	}
	// Exclude no se duplica.
	ex, _ := os.ReadFile(filepath.Join(root, ".git", "info", "exclude"))
	if n := strings.Count(string(ex), ".ai-workflow/"); n != 1 {
		t.Errorf("patrón .ai-workflow/ aparece %d veces en exclude, want 1", n)
	}
}
