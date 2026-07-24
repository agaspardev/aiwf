package sonar

import (
	"os"
	"path/filepath"
	"testing"
)

func TestParseGate(t *testing.T) {
	status, err := parseGate([]byte(`{"projectStatus":{"status":"ERROR"}}`))
	if err != nil {
		t.Fatal(err)
	}
	if status != "ERROR" {
		t.Errorf("status = %q, want ERROR", status)
	}
}

func TestParseIssues(t *testing.T) {
	data := []byte(`{"total":2,"issues":[
		{"severity":"BLOCKER","message":"npe","component":"a.go","line":10},
		{"severity":"CRITICAL","message":"leak","component":"b.go","line":5}
	]}`)
	total, issues, err := parseIssues(data)
	if err != nil {
		t.Fatal(err)
	}
	if total != 2 || len(issues) != 2 {
		t.Fatalf("total=%d issues=%d", total, len(issues))
	}
	if issues[0].Severity != "BLOCKER" || issues[0].Line != 10 {
		t.Errorf("issue[0] = %+v", issues[0])
	}
}

func TestLoadConfigFromVaultConfig(t *testing.T) {
	root := t.TempDir()
	envDir := filepath.Join(root, ".ai-workflow", "env")
	if err := os.MkdirAll(envDir, 0o755); err != nil {
		t.Fatal(err)
	}
	must(t, os.WriteFile(filepath.Join(envDir, "vault-config.local.json"),
		[]byte(`{"sonarqube":{"enabled":true,"host":"http://sonar:9000","projectKey":"demo"}}`), 0o644))
	must(t, os.WriteFile(filepath.Join(envDir, ".env.local"), []byte("SONAR_TOKEN=abc123\n"), 0o644))

	// Aislar de un SONAR_TOKEN heredado del entorno.
	t.Setenv("SONAR_TOKEN", "")

	cfg := LoadConfig(root)
	if !cfg.Enabled || cfg.Host != "http://sonar:9000" || cfg.ProjectKey != "demo" {
		t.Errorf("cfg = %+v", cfg)
	}
	if cfg.Token != "abc123" {
		t.Errorf("token = %q, want abc123 (desde .env.local)", cfg.Token)
	}
}

func TestLoadConfigDefaults(t *testing.T) {
	root := t.TempDir()
	cfg := LoadConfig(root)
	if cfg.Enabled {
		t.Error("sin config, enabled debería ser false")
	}
	if cfg.Host != "http://localhost:9000" {
		t.Errorf("host default = %q", cfg.Host)
	}
	if cfg.ProjectKey != filepath.Base(root) {
		t.Errorf("projectKey default = %q, want %q", cfg.ProjectKey, filepath.Base(root))
	}
}

func TestEnvTokenTakesPrecedence(t *testing.T) {
	root := t.TempDir()
	t.Setenv("SONAR_TOKEN", "from-env")
	if got := resolveToken(root, "SONAR_TOKEN"); got != "from-env" {
		t.Errorf("resolveToken = %q, want from-env", got)
	}
}

func must(t *testing.T, err error) {
	t.Helper()
	if err != nil {
		t.Fatal(err)
	}
}
