# Skill: /aiwf-doctor

Verifica el estado de todas las herramientas del AI Engineering Workflow.

## Comportamiento

Ejecutar `ai diagnostico` (que llama a `scripts/harness.ps1 --Doctor`) y reportar los resultados de forma clara.

## Categorías de verificación

1. **Claude Code**: versión y disponibilidad
2. **OmniRoute**: servidor HTTP en `127.0.0.1:20128`, API key, health endpoint
3. **SonarQube**: `docker ps` para el contenedor, endpoint `/api/system/status`
4. **Seguridad P1** (bloqueantes si faltan): `gitleaks`, `osv-scanner`
5. **Seguridad P2** (recomendados): `semgrep`, `trivy`
6. **Seguridad P3** (opcionales): `syft`
7. **MCP servers**: codegraph, context7, engram, security-adapter, plane
8. **Herramientas de desarrollo**: `node`, `docker`, `python`, `git`, `pnpm`

## Estados

- `[OK]` — Herramienta disponible y funcional
- `[WARN]` — Disponible pero con advertencia (versión antigua, config incompleta)
- `[DEGRADED]` — Faltante pero no bloqueante (P2/P3)
- `[CRITICO]` — Faltante y bloquea el workflow normal (P1)

## Post-diagnóstico

Resumir en una línea el estado general: `OPERATIVO`, `DEGRADADO` o `BLOQUEADO`, con la lista de ítems críticos si los hay.
