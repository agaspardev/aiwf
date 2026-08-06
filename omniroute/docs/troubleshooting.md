# Troubleshooting — Problemas Conocidos y Soluciones

---

## Problema: PUT /api/settings/compression rechaza campos

**Síntoma**: Error 400 al enviar `fidelityGate`, `riskGate`, o `pipelineCircuitBreaker` via PUT.

**Causa**: El schema Zod (`compressionSettingsUpdateSchema`) es `.strict()` — solo acepta campos listados explícitamente. Estos 3 campos NO están en el schema.

**Solución**: Escribir directo en SQLite con `setup-safety-gates.mjs`.

---

## Problema: PUT parcial sobrescribe rtkConfig completo

**Síntoma**: Al enviar `{"rtkConfig": {"applyToCodeBlocks": true}}`, todos los demás campos de rtkConfig vuelven a defaults.

**Causa**: `normalizeRtkConfig()` spreads `DEFAULT_RTK_CONFIG` primero, luego asigna campos del input. Si el input solo tiene `applyToCodeBlocks`, los demás campos usan defaults (que son diferentes a los configurados).

**Solución**: SIEMPRE enviar el objeto `rtkConfig` COMPLETO en cada PUT.

---

## Problema: Auth falla con `x-api-key` header

**Síntoma**: Error 401 al usar header `x-api-key` para el endpoint de settings.

**Causa**: El endpoint de settings usa `Authorization: Bearer <key>`, no `x-api-key`.

**Solución**: Usar `Authorization: Bearer $apiKey`.

---

## Problema: OmniRoute no inicia después de npm update

**Síntoma**: `omniroute restart` falla o el proceso no arranca.

**Causa**: npm update puede sobrescribir archivos en node_modules.

**Solución**:
```powershell
omniroute stop
omniroute restart
```

---

## Problema: Safety gates se pierden después de npm update

**Síntoma**: fidelityGate/riskGate/circuitBreaker no están en el config después de actualizar.

**Causa**: Los safety gates están en SQLite (`~/.omniroute/storage.sqlite`), NO en node_modules. Sobreviven updates. Pero si se borra el `.omniroute` directory, se pierden.

**Solución**: Re-ejecutar `node scripts/setup-safety-gates.mjs` + `omniroute restart`.

---

## Problema: Validation fallbacks altos (1,983)

**Síntoma**: Muchos requests caen en fallback de validación.

**Causa**: La validación post-compresión detecta que la compresión excedió umbrales y revierte a la versión anterior. Es un comportamiento esperado — significa que los safety gates están funcionando.

**Solución**: No es un problema. Es el sistema protegiéndose. Si se vuelve excesivo, ajustar `minTokenSurvivalPercent` en fidelityGate.

---

## Problema: `enableRenderers` siempre queda en `false`

**Síntoma**: Aunque se envía `enableRenderers: true` en rtkConfig, siempre queda `false`.

**Causa**: Bug en `normalizeRtkConfig()` — no incluye `enableRenderers` en las asignaciones explícitas, así que usa el default (`false`).

**Solución**: No es un problema para compresión. `enableRenderers` activa "semantic renderers" que REESCRIBEN código (añaden tokens, no los ahorran). `false` es el valor correcto.

---

## Problema: `targetRatio` y `memoizeCompressionResults` no se pueden cambiar

**Síntoma**: No aparecen en el schema PUT, no se leen de la DB.

**Causa**: Son constantes compile-time en `DEFAULT_COMPRESSION_CONFIG`. No están en el switch de lectura de DB ni en el schema de escritura.

**Solución**: Los defaults ya son óptimos: `targetRatio: 0.7`, `memoizeCompressionResults: true`.

---

## Comandos Útiles de Diagnóstico

```powershell
# Estado completo de compresión
omniroute compression status --output json

# Verificar si OmniRoute está corriendo
Get-NetTCPConnection -LocalPort 20128

# Ver process ID
Get-Process -Id <PID> | Select-Object ProcessName, StartTime

# Verificar DB
node -e "const Database = require('C:\\Users\\anton\\AppData\\Roaming\\npm\\node_modules\\omniroute\\node_modules\\better-sqlite3'); const db = new Database('C:\\Users\\anton\\.omniroute\\storage.sqlite', {readonly:true}); console.log(db.prepare('SELECT COUNT(*) as n FROM key_value').get()); db.close();"

# Ver analytics recientes
omniroute compression status --output json | Select-String "tokensSaved"
```
