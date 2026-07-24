# Skill: /aiwf-init

Inicializa el AI Workflow en el proyecto actual.

## Pasos

1. Detectar si el directorio actual es un repositorio git. Si no lo es, preguntar al usuario si desea inicializarlo.
2. Verificar si ya existe `.ai-workflow/` en este directorio. Si existe, indicar que ya está inicializado y ofrecer `--reinit` para sobreescribir.
3. Ejecutar `ai init` (que llama a `scripts/init-project.ps1`) para crear la estructura.
4. Confirmar al usuario qué se creó:
   - `.ai-workflow/` con subdirectorios (state, evidence, handoffs, etc.)
   - `.claude/knowledge/` con ARCHITECTURE.md, DECISIONS.md, CONVENTIONS.md, GOTCHAS.md, LEARNINGS.md
   - `.claude/CLAUDE.md` con reglas de contención
   - `sonar-project.properties` si SonarQube está habilitado
   - `.ai-workflow/` añadido a `.git/info/exclude` (NO a `.gitignore`)
5. Llamar `mem_save` con el hecho de que el proyecto fue inicializado.

## Resultado esperado

El proyecto está listo para sesiones con harness. El usuario puede ejecutar `ai` para abrir Claude Code con el contrato activo.
