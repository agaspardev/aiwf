package overlay

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
)

// Uninstall revierte lo aplicado por aiwf usando el manifiesto, SIN tocar el contenido
// de gentle-ai: borra archivos OWNED, remueve los bloques marcados de archivos
// compartidos, y des-fusiona JSONMerge revirtiendo las claves registradas en AddedKeys.
// Si un JSONMerge no tiene AddedKeys (manifiesto previo a la versión que los rastrea),
// se salta con aviso igual que antes.
func Uninstall(root, manifestPath string) (skippedJSON []string, err error) {
	m, err := LoadManifest(manifestPath)
	if err != nil {
		return nil, err
	}
	for _, r := range m.Records {
		full := filepath.Join(root, r.Path)
		switch r.Type {
		case Owned:
			if rmErr := os.Remove(full); rmErr != nil && !os.IsNotExist(rmErr) {
				return skippedJSON, rmErr
			}
		case MarkerBlock:
			content := readFileOrEmpty(full)
			if content == "" {
				continue
			}
			if wErr := os.WriteFile(full, []byte(RemoveBlock(content)), 0o644); wErr != nil {
				return skippedJSON, wErr
			}
		case JSONMerge:
			switch {
			case len(r.OverlayPayload) > 0:
				// Reversión precisa: resta del archivo exactamente lo que aiwf agregó
				// (claves nuevas, ítems de array) y restaura escalares al valor base.
				if err := uninstallMergePrecise(full, r.BaseSnapshot, r.OverlayPayload); err != nil {
					return skippedJSON, err
				}
			case len(r.AddedKeys) > 0:
				// Manifiesto previo a BaseSnapshot: revertir claves top-level.
				if err := uninstallMerge(full, r.AddedKeys); err != nil {
					return skippedJSON, err
				}
			default:
				// Manifiesto legacy sin datos para revertir: no tocar, avisar.
				skippedJSON = append(skippedJSON, r.Path)
			}
		}
	}
	// Borrar el manifiesto: aiwf ya no gestiona nada.
	if rmErr := os.Remove(manifestPath); rmErr != nil && !os.IsNotExist(rmErr) {
		return skippedJSON, rmErr
	}
	return skippedJSON, nil
}

// uninstallMerge remueve del archivo JSON en full las top-level keys listadas en
// addedKeys. Si addedKeys está vacío o el archivo no existe, es no-op. Errores de
// parseo JSON o escritura se devuelven.
func uninstallMerge(full string, addedKeys []string) error {
	if len(addedKeys) == 0 {
		return nil
	}
	content := readFileOrEmptyBytes(full)
	if len(content) == 0 {
		return nil
	}
	current := map[string]any{}
	if err := json.Unmarshal(content, &current); err != nil {
		return nil // archivo no JSON o corrupto — no podemos des-fusionar
	}
	for _, k := range addedKeys {
		delete(current, k)
	}
	cleaned, err := json.MarshalIndent(current, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(full, cleaned, 0o644)
}

// uninstallMergePrecise revierte del archivo en full exactamente lo que aiwf agregó,
// comparando la base congelada (baseSnap) con el payload que aiwf fusionó. Restaura
// arrays (quita solo los ítems que aiwf sumó) y escalares (al valor base), y elimina
// las claves que no existían en la base. Si el archivo no existe o no es JSON, no-op.
func uninstallMergePrecise(full string, baseSnap, overlayPayload []byte) error {
	content := readFileOrEmptyBytes(full)
	if len(content) == 0 {
		return nil
	}
	current := map[string]any{}
	if err := json.Unmarshal(content, &current); err != nil {
		return nil // archivo no JSON o corrupto — no podemos des-fusionar
	}
	base := map[string]any{}
	if len(bytes.TrimSpace(baseSnap)) > 0 {
		json.Unmarshal(baseSnap, &base) // error → base vacía: se quita todo lo del overlay
	}
	overlay := map[string]any{}
	if err := json.Unmarshal(overlayPayload, &overlay); err != nil {
		return nil
	}
	cleaned := deepUnmerge(current, overlay, base)
	out, err := json.MarshalIndent(cleaned, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(full, out, 0o644)
}

// deepUnmerge revierte de current las contribuciones de overlay usando base como
// referencia de lo que había antes de aiwf. No muta los argumentos.
//   - clave ausente en base → aiwf la agregó entera → se elimina;
//   - mapas → recursión;
//   - arrays → se quitan solo los ítems que aiwf sumó (overlay menos base);
//   - escalar preexistente → se restaura el valor base (aiwf pudo sobrescribirlo).
func deepUnmerge(current, overlay, base map[string]any) map[string]any {
	out := make(map[string]any, len(current))
	for k, v := range current {
		out[k] = v
	}
	for k, ov := range overlay {
		cv, inCur := out[k]
		if !inCur {
			continue
		}
		bv, inBase := base[k]
		if !inBase {
			delete(out, k) // aiwf agregó la clave entera
			continue
		}
		if cm, ok := cv.(map[string]any); ok {
			if om, ok := ov.(map[string]any); ok {
				bm, _ := bv.(map[string]any)
				if bm == nil {
					bm = map[string]any{}
				}
				out[k] = deepUnmerge(cm, om, bm)
				continue
			}
		}
		if ca, ok := cv.([]any); ok {
			if oa, ok := ov.([]any); ok {
				ba, _ := bv.([]any)
				out[k] = subtractAdded(ca, oa, ba)
				continue
			}
		}
		out[k] = bv // escalar (u otro cambio de tipo): restaurar valor base
	}
	return out
}

// subtractAdded devuelve current sin los ítems que aiwf agregó (los de overlay que no
// estaban en base). Preserva el orden y conserva los ítems que ya estaban en base
// aunque también figuren en overlay.
func subtractAdded(current, overlay, base []any) []any {
	baseSeen := make(map[string]bool, len(base))
	for _, v := range base {
		baseSeen[canonicalKey(v)] = true
	}
	addedByAiwf := make(map[string]bool, len(overlay))
	for _, v := range overlay {
		if key := canonicalKey(v); !baseSeen[key] {
			addedByAiwf[key] = true
		}
	}
	out := make([]any, 0, len(current))
	for _, v := range current {
		if addedByAiwf[canonicalKey(v)] {
			continue
		}
		out = append(out, v)
	}
	return out
}
