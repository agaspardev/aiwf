package diag

import "testing"

func TestDefaultChecksHasCriticalClaude(t *testing.T) {
	var claude *Check
	for i := range DefaultChecks() {
		if DefaultChecks()[i].Command == "claude" {
			c := DefaultChecks()[i]
			claude = &c
		}
	}
	if claude == nil || !claude.Critical {
		t.Fatal("Claude Code debería estar y ser crítica")
	}
}

func TestRunUsesLookPath(t *testing.T) {
	orig := lookPath
	defer func() { lookPath = orig }()
	lookPath = func(cmd string) (string, error) {
		if cmd == "claude" {
			return "/usr/bin/claude", nil
		}
		return "", execNotFound{}
	}
	got := Run([]Check{
		{Name: "Claude Code", Command: "claude", Critical: true},
		{Name: "Trivy", Command: "trivy"},
	})
	if !got[0].Present {
		t.Error("claude debería estar presente")
	}
	if got[1].Present {
		t.Error("trivy no debería estar presente")
	}
	if m := CriticalMissing(got); len(m) != 0 {
		t.Errorf("no debería faltar nada crítico: %v", m)
	}
}

func TestCriticalMissing(t *testing.T) {
	checks := []Check{
		{Name: "Claude Code", Command: "claude", Critical: true, Present: false},
		{Name: "Trivy", Command: "trivy", Present: false},
	}
	m := CriticalMissing(checks)
	if len(m) != 1 || m[0].Command != "claude" {
		t.Errorf("CriticalMissing = %v, want [claude]", m)
	}
}

type execNotFound struct{}

func (execNotFound) Error() string { return "not found" }
