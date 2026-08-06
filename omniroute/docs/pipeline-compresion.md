# Pipeline de Compresión — Detalle Completo

> Configuración actual: `stacked` mode con 5 engines secuenciales

---

## Pipeline Actual

```
Request entrante
       │
       ▼
┌─────────────────────┐
│   1. SESSION-DEDUP   │  Elimina mensajes duplicados en la sesión
│                      │  (cross-turn deduplication)
│  Riesgo: CERO       │  Solo colapsa mensajes idénticos exactos
└──────────┬──────────┘
           │
           ▼
┌─────────────────────┐
│   2. RTK (standard)  │  Compresión de tool outputs y código
│                      │
│  - enableGrouping    │  Agrupa líneas similares
│  - stripCodeComments │  Elimina comentarios de código
│  - applyToCodeBlocks │  Procesa bloques fenced
│  - dedupThreshold: 2 │  Colapsa 2+ líneas idénticas
│  - maxLines: 120     │  Trunca a 120 líneas
│  - maxChars: 12000   │  Trunca a 12K chars
│                      │
│  Riesgo: BAJO        │  SmartTruncate preserva head+tail
└──────────┬──────────┘
           │
           ▼
┌─────────────────────┐
│   3. CAVEMAN (full)  │  Compresión semántica de lenguaje natural
│                      │
│  - intensity: full   │  Reglas agresivas de condensación
│  - roles: [user,     │  Comprime ambos roles
│    assistant]        │
│  - minMsgLen: 30     │  Ignora mensajes cortos
│  - autoClarity: true │  Output mode con auto-clarity
│  - Idiomas: en + es  │  Reglas bilingües
│                      │
│  Riesgo: BAJO-MEDIO  │  Preserva URLs, código, paths
└──────────┬──────────┘
           │
           ▼
┌─────────────────────┐
│   4. HEADROOM        │  SmartCrusher — compactación tabular
│                      │  (listas, tablas, estructuras repetitivas)
│  Riesgo: CERO       │  Solo estructuras tabulares
└──────────┬──────────┘
           │
           ▼
┌─────────────────────┐
│   5. ULTRA (Tier-B)  │  Pruneo semántico TinyBERT ONNX
│                      │  (57MB, CPU-only, fail-open)
│  - ultraEngine: slm  │  Modelo MobileBERT ONNX
│  - compressionRate   │  0.5 (mantiene 50% de tokens)
│  - minScoreThreshold │  0.3 (tokens < 0.3 se eliminan)
│  - prewarm: true     │  Modelo pre-calentado al iniciar
│  - fallback: Tier-A  │  Si SLM falla → heuristic
│                      │
│  Riesgo: MEDIO       │  Prunea stopwords y tokens de bajo
│                      │  valor semántico. Code blocks protegidos
│                      │  por extractPreservedBlocks
└──────────┬──────────┘
           │
           ▼
┌─────────────────────┐
│  POST-PIPELINE       │
│                      │
│  Fidelity Gate ──────│─ Valida que tokens críticos sobrevivan
│                      │  (URLs, identificadores, JSON keys, números)
│                      │  Si falla → RECHAZA el paso del engine
│                      │
│  Risk Gate ──────────│─ Protege contenido sensible
│                      │  (private keys, k8s secrets, DDL, legal)
│                      │  Enmascara → engines lo saltan → restaura
│                      │
│  Circuit Breaker ────│─ Si un engine falla 3 veces seguidas
│                      │  → se salta por 30s → probe → recovery
│                      │
└──────────┬──────────┘
           │
           ▼
    Request comprimido
    enviado al provider
```

---

## Configuración Detallada por Engine

### Session-Dedup

```json
{
  "engine": "session-dedup"
}
```

- **Qué hace**: Compara mensajes actuales con los N mensajes anteriores. Si un mensaje es idéntico a uno ya enviado, lo reemplaza con un marker `[CCR retrieve hash=...]`
- **Riesgo**: CERO — los markers son recuperables
- **Ahorro**: 10-30% en sesiones con contexto repetitivo

### RTK (Rapid Token Kompressor)

```json
{
  "engine": "rtk",
  "intensity": "standard",
  "config": {
    "enabled": true,
    "intensity": "standard",
    "applyToToolResults": true,
    "applyToCodeBlocks": true,
    "applyToAssistantMessages": false,
    "enableGrouping": true,
    "groupingThreshold": 3,
    "stripCodeComments": true,
    "preserveDocstrings": true,
    "deduplicateThreshold": 2,
    "customFiltersEnabled": true,
    "trustProjectFilters": false,
    "rawOutputRetention": "failures",
    "rawOutputMaxBytes": 1048576,
    "maxLinesPerResult": 120,
    "maxCharsPerResult": 12000,
    "enabledFilters": [],
    "disabledFilters": []
  }
}
```

- **Qué hace**: Compresión de tool outputs (grep, bash, file reads). Agrupa líneas similares, elimina comentarios, trunca contenido largo
- **applyToCodeBlocks**: true — procesa bloques ` ```...``` `
- **applyToAssistantMessages**: false — NO comprime respuestas del asistente (riesgo alto)
- **deduplicateThreshold**: 2 — colapsa 2+ líneas idénticas (era 3)
- **rawOutputRetention**: "failures" — guarda output crudo solo en fallos

