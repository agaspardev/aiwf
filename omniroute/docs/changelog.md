# Changelog — Configuración de OmniRoute

---

## 2026-08-06 — Patch de tool names en converters de streaming (CRITICO)

### Problema
Los modelos OpenAI-compatibles (codex, openrouter, antigravity) devolvían tool
names en lowercase (`"name":"read"`) en el streaming, rompiendo las llamadas a
herramientas en Claude Code. Root cause: los converters de streaming de OmniRoute
emitían `name` raw sin aplicar el reverse-lookup del `REVERSE_MAP`.

### Parche aplicado (6 chunks en `dist\.build\next\server\chunks\`)
| Chunk | Converter | Patch | Sites |
|-------|-----------|-------|-------|
| `_0e8rngb._.js` | OPENROUTER→CLAUDE | `s.REVERSE_MAP` reverse-lookup | 3 |
| `_0fpntkz._.js` | OPENROUTER→CLAUDE | `s.REVERSE_MAP` reverse-lookup | 3 |
| `_0k-_ib1._.js` | OPENROUTER→CLAUDE | `s.REVERSE_MAP` reverse-lookup | 3 |
| `_1zjt2j7._.js` | OPENROUTER→CLAUDE | `s.REVERSE_MAP` reverse-lookup | 3 |
| `open-sse_0rryb_x._.js` | OPENAI→CLAUDE | `u.REVERSE_MAP` (módulo 75659, `u=e.i(786790)`) reverse-lookup | 2 (mid-stream + finish) |
| `open-sse_1n23zok._.js` | OPENAI→CLAUDE | idem | 2 |

Patrón aplicado:
```js
name:(Object.keys(s.REVERSE_MAP).find(q=>s.REVERSE_MAP[q]===a.name)??a.name)||""
```

### Verificación
- `node --check` OK en los 6 chunks
- Backups: `.bak` (×4 chunks), `.bak-opensse` + `.bak-opensse-2` (×2 open-sse)
- Test real post-patch: `cx/gpt-5.6-sol-low`, `openrouter/nvidia/nemotron-3-super-120b-a12b`,
  `antigravity/gemini-2.5-flash-lite` → `"name":"Read"` ✅ en `/v1/messages` y `/v1/chat/completions` (stream)
- Server reiniciado (PID 1972)

### Scripts de re-aplicación (idempotentes)
`patch-toolcloak-v1.ps1`, `patch-toolcloak-1zjt2j7.ps1`,
`patch-toolcloak-antigravity.ps1` (PATCH1, chunks s.REVERSE_MAP),
`patch-toolcloak-opensse.ps1` (PATCH2, open-sse u.REVERSE_MAP). Doble
ejecución verifica "NO CHANGES".

> ⚠️ `npm update -g omniroute` regenera `dist/` y BORRA estos patches. Re-aplicar.

### Nota aparte
`gemini-web` quedó con error de entorno no relacionado:
`Cannot find module '...playwright-core\browsers.json'`.

---

## 2026-08-05 — Reescritura de 10 combos: eliminado provider `claude`, todo `-low` vía antigravity

### Motivación
El provider `claude` (API directa Anthropic) quedó sin cuota. Todos los modelos
Claude pasan ahora por `antigravity/` con sufijos `-low` (lower-effort).

### Cambios aplicados (vía Dashboard/API, `updatedAt 2026-08-05`)
- Eliminado el provider `claude` de todos los combos; los Claude directos
  (`claude-opus-4-8`, `claude-sonnet-4-6`) pasan a `antigravity/claude-*-low`
- Combos reescritos: `coding-auto`, `architecture`, `docs`, `mimo-deep-work`,
  `agent-daily`, `agent-critical`, `agent-auto`, `antigravity-pro`, `claude-only`
- `agent-*` y `agent-daily` cambian a estrategia **round-robin**
  (`agent-daily`: gpt-5.6-sol-low | claude-sonnet-4-6-low)

### Renombrados
| Antiguo | Nuevo |
|---------|-------|
| `coding-web-first` | `coding-web-auxiliary` |
| `architecture-web-first` | `architecture-web-auxiliary` |
| `gemini-research` | `gemini-web-auxiliary` |
| `research` | `research-web-auxiliary` |
| `gpt-only` | `gpt-web-auxiliary` |

---

## 2026-07-30 — Combo `full-free` + downgrade a low-effort

### Cambios
- Creado combo **`full-free`** (id `cd8c6342`): 11 modelos 100% gratuitos
  (claude-haiku-4-5, gemini-2.5-flash-lite / 3.1-flash-lite / 3.5-flash-low,
  groq llama-3.3-70b / gpt-oss-20b, openrouter nemotron-free,
  oc deepseek / nemotron / laguna, ollama granite3.3:8b)
