// Package initproj inicializa la estructura de trabajo de aiwf en un proyecto
// (portado de init-project.ps1). REGLA DURA: el repo es del cliente; NADA del
// workflow se versiona — todo se excluye por-clon vía .git/info/exclude.
package initproj

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

// Report acumula las acciones realizadas, para el output del CLI.
type Report struct {
	Created  []string
	Skipped  []string
	Warnings []string
}

func (r *Report) created(s string) { r.Created = append(r.Created, s) }
func (r *Report) skipped(s string) { r.Skipped = append(r.Skipped, s) }
func (r *Report) warn(s string)    { r.Warnings = append(r.Warnings, s) }

// Init ejecuta la inicialización idempotente sobre root para el proyecto name.
func Init(root, name string, force bool) (*Report, error) {
	r := &Report{}
	aiDir := filepath.Join(root, ".ai-workflow")
	claudeDir := filepath.Join(root, ".claude")

	if err := step1Dirs(aiDir, r); err != nil {
		return r, err
	}
	if err := step2State(aiDir, name, force, r); err != nil {
		return r, err
	}
	if err := step3Env(aiDir, name, r); err != nil {
		return r, err
	}
	if err := step4VaultConfig(aiDir, name, r); err != nil {
		return r, err
	}
	if err := step5Knowledge(claudeDir, name, force, r); err != nil {
		return r, err
	}
	if err := step6Sonar(root, name, r); err != nil {
		return r, err
	}
	if err := step7GitExclude(root, r); err != nil {
		return r, err
	}
	if err := step9ClaudeMd(root, aiDir, name, r); err != nil {
		return r, err
	}
	if err := step10GitleaksHook(root, r); err != nil {
		return r, err
	}
	if err := step11OpenSpec(aiDir, r); err != nil {
		return r, err
	}
	return r, nil
}

// ── helpers de archivos ─────────────────────────────────────────────────────

func writeIfMissing(path, content string, force bool, r *Report, label string) error {
	if !force {
		if _, err := os.Stat(path); err == nil {
			r.skipped(label)
			return nil
		}
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		return err
	}
	r.created(label)
	return nil
}

// ── pasos ────────────────────────────────────────────────────────────────────

func step1Dirs(aiDir string, r *Report) error {
	dirs := []string{
		"state", "tasks", "handoffs",
		"evidence/security", "evidence/sonar", "evidence/fusion", "evidence/review", "evidence/document",
		"env", "notes", "scratch", "cache",
		"reports/coverage", "reports/lighthouse",
		"schemas",
		"openspec", "openspec/specs", "openspec/changes", "openspec/changes/archive",
	}
	for _, d := range dirs {
		p := filepath.Join(aiDir, filepath.FromSlash(d))
		if _, err := os.Stat(p); err == nil {
			continue
		}
		if err := os.MkdirAll(p, 0o755); err != nil {
			return err
		}
		r.created(".ai-workflow/" + d + "/")
	}
	return nil
}

func step2State(aiDir, name string, force bool, r *Report) error {
	now := time.Now()
	state := map[string]any{
		"schemaVersion":  1,
		"workflowId":     "wf-" + now.Format("20060102") + "-001",
		"project":        name,
		"phase":          "INIT",
		"activeTasks":    []any{},
		"completedTasks": []any{},
		"blockedTasks":   []any{},
		"risks":          []any{},
		"nextAction":     "Ejecutar /sdd-explore o /aiwf-audit para analizar el repositorio",
		"createdAt":      now.Format(time.RFC3339),
		"updatedAt":      now.Format(time.RFC3339),
	}
	path := filepath.Join(aiDir, "state", "workflow-state.json")
	if !force {
		if _, err := os.Stat(path); err == nil {
			r.skipped(".ai-workflow/state/workflow-state.json")
			return nil
		}
	}
	data, err := json.MarshalIndent(state, "", "  ")
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	if err := os.WriteFile(path, data, 0o644); err != nil {
		return err
	}
	r.created(".ai-workflow/state/workflow-state.json")
	return nil
}

func step3Env(aiDir, name string, r *Report) error {
	content := fmt.Sprintf(`# Variables de entorno locales — NUNCA versionar este archivo
# Completar con valores reales antes de usar el workflow

SONAR_TOKEN=
SONAR_HOST_URL=http://localhost:9000
SONAR_PROJECT_KEY=%s

# Plane CE — obtener token en: http://localhost:80/profile/api-tokens/
PLANE_API_TOKEN=your-plane-api-token-here

# API keys adicionales (si aplica)
# ANTHROPIC_API_KEY=
`, name)
	return writeIfMissing(filepath.Join(aiDir, "env", ".env.local"), content, false, r,
		".ai-workflow/env/.env.local (template — completar)")
}

