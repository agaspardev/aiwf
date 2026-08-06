# Sistema de Plugins en OmniRoute v3.8.48

## Stack

- **Runtime**: Node.js child_process (spawn) por plugin
- **Sandbox**: VM (worker_threads) sin `require()` salvo `crypto`
- **Persistencia**: SQLite (`storage.sqlite`), tabla `plugins`
- **Lifecycle**: 4 fases — install → load → activate → (opcional) persist

## Estructura de un Plugin

```
~/.omniroute/plugins/<name>/
├── plugin.json       # Manifest obligatorio
└── index.js          # Entry point (NODE_PATH incluye omniroute y deps)
```

### plugin.json

```json
{
  "name": "rate-limiter",
  "version": "1.0.0",
  "description": "Rate limiting per model using sliding window",
  "hooks": {
    "onRequest": true,
    "onResponse": true
  },
  "configSchema": {
    "type": "object",
    "properties": {
      "maxRequests": { "type": "number", "default": 20 }
    }
  },
  "entry": "index.js",
  "author": ""
}
```

### index.js — API

El `index.js` exporta un objeto con hooks. Cada hook recibe un `ctx` con setters:

```js
module.exports = {
  onRequest(ctx) {
    // Interceptar request. Llamar ctx.blockRequest(msg) para rechazar.
  },
  onResponse(ctx) {
    // Interceptar response. No puede bloquear.
  },
  onError(ctx) {
    // Interceptar error. Llamar ctx.blockRequest(msg) para sobreescribir.
  }
};
```

### ctx API

| Método | Descripción |
|--------|-------------|
| `ctx.blockRequest(msg)` | Bloquea/rechaza el request con mensaje |
| `ctx.setHeader(k, v)` | Setea header en la response |
| `ctx.getRequest()` | Objeto request singletons (model, messages, etc.) |
| `ctx.getResponse()` | Objeto response parcial |
| `ctx.getMetadata()` | Metadata parsed |
| `ctx.getLogger()` | Logger scoped al plugin |

## Hooks Disponibles

| Hook | Fase | Blocking | Uso |
|------|------|----------|-----|
| `onRequest` | Antes de procesar | Sí | Rate limiting, auth, validación |
| `onResponse` | Después de generar | No | Logging, tracking de costos |
| `onError` | En error | Sí | Error handling custom |

## Plugins del Repo

### rate-limiter

- **Sliding window** por modelo usando Set con timestamps
- Limpia ventanas viejas cada request
- Bloquea con 429 si excede `maxRequests` (default: 20/min)
- Hooks: `onRequest`

### cost-tracker

- Acumula costos por modelo en Map en memoria
- Calcula costo del request usando `ctx.getMetadata().cost`
- No persiste (se pierde en reinicio)
- Hooks: `onResponse`

### request-logger

- Mide timing (inicio en `onRequest`, duración en `onResponse`)
- Acumula stats en memoria (count, avg duration/min)
- Hooks: `onRequest`, `onResponse`

## API Endpoints

### `POST /api/plugins/scan`
**Descubre** plugins nuevos en `~/.omniroute/plugins/`. Los inserta en DB con status `installed`. Requiere restart para activación automática o activación manual via API.

### `POST /api/plugins/<name>/activate`
**Activa** un plugin en runtime: spawn child_process, carga `index.js`, registra hooks en el hook registry. Persiste `status=active` en DB.

### `GET /api/plugins/marketplace`
**Seed data** generado estáticamente. En v3.8.48 el marketplace no es funcional — los `downloadUrl` vienen vacíos (`""`). No usar.

### `GET /api/plugins`
**Devuelve 400** consistentemente si no se pasa query param correcto. No usar. Leer DB directamente en su lugar.

### `POST /api/plugins/<name>/deactivate`
**Desactiva** un plugin: quita hooks del registry, marca `status=installed`.

## Notas Críticas

1. **No existe marketplace funcional** — `GET /api/plugins/marketplace` devuelve seed data con `downloadUrl: ""`. Instalación es 100% por copia filesystem a `~/.omniroute/plugins/`.

2. **`omniroute plugin install <name>`** instala desde npm (busca `@omniroute/plugin-<name>`), no desde marketplace. NO usar para instalar plugins locales.

3. **Activación via API** — Usar .NET `WebRequest` en PowerShell (ver `setup-plugins.ps1`). `Invoke-RestMethod` de PowerShell tiene problemas con el stream de respuesta del endpoint de activación.

4. **Persistencia** — `status=active` se guarda en DB. `loadAll()` en startup carga plugins con status `active`. En fase sin restart, activar via API es necesario.

5. **Sandbox restringido** — `index.js` corre en una VM sin `require()`. Sólo `crypto` está disponible globalmente. No se puede importar módulos npm.

6. **No hay reload en caliente** — Para recargar el código de un plugin, hay que deactivar y reactivar, o restartear OmniRoute.

7. **Path separator bug** — En la validación de plugin path, `pluginDir.startsWith(r + "/")` usa `/` hardcoded. En Windows esto puede fallar si realpath no normaliza separadores. Los plugins se cargan igual porque el check usa `path.resolve()` antes.
