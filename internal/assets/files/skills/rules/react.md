---
name: rule-react
trigger: al escribir o revisar componentes React, hooks, estado
applies_to: ["**/*.jsx", "**/*.tsx"]
type: foundation
---

# Regla: React (la joya sin marco)

Del libro Gentleman Programming ("Dominando React").

## Componentes y estado
- **Componentes funcionales** como receta: mismas props → misma UI. Puros donde se pueda.
- `useState` para estado local. Derivá estado en render en vez de duplicarlo en más estado.
- Entendé el **Virtual DOM** y la detección de cambios: minimizá renders innecesarios.

## Hooks con disciplina
- **`useEffect` solo para sincronizar con sistemas externos** (fetch, subscripciones, DOM). NO para lógica que puede ser derivada en render ni para reaccionar a eventos (eso va en el handler).
- Declará **todas las dependencias** del efecto; no las mientas. Limpiá subscripciones en el return.
- **Custom hooks** para extraer y reutilizar lógica con estado (nombre `use...`).
- `useRef` para valores mutables que no disparan render; `useMemo`/`useCallback` solo cuando hay costo medible (no por default).

## Comunicación entre componentes (en este orden de preferencia)
1. **Composición** (`children`, patrón Composition) — la primera opción.
2. **Context** — para estado compartido entre componentes sin relación directa; no lo uses como store global de todo.
3. Herencia — evitar.

## Robustez
- **Error Boundaries** para contener errores de render y no romper toda la app.
- **Portals** para UI fuera del árbol (modales, tooltips).
- **Axios interceptors** para centralizar auth, errores y refresh de tokens (no repetir en cada request).

## Regla dura
Un `useEffect` con dependencias mal declaradas o usado para derivar estado es un bug esperando. Diseñá el flujo de datos antes de escribir el componente.
