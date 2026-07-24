---
name: rule-angular
trigger: al escribir o revisar Angular (componentes, servicios, DI, testing)
applies_to: ["**/*.component.ts", "**/*.service.ts", "**/*.module.ts", "angular.json"]
type: foundation
---

# Regla: Angular (dominando el framework)

Del libro Gentleman Programming ("Angular: Dominando el Framework").

## Fundamentos
- Angular es **TypeScript-first**: aprovechá tipado fuerte, decoradores y DI en todo.
- **Inyección de dependencias**: servicios inyectados, no instanciados a mano. Facilita test y desacople.
- **Control flow moderno** (`@if`, `@for`, `@switch`) sobre las viejas directivas estructurales; `@for` requiere `track`.
- Preferí **componentes standalone** y señales (signals) donde el proyecto lo permita.

## Datos y SSR
- **Interceptores HTTP** para auth, errores y logging centralizados (no repetir en cada llamada).
- Considerá **SSR** para SEO y first paint; cuidá el estado que viaja server→client.

## Testing (calidad desde el inicio)
- **Jest** para unit; **Playwright** para e2e. Buenas prácticas: manejar `await` correctamente, mockear dependencias y simular entornos.
- En Playwright: mock de red y manejo explícito de asincronía; no `sleep` arbitrarios.

## Regla dura
Lógica de negocio en servicios (testeables), no en componentes. El componente orquesta y presenta; el servicio decide.
