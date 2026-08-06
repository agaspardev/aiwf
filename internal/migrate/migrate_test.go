package migrate

import (
	"crypto/sha256"
	"encoding/hex"
	"os"
	"path/filepath"
	"testing"
)

func TestPlanDetectsLegacyChangesWithoutWriting(t *testing.T) {
	root := t.TempDir()
	write(t, filepath.Join(root, ".ai-workflow", "changes", "legacy-change", "proposal.md"), "legacy")
	before := treeDigest(t, root)

	plan, err := BuildPlan(root, "aiwf-core")
	if err != nil {
		t.Fatalf("BuildPlan: %v", err)
	}
	if len(plan.Operations) != 1 {
		t.Fatalf("operations = %+v", plan.Operations)
	}
	operation := plan.Operations[0]
	if operation.Source != ".ai-workflow/changes/legacy-change/proposal.md" || operation.Target != ".ai-workflow/projects/aiwf-core/changes/legacy-change/proposal.md" {
		t.Fatalf("operation = %+v", operation)
	}
	if operation.Checksum == "" || operation.Classification != ClassificationDeterministic {
		t.Fatalf("operation lacks checksum/classification: %+v", operation)
	}
	if after := treeDigest(t, root); after != before {
		t.Fatalf("BuildPlan wrote to filesystem: before=%s after=%s", before, after)
	}
}

func TestPlanMarksFlatSpecAndLooseNotesAmbiguous(t *testing.T) {
	root := t.TempDir()
	write(t, filepath.Join(root, ".ai-workflow", "changes", "legacy-change", "spec.md"), "spec")
	write(t, filepath.Join(root, ".ai-workflow", "notes", "idea.md"), "idea")

	plan, err := BuildPlan(root, "aiwf-core")
	if err != nil {
		t.Fatal(err)
	}
	if len(plan.Ambiguities) != 2 {
		t.Fatalf("ambiguities = %+v", plan.Ambiguities)
	}
}

func TestApplyIsIdempotentAndVerifyPasses(t *testing.T) {
	root := t.TempDir()
	write(t, filepath.Join(root, ".ai-workflow", "changes", "legacy-change", "proposal.md"), "legacy")
	plan, err := BuildPlan(root, "aiwf-core")
	if err != nil {
		t.Fatal(err)
	}

	report, err := Apply(root, plan)
	if err != nil {
		t.Fatalf("Apply: %v", err)
	}
	if len(report.Copied) != 1 || len(report.Skipped) != 0 {
		t.Fatalf("report = %+v", report)
	}
	second, err := Apply(root, plan)
	if err != nil {
		t.Fatalf("second Apply: %v", err)
	}
	if len(second.Copied) != 0 || len(second.Skipped) != 1 {
		t.Fatalf("second report = %+v", second)
	}
	if err := Verify(root, plan); err != nil {
		t.Fatalf("Verify: %v", err)
	}
}

func TestApplyBlocksCollision(t *testing.T) {
	root := t.TempDir()
	write(t, filepath.Join(root, ".ai-workflow", "changes", "a-change", "proposal.md"), "first")
	write(t, filepath.Join(root, ".ai-workflow", "changes", "z-change", "proposal.md"), "second")
	plan, err := BuildPlan(root, "aiwf-core")
	if err != nil {
		t.Fatal(err)
	}
	write(t, filepath.Join(root, filepath.FromSlash(plan.Operations[1].Target)), "different")
	if _, err := Apply(root, plan); err == nil {
		t.Fatal("Apply should block target collision")
	}
	if _, err := os.Stat(filepath.Join(root, filepath.FromSlash(plan.Operations[0].Target))); !os.IsNotExist(err) {
		t.Fatal("preflight failure must not leave partial copies")
	}
}

func TestRollbackRemovesUnmodifiedCopies(t *testing.T) {
	root := t.TempDir()
	write(t, filepath.Join(root, ".ai-workflow", "changes", "legacy-change", "proposal.md"), "legacy")
	plan, err := BuildPlan(root, "aiwf-core")
	if err != nil {
		t.Fatal(err)
	}
	report, err := Apply(root, plan)
	if err != nil {
		t.Fatal(err)
	}
	if err := Rollback(root, report); err != nil {
		t.Fatalf("Rollback: %v", err)
	}
	assertRolledBack(t, root, plan.Operations[0])
}

func TestRollbackAfterFinalizeRestoresSource(t *testing.T) {
	root := t.TempDir()
	write(t, filepath.Join(root, ".ai-workflow", "changes", "legacy-change", "proposal.md"), "legacy")
	plan, err := BuildPlan(root, "aiwf-core")
	if err != nil {
		t.Fatal(err)
	}
	report, err := Apply(root, plan)
	if err != nil {
		t.Fatal(err)
	}
	if err := Finalize(root, plan); err != nil {
		t.Fatalf("Finalize: %v", err)
	}
	if err := Rollback(root, report); err != nil {
		t.Fatalf("Rollback after finalize: %v", err)
	}
	assertRolledBack(t, root, plan.Operations[0])
}

func assertRolledBack(t *testing.T, root string, operation Operation) {
	t.Helper()
	if _, err := os.Stat(filepath.Join(root, filepath.FromSlash(operation.Target))); !os.IsNotExist(err) {
		t.Fatal("rollback did not remove copied target")
	}
	source := filepath.Join(root, filepath.FromSlash(operation.Source))
	if checksum, err := fileChecksum(source); err != nil || checksum != operation.Checksum {
		t.Fatalf("rollback source checksum=%s err=%v", checksum, err)
	}
}

func treeDigest(t *testing.T, root string) string {
	t.Helper()
	hash := sha256.New()
	err := filepath.WalkDir(root, func(path string, entry os.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		relative, err := filepath.Rel(root, path)
		if err != nil {
			return err
		}
		hash.Write([]byte(filepath.ToSlash(relative)))
		if !entry.IsDir() {
			data, err := os.ReadFile(path)
			if err != nil {
				return err
			}
			hash.Write(data)
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	return hex.EncodeToString(hash.Sum(nil))
}

func write(t *testing.T, path, content string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
}
