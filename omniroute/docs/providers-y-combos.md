# Providers, Combos y Model Aliases

> Configuración actual del routing de modelos — verificada contra el server en ejecución el 2026-08-06.
> Fuente: `GET /api/combos` (20 combos) + tabla `key_value` (namespace `modelAliases`) + `provider_connections`.

---

## Providers Activos

Los providers se configuran via Dashboard UI (`localhost:20128`) o API. Las credenciales se almacenan en `provider_connections` (encriptadas con AES-256-GCM).

| Provider ID | Estado | Cuentas | Uso en combos |
|-------------|--------|---------|---------------|
| `antigravity` | ✅ activo | 4 (antonio@innobyte.cl, ajgaspar1988@gmail.com, gaspar.antoniojesus@gmail.com, angelanrengifo@gmail.com) | Claude, Gemini, GPT-OSS |
| `codex` | ✅ activo | 1 (antonio@innobyte.cl) | `cx/gpt-5.6-sol-low`, `cx/gpt-5.5-low` |
| `groq` | ✅ activo | 4 (an.gaspar@duocuc.cl, angelanrengifo, sujeto89..., admin@antoniogaspar.dev) | llama-3.3-70b, gpt-oss, llama-4-scout, qwen |
| `gemini` | ✅ activo | 5 (Innobyte, ajgaspar1988, gaspar.antoniojesus, angela, sujeto) | Gemini vía API |
| `gemini-web` | ✅ activo (⚠️ error playwright-core browsers.json) | 4 | `gweb/gemini-*` vía web |
| `chatgpt-web` | ✅ activo | 1 (main) | `cgpt-web/gpt-5.5*` |
| `qwen-web` | ✅ activo | 3 (main, main-2, main-3) | `qwen-web/qwen3.7-max`, `qwen3.6-*` |
| `openrouter` | ✅ activo | 4 (ajgaspar1988, Innobyte, duoc, sujeto) | nemotron-free, modelos `:free` |
| `opencode` | ✅ activo | 1 | `oc/deepseek-v4-flash-free`, `oc/nemotron-3-ultra-free`, `oc/laguna-s-2.1-free` |
| `github` | ✅ activo | 1 | `gh/gpt-4.1` |
| `ollama-local` | ✅ activo | 1 | `ollama-local/granite3.3:8b` |
| `xiaomi-mimo` | ✅ activo | 1 (ajgaspar) | (sin uso en combos) |
| `perplexity-web` | ✅ activo | 1 (main) | (sin uso en combos) |
| `claude` | ⚠️ inactivo (cuenta agotada) | 1 | ❌ NO se usa en combos desde 05-08 |
| `kiro` | ❌ inactivo | 4 | — |
| `kimi-coding` | ❌ error token_expired | 1 | — |
| `mimocode` | ❌ error upstream | 1 | — |
| `jina-reader` | 1 activa / 2 inactivas | 3 | — |

> **Provider prefixes** usados en combos: `antigravity/`, `cx/` (codex), `oc/` (opencode),
> `cgpt-web/`, `gweb/`, `groq/`, `gh/` (github), `openrouter/`, `qwen-web/`, `ollama-local/`.

---

## Model Aliases (namespace `modelAliases` en `key_value`)

Mapean nombres cortos a modelos reales (con prefijo provider).

### Claude (prefijo `cc/` — directo Anthropic, SIN cuota; los combos usan `antigravity/claude-*-low`)

| Alias | Modelo real |
|-------|------------|
| `claude-fable-5` | `cc/claude-fable-5` |
| `claude-haiku-4-5-20251001` | `cc/claude-haiku-4-5-20251001` |
| `claude-opus-4-7` | `cc/claude-opus-4-7` |
| `claude-opus-4-8` | `cc/claude-opus-4-8` |
| `claude-sonnet-4-6` | `cc/claude-sonnet-4-6` |
| `claude-sonnet-5` | `cc/claude-sonnet-5` |
| `zzz-omni-probe` | `cc/claude-opus-4-8` |

