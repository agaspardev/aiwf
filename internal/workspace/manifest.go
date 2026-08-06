package workspace

import (
	"errors"
	"fmt"
	"path/filepath"
	"strings"
)

var (
	ErrInvalidManifest  = errors.New("invalid workspace manifest")
	ErrInvalidReference = errors.New("invalid workspace reference")
)

// WorkspaceManifest identifies the repository-level workspace contract.
type WorkspaceManifest struct {
	SchemaVersion     int    `json:"schemaVersion" yaml:"schemaVersion"`
	LayoutVersion     int    `json:"layoutVersion" yaml:"layoutVersion"`
	RepositoryID      string `json:"repositoryId" yaml:"repositoryId"`
	DefaultSubproject string `json:"defaultSubproject,omitempty" yaml:"defaultSubproject,omitempty"`
}

func (m WorkspaceManifest) Validate() error {
	if m.SchemaVersion < 1 || m.LayoutVersion < 1 {
		return fmt.Errorf("%w: schemaVersion and layoutVersion must be positive", ErrInvalidManifest)
	}
	if err := ValidateID(m.RepositoryID); err != nil {
		return fmt.Errorf("repositoryId: %w", err)
	}
	if m.DefaultSubproject != "" {
		if err := ValidateID(m.DefaultSubproject); err != nil {
			return fmt.Errorf("defaultSubproject: %w", err)
		}
	}
	return nil
}

// ProjectManifest defines one durable subproject scope.
type ProjectManifest struct {
	SchemaVersion  int      `json:"schemaVersion" yaml:"schemaVersion"`
	ID             string   `json:"id" yaml:"id"`
	CreatedAt      string   `json:"createdAt,omitempty" yaml:"createdAt,omitempty"`
	KnowledgeRoot  string   `json:"knowledgeRoot" yaml:"knowledgeRoot"`
	Description    string   `json:"description,omitempty" yaml:"description,omitempty"`
	Owners         []string `json:"owners,omitempty" yaml:"owners,omitempty"`
	Classification string   `json:"classification,omitempty" yaml:"classification,omitempty"`
	ArtifactStore  string   `json:"artifactStore,omitempty" yaml:"artifactStore,omitempty"`
}

func (m ProjectManifest) Validate() error {
	if m.SchemaVersion < 1 {
		return fmt.Errorf("%w: schemaVersion must be positive", ErrInvalidManifest)
	}
	if err := ValidateID(m.ID); err != nil {
		return fmt.Errorf("id: %w", err)
	}
	if err := validateWorkflowReference(m.KnowledgeRoot); err != nil {
		return fmt.Errorf("knowledgeRoot: %w", err)
	}
	if m.ArtifactStore != "" && m.ArtifactStore != "hybrid" && m.ArtifactStore != "engram" {
		return fmt.Errorf("%w: unsupported artifactStore %q", ErrInvalidManifest, m.ArtifactStore)
	}
	return nil
}

// Reference links a change to durable knowledge or immutable source material.
type Reference struct {
	Path     string `json:"path" yaml:"path"`
	Relation string `json:"relation,omitempty" yaml:"relation,omitempty"`
	Note     string `json:"note,omitempty" yaml:"note,omitempty"`
}

// Artifact declares an output owned by a change.
type Artifact struct {
	Path string `json:"path" yaml:"path"`
	Type string `json:"type,omitempty" yaml:"type,omitempty"`
}

// ChangeState persists SDD state without a repository-global active change.
type ChangeState struct {
	SchemaVersion  int         `json:"schemaVersion" yaml:"schemaVersion"`
	Subproject     string      `json:"subproject" yaml:"subproject"`
	Change         string      `json:"change" yaml:"change"`
	Status         string      `json:"status" yaml:"status"`
	Phase          string      `json:"phase" yaml:"phase"`
	KnowledgeLevel string      `json:"knowledgeLevel" yaml:"knowledgeLevel"`
	WorkItemRef    *Reference  `json:"workItemRef,omitempty" yaml:"workItemRef,omitempty"`
	KnowledgeRefs  []Reference `json:"knowledgeRefs,omitempty" yaml:"knowledgeRefs,omitempty"`
	SourceRefs     []Reference `json:"sourceRefs,omitempty" yaml:"sourceRefs,omitempty"`
	Artifacts      []Artifact  `json:"artifacts,omitempty" yaml:"artifacts,omitempty"`
	CreatedAt      string      `json:"createdAt,omitempty" yaml:"createdAt,omitempty"`
	UpdatedAt      string      `json:"updatedAt,omitempty" yaml:"updatedAt,omitempty"`
	LastGate       string      `json:"lastGate,omitempty" yaml:"lastGate,omitempty"`
}

