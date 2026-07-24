---
name: rules-index
trigger: SIEMPRE cargado — router de fundamentos de ingeniería
type: foundation-index
---

# Fundamentos de ingeniería (router+módulo)

Reglas destiladas del libro Gentleman Programming, para **hacer las cosas bien desde el diseño y
el código**, no esperar a los jueces/quality gates al final. Progressive disclosure: este índice
siempre está; cargá el módulo completo SOLO cuando la tarea coincide con su trigger.

| Módulo | Cargar cuando… |
|---|---|
| `clean-code-agile` | SIEMPRE (transversal): planificar, codificar, definir tareas/tests |
| `clean-architecture` | diseñar estructura, capas, features, servicios |
| `hexagonal` | diseñar servicios con puertos/adaptadores, separar negocio de infraestructura |
| `typescript` | escribir/revisar TypeScript (`.ts`/`.tsx`) |
| `react` | componentes React, hooks, estado (`.jsx`/`.tsx`) |
| `angular` | Angular (componentes, servicios, DI, testing) |
| `barrels` | crear `index.ts` de re-exportación |
| `algorithms` | bucles, búsquedas, ordenamientos, estructuras con volumen |
| `frontend-fundamentals` | HTML/CSS/UI (semántica, responsive, accesibilidad) |

## Cómo se aplican (aseguramiento desde el inicio)
1. En **planificación/diseño**: cargá `clean-code-agile` + el módulo de arquitectura aplicable.
   Distinguí lógica de negocio vs casos de uso; definí capas y alcance; historia con casos de uso.
2. Al **codificar**: cargá el módulo del lenguaje/framework que estás tocando (por `applies_to`).
3. En **verify/review**: los jueces confirman lo que ya construiste bien; no son la primera línea de calidad.

Estos módulos los indexa `skill-registry.ps1` y el orquestador los rutea por trigger.
