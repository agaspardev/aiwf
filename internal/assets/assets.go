// Package assets embebe el contenido PROPIO de aiwf (skills aiwf/rules, agents,
// harness, vault, security, templates) y lo traduce a entries del overlay. NO incluye
// contenido de gentle-ai ni de terceros: eso lo provee el instalador de gentle-ai.
package assets

import (
	"embed"
	"io/fs"
	"strings"

	"github.com/agaspardev/aiwf/internal/overlay"
)

//go:embed all:files
var files embed.FS

// claudePersonaImport es la línea que se inyecta en el CLAUDE.md compartido para
// referenciar nuestro persona propio sin sobrescribir el de gentle-ai.
const claudePersonaImport = "@CLAUDE.aiwf.md"

// ReadSecurityPolicy devuelve el contenido del block-warn-ignore.yaml embebido.
func ReadSecurityPolicy() ([]byte, error) {
	return files.ReadFile("files/security/policies/block-warn-ignore.yaml")
}

// BuildEntries recorre los assets embebidos y produce las entries del overlay,
// clasificando cada archivo por su ruta de deploy.
func BuildEntries() ([]overlay.Entry, error) {
	var entries []overlay.Entry
	err := fs.WalkDir(files, "files", func(p string, d fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if d.IsDir() {
			return nil
		}
		data, err := files.ReadFile(p)
		if err != nil {
			return err
		}
		// Ruta de deploy relativa a la raíz de instalación.
		rel := strings.TrimPrefix(p, "files/")

		switch rel {
		case "templates/CLAUDE.md":
			// Estrategia import: nuestro persona en archivo propio + línea de import
			// inyectada en el CLAUDE.md compartido (bloque marcado, idempotente).
			entries = append(entries,
				overlay.Entry{Path: "CLAUDE.aiwf.md", Type: overlay.Owned, Payload: data},
				overlay.Entry{Path: "CLAUDE.md", Type: overlay.MarkerBlock, Payload: []byte(claudePersonaImport)},
			)
		case "templates/settings.json":
			// settings compartido: deep-merge sin pisar lo de gentle-ai.
			entries = append(entries, overlay.Entry{Path: "settings.json", Type: overlay.JSONMerge, Payload: data})
		default:
			// Todo lo demás es OWNED en su ruta espejo (skills, agents, harness,
			// vault, security, plantillas por-proyecto).
			entries = append(entries, overlay.Entry{Path: rel, Type: overlay.Owned, Payload: data})
		}
		return nil
	})
	return entries, err
}
