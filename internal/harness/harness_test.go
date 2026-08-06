package harness

import (
	"strings"
	"testing"

	"github.com/agaspardev/aiwf/internal/assets"
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

func TestValidateLaunchRequirementsRejectsCapabilityModeWithoutOmniRoute(t *testing.T) {
	mode := Mode{Combo: "gpt-agent", CapabilityGate: true}
	if err := ValidateLaunchRequirements(mode, false); err == nil || !strings.Contains(err.Error(), "requiere OmniRoute") {
		t.Fatalf("ValidateLaunchRequirements error = %v", err)
	}
}

func TestValidateLaunchRequirementsAllowsGenericModeWithoutOmniRoute(t *testing.T) {
	mode := Mode{Combo: "agent-auto", CapabilityGate: false}
	if err := ValidateLaunchRequirements(mode, false); err != nil {
		t.Fatalf("ValidateLaunchRequirements: %v", err)
	}
}

func TestValidateLaunchRequirementsAllowsCapabilityModeWithOmniRoute(t *testing.T) {
	mode := Mode{Combo: "gpt-agent", CapabilityGate: true}
	if err := ValidateLaunchRequirements(mode, true); err != nil {
		t.Fatalf("ValidateLaunchRequirements: %v", err)
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

func TestValidateMutableModesRejectsAuxiliaryPrimary(t *testing.T) {
	modes, err := LoadModes([]byte(`{
		"schemaVersion": 1,
		"defaultMode": "gpt",
		"modes": {
			"gpt": {
				"combo": "gpt-only",
				"permissionMode": "acceptEdits",
				"auxiliaryCombos": ["gpt-web-auxiliary"]
			}
		}
	}`))
	if err != nil {
		t.Fatalf("LoadModes: %v", err)
	}
	capabilities, err := LoadModelCapabilities([]byte(`{
		"schemaVersion": 1,
		"models": {
			"cgpt-web/gpt-5.5": {
				"class": "auxiliary",
				"toolCall": false,
				"localRead": false,
				"localWrite": false
			}
		}
	}`))
	if err != nil {
		t.Fatalf("LoadModelCapabilities: %v", err)
	}

	violations := ValidateMutableModes(modes, []ComboDefinition{{
		Name:   "gpt-only",
		Models: []string{"cgpt-web/gpt-5.5"},
	}}, capabilities)

	if len(violations) != 1 {
		t.Fatalf("violations = %v, want 1", violations)
	}
	if got := violations[0]; got.Mode != "gpt" || got.Combo != "gpt-only" || got.Model != "cgpt-web/gpt-5.5" {
		t.Fatalf("unexpected violation: %+v", got)
	}
}

func TestValidateMutableModesFailsClosedForUnknownModel(t *testing.T) {
	modes, err := LoadModes([]byte(`{
		"schemaVersion": 1,
		"defaultMode": "codigo",
		"modes": {
			"codigo": {"combo": "coding", "permissionMode": "acceptEdits"}
		}
	}`))
	if err != nil {
		t.Fatalf("LoadModes: %v", err)
	}

	violations := ValidateMutableModes(modes, []ComboDefinition{{
		Name:   "coding",
		Models: []string{"cx/unknown"},
	}}, ModelCapabilities{Models: map[string]ModelCapability{}})

	if len(violations) != 1 || violations[0].Reason != "model capability is not registered" {
		t.Fatalf("violations = %+v", violations)
	}
}

func TestValidateMutableModesAcceptsCertifiedAgent(t *testing.T) {
	modes, err := LoadModes([]byte(`{
		"schemaVersion": 1,
		"defaultMode": "gpt",
		"modes": {
			"gpt": {
				"combo": "gpt-agent",
				"permissionMode": "acceptEdits",
				"auxiliaryCombos": ["gpt-web-auxiliary"]
			}
		}
	}`))
	if err != nil {
		t.Fatalf("LoadModes: %v", err)
	}
	capabilities := ModelCapabilities{Models: map[string]ModelCapability{
		"cx/gpt-5.6-sol": {
			Class:      CapabilityAgent,
			ToolCall:   true,
			LocalRead:  true,
			LocalWrite: true,
		},
	}}

	violations := ValidateMutableModes(modes, []ComboDefinition{{
		Name:   "gpt-agent",
		Models: []string{"cx/gpt-5.6-sol"},
	}}, capabilities)

	if len(violations) != 0 {
		t.Fatalf("violations = %+v, want none", violations)
	}
}

func TestValidateAuxiliaryPolicyRequiresMutationAndVerificationGuard(t *testing.T) {
	mode := Mode{
		Combo:           "gpt-agent",
		PermissionMode:  "acceptEdits",
		AuxiliaryCombos: []string{"gpt-web-auxiliary"},
		Directives: []string{
			"Use gpt-web-auxiliary for bounded read-only consultations",
		},
	}

	violations := ValidateAuxiliaryPolicy("gpt", mode)
	if len(violations) != 2 {
		t.Fatalf("violations = %+v, want mutation and verification guards", violations)
	}
}

func TestValidateAuxiliaryPolicyAcceptsExplicitGuards(t *testing.T) {
	mode := Mode{
		Combo:           "gpt-agent",
		PermissionMode:  "acceptEdits",
		AuxiliaryCombos: []string{"gpt-web-auxiliary"},
		Directives: []string{
			"Never delegate filesystem mutations to an auxiliary combo",
			"Never delegate final verification to an auxiliary combo",
		},
	}

	if violations := ValidateAuxiliaryPolicy("gpt", mode); len(violations) != 0 {
		t.Fatalf("violations = %+v, want none", violations)
	}
}

func TestEmbeddedModesDefineSafeGPTOrchestration(t *testing.T) {
	data, err := assets.ReadFile("harness/modes.json")
	if err != nil {
		t.Fatalf("ReadFile modes: %v", err)
	}
	modes, err := LoadModes(data)
	if err != nil {
		t.Fatalf("LoadModes: %v", err)
	}
	mode, err := modes.Resolve("gpt")
	if err != nil {
		t.Fatalf("Resolve gpt: %v", err)
	}
	if mode.Combo != "gpt-agent" {
		t.Fatalf("gpt combo = %q, want gpt-agent", mode.Combo)
	}
	if got := strings.Join(mode.AuxiliaryCombos, ","); got != "gpt-web-auxiliary" {
		t.Fatalf("gpt auxiliaries = %q", got)
	}
	if violations := ValidateAuxiliaryPolicy("gpt", mode); len(violations) != 0 {
		t.Fatalf("gpt policy violations = %+v", violations)
	}
}

func TestEmbeddedModelCapabilitiesCertifyCodexAndRestrictWeb(t *testing.T) {
	data, err := assets.ReadFile("harness/model-capabilities.json")
	if err != nil {
		t.Fatalf("ReadFile capabilities: %v", err)
	}
	capabilities, err := LoadModelCapabilities(data)
	if err != nil {
		t.Fatalf("LoadModelCapabilities: %v", err)
	}
	for _, model := range []string{"cx/gpt-5.6-sol", "cx/gpt-5.6-sol-high"} {
		capability := capabilities.Models[model]
		if capability.Class != CapabilityAgent || !capability.ToolCall || !capability.LocalRead || !capability.LocalWrite {
			t.Fatalf("%s is not a certified agent: %+v", model, capability)
		}
	}
	for _, model := range []string{"cgpt-web/gpt-5.5", "cgpt-web/gpt-5.5-pro"} {
		capability := capabilities.Models[model]
		if capability.Class != CapabilityAuxiliary || capability.ToolCall || capability.LocalRead || capability.LocalWrite {
			t.Fatalf("%s is not restricted to auxiliary: %+v", model, capability)
		}
	}
}

func TestValidateConfiguredModesCombinesCapabilityAndPolicyViolations(t *testing.T) {
	modes, err := LoadModes([]byte(`{
		"schemaVersion": 1,
		"defaultMode": "gpt",
		"modes": {
			"gpt": {
				"combo": "gpt-web",
				"permissionMode": "acceptEdits",
				"auxiliaryCombos": ["gpt-web"]
			}
		}
	}`))
	if err != nil {
		t.Fatalf("LoadModes: %v", err)
	}
	capabilities := ModelCapabilities{Models: map[string]ModelCapability{
		"cgpt-web/gpt-5.5": {Class: CapabilityAuxiliary},
	}}

	violations := ValidateConfiguredModes(modes, []ComboDefinition{{
		Name: "gpt-web", Models: []string{"cgpt-web/gpt-5.5"},
	}}, capabilities)

	if len(violations) != 3 {
		t.Fatalf("violations = %+v, want capability plus two policy violations", violations)
	}
}

func TestFormatCapabilityViolationsIsDeterministic(t *testing.T) {
	got := FormatCapabilityViolations([]CapabilityViolation{
		{Mode: "gpt", Combo: "gpt-agent", Reason: "z"},
		{Mode: "automatico", Combo: "agent-auto", Model: "m", Reason: "a"},
	})
	want := "automatico/agent-auto/m: a\ngpt/gpt-agent: z"
	if got != want {
		t.Fatalf("FormatCapabilityViolations() = %q, want %q", got, want)
	}
}

func TestValidateSelectedModeIgnoresUnselectedUnregisteredModels(t *testing.T) {
	modes, err := LoadModes([]byte(`{
		"schemaVersion": 1,
		"defaultMode": "gpt",
		"modes": {
			"gpt": {"combo": "gpt-agent", "permissionMode": "acceptEdits"},
			"other": {"combo": "other-agent", "permissionMode": "acceptEdits"}
		}
	}`))
	if err != nil {
		t.Fatalf("LoadModes: %v", err)
	}
	capabilities := ModelCapabilities{Models: map[string]ModelCapability{
		"cx/gpt-5.6-sol": {
			Class: CapabilityAgent, ToolCall: true, LocalRead: true, LocalWrite: true,
		},
	}}
	combos := []ComboDefinition{
		{Name: "gpt-agent", Models: []string{"cx/gpt-5.6-sol"}},
		{Name: "other-agent", Models: []string{"unregistered/model"}},
	}

	violations, err := ValidateSelectedMode("gpt", modes, combos, capabilities)
	if err != nil {
		t.Fatalf("ValidateSelectedMode: %v", err)
	}
	if len(violations) != 0 {
		t.Fatalf("violations = %+v, want none", violations)
	}
}

func TestValidateSelectedModeRejectsUnknownMode(t *testing.T) {
	modes, err := LoadModes([]byte(`{
		"schemaVersion": 1,
		"defaultMode": "gpt",
		"modes": {"gpt": {"combo": "gpt-agent", "permissionMode": "acceptEdits"}}
	}`))
	if err != nil {
		t.Fatalf("LoadModes: %v", err)
	}

	if _, err := ValidateSelectedMode("missing", modes, nil, ModelCapabilities{}); err == nil {
		t.Fatal("expected unknown mode error")
	}
}

func TestValidateSelectedModeSkipsCapabilityGateWhenModeDoesNotOptIn(t *testing.T) {
	modes, err := LoadModes([]byte(`{
		"schemaVersion": 1,
		"defaultMode": "automatico",
		"modes": {
			"automatico": {
				"combo": "agent-auto",
				"permissionMode": "acceptEdits"
			}
		}
	}`))
	if err != nil {
		t.Fatalf("LoadModes: %v", err)
	}

	violations, err := ValidateSelectedMode("automatico", modes, []ComboDefinition{{
		Name: "agent-auto", Models: []string{"unregistered/model"},
	}}, ModelCapabilities{})
	if err != nil {
		t.Fatalf("ValidateSelectedMode: %v", err)
	}
	if len(violations) != 0 {
		t.Fatalf("violations = %+v, want none for mode without capability gate", violations)
	}
}

func TestEmbeddedGPTOptsIntoCapabilityGate(t *testing.T) {
	data, err := assets.ReadFile("harness/modes.json")
	if err != nil {
		t.Fatalf("ReadFile modes: %v", err)
	}
	modes, err := LoadModes(data)
	if err != nil {
		t.Fatalf("LoadModes: %v", err)
	}
	mode, err := modes.Resolve("gpt")
	if err != nil {
		t.Fatalf("Resolve gpt: %v", err)
	}
	if !mode.CapabilityGate {
		t.Fatal("gpt mode must enable capabilityGate")
	}
}
