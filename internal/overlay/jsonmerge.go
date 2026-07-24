package overlay

import (
	"bytes"
	"encoding/json"
	"fmt"
)

// DeepMerge combina overlay sobre base recursivamente. Reglas:
//   - mapas: se fusionan recursivamente;
//   - arrays: se UNEN (items de base + items nuevos de overlay, con dedup) para no
//     perder contenido de gentle-ai (p. ej. permissions.allow/deny);
//   - escalares o tipos distintos: gana overlay (nuestra customización).
//
// No muta los argumentos.
func DeepMerge(base, overlay map[string]any) map[string]any {
	out := make(map[string]any, len(base))
	for k, v := range base {
		out[k] = v
	}
	for k, ov := range overlay {
		if bv, ok := out[k]; ok {
			out[k] = mergeValue(bv, ov)
		} else {
			out[k] = ov
		}
	}
	return out
}

// mergeValue fusiona dos valores según su tipo (ver reglas en DeepMerge).
func mergeValue(base, overlay any) any {
	if bm, ok := base.(map[string]any); ok {
		if om, ok := overlay.(map[string]any); ok {
			return DeepMerge(bm, om)
		}
	}
	if ba, ok := base.([]any); ok {
		if oa, ok := overlay.([]any); ok {
			return unionArrays(ba, oa)
		}
	}
	return overlay
}

// unionArrays devuelve base seguido de los items de overlay que no estén ya presentes
// (dedup por igualdad JSON canónica). Preserva el orden y nunca descarta items de base.
func unionArrays(base, overlay []any) []any {
	out := make([]any, 0, len(base)+len(overlay))
	seen := make(map[string]bool, len(base)+len(overlay))
	add := func(v any) {
		key := canonicalKey(v)
		if seen[key] {
			return
		}
		seen[key] = true
		out = append(out, v)
	}
	for _, v := range base {
		add(v)
	}
	for _, v := range overlay {
		add(v)
	}
	return out
}

// canonicalKey serializa v a JSON para usarlo como clave de dedup.
func canonicalKey(v any) string {
	b, err := json.Marshal(v)
	if err != nil {
		return fmt.Sprintf("%v", v)
	}
	return string(b)
}

// MergeJSON parsea base y overlay como objetos JSON y devuelve el merge indentado.
// base puede estar vacío (se trata como {}). Si base no está vacío y es JSON inválido,
// devuelve error (no queremos clobberear un archivo de gentle-ai que no entendemos).
func MergeJSON(base, overlay []byte) ([]byte, error) {
	baseMap := map[string]any{}
	if len(bytes.TrimSpace(base)) > 0 {
		if err := json.Unmarshal(base, &baseMap); err != nil {
			return nil, fmt.Errorf("base JSON inválido: %w", err)
		}
	}
	overlayMap := map[string]any{}
	if err := json.Unmarshal(overlay, &overlayMap); err != nil {
		return nil, fmt.Errorf("overlay JSON inválido: %w", err)
	}
	merged := DeepMerge(baseMap, overlayMap)
	return json.MarshalIndent(merged, "", "  ")
}
