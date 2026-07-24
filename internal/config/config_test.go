package config

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestDefault(t *testing.T) {
	cfg := Default()
	if cfg.GentleAIVersion != PinnedGentleAIVersion {
		t.Errorf("GentleAIVersion = %q, want %q", cfg.GentleAIVersion, PinnedGentleAIVersion)
	}
}

func TestLoadConfig(t *testing.T) {
	dir := t.TempDir()
	p := filepath.Join(dir, "config.json")
	must(t, os.WriteFile(p, []byte(`{"gentle_ai_version":"0.4.5"}`), 0o644))

	cfg, err := LoadConfig(p)
	must(t, err)
	if cfg.GentleAIVersion != "0.4.5" {
		t.Errorf("GentleAIVersion = %q, want 0.4.5", cfg.GentleAIVersion)
	}
}

func TestLoadConfigNoFile(t *testing.T) {
	cfg, err := LoadConfig("/nonexistent/path.json")
	must(t, err)
	if cfg.GentleAIVersion != "" {
		t.Errorf("sin archivo debería devolver Config vacío, got %+v", cfg)
	}
}

func TestLoadConfigInvalidJSON(t *testing.T) {
	dir := t.TempDir()
	p := filepath.Join(dir, "bad.json")
	must(t, os.WriteFile(p, []byte(`no-json`), 0o644))

	_, err := LoadConfig(p)
	if err == nil {
		t.Fatal("esperaba error por JSON inválido")
	}
}

func TestLoadOrDefaultPicksDefaults(t *testing.T) {
	cfg := LoadOrDefault("/nonexistent/path.json")
	if cfg.GentleAIVersion != PinnedGentleAIVersion {
		t.Errorf("sin archivo debería devolver default, got %q", cfg.GentleAIVersion)
	}
}

func TestLoadOrDefaultMergesFile(t *testing.T) {
	dir := t.TempDir()
	p := filepath.Join(dir, "config.json")
	must(t, os.WriteFile(p, []byte(`{"gentle_ai_version":"1.2.3"}`), 0o644))

	cfg := LoadOrDefault(p)
	if cfg.GentleAIVersion != "1.2.3" {
		t.Errorf("debería leer del archivo, got %q", cfg.GentleAIVersion)
	}
}

func TestLoadOrDefaultPreservesDefaultsOnPartialFile(t *testing.T) {
	dir := t.TempDir()
	p := filepath.Join(dir, "config.json")
	must(t, os.WriteFile(p, []byte(`{}`), 0o644))

	cfg := LoadOrDefault(p)
	if cfg.GentleAIVersion != PinnedGentleAIVersion {
		t.Errorf("archivo vacío debería usar default, got %q", cfg.GentleAIVersion)
	}
}

func TestConfigFilePathEnvOverride(t *testing.T) {
	t.Setenv("AIWF_CONFIG", "/custom/aiwf.json")
	if got := ConfigFilePath(); got != "/custom/aiwf.json" {
		t.Errorf("ConfigFilePath = %q, want /custom/aiwf.json", got)
	}
}

func TestConfigFilePathDefault(t *testing.T) {
	want := filepath.Join(".aiwf", "config.json")
	got := ConfigFilePath()
	if !strings.HasSuffix(got, want) {
		t.Errorf("ConfigFilePath = %q, should end with %q", got, want)
	}
}

func must(t *testing.T, err error) {
	t.Helper()
	if err != nil {
		t.Fatal(err)
	}
}
