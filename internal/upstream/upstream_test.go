package upstream

import (
	"testing"
	"time"
)

func TestParseVersion(t *testing.T) {
	cases := []struct {
		in      string
		want    Version
		wantErr bool
	}{
		{"1.2.3", Version{1, 2, 3}, false},
		{"v0.4.10", Version{0, 4, 10}, false},
		{"gentle-ai version v1.4.2 (build abc)", Version{1, 4, 2}, false},
		{"sin numeros", Version{}, true},
		{"", Version{}, true},
	}
	for _, c := range cases {
		got, err := ParseVersion(c.in)
		if (err != nil) != c.wantErr {
			t.Errorf("ParseVersion(%q) err=%v, wantErr=%v", c.in, err, c.wantErr)
			continue
		}
		if !c.wantErr && got != c.want {
			t.Errorf("ParseVersion(%q) = %v, want %v", c.in, got, c.want)
		}
	}
}

func TestVersionCompare(t *testing.T) {
	cases := []struct {
		a, b Version
		want int
	}{
		{Version{1, 0, 0}, Version{1, 0, 0}, 0},
		{Version{1, 2, 0}, Version{1, 1, 9}, 1},
		{Version{1, 1, 9}, Version{1, 2, 0}, -1},
		{Version{0, 9, 9}, Version{1, 0, 0}, -1},
		{Version{2, 0, 0}, Version{1, 9, 9}, 1},
		{Version{1, 2, 3}, Version{1, 2, 4}, -1},
	}
	for _, c := range cases {
		if got := c.a.Compare(c.b); got != c.want {
			t.Errorf("%v.Compare(%v) = %d, want %d", c.a, c.b, got, c.want)
		}
	}
}

func TestDecide(t *testing.T) {
	target := Version{1, 2, 0}
	cases := []struct {
		name      string
		installed bool
		current   Version
		want      Action
	}{
		{"ausente", false, Version{}, Install},
		{"desactualizado", true, Version{1, 1, 0}, Update},
		{"igual al pin", true, Version{1, 2, 0}, Skip},
		{"mayor al pin", true, Version{1, 3, 0}, Skip},
	}
	for _, c := range cases {
		if got := Decide(c.installed, c.current, target); got != c.want {
			t.Errorf("%s: Decide = %v, want %v", c.name, got, c.want)
		}
	}
}

func TestEligibleForAdoption(t *testing.T) {
	now := time.Date(2026, 7, 23, 0, 0, 0, 0, time.UTC)
	fresh := now.Add(-3 * 24 * time.Hour)   // 3 días: dentro de la ventana
	mature := now.Add(-20 * 24 * time.Hour) // 20 días: fuera de la ventana

	cases := []struct {
		name string
		r    Release
		want bool
	}{
		{"reciente no elegible", Release{Version{1, 1, 0}, fresh, false}, false},
		{"madura elegible", Release{Version{1, 1, 0}, mature, false}, true},
		{"reciente pero seguridad", Release{Version{1, 1, 0}, fresh, true}, true},
		{"justo en el borde", Release{Version{1, 1, 0}, now.Add(-ProbationWindow), false}, true},
	}
	for _, c := range cases {
		if got := EligibleForAdoption(c.r, now); got != c.want {
			t.Errorf("%s: EligibleForAdoption = %v, want %v", c.name, got, c.want)
		}
	}
}

func TestLatestEligible(t *testing.T) {
	now := time.Date(2026, 7, 23, 0, 0, 0, 0, time.UTC)
	fresh := now.Add(-2 * 24 * time.Hour)
	mature := now.Add(-30 * 24 * time.Hour)

	releases := []Release{
		{Version{1, 0, 0}, mature, false},
		{Version{1, 2, 0}, fresh, false},  // más nueva pero dentro de ventana → no elegible
		{Version{1, 1, 0}, mature, false}, // madura, debería ganar
	}
	got, ok := LatestEligible(releases, now)
	if !ok {
		t.Fatal("LatestEligible = none, want 1.1.0")
	}
	if got.Version != (Version{1, 1, 0}) {
		t.Errorf("LatestEligible = %v, want 1.1.0", got.Version)
	}

	// Con seguridad, la fresca 1.2.0 pasa a ser elegible y gana.
	releases[1].Security = true
	got, ok = LatestEligible(releases, now)
	if !ok || got.Version != (Version{1, 2, 0}) {
		t.Errorf("LatestEligible (con security) = %v (ok=%v), want 1.2.0", got.Version, ok)
	}

	// Ninguna elegible.
	none := []Release{{Version{9, 9, 9}, fresh, false}}
	if _, ok := LatestEligible(none, now); ok {
		t.Error("LatestEligible con solo frescas = ok, want none")
	}
}
