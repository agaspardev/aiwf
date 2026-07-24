---
name: rule-clean-code-agile
trigger: siempre — al planificar, escribir código, definir tareas/historias o tests
applies_to: ["**/*"]
type: foundation
---

# Regla: Código Limpio y Agilidad (fundamento transversal)

Del libro Gentleman Programming. Se aplica en TODA tarea, desde la planificación.

## Diseño atómico (pensar atómicamente)
- Separá el código en la **mínima unidad de lógica** posible: Átomo → Molécula → Organismo → Plantilla → Página → App.
- Unidades chicas = **testables, reutilizables, mantenibles, más performantes**. Acoplamiento bajo, cohesión alta.
- Si algo no es atómico, es casi imposible de testear (todo termina acoplado).

## Programación funcional
- Métodos **declarativos, sin efectos secundarios** (funciones puras: mismos parámetros → mismo resultado, no muta nada externo).
- Preferir `map/filter/find/reduce` sobre bucles imperativos con estado mutable.

## TDD (calidad desde el inicio, NO al final)
- Flujo: ¿qué quiero? → requisitos/casos de uso → **escribir tests que fallen** → codificar sabiendo qué querés → hacer pasar → refactorizar.
- Los tests derivan de los **casos de uso** de la historia. Testear comportamiento, no implementación.

## Historia de usuario
- Formato: **Como (quién) quiero (qué) para (por qué)** + casos de uso (cómo: paso 1, 2, 3...).
- Escribirla puede revelar que la feature no tiene sentido (objetivo poco claro). Incluir camino feliz y casos borde acotados.

## Definición de "listo" (DoD) y valor
- Toda feature debe tener **inicio y fin claros** y **aportar valor por sí sola** (independiente).
- Entregar la **pieza de valor más pequeña** posible → feedback rápido (Agile es mentalidad, no ceremonias).
- Preferí **simple y elegante** sobre complejo y hermoso: los requisitos cambian; iterá después.

## Fuentes de verdad (comunicación)
- Dos tipos: (1) realidad del negocio (Notion/Confluence: qué, por qué, requisitos, casos especiales); (2) carga de trabajo (tickets).
- Las reuniones NO agregan contexto nuevo: **actualizá la fuente de verdad** y compartila. Documentá decisiones y su porqué.
- Refactorizá siempre (la deuda técnica es inevitable; "si funciona no lo toques" lleva a legacy).

## Calidad de código = confiabilidad, mantenibilidad, testabilidad, portabilidad, reutilización
Primero alcanzá el objetivo; luego, con tiempo, subí calidad. El código COMUNICA: escribí como si otro lo leyera (aunque trabajes solo).