- `free-first` ampliado a 14 modelos (añadidos gemini-3.5-flash-low/medium,
  groq llama-4-scout, qwen3.6-27b)
- Downgrade de `codex`/`claude` a modo **low-effort** (`cx/gpt-5.6-sol-low`,
  `cx/gpt-5.5-low`, `antigravity/claude-*-low`)

---

## 2026-07-27 (noche) — Caveman `full` → `lite` (fidelidad sobre ratio)

### Motivación
Auditoría detectó **61.6% de reversiones** del fidelity gate (13,459/21,851 requests):
`caveman: "full"` + `minTokenSurvivalPercent: 95` son incompatibles por diseño.
Caveman full baja la supervivencia de tokens < 95%, el gate revierte, se paga
la compresión y se tira el resultado. Ahorro neto colapsado a ~9%.

### Decisión
Priorizar **fidelidad y ahorro neto** sobre ratio teórico por request. Menos
reversiones = menos CPU desperdiciado + menor riesgo de alucinación por
compresión agresiva.

### Cambios aplicados
- **caveman intensity**: `full` → `lite` (la API sólo acepta `lite|full|ultra`;
  no existe `standard`)
- Vía MCP `omniroute_set_compression_engine` + `omniroute_compression_configure`
- JSON sincronizados: `compression.json` (×3), `compression-api.json` (×1)

### Pendiente de validación
Medir `validationFallbacks` tras 24h de tráfico real vs. baseline 61.6%.
Muestra inicial (~77 requests): ~52%. No concluyente aún.

---

## 2026-07-27 (tarde) — Ultra Engine Tier-B (TinyBERT ONNX) activado

### Cambios aplicados

#### Via PUT API
1. **ultraEngine**: `"heuristic"` → `"slm"`
   - Activa Tier-B: TinyBERT ONNX (57MB, CPU-only)
   - Fail-open a Tier-A si SLM falla

2. **ultraSlmPrewarm**: `false` → `true`
   - Modelo pre-calentado al iniciar/reiniciar
   - Evita latencia de carga fría en primer request

### Configuración Ultra
```json
{
  "enabled": true,
  "compressionRate": 0.5,
  "minScoreThreshold": 0.3,
  "slmFallbackToAggressive": true,
  "maxTokensPerMessage": 0
}
```

### Tier-B SLM Details
- **Modelo**: TinyBERT (57MB, fp32 ONNX)
- **Repositorio**: `atjsh/llmlingua-2-js-tinybert-meetingbank`
- **Runtime**: ONNX via `@huggingface/transformers` en `worker_threads`
- **CPU-only**: Sin GPU requerida
- **RAM**: ~200-300MB mientras worker está vivo (auto-evict tras 5 min)
- **Cache**: `~/.omniroute/models/llmlingua/` (auto-descarga desde HuggingFace)
- **Timeouts**: 60s carga fría, 5s llamadas en caliente
- **Fail-open**: 10 niveles de protección en cadena

### Dependencias (ya instaladas)
- `@atjsh/llmlingua-2@2.0.3` ✓
- `@huggingface/transformers@3.5.2` ✓
- `@tensorflow/tfjs@4.22.0` ✓
- `js-tiktoken` ✓

### Lecciones aprendidas
- Las dependencias opcionales YA estaban instaladas en el `node_modules`
- Los modelos se auto-descargan en el primer request (no placement manual)
- `RUN_LLMLINGUA_INT` es solo documentación, no controla runtime
- El worker se auto-evicta tras 5 min de inactividad para liberar RAM

### Pipeline final
```
session-dedup → rtk(standard) → caveman(full) → headroom → ultra(Tier-B SLM)
```

---

## 2026-07-27 (mañana) — Optimización completa del pipeline

### Cambios aplicados

#### Via PUT API
1. **cavemanOutputMode**: `disabled` → `enabled: true, intensity: lite, autoClarity: true`
   - Post-procesamiento de output del asistente
   - Ahorra 5-15% en tokens de output

2. **applyToCodeBlocks**: `false` → `true`
   - RTK procesa bloques de código fenced
   - Ahorra 10-25% en resultados con código

3. **deduplicateThreshold**: `3` → `2`
   - Colapsa patrones de 2+ líneas idénticas
   - Ahorra 2-5% extra en logs repetidos

4. **preserveSystemPromptMode**: `always` → `whenNoCache`
   - Comprime system prompts cuando no están en caché

5. **rawOutputRetention**: `never` → `failures`
   - Guarda output crudo solo en fallos para recuperación

6. **RTK config completa re-aplicada** (hubo sobrescritura parcial):
   - enableGrouping: true
   - stripCodeComments: true
   - applyToCodeBlocks: true
   - deduplicateThreshold: 2
   - rawOutputRetention: failures

#### Via SQLite directo (safety gates)
7. **fidelityGate**: `no existía` → `enabled: true`
   - 95% token survival, 90% JSON key survival
   - Valida integridad numérica y diff hunks

