---
name: rule-algorithms
trigger: al escribir lógica con bucles, búsquedas, ordenamientos o estructuras de datos con volumen
applies_to: ["**/*"]
type: foundation
---

# Regla: Algoritmos y complejidad (Big O)

Del libro Gentleman Programming ("Algoritmos a la Manera Caballerosa").

## Conciencia de complejidad (Big O)
- Antes de escribir un bucle sobre datos, preguntá el **costo**: O(1), O(log n), O(n), O(n log n), O(n²).
- **Evitá O(n²) accidental**: bucles anidados o `array.includes/find` dentro de un loop sobre otro array grande.

## Estructuras y búsquedas correctas
- **Búsqueda lineal** O(n): datos sin orden o chicos. **Búsqueda binaria** O(log n): datos **ordenados**.
- Para lookups repetidos usá **Set/Map** (O(1)) en vez de `array.includes`/`find` (O(n)) dentro de un loop.
- Elegí la estructura por el acceso dominante: índice → array; pertenencia/clave → Set/Map; orden → estructura ordenada.

## Ordenamiento
- Conocé el costo del sort del lenguaje (típico O(n log n)); no reimplementes bubble sort O(n²) salvo aprendizaje.

## Regla dura
El algoritmo correcto se elige en el DISEÑO, no se "optimiza" después cuando el juez marca lentitud. Si el input crece, la complejidad importa.
