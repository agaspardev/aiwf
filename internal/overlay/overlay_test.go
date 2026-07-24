package overlay

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func manifestPath(root string) string { return filepath.Join(root, ".aiwf", "manifest.json") }

func TestApplyOwnedWritesFile(t *testing.T) {
	root := t.TempDir()
	o := New(root, manifestPath(root), []Entry{
		{Path: "skills/aiwf-doctor/SKILL.md", Type: Owned, Payload: []byte("skill propia")},
	})
	if err := o.Apply(); err != nil {
		t.Fatalf("Apply: %v", err)
	}
	got, err := os.ReadFile(filepath.Join(root, "skills/aiwf-doctor/SKILL.md"))
	if err != nil {
		t.Fatalf("no se escribió el archivo owned: %v", err)
	}
	if string(got) != "skill propia" {
		t.Errorf("contenido = %q", got)
	}
}

// TestReconcileSurvivesGentleAIUpdate es la garantía central del spec: tras un update
// de gentle-ai que reescribe un archivo compartido, la reconciliación restaura la
// customización SIN perder lo de gentle-ai.
func TestReconcileSurvivesGentleAIUpdate(t *testing.T) {
	root := t.TempDir()
	claude := filepath.Join(root, "CLAUDE.md")
	settings := filepath.Join(root, "settings.json")

	// gentle-ai instala su base.
	must(t, os.WriteFile(claude, []byte("# CLAUDE de gentle-ai\nreglas base\n"), 0o644))
	must(t, os.WriteFile(settings, []byte(`{"model":"base","permissions":{"allow":["g"]}}`), 0o644))

	o := New(root, manifestPath(root), []Entry{
		{Path: "CLAUDE.md", Type: MarkerBlock, Payload: []byte("@CLAUDE.aiwf.md")},
		{Path: "settings.json", Type: JSONMerge, Payload: []byte(`{"model":"aiwf-router"}`)},
	})

	// Primera aplicación.
	must(t, o.Apply())
	assertContains(t, claude, "reglas base")
	assertContains(t, claude, "@CLAUDE.aiwf.md")
	assertContains(t, settings, "aiwf-router")
	assertContains(t, settings, `"allow"`) // no se perdió lo de gentle-ai

	// gentle-ai se ACTUALIZA y reescribe ambos archivos (borra lo nuestro).
	must(t, os.WriteFile(claude, []byte("# CLAUDE de gentle-ai v2\nreglas base v2\n"), 0o644))
	must(t, os.WriteFile(settings, []byte(`{"model":"base2","permissions":{"allow":["g2"]}}`), 0o644))

	// Reconciliación: re-aplica el overlay.
	must(t, o.Reconcile())

	// La customización volvió y lo nuevo de gentle-ai sigue intacto.
	assertContains(t, claude, "reglas base v2")
	assertContains(t, claude, "@CLAUDE.aiwf.md")
	assertContains(t, settings, "aiwf-router")
	assertContains(t, settings, "g2")

	// Idempotencia: reconciliar de nuevo no cambia el resultado.
	before, _ := os.ReadFile(claude)
	must(t, o.Reconcile())
	after, _ := os.ReadFile(claude)
	if string(before) != string(after) {
		t.Error("Reconcile no es idempotente sobre CLAUDE.md")
	}
}

func TestManifestRecordsEntries(t *testing.T) {
	root := t.TempDir()
	mp := manifestPath(root)
	o := New(root, mp, []Entry{
		{Path: "a.md", Type: Owned, Payload: []byte("x")},
		{Path: "CLAUDE.md", Type: MarkerBlock, Payload: []byte("y")},
	})
	must(t, o.Apply())

	m, err := LoadManifest(mp)
	if err != nil {
		t.Fatalf("LoadManifest: %v", err)
	}
	if len(m.Records) != 2 {
		t.Fatalf("manifiesto con %d registros, want 2", len(m.Records))
	}
	// Re-aplicar no duplica registros (upsert por path).
	must(t, o.Apply())
	m2, _ := LoadManifest(mp)
	if len(m2.Records) != 2 {
		t.Errorf("tras re-apply hay %d registros, want 2 (upsert)", len(m2.Records))
	}
}

