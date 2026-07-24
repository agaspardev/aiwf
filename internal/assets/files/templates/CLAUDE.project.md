# CLAUDE.md — {{PROJECT_NAME}}

_Inicializado: {{DATE}}_

---

## Contención de artefactos locales

REGLA OBLIGATORIA: Todos los archivos generados durante la sesión que no sean
código fuente del proyecto deben ir en `.ai-workflow/`:

| Tipo de artefacto | Ruta destino |
|---|---|
| Screenshots, capturas de browser | `.ai-workflow/reports/playwright/screenshots/` |
| Reportes de Playwright | `.ai-workflow/reports/playwright/` |
| Reportes de cobertura | `.ai-workflow/reports/coverage/` |
| Archivos temporales y experimentos | `.ai-workflow/scratch/` |
| Notas de sesión | `.ai-workflow/notes/` |
| Evidencia de scans de seguridad | `.ai-workflow/evidence/security/` |
| Outputs de SonarQube | `.ai-workflow/evidence/sonar/` |
| Handoffs entre sesiones | `.ai-workflow/handoffs/` |

NUNCA crear archivos de trabajo fuera de estas rutas.
P8: si el resultado puede obtenerse por código (cache, grep, git), no invocar un agente.

## Knowledge center

`.claude/knowledge/` contiene la documentación viva del proyecto — versionar siempre.
Actualizar los archivos relevantes al final de cada sesión con descubrimientos no obvios.

## Comandos útiles en este proyecto

```bash
aiwf sonar changed    # SonarQube analysis of modified files
aiwf security sast    # Quick SAST scan
aiwf estado           # Workflow status
```
