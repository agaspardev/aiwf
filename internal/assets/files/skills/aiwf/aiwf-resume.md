# Skill: /aiwf-resume

Retoma una sesión de trabajo desde el último handoff guardado.

## Comportamiento

1. **Buscar el último handoff**:
   - Listar archivos en `.ai-workflow/handoffs/` ordenados por fecha (más reciente primero)
   - Si no hay handoffs: reportar "No hay handoffs previos. Usar `/aiwf-init` para inicializar."

2. **Leer el handoff**:
   - Qué se logró en la sesión anterior (`accomplished`)
   - Decisiones tomadas (`decisions`)
   - Bloqueos abiertos (`blockers`)
   - Próximos pasos (`next_steps`)

3. **Leer el estado del workflow**:
   - `.ai-workflow/state/workflow-state.json` → fase, task packet activo, último quality gate

4. **Leer evidencia reciente** (si existe):
   - Último security scan summary
   - Último sonar summary

5. **Reportar al usuario** en forma de briefing de sesión:
   ```
   RETOMANDO: <proyecto> — Fase: <fase>
   
   Logrado en sesión anterior:
   - <item>
   
   Próximos pasos (ordenados):
   1. <paso>
   
   Estado quality gate: <OK|ERROR|DESCONOCIDO>
   ```

6. **Sugerir primer paso** basado en `next_steps[0]` del handoff.

## Equivalente CLI

`ai seguir` (que llama a `scripts/harness.ps1 --Continue`)