### Gemini / Antigravity

| Alias | Modelo real |
|-------|------------|
| `gemini-3.1-pro` | `agy/gemini-pro-agent` |
| `gemini-3.1-flash-lite-preview` | `gemini/gemini-3.1-flash-lite` |

### OpenRouter (modelos gratuitos)

| Alias | Modelo real |
|-------|------------|
| `free` | `openrouter/openrouter/free` |
| `gemma-4-26b-a4b-it:free` | `openrouter/google/gemma-4-26b-a4b-it:free` |
| `gemma-4-31b-it:free` | `openrouter/google/gemma-4-31b-it:free` |
| `gpt-oss-20b:free` | `openrouter/openai/gpt-oss-20b:free` |
| `hy3:free` | `openrouter/tencent/hy3:free` |
| `kat-coder-air-v2.5:free` | `openrouter/kwaipilot/kat-coder-air-v2.5:free` |
| `laguna-m.1:free` | `openrouter/poolside/laguna-m.1:free` |
| `laguna-xs-2.1:free` | `openrouter/poolside/laguna-xs-2.1:free` |
| `lyria-3-clip-preview` | `openrouter/google/lyria-3-clip-preview` |
| `lyria-3-pro-preview` | `openrouter/google/lyria-3-pro-preview` |
| `nemotron-3-nano-30b-a3b:free` | `openrouter/nvidia/nemotron-3-nano-30b-a3b:free` |
| `nemotron-3-nano-omni-30b-a3b-reasoning:free` | `openrouter/nvidia/nemotron-3-nano-omni-30b-a3b-reasoning:free` |
| `nemotron-3-super-120b-a12b:free` | `openrouter/nvidia/nemotron-3-super-120b-a12b:free` |
| `nemotron-3-ultra-550b-a55b:free` | `openrouter/nvidia/nemotron-3-ultra-550b-a55b:free` |
| `nemotron-3.5-content-safety:free` | `openrouter/nvidia/nemotron-3.5-content-safety:free` |
| `nemotron-nano-9b-v2:free` | `openrouter/nvidia/nemotron-nano-9b-v2:free` |
| `nemotron-nano-12b-v2-vl:free` | `openrouter/nvidia/nemotron-nano-12b-v2-vl:free` |
| `north-mini-code:free` | `openrouter/cohere/north-mini-code:free` |

---

## Combos de Routing (20 activos)

Verificados `2026-08-06` via `GET /api/combos`. Orden = `sortOrder`.
Estrategias: `priority` (fallback en orden) o `round-robin`.

### 1. coding-auto `2efa7b0d` (sort 1, priority, upd 08-05) — Coding general
1. `antigravity/gemini-3.5-flash-high`
2. `cx/gpt-5.6-sol-low`
3. `antigravity/gemini-2.5-flash-lite`
4. `antigravity/claude-sonnet-4-6-low`
5. `groq/llama-3.3-70b-versatile`
6. `oc/deepseek-v4-flash-free`
7. `oc/nemotron-3-ultra-free`
8. `oc/laguna-s-2.1-free`
9. `openrouter/nvidia/nemotron-3-super-120b-a12b:free`
10. `gh/gpt-4.1`
11. `ollama-local/granite3.3:8b`

### 2. fast-fix `9871fcaf` (sort 2, priority, upd 08-06) — Fixes rápidos
1. `oc/deepseek-v4-flash-free`
2. `openrouter/nvidia/nemotron-3-super-120b-a12b:free`
3. `gh/gpt-4.1`
4. `antigravity/gemini-2.5-flash-lite`
5. `oc/laguna-s-2.1-free`
6. `groq/openai/gpt-oss-20b`
7. `antigravity/gemini-3.5-flash-medium`
8. `groq/meta-llama/llama-4-scout-17b-16e-instruct`

