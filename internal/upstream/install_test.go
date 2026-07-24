package upstream

import (
	"testing"

	"github.com/agaspardev/aiwf/internal/bootstrap"
)

func TestInstallCommand(t *testing.T) {
	target := Version{1, 4, 0}
	cases := []struct {
		name      string
		goos      string
		pm        bootstrap.PackageManager
		wantProg  string
		wantShell bool
	}{
		{"windows scoop", "windows", bootstrap.Scoop, "scoop", false},
		{"windows sin scoop → go install", "windows", bootstrap.Winget, "go", false},
		{"darwin brew", "darwin", bootstrap.Homebrew, "brew", false},
		{"linux sin brew → curl|bash", "linux", bootstrap.Unknown, "", true},
	}
	for _, c := range cases {
		cmd, err := InstallCommand(c.goos, c.pm, target)
		if err != nil {
			t.Errorf("%s: error inesperado %v", c.name, err)
			continue
		}
		if c.wantShell {
			if cmd.Shell == "" {
				t.Errorf("%s: esperaba comando de shell, got Prog=%q", c.name, cmd.Prog)
			}
			continue
		}
		if cmd.Prog != c.wantProg {
			t.Errorf("%s: Prog = %q, want %q", c.name, cmd.Prog, c.wantProg)
		}
	}
}

func TestInstallCommandUnsupportedOS(t *testing.T) {
	if _, err := InstallCommand("plan9", bootstrap.Unknown, Version{}); err == nil {
		t.Error("esperaba error para SO no soportado")
	}
}

func TestGoInstallRef(t *testing.T) {
	if got := goInstallRef(Version{}); got != "latest" {
		t.Errorf("goInstallRef(zero) = %q, want latest", got)
	}
	if got := goInstallRef(Version{1, 4, 2}); got != "v1.4.2" {
		t.Errorf("goInstallRef(1.4.2) = %q, want v1.4.2", got)
	}
}

func TestParseReleases(t *testing.T) {
	data := []byte(`[
		{"tag_name":"v1.2.0","published_at":"2026-01-01T00:00:00Z","prerelease":false,"draft":false},
		{"tag_name":"v1.3.0-rc1","published_at":"2026-02-01T00:00:00Z","prerelease":true,"draft":false},
		{"tag_name":"draft","published_at":"2026-03-01T00:00:00Z","prerelease":false,"draft":true},
		{"tag_name":"v1.1.0","published_at":"2025-12-01T00:00:00Z","prerelease":false,"draft":false}
	]`)
	got, err := parseReleases(data)
	if err != nil {
		t.Fatalf("parseReleases error: %v", err)
	}
	if len(got) != 2 {
		t.Fatalf("parseReleases devolvió %d releases, want 2 (%+v)", len(got), got)
	}
	if got[0].Version != (Version{1, 2, 0}) || got[1].Version != (Version{1, 1, 0}) {
		t.Errorf("versiones = %v, %v; want 1.2.0, 1.1.0", got[0].Version, got[1].Version)
	}
}
