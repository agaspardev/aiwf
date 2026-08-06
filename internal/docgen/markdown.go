package docgen

import (
	"fmt"
	"sort"
	"strings"
	"time"
)

func joinOr(items []string, empty string) string {
	if len(items) == 0 {
		return empty
	}
	return strings.Join(items, "\n")
}

func dirLines(dirs []DirCount) []string {
	out := make([]string, len(dirs))
	for i, d := range dirs {
		out[i] = fmt.Sprintf("  - `%s/` (%d)", d.Dir, d.Count)
	}
	return out
}

func extList(exts []ExtCount) string {
	parts := make([]string, len(exts))
	for i, e := range exts {
		parts[i] = fmt.Sprintf("%s (%d)", e.Ext, e.Count)
	}
	return strings.Join(parts, ", ")
}

func entryLines(eps []string, max int) []string {
	if len(eps) > max {
		eps = eps[:max]
	}
	out := make([]string, len(eps))
	for i, e := range eps {
		out[i] = "  - `" + e + "`"
	}
	return out
}

func connLines(conns map[string]int) []string {
	// orden estable
	names := make([]string, 0, len(conns))
	for k := range conns {
		names = append(names, k)
	}
	sort.Strings(names)
	out := make([]string, len(names))
	for i, n := range names {
		out[i] = fmt.Sprintf("  - %s : %d referencias", n, conns[n])
	}
	return out
}

func contextPack(rep *Report, today string) string {
	return fmt.Sprintf(`---
title: Context Pack — %s
type: guide
status: generated
updated: %s
verified_at: %s
summary: Índice curado del proyecto para consulta rápida de la IA (progressive disclosure).
---

# Context Pack — %s

> GENERADO por `+"`aiwf document`"+` el %s (determinista, cero tokens). LOCAL — no versionar.
> La IA lee ESTE archivo primero, en vez de escanear el repo. Para detalle: ARCHITECTURE.md.
> Para navegación de símbolos: %s.

## Qué es (en una línea)
Stack: **%s** · %d archivos · %d dependencias directas.

## Estructura (top-level)
%s

Extensiones dominantes: %s

## Entrypoints
%s

## Conexiones externas detectadas
%s

## Cómo profundizar (progressive disclosure — buscar barato, profundizar caro)
1. Este context-pack (índice).
2. `+"`ARCHITECTURE.md`"+` del knowledge scope explícito (detalle + evolución).
3. `+"`evidence/document/report-*.json`"+` del change owner (reporte crudo completo).
4. CodeGraph / grep para símbolos puntuales.
`,
		rep.Project, today, today, rep.Project, today, rep.CodeGraph,
		strings.Join(rep.Stack, ", "), rep.FilesTotal, rep.DepsTotal,
		joinOr(dirLines(rep.TopDirs), "  - (sin subdirectorios)"),
		extList(rep.ByExtension),
		joinOr(entryLines(rep.Entrypoints, 12), "  - (no detectados automáticamente)"),
		joinOr(connLines(rep.External), "  - (ninguna detectada por patrón)"),
	)
}

func architecture(rep *Report, today, ts string) string {
	reviewAfter := time.Now().AddDate(0, 3, 0).Format("2006-01-02")

	depsTop := rep.Dependencies
	if len(depsTop) > 30 {
		depsTop = depsTop[:30]
	}
	depLines := make([]string, len(depsTop))
	for i, d := range depsTop {
		depLines[i] = "- " + d
	}

	git := rep.Git
	gitLine := fmt.Sprintf("rama %s, %s commits, %s colaboradores",
		git["branch"], git["commits"], git["contributors"])

	return fmt.Sprintf(`---
title: Arquitectura — %s
type: source
domain: project
status: draft
updated: %s
verified_at: %s
review_after: %s
source_of_truth: [codebase, "aiwf document %s"]
supersedes: []
summary: Arquitectura derivada deterministamente del código. Revisar y curar.
---

# Arquitectura — %s

_Derivado por `+"`aiwf document`"+` el %s. Editá/curá lo que el análisis determinista no capta._

## Stack y tamaño
- Lenguajes/frameworks: %s
- Archivos: %d · Dependencias directas: %d
- Git: %s

## Módulos principales
%s

## Entrypoints
%s

## Dependencias clave (top 30)
%s

## Conexiones externas
%s

## Navegación de código
%s

## Decisiones y porqués
_(Curado humano — el análisis determinista no infiere intención. Registrar aquí ADRs y tradeoffs.)_
`,
		rep.Project, today, today, reviewAfter, ts, rep.Project, today,
		strings.Join(rep.Stack, ", "), rep.FilesTotal, rep.DepsTotal, gitLine,
		joinOr(dirLines(rep.TopDirs), "- (sin subdirectorios)"),
		joinOr(entryLines(rep.Entrypoints, 12), "- (no detectados)"),
		joinOr(depLines, "- (sin dependencias detectadas)"),
		joinOr(connLines(rep.External), "- (ninguna detectada por patrón)"),
		rep.CodeGraph,
	)
}
