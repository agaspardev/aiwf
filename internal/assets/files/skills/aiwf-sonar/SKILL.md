---
name: aiwf-sonar
description: SonarQube integration — code quality analysis and quality gate status.
---
# Skill: /aiwf-sonar [changed|full|gate|issues]

SonarQube integration for code quality analysis.

## Usage

```
/aiwf-sonar            # analyze files modified since HEAD (default)
/aiwf-sonar changed    # same as default
/aiwf-sonar full       # full project analysis
/aiwf-sonar gate       # query the current quality gate
/aiwf-sonar issues     # list new BLOCKER/CRITICAL issues
```

## Behavior

1. Delegate to `aiwf sonar <mode>`.
2. If SonarQube is not enabled in configuration, explain how to enable it.
3. After analysis, read the generated summary in `.ai-workflow/evidence/sonar/` and incorporate into context.
4. Report the Quality Gate status (OK / ERROR) with the most critical issues.
5. If the gate is in ERROR: list blockers and suggest `/quality-gate` for full evaluation.

## Integration with quality-gate

When running `/quality-gate`, this skill automatically includes the latest available
`sonar-summary-*.md` from `.ai-workflow/evidence/sonar/` as evidence.
