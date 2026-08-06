package migrate

import (
	"fmt"
	"os"
	"path/filepath"
	"time"
)

// Apply copies deterministic operations. Sources remain intact until explicit finalize.
func Apply(root string, plan Plan) (Report, error) {
	if len(plan.Ambiguities) > 0 {
		return Report{}, fmt.Errorf("plan has %d unresolved ambiguities", len(plan.Ambiguities))
	}
	if err := preflightApply(root, plan.Operations); err != nil {
		return Report{}, err
	}
	report := Report{SchemaVersion: 1, AppliedAt: time.Now().UTC().Format(time.RFC3339)}
	for _, operation := range plan.Operations {
		source, target, err := validateOperation(root, operation)
		if err != nil {
			return report, err
		}
		sourceChecksum, err := fileChecksum(source)
		if err != nil {
			return report, err
		}
		if sourceChecksum != operation.Checksum {
			return report, fmt.Errorf("source changed after plan: %s", operation.Source)
		}
		if targetChecksum, checksumErr := fileChecksum(target); checksumErr == nil {
			if targetChecksum != operation.Checksum {
				return report, fmt.Errorf("target collision: %s", operation.Target)
			}
			report.Skipped = append(report.Skipped, operation)
			continue
		} else if !os.IsNotExist(checksumErr) {
			return report, checksumErr
		}
		if err := os.MkdirAll(filepath.Dir(target), 0o755); err != nil {
			return report, err
		}
		data, err := os.ReadFile(source)
		if err != nil {
			return report, err
		}
		if err := os.WriteFile(target, data, 0o644); err != nil {
			return report, err
		}
		report.Copied = append(report.Copied, operation)
	}
	return report, nil
}

func preflightApply(root string, operations []Operation) error {
	for _, operation := range operations {
		source, target, err := validateOperation(root, operation)
		if err != nil {
			return err
		}
		sourceChecksum, err := fileChecksum(source)
		if err != nil {
			return err
		}
		if sourceChecksum != operation.Checksum {
			return fmt.Errorf("source changed after plan: %s", operation.Source)
		}
		if targetChecksum, checksumErr := fileChecksum(target); checksumErr == nil {
			if targetChecksum != operation.Checksum {
				return fmt.Errorf("target collision: %s", operation.Target)
			}
		} else if !os.IsNotExist(checksumErr) {
			return checksumErr
		}
	}
	return nil
}

// Verify checks that every planned target still matches its source checksum.
func Verify(root string, plan Plan) error {
	if len(plan.Ambiguities) > 0 {
		return fmt.Errorf("plan has %d unresolved ambiguities", len(plan.Ambiguities))
	}
	for _, operation := range plan.Operations {
		_, target, err := validateOperation(root, operation)
		if err != nil {
			return err
		}
		checksum, err := fileChecksum(target)
		if err != nil {
			return err
		}
		if checksum != operation.Checksum {
			return fmt.Errorf("target checksum mismatch: %s", operation.Target)
		}
	}
	return nil
}

// Rollback removes unmodified copies and restores finalized sources when needed.
func Rollback(root string, report Report) error {
	for i := len(report.Copied) - 1; i >= 0; i-- {
		operation := report.Copied[i]
		source, target, err := validateOperation(root, operation)
		if err != nil {
			return err
		}
		data, err := os.ReadFile(target)
		if err != nil {
			return err
		}
		checksum, err := fileChecksum(target)
		if err != nil {
			return err
		}
		if checksum != operation.Checksum {
			return fmt.Errorf("target changed after apply: %s", operation.Target)
		}
		if _, sourceErr := os.Stat(source); os.IsNotExist(sourceErr) {
			if err := os.MkdirAll(filepath.Dir(source), 0o755); err != nil {
				return err
			}
			if err := os.WriteFile(source, data, 0o644); err != nil {
				return err
			}
		} else if sourceErr != nil {
			return sourceErr
		} else if sourceChecksum, checksumErr := fileChecksum(source); checksumErr != nil || sourceChecksum != operation.Checksum {
			return fmt.Errorf("source changed after apply: %s", operation.Source)
		}
		if err := os.Remove(target); err != nil {
			return err
		}
		removeEmptyParents(filepath.Dir(target), filepath.Join(root, ".ai-workflow"))
	}
	return nil
}

func removeEmptyParents(path, stop string) {
	for path != stop && path != filepath.Dir(path) {
		entries, err := os.ReadDir(path)
		if err != nil || len(entries) != 0 {
			return
		}
		if os.Remove(path) != nil {
			return
		}
		path = filepath.Dir(path)
	}
}
