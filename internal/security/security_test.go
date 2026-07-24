package security

import (
	"os"
	"strings"
	"testing"
)

// newTestRunner crea un Runner con herramientas presentes y códigos de salida canned.
func newTestRunner(t *testing.T, present map[string]bool, codes map[string]int) *Runner {
	t.Helper()
	r := NewRunner(".", t.TempDir())
	r.LookPath = func(name string) (string, error) {
		if present[name] {
			return "/usr/bin/" + name, nil
		}
		return "", os.ErrNotExist
	}
	r.RunCmd = func(name string, args ...string) int {
		return codes[name]
	}
	return r
}

func TestSecretsBlocksOnFindings(t *testing.T) {
	r := newTestRunner(t, map[string]bool{"gitleaks": true}, map[string]int{"gitleaks": 1})
	res := r.Secrets()
	if !res.Blocked() {
		t.Errorf("gitleaks exit 1 debería BLOCK, got %s", res.Status)
	}
}

func TestSecretsOK(t *testing.T) {
	r := newTestRunner(t, map[string]bool{"gitleaks": true}, map[string]int{"gitleaks": 0})
	if r.Secrets().Status != "OK" {
		t.Error("gitleaks exit 0 debería OK")
	}
}

func TestSastWarnOnFindings(t *testing.T) {
	r := newTestRunner(t, map[string]bool{"semgrep": true}, map[string]int{"semgrep": 1})
	if r.Sast().Status != "WARN" {
		t.Error("semgrep exit 1 debería WARN")
	}
}

func TestMissingToolSkips(t *testing.T) {
	r := newTestRunner(t, map[string]bool{}, map[string]int{})
	res := r.Secrets()
	if !res.Skipped() || res.Note == "" {
		t.Errorf("gitleaks ausente debería SKIP con note, got %+v", res)
	}
}

func TestRunAllWritesSummaryAndExitCode(t *testing.T) {
	r := newTestRunner(t,
		map[string]bool{"gitleaks": true, "semgrep": true, "osv-scanner": true, "trivy": true},
		map[string]int{"gitleaks": 1, "semgrep": 0, "osv-scanner": 0, "trivy": 0},
	)
	results, err := r.Run(ScopeAll)
	if err != nil {
		t.Fatal(err)
	}
	// gitleaks bloqueó → exit 1.
	if ExitCode(results) != 1 {
		t.Errorf("ExitCode = %d, want 1 (gitleaks BLOCK)", ExitCode(results))
	}
	// Summary escrito y con contenido.
	data, err := os.ReadFile(r.SummaryPath)
	if err != nil {
		t.Fatalf("summary no escrito: %v", err)
	}
	s := string(data)
	for _, want := range []string{"Gitleaks", "Semgrep", "OSV-Scanner", "Trivy", "Bloqueantes: 1"} {
		if !strings.Contains(s, want) {
			t.Errorf("summary no contiene %q", want)
		}
	}
}

func TestExitCodeMissingTools(t *testing.T) {
	results := []ToolResult{{Tool: "X", Status: "SKIP"}, {Tool: "Y", Status: "OK"}}
	if ExitCode(results) != 2 {
		t.Errorf("ExitCode = %d, want 2 (faltan herramientas)", ExitCode(results))
	}
}

func TestParseScope(t *testing.T) {
	if s, _ := ParseScope(""); s != ScopeAll {
		t.Error("scope vacío debería ser all")
	}
	if _, err := ParseScope("bogus"); err == nil {
		t.Error("scope inválido debería error")
	}
}
