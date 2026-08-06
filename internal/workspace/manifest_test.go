package workspace

import (
	"errors"
	"testing"
)

func TestWorkspaceManifestValidate(t *testing.T) {
	manifest := WorkspaceManifest{SchemaVersion: 1, LayoutVersion: 1, RepositoryID: "aiwf"}
	if err := manifest.Validate(); err != nil {
		t.Fatalf("Validate: %v", err)
	}
	manifest.LayoutVersion = 0
	if err := manifest.Validate(); err == nil {
		t.Fatal("expected invalid layout version")
	}
}

func TestProjectManifestValidate(t *testing.T) {
	manifest := ProjectManifest{
		SchemaVersion: 1,
		ID:            "aiwf-core",
		KnowledgeRoot: ".ai-workflow/knowledge/projects/aiwf-core",
	}
	if err := manifest.Validate(); err != nil {
		t.Fatalf("Validate: %v", err)
	}
	manifest.KnowledgeRoot = "../knowledge"
	if err := manifest.Validate(); !errors.Is(err, ErrInvalidReference) {
		t.Fatalf("error = %v, want ErrInvalidReference", err)
	}
}

func TestChangeStateValidate(t *testing.T) {
	state := ChangeState{
		SchemaVersion:  1,
		Subproject:     "aiwf-core",
		Change:         "f0-workspace-information-architecture",
		Status:         "active",
		Phase:          "tasks-ready",
		KnowledgeLevel: "K4",
		KnowledgeRefs:  []Reference{{Path: ".ai-workflow/knowledge/projects/aiwf-core/tech/current-state.md", Relation: "context"}},
		SourceRefs:     []Reference{{Path: ".ai-workflow/projects/aiwf-core/intake/requirements.md", Relation: "source"}},
		Artifacts:      []Artifact{{Path: "proposal.md", Type: "proposal"}},
	}
	if err := state.Validate(); err != nil {
		t.Fatalf("Validate: %v", err)
	}
}

func TestChangeStateRejectsInvalidValues(t *testing.T) {
	valid := ChangeState{
		SchemaVersion:  1,
		Subproject:     "aiwf-core",
		Change:         "change-one",
		Status:         "active",
		Phase:          "proposal",
		KnowledgeLevel: "K2",
	}
	tests := []struct {
		name   string
		mutate func(*ChangeState)
		want   error
	}{
		{name: "invalid subproject", mutate: func(s *ChangeState) { s.Subproject = "../bad" }, want: ErrInvalidIdentity},
		{name: "invalid change", mutate: func(s *ChangeState) { s.Change = "Bad" }, want: ErrInvalidIdentity},
		{name: "invalid knowledge level", mutate: func(s *ChangeState) { s.KnowledgeLevel = "K5" }, want: ErrInvalidManifest},
		{name: "absolute reference", mutate: func(s *ChangeState) { s.SourceRefs = []Reference{{Path: "/tmp/source"}} }, want: ErrInvalidReference},
		{name: "traversal reference", mutate: func(s *ChangeState) { s.KnowledgeRefs = []Reference{{Path: "../knowledge"}} }, want: ErrInvalidReference},
		{name: "absolute artifact", mutate: func(s *ChangeState) { s.Artifacts = []Artifact{{Path: `C:\tmp\proposal.md`}} }, want: ErrInvalidReference},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			state := valid
			tt.mutate(&state)
			if err := state.Validate(); !errors.Is(err, tt.want) {
				t.Fatalf("error = %v, want %v", err, tt.want)
			}
		})
	}
}

func TestEngramTopic(t *testing.T) {
	got, err := EngramTopic("aiwf-core", "add-auth", "proposal")
	if err != nil {
		t.Fatalf("EngramTopic: %v", err)
	}
	if want := "sdd/aiwf-core/add-auth/proposal"; got != want {
		t.Fatalf("EngramTopic = %q, want %q", got, want)
	}
	if _, err := EngramTopic("aiwf-core", "add-auth", "../proposal"); !errors.Is(err, ErrInvalidIdentity) {
		t.Fatalf("error = %v, want ErrInvalidIdentity", err)
	}
}
