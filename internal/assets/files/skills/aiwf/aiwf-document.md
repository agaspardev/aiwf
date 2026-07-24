# Skill: /aiwf-document [full|update]

Documenta un proyecto de forma **determinista y sin gastar tokens**, y lo deja como
segundo cerebro local para la IA. Pensado para proyectos legacy o sin documentar.

## Cuándo usar (automático)

- **Al empezar a trabajar en un proyecto que no tiene** `.claude/knowledge/context-pack.md` —
  córrelo ANTES de explorar el repo a mano (P8: el código extrae, no el agente escaneando).
- Cuando el usuario pide "documentá el proyecto", "qué hace esto", "levantá la arquitectura".
- Tras un cambio de arquitectura o de flujos: `ai document update` (registra la evolución).

## Comportamiento

1. **P8 gate**: si ya existe `context-pack.md` fresco, léelo en vez de re-escanear.
2. Ejecutar `ai document full` (primera vez) o `ai document update` (posteriores).
   - Extracción **determinista, cero tokens**: stack, estructura, dependencias, entrypoints,
     conexiones externas (redactadas), CodeGraph si está, metadatos de git.
   - Escribe LOCAL y excluido de git: `context-pack.md` (índice AI-first), `ARCHITECTURE.md`
     (detalle humano), reporte crudo en `.ai-workflow/evidence/document/`.
3. **Leer el `context-pack.md`** generado e incorporarlo al contexto — es el índice curado
   que reemplaza escanear el repo (ahorro de tokens por diseño).
4. **Curar** `ARCHITECTURE.md`: el análisis determinista no infiere intención ni decisiones.
   Completar "Decisiones y porqués" con lo que sepas o preguntar al usuario.
5. Si el usuario quiere prosa sintetizada: `ai document -Synthesize` (opt-in) — **rutea a
   OmniRoute** (combo `agent-daily`), nunca IA local.

## Control de evolución

`ai document update` archiva la `ARCHITECTURE.md` previa en `.claude/knowledge/history/`
y enlaza `supersedes`. La historia de versiones vive **local** (el repo es del cliente, no
se versiona en git). Mantener actualizado tras cambios de arquitectura/flujo es parte de
cerrar el trabajo.

## Restricciones

- NADA se versiona en git. Todo va a `.claude/` y `.ai-workflow/`, ya excluidos por `init`.
- Ninguna IA local (sin Ollama). Cualquier paso con LLM va por OmniRoute.
- El reporte nunca incluye valores de secretos — solo conteos y patrones redactados.
