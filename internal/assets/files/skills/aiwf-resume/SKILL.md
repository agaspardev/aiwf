---
name: aiwf-resume
description: Resume a work session from the last saved handoff.
---
# Skill: /aiwf-resume

Resume a work session from the last saved handoff.

## Behavior

1. **Find the last handoff**:
   - List files in `.ai-workflow/handoffs/` sorted by date (most recent first)
   - If no handoffs: report "No previous handoffs. Use `/aiwf-init` to initialize."

2. **Read the handoff**:
   - What was accomplished in the previous session (`accomplished`)
   - Decisions made (`decisions`)
   - Open blockers (`blockers`)
   - Next steps (`next_steps`)

3. **Read the workflow state**:
   - `.ai-workflow/state/workflow-state.json` → current phase, active task packet, last quality gate

4. **Read recent evidence** (if available):
   - Last security scan summary
   - Last sonar summary

5. **Report to the user** as a session briefing:
   ```
   RESUMING: <project> — Phase: <phase>

   Accomplished in previous session:
   - <item>

   Next steps (ordered):
   1. <step>

   Quality gate status: <OK|ERROR|UNKNOWN>
   ```

6. **Suggest first step** based on `next_steps[0]` from the handoff.
