package upstream

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"runtime"

	"github.com/agaspardev/aiwf/internal/bootstrap"
)

// goModule es el path del entrypoint de gentle-ai para `go install`.
const goModule = "github.com/gentleman-programming/gentle-ai/cmd/gentle-ai"

// installScriptShell es la instalación oficial para macOS/Linux (curl | bash).
const installScriptShell = "curl -fsSL https://raw.githubusercontent.com/Gentleman-Programming/gentle-ai/main/scripts/install.sh | bash"

// Command representa un comando a ejecutar. Si Shell != "" se corre como línea de
// shell (para pipes como curl|bash); si no, se ejecuta Prog con Args.
type Command struct {
	Shell string
	Prog  string
	Args  []string
}

// goInstallRef devuelve la referencia de versión para `go install`: "latest" si el
// target es cero (placeholder), o "vX.Y.Z".
func goInstallRef(target Version) string {
	if target == (Version{}) {
		return "latest"
	}
	return "v" + target.String()
}

// InstallCommand construye el comando de instalación oficial de gentle-ai según el SO
// y el gestor de paquetes detectado. Es una función pura (no ejecuta nada).
func InstallCommand(goos string, pm bootstrap.PackageManager, target Version) (Command, error) {
	switch goos {
	case "windows":
		if pm == bootstrap.Scoop {
			return Command{Prog: "scoop", Args: []string{"install", "gentle-ai"}}, nil
		}
		// Vía oficial en Windows: go install (requiere Go, garantizado por bootstrap).
		return Command{Prog: "go", Args: []string{"install", goModule + "@" + goInstallRef(target)}}, nil
	case "darwin", "linux":
		if pm == bootstrap.Homebrew {
			return Command{Prog: "brew", Args: []string{"install", "gentle-ai"}}, nil
		}
		return Command{Shell: installScriptShell}, nil
	default:
		return Command{}, fmt.Errorf("SO no soportado para instalación automática: %s", goos)
	}
}

// Run ejecuta el comando, heredando stdout/stderr del proceso.
func (c Command) Run(ctx context.Context) error {
	var cmd *exec.Cmd
	if c.Shell != "" {
		cmd = commandContext(ctx, "sh", "-c", c.Shell)
	} else {
		cmd = commandContext(ctx, c.Prog, c.Args...)
	}
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

// Ensure aplica la matriz de idempotencia: detecta gentle-ai, decide y ejecuta la
// instalación/actualización si corresponde. Devuelve la acción tomada.
func Ensure(ctx context.Context, target Version) (Action, error) {
	st, _ := Detect(ctx)
	action := Decide(st.Installed, st.Version, target)
	if action == Skip {
		return Skip, nil
	}
	cmd, err := InstallCommand(runtime.GOOS, bootstrap.DetectPackageManager(), target)
	if err != nil {
		return action, err
	}
	return action, cmd.Run(ctx)
}