func TestJSONMergeRastreaAddedKeys(t *testing.T) {
	root := t.TempDir()
	mp := manifestPath(root)
	must(t, os.WriteFile(filepath.Join(root, "settings.json"), []byte(`{"a":1,"b":{"c":2}}`), 0o644))
	o := New(root, mp, []Entry{
		{Path: "settings.json", Type: JSONMerge, Payload: []byte(`{"b":{"d":3},"e":4}`)},
	})
	must(t, o.Apply())

	m, err := LoadManifest(mp)
	must(t, err)
	if len(m.Records) != 1 {
		t.Fatalf("manifiesto: %d records", len(m.Records))
	}
	rec := m.Records[0]
	if rec.Type != JSONMerge {
		t.Fatalf("tipo = %d, want JSONMerge", rec.Type)
	}
	// "e" es nueva, "b" ya existe → solo "e" debe estar en AddedKeys.
	if len(rec.AddedKeys) != 1 || rec.AddedKeys[0] != "e" {
		t.Errorf("AddedKeys = %v, want [e] (b ya existía en base)", rec.AddedKeys)
	}
}

func TestComputeAddedKeys(t *testing.T) {
	tests := []struct {
		name    string
		base    string
		overlay string
		want    []string
	}{
		{"nueva key", `{"a":1}`, `{"b":2}`, []string{"b"}},
		{"key existente", `{"a":1}`, `{"a":2}`, nil},
		{"mixto", `{"a":1}`, `{"a":2,"b":3}`, []string{"b"}},
		{"base vacío", ``, `{"a":1}`, []string{"a"}},
		{"overlay vacío", `{}`, `{}`, nil},
		{"base inválido ok", `no-json`, `{"a":1}`, []string{"a"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := computeAddedKeys([]byte(tt.base), []byte(tt.overlay))
			if len(got) != len(tt.want) {
				t.Fatalf("computeAddedKeys = %v, want %v", got, tt.want)
			}
			for i := range got {
				if got[i] != tt.want[i] {
					t.Fatalf("computeAddedKeys = %v, want %v", got, tt.want)
				}
			}
		})
	}
}

// helpers

func must(t *testing.T, err error) {
	t.Helper()
	if err != nil {
		t.Fatal(err)
	}
}

func assertContains(t *testing.T, path, sub string) {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("leyendo %s: %v", path, err)
	}
	if !strings.Contains(string(data), sub) {
		t.Errorf("%s no contiene %q\n---\n%s", filepath.Base(path), sub, data)
	}
}

// ── Prune regression tests ──────────────────────────────────────────────

// TestPruneRemovesOrphanedOwned verifies that Apply removes Owned files
// that were previously managed but are no longer in the desired entries.
func TestPruneRemovesOrphanedOwned(t *testing.T) {
	root := t.TempDir()
	mp := manifestPath(root)

	// First Apply: install two Owned skills.
	o1 := New(root, mp, []Entry{
		{Path: "skills/aiwf/old-skill.md", Type: Owned, Payload: []byte("old")},
		{Path: "skills/aiwf-init/SKILL.md", Type: Owned, Payload: []byte("init")},
	})
	must(t, o1.Apply())

	// Verify both exist.
	assertFileExists(t, filepath.Join(root, "skills/aiwf/old-skill.md"))
	assertFileExists(t, filepath.Join(root, "skills/aiwf-init/SKILL.md"))

	// Second Apply: only the new format skill ships.
	o2 := New(root, mp, []Entry{
		{Path: "skills/aiwf-init/SKILL.md", Type: Owned, Payload: []byte("init v2")},
	})
	must(t, o2.Apply())

	// Old skill should be gone.
	assertFileNotExists(t, filepath.Join(root, "skills/aiwf/old-skill.md"))
	// New skill should be updated.
	got, _ := os.ReadFile(filepath.Join(root, "skills/aiwf-init/SKILL.md"))
	if string(got) != "init v2" {
		t.Errorf("expected 'init v2', got %q", got)
	}

	// Manifest should only have 1 record.
	m, _ := LoadManifest(mp)
	if len(m.Records) != 1 {
		t.Errorf("manifest has %d records, want 1", len(m.Records))
	}
}

