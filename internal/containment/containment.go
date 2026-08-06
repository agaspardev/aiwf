// Package containment verifica de forma determinista (cero tokens, cero red) que
// los artefactos de trabajo vivan bajo .ai-workflow/. Es el enforcement de la
// regla de contención que F1 fijó como prompt invariante (governanceCore).
package containment

import (
	"io/fs"
	"path/filepath"
	"strings"
)

// Violation es un artefacto de trabajo hallado fuera de .ai-workflow/.
type Violation struct {
	Path string // relativa a root, POSIX
	Kind string // "dir" | "file"
}

// forbiddenDirs son directorios de trabajo que solo pueden vivir bajo .ai-workflow/.
var forbiddenDirs = map[string]bool{
	"scratch": true, "notes": true, "evidence": true, "reports": true,
	"screenshots": true, "playwright-report": true, "coverage": true, "handoffs": true,
}

// skipDirs son raíces que el scan nunca recorre.
var skipDirs = map[string]bool{
	".git": true, ".ai-workflow": true, "node_modules": true, "dist": true, "omniroute": true,
}

// forbiddenFiles son nombres exactos de archivo-artefacto.
var forbiddenFiles = map[string]bool{
	"coverage.out": true, "lcov.info": true,
}

// forbiddenExts son extensiones de archivo-artefacto.
var forbiddenExts = map[string]bool{
	".cov": true, ".log": true,
}

// Scan recorre root y devuelve los artefactos de trabajo fuera de .ai-workflow/.
// allow exime rutas relativas (POSIX) puntuales. Determinista y sin I/O de red.
func Scan(root string, allow map[string]bool) ([]Violation, error) {
	var violations []Violation
	err := filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		rel, relErr := filepath.Rel(root, path)
		if relErr != nil {
			return relErr
		}
		if rel == "." {
			return nil
		}
		rel = filepath.ToSlash(rel)
		name := d.Name()

		if d.IsDir() {
			if skipDirs[name] {
				return fs.SkipDir
			}
			if forbiddenDirs[name] && !allow[rel] {
				violations = append(violations, Violation{Path: rel, Kind: "dir"})
				return fs.SkipDir // no doble-reportar el contenido
			}
			return nil
		}

		if allow[rel] {
			return nil
		}
		if forbiddenFiles[name] || forbiddenExts[strings.ToLower(filepath.Ext(name))] {
			violations = append(violations, Violation{Path: rel, Kind: "file"})
		}
		return nil
	})
	return violations, err
}
