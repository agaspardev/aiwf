---
name: rule-hexagonal
trigger: al diseñar servicios/backend con puertos y adaptadores, o separar lógica de negocio de infraestructura
applies_to: ["**/*service*", "**/*port*", "**/*adapter*", "src/**"]
type: foundation
---

# Regla: Arquitectura Hexagonal (puertos y adaptadores)

Del libro Gentleman Programming. Separación de preocupaciones: la lógica de negocio vive en el hexágono y habla con el exterior SOLO por adaptadores.

## Actores y piezas
- **Actores primarios (izquierda, drivers)**: inician la acción. Nunca hablan directo con el servicio — usan un adaptador **driver**.
- **Actores secundarios (derecha, drivens)**: proveen recursos (DB, pagos, otro servicio). El servicio los usa vía adaptador **driven**.
- **Puertos**: interfaces que definen qué acciones ofrece/necesita el hexágono. Un servicio puede ser actor primario de otro.

## Tres tipos de lógica (distinguir SIEMPRE al analizar requisitos)
1. **Lógica de negocio**: viene del producto, no cambia por lo técnico (ej: "usuarios ≥18"). Vive en el dominio.
2. **Lógica de organización**: negocio reutilizado entre proyectos de la misma empresa (ej: validación de tarjetas).
3. **Casos de uso**: tienen limitación técnica, cambian con el uso (ej: layout/UX del error, posición de campos).

## Convenciones de nombres
- Puertos: **`For<Acción>`** (`ForRegistering`, `ForAuthenticating` que agrupa registro+login).
- Adaptadores: **`<Actor><Acción>`** (`Registerer`, `Authenticator`).

## Proceso (TDD-first — calidad desde el inicio)
1. Leer requisitos con cuidado. 2. Identificar lógica de negocio. 3. Identificar acciones que ofrece el hexágono.
4. Identificar recursos necesarios y quién los provee. 5. Crear puertos (drivers/drivens).
6. Crear adaptadores **stub/mock** para cerrar el hexágono. 7. Escribir tests de los casos de uso (fallan primero).
8. Implementar la lógica dentro del hexágono, **separada** de los detalles de drivers/drivens.

## Regla dura
La lógica del hexágono nunca conoce la implementación concreta de un adaptador. Cambiar la DB o el proveedor de pagos = cambiar un adaptador, nunca el servicio.
