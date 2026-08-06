# Arquitectura de OmniRoute

> Versión documentada: v3.8.48 | Fecha: 2026-07-27

---

## Visión General

OmniRoute es un **proxy de routing inteligente** para modelos de IA. Actúa como intermediario entre herramientas de desarrollo (OpenCode, Claude Code, Cursor, etc.) y múltiples proveedores de modelos (Anthropic, OpenAI, Google, etc.).

> **Documentación de plugins**: Ver [plugins.md](plugins.md) para sistema completo.
> **Plugins del repo**: `../plugins/` — rate-limiter, cost-tracker, request-logger.

```
┌─────────────────────────────────────────────────────────────┐
│                     HERRAMIENTAS IA                         │
│  OpenCode │ Claude Code │ Cursor │ Codex │ Cline │ ...     │
└──────┬──────┬──────┬──────┬──────┬──────┬──────────────────┘
       │      │      │      │      │      │
       ▼      ▼      ▼      ▼      ▼      ▼
┌─────────────────────────────────────────────────────────────┐
│                    OMNIRITE PROXY                           │
│                  localhost:20128                             │
│                                                             │
│  ┌──────────────┐  ┌──────────────┐  ┌──────────────────┐  │
│  │   Routing    │  │  Compresión  │  │   Quota/Rate     │  │
│  │   Engine     │  │   Pipeline   │  │   Limiting       │  │
│  │              │  │              │  │                  │  │
│  │ - Combos     │  │ - session    │  │ - Per-key        │  │
│  │ - Fallback   │  │   -dedup     │  │ - Per-provider   │  │
│  │ - Retry      │  │ - RTK        │  │ - Circuit breaker│  │
│  │ - Auto       │  │ - Caveman    │  │ - Quota pools    │  │
│  └──────────────┘  │ - Headroom   │  └──────────────────┘  │
│                    │ - Fidelity    │                        │
│                    │   Gate       │  ┌──────────────────┐  │
│                    │ - Risk Gate  │  │   Provider       │  │
│                    └──────────────┘  │   Registry       │  │
│                                      │                  │  │
│                    ┌──────────────┐  │ - Anthropic      │  │
│                    │  Model       │  │ - OpenAI         │  │
│                    │  Aliases     │  │ - Google         │  │
│                    │              │  │ - OpenRouter     │  │
│                    │ opus → cc/   │  │ - Groq           │  │
│                    │ sonnet → agy/│  │ - 160+ providers │  │
│                    └──────────────┘  └──────────────────┘  │
│                                                             │
│  ┌──────────────────────────────────────────────────────┐  │
│  │              SQLite Persistence                       │  │
│  │  ~/.omniroute/storage.sqlite                          │  │
│  │  - key_value (config, compression, settings)          │  │
│  │  - combos + combo_targets (routing combos)            │  │
│  │  - compression_analytics (métricas)                   │  │
│  │  - call_logs (logs de requests)                       │  │
│  │  - api_keys (autenticación)                           │  │
│  │  - model_aliases (mapeo de modelos)                   │  │
│  │  - 100+ tablas totales                                │  │
│  └──────────────────────────────────────────────────────┘  │
└─────────────────────────────────────────────────────────────┘
       │      │      │      │      │      │
       ▼      ▼      ▼      ▼      ▼      ▼
┌─────────────────────────────────────────────────────────────┐
│                   PROVEEDORES EXTERNOS                      │
│  Anthropic │ OpenAI │ Google │ OpenRouter │ Groq │ ...     │
└─────────────────────────────────────────────────────────────┘
```

---

## Componentes Principales

### 1. Routing Engine

| Componente | Archivo | Función |
|-----------|---------|---------|
| Combo Router | `open-sse/services/combo.ts` | Itera targets en orden hasta éxito o fallo |
| Account Fallback | `open-sse/services/accountFallback.ts` | Cambio de cuenta en quota/rate-limit |
| Model Deprecation | `open-sse/services/modelDeprecation.ts` | Detección de modelos deprecados |
| Wildcard Router | `open-sse/services/wildcardRouter.ts` | Matching de rutas wildcard |
| Circuit Breaker | `src/shared/utils/circuitBreaker.ts` | Resiliencia por provider |

**17 estrategias de routing**: priority, weighted, fill-first, round-robin, P2C, random, least-used, reset-aware, reset-window, cost-optimized, strict-random, auto, lkgp, context-optimized, context-relay, headroom, fusion.

### 2. Compression Pipeline

Ver [pipeline-compresion.md](pipeline-compresion.md) para detalle completo.

### 3. Persistence Layer

