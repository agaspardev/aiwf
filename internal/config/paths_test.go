package config

import (
	"path/filepath"
	"strings"
	"testing"
)

func TestInstallRootEnvOverride(t *testing.T) {
	t.Setenv("AIWF_INSTALL_ROOT", "/custom/root")
	got, err := InstallRoot()
	if err != nil || got != "/custom/root" {
		t.Fatalf("InstallRoot() = %q, %v", got, err)
	}
}

func TestInstallRootDefault(t *testing.T) {
	t.Setenv("AIWF_INSTALL_ROOT", "")
	got, err := InstallRoot()
	if err != nil {
		t.Fatalf("InstallRoot: %v", err)
	}
	if filepath.Base(got) != ".claude" {
		t.Errorf("InstallRoot default = %q, esperaba terminar en .claude", got)
	}
}

func TestManifestPathEnvOverride(t *testing.T) {
	t.Setenv("AIWF_MANIFEST", "/custom/manifest.json")
	got, err := ManifestPath()
	if err != nil || got != "/custom/manifest.json" {
		t.Fatalf("ManifestPath() = %q, %v", got, err)
	}
}

func TestManifestPathDefault(t *testing.T) {
	t.Setenv("AIWF_MANIFEST", "")
	got, err := ManifestPath()
	if err != nil {
		t.Fatalf("ManifestPath: %v", err)
	}
	if !strings.HasSuffix(filepath.ToSlash(got), ".aiwf/manifest.json") {
		t.Errorf("ManifestPath default = %q", got)
	}
}
