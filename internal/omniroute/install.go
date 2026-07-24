package omniroute

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"regexp"
	"strings"
)

// binaryName es el ejecutable de omniroute en PATH.
const binaryName = "omniroute"

// semverRe extrae el primer X.Y.Z de una cadena. omniroute imprime líneas de log de
// carga de entorno (con códigos ANSI) por STDOUT antes de la versión, así que TrimSpace
// no alcanza: hay que extraer el semver del ruido.
var semverRe = regexp.MustCompile(`\d+\.\d+\.\d+`)

// extractVersion devuelve el primer semver de raw, o el raw recortado si no hay ninguno.
func extractVersion(raw string) string {
	if m := semverRe.FindString(raw); m != "" {
		return m
	}
	return strings.TrimSpace(raw)
}

// Action es la decisión del instalador respecto a omniroute. Con "latest directo"
// (sin pin de versión) la matriz se reduce a Install/Skip: no hay rama Update.
type Action int

const (
	// Skip: omniroute ya presente; no se hace nada.
	Skip Action = iota
	// Install: omniroute ausente; hay que instalarlo.
	Install
)

func (a Action) String() string {
	switch a {
	case Install:
		return "install"
	case Skip:
		return "skip"
	default:
		return "unknown"
	}
}

// JSPackageManager identifica el gestor de paquetes JavaScript usado para instalar
// omniroute (Node.js). Es un eje DISTINTO al gestor del SO (winget/brew/apt): omniroute
// se distribuye por npm, no por el package manager del sistema.
type JSPackageManager string

const (
	Pnpm   JSPackageManager = "pnpm"
	Npm    JSPackageManager = "npm"
	NoJSPM JSPackageManager = ""
)

// Command representa el comando de instalación a ejecutar.
type Command struct {
	Prog string
	Args []string
}

// State es el estado detectado de omniroute en el sistema.
type State struct {
	Installed  bool
	Path       string
	RawVersion string // salida de `omniroute --version`; NO se parsea (latest directo)
}

// Seams inyectables para test (nivel paquete), al estilo de internal/security.
var (
	lookPath = exec.LookPath

	// runVersion devuelve la versión cruda de omniroute (stdout de `--version`).
	// omniroute emite logs de carga de entorno por stderr; .Output() toma solo stdout.
	runVersion = func(ctx context.Context) (string, error) {
		out, err := exec.CommandContext(ctx, binaryName, "--version").Output()
		return extractVersion(string(out)), err
	}

	// runCommand ejecuta un comando heredando stdout/stderr del proceso.
	runCommand = func(ctx context.Context, prog string, args ...string) error {
		cmd := exec.CommandContext(ctx, prog, args...)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		return cmd.Run()
	}
)

// Run ejecuta el comando de instalación.
func (c Command) Run(ctx context.Context) error {
	return runCommand(ctx, c.Prog, c.Args...)
}

// Detect busca omniroute en PATH y, si está, captura su versión cruda. No parsea
// semver: con "latest directo" la decisión no depende de la versión instalada.
func Detect(ctx context.Context) State {
	path, err := lookPath(binaryName)
	if err != nil {
		return State{Installed: false}
	}
	st := State{Installed: true, Path: path}
	if v, verr := runVersion(ctx); verr == nil {
		st.RawVersion = v
	}
	return st
}

// Decide instala si falta; si ya está, Skip. Sin rama Update (latest directo).
func Decide(installed bool) Action {
	if installed {
		return Skip
	}
	return Install
}

// DetectJSPackageManager devuelve el gestor JS preferido disponible: pnpm sobre npm
// (regla de dependencias del proyecto), o NoJSPM si ninguno está en PATH.
func DetectJSPackageManager() JSPackageManager {
	if _, err := lookPath(string(Pnpm)); err == nil {
		return Pnpm
	}
	if _, err := lookPath(string(Npm)); err == nil {
		return Npm
	}
	return NoJSPM
}

// InstallCommand construye el comando de instalación global de omniroute según el
// gestor JS. pnpm necesita --allow-build para better-sqlite3 y @swc/core (dependencias
// nativas de omniroute). Función pura: no ejecuta nada.
func InstallCommand(pm JSPackageManager) (Command, error) {
	switch pm {
	case Pnpm:
		return Command{Prog: "pnpm", Args: []string{
			"add", "-g", "omniroute@latest",
			"--allow-build=better-sqlite3", "--allow-build=@swc/core",
		}}, nil
	case Npm:
		return Command{Prog: "npm", Args: []string{"install", "-g", "omniroute"}}, nil
	default:
		return Command{}, fmt.Errorf("instalar omniroute requiere pnpm o npm en PATH")
	}
}

// Ensure detecta omniroute y lo instala si falta (idempotente). Devuelve la acción
// tomada. Si no hay gestor JS disponible, devuelve Install junto al error explicativo.
func Ensure(ctx context.Context) (Action, error) {
	if Decide(Detect(ctx).Installed) == Skip {
		return Skip, nil
	}
	cmd, err := InstallCommand(DetectJSPackageManager())
	if err != nil {
		return Install, err
	}
	return Install, cmd.Run(ctx)
}
