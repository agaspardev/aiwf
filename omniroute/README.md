# OmniRoute — AI Proxy Hub

v3.8.49 | localhost:20128 | Dashboard: http://localhost:20128

> Documento de configuración actual — verificado contra el server en ejecución el 2026-08-06.
> Cambios post-instalación: ver [docs/changelog.md](docs/changelog.md).

## Architecture

OmniRoute es un proxy multi-provider con routing, caching, memoria vectorial, guardrails de seguridad, y pipeline de compresión.

```
User/CLI  →  OmniRoute (localhost:20128)  →  20 combos / ~15 providers (OpenRouter, Groq, Gemini, Ollama, Codex, etc.)
```

**Arranque**: Scheduled Task `OmniRoute` (`Running`) → `node.exe "...\omniroute.mjs" serve --no-open --no-tray`.
Workdir: `C:\Users\anton\.omniroute`. Puerto 20128. Restart:
`Stop-ScheduledTask -TaskName OmniRoute; Start-ScheduledTask -TaskName OmniRoute`.
El entry real del runtime es `dist\server.js` (Next standalone, distDir `./.build/next`).

## Active Features

### Memory System ✅ `PUT /api/settings/memory`
- **Estado**: ENABLED
- **Estrategia**: hybrid (FTS5 keyword + vector embeddings)
- **Vector store**: sqlite-vec (384 dims, Xenova/all-MiniLM-L6-v2)
- **Max tokens**: 2000 | **Retención**: 30 días | **skillsEnabled**: true
- **Engine**: FTS5+Embeddings+VectorStore disponibles

### Cache ✅ `PUT /api/settings/cache-config`
- **Semantic cache**: ENABLED, max 100 entries, TTL 1,800,000ms (30 min)
- **Prompt cache**: enabled (strategy: auto)
- **Idempotency window**: 5,000ms
- **Model catalog cache**: 60,000ms

### Guardrails ✅ `GET /api/guardrails`
| Guardrail | Estado |
|-----------|--------|
| vision-bridge | ENABLED (priority 5) |
| pii-masker | ENABLED (priority 10) |
| prompt-injection | ENABLED (priority 20) |
| credential-masker | ENABLED (priority 95) |

### Feature Flags relevantes ✅ `GET /api/settings/feature-flags`
15 activos de 42. Los no-default:

| Flag | Valor | Fuente |
|------|-------|--------|
| `INJECTION_GUARD_MODE` | warn | env |
| `PII_REDACTION_ENABLED` | true | env |
| `PII_RESPONSE_SANITIZATION` | true | env |
| `PII_RESPONSE_SANITIZATION_MODE` | redact | default |
| `OMNIROUTE_MCP_COMPRESS_DESCRIPTIONS` | true | env |
| `STREAM_RECOVERY_ENABLED` | true | env |
| `OMNIROUTE_EMERGENCY_FALLBACK` | true | default |
| `RESPONSES_PASSTHROUGH_DROP_COMMENTARY` | true | default |
| `OMNIROUTE_MCP_ENFORCE_SCOPES` | true | default |
| `OMNIROUTE_CODEX_WS_ENABLED` | true | default |
| `OMNIROUTE_ENABLE_LIVE_WS` | true | default |
| `OMNIROUTE_ALLOW_LOCAL_PROVIDER_URLS` | true | default |
| `ARENA_ELO_SYNC_ENABLED` | true | default |
| `ONEPROXY_ENABLED` | true | default |
| `OUTBOUND_SSRF_GUARD_ENABLED` | true | default |

### Plugins ✅ `GET /api/plugins` — 3 plugins ACTIVOS
| Plugin | Hooks | Descripción |
|--------|-------|-------------|
| cost-tracker | onResponse | Acumula costos por modelo |
| rate-limiter | onRequest | Sliding window por modelo |
| request-logger | onRequest, onResponse, onError | Timing + stats |