### Caveman

```json
{
  "engine": "caveman",
  "intensity": "full",
  "config": {
    "enabled": true,
    "compressRoles": ["user", "assistant"],
    "skipRules": [],
    "minMessageLength": 30,
    "preservePatterns": [],
    "intensity": "full"
  }
}
```

- **Qué hace**: Compresión semántica de lenguaje natural. Elimina frases relleno, condensa explicaciones
- **intensity: full** — reglas agresivas (era lite)
- **compressRoles**: user + assistant — comprime ambos lados
- **minMessageLength**: 30 — ignora mensajes de menos de 30 chars
- **Language packs**: en + es — reglas bilingües

### Headroom (SmartCrusher)

```json
{
  "engine": "headroom"
}
```

- **Qué hace**: Compactación de estructuras tabulares (listas, tablas JSON, outputs de herramientas con estructura repetitiva)
- **Riesgo**: CERO — solo toca estructuras tabulares

### Ultra (Tier-B SLM/ONNX)

```json
{
  "engine": "ultra",
  "config": {
    "enabled": true,
    "compressionRate": 0.5,
    "minScoreThreshold": 0.3,
    "slmFallbackToAggressive": true,
    "maxTokensPerMessage": 0
  }
}
```

- **Qué hace**: Pruneo semántico de tokens usando TinyBERT ONNX (57MB, CPU-only). Analiza cada token con un modelo MobileBERT y asigna un score de información. Tokens con bajo score se eliminan preservando el significado
- **ultraEngine**: `"slm"` — usa TinyBERT ONNX (Tier-B). Si falla, cae a Tier-A (heuristic) automáticamente
- **ultraSlmPrewarm**: `true` — pre-calienta el modelo al iniciar
- **Modelo**: TinyBERT (57MB, fp32, CPU). Cache en `~/.omniroute/models/llmlingua/`
- **RAM**: ~200-300MB mientras el worker está vivo. Se auto-evicta tras 5 min sin uso
- **Protección de código**: Usa `extractPreservedBlocks` para tombstonear bloques fenced, inline code, URLs, etc. — NUNCA toca código
- **Fail-open**: 10 niveles de protección. Si cualquier cosa falla, retorna el texto original
- **Riesgo**: MEDIO — prunea tokens semánticamente prescindibles pero preserva significado mejor que Tier-A

---

## Caveman Output Mode

```json
{
  "cavemanOutputMode": {
    "enabled": true,
    "intensity": "lite",
    "autoClarity": true
  }
}
```

Post-procesamiento de la **salida del asistente** antes de que llegue al usuario. Elimina frases relleno ("Here's", "Below is"). `autoClarity` decide inteligentemente cuándo comprimir.

---

## Safety Gates (SQLite directo)

### Fidelity Gate

```json
{
  "fidelityGate": {
    "enabled": true,
    "minTokenSurvivalPercent": 95,
    "minJsonKeyPercent": 90,
    "checkNumericIntegrity": true,
    "checkDiffHunks": true
  }
}
```

Post-validación después de cada engine. Si un engine elimina >5% de tokens protegidos (URLs, identificadores, paths), >10% de JSON keys, algún número literal, o algún hunk header `@@` → **rechaza ese paso** y mantiene el body anterior.

### Risk Gate

```json
{
  "riskGate": {
    "enabled": true,
    "categories": [
      "stack_trace",
      "private_key",
      "secret_assignment",
      "k8s_secret",
      "db_migration",
      "legal"
    ]
  }
}
```

Pre-compresión: detecta y enmascara contenido sensible para que los engines lo salten.

### Pipeline Circuit Breaker

```json
{
  "pipelineCircuitBreaker": {
    "enabled": true,
    "failureThreshold": 3,
    "cooldownMs": 30000
  }
}
```

Si un engine falla 3 veces consecutivas跨requests → se salta automáticamente por 30 segundos → probe → recovery.

---

## Context Editing (Claude/Anthropic)

```json
{
  "contextEditing": {
    "enabled": true
  }
}
```

Feature delegada al provider: Anthropic limpia bloques tool-use viejos server-side. Solo funciona con modelos Claude.

---

## Language Config

```json
{
  "languageConfig": {
    "enabled": true,
    "defaultLanguage": "es",
    "autoDetect": true,
    "enabledPacks": ["en", "es"]
  }
}
```

Carga reglas de compresión en español e inglés. `autoDetect` detecta el idioma de cada mensaje.

---

## Stats de Compresión (en tiempo real)

| Métrica | Valor |
|---------|-------|
| Total requests | 2,715+ |
| Tokens ahorrados | 182.5M |
| Ahorro promedio | 17% |
| Engines activos | 5 (session-dedup, RTK, caveman, headroom, ultra) |
| Providers que más ahorran | antigravity (67M), opencode (46M), claude (32M) |
| Validation fallbacks | 1,860 |
| Duración promedio | 274ms |