// TestPruneNeverTouchesShared verifies that MarkerBlock and JSONMerge entries
// are never pruned, even if they are not in the desired set.
func TestPruneNeverTouchesShared(t *testing.T) {
	root := t.TempDir()
	mp := manifestPath(root)

	// First Apply: includes Shared types.
	must(t, os.WriteFile(filepath.Join(root, "CLAUDE.md"), []byte("base"), 0o644))
	must(t, os.WriteFile(filepath.Join(root, "settings.json"), []byte(`{"a":1}`), 0o644))

	o1 := New(root, mp, []Entry{
		{Path: "CLAUDE.md", Type: MarkerBlock, Payload: []byte("@CLAUDE.aiwf.md")},
		{Path: "settings.json", Type: JSONMerge, Payload: []byte(`{"b":2}`)},
		{Path: "skills/old.md", Type: Owned, Payload: []byte("old")},
	})
	must(t, o1.Apply())

	// Second Apply: only the Owned entry is dropped.
	o2 := New(root, mp, []Entry{
		{Path: "CLAUDE.md", Type: MarkerBlock, Payload: []byte("@CLAUDE.aiwf.md")},
		{Path: "settings.json", Type: JSONMerge, Payload: []byte(`{"b":2}`)},
	})
	must(t, o2.Apply())

	// Shared files must still exist.
	assertFileExists(t, filepath.Join(root, "CLAUDE.md"))
	assertFileExists(t, filepath.Join(root, "settings.json"))
	// Owned orphan must be gone.
	assertFileNotExists(t, filepath.Join(root, "skills/old.md"))

	// Even if we drop Shared from entries, they must NOT be pruned.
	o3 := New(root, mp, []Entry{})
	must(t, o3.Apply())
	assertFileExists(t, filepath.Join(root, "CLAUDE.md"))
	assertFileExists(t, filepath.Join(root, "settings.json"))

	// Manifest still has the Shared records.
	m, _ := LoadManifest(mp)
	sharedCount := 0
	for _, r := range m.Records {
		if r.Type == MarkerBlock || r.Type == JSONMerge {
			sharedCount++
		}
	}
	if sharedCount != 2 {
		t.Errorf("shared records = %d, want 2", sharedCount)
	}
}

// TestPruneTransactional verifies that if Apply fails mid-way, the previous
// manifest remains valid (prune happens before Save).
func TestPruneTransactional(t *testing.T) {
	root := t.TempDir()
	mp := manifestPath(root)

	// Install one Owned file.
	o1 := New(root, mp, []Entry{
		{Path: "skills/keep.md", Type: Owned, Payload: []byte("keep")},
	})
	must(t, o1.Apply())

	// Verify manifest has 1 record.
	m1, _ := LoadManifest(mp)
	if len(m1.Records) != 1 {
		t.Fatalf("initial manifest: %d records", len(m1.Records))
	}

	// Second Apply with an entry that will fail (invalid type).
	o2 := New(root, mp, []Entry{
		{Path: "skills/keep.md", Type: Owned, Payload: []byte("keep v2")},
		{Path: "bad.md", Type: FileType(99), Payload: []byte("boom")},
	})
	err := o2.Apply()
	if err == nil {
		t.Fatal("expected error from invalid type")
	}

	// Manifest should not have been updated (Save was never called).
	m2, _ := LoadManifest(mp)
	if len(m2.Records) != 1 {
		t.Errorf("after failed Apply: %d records, want 1 (unchanged)", len(m2.Records))
	}
}

