---
name: aiwf-opinion
description: Parallel multi-model perspectives on a scoped question. Read-only.
---
# Skill: /aiwf-opinion <question>

Parallel perspectives from multiple models on a scoped question. **Read-only**:
none of the opinions touch the filesystem. For design decisions, "is this approach
solid?", second reads before committing.

Applies the **fusion contract**.

## When to use

- Before a design decision with non-obvious tradeoffs.
- When you want a second (or third) read of an approach, not someone to implement it.
- When the cost of being wrong exceeds the cost of N× tokens.

**Do not use** for what a deterministic check can answer (P8), or for trivial tasks.

## Combos (configurable)

Default: **`agent-critical` + `free-auxiliary`**.

## Behavior

1. **P8 gate**: confirm the question cannot be answered by code/grep/docs. If it can, say so and skip fusion.
2. **Announce**: which combos will be consulted and approximate token cost.
3. **Query in parallel** the SAME question to each combo via `omniroute_route_request`, read-only.
4. **Present** each response labeled with its combo, with fact/inference/hypothesis marked.
5. **Comparison table** + **Convergences** / **Divergences** / **synthesized recommendation** with confidence.
6. **Do NOT synthesize action or edit anything** — deliver perspectives for you or the primary agent to decide.
7. **Save** to `${AIWF_CHANGE_ROOT}/evidence/fusion/opinion-<timestamp>.md`.

## Output

Side-by-side report + table + recommendation. Divergence between models is flagged as
the most valuable thing to review, never averaged away.
