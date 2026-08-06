---
name: aiwf-init
description: Initialize the AI Engineering Workflow in the current project.
---
# Skill: /aiwf-init

Initialize the AI Engineering Workflow in the current project.

## Steps

1. Detect if the current directory is a git repository. If not, ask whether to initialize one.
2. Check for `.ai-workflow/config/workspace.yaml`. If it exists, report already initialized; do not overwrite unless the user explicitly passes `--force`.
3. Run `aiwf init` to create the minimal control plane.
4. Confirm exactly what was created:
   - `.ai-workflow/config/workspace.yaml`
   - `.ai-workflow/` in `.git/info/exclude` (NOT `.gitignore`)
5. Confirm that init did NOT create empty project/change/knowledge/evidence directories. Their owning commands create them on demand.
6. Call `mem_save` recording that the workspace was initialized.

## Expected result

The repository control plane is ready. Create a subproject, then launch `aiwf <mode> <subproject>`; `/sdd-new <change>` owns change-specific artifacts.
