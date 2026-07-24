# Skill: /aiwf-opinion <pregunta>

Perspectivas paralelas de varios modelos sobre una pregunta acotada. **Read-only**:
ninguna de las opiniones toca el filesystem. Para decisiones de diseño, "¿este enfoque
es sólido?", segundas lecturas antes de comprometerse.

Aplica el **[contrato de fusión](../_shared/fusion-contract.md)**.

## Cuándo usar

- Antes de una decisión de diseño con tradeoffs no obvios.
- Cuando quieres una segunda (o tercera) lectura de un enfoque, no que alguien lo implemente.
- Cuando el costo de equivocarse supera el costo de N× tokens.

**No usar** para lo que un check determinista responde (P8), ni para tareas triviales.

## Combos (configurables)

Resuelve los combos con el script determinista (P8):
`scripts/fusion-combos.ps1 -Skill opinion`. Lee `vault-config.local.json` →
`fusion.opinion`; si no hay override, default: **`agent-critical` + `free-auxiliary`**.

## Comportamiento

1. **P8 gate**: confirma que la pregunta no se responde por código/grep/docs. Si se responde, dilo y no gastes fusión.
2. **Anuncia**: qué combos consultarás y que cuesta ~N× tokens. Procede sin pedir permiso salvo que el usuario haya pedido confirmación.
3. **Consulta en paralelo** la MISMA pregunta a cada combo vía `omniroute_route_request` (o `gemini-consult.ps1 -Combo <x>`), read-only.
4. **Presenta** cada respuesta etiquetada con su combo, con hecho/inferencia/hipótesis marcados.
5. **Tabla de comparación** (formato del contrato) + **Convergencias** / **Divergencias** / **recomendación sintetizada** con confianza.
6. **NO sintetiza acción ni edita nada** — entrega perspectivas para que tú o el agente primario decidan.
7. **Guarda** en `.ai-workflow/evidence/fusion/opinion-<timestamp>.md`.

## Salida

Reporte lado a lado + tabla + recomendación. La divergencia entre modelos se señala como
lo más valioso a revisar, no se promedia.
