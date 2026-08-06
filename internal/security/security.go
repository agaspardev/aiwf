// Package security orquesta el pipeline AppSec (gitleaks/semgrep/osv-scanner/trivy/syft),
// portado de security-scan.ps1: ejecuta cada herramienta como subproceso, escribe reportes
// en el evidence/security owner inyectado y produce un summary. Herramientas externas y
// LookPath son inyectables para testear.
package security

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

// Scope define qué se escanea.
type Scope string

const (
	ScopeSecrets Scope = "secrets"
	ScopeSast    Scope = "sast"
	ScopeSca     Scope = "sca"
	ScopeSbom    Scope = "sbom"
	ScopeAll     Scope = "all"
)

// ParseScope valida un scope ("" = all).
func ParseScope(s string) (Scope, error) {
	switch Scope(s) {
	case ScopeSecrets, ScopeSast, ScopeSca, ScopeSbom, ScopeAll:
		return Scope(s), nil
	case "":
		return ScopeAll, nil
	default:
		return "", fmt.Errorf("scope inválido: %q (secrets|sast|sca|sbom|all)", s)
	}
}

// ToolResult es el resultado de correr una herramienta.
type ToolResult struct {
	Tool     string
	Report   string
	ExitCode int
	Status   string // OK | WARN | BLOCK | ERROR | SKIP
	Note     string
}

func (r ToolResult) Blocked() bool { return r.Status == "BLOCK" }
func (r ToolResult) Skipped() bool { return r.Status == "SKIP" }

// Runner ejecuta el pipeline. LookPath y RunCmd son inyectables (tests).
type Runner struct {
	Target      string
	ReportsDir  string
	SummaryPath string
	ts          string
	LookPath    func(string) (string, error)
	RunCmd      func(name string, args ...string) int
	Policy      *Policy // nil = usar DefaultPolicy
}

// policy devuelve el Policy del Runner o DefaultPolicy si no se configuró.
func (r *Runner) policy() *Policy {
	if r.Policy != nil {
		return r.Policy
	}
	return DefaultPolicy()
}

// NewRunner construye un Runner con las dependencias reales.
func NewRunner(target, reportsDir string) *Runner {
	return &Runner{
		Target:     target,
		ReportsDir: reportsDir,
		ts:         time.Now().Format("20060102-150405"),
		LookPath:   exec.LookPath,
		RunCmd:     defaultRunCmd,
	}
}

func defaultRunCmd(name string, args ...string) int {
	cmd := exec.Command(name, args...)
	cmd.Stdout = os.Stderr // el ruido va a stderr; el artefacto es el reporte
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		var ee *exec.ExitError
		if errors.As(err, &ee) {
			return ee.ExitCode()
		}
		return -1
	}
	return 0
}

func (r *Runner) has(tool string) bool {
	_, err := r.LookPath(tool)
	return err == nil
}

func (r *Runner) reportPath(name, ext string) string {
	return filepath.Join(r.ReportsDir, fmt.Sprintf("%s-%s.%s", name, r.ts, ext))
}

func skip(tool, note string) ToolResult {
	return ToolResult{Tool: tool, Status: "SKIP", Note: note}
}

// Secrets — Gitleaks sobre el historial git. Exit != 0 => BLOCK (por defecto policy).
func (r *Runner) Secrets() ToolResult {
	if !r.has("gitleaks") {
		return skip("Gitleaks", "no instalado (P1 — winget install gitleaks)")
	}
	out := r.reportPath("gitleaks", "json")
	code := r.RunCmd("gitleaks", "git", "--redact", "--report-format", "json", "--report-path", out)
	return ToolResult{
		Tool: "Gitleaks", Report: out, ExitCode: code,
		Status: r.policy().ClassifyByExitCode("gitleaks", code),
	}
}

// Sast — Semgrep. 0=OK, 1=WARN (findings), otro=ERROR.
func (r *Runner) Sast() ToolResult {
	if !r.has("semgrep") {
		return skip("Semgrep", "no instalado (P2 — pip install semgrep)")
	}
	out := r.reportPath("semgrep", "json")
	code := r.RunCmd("semgrep", "scan", "--config", "p/default", "--json", "-o", out, r.Target)
	status := r.policy().ClassifyByExitCode("semgrep", code)
	if code != 0 && code != 1 {
		status = "ERROR"
	}
	return ToolResult{Tool: "Semgrep", Report: out, ExitCode: code, Status: status}
}

