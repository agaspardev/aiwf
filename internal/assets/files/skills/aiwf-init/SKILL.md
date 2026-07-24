---
name: aiwf-init
description: Initialize the AI Engineering Workflow in the current project.
---
# Skill: /aiwf-init

Initialize the AI Engineering Workflow in the current project.

## Steps

1. Detect if the current directory is a git repository. If not, ask whether to initialize one.
2. Check for an existing `.ai-workflow/` directory. If present, report already initialized and offer `--reinit` to overwrite.
3. Run `aiwf init` to create the project structure.
4. Confirm what was created:
   - `.ai-workflow/` with subdirectories (state, evidence, handoffs, etc.)
   - `.claude/knowledge/` with ARCHITECTURE.md, DECISIONS.md, CONVENTIONS.md, GOTCHAS.md, LEARNINGS.md
   - `.claude/CLAUDE.md` with containment rules
   - `sonar-project.properties` if SonarQube is enabled
   - `.ai-workflow/` added to `.git/info/exclude` (NOT `.gitignore`)
5. Call `mem_save` recording that the project was initialized.

## Expected result

The project is ready for harness sessions. The user can run `aiwf` to open Claude Code with the active contract.
