package main

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/agaspardev/aiwf/internal/omniroute"
)

// omnirouteCmd maneja `aiwf omniroute [subcomando]`. Default: status.
func omnirouteCmd(args []string) int {
	action := "status"
	if len(args) > 0 {
		action = args[0]
	}
	switch action {
	case "ensure":
		return omnirouteEnsure()
	case "configure":
		return omnirouteConfigure(args[1:])
	case "status":
		return omnirouteStatus(args[1:])
	case "providers":
		return omnirouteProviders()
	case "usage":
		return omnirouteUsage()
	case "doctor":
		return omnirouteDoctor()
	case "combos":
		return omnirouteCombos()
	default:
		fmt.Fprintf(os.Stderr, "aiwf omniroute: subcomando desconocido %q (usá: status | configure | providers | combos | usage | doctor | ensure)\n", action)
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

// omnirouteStatus muestra el estado detallado de OmniRoute.
func omnirouteStatus(args []string) int {
	ctx := context.Background()
	d := omniroute.CheckDetailedStatus(ctx)

	fmt.Println("OmniRoute status")
	fmt.Printf("  Server:        %s\n", upDown(d.ServerUp))
	fmt.Printf("  API key:       %s\n", presentAbsent(d.KeyPresent))
	fmt.Printf("  MCP:           %s (%s)\n", boolStr(d.MCPOnline), d.MCPTransport)
	fmt.Printf("  Compression:   %s (mode: %s)\n", boolStr(d.CompressionOn), d.CompressionMode)
	fmt.Printf("  Providers:     %d active / %d total\n", d.ProvidersActive, d.ProviderCount)
	fmt.Printf("  Cache:         %d hits, %d misses (%d entries)\n", d.CacheHits, d.CacheMisses, d.CacheSize)
	fmt.Printf("  Calls (24h):   %d\n", d.TotalCalls24h)

	if hasFlag(args, "--doctor") {
		fmt.Println("  Doctor:")
		if err := omniroute.RunDoctor(ctx); err != nil {
			fmt.Fprintf(os.Stderr, "    doctor error: %v\n", err)
			return 1
		}
	}
	return 0
}

// omnirouteProviders lista los providers configurados.
func omnirouteProviders() int {
	ctx := context.Background()
	apiKey := omniroute.ReadKey()
	if apiKey == "" {
		fmt.Fprintln(os.Stderr, "OmniRoute API key no encontrada (corré `aiwf omniroute configure`)")
		return 1
	}
	providers, err := omniroute.ListProviders(ctx, omniroute.DefaultURL, apiKey)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error listing providers: %v\n", err)
		return 1
	}
	fmt.Printf("Providers (%d):\n", len(providers))
	fmt.Print(omniroute.PrintProvidersTable(providers))
	fmt.Println()
	for _, s := range omniroute.SummarizeProviders(providers) {
		fmt.Printf("  %-20s %d active / %d inactive\n", s.Type, s.Active, s.Inactive)
	}
	return 0
}

// omnirouteUsage muestra el reporte de uso/costos.
func omnirouteUsage() int {
	ctx := context.Background()
	apiKey := omniroute.ReadKey()
	if apiKey == "" {
		fmt.Fprintln(os.Stderr, "OmniRoute API key no encontrada")
		return 1
	}
	report, err := omniroute.GetUsage(ctx, omniroute.DefaultURL, apiKey)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error getting usage: %v\n", err)
		return 1
	}
	fmt.Print(omniroute.PrintUsage(report, startTime))
	return 0
}

// startTime se usa para reportes de uso desde el inicio de la sesión/instalación.
var startTime = time.Now()

// omnirouteCombos lista los combos de routing configurados.
func omnirouteCombos() int {
	ctx := context.Background()
	apiKey := omniroute.ReadKey()
	if apiKey == "" {
		fmt.Fprintln(os.Stderr, "OmniRoute API key no encontrada")
		return 1
	}
	combos, err := omniroute.ListCombos(ctx, omniroute.DefaultURL, apiKey)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error listing combos: %v\n", err)
		return 1
	}
	fmt.Printf("Combos (%d):\n", len(combos))
	for _, c := range combos {
		fmt.Printf("  %-30s %v\n", c.Name, c.Models)
	}
	return 0
}

// omnirouteDoctor ejecuta el diagnóstico de OmniRoute.
func omnirouteDoctor() int {
	ctx := context.Background()
	if err := omniroute.RunDoctor(ctx); err != nil {
		fmt.Fprintf(os.Stderr, "doctor error: %v\n", err)
		return 1
	}
	return 0
}

func boolStr(v bool) string {
	if v {
		return "✓"
	}
	return "✗"
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