// TestPruneCleansEmptyDirs verifies that after pruning files, empty parent
// directories are cleaned up.
func TestPruneCleansEmptyDirs(t *testing.T) {
	root := t.TempDir()
	mp := manifestPath(root)

	// Install a deeply nested Owned file.
	o1 := New(root, mp, []Entry{
		{Path: "skills/aiwf/sub/deep.md", Type: Owned, Payload: []byte("deep")},
		{Path: "keep.md", Type: Owned, Payload: []byte("stay")},
	})
	must(t, o1.Apply())
	assertFileExists(t, filepath.Join(root, "skills/aiwf/sub/deep.md"))

	// Second Apply: remove the deep file.
	o2 := New(root, mp, []Entry{
		{Path: "keep.md", Type: Owned, Payload: []byte("stay")},
	})
	must(t, o2.Apply())

	// File should be gone.
	assertFileNotExists(t, filepath.Join(root, "skills/aiwf/sub/deep.md"))
	// Empty parents should be cleaned.
	if _, err := os.Stat(filepath.Join(root, "skills/aiwf/sub")); !os.IsNotExist(err) {
		t.Error("expected skills/aiwf/sub to be removed")
	}
	if _, err := os.Stat(filepath.Join(root, "skills/aiwf")); !os.IsNotExist(err) {
		t.Error("expected skills/aiwf to be removed")
	}
	if _, err := os.Stat(filepath.Join(root, "skills")); !os.IsNotExist(err) {
		t.Error("expected skills/ to be removed")
	}
	// Root itself must not be removed.
	if _, err := os.Stat(root); os.IsNotExist(err) {
		t.Error("root should still exist")
	}
}

// TestSkillMigrationPrune simulates the real migration: old flat skills are
// replaced by new packaged skills, and the old paths are cleaned up.
func TestSkillMigrationPrune(t *testing.T) {
	root := t.TempDir()
	mp := manifestPath(root)

	// Simulate current state: 3 old-format skills.
	oldEntries := []Entry{
		{Path: "skills/aiwf/aiwf-init.md", Type: Owned, Payload: []byte("old init")},
		{Path: "skills/aiwf/aiwf-audit.md", Type: Owned, Payload: []byte("old audit")},
		{Path: "skills/aiwf/aiwf-doctor.md", Type: Owned, Payload: []byte("old doctor")},
	}
	o1 := New(root, mp, oldEntries)
	must(t, o1.Apply())

	// New Apply: migrated format (2 skills, doctor curated out).
	newEntries := []Entry{
		{Path: "skills/aiwf-init/SKILL.md", Type: Owned, Payload: []byte("new init")},
		{Path: "skills/aiwf-audit/SKILL.md", Type: Owned, Payload: []byte("new audit")},
	}
	o2 := New(root, mp, newEntries)
	must(t, o2.Apply())

	// All old paths should be gone.
	assertFileNotExists(t, filepath.Join(root, "skills/aiwf/aiwf-init.md"))
	assertFileNotExists(t, filepath.Join(root, "skills/aiwf/aiwf-audit.md"))
	assertFileNotExists(t, filepath.Join(root, "skills/aiwf/aiwf-doctor.md"))

	// New paths should exist.
	assertFileExists(t, filepath.Join(root, "skills/aiwf-init/SKILL.md"))
	assertFileExists(t, filepath.Join(root, "skills/aiwf-audit/SKILL.md"))

	// Manifest should only have the 2 new records.
	m, _ := LoadManifest(mp)
	if len(m.Records) != 2 {
		t.Errorf("manifest has %d records, want 2", len(m.Records))
	}
}

// ── Test helpers ─────────────────────────────────────────────────────────

func assertFileExists(t *testing.T, path string) {
	t.Helper()
	if _, err := os.Stat(path); os.IsNotExist(err) {
		t.Errorf("expected %s to exist", path)
	}
}

func assertFileNotExists(t *testing.T, path string) {
	t.Helper()
	if _, err := os.Stat(path); err == nil {
		t.Errorf("expected %s to not exist", path)
	}
}
