package migrate

import (
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

func SavePlan(path string, plan Plan) error {
	data, err := yaml.Marshal(plan)
	if err != nil {
		return err
	}
	return writeFile(path, data)
}

func LoadPlan(path string) (Plan, error) {
	var plan Plan
	data, err := os.ReadFile(path)
	if err != nil {
		return Plan{}, err
	}
	if err := yaml.Unmarshal(data, &plan); err != nil {
		return Plan{}, err
	}
	return plan, nil
}

func SaveReport(path string, report Report) error {
	data, err := yaml.Marshal(report)
	if err != nil {
		return err
	}
	return writeFile(path, data)
}

func LoadReport(path string) (Report, error) {
	var report Report
	data, err := os.ReadFile(path)
	if err != nil {
		return Report{}, err
	}
	if err := yaml.Unmarshal(data, &report); err != nil {
		return Report{}, err
	}
	return report, nil
}

// Finalize removes verified sources. It is intentionally separate from Apply.
func Finalize(root string, plan Plan) error {
	lock, err := acquireLock(root)
	if err != nil {
		return err
	}
	defer releaseLock(lock)

	if err := Verify(root, plan); err != nil {
		return err
	}
	for _, operation := range plan.Operations {
		source, _, err := validateOperation(root, operation)
		if err != nil {
			return err
		}
		checksum, err := fileChecksum(source)
		if err != nil {
			return err
		}
		if checksum != operation.Checksum {
			return fmt.Errorf("source changed after plan: %s", operation.Source)
		}
		if err := os.Remove(source); err != nil {
			return err
		}
		removeEmptyParents(filepath.Dir(source), filepath.Join(root, ".ai-workflow"))
	}
	return nil
}

func writeFile(path string, data []byte) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	return os.WriteFile(path, data, 0o644)
}
