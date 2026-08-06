// Package initproj initializes aiwf's minimal repository control plane.
// Generated workflow artifacts remain local and are excluded per clone.
package initproj

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"

	"github.com/agaspardev/aiwf/internal/workspace"
)

// Report accumulates observable initialization actions.
type Report struct {
	Created  []string
	Skipped  []string
	Warnings []string
}

func (r *Report) created(path string) { r.Created = append(r.Created, path) }
func (r *Report) skipped(path string) { r.Skipped = append(r.Skipped, path) }
func (r *Report) warn(message string) { r.Warnings = append(r.Warnings, message) }

// Init creates only the repository workspace manifest. Feature-specific directories
// are materialized on demand by their owning command.
func Init(root, name string, force bool) (*Report, error) {
	report := &Report{}
	manifest := workspace.WorkspaceManifest{
		SchemaVersion: 1,
		LayoutVersion: 1,
		RepositoryID:  name,
	}
	if err := manifest.Validate(); err != nil {
		return report, err
	}

	data, err := yaml.Marshal(manifest)
	if err != nil {
		return report, fmt.Errorf("serializar workspace manifest: %w", err)
	}
	path := filepath.Join(root, ".ai-workflow", "config", "workspace.yaml")
	if err := writeIfMissing(path, string(data), force, report, ".ai-workflow/config/workspace.yaml"); err != nil {
		return report, err
	}
	if err := ensureGitExclude(root, report); err != nil {
		return report, err
	}
	return report, nil
}

// InitProject creates one subproject manifest without speculative directories.
func InitProject(root, subproject string, force bool) (*Report, error) {
	report := &Report{}
	paths, err := workspace.NewPaths(root, subproject, "")
	if err != nil {
		return report, err
	}
	manifest := workspace.ProjectManifest{
		SchemaVersion: 1,
		ID:            subproject,
		KnowledgeRoot: filepath.ToSlash(filepath.Join(".ai-workflow", "knowledge", "projects", subproject)),
		ArtifactStore: "hybrid",
	}
	if err := manifest.Validate(); err != nil {
		return report, err
	}
	data, err := yaml.Marshal(manifest)
	if err != nil {
		return report, fmt.Errorf("serializar project manifest: %w", err)
	}
	label := filepath.ToSlash(filepath.Join(".ai-workflow", "projects", subproject, "project.yaml"))
	if err := writeIfMissing(paths.ProjectManifest, string(data), force, report, label); err != nil {
		return report, err
	}
	return report, nil
}

func writeIfMissing(path, content string, force bool, report *Report, label string) error {
	if _, err := os.Stat(path); err == nil && !force {
		report.skipped(label)
		return nil
	} else if err != nil && !os.IsNotExist(err) {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		return err
	}
	report.created(label)
	return nil
}

func ensureGitExclude(root string, report *Report) error {
	gitDir := filepath.Join(root, ".git")
	if _, err := os.Stat(gitDir); err != nil {
		report.warn("No se detectó .git/. Añadí manualmente .ai-workflow/ a .git/info/exclude")
		return nil
	}

	infoDir := filepath.Join(gitDir, "info")
	if err := os.MkdirAll(infoDir, 0o755); err != nil {
		return err
	}
	excludePath := filepath.Join(infoDir, "exclude")
	existing, err := os.ReadFile(excludePath)
	if err != nil && !os.IsNotExist(err) {
		return err
	}
	if linePresent(string(existing), ".ai-workflow/") {
		return nil
	}

	const marker = "# AI Workflow (local, no versionar) — aiwf init"
	content := string(existing)
	if content != "" && !strings.HasSuffix(content, "\n") {
		content += "\n"
	}
	content += marker + "\n.ai-workflow/\n"
	return os.WriteFile(excludePath, []byte(content), 0o644)
}

func linePresent(content, target string) bool {
	for _, line := range strings.Split(content, "\n") {
		if strings.TrimSpace(line) == target {
			return true
		}
	}
	return false
}

// gitTracked remains available for callers that need to guard client-owned files.
func gitTracked(root, path string) bool {
	if _, err := exec.LookPath("git"); err != nil {
		return false
	}
	command := exec.Command("git", "-C", root, "ls-files", "--error-unmatch", path)
	return command.Run() == nil
}
