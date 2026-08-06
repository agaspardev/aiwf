package main

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/agaspardev/aiwf/internal/migrate"
)

func migrateCmd(args []string) int {
	if len(args) == 0 {
		fmt.Fprintln(os.Stderr, "uso: aiwf migrate layout|apply|verify|finalize|rollback ...")
		return 2
	}
	root, err := os.Getwd()
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		return 1
	}
	switch args[0] {
	case "layout":
		return migrateLayout(root, args[1:])
	case "apply":
		return migrateApply(root, args[1:])
	case "verify":
		return migrateVerify(root, args[1:])
	case "finalize":
		return migrateFinalize(root, args[1:])
	case "rollback":
		return migrateRollback(root, args[1:])
	default:
		fmt.Fprintf(os.Stderr, "error: acción migrate desconocida: %s\n", args[0])
		return 2
	}
}

func migrateLayout(root string, args []string) int {
	subproject, save, deterministicOnly, err := parseMigrationPlanArgs(args)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		return 2
	}
	plan, err := migrate.BuildPlan(root, subproject)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		return 1
	}
	if deterministicOnly {
		plan.Ambiguities = nil
	}
	fmt.Printf("[migrate] %d deterministas · %d ambiguos\n", len(plan.Operations), len(plan.Ambiguities))
	for _, operation := range plan.Operations {
		fmt.Printf("  COPY %s -> %s\n", operation.Source, operation.Target)
	}
	for _, ambiguity := range plan.Ambiguities {
		fmt.Printf("  BLOCK %s — %s\n", ambiguity.Source, ambiguity.Reason)
	}
	if !save {
		fmt.Println("  dry-run: cero escrituras. Usá --save-plan para persistir el plan.")
		return 0
	}
	path := filepath.Join(root, ".ai-workflow", "migrations", "plans", time.Now().UTC().Format("20060102-150405")+"-layout.yaml")
	if err := migrate.SavePlan(path, plan); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		return 1
	}
	fmt.Printf("  plan: %s\n", path)
	return 0
}

func migrateApply(root string, args []string) int {
	path, err := requiredFlag(args, "--plan")
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		return 2
	}
	plan, err := migrate.LoadPlan(path)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		return 1
	}
	report, err := migrate.Apply(root, plan)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		return 1
	}
	reportPath := filepath.Join(root, ".ai-workflow", "migrations", "runs", time.Now().UTC().Format("20060102-150405")+"-apply.yaml")
	if err := migrate.SaveReport(reportPath, report); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		return 1
	}
	fmt.Printf("[migrate] copied=%d skipped=%d report=%s\n", len(report.Copied), len(report.Skipped), reportPath)
	return 0
}

func migrateVerify(root string, args []string) int {
	path, err := requiredFlag(args, "--plan")
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		return 2
	}
	plan, err := migrate.LoadPlan(path)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		return 1
	}
	if err := migrate.Verify(root, plan); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		return 1
	}
	fmt.Println("[migrate] verify PASS")
	return 0
}

func migrateFinalize(root string, args []string) int {
	path, err := requiredFlag(args, "--plan")
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		return 2
	}
	plan, err := migrate.LoadPlan(path)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		return 1
	}
	if err := migrate.Finalize(root, plan); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		return 1
	}
	fmt.Println("[migrate] finalize completado")
	return 0
}

func migrateRollback(root string, args []string) int {
	path, err := requiredFlag(args, "--run")
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		return 2
	}
	report, err := migrate.LoadReport(path)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		return 1
	}
	if err := migrate.Rollback(root, report); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		return 1
	}
	fmt.Println("[migrate] rollback completado")
	return 0
}

func parseMigrationPlanArgs(args []string) (string, bool, bool, error) {
	subproject, save, deterministicOnly := "", false, false
	for i := 0; i < len(args); i++ {
		switch args[i] {
		case "--subproject":
			if i+1 >= len(args) {
				return "", false, false, fmt.Errorf("--subproject requiere un valor")
			}
			subproject = args[i+1]
			i++
		case "--save-plan":
			save = true
		case "--deterministic-only":
			deterministicOnly = true
		default:
			return "", false, false, fmt.Errorf("argumento desconocido: %s", args[i])
		}
	}
	if subproject == "" {
		return "", false, false, fmt.Errorf("--subproject es obligatorio")
	}
	return subproject, save, deterministicOnly, nil
}

func requiredFlag(args []string, flag string) (string, error) {
	if len(args) != 2 || args[0] != flag || args[1] == "" {
		return "", fmt.Errorf("uso: %s <path>", flag)
	}
	return args[1], nil
}
