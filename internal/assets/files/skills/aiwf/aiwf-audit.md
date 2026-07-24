# Skill: /aiwf-audit

Auditoría completa del proyecto: arquitectura, calidad, seguridad y migrabilidad.

## Cuándo usar

- Al iniciar trabajo en un proyecto heredado (para entender el estado real)
- Antes de un release importante
- Cuando se sospecha deuda técnica acumulada
- Para evaluar si el proyecto cumple los criterios de distribución

## Comportamiento

### Fase 1 — Recopilación (P8: código antes que agente)

1. **Estructura**: `git ls-files | head -50` para mapeo rápido
2. **Dependencias**: detectar lock files y herramientas (npm, pnpm, go.sum, etc.)
3. **Configuración**: leer `sonar-project.properties`, `.security/policies/`
4. **Estado workflow**: `.ai-workflow/state/workflow-state.json`

### Fase 2 — Análisis por dimensión

- **Arquitectura**: ¿Hay separación clara de responsabilidades? ¿Acoplamiento oculto? ¿Decisiones sin documentar?
- **Calidad**: ¿Existe cobertura de tests? ¿SonarQube configurado? ¿Deuda técnica documentada?
- **Seguridad**: ¿Último security scan? ¿Gitleaks pre-commit hook? ¿Dependencias con CVEs críticos?
- **Migrabilidad**: ¿Hay rutas absolutas hardcodeadas? ¿Secretos en el repo? ¿`.gitignore` correcto?

### Fase 3 — Reporte

Formato:
```
AUDITORÍA: <proyecto>
Fecha: <hoy>

## Arquitectura     [OK|WARN|BLOCKER]
## Calidad          [OK|WARN|BLOCKER]
## Seguridad        [OK|WARN|BLOCKER]
## Migrabilidad     [OK|WARN|BLOCKER]

## Acciones recomendadas (ordenadas por prioridad)
1. [BLOCKER] <acción>
2. [WARN]    <acción>
```

Guardar resultado en `.ai-workflow/evidence/audit-<timestamp>.md`.
