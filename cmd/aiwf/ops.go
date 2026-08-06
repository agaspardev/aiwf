package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/agaspardev/aiwf/internal/diag"
	"github.com/agaspardev/aiwf/internal/initproj"
	"github.com/agaspardev/aiwf/internal/state"
)

// initCmd inicializa la estructura de aiwf en el proyecto actual.
func initCmd(args []string) int {
	root, err := os.Getwd()
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		return 1
	}
	name := filepath.Base(root)
	force := false
	for i := 0; i < len(args); i++ {
		switch args[i] {
		case "--force", "-f":
			force = true
		case "--name":
			if i+1 >= len(args) {
				fmt.Fprintln(os.Stderr, "error: --name requiere un valor")
				return 2
			}
			name = args[i+1]
			i++
		default:
			fmt.Fprintf(os.Stderr, "error: argumento desconocido: %s\n", args[i])
			return 2
		}
	}

	fmt.Printf("aiwf init — inicializando: %s\n", name)
	rep, err := initproj.Init(root, name, force)
	if err != nil {
		fmt.Fprintf(os.Stderr, "  error: %v\n", err)
		return 1
	}
	for _, c := range rep.Created {
		fmt.Printf("  + %s\n", c)
	}
	for _, w := range rep.Warnings {
		fmt.Printf("  ! %s\n", w)
	}
	fmt.Printf("\n  init completado (%d creados, %d ya existían, %d avisos)\n",
		len(rep.Created), len(rep.Skipped), len(rep.Warnings))
	fmt.Println("  Próximo paso: aiwf project new <subproject>")
	return 0
}

func projectCmd(args []string) int {
	if len(args) < 2 || args[0] != "new" {
		fmt.Fprintln(os.Stderr, "uso: aiwf project new <subproject> [--force]")
		return 2
	}
	name, force := args[1], false
	for _, arg := range args[2:] {
		if arg != "--force" && arg != "-f" {
			fmt.Fprintf(os.Stderr, "error: argumento desconocido: %s\n", arg)
			return 2
		}
		force = true
	}
	root, err := os.Getwd()
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		return 1
	}
	report, err := initproj.InitProject(root, name, force)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		return 1
	}
	for _, created := range report.Created {
		fmt.Printf("  + %s\n", created)
	}
	for _, skipped := range report.Skipped {
		fmt.Printf("  = %s\n", skipped)
	}
	return 0
}

// estado muestra el estado scoped de un subproyecto o change.
func estado(args []string) int {
	scopeRequest, rest, err := parseScopedArgs(args)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		return 2
	}
	if len(rest) > 1 || (len(rest) == 1 && scopeRequest.Subproject != "") {
		fmt.Fprintln(os.Stderr, "uso: aiwf estado <subproject> [--change <change>]")
		return 2
	}
	if len(rest) == 1 {
		scopeRequest.Subproject = rest[0]
	}
	cwd, err := os.Getwd()
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		return 1
	}
	requireChange := scopeRequest.Change != ""
	scope, err := resolveScope(cwd, scopeRequest, requireChange)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		return 2
	}
	if scope.Change == "" {
		sums, listErr := state.Summarize(scope.Paths.Changes)
		if listErr != nil {
			fmt.Fprintf(os.Stderr, "error leyendo changes: %v\n", listErr)
			return 1
		}
		fmt.Printf("Estado del subproyecto — %s\n", scope.Subproject)
		active, completed := 0, 0
		for _, s := range sums {
			marker := ""
			if s.Inferred {
				marker = " (inferida)"
			}
			status := s.Status
			if status == "" {
				status = "-"
			}
			fmt.Printf("  - %-40s fase=%s%s  estado=%s\n", s.Name, s.Phase, marker, status)
			switch {
			case s.Phase == "archive" || s.Status == "archived" || s.Status == "completed":
				completed++
			default:
				active++
			}
		}
		fmt.Printf("\n  Resumen: %d activos, %d completados (%d total)\n", active, completed, len(sums))
		return 0
	}

	s, ok, err := state.Load(scope.Paths.Change)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error leyendo estado: %v\n", err)
		return 1
	}
	if !ok {
		fmt.Printf("No hay state.yaml para %s/%s.\n", scope.Subproject, scope.Change)
		return 0
	}
	fmt.Printf("Estado del workflow — %s/%s\n", s.Subproject, s.Change)
	fmt.Printf("  Fase:  %s\n", s.Phase)
	fmt.Printf("  Estado: %s\n", s.Status)
	return 0
}

// diagnostico verifica el toolchain operativo con degradación graceful (harness -Doctor).
func diagnostico() int {
	checks := diag.Run(diag.DefaultChecks())
	fmt.Println("aiwf diagnostico")
	for _, c := range checks {
		icon := "[OK]      "
		if !c.Present {
			icon = "[DEGRADED]"
		}
		fmt.Printf("  %s %s\n", icon, c.Name)
	}
	if missing := diag.CriticalMissing(checks); len(missing) > 0 {
		names := make([]string, len(missing))
		for i, m := range missing {
			names[i] = m.Name
		}
		fmt.Printf("\n  CRITICO: %s no disponible. El workflow no puede operar.\n", strings.Join(names, ", "))
		return 1
	}
	fmt.Println("\n  Estado: OPERATIVO")
	fmt.Println("  Herramientas ausentes aparecen como DEGRADED (no ERROR); degradación graceful.")
	return 0
}
