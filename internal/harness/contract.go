package harness

import (
	"encoding/json"
	"fmt"
	"strings"
)

// ContractVersion es la versión del contrato (debe seguir a la del workflow).
const ContractVersion = "1.0.7"

// Gates es el contenido de quality-gates.json: gateSet -> lista de gates.
type Gates map[string][]string

// LoadGates parsea quality-gates.json, ignorando claves que no sean listas de strings
// (p. ej. "schemaVersion"), de modo que solo queden los gateSets.
func LoadGates(data []byte) (Gates, error) {
	var raw map[string]any
	if err := json.Unmarshal(data, &raw); err != nil {
		return nil, fmt.Errorf("quality-gates.json inválido: %w", err)
	}
	g := Gates{}
	for k, v := range raw {
		arr, ok := v.([]any)
		if !ok {
			continue // saltea schemaVersion u otros escalares
		}
		list := make([]string, 0, len(arr))
		for _, item := range arr {
			if s, ok := item.(string); ok {
				list = append(list, s)
			}
		}
		g[k] = list
	}
	return g, nil
}

// Resolve devuelve los gates globales seguidos de los del gateSet indicado.
func (g Gates) Resolve(gateSet string) []string {
	out := make([]string, 0)
	out = append(out, g["global"]...)
	out = append(out, g[gateSet]...)
	return out
}

// ContractParams son los datos variables del contract prompt.
type ContractParams struct {
	InstanceRoot string
	ModeName     string
	Mode         Mode
	OmniStatus   string // "ONLINE" | "OFFLINE (using default model)"
	SonarStatus  string
	Gates        []string
	ResumeBlock  string // contexto de handoff opcional (ya formateado)
}

// contractBody es el cuerpo estático del contrato (comportamiento obligatorio,
// fundamentos de ingeniería y stop rules), portado fielmente de harness.ps1.
const contractBody = `MANDATORY BEHAVIOR:
- Classify the task before acting. Announce classification in one line, then proceed.
- Never accept a claim without verifiable evidence (command output, exit code, file path, test result).
- Keep facts, inferences and decisions visibly separated.
- Never run a production build. Never commit or push without explicit user authorization.
- Before any destructive action: stop and request explicit approval.
- For code tasks: run lint, typecheck and tests only — no build.
- Save significant decisions, discoveries and bugfixes to Engram (mem_save).
- Before ending the session: mem_session_summary is MANDATORY.
- Context policy: operate within 20% of context window. At 15%: no new broad investigations.
  At 18%: persist state, write handoff to .ai-workflow/handoffs/, end session.
- P8 (Code before agent): if the result can be obtained by code (cache, grep, git, script),
  do not invoke an agent. Agents are last resort for creative generation only.

ENGINEERING FOUNDATIONS (quality from the START, not deferred to end-of-line judges):
- Before designing or coding, load the applicable rule module from skills/rules/ (router+module,
  progressive disclosure). INDEX.md is the always-loaded router; load a module when its trigger matches.
- Planning/design: apply clean-code-agile + the architecture module (clean-architecture / hexagonal).
  Distinguish business logic vs use cases; define layers, scope and the dependency rule (deps point inward).
- Coding: load the language/framework module for what you touch (typescript / react / angular / barrels /
  algorithms / frontend-fundamentals). Atomic units, functional style, TDD from use cases, no 'any'/'as' to shut errors.
- Judges and quality gates CONFIRM good work; they are not the first line of quality. Build it right first.

DELEGATION STOP RULES (numbers, not vibes — protect the orchestrator's context):
- 4+ files needed to understand something -> delegate an explorer minion; keep only its summary.
- 2+ non-trivial files to write -> delegate a dedicated writer with the full task.
- ~20 tool calls or 5 exploratory reads without delegating -> stop and delegate what remains.
- Review ALWAYS runs in fresh context — never in the thread that wrote the code.
- Reads parallelize freely; writes never parallelize (single writer unless isolated worktrees).
- Register each delegated task by fingerprint (phase+description); never launch the same one twice.`

// BuildContract construye el contract prompt completo, replicando harness.ps1.
func BuildContract(p ContractParams) string {
	var b strings.Builder
	fmt.Fprintf(&b, "AI WORKFLOW CONTRACT (authoritative — version %s)\n", ContractVersion)
	fmt.Fprintf(&b, "Instance: %s\n", p.InstanceRoot)
	fmt.Fprintf(&b, "Mode: %s | Combo: %s | Risk: %s | Gate: %s\n",
		p.ModeName, p.Mode.Combo, p.Mode.Risk, p.Mode.GateSet)
	fmt.Fprintf(&b, "OmniRoute: %s | SonarQube: %s\n\n", p.OmniStatus, p.SonarStatus)

	b.WriteString(contractBody)

	b.WriteString("\n\nMODE DIRECTIVES:\n")
	b.WriteString(bulletList(p.Mode.Directives))

	fmt.Fprintf(&b, "\n\nQUALITY GATES (%s):\n", p.Mode.GateSet)
	b.WriteString(bulletList(p.Gates))

	fmt.Fprintf(&b, "\n\nAIWF | mode: %s | sonar: %s | context-policy: 20pct-max",
		p.ModeName, p.SonarStatus)

	if p.ResumeBlock != "" {
		b.WriteString(p.ResumeBlock)
	}
	return b.String()
}

// bulletList formatea items como "- item" separados por saltos de línea.
func bulletList(items []string) string {
	lines := make([]string, 0, len(items))
	for _, it := range items {
		if strings.TrimSpace(it) == "" {
			continue
		}
		lines = append(lines, "- "+it)
	}
	return strings.Join(lines, "\n")
}