### 3. free-first `8e80e9d0` (sort 3, priority, upd 07-30) — Priorizar gratuitos
1. `antigravity/claude-haiku-4-5-20251001`
2. `antigravity/gemini-2.5-flash-lite`
3. `antigravity/gemini-3.1-flash-lite`
4. `antigravity/gemini-3.5-flash-low`
5. `antigravity/gemini-3.5-flash-medium`
6. `groq/llama-3.3-70b-versatile`
7. `groq/openai/gpt-oss-20b`
8. `groq/meta-llama/llama-4-scout-17b-16e-instruct`
9. `groq/qwen/qwen3.6-27b`
10. `openrouter/nvidia/nemotron-3-super-120b-a12b:free`
11. `oc/deepseek-v4-flash-free`
12. `oc/nemotron-3-ultra-free`
13. `oc/laguna-s-2.1-free`
14. `ollama-local/granite3.3:8b`

### 4. architecture `79ee7bfa` (sort 4, priority, upd 08-05) — Arquitectura
1. `cx/gpt-5.6-sol-low`
2. `antigravity/gemini-3.1-pro-high`
3. `antigravity/claude-opus-4-8-low`
4. `antigravity/claude-sonnet-4-6-low`
5. `oc/deepseek-v4-flash-free`
6. `oc/nemotron-3-ultra-free`
7. `oc/laguna-s-2.1-free`

### 5. docs `57819b9f` (sort 5, priority, upd 08-05) — Documentación
1. `antigravity/gemini-3.5-flash-medium`
2. `oc/deepseek-v4-flash-free`
3. `oc/nemotron-3-ultra-free`
4. `oc/laguna-s-2.1-free`
5. `groq/openai/gpt-oss-120b`
6. `openrouter/nvidia/nemotron-3-super-120b-a12b:free`

### 6. coding-web-auxiliary `2d2c2061` (sort 6, priority, upd 07-27) — [antes coding-web-first]
1. `cgpt-web/gpt-5.5-pro`
2. `cgpt-web/gpt-5.5`
3. `gweb/gemini-3.5-flash`
4. `qwen-web/qwen3.7-max`

### 7. architecture-web-auxiliary `1eb1c58d` (sort 7, priority, upd 07-27) — [antes architecture-web-first]
1. `cgpt-web/gpt-5.5-pro-extended`
2. `cgpt-web/gpt-5.6-thinking`
3. `gweb/gemini-3.1-pro`

### 8. mimo-deep-work `743c74b2` (sort 8, priority, upd 08-05) — Trabajo profundo
1. `cx/gpt-5.6-sol-low`
2. `antigravity/gemini-3.1-pro-high`
3. `antigravity/claude-opus-4-8-low`
4. `antigravity/claude-sonnet-4-6-low`
5. `oc/deepseek-v4-flash-free`
6. `oc/nemotron-3-ultra-free`
7. `oc/laguna-s-2.1-free`

### 9. agent-daily `c8ffc742` (sort 9, **round-robin**, upd 08-05) — Agentes diarios
1. `cx/gpt-5.6-sol-low`
2. `antigravity/claude-sonnet-4-6-low`

### 10. agent-critical `0343c90a` (sort 10, **round-robin**, upd 08-05) — Agentes críticos
1. `cx/gpt-5.6-sol-low`
2. `ollama-local/granite3.3:8b`
3. `antigravity/claude-opus-4-8-low`
4. `antigravity/claude-sonnet-4-6-low`

### 11. agent-auto `7cec8d40` (sort 11, **round-robin**, upd 08-05) — Agentes auto
1. `cx/gpt-5.5-low`
2. `ollama-local/granite3.3:8b`
3. `antigravity/claude-opus-4-8-low`
4. `antigravity/claude-sonnet-4-6-low`

### 12. free-auxiliary `f88234eb` (sort 12, priority, upd 07-27) — Auxiliar gratuito
1. `qwen-web/qwen3.6-plus`
2. `cgpt-web/gpt-5.5`

### 13. gemini-web-auxiliary `20f8fff6` (sort 13, priority, upd 07-27) — [antes gemini-research]
1. `gweb/gemini-3.5-flash`
2. `gweb/gemini-3.1-pro`

