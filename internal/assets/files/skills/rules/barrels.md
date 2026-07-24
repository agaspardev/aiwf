---
name: rule-barrels
trigger: al crear archivos index.ts de re-exportación (barrel exports) o estructurar imports de un dominio
applies_to: ["**/index.ts", "**/index.tsx"]
type: foundation
---

# Regla: Barrel Exports (con criterio)

Del libro Gentleman Programming ("Estructurar un Proyecto con Barrel Exports").

## Ventajas (por qué existen)
- Imports limpios: `import { A, B } from '@/features/auth'` en vez de rutas profundas repetidas.
- Encapsulan el API público de un dominio/feature: el resto importa del barrel, no de archivos internos.

## Problemas potenciales (vigilar SIEMPRE)
- **Dependencias circulares**: un barrel que re-exporta módulos que se importan entre sí → ciclos difíciles de debuggear.
- **Tree-shaking / bundle size**: un barrel que junta todo puede arrastrar código no usado si el bundler no lo poda bien.
- **Carga innecesaria**: importar del barrel raíz puede traer más de lo que necesitás.

## Estrategias de mitigación
- **Barrels por dominio específico**, no un barrel gigante en la raíz que re-exporte todo.
- Evitá **barrels de barrels** (index que re-exporta otros index) — multiplican ciclos y peso.
- Exportá solo el **API público** del dominio; lo interno no va al barrel.
- Si hay ciclos o problemas de bundle, **importá directo del archivo** — está bien no usar barrel.

## Regla dura
El barrel es conveniencia, no obligación. Ante ciclo o peso de bundle, el import directo gana.
