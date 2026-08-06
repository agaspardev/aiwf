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

// ChangeSummary is one change's reported phase/status for `aiwf status`.
// Inferred is true when Phase did not come from state.yaml but was derived
// from the most advanced artifact present in the change directory.
type ChangeSummary struct {
	Name     string
	Phase    string
	Status   string
	Inferred bool
}

// phaseOrder is the monotonic SDD phase progression; later = more advanced.
var phaseOrder = []struct {
	artifact string // file (or dir) whose presence implies this phase
	phase    string
	isDir    bool
}{
	{"exploration.md", "explore", false},
	{"proposal.md", "propose", false},
	{"design.md", "design", false},
	{"tasks.md", "tasks", false},
	{"verify.md", "verify", false},
	{"archive", "archive", true},
}

// inferPhase returns the most advanced phase implied by artifacts in changeDir,
// or "unknown" when none are present. Deterministic (fixed order, not fs order).
func inferPhase(changeDir string) string {
	phase := "unknown"
	for _, p := range phaseOrder {
		info, err := os.Stat(filepath.Join(changeDir, p.artifact))
		if err != nil {
			continue
		}
		if info.IsDir() == p.isDir {
			phase = p.phase
		}
	}
	return phase
}

// Summarize returns one ChangeSummary per non-archive change under changesDir.
// Phase/Status come from state.yaml when present; otherwise phase is inferred
// from artifacts (Inferred=true) and status is left empty.
func Summarize(changesDir string) ([]ChangeSummary, error) {
	names, err := ListChanges(changesDir)
	if err != nil {
		return nil, err
	}
	sums := make([]ChangeSummary, 0, len(names))
	for _, name := range names {
		dir := filepath.Join(changesDir, name)
		s, ok, err := Load(dir)
		if err != nil {
			return nil, err
		}
		if ok {
			sums = append(sums, ChangeSummary{Name: name, Phase: s.Phase, Status: s.Status})
			continue
		}
		sums = append(sums, ChangeSummary{Name: name, Phase: inferPhase(dir), Inferred: true})
	}
	return sums, nil
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
