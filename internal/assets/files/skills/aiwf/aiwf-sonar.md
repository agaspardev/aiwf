# Skill: /aiwf-sonar [changed|full|gate|issues]

Integración con SonarQube para análisis de calidad de código.

## Uso

```
/aiwf-sonar            # análisis de archivos modificados desde HEAD (default)
/aiwf-sonar changed    # igual que default
/aiwf-sonar full       # análisis completo del proyecto
/aiwf-sonar gate       # consulta el quality gate actual
/aiwf-sonar issues     # lista issues BLOCKER/CRITICAL nuevos
```

## Comportamiento

1. Delegar al script `ai sonar <modo>` (`scripts/sonar-scan.ps1`).
2. Si SonarQube no está habilitado en `vault-config.local.json`, informar cómo habilitarlo.
3. Después del análisis, leer el summary generado en `.ai-workflow/evidence/sonar/` e incorporarlo al contexto.
4. Reportar el Quality Gate status (OK / ERROR) con los issues más críticos.
5. Si el gate está en ERROR: listar los bloqueantes y sugerir `/quality-gate` para evaluación completa.

## Integración con quality-gate

Al ejecutar `/quality-gate`, esta skill incluye automáticamente el último `sonar-summary-*.md` disponible en `.ai-workflow/evidence/sonar/` como evidencia.
