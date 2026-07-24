package omniroute

import (
	"context"
	"os/exec"
	"reflect"
	"testing"
)

// withFakes reemplaza los seams inyectables durante t y los restaura al terminar.
func withFakes(t *testing.T, lp func(string) (string, error), rv func(context.Context) (string, error), rc func(context.Context, string, ...string) error) {
	t.Helper()
	origLP, origRV, origRC := lookPath, runVersion, runCommand
	if lp != nil {
		lookPath = lp
	}
	if rv != nil {
		runVersion = rv
	}
	if rc != nil {
		runCommand = rc
	}
	t.Cleanup(func() { lookPath, runVersion, runCommand = origLP, origRV, origRC })
}

// found simula que solo los binarios en present están en PATH.
func found(present ...string) func(string) (string, error) {
	set := map[string]bool{}
	for _, p := range present {
		set[p] = true
	}
	return func(name string) (string, error) {
		if set[name] {
			return "/fake/" + name, nil
		}
		return "", exec.ErrNotFound
	}
}

func TestDetect(t *testing.T) {
	// Presente: captura RawVersion.
	withFakes(t,
		found("omniroute"),
		func(context.Context) (string, error) { return "3.8.48", nil },
		nil,
	)
	st := Detect(context.Background())
	if !st.Installed || st.RawVersion != "3.8.48" {
		t.Fatalf("Detect presente = %+v, want Installed con RawVersion 3.8.48", st)
	}

	// Ausente: Installed=false, sin llamar runVersion.
	withFakes(t,
		found(), // nada en PATH
		func(context.Context) (string, error) {
			t.Fatal("runVersion no debía llamarse si omniroute está ausente")
			return "", nil
		},
		nil,
	)
	if st := Detect(context.Background()); st.Installed {
		t.Errorf("Detect ausente = %+v, want Installed=false", st)
	}
}

func TestExtractVersion(t *testing.T) {
	// Salida real: omniroute mete logs de env (con ANSI) por stdout antes del semver.
	noisy := "  \x1b[2m📋 Loaded env from C:\\Users\\anton\\.omniroute\\.env\x1b[0m\n  \x1b[2m📋 Loaded env\x1b[0m\n3.8.48\n"
	if got := extractVersion(noisy); got != "3.8.48" {
		t.Errorf("extractVersion(ruidoso) = %q, want 3.8.48", got)
	}
	if got := extractVersion("v1.2.3\n"); got != "1.2.3" {
		t.Errorf("extractVersion(v1.2.3) = %q, want 1.2.3", got)
	}
	if got := extractVersion("sin version"); got != "sin version" {
		t.Errorf("extractVersion(sin semver) = %q, want fallback recortado", got)
	}
}

func TestDecide(t *testing.T) {
	if Decide(true) != Skip {
		t.Error("Decide(true) debe ser Skip")
	}
	if Decide(false) != Install {
		t.Error("Decide(false) debe ser Install")
	}
}

func TestDetectJSPackageManager(t *testing.T) {
	cases := []struct {
		name    string
		present []string
		want    JSPackageManager
	}{
		{"pnpm presente gana", []string{"pnpm", "npm"}, Pnpm},
		{"solo npm", []string{"npm"}, Npm},
		{"ninguno", nil, NoJSPM},
	}
	for _, c := range cases {
		withFakes(t, found(c.present...), nil, nil)
		if got := DetectJSPackageManager(); got != c.want {
			t.Errorf("%s: got %q, want %q", c.name, got, c.want)
		}
	}
}

func TestInstallCommand(t *testing.T) {
	pnpm, err := InstallCommand(Pnpm)
	if err != nil {
		t.Fatalf("pnpm: error inesperado %v", err)
	}
	wantPnpm := []string{"add", "-g", "omniroute@latest", "--allow-build=better-sqlite3", "--allow-build=@swc/core"}
	if pnpm.Prog != "pnpm" || !reflect.DeepEqual(pnpm.Args, wantPnpm) {
		t.Errorf("pnpm cmd = %+v, want prog=pnpm args=%v", pnpm, wantPnpm)
	}

	npm, err := InstallCommand(Npm)
	if err != nil {
		t.Fatalf("npm: error inesperado %v", err)
	}
	if npm.Prog != "npm" || !reflect.DeepEqual(npm.Args, []string{"install", "-g", "omniroute"}) {
		t.Errorf("npm cmd = %+v, want npm install -g omniroute", npm)
	}

	if _, err := InstallCommand(NoJSPM); err == nil {
		t.Error("InstallCommand(NoJSPM) debía devolver error")
	}
}

func TestEnsureAusenteConPnpm(t *testing.T) {
	var gotProg string
	var gotArgs []string
	withFakes(t,
		found("pnpm"), // omniroute ausente, pnpm presente
		nil,
		func(_ context.Context, prog string, args ...string) error {
			gotProg, gotArgs = prog, args
			return nil // NO instala de verdad
		},
	)
	act, err := Ensure(context.Background())
	if err != nil {
		t.Fatalf("Ensure error: %v", err)
	}
	if act != Install {
		t.Errorf("action = %v, want Install", act)
	}
	if gotProg != "pnpm" || len(gotArgs) == 0 || gotArgs[0] != "add" {
		t.Errorf("comando ejecutado = %q %v, want pnpm add ...", gotProg, gotArgs)
	}
}

func TestEnsurePresenteSkip(t *testing.T) {
	withFakes(t,
		found("omniroute"),
		func(context.Context) (string, error) { return "3.8.48", nil },
		func(context.Context, string, ...string) error {
			t.Fatal("no debía ejecutarse instalación si omniroute ya está presente")
			return nil
		},
	)
	act, err := Ensure(context.Background())
	if err != nil || act != Skip {
		t.Errorf("Ensure presente = (%v, %v), want (Skip, nil)", act, err)
	}
}

func TestEnsureSinPackageManager(t *testing.T) {
	withFakes(t,
		found(), // ni omniroute ni pnpm/npm
		nil,
		func(context.Context, string, ...string) error {
			t.Fatal("no debía intentar ejecutar sin gestor JS")
			return nil
		},
	)
	act, err := Ensure(context.Background())
	if err == nil {
		t.Error("Ensure sin pnpm/npm debía devolver error")
	}
	if act != Install {
		t.Errorf("action = %v, want Install (intención, aunque falló)", act)
	}
}
