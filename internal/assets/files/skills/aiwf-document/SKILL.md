---
name: aiwf-document
description: Deterministic project documentation — zero tokens, AI-first context pack.
---
# Skill: /aiwf-document [full|update]

Document a project deterministically without spending tokens, producing a local
second brain for the AI. Designed for legacy or undocumented projects.

## When to use (automatic)

- **Starting work on a project without** `.claude/knowledge/context-pack.md` —
  run BEFORE manually exploring the repo (P8: code extracts, not the agent scanning).
- When the user asks to "document the project", "what does this do", "map the architecture".
- After an architecture or flow change: `aiwf document update` (records evolution).

## Behavior

1. **P8 gate**: if a fresh `context-pack.md` already exists, read it instead of re-scanning.
2. Run `aiwf document full` (first time) or `aiwf document update` (subsequent).
   - **Deterministic, zero-token extraction**: stack, structure, dependencies, entrypoints,
     external connections (redacted), CodeGraph if available, git metadata.
   - Writes LOCAL and git-excluded: `context-pack.md` (AI-first index), `ARCHITECTURE.md`
     (human-readable detail), raw report in `.ai-workflow/evidence/document/`.
3. **Read the generated `context-pack.md`** and incorporate into context — it is the curated
   index that replaces scanning the repo (token savings by design).
4. **Curate** `ARCHITECTURE.md`: deterministic analysis does not infer intent or decisions.
   Complete "Decisions and rationale" with what you know or ask the user.
5. If the user wants synthesized prose: `aiwf document -Synthesize` (opt-in) — **routes to
   OmniRoute** (combo `agent-daily`), never local AI.

## Evolution control

`aiwf document update` archives the previous `ARCHITECTURE.md` in `.claude/knowledge/history/`
and links `supersedes`. Version history lives **locally** (the repo belongs to the client,
not versioned in git). Keeping it updated after architecture/flow changes is part of closing
the work.

## Constraints

- NOTHING is versioned in git. Everything goes to `.claude/` and `.ai-workflow/`, already excluded by `init`.
- No local AI (no Ollama). Any LLM step goes through OmniRoute.
- The report never includes secret values — only counts and redacted patterns.
