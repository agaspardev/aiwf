// Package sonar integra SonarQube (portado de sonar-scan.ps1): consulta quality gate e
// issues vía API y lanza sonar-scanner para análisis. Credenciales SIEMPRE desde entorno,
// nunca hardcodeadas. HTTP y exec son inyectables para tests.
package sonar

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

// Config resuelve el destino SonarQube del proyecto.
type Config struct {
	Enabled    bool
	Host       string
	ProjectKey string
	Token      string
}

type vaultConfig struct {
	SonarQube struct {
		Enabled     bool   `json:"enabled"`
		Host        string `json:"host"`
		ProjectKey  string `json:"projectKey"`
		TokenEnvVar string `json:"tokenEnvVar"`
	} `json:"sonarqube"`
}

// LoadConfig lee la config de SonarQube del proyecto (vault-config.local.json) y
// resuelve el token desde entorno o .env.local.
func LoadConfig(projectRoot string) Config {
	cfg := Config{Host: "http://localhost:9000"}
	tokenVar := "SONAR_TOKEN"

	vcPath := filepath.Join(projectRoot, ".ai-workflow", "env", "vault-config.local.json")
	if data, err := os.ReadFile(vcPath); err == nil {
		var vc vaultConfig
		if json.Unmarshal(data, &vc) == nil {
			cfg.Enabled = vc.SonarQube.Enabled
			if vc.SonarQube.Host != "" {
				cfg.Host = vc.SonarQube.Host
			}
			cfg.ProjectKey = vc.SonarQube.ProjectKey
			if vc.SonarQube.TokenEnvVar != "" {
				tokenVar = vc.SonarQube.TokenEnvVar
			}
		}
	}
	if cfg.ProjectKey == "" {
		cfg.ProjectKey = filepath.Base(projectRoot)
	}
	cfg.Token = resolveToken(projectRoot, tokenVar)
	return cfg
}

func resolveToken(projectRoot, tokenVar string) string {
	if v := os.Getenv(tokenVar); v != "" {
		return v
	}
	data, err := os.ReadFile(filepath.Join(projectRoot, ".ai-workflow", "env", ".env.local"))
	if err != nil {
		return ""
	}
	for _, line := range strings.Split(string(data), "\n") {
		if v, ok := strings.CutPrefix(strings.TrimSpace(line), tokenVar+"="); ok {
			return strings.TrimSpace(v)
		}
	}
	return ""
}

// ── HTTP (inyectable) ────────────────────────────────────────────────────────

var httpClient = &http.Client{Timeout: 10 * time.Second}

func doGet(ctx context.Context, url, token string) ([]byte, int, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, 0, err
	}
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}
	resp, err := httpClient.Do(req)
	if err != nil {
		return nil, 0, err
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(io.LimitReader(resp.Body, 5<<20))
	return body, resp.StatusCode, nil
}

// ── Parseo (puro, testeable) ─────────────────────────────────────────────────

func parseGate(data []byte) (string, error) {
	var r struct {
		ProjectStatus struct {
			Status string `json:"status"`
		} `json:"projectStatus"`
	}
	if err := json.Unmarshal(data, &r); err != nil {
		return "", err
	}
	return r.ProjectStatus.Status, nil
}

// Issue es un issue de SonarQube.
type Issue struct {
	Severity  string
	Message   string
	Component string
	Line      int
}

func parseIssues(data []byte) (int, []Issue, error) {
	var r struct {
		Total  int `json:"total"`
		Issues []struct {
			Severity  string `json:"severity"`
			Message   string `json:"message"`
			Component string `json:"component"`
			Line      int    `json:"line"`
		} `json:"issues"`
	}
	if err := json.Unmarshal(data, &r); err != nil {
		return 0, nil, err
	}
	out := make([]Issue, len(r.Issues))
	for i, s := range r.Issues {
		out[i] = Issue{Severity: s.Severity, Message: s.Message, Component: s.Component, Line: s.Line}
	}
	return r.Total, out, nil
}

// ── Operaciones ──────────────────────────────────────────────────────────────

// GateStatus consulta el quality gate del proyecto (OK / ERROR / ...).
func (c Config) GateStatus(ctx context.Context) (string, error) {
	url := fmt.Sprintf("%s/api/qualitygates/project_status?projectKey=%s", c.Host, c.ProjectKey)
	body, code, err := doGet(ctx, url, c.Token)
	if err != nil {
		return "", err
	}
	if code != http.StatusOK {
		return "", fmt.Errorf("sonar gate: status %d", code)
	}
	return parseGate(body)
}

// Issues lista los issues BLOCKER/CRITICAL sin resolver.
func (c Config) Issues(ctx context.Context) (int, []Issue, error) {
	url := fmt.Sprintf("%s/api/issues/search?projectKeys=%s&resolved=false&severities=BLOCKER,CRITICAL&ps=50",
		c.Host, c.ProjectKey)
	body, code, err := doGet(ctx, url, c.Token)
	if err != nil {
		return 0, nil, err
	}
	if code != http.StatusOK {
		return 0, nil, fmt.Errorf("sonar issues: status %d", code)
	}
	return parseIssues(body)
}

// Reachable indica si el servidor SonarQube responde (para D6: status en el contract).
func Reachable(ctx context.Context, host string) bool {
	ctx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()
	_, code, err := doGet(ctx, host+"/api/system/status", "")
	return err == nil && code == http.StatusOK
}

// runScanner se puede sustituir en tests.
var runScanner = func(args ...string) int {
	cmd := exec.Command("sonar-scanner", args...)
	cmd.Stdout, cmd.Stderr = os.Stdout, os.Stderr
	if err := cmd.Run(); err != nil {
		var ee *exec.ExitError
		if errors.As(err, &ee) {
			return ee.ExitCode()
		}
		return -1
	}
	return 0
}

// Scan ejecuta sonar-scanner (full = análisis completo; si no, incremental sobre main).
func (c Config) Scan(full bool, sources string) int {
	args := []string{
		"-Dsonar.projectKey=" + c.ProjectKey,
		"-Dsonar.sources=" + sources,
		"-Dsonar.host.url=" + c.Host,
		"-Dsonar.token=" + c.Token,
	}
	if !full {
		args = append(args, "-Dsonar.newCode.referenceBranch=main")
	}
	return runScanner(args...)
}
