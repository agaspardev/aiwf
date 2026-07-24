---
name: rule-typescript
trigger: al escribir o revisar TypeScript (tipos, interfaces, generics)
applies_to: ["**/*.ts", "**/*.tsx"]
type: foundation
---

# Regla: TypeScript (tipado con criterio)

Del libro Gentleman Programming ("TypeScript con De Tuti").

## Fundamentos
- **`any` es el enemigo**: apaga el chequeo de tipos. Evitalo; si no sabés el tipo, usá `unknown` y estrechá con guards.
- **`unknown` vs `any`**: `unknown` obliga a verificar antes de usar (seguro); `any` desactiva todo (peligroso).
- Aprovechá la **inferencia**: no anotes lo obvio, pero tipá fronteras públicas (params, retornos, APIs).
- **Shape/estructural**: TS tipa por forma, no por nombre. Dos objetos con la misma forma son compatibles.

## Herramientas de tipado
- **union** (`A | B`) e **intersección** (`A & B`) para modelar estados válidos y combinaciones.
- **`as const`** para literales inmutables y derivar tipos exactos de valores.
- **Generics** para reutilización con seguridad de tipos; usar constraints (`<T extends ...>`) y `keyof`/`in` para tipos derivados.
- **Utility types** (`Partial`, `Pick`, `Omit`, `Record`, `Readonly`, `ReturnType`...) antes de escribir tipos a mano.
- **Enums**: preferir enums de string o `as const` objects sobre enums numéricos (más legibles y seguros).

## Cuidados
- **Type assertion (`as`)** es una promesa al compilador, no una conversión: úsala poco y con certeza. No la uses para callar errores reales.
- Modelá **estados inválidos como imposibles** con union types (evita banderas booleanas contradictorias).

## Regla dura
Un `any` o un `as` para "que compile" es deuda que un juez encontrará después. Tipá bien desde el inicio: el tipo es la primera prueba.
