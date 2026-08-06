package omniroute

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

// withConfigFakes reemplaza los seams de configuración durante t.
func withConfigFakes(t *testing.T, key string, serverUp bool) {
	t.Helper()
	origKey, origPing := keyReader, pingServer
	keyReader = func() string { return key }
	pingServer = func(context.Context) bool { return serverUp }
	t.Cleanup(func() { keyReader, pingServer = origKey, origPing })
}

func TestCheckStatus(t *testing.T) {
	cases := []struct {
		name         string
		key          string
		serverUp     bool
		wantKey      bool
		wantUp       bool
		wantConfigOK bool
	}{
		{"todo OK", "k123", true, true, true, true},
		{"sin key", "", true, false, true, false},
		{"server abajo", "k123", false, true, false, false},
		{"nada", "", false, false, false, false},
	}
	for _, c := range cases {
		withConfigFakes(t, c.key, c.serverUp)
		st := CheckStatus(context.Background())
		if st.KeyPresent != c.wantKey || st.ServerUp != c.wantUp {
			t.Errorf("%s: Status = %+v, want KeyPresent=%v ServerUp=%v", c.name, st, c.wantKey, c.wantUp)
		}
		if st.Configured() != c.wantConfigOK {
			t.Errorf("%s: Configured() = %v, want %v", c.name, st.Configured(), c.wantConfigOK)
		}
	}
}

func TestGuidance(t *testing.T) {
	// Todo OK: un solo mensaje de confirmación.
	ok := Guidance(Status{KeyPresent: true, ServerUp: true})
	if len(ok) != 1 || !strings.Contains(ok[0], "configurado") {
		t.Errorf("Guidance(OK) = %v, want mensaje de configurado", ok)
	}

	// Server abajo: debe sugerir arrancar omniroute.
	down := strings.Join(Guidance(Status{KeyPresent: true, ServerUp: false}), " ")
	if !strings.Contains(down, "no responde") {
		t.Errorf("Guidance(server abajo) = %q, want paso de arranque", down)
	}

	// Key ausente: debe apuntar al Dashboard/Endpoints y al .env.
	noKey := strings.Join(Guidance(Status{KeyPresent: false, ServerUp: true}), " ")
	if !strings.Contains(noKey, "Endpoints") || !strings.Contains(noKey, ".env") {
		t.Errorf("Guidance(sin key) = %q, want pasos de generación de key", noKey)
	}

	// Ambos faltan: dos pasos.
	if got := Guidance(Status{}); len(got) != 2 {
		t.Errorf("Guidance(nada) devolvió %d pasos, want 2", len(got))
	}
}

// TestGuidanceNuncaExponeKey verifica la garantía de seguridad: la guía jamás incluye
// el valor de la API key, aunque esté presente.
func TestGuidanceNuncaExponeKey(t *testing.T) {
	const secret = "sk-supersecreto-12345"
	withConfigFakes(t, secret, true)
	st := CheckStatus(context.Background())
	for _, line := range Guidance(st) {
		if strings.Contains(line, secret) {
			t.Fatalf("Guidance filtró la API key en %q", line)
		}
	}
}

func TestListCombosReturnsCapabilityDefinitions(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/combos" {
			http.NotFound(w, r)
			return
		}
		if got := r.Header.Get("Authorization"); got != "Bearer test-key" {
			t.Fatalf("Authorization = %q", got)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"combos":[{"name":"gpt-agent","models":[{"model":"cx/gpt-5.6-sol"},{"model":"cx/gpt-5.6-sol-high"}]}]}`))
	}))
	defer server.Close()

	combos, err := ListCombos(context.Background(), server.URL, "test-key")
	if err != nil {
		t.Fatalf("ListCombos: %v", err)
	}
	if len(combos) != 1 || combos[0].Name != "gpt-agent" {
		t.Fatalf("combos = %+v", combos)
	}
	if got := strings.Join(combos[0].Models, ","); got != "cx/gpt-5.6-sol,cx/gpt-5.6-sol-high" {
		t.Fatalf("models = %q", got)
	}
}

func TestListCombosRejectsNonSuccessStatus(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		http.Error(w, "nope", http.StatusBadGateway)
	}))
	defer server.Close()

	if _, err := ListCombos(context.Background(), server.URL, "test-key"); err == nil || !strings.Contains(err.Error(), "status 502") {
		t.Fatalf("ListCombos error = %v", err)
	}
}