### 14. antigravity-pro `114e83a6` (sort 15, priority, upd 08-05) — Pro Gemini
1. `antigravity/gemini-3.1-pro-high`
2. `antigravity/gemini-3.1-pro-low`
3. `antigravity/gemini-3.5-flash-high`
4. `antigravity/gemini-3.5-flash-medium`
5. `antigravity/gemini-2.5-flash-lite`
6. `antigravity/gemini-3.1-flash-lite`
7. `antigravity/gemini-3.5-flash-low`
8. `antigravity/gpt-oss-120b-medium`
9. `oc/deepseek-v4-flash-free`
10. `antigravity/claude-opus-4-8-low`
11. `antigravity/claude-sonnet-4-6-low`

### 15. research-web-auxiliary `66ac1bea` (sort 16, priority, upd 07-27) — [antes research]
1. `gweb/gemini-3.1-pro`
2. `cgpt-web/gpt-5.5-thinking`
3. `qwen-web/qwen3.7-max`

### 16. claude-only `81312116` (sort 17, priority, upd 08-05) — Solo Claude
1. `antigravity/claude-opus-4-8-low`
2. `antigravity/claude-sonnet-4-6-low`

### 17. gpt-web-auxiliary `ea60de07` (sort 18, priority, upd 07-27) — [antes gpt-only]
1. `cgpt-web/gpt-5.5`

### 18. gpt-agent `33fd8674` (sort 19, priority, upd 07-29) — GPT agente
1. `cx/gpt-5.6-sol-low`
2. `cx/gpt-5.6-sol-low` (duplicado en config real)
3. `ollama-local/granite3.3:8b`

### 19. free-web-auxiliary `db6ac62a` (sort 20, priority, upd 07-27) — Auxiliar web gratuito
1. `qwen-web/qwen3.7-max`
2. `qwen-web/qwen3.6-27b`

### 20. full-free `cd8c6342` (sort 21, priority, upd 07-30) — 100% gratuito
1. `antigravity/claude-haiku-4-5-20251001`
2. `antigravity/gemini-2.5-flash-lite`
3. `antigravity/gemini-3.1-flash-lite`
4. `antigravity/gemini-3.5-flash-low`
5. `groq/llama-3.3-70b-versatile`
6. `groq/openai/gpt-oss-20b`
7. `openrouter/nvidia/nemotron-3-super-120b-a12b:free`
8. `oc/deepseek-v4-flash-free`
9. `oc/nemotron-3-ultra-free`
10. `oc/laguna-s-2.1-free`
11. `ollama-local/granite3.3:8b`

---

## OpenCode Provider Config

OpenCode se conecta a OmniRoute via provider personalizado:

```json
{
  "provider": {
    "omniroute": {
      "npm": "@ai-sdk/openai-compatible",
      "name": "OmniRoute",
      "options": {
        "baseURL": "http://localhost:20128/v1",
        "apiKey": "sk-898af3cd7ede926e-a76154-6891fe35"
      },
      "models": {
        "auto": { "name": "Auto (OmniRoute)" },
        "auto/coding": { "name": "Auto Coding" },
        "auto/cheap": { "name": "Auto Cheap" },
        "auto/fast": { "name": "Auto Fast" },
        "auto/best-free": { "name": "Auto Best Free" }
      }
    }
  }
}
```

---

## Modelo "Auto" de OmniRoute

Cuando se usa `auto` como modelo, OmniRoute selecciona automáticamente el mejor modelo basado en:
1. Quota disponible
2. Costo
3. Calidad requerida
4. Combo configurado

Los combos `coding-auto`, `free-first`, etc. definen las prioridades de selección.

---

## Notas de configuración por combo (config compartida)

Todos los combos activos usan por defecto:
```json
{
  "maxRetries": 0,
  "timeoutMs": 90000,
  "targetTimeoutMs": 35000,
  "trackMetrics": true,
  "compressionMode": "stacked",
  "failoverBeforeRetry": true
}
```
`computed_context_length`: 131072 (128K) en todos los combos.