// Sca — OSV-Scanner + Trivy sobre dependencias/filesystem.
func (r *Runner) Sca() []ToolResult {
	var res []ToolResult
	if !r.has("osv-scanner") {
		res = append(res, skip("OSV-Scanner", "no instalado (P1 — go install github.com/google/osv-scanner/cmd/osv-scanner@latest)"))
	} else {
		out := r.reportPath("osv", "sarif")
		code := r.RunCmd("osv-scanner", "scan", "source", "--recursive", r.Target, "--format", "sarif", "--output-file", out)
		res = append(res, ToolResult{
			Tool: "OSV-Scanner", Report: out, ExitCode: code,
			Status: r.policy().ClassifyByExitCode("osv-scanner", code),
		})
	}
	if !r.has("trivy") {
		res = append(res, skip("Trivy", "no instalado (P2 — winget install Aqua.Trivy)"))
	} else {
		out := r.reportPath("trivy", "json")
		code := r.RunCmd("trivy", "fs", "--scanners", "vuln,misconfig,secret", "--format", "json", "--output", out, r.Target)
		res = append(res, ToolResult{
			Tool: "Trivy", Report: out, ExitCode: code,
			Status: r.policy().ClassifyByExitCode("trivy", code),
		})
	}
	return res
}

// Sbom — Syft (CycloneDX), bajo demanda.
func (r *Runner) Sbom() ToolResult {
	if !r.has("syft") {
		return skip("Syft", "no instalado (winget install anchore.syft)")
	}
	out := r.reportPath("sbom", "json")
	code := r.RunCmd("syft", r.Target, "-o", "cyclonedx-json="+out)
	return ToolResult{Tool: "Syft", Report: out, ExitCode: code, Status: "OK"}
}

// Run ejecuta el scope y escribe el summary.
func (r *Runner) Run(scope Scope) ([]ToolResult, error) {
	if err := os.MkdirAll(r.ReportsDir, 0o755); err != nil {
		return nil, err
	}
	var results []ToolResult
	switch scope {
	case ScopeSecrets:
		results = append(results, r.Secrets())
	case ScopeSast:
		results = append(results, r.Sast())
	case ScopeSca:
		results = append(results, r.Sca()...)
	case ScopeSbom:
		results = append(results, r.Sbom())
	case ScopeAll:
		results = append(results, r.Secrets(), r.Sast())
		results = append(results, r.Sca()...)
	}
	if err := r.writeSummary(scope, results); err != nil {
		return results, err
	}
	return results, nil
}

func (r *Runner) writeSummary(scope Scope, results []ToolResult) error {
	var b strings.Builder
	fmt.Fprintf(&b, "# Security Scan Summary — %s\n- Scope: %s\n- Target: %s\n\n", r.ts, scope, r.Target)
	blocks, missing := 0, 0
	for _, res := range results {
		fmt.Fprintf(&b, "## %s\n- status: %s\n- exit_code: %d\n", res.Tool, res.Status, res.ExitCode)
		if res.Report != "" {
			fmt.Fprintf(&b, "- report: %s\n", res.Report)
		}
		if res.Note != "" {
			fmt.Fprintf(&b, "- note: %s\n", res.Note)
		}
		b.WriteString("\n")
		if res.Blocked() {
			blocks++
		}
		if res.Skipped() {
			missing++
		}
	}
	fmt.Fprintf(&b, "## Resultado\n- Bloqueantes: %d\n- Herramientas faltantes: %d\n", blocks, missing)

	r.SummaryPath = filepath.Join(r.ReportsDir, "scan-summary-"+r.ts+".md")
	return os.WriteFile(r.SummaryPath, []byte(b.String()), 0o644)
}

// ExitCode traduce los resultados a un código de salida: 1 si hay bloqueantes,
// 2 si faltan herramientas, 0 si todo ok.
func ExitCode(results []ToolResult) int {
	blocked, missing := false, false
	for _, r := range results {
		if r.Blocked() {
			blocked = true
		}
		if r.Skipped() {
			missing = true
		}
	}
	switch {
	case blocked:
		return 1
	case missing:
		return 2
	default:
		return 0
	}
}
