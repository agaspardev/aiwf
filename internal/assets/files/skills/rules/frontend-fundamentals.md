---
name: rule-frontend-fundamentals
trigger: al escribir HTML, CSS o UI (accesibilidad, responsive, semántica)
applies_to: ["**/*.html", "**/*.css", "**/*.scss", "**/*.tsx", "**/*.vue"]
type: foundation
---

# Regla: Fundamentos Frontend

Del libro Gentleman Programming ("El Manual Definitivo del Frontend Developer").

## HTML semántico y DOM
- Usá etiquetas **semánticas** (`header`, `nav`, `main`, `section`, `article`, `footer`, `button`) — accesibilidad + SEO.
- Un `div` con `onClick` NO es un botón: usá el elemento correcto (teclado, lectores de pantalla, foco).
- El **CLS (content layout shift)** afecta UX y SEO: reservá espacio para imágenes/contenido que carga tarde.

## CSS
- Móvil primero, unidades relativas (`rem`, `%`, `fr`), flexbox/grid para layout.
- Evitá `!important` y selectores frágiles; preferí clases y variables CSS.

## JavaScript
- Fundamentos sólidos: tipos, coerción, `this`, closures, async/await, event loop. No dependas de magia del framework.

## Diseño responsivo
- Breakpoints según contenido, no según dispositivos concretos. Probá teclado y lectores de pantalla.

## Regla dura
La accesibilidad y la semántica se diseñan desde el markup inicial; parchearlas al final es caro y peor.