func step4VaultConfig(aiDir, name string, r *Report) error {
	cfg := map[string]any{
		"_comment":         "Overrides personales — no versionar. Completar con rutas reales.",
		"vaultPaths":       map[string]any{"operational": "", "personal": "", "learning": "", "work": ""},
		"restrictedScopes": []string{"work"},
		"sonarqube":        map[string]any{"enabled": false, "projectKey": name},
		"omniroute":        map[string]any{"enabled": true, "url": "http://127.0.0.1:20128"},
	}
	path := filepath.Join(aiDir, "env", "vault-config.local.json")
	if _, err := os.Stat(path); err == nil {
		r.skipped(".ai-workflow/env/vault-config.local.json")
		return nil
	}
	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	if err := os.WriteFile(path, data, 0o644); err != nil {
		return err
	}
	r.created(".ai-workflow/env/vault-config.local.json (template — completar)")
	return nil
}

func step5Knowledge(claudeDir, name string, force bool, r *Report) error {
	today := time.Now().Format("2006-01-02")
	files := map[string]string{
		"ARCHITECTURE.md": fmt.Sprintf("# Arquitectura de %s\n\n_Actualizado: %s_\n\n## Vision general\n\n## Componentes principales\n\n## Decisiones de diseno\n", name, today),
		"DECISIONS.md":    fmt.Sprintf("# Decisiones — %s\n\n_Registro de ADRs y decisiones tecnicas significativas._\n\n| ID | Decision | Estado | Fecha |\n|---|---|---|---|\n", name),
		"CONVENTIONS.md":  fmt.Sprintf("# Convenciones — %s\n\n## Nomenclatura\n\n## Estructura de archivos\n\n## Patrones preferidos\n", name),
		"GOTCHAS.md":      fmt.Sprintf("# Gotchas y trampas — %s\n\n_Comportamientos no obvios que han causado problemas._\n", name),
		"LEARNINGS.md":    fmt.Sprintf("# Aprendizajes de sesion — %s\n\n_Descubrimientos tecnicos relevantes para futuras sesiones._\n", name),
	}
	for fname, content := range files {
		if err := writeIfMissing(filepath.Join(claudeDir, "knowledge", fname), content, force, r,
			".claude/knowledge/"+fname); err != nil {
			return err
		}
	}
	return nil
}

func step6Sonar(root, name string, r *Report) error {
	content := fmt.Sprintf(`sonar.projectKey=%s
sonar.projectName=%s
sonar.sources=.
sonar.exclusions=node_modules/**,.ai-workflow/**,dist/**,coverage/**
sonar.host.url=${env.SONAR_HOST_URL}
sonar.token=${env.SONAR_TOKEN}
`, name, name)
	return writeIfMissing(filepath.Join(root, "sonar-project.properties"), content, false, r,
		"sonar-project.properties")
}

// step7GitExclude añade los patrones del workflow a .git/info/exclude sin tocar
// .gitignore. Idempotente: no duplica líneas ya presentes.
func step7GitExclude(root string, r *Report) error {
	gitDir := filepath.Join(root, ".git")
	if _, err := os.Stat(gitDir); err != nil {
		r.warn("No se detectó .git/. Añadí manualmente a .git/info/exclude: .ai-workflow/, .claude/, sonar-project.properties")
		return nil
	}
	patterns := []string{".ai-workflow/", ".claude/", "sonar-project.properties"}
	const marker = "# AI Workflow (local, no versionar) — aiwf init"

	infoDir := filepath.Join(gitDir, "info")
	if err := os.MkdirAll(infoDir, 0o755); err != nil {
		return err
	}
	excludeFile := filepath.Join(infoDir, "exclude")
	existing, _ := os.ReadFile(excludeFile)

	var missing []string
	for _, p := range patterns {
		if !linePresent(string(existing), p) {
			missing = append(missing, p)
		}
	}
	if len(missing) == 0 {
		r.skipped(".git/info/exclude (ya excluye todo lo del workflow)")
		return nil
	}
	block := "\n" + marker + "\n" + strings.Join(missing, "\n") + "\n"
	f, err := os.OpenFile(excludeFile, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o644)
	if err != nil {
		return err
	}
	defer f.Close()
	if _, err := f.WriteString(block); err != nil {
		return err
	}
	r.created(".git/info/exclude <- " + strings.Join(missing, ", "))
	return nil
}

// linePresent indica si content tiene una línea (trim) igual a pattern.
func linePresent(content, pattern string) bool {
	for _, line := range strings.Split(content, "\n") {
		if strings.TrimSpace(line) == pattern {
			return true
		}
	}
	return false
}

