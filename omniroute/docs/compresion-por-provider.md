# Compresión por Provider — Comportamiento

> La compresión es **provider-agnostic** — aplica igual a todos los providers

---

## Princípio Fundamental

La compresión es **pre-processing**: se ejecuta ANTES de que el request llegue al provider. El body comprimido es lo que se envía. Los engines de compresión no saben qué provider están usando.

```
Request entrante (cualquier provider)
       │
       ▼
  bodyAdapter() ← normaliza formato (Chat Completions, Responses API, Kiro)
       │
       ▼
  session-dedup → rtk → caveman → headroom → ultra(Tier-B SLM)
       │
       ▼
  Body comprimido → Provider destino
```

---

## Compatibilidad por Feature

| Feature | Anthropic/Claude | OpenAI/GPT | Gemini | Groq | DeepSeek | Otros |
|---------|------------------|------------|--------|------|----------|-------|
| **session-dedup** | ✓ | ✓ | ✓ | ✓ | ✓ | ✓ |
| **RTK** | ✓ | ✓ | ✓ | ✓ | ✓ | ✓ |
| **caveman** | ✓ | ✓ | ✓ | ✓ | ✓ | ✓ |
| **headroom** | ✓ | ✓ | ✓ | ✓ | ✓ | ✓ |
| **ultra (Tier-B SLM)** | ✓ | ✓ | ✓ | ✓ | ✓ | ✓ |
| **Context Editing** | ✓ | ✗ | ✗ | ✗ | ✗ | ✗ |
| **OmniGlyph** | ✓ (solo Claude) | ✗ | ✗ | ✗ | ✗ | ✗ |
| **Cache-aware adjustment** | ✓ | ✓ | ✗ | ✗ | ✓ | ✓ |

---

## Detalles por Feature

### Compresión (Stacked Pipeline)

**Todos los providers.** Los engines no tienen awareness del provider:

- `session-dedup` — colapsa mensajes duplicados
- `RTK` — comprime tool outputs y código
- `caveman` — condensación semántica
- `headroom` — compactación tabular
- `ultra` — pruneo semántico TinyBERT ONNX

El `bodyAdapter()` normaliza tres formatos de body en `{ messages: [...] }` antes de que los engines ejecuten:
1. **Chat Completions** (`body.messages`) — passthrough
2. **OpenAI Responses API** (`body.input`) — flatten a role/content
3. **Kiro/AWS CodeWhisperer** (`body.conversationState`) — extrae tool results anidados

### Context Editing (Solo Claude)

**Claude/Anthropic exclusivo.** No es un engine de compresión — es una feature delegada al provider:

```json
{
  "contextEditing": { "enabled": true }
}
```

- Inyecta `body.context_management.edits[]` en el request
- Claude limpia bloques tool-use viejos server-side usando su propio tokenizer
- Para providers no-Claude: `context_management` se elimina automáticamente (`providerFieldStrips.ts`)
- Beta header `context-management-2025-06-27` se envía siempre en requests Claude

**Gating en el executor:**
```typescript
if (
  (this.provider === "claude" || isClaudeCodeCompatible(this.provider)) &&
  contextEditing?.enabled
) {
  applyContextEditingToBody(transformedBody, { enabled: true });
}
```

### OmniGlyph (Solo Claude con modelo específico)

**Único engine con hard provider gate.** Se auto-skips para non-Anthropic:

```typescript
// omniglyphAdapter.ts
if (providerTransport !== "direct") return { skip: "transport_not_direct" };
if (!isClaudeFormat(body)) return { skip: "source_format_not_claude" };
if (!isOmniGlyphSupportedModel()) return { skip: "model_not_approved" };
```

`providerTransport` se setea en chatCore:
```typescript
providerTransport: provider === "anthropic" ? "direct" : "aggregator"
```

### Cache-Aware Adjustment

**Protección de cache, no limitación.** Para providers con caching, degrada ultra/aggressive a "standard" para proteger el prefix cacheable:

| Provider | Tipo de Cache |
|----------|---------------|
| Claude/Anthropic | Claude-protocol cache |
| OpenAI/Codex/Azure | Automatic prefix caching |
| DeepSeek | Claude-protocol compatible |
| Alibaba/Qwen | OpenAI-compatible |

**Comportamiento:**
- Si el provider soporta caching → system prompt NUNCA se comprime (preserva cache hit)
- Si el provider NO soporta caching → compresión libre incluyendo system prompt
- `preserveSystemPromptMode: "whenNoCache"` — comprime system prompt solo si no hay caché

---

## Flujo Completo con Cache-Aware

```
Request entrante
       │
       ▼
  ¿Provider soporta cache?
       │
   ┌───┴───┐
   SÍ      NO
   │       │
   ▼       ▼
  System prompt   Compresión libre
  preservado      (todos los engines)
   │
   ▼
  Compresión de mensajes
  (session-dedup → rtk → caveman → headroom → ultra)
```

---

## comboOverrides (NO son por provider)

`comboOverrides` mapea **routing combos** a modos de compresión, NO providers:

```json
{
  "comboOverrides": {
    "fast-gpt": "lite",
    "deep-anthropic": "stacked"
  }
}
```

Un combo puede usar múltiples providers (fallback chain) — la misma compresión aplica a todos los targets del combo.

---

## Stats de Compresión por Provider (actuales)

| Provider | Requests | Tokens Saved |
|----------|----------|--------------|
| antigravity | 879 | 67.2M |
| opencode | 479 | 46.2M |
| claude | 797 | 31.9M |
| codex | 342 | 31.0M |
| gemini-web | 97 | 2.7M |
| chatgpt-web | 57 | 1.7M |
| groq | 53 | 1.1M |

**Total:** 2,715 requests, 182.5M tokens ahorrados, 17% promedio

---

## Referencias

| Archivo | Qué contiene |
|---------|--------------|
| `handlers/chatCore.ts` | Orquestación de compresión pre-provider |
| `handlers/chatCore/headers.ts` | Resolución de modo por request |
| `services/compression/strategySelector.ts` | Selección de plan (provider-agnostic) |
| `services/compression/bodyAdapter.ts` | Normalización de formatos de body |
| `services/compression/cachingAware.ts` | Ajuste de cache por provider |
| `services/compression/engines/omniglyphAdapter.ts` | Único engine con provider gate |
| `config/contextEditing.ts` | Context editing (Claude-only) |
| `utils/cacheControlPolicy.ts` | Detección de providers con caching |
