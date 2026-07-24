package main

import (
	"context"
	"fmt"
	"os"

	"github.com/agaspardev/aiwf/internal/omniroute"
)

// omnirouteCmd maneja `aiwf omniroute [ensure]`. Default: ensure.
func omnirouteCmd(args []string) int {
	action := "ensure"
	if len(args) > 0 {
		action = args[0]
	}
	switch action {
	case "ensure":
		return omnirouteEnsure()
	default:
		fmt.Fprintf(os.Stderr, "aiwf omniroute: subcomando desconocido %q (usá: ensure)\n", action)
		return 2
	}
}

// omnirouteEnsure ejecuta la matriz de instalación y reporta la acción tomada.
func omnirouteEnsure() int {
	fmt.Println("aiwf omniroute ensure")
	act, err := omniroute.Ensure(context.Background())
	if err != nil {
		fmt.Fprintf(os.Stderr, "  error: %v\n", err)
		return 1
	}
	printOmnirouteAction("  ", act)
	return 0
}

// ensureOmniroute es el paso complementario dentro de `aiwf install`: NO aborta el
// install si omniroute falla (es complementario a gentle-ai) — reporta y continúa.
func ensureOmniroute() {
	act, err := omniroute.Ensure(context.Background())
	if err != nil {
		fmt.Fprintf(os.Stderr, "    aviso: omniroute no asegurado: %v (continuando)\n", err)
		return
	}
	printOmnirouteAction("    ", act)
}

// printOmnirouteAction imprime la acción con la sangría dada.
func printOmnirouteAction(indent string, act omniroute.Action) {
	switch act {
	case omniroute.Skip:
		fmt.Printf("%somniroute ya presente (skip)\n", indent)
	case omniroute.Install:
		fmt.Printf("%somniroute instalado\n", indent)
	}
}
