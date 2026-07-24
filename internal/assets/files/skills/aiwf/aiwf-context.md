# Skill: /aiwf-context

Reporta el uso de contexto de la sesión actual y advierte sobre riesgo de compactación.

## Comportamiento

Calcular o estimar el uso de contexto de la sesión actual y reportar:

1. **Presupuesto de contexto** (política del contract prompt):
   - Código/archivos: máximo 20% del contexto disponible
   - Resultados de herramientas: máximo 15%
   - Historial de conversación: máximo 18%

2. **Estado actual estimado**:
   - Mensajes en sesión
   - Archivos leídos
   - Llamadas a herramientas recientes

3. **Recomendación**:
   - Si el contexto está bajo riesgo: sugerir `ai seguir` para hacer handoff antes de compactar
   - Si el contexto es moderado: continuar normalmente
   - Si el contexto es bajo: nada que reportar

4. **Acción preventiva**: Si detectas que el contexto supera el 60% del límite, sugerir proactivamente hacer un handoff con `mem_session_summary` antes de que ocurra la compactación forzada.

## Nota

Esta skill es informativa. No puede medir el contexto con precisión exacta — solo estima basado en la actividad visible de la sesión. Para límites precisos, consultar la documentación del modelo activo.
