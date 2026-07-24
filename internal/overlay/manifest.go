package overlay

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"os"
	"path/filepath"
	"time"
)

// Record es una entrada del manifiesto: qué instaló/parchó aiwf en una ruta.
type Record struct {
	Path      string    `json:"path"`
	Type      FileType  `json:"type"`
	Checksum  string    `json:"checksum"` // sha256 del payload aplicado
	AppliedAt time.Time `json:"applied_at"`
	// AddedKeys lista las top-level keys JSON que aiwf inyectó (JSONMerge).
	// Vacío para Owned/MarkerBlock o si no había keys nuevas. Compatibilidad con
	// manifiestos previos; el uninstall preciso usa BaseSnapshot+OverlayPayload.
	AddedKeys []string `json:"added_keys,omitempty"`
	// BaseSnapshot es el contenido del archivo JSON ANTES de que aiwf lo tocara por
	// primera vez (JSONMerge). Se CONGELA en el primer Apply y se preserva en cada
	// reconcile — nunca se re-captura sobre un archivo ya modificado por aiwf.
	// Permite revertir con precisión (restaurar arrays y escalares a su valor base).
	BaseSnapshot []byte `json:"base_snapshot,omitempty"`
	// OverlayPayload es el fragmento JSON que aiwf fusionó (JSONMerge). Junto a
	// BaseSnapshot permite calcular exactamente qué agregó aiwf para des-fusionar.
	OverlayPayload []byte `json:"overlay_payload,omitempty"`
}

// Manifest registra todo lo que aiwf aplicó, para detectar drift, reconciliar y
// desinstalar limpio sin tocar lo de gentle-ai.
type Manifest struct {
	path            string
	GentleAIVersion string   `json:"gentle_ai_version_seen,omitempty"`
	Records         []Record `json:"records"`
}

// checksum devuelve el sha256 hex de payload.
func checksum(payload []byte) string {
	sum := sha256.Sum256(payload)
	return hex.EncodeToString(sum[:])
}

// LoadManifest lee el manifiesto de path. Si no existe, devuelve uno vacío.
func LoadManifest(path string) (*Manifest, error) {
	m := &Manifest{path: path}
	data, err := os.ReadFile(path)
	if os.IsNotExist(err) {
		return m, nil
	}
	if err != nil {
		return nil, err
	}
	if err := json.Unmarshal(data, m); err != nil {
		return nil, err
	}
	m.path = path
	return m, nil
}

// Find devuelve el registro de una ruta y si existía.
func (m *Manifest) Find(path string) (Record, bool) {
	for i := range m.Records {
		if m.Records[i].Path == path {
			return m.Records[i], true
		}
	}
	return Record{}, false
}

// Upsert inserta o reemplaza el registro de una ruta.
func (m *Manifest) Upsert(r Record) {
	for i := range m.Records {
		if m.Records[i].Path == r.Path {
			m.Records[i] = r
			return
		}
	}
	m.Records = append(m.Records, r)
}

// Save persiste el manifiesto (creando el directorio si hace falta).
func (m *Manifest) Save() error {
	if err := os.MkdirAll(filepath.Dir(m.path), 0o755); err != nil {
		return err
	}
	data, err := json.MarshalIndent(m, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(m.path, data, 0o644)
}
