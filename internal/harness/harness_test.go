package harness

import (
	"strings"
	"testing"
)

const sampleModes = `{
  "schemaVersion": 1,
  "defaultMode": "automatico",
  "modes": {
    "automatico": {"combo":"agent-auto","permissionMode":"acceptEdits","risk":"adaptive","gateSet":"code","directives":["d1","d2"]},
    "gratis": {"combo":"free-first","permissionMode":"default","risk":"experimental","gateSet":"documentation","directives":["e1"]}
  }
}`

const sampleGates = `{"global":["g1","g2"],"code":["c1"],"documentation":["doc1"]}`

func TestResolveMode(t *testing.T) {
	m, err := LoadModes([]byte(sampleModes))
	if err != nil {
		t.Fatalf("LoadModes: %v", err)
	}
	mode, err := m.Resolve("automatico")
	if err != nil {
		t.Fatalf("Resolve: %v", err)
	}
	if mode.Combo != "agent-auto" || mode.PermissionMode != "acceptEdits" {
		t.Errorf("modo mal parseado: %+v", mode)
	}
	if _, err := m.Resolve("inexistente"); err == nil {
		t.Error("esperaba error para modo inexistente")
	}
}

func TestBuildClaudeArgsWithOmni(t *testing.T) {
	mode := Mode{Combo: "agent-auto", PermissionMode: "acceptEdits"}
	args := BuildClaudeArgs(mode, LaunchOptions{
		OmniActive: true,
		VaultDir:   "/vault",
		MCPConfig:  "/mcp.json",
		Contract:   "CONTRACT",
	})
	joined := strings.Join(args, " ")
	for _, want := range []string{"--model agent-auto", "--permission-mode acceptEdits", "--add-dir /vault", "--mcp-config /mcp.json", "--append-system-prompt CONTRACT"} {
		if !strings.Contains(joined, want) {
			t.Errorf("args no contienen %q\n got: %s", want, joined)
		}
	}
}

func TestBuildClaudeArgsWithoutOmniAndSkipPerms(t *testing.T) {
	mode := Mode{Combo: "agent-auto", PermissionMode: "acceptEdits"}
	args := BuildClaudeArgs(mode, LaunchOptions{OmniActive: false, SkipPermissions: true})
	joined := strings.Join(args, " ")
	if strings.Contains(joined, "--model") {
		t.Error("sin OmniRoute no debería pasar --model")
	}
	if !strings.Contains(joined, "--dangerously-skip-permissions") {
		t.Error("con SkipPermissions falta --dangerously-skip-permissions")
	}
	if strings.Contains(joined, "--permission-mode") {
		t.Error("con SkipPermissions no debería pasar --permission-mode")
	}
}

func TestGatesResolve(t *testing.T) {
	g, err := LoadGates([]byte(sampleGates))
	if err != nil {
		t.Fatalf("LoadGates: %v", err)
	}
	got := g.Resolve("code")
	want := []string{"g1", "g2", "c1"}
	if strings.Join(got, ",") != strings.Join(want, ",") {
		t.Errorf("Resolve(code) = %v, want %v", got, want)
	}
}

func TestBuildContractIncludesModeAndGates(t *testing.T) {
	mode := Mode{Combo: "agent-auto", Risk: "adaptive", GateSet: "code", Directives: []string{"do X", "do Y"}}
	c := BuildContract(ContractParams{
		InstanceRoot: "/root",
		ModeName:     "automatico",
		Mode:         mode,
		OmniStatus:   "ONLINE",
		SonarStatus:  "OFFLINE (degraded)",
		Gates:        []string{"g1", "c1"},
	})
	for _, want := range []string{
		"version " + ContractVersion,
		"Mode: automatico | Combo: agent-auto",
		"- do X", "- do Y",
		"QUALITY GATES (code):",
		"- g1", "- c1",
		"MANDATORY BEHAVIOR:",
	} {
		if !strings.Contains(c, want) {
			t.Errorf("contract no contiene %q", want)
		}
	}
}
