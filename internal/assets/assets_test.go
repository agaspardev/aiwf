package assets

import (
	"strings"
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

func TestAssetsDoNotReferenceLegacyArtifactRoots(t *testing.T) {
	forbidden := []string{
		".claude/knowledge",
		".ai-workflow/openspec",
		".ai-workflow/state/",
		".ai-workflow/notes/",
		".ai-workflow/handoffs/",
		".ai-workflow/evidence/",
		".ai-workflow/reports/",
		".ai-workflow/scratch/",
		".atl/",
	}
	entries, err := BuildEntries()
	if err != nil {
		t.Fatalf("BuildEntries: %v", err)
	}
	for _, entry := range entries {
		content := string(entry.Payload)
		for _, legacy := range forbidden {
			if strings.Contains(content, legacy) {
				t.Errorf("asset %s referencia ruta legacy %q", entry.Path, legacy)
			}
		}
	}
}

func TestAssetsUseChangeRootForGeneratedOutputs(t *testing.T) {
	entries, err := BuildEntries()
	if err != nil {
		t.Fatalf("BuildEntries: %v", err)
	}
	for _, entry := range entries {
		content := string(entry.Payload)
		for _, output := range []string{"evidence/", "reports/", "handoffs/", "notes/", "scratch/"} {
			if strings.Contains(content, output) && !strings.Contains(content, "${AIWF_CHANGE_ROOT}/"+output) {
				t.Errorf("asset %s menciona %s sin AIWF_CHANGE_ROOT", entry.Path, output)
			}
		}
	}
}
