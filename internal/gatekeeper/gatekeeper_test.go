package gatekeeper

import (
	"os"
	"path/filepath"
	"testing"
)

func writeContract(t *testing.T, dir, body string) string {
	t.Helper()
	p := filepath.Join(dir, "contract.json")
	if err := os.WriteFile(p, []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
	return p
}

func TestValidatePasses(t *testing.T) {
	root := t.TempDir()
	// artefacto declarado que SÍ existe
	must(t, os.WriteFile(filepath.Join(root, "out.md"), []byte("x"), 0o644))
	c := writeContract(t, root, `{
		"phase":"apply","status":"passed","executive_summary":"ok",
		"artifacts":[{"path":"out.md"}],"next_phase":"verify","risks":[]
	}`)
	res, err := Validate(root, c)
	if err != nil {
		t.Fatal(err)
	}
	if !res.Passed() {
		t.Errorf("debería pasar, problems=%v", res.Problems)
	}
	if res.ArtifactCount != 1 || res.NextPhase != "verify" {
		t.Errorf("res = %+v", res)
	}
}

func TestValidateMissingArtifact(t *testing.T) {
	root := t.TempDir()
	c := writeContract(t, root, `{
		"phase":"apply","status":"passed","executive_summary":"ok",
		"artifacts":[{"path":"nope.md"}],"next_phase":"verify"
	}`)
	res, _ := Validate(root, c)
	if res.Passed() {
		t.Error("no debería pasar con artefacto inexistente")
	}
}

func TestValidatePassedWithoutArtifacts(t *testing.T) {
	root := t.TempDir()
	c := writeContract(t, root, `{
		"phase":"apply","status":"passed","executive_summary":"ok",
		"artifacts":[],"next_phase":"verify"
	}`)
	res, _ := Validate(root, c)
	if res.Passed() {
		t.Error("status=passed sin artefactos (fase!=explore) debería bloquear")
	}
}

func TestValidateCriticalRisk(t *testing.T) {
	root := t.TempDir()
	must(t, os.WriteFile(filepath.Join(root, "a"), []byte("x"), 0o644))
	c := writeContract(t, root, `{
		"phase":"apply","status":"passed","executive_summary":"ok",
		"artifacts":[{"path":"a"}],"next_phase":"verify",
		"risks":[{"severity":"critical"}]
	}`)
	res, _ := Validate(root, c)
	if res.Passed() {
		t.Error("riesgo critico con status=passed debería bloquear")
	}
}

func must(t *testing.T, err error) {
	t.Helper()
	if err != nil {
		t.Fatal(err)
	}
}
