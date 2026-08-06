// Package state reads SDD state scoped to one change.
package state

import (
	"os"
	"path/filepath"
	"sort"

	"gopkg.in/yaml.v3"

	"github.com/agaspardev/aiwf/internal/workspace"
)

const FileName = "state.yaml"

type State = workspace.ChangeState

// Load reads state.yaml from a resolved change directory.
func Load(changeDir string) (*State, bool, error) {
	data, err := os.ReadFile(filepath.Join(changeDir, FileName))
	if os.IsNotExist(err) {
		return nil, false, nil
	}
	if err != nil {
		return nil, false, err
	}
	var state State
	if err := yaml.Unmarshal(data, &state); err != nil {
		return nil, false, err
	}
	if err := state.Validate(); err != nil {
		return nil, false, err
	}
	return &state, true, nil
}

// ListChanges returns non-archive change directories in deterministic order.
func ListChanges(changesDir string) ([]string, error) {
	entries, err := os.ReadDir(changesDir)
	if os.IsNotExist(err) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	changes := make([]string, 0, len(entries))
	for _, entry := range entries {
		if entry.IsDir() && entry.Name() != "archive" {
			changes = append(changes, entry.Name())
		}
	}
	sort.Strings(changes)
	return changes, nil
}
