package overlay

import "strings"

import "testing"

func TestInjectBlockAppendsWhenAbsent(t *testing.T) {
	base := "contenido de gentle-ai\n"
	got := InjectBlock(base, "reglas aiwf")
	if !strings.Contains(got, "contenido de gentle-ai") {
		t.Error("se perdió el contenido base de gentle-ai")
	}
	if !strings.Contains(got, "reglas aiwf") || !HasBlock(got) {
		t.Error("no se inyectó el bloque aiwf")
	}
}

func TestInjectBlockIdempotent(t *testing.T) {
	base := "base gentle-ai\n"
	once := InjectBlock(base, "payload")
	twice := InjectBlock(once, "payload")
	if once != twice {
		t.Errorf("InjectBlock no es idempotente:\n once=%q\n twice=%q", once, twice)
	}
}

func TestInjectBlockReplacesPayload(t *testing.T) {
	base := "base\n"
	v1 := InjectBlock(base, "viejo")
	v2 := InjectBlock(v1, "nuevo")
	if strings.Contains(v2, "viejo") {
		t.Error("no se reemplazó el payload viejo")
	}
	if !strings.Contains(v2, "nuevo") {
		t.Error("falta el payload nuevo")
	}
	if strings.Count(v2, markerStart) != 1 {
		t.Errorf("debería haber exactamente un bloque, hay %d", strings.Count(v2, markerStart))
	}
}

func TestInjectBlockEmptyBase(t *testing.T) {
	got := InjectBlock("", "solo aiwf")
	if !HasBlock(got) || !strings.Contains(got, "solo aiwf") {
		t.Error("bloque no inyectado sobre base vacía")
	}
}

func TestRemoveBlockKeepsBase(t *testing.T) {
	base := "linea gentle-ai"
	withBlock := InjectBlock(base, "aiwf")
	removed := RemoveBlock(withBlock)
	if HasBlock(removed) {
		t.Error("el bloque no fue removido")
	}
	if !strings.Contains(removed, "linea gentle-ai") {
		t.Error("se perdió el contenido base al remover el bloque")
	}
}
