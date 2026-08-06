package main

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"

	"github.com/agaspardev/aiwf/internal/initproj"
	"github.com/agaspardev/aiwf/internal/workspace"
)

// errSubprojectCanceled indica que el usuario declinó crear un subproject inexistente.
var errSubprojectCanceled = errors.New("no se creó el subproject")

// isInteractive devuelve true si stdin es un terminal (no un pipe de CI/script).
var isInteractive = func() bool {
	info, err := os.Stdin.Stat()
	if err != nil {
		return false
	}
	return info.Mode()&os.ModeCharDevice != 0
}

// confirmCreate pide confirmación sí/no y devuelve true solo ante s/y (case-insensitive).
var confirmCreate = func() bool {
	fmt.Fprint(os.Stderr, "¿Crearlo ahora? [s/N]: ")
	answer, err := readLine()
	if err != nil {
		return false
	}
	switch strings.ToLower(strings.TrimSpace(answer)) {
	case "s", "si", "sí", "y", "yes":
		return true
	default:
		return false
	}
}

// resolveSessionSubproject resuelve el subproject de una sesión:
//   - requested vacío: usa el default persistido (base/personalizado); si no hay,
//     pregunta por un nombre (Enter = base) o usa "base" en CI/scripts.
//   - requested existente: lo usa sin tocar el default.
//   - requested inexistente: confirma antes de crear (protege contra typos).
func resolveSessionSubproject(root, requested string) (string, error) {
	if requested != "" {
		return resolveExplicitSubproject(root, requested)
	}
	return resolveDefaultSubproject(root)
}

func resolveExplicitSubproject(root, requested string) (string, error) {
	if err := workspace.ValidateID(requested); err != nil {
		return "", err
	}
	if subprojectExists(root, requested) {
		return requested, nil
	}
	if isInteractive() {
		fmt.Fprintf(os.Stderr, "El subproject %q no existe.\n", requested)
		if !confirmCreate() {
			return "", fmt.Errorf("%w %q", errSubprojectCanceled, requested)
		}
	} else {
		return "", fmt.Errorf("el subproject %q no existe; crealo antes con `aiwf project new %s` o usá un nombre existente", requested, requested)
	}
	if _, err := initproj.InitProject(root, requested, false); err != nil {
		return "", err
	}
	fmt.Printf("  + subproject %q creado\n", requested)
	return requested, nil
}

func resolveDefaultSubproject(root string) (string, error) {
	if m, ok, err := loadWorkspaceManifest(root); err == nil && ok && m.DefaultSubproject != "" {
		if subprojectExists(root, m.DefaultSubproject) {
			return m.DefaultSubproject, nil
		}
		// El default persiste pero su subproject no existe: lo recreamos.
		if _, err := initproj.InitProject(root, m.DefaultSubproject, false); err != nil {
			return "", err
		}
		return m.DefaultSubproject, nil
	}

	name := "base"
	if isInteractive() {
		fmt.Fprintln(os.Stderr, "No hay subproject por defecto en este workspace.")
		prompted, err := promptSubproject(root)
		if err != nil {
			return "", err
		}
		name = prompted
	}
	if err := workspace.ValidateID(name); err != nil {
		return "", fmt.Errorf("subproject por defecto: %w", err)
	}
	if !subprojectExists(root, name) {
		if _, err := initproj.InitProject(root, name, false); err != nil {
			return "", err
		}
	}
	if err := persistDefaultSubproject(root, name); err != nil {
		return "", err
	}
	return name, nil
}

func persistDefaultSubproject(root, name string) error {
	m, _, err := loadWorkspaceManifest(root)
	if err != nil {
		return err
	}
	if m == nil {
		m = &workspace.WorkspaceManifest{
			SchemaVersion: 1,
			LayoutVersion: 1,
			RepositoryID:  defaultRepositoryID(root),
		}
	}
	m.DefaultSubproject = name
	return saveWorkspaceManifest(root, m)
}

func defaultRepositoryID(root string) string {
	id := strings.ToLower(strings.TrimSpace(filepath.Base(root)))
	var b strings.Builder
	for _, r := range id {
		if r == '-' || (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') {
			b.WriteRune(r)
		}
	}
	cleaned := strings.Trim(b.String(), "-")
	if cleaned == "" {
		return "workspace"
	}
	return cleaned
}

func subprojectExists(root, name string) bool {
	_, err := os.Stat(filepath.Join(root, ".ai-workflow", "projects", name, "project.yaml"))
	return err == nil
}

// promptSubproject lee un nombre de subproject. Enter vacío = base.
func promptSubproject(root string) (string, error) {
	existing, _ := childDirectories(filepath.Join(root, ".ai-workflow", "projects"))
	hint := ""
	if len(existing) > 0 {
		hint = " (" + strings.Join(existing, ", ") + ")"
	}
	fmt.Fprintf(os.Stderr, "Subproject%s (Enter = base): ", hint)
	name, err := readLine()
	if err != nil {
		return "", err
	}
	if strings.TrimSpace(name) == "" {
		return "base", nil
	}
	return strings.TrimSpace(name), nil
}

func readLine() (string, error) {
	line, err := bufio.NewReader(os.Stdin).ReadString('\n')
	if err != nil && !errors.Is(err, io.EOF) {
		return "", err
	}
	return line, nil
}

func loadWorkspaceManifest(root string) (*workspace.WorkspaceManifest, bool, error) {
	path := filepath.Join(root, ".ai-workflow", "config", "workspace.yaml")
	data, err := os.ReadFile(path)
	if os.IsNotExist(err) {
		return nil, false, nil
	}
	if err != nil {
		return nil, false, err
	}
	var m workspace.WorkspaceManifest
	if err := yaml.Unmarshal(data, &m); err != nil {
		return nil, false, fmt.Errorf("parsear %s: %w", path, err)
	}
	if err := m.Validate(); err != nil {
		return nil, false, err
	}
	return &m, true, nil
}

func saveWorkspaceManifest(root string, m *workspace.WorkspaceManifest) error {
	if err := m.Validate(); err != nil {
		return err
	}
	data, err := yaml.Marshal(m)
	if err != nil {
		return err
	}
	path := filepath.Join(root, ".ai-workflow", "config", "workspace.yaml")
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	return os.WriteFile(path, data, 0o644)
}
