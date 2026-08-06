package omniroute

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
	"testing"
)

func TestCheckDetailedStatusNotConfiguredReturnsEarly(t *testing.T) {
	withConfigFakes(t, "", false) // sin key ni server
	d := CheckDetailedStatus(context.Background())
	if d.Configured() || d.MCPOnline || d.ProviderCount != 0 {
		t.Fatalf("esperaba Detailed vacío cuando no está configurado: %+v", d)
	}
}

func TestCheckDetailedStatusEnrichesFromAPIs(t *testing.T) {
	withConfigFakes(t, "k", true)
	origGet := httpGetJSON
	httpGetJSON = func(_ context.Context, url, _ string, dst any) error {
		var payload string
		switch {
		case strings.Contains(url, "/api/mcp/status"):
			payload = `{"online":true,"transport":"stdio","activity":{"totalCalls24h":42}}`
		case strings.Contains(url, "/api/settings/compression"):
			payload = `{"enabled":true,"defaultMode":"aggressive"}`
		case strings.Contains(url, "/api/providers"):
			payload = `{"connections":[{"isActive":true},{"isActive":false}]}`
		case strings.Contains(url, "/api/cache/stats"):
			payload = `{"hits":10,"misses":2,"size":5}`
		default:
			return errors.New("url inesperada: " + url)
		}
		return json.Unmarshal([]byte(payload), dst)
	}
	t.Cleanup(func() { httpGetJSON = origGet })

	d := CheckDetailedStatus(context.Background())
	if !d.Configured() {
		t.Fatal("esperaba Configured=true")
	}
	if !d.MCPOnline || d.MCPTransport != "stdio" || d.TotalCalls24h != 42 {
		t.Errorf("MCP mal poblado: %+v", d)
	}
	if !d.CompressionOn || d.CompressionMode != "aggressive" {
		t.Errorf("compresión mal poblada: %+v", d)
	}
	if d.ProviderCount != 2 || d.ProvidersActive != 1 {
		t.Errorf("providers mal contados: count=%d active=%d", d.ProviderCount, d.ProvidersActive)
	}
	if d.CacheHits != 10 || d.CacheMisses != 2 || d.CacheSize != 5 {
		t.Errorf("cache mal poblado: %+v", d)
	}
}

func TestRunDoctorDelegatesToCommand(t *testing.T) {
	origRun := runCommand
	var gotProg string
	var gotArgs []string
	runCommand = func(_ context.Context, prog string, args ...string) error {
		gotProg, gotArgs = prog, args
		return nil
	}
	t.Cleanup(func() { runCommand = origRun })

	if err := RunDoctor(context.Background()); err != nil {
		t.Fatalf("RunDoctor: %v", err)
	}
	if gotProg != binaryName || len(gotArgs) != 1 || gotArgs[0] != "doctor" {
		t.Errorf("RunDoctor invocó %q %v, esperaba %q doctor", gotProg, gotArgs, binaryName)
	}
}

func TestRunDoctorPropagatesError(t *testing.T) {
	origRun := runCommand
	runCommand = func(context.Context, string, ...string) error { return errors.New("boom") }
	t.Cleanup(func() { runCommand = origRun })

	if err := RunDoctor(context.Background()); err == nil {
		t.Error("esperaba que RunDoctor propague el error del comando")
	}
}
