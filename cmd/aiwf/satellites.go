package main

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/agaspardev/aiwf/internal/assets"
	"github.com/agaspardev/aiwf/internal/config"
	"github.com/agaspardev/aiwf/internal/containment"
	"github.com/agaspardev/aiwf/internal/docgen"
	"github.com/agaspardev/aiwf/internal/gatekeeper"
	"github.com/agaspardev/aiwf/internal/govuln"
	"github.com/agaspardev/aiwf/internal/omniroute"
	"github.com/agaspardev/aiwf/internal/report"
	"github.com/agaspardev/aiwf/internal/security"
	"github.com/agaspardev/aiwf/internal/skillreg"
	"github.com/agaspardev/aiwf/internal/sonar"
)

// sonarCmd integra SonarQube: gate | issues | changed | full (default gate).
func sonarCmd(args []string) int {
	scopeRequest, args, err := parseScopedArgs(args)
	if err != nil {
		fmt.Fprintf(os.Stderr, "[sonar] error: %v\n", err)
		return 2
	}
	mode := "gate"
	for _, arg := range args {
		if strings.HasPrefix(arg, "-") {
			fmt.Fprintf(os.Stderr, "[sonar] bandera desconocida: %s\n", arg)
			return 2
		}
		mode = arg
		break
	}
	cwd, _ := os.Getwd()
	cfg := sonar.LoadConfig(cwd)
	if !cfg.Enabled {
		fmt.Println("[sonar] SonarQube no está habilitado.")
		fmt.Println("  Habilitar en .ai-workflow/config/local/vault-config.local.json: \"sonarqube\": { \"enabled\": true, ... }")
		return 0
	}

	ctx := context.Background()
	switch mode {
	case "gate":
		if cfg.Token == "" {
			fmt.Fprintln(os.Stderr, "[sonar] falta SONAR_TOKEN (completá .ai-workflow/config/local/env.local)")
			return 1
		}
		status, err := cfg.GateStatus(ctx)
		if err != nil {
			fmt.Fprintf(os.Stderr, "[sonar] error consultando gate: %v\n", err)
			return 1
		}
		fmt.Printf("[sonar] Quality Gate: %s\n", status)
		if status != "OK" {
			return 1
		}
		return 0
	case "issues":
		if cfg.Token == "" {
			fmt.Fprintln(os.Stderr, "[sonar] falta SONAR_TOKEN")
			return 1
		}
		total, issues, err := cfg.Issues(ctx)
		if err != nil {
			fmt.Fprintf(os.Stderr, "[sonar] error consultando issues: %v\n", err)
			return 1
		}
		fmt.Printf("[sonar] Issues BLOCKER/CRITICAL sin resolver: %d\n", total)
		for i, is := range issues {
			if i >= 10 {
				fmt.Printf("  ... y %d más\n", total-10)
				break
			}
			fmt.Printf("  [%s] %s — %s:%d\n", is.Severity, is.Message, is.Component, is.Line)
		}
		return 0
	case "changed", "full":
		resolved, err := resolveScope(cwd, scopeRequest, true)
		if err != nil {
			fmt.Fprintf(os.Stderr, "[sonar] error: %v\n", err)
			return 2
		}
		fmt.Printf("[sonar] análisis %s con sonar-scanner...\n", mode)
		code := cfg.Scan(mode == "full", ".", filepath.Join(resolved.Paths.Evidence, "sonar"))
		return code
	default:
		fmt.Fprintf(os.Stderr, "[sonar] modo desconocido: %q (gate|issues|changed|full)\n", mode)
		return 1
	}
}

// securityCmd corre el pipeline AppSec (secrets|sast|sca|sbom|all).
func securityCmd(args []string) int {
	scopeRequest, args, err := parseScopedArgs(args)
	if err != nil {
		fmt.Fprintf(os.Stderr, "  error: %v\n", err)
		return 2
	}
	scopeStr, target := "", "."
	for i := 0; i < len(args); i++ {
		switch args[i] {
		case "--target":
			if i+1 >= len(args) {
				fmt.Fprintln(os.Stderr, "  error: --target requiere un valor")
				return 2
			}
			target = args[i+1]
			i++
		default:
			if strings.HasPrefix(args[i], "-") {
				fmt.Fprintf(os.Stderr, "  error: bandera desconocida: %s\n", args[i])
				return 2
			}
			if scopeStr != "" {
				fmt.Fprintln(os.Stderr, "  error: solo se admite un scope de security")
				return 2
			}
			scopeStr = args[i]
		}
	}
	scope, err := security.ParseScope(scopeStr)
	if err != nil {
		fmt.Fprintf(os.Stderr, "  %v\n", err)
		return 1
	}

	cwd, _ := os.Getwd()
	resolved, err := resolveScope(cwd, scopeRequest, true)
	if err != nil {
		fmt.Fprintf(os.Stderr, "  error: %v\n", err)
		return 2
	}
	reports := filepath.Join(resolved.Paths.Evidence, "security")
	runner := security.NewRunner(target, reports)

	// Cargar política de clasificación desde assets embebidos.
	if data, err := assets.ReadSecurityPolicy(); err == nil {
		if p, pErr := security.LoadPolicy(data); pErr == nil {
			runner.Policy = p
		}
	}

	fmt.Printf("[security] scan '%s' (target %s)...\n", scope, target)
	results, err := runner.Run(scope)
	if err != nil {
		fmt.Fprintf(os.Stderr, "  error: %v\n", err)
		return 1
	}
	for _, res := range results {
		line := fmt.Sprintf("  [%s] %s", res.Status, res.Tool)
		if res.Note != "" {
			line += " — " + res.Note
		}
		fmt.Println(line)
	}
	fmt.Printf("[security] summary: %s\n", runner.SummaryPath)
	return security.ExitCode(results)
}

