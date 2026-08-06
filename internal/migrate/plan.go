// Package migrate plans and executes explicit, checksum-aware layout migrations.
package migrate

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/agaspardev/aiwf/internal/workspace"
)

const (
	ClassificationDeterministic = "deterministic"
	ClassificationAmbiguous     = "ambiguous"
)

type Operation struct {
	Source         string `json:"source" yaml:"source"`
	Target         string `json:"target" yaml:"target"`
	Checksum       string `json:"checksum" yaml:"checksum"`
	Classification string `json:"classification" yaml:"classification"`
}

type Ambiguity struct {
	Source string `json:"source" yaml:"source"`
	Reason string `json:"reason" yaml:"reason"`
}

type Plan struct {
	SchemaVersion int         `json:"schemaVersion" yaml:"schemaVersion"`
	Subproject    string      `json:"subproject" yaml:"subproject"`
	CreatedAt     string      `json:"createdAt" yaml:"createdAt"`
	Operations    []Operation `json:"operations" yaml:"operations"`
	Ambiguities   []Ambiguity `json:"ambiguities,omitempty" yaml:"ambiguities,omitempty"`
}

type Report struct {
	SchemaVersion int         `json:"schemaVersion" yaml:"schemaVersion"`
	AppliedAt     string      `json:"appliedAt" yaml:"appliedAt"`
	Copied        []Operation `json:"copied,omitempty" yaml:"copied,omitempty"`
	Skipped       []Operation `json:"skipped,omitempty" yaml:"skipped,omitempty"`
}

// BuildPlan is read-only. It classifies legacy files and computes stable targets.
func BuildPlan(root, subproject string) (Plan, error) {
	if err := workspace.ValidateID(subproject); err != nil {
		return Plan{}, err
	}
	plan := Plan{SchemaVersion: 1, Subproject: subproject, CreatedAt: time.Now().UTC().Format(time.RFC3339)}

	sources := []struct {
		root   string
		mapper func(string) (string, bool, string)
	}{
		{root: ".ai-workflow/changes", mapper: changeMapper(subproject)},
		{root: ".ai-workflow/openspec/changes", mapper: changeMapper(subproject)},
		{root: ".ai-workflow/openspec/specs", mapper: specsMapper(subproject)},
		{root: ".ai-workflow/notes", mapper: ambiguousMapper("loose note requires a knowledge, intake, change, or archive owner")},
		{root: ".claude/knowledge", mapper: ambiguousMapper("knowledge artifact requires KDD classification; placeholders may be discarded")},
	}
	for _, source := range sources {
		if err := collect(root, source.root, source.mapper, &plan); err != nil {
			return Plan{}, err
		}
	}
	sort.Slice(plan.Operations, func(i, j int) bool { return plan.Operations[i].Source < plan.Operations[j].Source })
	sort.Slice(plan.Ambiguities, func(i, j int) bool { return plan.Ambiguities[i].Source < plan.Ambiguities[j].Source })
	return plan, nil
}

func collect(root, relativeRoot string, mapper func(string) (string, bool, string), plan *Plan) error {
	base := filepath.Join(root, filepath.FromSlash(relativeRoot))
	if _, err := os.Stat(base); os.IsNotExist(err) {
		return nil
	} else if err != nil {
		return err
	}
	return filepath.WalkDir(base, func(path string, entry fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if entry.IsDir() {
			if entry.Name() == "archive" {
				return filepath.SkipDir
			}
			return nil
		}
		relative, err := filepath.Rel(root, path)
		if err != nil {
			return err
		}
		relative = filepath.ToSlash(relative)
		target, deterministic, reason := mapper(relative)
		if !deterministic {
			plan.Ambiguities = append(plan.Ambiguities, Ambiguity{Source: relative, Reason: reason})
			return nil
		}
		checksum, err := fileChecksum(path)
		if err != nil {
			return err
		}
		plan.Operations = append(plan.Operations, Operation{
			Source: relative, Target: target, Checksum: checksum, Classification: ClassificationDeterministic,
		})
		return nil
	})
}

func changeMapper(subproject string) func(string) (string, bool, string) {
	return func(source string) (string, bool, string) {
		parts := strings.Split(filepath.ToSlash(source), "/")
		changeIndex := 2
		if len(parts) > 2 && parts[1] == "openspec" {
			changeIndex = 3
		}
		if len(parts) <= changeIndex+1 {
			return "", false, "change path is incomplete"
		}
		change := parts[changeIndex]
		fileParts := parts[changeIndex+1:]
		if len(fileParts) == 1 && fileParts[0] == "spec.md" {
			return "", false, "flat spec requires an explicit domain"
		}
		targetParts := append([]string{".ai-workflow", "projects", subproject, "changes", change}, fileParts...)
		return strings.Join(targetParts, "/"), true, ""
	}
}

func specsMapper(subproject string) func(string) (string, bool, string) {
	return func(source string) (string, bool, string) {
		const prefix = ".ai-workflow/openspec/specs/"
		if !strings.HasPrefix(source, prefix) {
			return "", false, "spec path is outside legacy specs root"
		}
		return ".ai-workflow/projects/" + subproject + "/specs/" + strings.TrimPrefix(source, prefix), true, ""
	}
}

func ambiguousMapper(reason string) func(string) (string, bool, string) {
	return func(string) (string, bool, string) { return "", false, reason }
}

func fileChecksum(path string) (string, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}
	sum := sha256.Sum256(data)
	return hex.EncodeToString(sum[:]), nil
}

func validateOperation(root string, operation Operation) (string, string, error) {
	source := filepath.Join(root, filepath.FromSlash(operation.Source))
	target := filepath.Join(root, filepath.FromSlash(operation.Target))
	for label, path := range map[string]string{"source": source, "target": target} {
		relative, err := filepath.Rel(root, path)
		if err != nil || relative == ".." || strings.HasPrefix(relative, ".."+string(filepath.Separator)) {
			return "", "", fmt.Errorf("%s escapes repository: %s", label, path)
		}
	}
	return source, target, nil
}
