package overlay

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

// FileType define cómo aiwf aplica un archivo.
type FileType int

const (
	// Owned: archivo exclusivo de aiwf; se escribe/reemplaza directo.
	Owned FileType = iota
	// MarkerBlock: archivo compartido; el payload se inyecta en un bloque marcado.
	MarkerBlock
	// JSONMerge: archivo compartido JSON; el payload se fusiona (deep-merge).
	JSONMerge
)

// Entry describe un archivo gestionado por el overlay.
type Entry struct {
	// Path es relativo a la raíz de instalación (p. ej. "CLAUDE.md").
	Path string
	Type FileType
	// Payload es el contenido propio: archivo completo (Owned), bloque (MarkerBlock)
	// o fragmento JSON (JSONMerge).
	Payload []byte
}

// Overlay aplica un conjunto de entries sobre una raíz de instalación, registrando
// todo en un manifiesto.
type Overlay struct {
	Root         string
	ManifestPath string
	Entries      []Entry
}

// New construye un Overlay.
func New(root, manifestPath string, entries []Entry) *Overlay {
	return &Overlay{Root: root, ManifestPath: manifestPath, Entries: entries}
}

// Apply aplica todas las entries de forma idempotente y guarda el manifiesto.
// After upserting current entries, it prunes Owned files that were previously
// managed but are no longer in the desired set (orphan cleanup).
// Reconcile es un alias semántico: re-aplica para restaurar la coexistencia tras un
// update de gentle-ai.
func (o *Overlay) Apply() error {
	m, err := LoadManifest(o.ManifestPath)
	if err != nil {
		return fmt.Errorf("cargando manifiesto: %w", err)
	}

	// Build the desired path set before modifying the manifest.
	desired := make(map[string]bool, len(o.Entries))
	for _, e := range o.Entries {
		desired[e.Path] = true
	}

	for _, e := range o.Entries {
		// Para JSONMerge, leer el contenido actual ANTES de aplicar para poder
		// detectar qué keys son nuevas (AddedKeys).
		var existingBefore []byte
		if e.Type == JSONMerge {
			existingBefore = readFileOrEmptyBytes(filepath.Join(o.Root, e.Path))
		}

		if err := o.applyEntry(e); err != nil {
			return fmt.Errorf("aplicando %s: %w", e.Path, err)
		}

		r := Record{
			Path:      e.Path,
			Type:      e.Type,
			Checksum:  checksum(e.Payload),
			AppliedAt: time.Now().UTC(),
		}
		if e.Type == JSONMerge {
			r.OverlayPayload = e.Payload
			// CONGELAR la base: si ya hay un registro previo para esta ruta, el archivo
			// en disco ya contiene lo de aiwf (estamos en un reconcile), así que NO es
			// una base limpia. Preservamos el snapshot y AddedKeys del primer Apply.
			// Sin esto, reconcile colapsaría AddedKeys a [] y el uninstall no revertiría.
			if prev, ok := m.Find(e.Path); ok && prev.Type == JSONMerge {
				r.BaseSnapshot = prev.BaseSnapshot
				r.AddedKeys = prev.AddedKeys
			} else {
				r.BaseSnapshot = existingBefore
				r.AddedKeys = computeAddedKeys(existingBefore, e.Payload)
			}
		}
		m.Upsert(r)
	}

	// Prune: remove Owned files that were previously managed but are no longer
	// shipped. Only Owned entries are eligible — Shared types (MarkerBlock,
	// JSONMerge) are never deleted by prune because their content is
	// interleaved with other tools' data.
	pruned := m.PruneOwned(desired)
	for _, path := range pruned {
		full := filepath.Join(o.Root, path)
		_ = os.Remove(full)
		// Best-effort: remove empty parent dirs up to root.
		removeEmptyParents(o.Root, filepath.Dir(full))
	}

	return m.Save()
}

// Reconcile re-aplica el overlay (idempotente). Se ejecuta tras instalar/actualizar
// gentle-ai para garantizar que las customizaciones sobreviven.
func (o *Overlay) Reconcile() error { return o.Apply() }

// applyEntry aplica una sola entry según su tipo, sin sobrescribir contenido ajeno
// en archivos compartidos.
func (o *Overlay) applyEntry(e Entry) error {
	full := filepath.Join(o.Root, e.Path)
	if err := os.MkdirAll(filepath.Dir(full), 0o755); err != nil {
		return err
	}
	switch e.Type {
	case Owned:
		return os.WriteFile(full, e.Payload, 0o644)
	case MarkerBlock:
		existing := readFileOrEmpty(full)
		merged := InjectBlock(existing, string(e.Payload))
		return os.WriteFile(full, []byte(merged), 0o644)
	case JSONMerge:
		existing := readFileOrEmpty(full)
		merged, err := MergeJSON([]byte(existing), e.Payload)
		if err != nil {
			return err
		}
		return os.WriteFile(full, merged, 0o644)
	default:
		return fmt.Errorf("tipo de archivo desconocido: %d", e.Type)
	}
}

// readFileOrEmpty devuelve el contenido de path, o "" si no existe.
func readFileOrEmpty(path string) string {
	data, err := os.ReadFile(path)
	if err != nil {
		return ""
	}
	return string(data)
}

// readFileOrEmptyBytes es como readFileOrEmpty pero devuelve []byte para usar en
// computeAddedKeys sin reconversión.
func readFileOrEmptyBytes(path string) []byte {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil
	}
	return data
}

// computeAddedKeys devuelve las top-level keys del overlay que NO existen en base.
// Si base está vacío o es JSON inválido, todas las keys del overlay se consideran nuevas.
func computeAddedKeys(baseJSON, overlayJSON []byte) []string {
	baseMap := map[string]any{}
	if len(bytes.TrimSpace(baseJSON)) > 0 {
		json.Unmarshal(baseJSON, &baseMap) // error → todas las keys son nuevas
	}
	overlayMap := map[string]any{}
	json.Unmarshal(overlayJSON, &overlayMap)

	var added []string
	for k := range overlayMap {
		if _, exists := baseMap[k]; !exists {
			added = append(added, k)
		}
	}
	return added
}

// removeEmptyParents removes empty directories between dir and root (exclusive).
// Stops at the first non-empty directory. Best-effort: errors are ignored.
func removeEmptyParents(root, dir string) {
	absRoot, _ := filepath.Abs(root)
	for {
		absDir, _ := filepath.Abs(dir)
		if absDir == absRoot || len(absDir) <= len(absRoot) {
			return
		}
		entries, err := os.ReadDir(dir)
		if err != nil || len(entries) > 0 {
			return
		}
		os.Remove(dir)
		dir = filepath.Dir(dir)
	}
}