8. **riskGate**: `no existía` → `enabled: true`
   - 6 categorías: stack_trace, private_key, secret_assignment, k8s_secret, db_migration, legal

9. **pipelineCircuitBreaker**: `no existía` → `enabled: true`
   - 3 fallos → skip 30s → probe → recovery

### Cambios previos (antes de esta sesión)
- Caveman intensity: `lite` → `full`
- Caveman compressRoles: `["user"]` → `["user","assistant"]`
- Caveman minMessageLength: `50` → `30`
- Language packs: `enabled=false` → `enabled=true, es+en`
- RTK enableGrouping: `false` → `true`
- RTK stripCodeComments: `false` → `true`
- Context Editing: habilitado (Claude)
- Pipeline: `[rtk, caveman]` → `[session-dedup, rtk, caveman, headroom]`
- session-dedup: habilitado
- headroom: habilitado

### Lecciones aprendidas
1. El PUT schema es `.strict()` — rechaza campos no listados
2. Enviar `rtkConfig` parcial SOBRESCRIBE todo el objeto
3. Auth para PUT es `Authorization: Bearer`, no `x-api-key`
4. Los safety gates se escriben directo en SQLite
5. `normalizeRtkConfig` no maneja `enableRenderers` (bug conocido)
6. `targetRatio` y `memoizeCompressionResults` son compile-time defaults

### Métricas finales
- Pipeline: session-dedup → rtk(standard) → caveman(full) → headroom → ultra(heuristic)
- Safety gates: fidelityGate + riskGate + circuitBreaker
- 182.5M tokens ahorrados, 17% promedio, $17.12 USD ahorrados

---

## 2026-07-29 — Plugin System: rate-limiter, cost-tracker, request-logger

### Contexto
Instalación limpia de OmniRoute v3.8.48 en Windows. Se descubrió que el sistema de
plugins no funciona via `omniroute plugin install` (instala desde npm) ni via
marketplace API (`/api/plugins/marketplace` devuelve seed data con `downloadUrl: ""`).

### Decisión
Implementar 3 plugins locales por copia filesystem directa a `~/.omniroute/plugins/`,
usando la API `POST /api/plugins/scan` + `POST /api/plugins/<name>/activate` para
activación en runtime.

### Plugins creados
| Plugin | Hooks | Función |
|--------|-------|---------|
| **rate-limiter** | `onRequest` | Sliding window rate limit por modelo (20 req/min) |
| **cost-tracker** | `onResponse` | Acumula costos por modelo en memoria |
| **request-logger** | `onRequest`, `onResponse` | Mide timing + acumula stats |

### Archivos
- `plugins/rate-limiter/plugin.json` + `index.js`
- `plugins/cost-tracker/plugin.json` + `index.js`
- `plugins/request-logger/plugin.json` + `index.js`
- `scripts/setup-plugins.ps1` — script de escaneo + activación
- `docs/plugins.md` — documentación completa del sistema de plugins

### API endpoints descubiertos
| Endpoint | Método | Status | Uso |
|----------|--------|--------|-----|
| `/api/plugins/scan` | POST | ✅ Funciona | Descubre plugins en `~/.omniroute/plugins/` |
| `/api/plugins/<name>/activate` | POST | ✅ Funciona | Activa plugin en runtime (+ persistencia) |
| `/api/plugins/marketplace` | GET | ⚠️ Seed data | No es funcional en v3.8.48 |
| `/api/plugins` | GET | ❌ 400 Bad Request | No usar; leer DB directo |
| `/api/plugins/<name>/deactivate` | POST | ✅ Documentado | Desactiva plugin |

### Hallazgos técnicos
1. `POST /api/plugins/<name>/activate` requiere body `{}` (Content-Type: application/json)
2. PowerShell `Invoke-RestMethod` tiene problemas con response stream del activate endpoint
3. Usar .NET `WebRequest` en su lugar para llamadas POST a la API de plugins
4. Los plugins corren en sandbox VM sin `require()` (solo `crypto`)
5. El código JS del plugin se evalua en Node.js child_process vía `loadPlugin`
6. La activación es dual: runtime (hooks registry) + persistente (status en DB)
7. `loadAll()` en startup carga plugins con `status = 'active'` de la DB
8. `pluginDir.startsWith(r + "/")` tiene bug de path separator en Windows
9. `plugin.json` tiene `hooks` field específico (ej: `{"onRequest": true}`)

### restore-all.ps1 ahora incluye:
1. Copia plugins del repo → `~/.omniroute/plugins/`
2. Ejecuta `setup-plugins.ps1` (scan + activate)
3. Luego `omniroute restart` para persistencia

### Dependencias
- Plugin sandbox usa hooks registry integrado de OmniRoute (sin deps externas)
