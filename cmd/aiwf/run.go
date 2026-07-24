package main

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/agaspardev/aiwf/internal/assets"
	"github.com/agaspardev/aiwf/internal/config"
	"github.com/agaspardev/aiwf/internal/harness"
	"github.com/agaspardev/aiwf/internal/omniroute"
	"github.com/agaspardev/aiwf/internal/sonar"
)

// loadModes lee modes.json desde la instalación (si existe) o desde los assets embebidos.
func loadModes() (*harness.Modes, error) {
	if data, ok := readInstalledOrEmbedded("harness/modes.json"); ok {
		return harness.LoadModes(data)
	}
	return nil, fmt.Errorf("no se pudo cargar modes.json")
}

// loadGates lee quality-gates.json desde la instalación o los assets embebidos.
func loadGates() (harness.Gates, error) {
	if data, ok := readInstalledOrEmbedded("harness/quality-gates.json"); ok {
		return harness.LoadGates(data)
	}
	return nil, fmt.Errorf("no se pudo cargar quality-gates.json")
}

// readInstalledOrEmbedded devuelve el contenido de deployPath preferentemente desde la
// instalación (para respetar ediciones del usuario) y si no, desde los assets embebidos.
func readInstalledOrEmbedded(deployPath string) ([]byte, bool) {
	if root, err := config.InstallRoot(); err == nil {
		if data, rerr := os.ReadFile(filepath.Join(root, filepath.FromSlash(deployPath))); rerr == nil {
			return data, true
		}
	}
	if data, err := assets.ReadFile(deployPath); err == nil {
		return data, true
	}
	return nil, false
}

// runMode resuelve un modo y lanza (o simula, con dryRun) una sesión de claude.
func runMode(modeName string, dryRun, skipPerms bool) int {
	modes, err := loadModes()
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		return 1
	}
	if modeName == "" {
		modeName = modes.DefaultMode
	}
	mode, err := modes.Resolve(modeName)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		return 1
	}
	gates, err := loadGates()
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		return 1
	}

	root, _ := config.InstallRoot()
	vaultDir := filepath.Join(root, "vault")

	mcpArg := ""
	mcp := filepath.Join(root, "mcp", "servers.json")
	if _, statErr := os.Stat(mcp); statErr == nil {
		mcpArg = mcp
	}

	omniActive, baseURL, apiKey := detectOmniRoute()

	contract := harness.BuildContract(harness.ContractParams{
		InstanceRoot: root,
		ModeName:     modeName,
		Mode:         mode,
		OmniStatus:   omniStatus(omniActive),
		SonarStatus:  sonarStatus(),
		Gates:        gates.Resolve(mode.GateSet),
	})

	args := harness.BuildClaudeArgs(mode, harness.LaunchOptions{
		OmniActive:      omniActive,
		SkipPermissions: skipPerms,
		VaultDir:        vaultDir,
		MCPConfig:       mcpArg,
		Contract:        contract,
	})

	if dryRun {
		printDryRun(modeName, args, contract)
		return 0
	}

	if omniActive {
		os.Setenv("ANTHROPIC_BASE_URL", baseURL)
		os.Setenv("ANTHROPIC_API_KEY", apiKey)
	}
	cmd := exec.CommandContext(context.Background(), "claude", args...)
	cmd.Stdin, cmd.Stdout, cmd.Stderr = os.Stdin, os.Stdout, os.Stderr
	if err := cmd.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "error lanzando claude: %v\n", err)
		return 1
	}
	return 0
}

// printDryRun muestra los argumentos y el contract sin lanzar claude.
func printDryRun(modeName string, args []string, contract string) {
	fmt.Printf("[harness] DryRun modo %q — argumentos para claude:\n", modeName)
	for _, a := range args {
		if a == contract {
			fmt.Println("  <append-system-prompt: contract (abajo)>")
			continue
		}
		fmt.Printf("  %s\n", a)
	}
	fmt.Println("\nContract prompt:")
	fmt.Println(contract)
}

// sonarStatus resuelve el estado real de SonarQube para el proyecto actual (cierra D6).
func sonarStatus() string {
	cwd, err := os.Getwd()
	if err != nil {
		return "OFFLINE (degraded)"
	}
	cfg := sonar.LoadConfig(cwd)
	if cfg.Enabled && sonar.Reachable(context.Background(), cfg.Host) {
		return "ONLINE at " + cfg.Host
	}
	return "OFFLINE (degraded)"
}

func omniStatus(active bool) string {
	if active {
		return "ONLINE"
	}
	return "OFFLINE (using default model)"
}

// detectOmniRoute indica si OmniRoute está disponible: binario en PATH + API key en
// ~/.omniroute/.env. Devuelve (activo, baseURL, apiKey).
func detectOmniRoute() (bool, string, string) {
	if _, err := exec.LookPath("omniroute"); err != nil {
		return false, "", ""
	}
	key := omniroute.ReadKey()
	if key == "" {
		return false, "", ""
	}
	return true, omniroute.DefaultURL, key
}
