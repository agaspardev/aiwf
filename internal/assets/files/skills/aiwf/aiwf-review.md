# Skill: /aiwf-review

Revisión adversarial independiente del trabajo realizado en la sesión actual.

## Cuándo usar

Al completar una implementación, refactoring o arquitectura importante. Esta skill
actúa como revisor independiente que NO sabe qué se hizo — evalúa el resultado, no la intención.

## Comportamiento

1. **Recopilar artefactos** a revisar:
   - Archivos modificados (`git diff --name-only HEAD`)
   - Último task packet activo en `.ai-workflow/task-packets/`
   - Criterios de aceptación definidos

2. **Revisión adversarial** (revisar en este orden):
   - ¿Qué está MAL o incompleto? (antes de lo que está bien)
   - ¿Qué supuesto es débil o no verificado?
   - ¿Qué puede fallar en producción bajo carga o casos borde?
   - ¿Qué decisión carece de justificación técnica?
   - ¿Hay acoplamiento oculto, deuda técnica o regresiones posibles?

3. **Reportar** en formato:
   - Problema principal (primera oración)
   - Supuestos débiles
   - Contraargumento
   - Recomendación concreta
   - Criterios de aceptación (verificados vs. pendientes)

4. **Clasificar hallazgos**: BLOCKER / WARN / OK

5. **Sugerir** (no ejecutar): `/aiwf-security` si hay riesgo de seguridad, `/aiwf-sonar` si hay deuda de calidad.

## Restricción

No validar el criterio "el código compila" como evidencia de corrección. Verificar comportamiento, no sintaxis.
