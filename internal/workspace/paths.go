package workspace

import (
	"fmt"
	"path/filepath"
)

// Paths contains canonical workspace locations for one subproject and optional change.
type Paths struct {
	Repository        string
	Workflow          string
	Config            string
	WorkspaceManifest string
	KnowledgeShared   string
	KnowledgeProject  string
	Project           string
	ProjectManifest   string
	Specs             string
	Changes           string
	Change            string
	Evidence          string
	Reports           string
	Handoffs          string
	Notes             string
	Scratch           string
	Migrations        string
}

// NewPaths derives paths without touching the filesystem.
func NewPaths(repository, subproject, change string) (Paths, error) {
	if err := ValidateID(subproject); err != nil {
		return Paths{}, fmt.Errorf("subproject: %w", err)
	}
	if change != "" {
		if err := ValidateID(change); err != nil {
			return Paths{}, fmt.Errorf("change: %w", err)
		}
	}

	workflow := filepath.Join(repository, ".ai-workflow")
	config := filepath.Join(workflow, "config")
	project := filepath.Join(workflow, "projects", subproject)
	paths := Paths{
		Repository:        repository,
		Workflow:          workflow,
		Config:            config,
		WorkspaceManifest: filepath.Join(config, "workspace.yaml"),
		KnowledgeShared:   filepath.Join(workflow, "knowledge", "shared"),
		KnowledgeProject:  filepath.Join(workflow, "knowledge", "projects", subproject),
		Project:           project,
		ProjectManifest:   filepath.Join(project, "project.yaml"),
		Specs:             filepath.Join(project, "specs"),
		Changes:           filepath.Join(project, "changes"),
		Migrations:        filepath.Join(workflow, "migrations"),
	}
	if change == "" {
		return paths, nil
	}

	paths.Change = filepath.Join(paths.Changes, change)
	paths.Evidence = filepath.Join(paths.Change, "evidence")
	paths.Reports = filepath.Join(paths.Change, "reports")
	paths.Handoffs = filepath.Join(paths.Change, "handoffs")
	paths.Notes = filepath.Join(paths.Change, "notes")
	paths.Scratch = filepath.Join(paths.Change, "scratch")
	return paths, nil
}
