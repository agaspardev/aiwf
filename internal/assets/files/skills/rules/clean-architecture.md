---
name: rule-clean-architecture
trigger: al diseñar estructura, capas, módulos, o al crear features/servicios (arquitectura, diseño)
applies_to: ["**/*.ts", "**/*.tsx", "**/*.js", "**/*.jsx", "src/**"]
type: foundation
---

# Regla: Clean Architecture (fundamento, aplicar desde el diseño)

Del libro Gentleman Programming. **Fundamento a incorporar ANTES de codificar**, no a verificar al final.

## Capas (dependencias apuntan HACIA ADENTRO — Dependency Rule)
1. **Dominio**: entidades + reglas de negocio. **Independiente** de UI, DB, frameworks. Los conceptos fundamentales.
2. **Casos de uso**: lógica de aplicación que cumple requisitos funcionales; opera sobre el dominio, usa adaptadores.
3. **Adaptadores**: conectan casos de uso con el mundo externo (presentar datos / hablar con DB/API).
4. **Frameworks & Drivers**: capa más externa; React/Angular/DB/librerías concretas.

**REGLA DURA**: el dominio NO importa nada de las capas externas. La dependencia siempre va de afuera hacia adentro. Un cambio de framework o DB no debe tocar dominio ni casos de uso.

## Regla del Alcance (scope)
- **Root / global**: componentes y servicios reutilizables en toda la app (UI genérica, auth). Viven en la raíz.
- **Feature-específicos**: viven dentro de su módulo/contenedor; solo se usan en ese contexto. Candidatos a **lazy loading**.

## Estructura por funcionalidad (Screaming Architecture)
- Cada feature en su **propia carpeta con el nombre de la feature** (la estructura grita qué HACE la app, no qué framework usa).
- **Estructura orgánica**: no sobre-estructurar carpetas al inicio; que emerja de las necesidades. No inventar 8 niveles vacíos "por las dudas".

## Container / Presentational
- Un **componente contenedor** por feature (mismo nombre que la carpeta) con DOS responsabilidades: (a) estructura de presentación/layout, (b) lógica de negocio + obtención de datos (estado, servicios).
- Los **componentes presentacionales** hijos son lo más autónomos posible: reciben datos, renderizan, sin acoplarse a servicios.

## Al planificar/diseñar, preguntar SIEMPRE
- ¿Cuál es la lógica de DOMINIO (no cambia con la tecnología)? ¿Cuáles son casos de uso (sí cambian)?
- ¿Qué es global (root) y qué es feature-específico (lazy)?
- ¿La dependencia apunta hacia adentro? Si el dominio conoce React/DB, está MAL.
