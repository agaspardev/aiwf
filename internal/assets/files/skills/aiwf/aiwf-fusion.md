# Skill: /aiwf-fusion <tarea>

Consenso ejecutable: corre una tarea concreta por varios combos, produce la **tabla de
acuerdos/discrepancias**, sintetiza un resultado único marcando las divergencias como
riesgo, e itera hasta converger o escala tras 2 rondas. Generaliza `judgment-day`
(paralelo → sintetizar → iterar) de *review* a **fusión de soluciones cross-combo**.

Aplica el **[contrato de fusión](../_shared/fusion-contract.md)**.

## Cuándo usar

- Decisiones o generaciones de **alto riesgo / irreversibles** (arquitectura, un algoritmo
  crítico, una migración) donde una sola perspectiva puede tener puntos ciegos.
- **Opt-in y de alto valor** — nunca para tareas triviales (viola P8).

## Combos (configurables)

Resuelve los combos con `scripts/fusion-combos.ps1 -Skill fusion` (P8). Lee
`vault-config.local.json` → `fusion.fusion`; si no hay override, default:
**`agent-critical` + `agent-daily`** (dos tiers distintos = perspectivas más diversas).

## Comportamiento

1. **P8 gate + anuncio de costo** (como en el contrato).
2. **Genera en paralelo**: la misma tarea a cada combo vía `omniroute_route_request`.
   Cada combo produce su propuesta de forma independiente (sin verse entre sí).
3. **Tabla de acuerdos/discrepancias** (formato del contrato) punto por punto.
4. **Síntesis**: fusiona en una sola propuesta tomando lo mejor de cada una, y **marca
   explícitamente cada divergencia como hotspot de riesgo** (no la promedia ni la esconde).
5. **Iteración (estilo judgment-day)**: si hay divergencias sustanciales, corre una 2ª ronda
   donde cada combo revisa la síntesis. Máximo **2 rondas**; si no convergen, **escala al
   usuario** con las divergencias abiertas — no fuerces un falso consenso.
6. **El resultado es una PROPUESTA, no se auto-aplica.** Los gates deterministas (tests,
   `/quality-gate`, `/aiwf-security`) siguen decidiendo. El combo auxiliar nunca hace la
   escritura final ni la verificación.
7. **Guarda** en `.ai-workflow/evidence/fusion/fusion-<timestamp>.md` con tabla + propuesta
   + hotspots + combos y costo.

## Relación con judgment-day

- `judgment-day` = dos jueces del MISMO modelo revisando código (adversarial review).
- `/aiwf-fusion` = combos DISTINTOS generando/decidiendo, con síntesis cross-modelo.

Reutiliza su bucle de iteración/escalado; cambia el objetivo (fusión de soluciones) y el eje
(cross-combo en vez de cross-juez).
