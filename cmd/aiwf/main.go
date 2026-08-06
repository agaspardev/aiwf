// Command aiwf es el instalador y CLI cross-platform del entorno de ingeniería
// asistida por IA. Es una customización declarada de gentle-ai (ver README).
package main

import (
	"context"
	"fmt"
	"os"
	"runtime"

	"github.com/agaspardev/aiwf/internal/assets"
	"github.com/agaspardev/aiwf/internal/bootstrap"
	"github.com/agaspardev/aiwf/internal/config"
	"github.com/agaspardev/aiwf/internal/overlay"
	"github.com/agaspardev/aiwf/internal/upstream"
)

func main() {
	if len(os.Args) < 2 {
		usage()
		os.Exit(2)
	}
	switch os.Args[1] {
	case "doctor":
		os.Exit(doctor())
	case "install":
		os.Exit(install())
	case "reconcile":
		os.Exit(reconcile())
	case "uninstall":
		os.Exit(uninstall())
	case "init":
		os.Exit(initCmd(os.Args[2:]))
	case "project":
		os.Exit(projectCmd(os.Args[2:]))
	case "status":
		os.Exit(statusCmd(os.Args[2:]))
	case "check":
		os.Exit(checkCmd())
	case "document":
		os.Exit(documentCmd(os.Args[2:]))
	case "gate":
		os.Exit(gateCmd(os.Args[2:]))
	case "skills":
		os.Exit(skillsCmd(os.Args[2:]))
	case "gemini":
		os.Exit(geminiCmd(os.Args[2:]))
	case "omniroute":
		os.Exit(omnirouteCmd(os.Args[2:]))
	case "security":
		os.Exit(securityCmd(os.Args[2:]))
	case "sonar":
		os.Exit(sonarCmd(os.Args[2:]))
	case "migrate":
		os.Exit(migrateCmd(os.Args[2:]))
	case "-h", "--help", "help":
		usage()
	default:
		// Cualquier otro token se interpreta como un modo de trabajo.
		subproject, dryRun, skipPerms, err := parseRunArgs(os.Args[2:])
		if err != nil {
			fmt.Fprintf(os.Stderr, "error: %v\n", err)
			os.Exit(2)
		}
		os.Exit(runMode(os.Args[1], subproject, dryRun, skipPerms))
	}
}

// parseRunArgs extrae el subproyecto obligatorio y banderas de sesión.
func parseRunArgs(args []string) (subproject string, dryRun, skipPerms bool, err error) {
	for _, arg := range args {
		switch arg {
		case "--dry-run", "--dryrun":
			dryRun = true
		case "--skip-perms":
			skipPerms = true
		default:
			if len(arg) > 0 && arg[0] == '-' {
				return "", false, false, fmt.Errorf("bandera de sesión desconocida: %s", arg)
			}
			if subproject != "" {
				return "", false, false, fmt.Errorf("uso: aiwf <modo> <subproject> [--dry-run] [--skip-perms]")
			}
			subproject = arg
		}
	}
	if subproject == "" {
		return "", false, false, fmt.Errorf("falta subproject: uso aiwf <modo> <subproject>")
	}
	return subproject, dryRun, skipPerms, nil
}

func usage() {
	fmt.Print(`aiwf — instalador y entorno de ingeniería asistida por IA

Instalación:
  aiwf doctor      Verifica prerequisitos y entorno
  aiwf install     Instala gentle-ai (base) + la capa de aiwf
  aiwf reconcile   Re-aplica la capa de aiwf tras updates de gentle-ai
  aiwf uninstall   Revierte solo lo instalado por aiwf

Sesiones (modos de trabajo):
  aiwf                 Abre una sesión en el modo por defecto
  aiwf <modo> <subproject>          Abre una sesión aislada por subproyecto
                                    (automatico, codigo, arreglar, arquitectura,
                                     documentos, profundo, seguridad, gratis, gpt,
                                     fast-fix, deep-work, research y auxiliares)
  aiwf <modo> <subproject> --dry-run     Muestra argumentos y contract sin lanzar
  aiwf <modo> <subproject> --skip-perms  Usa --dangerously-skip-permissions

Herramientas:
  aiwf init [--name N] [--force]   Inicializa control plane mínimo .ai-workflow/
  aiwf project new <subproject>    Crea únicamente el manifest del subproyecto
  aiwf status <subproject> [--change C]  Estado scoped del workflow
  aiwf check           Verifica el toolchain operativo (claude, omniroute, ...)
  aiwf document [full|update] [-s]  Documenta el proyecto (determinista, cero tokens)
  aiwf gate <c.json>   Valida un phase-contract contra el repo (determinista)
  aiwf skills [--lint] Genera el registry de skills / detecta drift
  aiwf gemini "..."    Consulta al modelo auxiliar (Gemini) vía OmniRoute
  aiwf security [scope] Pipeline AppSec: secrets|sast|sca|sbom|all (default all)
  aiwf sonar [modo]    SonarQube: gate|issues|changed|full (default gate)
  aiwf migrate layout --subproject S  Dry-run de migración legacy asistida
`)
}

