---
name: aiwf-audit
description: Full project audit covering architecture, quality, security and migration readiness.
---
# Skill: /aiwf-audit

Full project audit: architecture, quality, security and migration readiness.

## When to use

- Starting work on a legacy or undocumented project (to understand actual state)
- Before an important release
- When accumulated technical debt is suspected
- To evaluate whether the project meets distribution criteria

## Behavior

### Phase 1 — Collection (P8: code before agent)

1. **Structure**: `git ls-files | head -50` for quick mapping
2. **Dependencies**: detect lock files and tools (npm, pnpm, go.sum, etc.)
3. **Configuration**: read `sonar-project.properties`, `.security/policies/`
4. **Workflow state**: `${AIWF_CHANGE_ROOT}/state.yaml`

### Phase 2 — Analysis by dimension

- **Architecture**: clear separation of concerns? Hidden coupling? Undocumented decisions?
- **Quality**: test coverage? SonarQube configured? Documented tech debt?
- **Security**: last security scan? Gitleaks pre-commit hook? Dependencies with critical CVEs?
- **Migration readiness**: hardcoded absolute paths? Secrets in repo? Correct `.gitignore`?

### Phase 3 — Report

Structured report with findings per dimension, severity, and recommended actions.
Save to `${AIWF_CHANGE_ROOT}/evidence/audit/`.