| Tabla | Propósito |
|-------|-----------|
| `key_value` | KV store para configuración (namespace: compression, settings, etc.) |
| `combos` | Configuraciones de routing combo |
| `combo_targets` | Targets ordenados por combo |
| `provider_connections` | Conexiones a proveedores con credenciales |
| `api_keys` | API keys con scopes y quota |
| `compression_analytics` | Métricas de compresión por request |
| `call_logs` | Logs detallados de cada request |
| `model_aliases` | Mapeo de alias a modelos reales |
| `quota_snapshots` | Snapshots históricos de quota |

### 4. API & Auth

| Endpoint | Método | Propósito |
|----------|--------|-----------|
| `/v1/chat/completions` | POST | Proxy principal (OpenAI-compatible) |
| `/api/settings/compression` | GET/PUT | Configuración de compresión |
| `/api/compression/engines` | GET | Estado de engines |
| `/v1/models` | GET | Lista de modelos disponibles |
| Dashboard | GET | UI web en `localhost:20128` |

**Auth**: `Authorization: Bearer <OMNIROUTE_API_KEY>`

---

## Infraestructura

| Componente | Detalle |
|-----------|---------|
| SO | Windows 11 |
| Node.js | v24.15.0 |
| OmniRoute | v3.8.48 (npm global) |
| Puerto | 20128 (TCP) |
| DB | SQLite (better-sqlite3) |
| DB Path | `C:\Users\anton\.omniroute\storage.sqlite` |
| Env | `C:\Users\anton\.omniroute\.env` |
| Autostart | Deshabilitado |
| Instalación | `npm install -g omniroute` |

---

## Plugin System

OmniRoute v3.8.48 tiene un sistema de plugins vía **child_process spawn** + **sandbox VM**.
No existe marketplace funcional — los plugins se instalan por copia filesystem.

### Arquitectura de plugins

```
~/.omniroute/plugins/<name>/
├── plugin.json       # Manifest: name, version, hooks, entry
└── index.js          # Sandbox VM (sin require, solo crypto)

API REST:
  POST /api/plugins/scan          → Descubre plugins nuevos
  POST /api/plugins/<name>/activate → Activa en runtime + DB (status=active)
  POST /api/plugins/<name>/deactivate → Desactiva

Startup flow:
  1. loadAll() lee SQLite → plugins con status=active
  2. loadPlugin() → spawn child_process → eval sandbox
  3. registerHook(name, hookFn) → hooks registry
  4. ChatHandler → emitHookBlocking (onRequest) / emitHook (onResponse)
```

### Ciclo de vida de un request con plugins

```
Request → emitHookBlocking("onRequest") → rate-limiter (block if exceeded)
  → ChatHandler.process() → provider call
  → emitHook("onResponse") → cost-tracker + request-logger
  → Response
```

### Plugins del repo

Ver `plugins/` para implementaciones. Documentación completa en [plugins.md](plugins.md).

| Plugin | Hooks | Archivos |
|--------|-------|----------|
| rate-limiter | onRequest | `plugins/rate-limiter/` |
| cost-tracker | onResponse | `plugins/cost-tracker/` |
| request-logger | onRequest, onResponse | `plugins/request-logger/` |

### Notas sobre implementación

1. El sandbox de `index.js` corre en `worker_threads` sin acceso a `require()` salvo `crypto`
2. `loadPlugin()` para cada plugin crea un child_process separado
3. Los hooks se registran en el Map global `hooks` en `hooks.ts`
4. `emitHookBlocking` es async — corre hooks blocking en secuencia
5. `emitHook` es fire-and-forget para hooks no-blocking
6. La activación persiste (`status=active` en SQLite) y sobrevive reinicios
7. PowerShell `Invoke-RestMethod` tiene bugs con response streams del activate endpoint — usar .NET `WebRequest`

## Archivos Clave en node_modules

| Ruta | Propósito |
|------|-----------|
| `open-sse/services/compression/types.ts` | Tipos y defaults de compresión |
| `open-sse/services/compression/strategySelector.ts` | Orquestación del pipeline stacked |
| `open-sse/services/compression/cavemanRules.ts` | Reglas hardcoded de caveman |
| `open-sse/services/compression/fidelityGate.ts` | Red de seguridad post-compresión |
| `open-sse/services/compression/riskGate/riskGate.ts` | Protección de contenido sensible |
| `open-sse/services/compression/pipelineEngineBreaker.ts` | Circuit breaker del pipeline |
| `src/lib/db/compression.ts` | Persistencia de config en SQLite |
| `src/shared/validation/compressionConfigSchemas.ts` | Schema Zod del PUT endpoint |

---

## Datos de Rendimiento (2026-07-27)

| Métrica | Valor |
|---------|-------|
| Total requests | 2,850+ |
| Tokens ahorrados | 194.9M |
| Ahorro promedio | 17% |
| USD ahorrados | $17.57 |
| Providers activos | 9 (antigravity, claude, opencode, codex, gemini-web, chatgpt-web, groq, qwen-web, openrouter) |
| Validation fallbacks | 1,983 |
| Duración promedio | 268ms |
