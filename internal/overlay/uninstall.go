package overlay

import (
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
			if err := uninstallMerge(full, r.AddedKeys); err != nil {
				return skippedJSON, err
			}
			if len(r.AddedKeys) == 0 {
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