Detalle: `docs/plugins.md`. Código: `C:\Users\anton\aiwf\omniroute\plugins\`.

### Rate Limits ✅ `GET /api/rate-limits`
- Conexiones con `rateLimitProtection: true`, `maxConcurrent: 5`
- 0 lockouts registrados
- 42 conexiones configuradas total

### Compression Pipeline ✅ `PUT /api/settings/compression`
- **Default mode**: stacked
- **Stacked pipeline**: `session-dedup → rtk(standard) → caveman(lite) → ultra`
- **Auto-trigger**: lite a 16K tokens
- **Preserve system prompt**: true (mode: whenNoCache)
- **MCP compression**: habilitado (`OMNIROUTE_MCP_COMPRESS_DESCRIPTIONS=true`)
- **Ultra engine**: SLM (TinyBERT ONNX) con prewarming

### Env Vars ✅ (cargadas desde `~/.omniroute/.env`)

| Variable | Valor | Efecto |
|----------|-------|--------|
| `INJECTION_GUARD_MODE` | warn | Detección de prompt injection |
| `PII_REDACTION_ENABLED` | true | Enmascarar PII en requests |
| `INPUT_SANITIZER_MODE` | redact | Sanitizar input (redact) |
| `PII_RESPONSE_SANITIZATION` | true | Sanitizar PII en responses |
| `STREAM_RECOVERY_ENABLED` | true | Recuperación automática de streams |
| `OMNIROUTE_MCP_COMPRESS_DESCRIPTIONS` | true | Comprimir descripciones MCP |

## Routing — 20 Combos ✅ `GET /api/combos`

Ver `docs/providers-y-combos.md` para la lista completa de modelos por combo.

| Combo | Models (resumen) | Uso |
|-------|------------------|-----|
| coding-auto | antigravity/gemini-3.5-flash-high, cx/gpt-5.6-sol-low, gemini-2.5-flash-lite, claude-sonnet-4-6-low, groq/llama, oc/deepseek, gh/gpt-4.1, ollama | Codificación principal |
| fast-fix | oc/deepseek-v4-flash-free, openrouter/nemotron-free, gh/gpt-4.1, gemini, groq | Fixes rápidos |
| free-first | claude-haiku, gemini-2.5-flash-lite/3.1-flash-lite/3.5-flash-low/medium, groq, openrouter-free, oc-free, ollama | Priorizar gratuitos |
| full-free | claude-haiku, gemini-lite, groq, openrouter-free, oc-free, ollama | 100% gratuito |
| antigravity-pro | gemini-3.1-pro-high/low, 3.5-flash-high/medium/low, gpt-oss, claude-opus/sonnet | Gemini Pro |
| claude-only | antigravity/claude-opus-4-8-low, claude-sonnet-4-6-low | Solo Claude (vía antigravity) |
| architecture | cx/gpt-5.6-sol-low, gemini-3.1-pro-high, claude, oc-free | Arquitectura |
| docs | gemini-3.5-flash-medium, oc-free, groq/gpt-oss-120b, openrouter | Documentación |
| mimo-deep-work | cx/gpt-5.6-sol-low, gemini-3.1-pro-high, claude, oc-free | Trabajo profundo |
| agent-daily / agent-critical / agent-auto | cx/gpt-*-low, claude, ollama | Agentes AI |
| coding-web-auxiliary | cgpt-web/gpt-5.5-pro/5.5, gweb/gemini-3.5-flash, qwen-web | Auxiliares web coding |
| architecture-web-auxiliary | cgpt-web/gpt-5.5-pro-extended, gpt-5.6-thinking, gweb/gemini-3.1-pro | Web arquitectura |
| free-auxiliary | qwen-web/qwen3.6-plus, cgpt-web/gpt-5.5 | Auxiliar gratuito |
| free-web-auxiliary | qwen-web/qwen3.7-max, qwen3.6-27b | Auxiliar web gratuito |
| gemini-web-auxiliary | gweb/gemini-3.5-flash, gweb/gemini-3.1-pro | Research vía Gemini |
| research-web-auxiliary | gweb/gemini-3.1-pro, cgpt-web/gpt-5.5-thinking, qwen-web/qwen3.7-max | Research |
| gpt-web-auxiliary | cgpt-web/gpt-5.5 | GPT web |
| gpt-agent | cx/gpt-5.6-sol-low, ollama | GPT agente |

## Conectividad ✅ `GET /api/rate-limits` + SQLite

- **~42 conexiones** configuradas (ver tabla `provider_connections` en `storage.sqlite`)
- **Providers activos** (is_active=1): antigravity (4 cuentas), codex, chatgpt-web, gemini (5), gemini-web (4), groq (4), github, jina-reader (1 de 3), mimocode (error upstream), ollama-local, opencode, openrouter (4), perplexity-web, qwen-web (3), xiaomi-mimo
- **Providers desactivados/error**: claude (inactivo, cuenta agotada), kiro (inactivo), kimi-coding (token_expired), jina-reader (2 inactivas)
- Modelos usados en combos llevan prefijo provider: `antigravity/`, `cx/`, `oc/`, `cgpt-web/`, `gweb/`, `groq/`, `gh/`, `openrouter/`, `qwen-web/`, `ollama-local/`

## Scripts

### `scripts/verify.ps1`
Verificación: 14 checks + status notes.

### `scripts/setup-plugins.ps1`
Escanea + activa los 3 plugins locales.

### Patch scripts (tool names)
Ver `docs/toolcloak-patch.md`. Scripts idempotentes en
`C:\Users\anton\aiwf\omniroute\scripts\patches\`:
`patch-toolcloak-opensse.ps1`, `patch-toolcloak-v1.ps1`,
`patch-toolcloak-1zjt2j7.ps1`, `patch-toolcloak-antigravity.ps1`.

## Disaster Recovery (DR)

> Si el equipo muere, restaurar OmniRoute y dejarlo **exactamente como está hoy** con UN comando:

```powershell
# 1. (Preventivo) Snapshot del estado actual → ~/.omniroute/dr-snapshots/latest/
.\scripts\dr-snapshot.ps1

