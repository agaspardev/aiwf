---
name: aiwf-init
description: Initialize the AI Engineering Workflow in the current project.
---
# Skill: /aiwf-init

Initialize the AI Engineering Workflow control plane in the current repository.

## How this works (code before agent — P8)

The deterministic orchestration lives in the native command `aiwf init`, NOT in
this prompt. `aiwf init` is idempotent and self-verifying: it detects the git
repo, writes only `.ai-workflow/config/workspace.yaml`, excludes `.ai-workflow/`
via `.git/info/exclude` (never `.gitignore`), creates NO empty directories, and
reports git status + whether the workspace already existed. Do not re-implement
any of that here.

## Steps

1. Run `aiwf init` (add `--force` only if the user explicitly wants to regenerate
   an existing workspace; add `--name <id>` to override the repository id).
2. Relay its structured output verbatim to the user — created files, git-repo
   status, `already initialized` notice, and any warnings (e.g. missing `.git/`).
3. If the output reports no `.git/`, tell the user the workspace is ready but
   `.ai-workflow/` was not auto-excluded; ask whether they want to `git init`
   (do NOT initialize git for them without confirmation).
4. Call `mem_save` recording that the workspace was initialized (this is the only
   step that is genuinely agent-side, not covered by the command).

## Expected result

The repository control plane is ready. Next: `aiwf project new <subproject>`, then
launch a mode; `/sdd-new <change>` owns change-specific artifacts.
