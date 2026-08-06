package migrate

import (
	"fmt"
	"os"
	"path/filepath"
	"time"
)

// acquireLock impone exclusividad inter-proceso para evitar migraciones concurrentes.
func acquireLock(root string) (*os.File, error) {
	lockPath := filepath.Join(root, ".ai-workflow", "migrations", "MIGRATION.lock")
	if err := os.MkdirAll(filepath.Dir(lockPath), 0o755); err != nil {
		return nil, err
	}
	// O_EXCL crea el archivo solo si no existe; es atómico en SO.
	f, err := os.OpenFile(lockPath, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0o600)
	if err != nil {
		return nil, fmt.Errorf("migración ya en progreso (lock existente): %w", err)
	}
	return f, nil
}

func releaseLock(f *os.File) {
	if f != nil {
		name := f.Name()
		f.Close()
		// Close antes de Remove: en Windows un handle abierto bloquea el borrado.
		os.Remove(name)
	}
}

// Apply copies deterministic operations. Sources remain intact until explicit finalize.
func Apply(root string, plan Plan) (Report, error) {
	lock, err := acquireLock(root)
	if err != nil {
		return Report{}, err
	}
	defer releaseLock(lock)

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
		tmpTarget := target + ".tmp"
		if err := os.WriteFile(tmpTarget, data, 0o644); err != nil {
			return report, err
		}
		if err := os.Rename(tmpTarget, target); err != nil {
			os.Remove(tmpTarget)
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
	lock, err := acquireLock(root)
	if err != nil {
		return err
	}
	defer releaseLock(lock)

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
		// Limpia un posible .tmp interrumpido de un Apply previo.
		os.Remove(target + ".tmp")
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