const containmentBlock = `
## Contencion de artefactos locales

REGLA OBLIGATORIA: Todos los archivos generados durante la sesion que no sean
codigo fuente del proyecto deben ir en ` + "`.ai-workflow/`" + `:

- Screenshots, capturas de browser    -> .ai-workflow/reports/playwright/screenshots/
- Reportes de cobertura               -> .ai-workflow/reports/coverage/
- Archivos temporales y experimentos  -> .ai-workflow/scratch/
- Notas de sesion                     -> .ai-workflow/notes/
- Evidencia de scans de seguridad     -> .ai-workflow/evidence/security/

NUNCA crear archivos de trabajo fuera de estas rutas.
P8 (Codigo antes que agente): si el resultado puede obtenerse por codigo, no invocar agente.
`

// step9ClaudeMd añade el bloque de contención a .claude/CLAUDE.md, salvo que el
// cliente ya lo versione (en ese caso va a .ai-workflow/notes/containment.md).
func step9ClaudeMd(root, aiDir, name string, r *Report) error {
	claudeMd := filepath.Join(root, ".claude", "CLAUDE.md")

	if _, err := os.Stat(claudeMd); err == nil {
		if gitTracked(root, ".claude/CLAUDE.md") {
			r.warn(".claude/CLAUDE.md está versionado por el cliente — NO se modifica; regla en .ai-workflow/notes/containment.md")
			return writeIfMissing(filepath.Join(aiDir, "notes", "containment.md"), containmentBlock, true, r,
				".ai-workflow/notes/containment.md")
		}
		existing, _ := os.ReadFile(claudeMd)
		if strings.Contains(string(existing), "Contencion de artefactos") {
			r.skipped(".claude/CLAUDE.md (ya contiene regla de contención)")
			return nil
		}
		f, err := os.OpenFile(claudeMd, os.O_WRONLY|os.O_APPEND, 0o644)
		if err != nil {
			return err
		}
		defer f.Close()
		if _, err := f.WriteString(containmentBlock); err != nil {
			return err
		}
		r.created(".claude/CLAUDE.md <- bloque de contención (local, excluido)")
		return nil
	}
	return writeIfMissing(claudeMd, "# CLAUDE.md — "+name+"\n"+containmentBlock, false, r,
		".claude/CLAUDE.md (creado, local, excluido)")
}

const gitleaksHook = `#!/bin/sh
# pre-commit hook — instalado por aiwf init
if ! command -v gitleaks >/dev/null 2>&1; then
    echo "[pre-commit] gitleaks no instalado — se omite el escaneo de secretos."
    exit 0
fi
echo "[pre-commit] Gitleaks — verificando secretos en staged..."
gitleaks protect --staged --redact || {
    echo "[pre-commit] BLOQUEADO: secretos detectados."
    echo "             Emergencia: git commit --no-verify"
    exit 1
}
`

func step10GitleaksHook(root string, r *Report) error {
	hooksDir := filepath.Join(root, ".git", "hooks")
	if _, err := os.Stat(hooksDir); err != nil {
		r.warn("No se encontró .git/hooks/ — pre-commit hook no instalado.")
		return nil
	}
	if !commandExists("gitleaks") {
		r.warn("Gitleaks no encontrado — pre-commit hook no instalado (winget install gitleaks; re-ejecutar 'aiwf init').")
		return nil
	}
	return writeIfMissing(filepath.Join(hooksDir, "pre-commit"), gitleaksHook, false, r,
		".git/hooks/pre-commit (Gitleaks)")
}

const openspecConfig = `# .ai-workflow/openspec/config.yaml
# NO editar artifact_store — el orquestador SDD lo lee y fuerza hybrid sin preguntar.
version: "1.0"
artifact_store: hybrid

rules:
  specs:
    - "Cada requisito MUST tener al menos un escenario GWT"
    - "Scenarios deben ser testables"
    - "Specs bajo 650 palabras"
  tasks:
    - "IDs inmutables (FEAT-T001). Nunca renombrar"

plane:
  enabled: true
  base_url: "http://localhost:80"
`

func step11OpenSpec(aiDir string, r *Report) error {
	return writeIfMissing(filepath.Join(aiDir, "openspec", "config.yaml"), openspecConfig, false, r,
		".ai-workflow/openspec/config.yaml")
}

// ── utilidades de sistema ────────────────────────────────────────────────────

func commandExists(name string) bool {
	_, err := exec.LookPath(name)
	return err == nil
}

// gitTracked indica si path está versionado en el repo de root.
func gitTracked(root, path string) bool {
	if !commandExists("git") {
		return false
	}
	cmd := exec.Command("git", "-C", root, "ls-files", "--error-unmatch", path)
	return cmd.Run() == nil
}
