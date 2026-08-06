package docgen

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func mk(t *testing.T, path, content string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
}

func TestDetectStack(t *testing.T) {
	root := t.TempDir()
	mk(t, filepath.Join(root, "go.mod"), "module x\n")
	mk(t, filepath.Join(root, "package.json"), "{}")
	got := detectStack(root)
	joined := strings.Join(got, ",")
	if !strings.Contains(joined, "Go") || !strings.Contains(joined, "Node/JS-TS") {
		t.Errorf("detectStack = %v, want Go + Node", got)
	}
}

func TestDetectStackUnknown(t *testing.T) {
	if got := detectStack(t.TempDir()); len(got) != 1 || got[0] != "desconocido" {
		t.Errorf("detectStack vacío = %v, want [desconocido]", got)
	}
}

func TestByExtensionAndTopDirs(t *testing.T) {
	files := []string{
		"cmd/main.go", "internal/a.go", "internal/b.go",
		"README.md", "docs/x.md", "internal/sub/c.go",
	}
	ext := byExtension(files, 15)
	if len(ext) == 0 || ext[0].Ext != ".go" || ext[0].Count != 4 {
		t.Errorf("byExtension top = %+v, want .go=4 primero", ext)
	}
	dirs := topDirs(files, 20)
	if dirs[0].Dir != "internal" || dirs[0].Count != 3 {
		t.Errorf("topDirs top = %+v, want internal=3", dirs)
	}
}

func TestParseDepsGoMod(t *testing.T) {
	root := t.TempDir()
	mk(t, filepath.Join(root, "go.mod"), "module x\n\ngo 1.25\n\nrequire (\n\tgithub.com/foo/bar v1.2.3\n\tgithub.com/baz/qux v0.1.0\n)\n")
	deps := parseDeps(root)
	joined := strings.Join(deps, "|")
	if !strings.Contains(joined, "github.com/foo/bar v1.2.3") {
		t.Errorf("parseDeps go.mod = %v", deps)
	}
}

func TestParseDepsPackageJSON(t *testing.T) {
	root := t.TempDir()
	mk(t, filepath.Join(root, "package.json"), `{"dependencies":{"react":"^18.0.0"},"devDependencies":{"vitest":"^1.0.0"}}`)
	deps := parseDeps(root)
	joined := strings.Join(deps, "|")
	if !strings.Contains(joined, "react@^18.0.0") || !strings.Contains(joined, "vitest@^1.0.0") {
		t.Errorf("parseDeps package.json = %v", deps)
	}
}

func TestGenerateWritesArtifacts(t *testing.T) {
	root := t.TempDir()
	mk(t, filepath.Join(root, "go.mod"), "module demo\n\nrequire github.com/foo/bar v1.0.0\n")
	mk(t, filepath.Join(root, "main.go"), "package main\nfunc main(){}\n")

	knowledge := filepath.Join(root, ".ai-workflow", "knowledge", "projects", "demo")
	evidence := filepath.Join(root, ".ai-workflow", "projects", "demo", "changes", "baseline", "evidence")
	rep, res, err := Generate(root, knowledge, evidence, "full", false)
	if err != nil {
		t.Fatalf("Generate: %v", err)
	}
	if rep.Project != filepath.Base(root) || len(rep.Stack) == 0 {
		t.Errorf("report = %+v", rep)
	}
	for _, p := range []string{res.ReportPath, res.ContextPackPath, res.ArchitecturePath} {
		if _, err := os.Stat(p); err != nil {
			t.Errorf("no se generó %s: %v", p, err)
		}
	}
	if filepath.Dir(res.ContextPackPath) != knowledge || filepath.Dir(res.ArchitecturePath) != knowledge {
		t.Fatalf("knowledge fuera del scope explícito: %+v", res)
	}
	if !strings.HasPrefix(res.ReportPath, evidence+string(filepath.Separator)) {
		t.Fatalf("evidence fuera del owner de change: %s", res.ReportPath)
	}
	if _, err := os.Stat(filepath.Join(root, ".claude", "knowledge")); !os.IsNotExist(err) {
		t.Fatal("docgen no debe crear .claude/knowledge")
	}
	// context-pack menciona el stack.
	cp, _ := os.ReadFile(res.ContextPackPath)
	if !strings.Contains(string(cp), "Context Pack — "+rep.Project) {
		t.Error("context-pack sin título esperado")
	}
	// ARCHITECTURE lista la dependencia.
	arch, _ := os.ReadFile(res.ArchitecturePath)
	if !strings.Contains(string(arch), "github.com/foo/bar v1.0.0") {
		t.Error("ARCHITECTURE.md sin la dependencia esperada")
	}
}

func TestGenerateRequiresExplicitOwners(t *testing.T) {
	root := t.TempDir()
	if _, _, err := Generate(root, "", "", "full", false); err == nil {
		t.Fatal("Generate debe rechazar owners vacíos")
	}
	if _, err := os.Stat(filepath.Join(root, ".ai-workflow")); !os.IsNotExist(err) {
		t.Fatal("Generate inválido no debe escribir")
	}
}

func TestUpdateModeArchivesPrevious(t *testing.T) {
	root := t.TempDir()
	mk(t, filepath.Join(root, "go.mod"), "module demo\n")

	knowledge := filepath.Join(root, ".ai-workflow", "knowledge", "projects", "demo")
	evidence := filepath.Join(root, ".ai-workflow", "projects", "demo", "changes", "baseline", "evidence")
	// Primera generación.
	if _, _, err := Generate(root, knowledge, evidence, "full", false); err != nil {
		t.Fatal(err)
	}
	// Cambiar el proyecto (añadir dep) y correr update → debería archivar la previa.
	mk(t, filepath.Join(root, "package.json"), `{"dependencies":{"react":"^18.0.0"}}`)
	_, res, err := Generate(root, knowledge, evidence, "update", false)
	if err != nil {
		t.Fatal(err)
	}
	if res.ArchivedPrev == "" {
		t.Error("update con cambios debería archivar la versión previa")
	}
	if _, err := os.Stat(res.ArchivedPrev); err != nil {
		t.Errorf("archivo previo no existe: %v", err)
	}
}