// documentCmd genera documentación determinista del proyecto (cero tokens).
func documentCmd(args []string) int {
	scopeRequest, args, err := parseScopedArgs(args)
	if err != nil {
		fmt.Fprintf(os.Stderr, "  error: %v\n", err)
		return 2
	}
	mode, synthesize := "full", false
	for _, a := range args {
		switch a {
		case "update":
			mode = "update"
		case "full":
			mode = "full"
		case "--synthesize", "-s":
			synthesize = true
		}
	}
	root, _ := os.Getwd()
	scope, err := resolveScope(root, scopeRequest, true)
	if err != nil {
		fmt.Fprintf(os.Stderr, "  error: %v\n", err)
		return 2
	}
	fmt.Printf("[document] extracción determinista (%s) — %s/%s\n", mode, scope.Subproject, scope.Change)
	rep, res, err := docgen.Generate(root, scope.Paths.KnowledgeProject, scope.Paths.Evidence, mode, synthesize)
	if err != nil {
		fmt.Fprintf(os.Stderr, "  error: %v\n", err)
		return 1
	}
	fmt.Printf("  stack: %s · %d archivos · %d deps\n", strings.Join(rep.Stack, ", "), rep.FilesTotal, rep.DepsTotal)
	fmt.Printf("  + %s (índice AI-first)\n", res.ContextPackPath)
	fmt.Printf("  + %s\n", res.ArchitecturePath)
	fmt.Printf("  + %s (crudo)\n", res.ReportPath)
	if res.ArchivedPrev != "" {
		fmt.Printf("  + %s (versión previa archivada — evolución)\n", res.ArchivedPrev)
	}
	fmt.Println("  TODO local y excluido de git (el repo es del cliente).")
	return 0
}

// gateCmd valida un phase-contract contra el repo (determinista, cero tokens).
func gateCmd(args []string) int {
	if len(args) == 0 {
		fmt.Fprintln(os.Stderr, "uso: aiwf gate <contract.json> | --containment")
		return 1
	}
	root, _ := os.Getwd()
	if args[0] == "--containment" {
		return containmentGate(root)
	}
	res, err := gatekeeper.Validate(root, args[0])
	if err != nil {
		fmt.Fprintf(os.Stderr, "  error: %v\n", err)
		return 1
	}
	fmt.Printf("[gate] fase %q — validando contra el repo...\n", res.Phase)
	if !res.Passed() {
		fmt.Println("  GATE BLOQUEADO:")
		for _, p := range res.Problems {
			fmt.Printf("  - %s\n", p)
		}
		return 1
	}
	fmt.Printf("  [OK] GATE PASA — %d artefacto(s) verificado(s). Próxima fase: %s\n", res.ArtifactCount, res.NextPhase)
	return 0
}

// containmentGate verifica la contención de artefactos bajo .ai-workflow/ (F2).
func containmentGate(root string) int {
	allow := loadContainmentAllow(filepath.Join(root, "scripts", "containment-allow.txt"))
	violations, err := containment.Scan(root, allow)
	if err != nil {
		fmt.Fprintf(os.Stderr, "  error: %v\n", err)
		return 1
	}
	if len(violations) > 0 {
		fmt.Println("[containment] BLOQUEADO — artefactos de trabajo fuera de .ai-workflow/:")
		for _, v := range violations {
			fmt.Printf("  - %s (%s)\n", v.Path, v.Kind)
		}
		return 1
	}
	fmt.Println("[containment] OK — todos los artefactos de trabajo están bajo .ai-workflow/.")
	return 0
}

// loadContainmentAllow parsea una allowlist (una ruta POSIX por línea, # comentarios).
func loadContainmentAllow(path string) map[string]bool {
	allow := map[string]bool{}
	data, err := os.ReadFile(path)
	if err != nil {
		return allow
	}
	for _, line := range strings.Split(string(data), "\n") {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		allow[line] = true
	}
	return allow
}

