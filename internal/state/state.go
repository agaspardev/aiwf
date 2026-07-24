// Package state lee el estado del workflow de un proyecto
// (.ai-workflow/state/workflow-state.json), usado por `aiwf estado`.
package state

import (
	"encoding/json"
	"os"
	"path/filepath"
)

// RelPath es la ubicación del estado relativa a la raíz del proyecto.
const RelPath = ".ai-workflow/state/workflow-state.json"

// State refleja workflow-state.json. Las listas de tareas se guardan sin interpretar
// (solo se cuentan).
type State struct {
	Project        string            `json:"project"`
	Phase          string            `json:"phase"`
	ActiveTasks    []json.RawMessage `json:"activeTasks"`
	CompletedTasks []json.RawMessage `json:"completedTasks"`
	BlockedTasks   []json.RawMessage `json:"blockedTasks"`
	NextAction     string            `json:"nextAction"`
}

// Load lee el estado del workflow desde dir. El segundo valor es false si no existe.
func Load(dir string) (*State, bool, error) {
	p := filepath.Join(dir, filepath.FromSlash(RelPath))
	data, err := os.ReadFile(p)
	if os.IsNotExist(err) {
		return nil, false, nil
	}
	if err != nil {
		return nil, false, err
	}
	var s State
	if err := json.Unmarshal(data, &s); err != nil {
		return nil, false, err
	}
	return &s, true, nil
}
