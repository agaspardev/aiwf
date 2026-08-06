package containment

import (
	"os"
	"path/filepath"
	"sort"
	"testing"
)

// write crea un archivo (con sus dirs) bajo root.
func write(t *testing.T, root, rel string) {
	t.Helper()
	p := filepath.Join(root, filepath.FromSlash(rel))
	if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(p, []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}
}

func paths(vs []Violation) []string {
	out := make([]string, len(vs))
	for i, v := range vs {
		out[i] = v.Path
	}
	sort.Strings(out)
	return out
}

func TestScanDetectsWorkArtifactsOutsideAiWorkflow(t *testing.T) {
	root := t.TempDir()
	write(t, root, "scratch/experiment.txt")        // dir de trabajo fuera → violación
	write(t, root, "coverage.out")                  // archivo-artefacto → violación
	write(t, root, "internal/foo/foo.go")           // fuente → OK
	write(t, root, ".ai-workflow/scratch/ok.txt")   // dentro de .ai-workflow → OK
	write(t, root, ".git/config")                   // excluido
	write(t, root, "node_modules/pkg/reports/y.js") // excluido

	vs, err := Scan(root, nil)
	if err != nil {
		t.Fatalf("Scan: %v", err)
	}
	got := paths(vs)
	want := []string{"coverage.out", "scratch"}
	if len(got) != len(want) {
		t.Fatalf("violaciones = %v, want %v", got, want)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("violaciones = %v, want %v", got, want)
		}
	}
}

func TestScanForbiddenDirVariants(t *testing.T) {
	root := t.TempDir()
	for _, d := range []string{"notes", "evidence", "reports", "screenshots", "playwright-report", "coverage", "handoffs"} {
		write(t, root, d+"/f.txt")
	}
	vs, err := Scan(root, nil)
	if err != nil {
		t.Fatalf("Scan: %v", err)
	}
	if len(vs) != 7 {
		t.Fatalf("esperaba 7 violaciones de dir, got %v", paths(vs))
	}
}

func TestScanFilePatterns(t *testing.T) {
	root := t.TempDir()
	write(t, root, "a.log")
	write(t, root, "sub/b.cov")
	write(t, root, "lcov.info")
	write(t, root, "keep.go")
	vs, err := Scan(root, nil)
	if err != nil {
		t.Fatalf("Scan: %v", err)
	}
	got := paths(vs)
	want := []string{"a.log", "lcov.info", "sub/b.cov"}
	if len(got) != 3 || got[0] != want[0] || got[1] != want[1] || got[2] != want[2] {
		t.Fatalf("violaciones = %v, want %v", got, want)
	}
}

func TestScanRespectsAllowlist(t *testing.T) {
	root := t.TempDir()
	write(t, root, "scratch/x.txt")
	write(t, root, "coverage.out")
	allow := map[string]bool{"scratch": true}
	vs, err := Scan(root, allow)
	if err != nil {
		t.Fatalf("Scan: %v", err)
	}
	got := paths(vs)
	if len(got) != 1 || got[0] != "coverage.out" {
		t.Fatalf("con allowlist esperaba solo coverage.out, got %v", got)
	}
}

func TestScanCleanTreeHasNoViolations(t *testing.T) {
	root := t.TempDir()
	write(t, root, "internal/foo.go")
	write(t, root, "README.md")
	write(t, root, ".ai-workflow/reports/coverage/c.txt")
	vs, err := Scan(root, nil)
	if err != nil {
		t.Fatalf("Scan: %v", err)
	}
	if len(vs) != 0 {
		t.Fatalf("árbol limpio esperaba 0 violaciones, got %v", paths(vs))
	}
}
