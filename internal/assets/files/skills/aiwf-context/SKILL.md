---
name: aiwf-context
description: Report current session context usage and warn about compaction risk.
---
# Skill: /aiwf-context

Report current session context usage and warn about compaction risk.

## Behavior

Calculate or estimate current session context usage and report:

1. **Context budget** (contract prompt policy):
   - Code/files: max 20% of available context
   - Tool results: max 15%
   - Conversation history: max 18%

2. **Estimated current state**:
   - Messages in session
   - Files read
   - Recent tool calls

3. **Recommendation**:
   - If context is at risk: suggest handoff before compaction
   - If context is moderate: continue normally
   - If context is low: nothing to report

4. **Preventive action**: if context exceeds 60% of the limit, proactively suggest a handoff with `mem_session_summary` before forced compaction occurs.

## Note

This skill is informational. It cannot measure context with exact precision — it estimates based on visible session activity. For precise limits, consult the active model documentation.
