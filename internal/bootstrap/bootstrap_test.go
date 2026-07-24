package bootstrap

import "testing"

func TestPackageManagerCandidatesByOS(t *testing.T) {
	cases := map[string]PackageManager{
		"windows": Winget,
		"darwin":  Homebrew,
		"linux":   Apt, // primera preferencia en linux
	}
	for goos, wantFirst := range cases {
		got := packageManagerCandidates(goos)
		if len(got) == 0 {
			t.Fatalf("%s: sin candidatos", goos)
		}
		if got[0] != wantFirst {
			t.Errorf("%s: primer candidato = %q, quiero %q", goos, got[0], wantFirst)
		}
	}
}

func TestCheckPrereqsReturnsAllRequired(t *testing.T) {
	got := CheckPrereqs()
	if len(got) != len(requiredCommands) {
		t.Fatalf("CheckPrereqs devolvió %d, quiero %d", len(got), len(requiredCommands))
	}
	for i, p := range got {
		if p.Name != requiredCommands[i].Name {
			t.Errorf("posición %d: nombre = %q, quiero %q", i, p.Name, requiredCommands[i].Name)
		}
	}
}

func TestMissingFiltersPresent(t *testing.T) {
	in := []Prereq{
		{Name: "A", Present: true},
		{Name: "B", Present: false},
		{Name: "C", Present: false},
	}
	got := Missing(in)
	if len(got) != 2 {
		t.Fatalf("Missing devolvió %d, quiero 2", len(got))
	}
	if got[0].Name != "B" || got[1].Name != "C" {
		t.Errorf("Missing = %+v, quiero [B C]", got)
	}
}

func TestMissingEmptyWhenAllPresent(t *testing.T) {
	in := []Prereq{{Name: "A", Present: true}, {Name: "B", Present: true}}
	if got := Missing(in); len(got) != 0 {
		t.Errorf("Missing = %+v, quiero vacío", got)
	}
}
