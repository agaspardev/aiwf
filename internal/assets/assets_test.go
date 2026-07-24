package assets

import (
	"testing"

	"github.com/agaspardev/aiwf/internal/overlay"
)

func TestBuildEntries(t *testing.T) {
	entries, err := BuildEntries()
	if err != nil {
		t.Fatalf("BuildEntries: %v", err)
	}
	if len(entries) == 0 {
		t.Fatal("BuildEntries no devolvió entries")
	}

	byPath := make(map[string]overlay.Entry, len(entries))
	for _, e := range entries {
		byPath[e.Path] = e
	}

	// CLAUDE.md: estrategia import → archivo propio OWNED + bloque en el compartido.
	if e, ok := byPath["CLAUDE.aiwf.md"]; !ok || e.Type != overlay.Owned {
		t.Error("falta CLAUDE.aiwf.md como OWNED")
	}
	if e, ok := byPath["CLAUDE.md"]; !ok || e.Type != overlay.MarkerBlock {
		t.Error("falta CLAUDE.md como MarkerBlock (import)")
	}
	// settings.json compartido → JSONMerge.
	if e, ok := byPath["settings.json"]; !ok || e.Type != overlay.JSONMerge {
		t.Error("falta settings.json como JSONMerge")
	}
	// Contenido propio OWNED en ruta espejo (new packaged format).
	if e, ok := byPath["skills/aiwf-init/SKILL.md"]; !ok || e.Type != overlay.Owned {
		t.Error("falta skills/aiwf-init/SKILL.md como OWNED")
	}
	if e, ok := byPath["agents/security-reviewer.md"]; !ok || e.Type != overlay.Owned {
		t.Error("falta un agent propio como OWNED")
	}

	// Ningún path debe filtrar el prefijo de embed.
	for _, e := range entries {
		if len(e.Path) == 0 || e.Path[0] == '/' {
			t.Errorf("path inválido: %q", e.Path)
		}
	}
}
