# Arquitectura del segundo cerebro

## Decisión canónica

La arquitectura es **configurable por instancia**. Cada instalador define sus propios vaultPaths
en `.ai-workflow/env/vault-config.local.json`. No hay rutas absolutas en el repositorio distribuible.

## Estructura recomendada (ejemplo)

```text
<tu-root-de-vaults>/
├── Personal/
├── Learning/
│   └── Certificaciones/
├── Projects/
│   └── AgenticOS/
├── Work/
│   └── <cliente-o-proyecto>/     ← scope restringido
└── System/                       ← gobernanza y registros
```

## Reglas de identidad

- Un vault de Obsidian es un directorio con `.obsidian/`.
- Un proyecto Git puede contener un vault como subdirectorio intencional.
- `System/` contiene gobernanza y registros — no es un vault de notas de contenido.

## Clasificación de sensibilidad

- **Privado:** `Personal`, `Learning`, `System`.
- **Restringido:** cualquier scope declarado en `restrictedScopes` en vault-config.local.json.
- Los permisos de hashing o copia técnica no implican permiso para analizar contenido restringido.

## Configuración por instancia

Completar `.ai-workflow/env/vault-config.local.json` en cada máquina:

```json
{
  "vaultPaths": {
    "operational": "~/.ai-workflow/vault",
    "personal":    "/path/absoluto/a/Personal",
    "learning":    "/path/absoluto/a/Learning",
    "work":        "/path/absoluto/a/Work"
  },
  "restrictedScopes": ["work"]
}
```

Los scripts `vault-search.ps1`, `vault-index.ps1` y `vault-lint.ps1` leen esta configuración
automáticamente via `scripts/shared/resolve-paths.ps1` → `Get-VaultConfig`.

## Convención mínima de notas (NOTE-SCHEMA)

```yaml
---
title: Nombre claro
type: source | concept | decision | finding | guide
domain: personal | learning | project | work | system
status: draft | verified | stale | superseded
updated: YYYY-MM-DD
verified_at: YYYY-MM-DD | null
review_after: YYYY-MM-DD | null
source_of_truth: []
supersedes: []
superseded_by: []
summary: Una frase que explica por qué importa.
---
```

## Principio de vigencia

- Toda información nueva entra como `draft`.
- Solo `verified` dentro de su ventana de revisión se considera conocimiento vigente.
- Indexar no es verificar.
- Una afirmación reemplazada se marca `superseded` y enlaza bidireccionalmente la nota sucesora.
- Los agentes deben usar búsquedas acotadas, validar hechos sensibles al tiempo y respetar
  el aislamiento de los scopes restringidos.

## Gobernanza

Las definiciones operativas de gobernanza de knowledge se ubican en `vault/KNOWLEDGE-SOURCES.md`
y en los archivos de gobernanza del scope `system` de cada instancia.