func (s ChangeState) Validate() error {
	if s.SchemaVersion < 1 {
		return fmt.Errorf("%w: schemaVersion must be positive", ErrInvalidManifest)
	}
	if err := ValidateID(s.Subproject); err != nil {
		return fmt.Errorf("subproject: %w", err)
	}
	if err := ValidateID(s.Change); err != nil {
		return fmt.Errorf("change: %w", err)
	}
	if !validKnowledgeLevel(s.KnowledgeLevel) {
		return fmt.Errorf("%w: unsupported knowledgeLevel %q", ErrInvalidManifest, s.KnowledgeLevel)
	}
	if s.WorkItemRef != nil {
		if err := validateWorkflowReference(s.WorkItemRef.Path); err != nil {
			return fmt.Errorf("workItemRef: %w", err)
		}
	}
	for i, ref := range s.KnowledgeRefs {
		if err := validateWorkflowReference(ref.Path); err != nil {
			return fmt.Errorf("knowledgeRefs[%d]: %w", i, err)
		}
	}
	for i, ref := range s.SourceRefs {
		if err := validateWorkflowReference(ref.Path); err != nil {
			return fmt.Errorf("sourceRefs[%d]: %w", i, err)
		}
	}
	for i, artifact := range s.Artifacts {
		if err := validateChangeRelativeReference(artifact.Path); err != nil {
			return fmt.Errorf("artifacts[%d]: %w", i, err)
		}
	}
	return nil
}

func validKnowledgeLevel(level string) bool {
	switch level {
	case "K0", "K1", "K2", "K3", "K4":
		return true
	default:
		return false
	}
}

func validateWorkflowReference(path string) error {
	if err := validateRelativeReference(path); err != nil {
		return err
	}
	normalized := filepath.ToSlash(filepath.Clean(path))
	if normalized != ".ai-workflow" && !strings.HasPrefix(normalized, ".ai-workflow/") {
		return fmt.Errorf("%w: %q must stay inside .ai-workflow", ErrInvalidReference, path)
	}
	for _, component := range strings.Split(strings.TrimPrefix(normalized, ".ai-workflow/"), "/") {
		if component == "" {
			continue
		}
		base := strings.TrimSuffix(component, filepath.Ext(component))
		if err := ValidateID(base); err != nil {
			return fmt.Errorf("%w: component %q is not kebab-case", ErrInvalidReference, component)
		}
	}
	return nil
}

func validateChangeRelativeReference(path string) error {
	return validateRelativeReference(path)
}

func validateRelativeReference(path string) error {
	if strings.TrimSpace(path) == "" {
		return fmt.Errorf("%w: path is empty", ErrInvalidReference)
	}
	if filepath.IsAbs(path) || filepath.VolumeName(path) != "" || strings.HasPrefix(path, `\`) {
		return fmt.Errorf("%w: %q is absolute", ErrInvalidReference, path)
	}
	normalized := filepath.ToSlash(filepath.Clean(path))
	if normalized == ".." || strings.HasPrefix(normalized, "../") {
		return fmt.Errorf("%w: %q escapes its owner", ErrInvalidReference, path)
	}
	return nil
}

// EngramTopic returns the collision-free key for one SDD artifact.
func EngramTopic(subproject, change, artifact string) (string, error) {
	for label, value := range map[string]string{
		"subproject": subproject,
		"change":     change,
		"artifact":   artifact,
	} {
		if err := ValidateID(value); err != nil {
			return "", fmt.Errorf("%s: %w", label, err)
		}
	}
	return "sdd/" + subproject + "/" + change + "/" + artifact, nil
}