// skillsCmd genera el registry de skills o (con --lint) solo verifica drift.
func skillsCmd(args []string) int {
	lint := false
	for _, a := range args {
		if a == "--lint" {
			lint = true
		}
	}
	root, _ := config.InstallRoot()
	skillsDir := filepath.Join(root, "skills")

	entries, err := skillreg.Scan(skillsDir)
	if err != nil {
		fmt.Fprintf(os.Stderr, "  error: %v\n", err)
		return 1
	}
	problems := skillreg.Lint(entries)

	if lint {
		fmt.Printf("[skills] lint de %d skills...\n", len(entries))
		if len(problems) > 0 {
			for _, p := range problems {
				fmt.Printf("  - %s\n", p)
			}
			return 1
		}
		fmt.Println("  [OK] sin drift: sin duplicados, todos con trigger.")
		return 0
	}

	dst := filepath.Join(root, "skill-registry.md")
	if err := os.WriteFile(dst, []byte(skillreg.Generate(entries)), 0o644); err != nil {
		fmt.Fprintf(os.Stderr, "  error escribiendo registry: %v\n", err)
		return 1
	}
	fmt.Printf("[skills] %d skills -> %s\n", len(entries), dst)
	if len(problems) > 0 {
		fmt.Printf("  (%d drift — correr 'aiwf skills --lint' para detalle)\n", len(problems))
	}
	return 0
}

// geminiCmd consulta al modelo auxiliar (Gemini) vía OmniRoute (solo texto).
func geminiCmd(args []string) int {
	if len(args) == 0 {
		fmt.Fprintln(os.Stderr, `uso: aiwf gemini "tu consulta"`)
		return 1
	}
	text, err := omniroute.Consult(context.Background(), strings.Join(args, " "), "gemini-research", 2048)
	if err != nil {
		fmt.Fprintf(os.Stderr, "  error: %v\n", err)
		return 1
	}
	fmt.Println(text)
	return 0
}

// reportCmd agrega reportes SARIF/CycloneDX + govulncheck en ai-report.json y
// ai-summary.md (F1 de ci-parity). Uso: aiwf report --reports <dir> [--out <dir>].
func reportCmd(args []string) int {
	reportsDir, outDir := "", ""
	for i := 0; i < len(args); i++ {
		switch args[i] {
		case "--reports":
			if i+1 >= len(args) {
				fmt.Fprintln(os.Stderr, "error: --reports requiere un directorio")
				return 2
			}
			reportsDir = args[i+1]
			i++
		case "--out":
			if i+1 >= len(args) {
				fmt.Fprintln(os.Stderr, "error: --out requiere un directorio")
				return 2
			}
			outDir = args[i+1]
			i++
		default:
			fmt.Fprintf(os.Stderr, "error: argumento desconocido: %s\n", args[i])
			return 2
		}
	}
	if reportsDir == "" {
		fmt.Fprintln(os.Stderr, "uso: aiwf report --reports <dir> [--out <dir>]")
		return 2
	}
	if outDir == "" {
		outDir = reportsDir
	}

	var findings []report.Finding
	entries, err := os.ReadDir(reportsDir)
	if err != nil {
		fmt.Fprintf(os.Stderr, "  error leyendo reportes: %v\n", err)
		return 1
	}
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		name := e.Name()
		data, readErr := os.ReadFile(filepath.Join(reportsDir, name))
		if readErr != nil {
			continue
		}
		switch {
		case strings.HasSuffix(name, ".sarif"):
			fs, perr := report.ParseSARIF(data)
			if perr != nil {
				fmt.Fprintf(os.Stderr, "  ! %s: %v\n", name, perr)
				continue
			}
			findings = append(findings, fs...)
		case strings.HasPrefix(name, "sbom") && strings.HasSuffix(name, ".json"):
			fs, perr := report.ParseCycloneDX(data)
			if perr != nil {
				fmt.Fprintf(os.Stderr, "  ! %s: %v\n", name, perr)
				continue
			}
			findings = append(findings, fs...)
		}
	}

	// govulncheck (P1): opcional, solo si está en PATH.
	if _, lookErr := exec.LookPath("govulncheck"); lookErr == nil {
		out, _ := exec.Command("govulncheck", "-json", "./...").Output()
		if len(out) > 0 {
			if fs, perr := govuln.ParseGovulncheck(out); perr == nil {
				findings = append(findings, fs...)
			}
		}
	}

	cwd, _ := os.Getwd()
	rep := report.BuildReport(filepath.Base(cwd), findings)
	if err := report.WriteReport(outDir, rep); err != nil {
		fmt.Fprintf(os.Stderr, "  error emitiendo reporte: %v\n", err)
		return 1
	}
	fmt.Printf("[report] %d finding(s) de %d tool(s) -> %s/ai-report.json + ai-summary.md\n",
		len(rep.Findings), len(rep.Tools), outDir)
	return 0
}
