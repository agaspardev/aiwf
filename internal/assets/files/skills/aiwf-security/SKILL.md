---
name: aiwf-security
description: AppSec pipeline — secrets, SAST, SCA and SBOM for the current project.
---
# Skill: /aiwf-security [sast|sca|secrets|sbom|all]

Run the AppSec security pipeline for the current project.

## Usage

```
/aiwf-security           # full pipeline: secrets + sast + sca
/aiwf-security secrets   # Gitleaks only (secrets in git)
/aiwf-security sast      # Semgrep only (static analysis)
/aiwf-security sca       # OSV-Scanner + Trivy only (dependencies)
/aiwf-security sbom      # generate CycloneDX SBOM with Syft
/aiwf-security all       # same as default (secrets + sast + sca)
```

## Behavior

1. Ask for user confirmation before running (the scan may take several minutes).
2. Delegate to `aiwf security <scope>`.
3. Read the generated `scan-summary-*.md` in `.ai-workflow/evidence/security/`.
4. Report:
   - BLOCK findings (block the quality gate)
   - WARN findings (document as technical debt)
   - Missing tools with installation instructions
5. If BLOCK findings exist: indicate that `/quality-gate` will not pass until resolved.

## Automatic checkpoints

This skill is suggested automatically (but never run without confirmation) when:
- Dependencies are added or removed → suggest `sca`
- Files in `auth/`, `crypto/`, `middleware/` are modified → suggest `sast`
- A feature task is completed → suggest `sonar`
- A release or tag is being prepared → suggest `all`
