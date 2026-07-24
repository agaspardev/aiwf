---
name: aiwf-fusion
description: Cross-combo consensus — run a task through multiple combos, synthesize results.
---
# Skill: /aiwf-fusion <task>

Executable consensus: run a concrete task through multiple combos, produce the
**agreement/discrepancy table**, synthesize a single result marking divergences
as risk, and iterate until convergence or escalate after 2 rounds.

Applies the **fusion contract**.

## When to use

- High-risk or irreversible decisions (architecture, critical algorithm, migration)
  where a single perspective may have blind spots.
- **Opt-in and high-value** — never for trivial tasks (violates P8).

## Combos (configurable)

Default: **`agent-critical` + `agent-daily`** (two distinct tiers = more diverse perspectives).

## Behavior

1. **P8 gate + cost announcement**.
2. **Generate in parallel**: same task to each combo via `omniroute_route_request`.
   Each combo produces its proposal independently (without seeing each other).
3. **Agreement/discrepancy table** point by point.
4. **Synthesis**: merge into a single proposal taking the best from each, and **explicitly
   mark each divergence as a risk hotspot** (don't average or hide it).
5. **Iteration**: if substantial divergences remain, run a 2nd round where each combo
   reviews the synthesis. Maximum **2 rounds**; if no convergence, **escalate to the user**
   with open divergences — never force a false consensus.
6. **The result is a PROPOSAL, not self-applied.** Deterministic gates (tests,
   `/quality-gate`, `/aiwf-security`) still decide. The auxiliary combo never writes
   the final output or runs verification.
7. **Save** to `.ai-workflow/evidence/fusion/fusion-<timestamp>.md` with table + proposal
   + hotspots + combos and cost.
