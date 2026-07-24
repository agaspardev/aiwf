# Skill: /aiwf-security [sast|sca|secrets|sbom|all]

Ejecuta el pipeline de seguridad AppSec para el proyecto actual.

## Uso

```
/aiwf-security           # pipeline completo: secrets + sast + sca
/aiwf-security secrets   # solo Gitleaks (secretos en git)
/aiwf-security sast      # solo Semgrep (análisis estático)
/aiwf-security sca       # solo OSV-Scanner + Trivy (dependencias)
/aiwf-security sbom      # genera SBOM CycloneDX con Syft
/aiwf-security all       # igual que default (secrets + sast + sca)
```

## Comportamiento

1. Pedir confirmación al usuario antes de ejecutar (el scan puede tardar varios minutos).
2. Delegar a `ai security <scope>` (`scripts/security-scan.ps1`).
3. Leer el `scan-summary-*.md` generado en `.ai-workflow/evidence/security/`.
4. Reportar:
   - Hallazgos BLOCK (bloquean quality-gate)
   - Hallazgos WARN (documentar como deuda técnica)
   - Herramientas faltantes con instrucciones de instalación
5. Si hay BLOCK: indicar que `/quality-gate` no pasará hasta resolverlos.

## Checkpoints automáticos

Esta skill se sugiere automáticamente (pero no se ejecuta sin confirmación) cuando:
- Se añaden o quitan dependencias → sugerir `sca`
- Se modifican archivos en `auth/`, `crypto/`, `middleware/` → sugerir `sast`
- Se completa una feature task → sugerir `sonar`
- Se prepara un release o tag → sugerir `all`
