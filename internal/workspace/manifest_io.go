package workspace

import (
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

// LoadWorkspaceManifest reads .ai-workflow/config/workspace.yaml. Returns
// ok=false when the manifest does not exist yet.
func LoadWorkspaceManifest(root string) (*WorkspaceManifest, bool, error) {
	path := filepath.Join(root, ".ai-workflow", "config", "workspace.yaml")
	data, err := os.ReadFile(path)
	if os.IsNotExist(err) {
		return nil, false, nil
	}
	if err != nil {
		return nil, false, err
	}
	var m WorkspaceManifest
	if err := yaml.Unmarshal(data, &m); err != nil {
		return nil, false, fmt.Errorf("parsear %s: %w", path, err)
	}
	if err := m.Validate(); err != nil {
		return nil, false, err
	}
	return &m, true, nil
}

// SaveWorkspaceManifest writes .ai-workflow/config/workspace.yaml.
func SaveWorkspaceManifest(root string, m *WorkspaceManifest) error {
	if err := m.Validate(); err != nil {
		return err
	}
	data, err := yaml.Marshal(m)
	if err != nil {
		return err
	}
	path := filepath.Join(root, ".ai-workflow", "config", "workspace.yaml")
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	return os.WriteFile(path, data, 0o644)
}
