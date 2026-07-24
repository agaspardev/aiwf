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
			if i+1 < len(args) {
				name = args[i+1]
				i++
			}
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
	fmt.Println("  Próximos pasos: completá .ai-workflow/env/.env.local y vault-config.local.json")
	return 0
}

// estado muestra el estado del workflow del proyecto actual (harness -ShowState).
func estado() int {
	cwd, err := os.Getwd()
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		return 1
	}
	s, ok, err := state.Load(cwd)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error leyendo estado: %v\n", err)
		return 1
	}
	if !ok {
		fmt.Println("No hay estado de workflow en el directorio actual.")
		fmt.Println("Ejecutá 'aiwf init' para inicializar el proyecto.")
		return 0
	}
	fmt.Printf("Estado del workflow — %s\n", s.Project)
	fmt.Printf("  Fase:           %s\n", s.Phase)
	fmt.Printf("  Tareas activas: %d\n", len(s.ActiveTasks))
	fmt.Printf("  Completadas:    %d\n", len(s.CompletedTasks))
	fmt.Printf("  Bloqueadas:     %d\n", len(s.BlockedTasks))
	fmt.Printf("  Siguiente:      %s\n", s.NextAction)
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
