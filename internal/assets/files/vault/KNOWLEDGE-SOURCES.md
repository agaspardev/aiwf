# Fuentes de conocimiento para los agentes

Esta nota es el **registro operativo** de los alcances que un agente puede consultar.
Los paths concretos se configuran en `.ai-workflow/env/vault-config.local.json` — esta nota
documenta los scopes posibles y las reglas de acceso, no las rutas absolutas (que son personales).

| Alcance | Configuración | Contenido | Cuándo consultar | Regla de acceso |
|---|---|---|---|---|
| `operational` | `vaultPaths.operational` en vault-config.local.json | Workflow, guías del entorno, decisiones AI | Cuando la tarea trate del entorno de IA | Permitido por defecto |
| `personal` | `vaultPaths.personal` | Hábitos, crecimiento y conocimiento personal | Solo si el tema es personal | Privado; consulta acotada |
| `learning` | `vaultPaths.learning` | Certificaciones, investigación técnica | Aprendizaje, herramientas, estudio | Consulta acotada |
| `work` | `vaultPaths.work` | Documentación laboral | Solo para una tarea laboral relacionada o por orden explícita | **Restringido**; declarar en `restrictedScopes` |

## Cómo configurar los vaultPaths

Editar `.ai-workflow/env/vault-config.local.json` en cada instancia instalada:

```json
{
  "vaultPaths": {
    "operational": "~/.ai-workflow/vault",
    "personal": "/ruta/a/tu/vault/personal",
    "learning": "/ruta/a/tu/vault/learning"
  },
  "restrictedScopes": ["work"]
}
```

## Cómo recuperar información

1. Elegir el alcance más pequeño que pueda responder la tarea.
2. Leer el `index.md` del scope (generado por `vault-index.ps1`) para elegir la nota exacta.
3. Si el index no es suficiente, buscar con `scripts/vault-search.ps1 -Query "..." -Scope <scope>`.
4. Leer solo las notas encontradas que sean necesarias.
5. Priorizar notas `verified`; tratar `draft` como hipótesis, `stale` con advertencia y `superseded` como historial.
6. Validar los datos que puedan haber cambiado contra `source_of_truth`.

## Vigencia y actualización

- Estar indexado NO significa estar vigente.
- `verified_at` registra la última comprobación y `review_after` su ventana de revisión.
- Al vencer `review_after`, la información se trata como `stale` hasta revalidarla.
- Una afirmación reemplazada se marca `superseded` y enlaza bidireccionalmente mediante `supersedes` y `superseded_by`.
- Las reglas completas están en `vault/KNOWLEDGE-ARCHITECTURE.md`.

## Regla de seguridad

Un vault no es permiso para divulgar su contenido. Las notas personales, laborales, datos de
clientes, secretos y transcripciones permanecen dentro de su alcance. Un permiso para copiar
datos o calcular hashes no equivale a permiso para analizar contenido.
