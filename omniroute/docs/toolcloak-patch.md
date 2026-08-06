# Patch de Tool Names en los Converters de Streaming (CRITICO)

> Fecha: 2026-08-06. Verificado contra el server v3.8.49 en ejecución.
> ⚠️ `npm update -g omniroute` regenera `dist/` y **BORRA estos patches**. Re-aplicar tras cada update.

---

## Problema

Los modelos **OpenAI-compatibles** (codex `cx/`, openrouter, antigravity) devolvían
los tool names en **lowercase** (`"name":"read"`) en el streaming. Los converters de
OmniRoute emitían el nombre raw sin aplicar el reverse-lookup del `REVERSE_MAP`,
rompiendo las llamadas a herramientas de Claude Code (que espera `"Read"`, `"Edit"`, etc.).

Root cause: en el converter OPENAI→CLAUDE del open-sse había 2 sitios raw
(`name:a.name||""` en mid-stream y `name:e.name||""` en finish) y en los
converters OPENROUTER→CLAUDE otros sitios con el mismo problema.

---

## Parche aplicado (6 chunks en `dist\.build\next\server\chunks\`)

| Chunk | Converter | Variable REVERSE_MAP | Sites parcheados |
|-------|-----------|----------------------|------------------|
| `_0e8rngb._.js` | OPENROUTER→CLAUDE | `s.REVERSE_MAP` | 3 |
| `_0fpntkz._.js` | OPENROUTER→CLAUDE | `s.REVERSE_MAP` | 3 |
| `_0k-_ib1._.js` | OPENROUTER→CLAUDE | `s.REVERSE_MAP` | 3 |
| `_1zjt2j7._.js` | OPENROUTER→CLAUDE | `s.REVERSE_MAP` | 3 |
| `open-sse_0rryb_x._.js` | OPENAI→CLAUDE | `u.REVERSE_MAP` (módulo 75659, `u=e.i(786790)`) | 2 (mid-stream `a.name` + finish `e.name`) |
| `open-sse_1n23zok._.js` | OPENAI→CLAUDE | `u.REVERSE_MAP` (módulo 75659, `u=e.i(786790)`) | 2 |

**Patrón de reemplazo aplicado:**

```js
// ANTES (raw):
name:a.name||""
// DESPUÉS (reverse-lookup):
name:(Object.keys(s.REVERSE_MAP).find(q=>s.REVERSE_MAP[q]===a.name)??a.name)||""
```

Para open-sse (`u.REVERSE_MAP`):
```js
name:(Object.keys(u.REVERSE_MAP).find(q=>u.REVERSE_MAP[q]===a.name)??a.name)||""
```

---

## Backups

| Chunk | Backup |
|-------|--------|
| `_0e8rngb._.js` | `_0e8rngb._.js.bak` |
| `_0fpntkz._.js` | `_0fpntkz._.js.bak` |
| `_0k-_ib1._.js` | `_0k-_ib1._.js.bak` |
| `_1zjt2j7._.js` | `_1zjt2j7._.js.bak` |
| `open-sse_0rryb_x._.js` | `.bak-opensse`, `.bak-opensse-2` |
| `open-sse_1n23zok._.js` | `.bak-opensse`, `.bak-opensse-2` |

Los `.bak-opensse-2` son el estado **después del PATCH1 y antes del PATCH2** (un stage más limpio).

---

## Scripts de re-aplicación (idempotentes)

Ubicación: `C:\Users\anton\aiwf\omniroute\scripts\patches\`
(antes en `C:\Users\anton\AppData\Local\Temp\opencode\`, movidos al repo el 2026-08-06;
los re-aplica automáticamente `scripts\dr-restore.ps1`)

| Script | Cubre | Detecta aplicado vía |
|--------|-------|----------------------|
| `patch-toolcloak-v1.ps1` | chunks `_0*.js` (OPENROUTER→CLAUDE, `s.REVERSE_MAP`) | patrón `Object.keys(s.REVERSE_MAP).find` |
| `patch-toolcloak-1zjt2j7.ps1` | `_1zjt2j7._.js` | idem |
| `patch-toolcloak-antigravity.ps1` | chunks `_0*.js` | idem |
| `patch-toolcloak-opensse.ps1` | `open-sse_*.js` (OPENAI→CLAUDE, `u.REVERSE_MAP`, 2 sites) | patrón `Object.keys(u.REVERSE_MAP).find` |

**Verificación de idempotencia**: ejecutar dos veces; la segunda debe reportar
"NO CHANGES".

---

## Verificación

### Sintaxis
- `node --check` OK en los 6 chunks parcheados.

### Functional (post-patch, server reiniciado PID 1972)
| Modelo | Endpoint | Resultado |
|--------|----------|-----------|
| `cx/gpt-5.6-sol-low` (codex) | `/v1/messages` stream | `"name":"Read"` ✅ |
| `openrouter/nvidia/nemotron-3-super-120b-a12b:free` | `/v1/messages` stream | `"name":"Read"` ✅ |
| `antigravity/gemini-2.5-flash-lite` | `/v1/messages` stream | `"name":"Read"` ✅ |
| `cx/gpt-5.6-sol-low` | `/v1/chat/completions` stream | `"name":"Read"` ✅ |

### Quick check manual
```powershell
# Verificar que ambos patches están vivos en cada chunk
$base = "C:\Users\anton\AppData\Roaming\npm\node_modules\omniroute\dist\.build\next\server\chunks"
foreach ($f in @("_0e8rngb._.js","_0fpntkz._.js","_0k-_ib1._.js","_1zjt2j7._.js")) {
  $c = Get-Content -Raw "$base\$f"
  "{0} : s.REVERSE_MAP={1}" -f $f, $c.Contains("Object.keys(s.REVERSE_MAP).find")
}
foreach ($f in @("open-sse_0rryb_x._.js","open-sse_1n23zok._.js")) {
  $c = Get-Content -Raw "$base\$f"
  "{0} : u.REVERSE_MAP={1}" -f $f, $c.Contains("Object.keys(u.REVERSE_MAP).find")
}
```

---

## Notas

- `gemini-web` falla con error de entorno NO relacionado:
  `Cannot find module '...\playwright\node_modules\playwright-core\browsers.json'`
  (falta el package anidado de playwright-core en la instalación npm global).
- Tras un `npm update -g omniroute`, re-aplicar los 4 scripts y reiniciar el server.