// doctor verifica el entorno y devuelve un exit code (0 = OK, 1 = faltan prereqs).
func doctor() int {
	cfg := config.LoadOrDefault(config.ConfigFilePath())
	fmt.Println("aiwf doctor")
	fmt.Printf("  SO/arch:          %s/%s\n", runtime.GOOS, runtime.GOARCH)
	fmt.Printf("  gestor paquetes:  %s\n", pmDisplay(bootstrap.DetectPackageManager()))
	fmt.Printf("  gentle-ai target: %s\n", cfg.GentleAIVersion)
	fmt.Println("  prerequisitos:")

	prereqs := bootstrap.CheckPrereqs()
	for _, p := range prereqs {
		mark := "OK   "
		if !p.Present {
			mark = "FALTA"
		}
		fmt.Printf("    [%s] %s\n", mark, p.Name)
	}

	// Estado de gentle-ai y decisión de idempotencia contra el pin.
	target, _ := upstream.ParseVersion(cfg.GentleAIVersion)
	st, err := upstream.Detect(context.Background())
	fmt.Println("  gentle-ai:")
	switch {
	case err != nil:
		fmt.Printf("    instalado pero versión ilegible: %v\n", err)
	case !st.Installed:
		fmt.Println("    no instalado")
	default:
		fmt.Printf("    instalado %s (%s)\n", st.Version, st.Path)
	}
	action := upstream.Decide(st.Installed, st.Version, target)
	fmt.Printf("    acción vs pin %s: %s\n", target, action)

	if missing := bootstrap.Missing(prereqs); len(missing) > 0 {
		fmt.Printf("\n  Faltan %d prerequisito(s). Instalalos antes de continuar.\n", len(missing))
		return 1
	}
	fmt.Println("\n  Entorno OK.")
	return 0
}

// install ejecuta la matriz de idempotencia sobre gentle-ai. El overlay propio
// (tasks 0.5-0.7) se encadenará aquí una vez implementado.
func install() int {
	cfg := config.LoadOrDefault(config.ConfigFilePath())
	target, _ := upstream.ParseVersion(cfg.GentleAIVersion)

	fmt.Println("aiwf install")
	fmt.Println("  paso 1: asegurar gentle-ai (base)")
	action, err := upstream.Ensure(context.Background(), target)
	if err != nil {
		fmt.Fprintf(os.Stderr, "    error asegurando gentle-ai: %v\n", err)
		return 1
	}
	switch action {
	case upstream.Skip:
		fmt.Println("    gentle-ai ya presente y actualizado (skip)")
	case upstream.Install:
		fmt.Println("    gentle-ai instalado")
	case upstream.Update:
		fmt.Println("    gentle-ai actualizado")
	}

	fmt.Println("  paso 1b: asegurar omniroute (complementario)")
	ensureOmniroute()

	fmt.Println("  paso 2: aplicar capa aiwf (overlay)")
	if err := applyOverlay(); err != nil {
		fmt.Fprintf(os.Stderr, "    error aplicando overlay: %v\n", err)
		return 1
	}
	fmt.Println("    capa aiwf aplicada")
	return 0
}

// applyOverlay construye el overlay desde los assets embebidos y lo aplica sobre la
// raíz de instalación.
func applyOverlay() error {
	root, err := config.InstallRoot()
	if err != nil {
		return err
	}
	mp, err := config.ManifestPath()
	if err != nil {
		return err
	}
	entries, err := assets.BuildEntries()
	if err != nil {
		return err
	}
	return overlay.New(root, mp, entries).Apply()
}

// reconcile re-aplica la capa aiwf tras un update de gentle-ai (idempotente).
func reconcile() int {
	fmt.Println("aiwf reconcile")
	if err := applyOverlay(); err != nil {
		fmt.Fprintf(os.Stderr, "  error: %v\n", err)
		return 1
	}
	fmt.Println("  capa aiwf reconciliada")
	return 0
}

// uninstall revierte solo lo aplicado por aiwf, sin tocar gentle-ai.
func uninstall() int {
	fmt.Println("aiwf uninstall")
	root, err := config.InstallRoot()
	if err != nil {
		fmt.Fprintf(os.Stderr, "  error: %v\n", err)
		return 1
	}
	mp, err := config.ManifestPath()
	if err != nil {
		fmt.Fprintf(os.Stderr, "  error: %v\n", err)
		return 1
	}
	skipped, err := overlay.Uninstall(root, mp)
	if err != nil {
		fmt.Fprintf(os.Stderr, "  error: %v\n", err)
		return 1
	}
	fmt.Println("  capa aiwf removida (gentle-ai intacto)")
	for _, p := range skipped {
		fmt.Printf("  nota: %s (JSON) no se des-fusionó automáticamente; revisá manualmente\n", p)
	}
	return 0
}

func pmDisplay(pm bootstrap.PackageManager) string {
	if pm == bootstrap.Unknown {
		return "(no detectado)"
	}
	return string(pm)
}
