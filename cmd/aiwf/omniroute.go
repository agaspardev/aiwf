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
	case "configure":
		return omnirouteConfigure(args[1:])
	default:
		fmt.Fprintf(os.Stderr, "aiwf omniroute: subcomando desconocido %q (usá: ensure | configure)\n", action)
		return 2
	}
}

// omnirouteConfigure reporta el estado de configuración de omniroute y los pasos
// faltantes (read-only). Con --doctor, delega en `omniroute doctor`.
func omnirouteConfigure(args []string) int {
	ctx := context.Background()
	fmt.Println("aiwf omniroute configure")
	st := omniroute.CheckStatus(ctx)
	fmt.Printf("  server: %s | api key: %s\n", upDown(st.ServerUp), presentAbsent(st.KeyPresent))
	for _, step := range omniroute.Guidance(st) {
		fmt.Printf("  - %s\n", step)
	}

	if hasFlag(args, "--doctor") {
		fmt.Println("  omniroute doctor:")
		if err := omniroute.RunDoctor(ctx); err != nil {
			fmt.Fprintf(os.Stderr, "  error corriendo omniroute doctor: %v\n", err)
			return 1
		}
	}
	return 0
}

func upDown(up bool) string {
	if up {
		return "arriba"
	}
	return "abajo"
}

func presentAbsent(p bool) string {
	if p {
		return "presente"
	}
	return "ausente"
}

func hasFlag(args []string, flag string) bool {
	for _, a := range args {
		if a == flag {
			return true
		}
	}
	return false
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
