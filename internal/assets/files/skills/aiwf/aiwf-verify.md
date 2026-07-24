# Skill: /aiwf-verify

Quality gate final con evidencia. Verifica que el trabajo cumple todos los criterios antes de cerrar una fase SDD.

## Comportamiento

1. **Leer el estado actual**:
   - `.ai-workflow/state/workflow-state.json` → fase actual, task packet activo
   - Criterios de aceptación del task packet

2. **Recopilar evidencia disponible**:
   - Último `scan-summary-*.md` en `.ai-workflow/evidence/security/`
   - Último `summary-*.md` en `.ai-workflow/evidence/sonar/`
   - Resultados de tests (si existe `coverage/` o similar)

3. **Evaluar cada criterio de aceptación** como PASS / FAIL / SKIP:
   - PASS: evidencia verificable en archivos o comportamiento observable
   - FAIL: criterio no cumplido — requiere acción
   - SKIP: criterio no aplicable al scope actual (documentar razón)

4. **Gate final**:
   - Si todos los criterios son PASS o SKIP justificado: `GATE PASS` → proceder a la siguiente fase
   - Si hay algún FAIL: `GATE FAIL` → listar bloqueantes con acción específica requerida

5. **Guardar evidencia**: escribir resultado en `.ai-workflow/evidence/verify-<timestamp>.md`.

6. **Llamar `mem_session_summary`** al final de cada fase completada.

## Modo `--cross-model` (advisory, opt-in)

Validación cruzada: un combo **distinto** al que escribió el código revisa el `git diff`
y emite notas. Aplica el **[contrato de fusión](../_shared/fusion-contract.md)**.

- **Cuándo**: al invocar `/aiwf-verify --cross-model`. No es el default (cuesta tokens).
- **Combo (configurable)**: resuelto por `scripts/fusion-combos.ps1 -Skill verify`
  (`vault-config.local.json` → `fusion.verify`; default **`free-auxiliary`**, barato). El
  combo revisa el `git diff HEAD` read-only vía `omniroute_route_request` y devuelve
  hallazgos: posibles bugs, casos borde omitidos, supuestos débiles.
- **STRICTAMENTE ADVISORY**: los hallazgos se etiquetan `ADVISORY` y van a una sección
  "Advisory cross-model" del reporte. **Nunca** cambian el veredicto PASS/FAIL del gate.
- **Los checks deterministas mandan**: tests, gitleaks, sonar, security siguen siendo la
  única fuente de verdad. Un "se ve bien" cross-modelo **no** convierte un FAIL en PASS,
  ni un PASS depende de él. El combo auxiliar nunca hace la verificación final.
- **Salida**: sección extra en `.ai-workflow/evidence/verify-<timestamp>.md`.

## Gatekeeper de fase (result contract)

Al cerrar una fase SDD, produce un **phase-contract** (schema `phase-contract.schema.json`:
status, executive_summary, artifacts[], risks[], next_phase) y valídalo con el gatekeeper
**determinista** antes de avanzar:

```
ai gate -Contract .ai-workflow/state/<fase>-contract.json
```

El gatekeeper (P8, cero tokens) verifica contra el repo: **¿los artefactos declarados
existen? ¿status=passed sin artefactos = alucinación? ¿riesgos críticos abiertos?** Sin
contrato validado, la fase no avanza — el fallo se detecta donde ocurrió, no tres fases después.

## Findings ledger (loop until dry)

En review adversarial, registra cada hallazgo en un **ledger** (schema `findings-ledger.schema.json`:
id, lens, location, severity, state, evidence_class, causal_disposition, evidence) en
`.ai-workflow/evidence/review/`. Barre hasta **2 pasadas secas consecutivas**, con **techo duro
de 4**. Solo hallazgos `introduced`/`behavior-activated`/`worsened` entran a corrección; el resto
es follow-up o escala. El re-review es quirúrgico (verifica el ledger contra el diff de fixes,
no re-lee todo).

## Nota

Esta skill NO reemplaza los tests automáticos — los complementa. Un gate que pasa aquí con
tests fallando NO es un gate que pasa. `--cross-model` añade una lectura extra, nunca
autoridad extra.