# 2. (En el equipo nuevo) Restauración completa:
.\scripts\dr-restore.ps1
```

**`dr-restore.ps1`** hace, en orden:
1. Instala `omniroute@3.8.49` (versión **pineada**, no la última)
2. Detiene el server (scheduled task `OmniRoute`)
3. Restaura `storage.sqlite` + `.env` desde el snapshot (DEBEN ir juntos: `STORAGE_ENCRYPTION_KEY` descifra las credenciales de providers en la DB)
4. Re-aplica los 4 patch scripts de tool names (idempotentes)
5. Copia + activa los 3 plugins desde `..\plugins\`
6. Reinicia el server y ejecuta `verify.ps1`

Flags: `-Version` (pinear otra versión), `-SnapshotDir` (snapshot específico),
`-SkipInstall`, `-SkipRestart`, `-SkipPlugins`, `-DryRun` (ensayo, no toca nada), `-Force`.

**`dr-snapshot.ps1`** crea un snapshot autocontenido y consistente (backup online de
`storage.sqlite` vía better-sqlite3, sin detener el server) + `.env` + `export-json` + `info.json`.
`-OutDir` permite guardar el snapshot fuera de `~/.omniroute` (e.g. USB/disco).

> ⚠️ El snapshot vive en `~/.omniroute/dr-snapshots/`. Para un DR real, copiar ese snapshot
> a un medio externo (no está versionado en git — la carpeta `omniroute/` está gitignoreada).

> ⚠️ `scripts/restore-all.ps1` está **obsoleto** (restaura memory/cache/env pero NO combos,
> credenciales, plugins ni patches, y tiene valores de cache viejos). Usar `dr-restore.ps1`.

## Backup

- **Export JSON**: `GET /api/settings/export-json` → backup manual
- Último backup: `~/.omniroute/backup-YYYYMMDD-HHmmss.json`
- `db_backups/` + `backups/` en `~/.omniroute/`

## Config Files

- `~/.omniroute/.env` — Environment vars (6 no-default)
- `~/.omniroute/storage.sqlite` (+ `-wal`/`-shm`) — DB principal (82MB) — memoria, cache, conexiones, rate limits, combos
- `~/.omniroute/plugins/` — plugins activos
- `C:\Users\anton\aiwf\omniroute\` — esta documentación

## API Reference

| Endpoint | Method | Auth | Purpose |
|----------|--------|------|---------|
| `/v1/chat/completions` | POST | Bearer | Chat completions |
| `/v1/messages` | POST | Bearer | Anthropic Messages (Claude Code) |
| `/v1/models` | GET | Bearer | Listar modelos |
| `/api/settings/memory` | GET/PUT | Bearer | Memoria |
| `/api/memory/health` | GET | Bearer | Health check memoria |
| `/api/memory/engine-status` | GET | Bearer | FTS5/embeddings status |
| `/api/settings/cache-config` | GET/PUT | Bearer | Cache |
| `/api/settings/feature-flags` | GET | Bearer | Feature flags |
| `/api/settings/compression` | GET/PUT | Bearer | Pipeline compresión |
| `/api/guardrails` | GET | Bearer | Guardrails registrados |
| `/api/rate-limits` | GET/POST | Bearer | Rate limits |
| `/api/plugins` | GET | Bearer | Plugins activos |
| `/api/plugins/scan` | POST | Bearer | Descubrir plugins |
| `/api/plugins/<name>/activate` | POST | Bearer | Activar plugin |
| `/api/providers` | GET | Bearer | Conexiones |
| `/api/providers/{id}` | GET/PUT | Bearer | Actualizar conexión |
| `/api/combos` | GET | Bearer | Combos routing |
| `/api/model-combo-mappings` | GET | Bearer | Model → combo mappings |
| `/api/settings/export-json` | GET | Bearer | Backup completo |
| `/api/usage/history` | GET | Bearer | Historial de uso |
| `/api/webhooks` | GET/POST | Bearer | Webhooks (sin configurar) |

## Restart

```powershell
# Recomendado (scheduled task):
Stop-ScheduledTask -TaskName OmniRoute
Start-ScheduledTask -TaskName OmniRoute

# Alternativas CLI:
omniroute restart    # Puede fallar con EADDRINUSE
omniroute stop       # Shutdown graceful
omniroute start      # Fresh start
```

> ⚠️ `npm update -g omniroute` regenera `dist/` y **borra los patches** de tool names.
> Re-aplicar con los patch scripts (ver `docs/toolcloak-patch.md`).
