package main

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"time"

	"github.com/agaspardev/aiwf/internal/assets"
	"github.com/agaspardev/aiwf/internal/config"
	"github.com/agaspardev/aiwf/internal/harness"
	"github.com/agaspardev/aiwf/internal/omniroute"
	"github.com/agaspardev/aiwf/internal/sonar"
	"github.com/agaspardev/aiwf/internal/workspace"
)

// loadModes lee modes.json desde la instalación (si existe) o desde los assets embebidos.
func loadModes() (*harness.Modes, error) {
	if data, ok := readInstalledOrEmbedded("harness/modes.json"); ok {
		return harness.LoadModes(data)
	}
	return nil, fmt.Errorf("no se pudo cargar modes.json")
}

// loadCapabilities lee model-capabilities.json desde la instalación o assets embebidos.
func loadCapabilities() (harness.ModelCapabilities, error) {
	if data, ok := readInstalledOrEmbedded("harness/model-capabilities.json"); ok {
		return harness.LoadModelCapabilities(data)
	}
	return harness.ModelCapabilities{}, fmt.Errorf("no se pudo cargar model-capabilities.json")
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
func runMode(modeName, subproject string, dryRun, skipPerms bool) int {
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
	repositoryRoot, cwdErr := os.Getwd()
	if cwdErr != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", cwdErr)
		return 1
	}
	paths, pathErr := workspace.NewPaths(repositoryRoot, subproject, "")
	if pathErr != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", pathErr)
		return 2
	}
	if _, statErr := os.Stat(paths.ProjectManifest); statErr != nil {
		if os.IsNotExist(statErr) {
			fmt.Fprintf(os.Stderr, "error: subproject %q no existe; falta %s\n", subproject, paths.ProjectManifest)
			return 2
		}
		fmt.Fprintf(os.Stderr, "error: %v\n", statErr)
		return 1
	}

	mcpArg := ""
	mcp := filepath.Join(root, "mcp", "servers.json")
	if _, statErr := os.Stat(mcp); statErr == nil {
		mcpArg = mcp
	}

	omniActive, baseURL, apiKey := detectOmniRoute()
	if launchErr := harness.ValidateLaunchRequirements(mode, omniActive); launchErr != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", launchErr)
		return 1
	}
	// F1/1B: permission-mode DERIVADO de certificación. Sin OmniRoute no hay
	// combos que certificar -> supervisado (fail-closed).
	permMode := "default"
	if omniActive {
		capabilities, capabilitiesErr := loadCapabilities()
		if capabilitiesErr != nil {
			fmt.Fprintf(os.Stderr, "error: %v\n", capabilitiesErr)
			return 1
		}
		comboContext, cancelComboRequest := context.WithTimeout(context.Background(), 5*time.Second)
		comboState, comboErr := omniroute.ListCombos(comboContext, baseURL, apiKey)
		cancelComboRequest()
		if comboErr != nil {
			fmt.Fprintf(os.Stderr, "error validando combos: %v\n", comboErr)
			return 1
		}
		combos := make([]harness.ComboDefinition, 0, len(comboState))
		for _, combo := range comboState {
			combos = append(combos, harness.ComboDefinition{Name: combo.Name, Models: combo.Models})
		}
		violations, validationErr := harness.ValidateSelectedMode(modeName, modes, combos, capabilities)
		if validationErr != nil {
			fmt.Fprintf(os.Stderr, "error validando modo: %v\n", validationErr)
			return 1
		}
		if len(violations) > 0 {
			fmt.Fprintf(os.Stderr, "error: configuración de routing insegura:\n%s\n", harness.FormatCapabilityViolations(violations))
			return 1
		}
		permMode = harness.DerivePermissionMode(mode, combos, capabilities)
	}

	contract := harness.BuildContract(harness.ContractParams{
		InstanceRoot:     root,
		ModeName:         modeName,
		Mode:             mode,
		OmniStatus:       omniStatus(omniActive),
		SonarStatus:      sonarStatus(),
		Gates:            gates.Resolve("code"),
		Subproject:       subproject,
		ProjectRoot:      paths.Project,
		KnowledgeShared:  paths.KnowledgeShared,
		KnowledgeProject: paths.KnowledgeProject,
		ChangeRoot:       paths.Changes,
		PermissionMode:   permMode,
	})

	args := harness.BuildClaudeArgs(mode, harness.LaunchOptions{
		OmniActive:      omniActive,
		SkipPermissions: skipPerms,
		VaultDir:        vaultDir,
		MCPConfig:       mcpArg,
		Contract:        contract,
		PermissionMode:  permMode,
	})

	if dryRun {
		printDryRun(modeName, args, contract)
		return 0
	}

	if omniActive {
		os.Setenv("ANTHROPIC_BASE_URL", baseURL)
		os.Setenv("ANTHROPIC_API_KEY", apiKey)
	}
	os.Setenv("AIWF_WORKSPACE_ROOT", paths.Workflow)
	os.Setenv("AIWF_SUBPROJECT", subproject)
	os.Setenv("AIWF_PROJECT_ROOT", paths.Project)
	os.Setenv("AIWF_KNOWLEDGE_SHARED_ROOT", paths.KnowledgeShared)
	os.Setenv("AIWF_KNOWLEDGE_PROJECT_ROOT", paths.KnowledgeProject)
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
