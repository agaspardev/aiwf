# CLAUDE.md — {{PROJECT_NAME}}

_Inicializado: {{DATE}}_

---

## Contención de artefactos locales

REGLA OBLIGATORIA: Todos los archivos generados durante la sesión que no sean
código fuente del proyecto deben ir en `.ai-workflow/`:

| Tipo de artefacto | Ruta destino |
|---|---|
| Screenshots, capturas de browser | `${AIWF_CHANGE_ROOT}/reports/playwright/screenshots/` |
| Reportes de Playwright | `${AIWF_CHANGE_ROOT}/reports/playwright/` |
| Reportes de cobertura | `${AIWF_CHANGE_ROOT}/reports/coverage/` |
| Archivos temporales y experimentos | `${AIWF_CHANGE_ROOT}/scratch/` |
| Notas de sesión | `${AIWF_CHANGE_ROOT}/notes/` |
| Evidencia de scans de seguridad | `${AIWF_CHANGE_ROOT}/evidence/security/` |
| Outputs de SonarQube | `${AIWF_CHANGE_ROOT}/evidence/sonar/` |
| Handoffs entre sesiones | `${AIWF_CHANGE_ROOT}/handoffs/` |

`AIWF_CHANGE_ROOT` DEBE estar resuelto por `/sdd-new` o `/sdd-continue` antes de generar cualquiera de estos artefactos. Si falta o hay múltiples changes abiertos, detenerse; nunca elegir el último modificado.
NUNCA crear archivos de trabajo fuera de estas rutas.
P8: si el resultado puede obtenerse por código (cache, grep, git), no invocar un agente.

## Knowledge center

`${AIWF_KNOWLEDGE_PROJECT_ROOT}/` contiene conocimiento vivo local del subproyecto. `${AIWF_KNOWLEDGE_SHARED_ROOT}/` contiene únicamente conocimiento transversal. Los changes enlazan estos artefactos; nunca los copian.
Actualizar solo el conocimiento impactado y mantener todo `.ai-workflow/` excluido de git.

## Comandos útiles en este proyecto

```bash
aiwf sonar changed    # SonarQube analysis of modified files
aiwf security sast    # Quick SAST scan
aiwf estado           # Workflow status
```
