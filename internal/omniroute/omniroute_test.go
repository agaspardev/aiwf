package omniroute

import (
	"context"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
)

// withKeyAndServer prepara un home temp con la API key y apunta baseURL al server.
func withKeyAndServer(t *testing.T, handler http.HandlerFunc) {
	t.Helper()
	home := t.TempDir()
	setHome(t, home)
	dir := filepath.Join(home, ".omniroute")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, ".env"), []byte("OMNIROUTE_API_KEY=k\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	server := httptest.NewServer(handler)
	old := baseURL
	baseURL = server.URL
	t.Cleanup(func() {
		baseURL = old
		server.Close()
	})
}

// setHome apunta el home del proceso a un temp dir (cross-platform: HOME en
// unix, USERPROFILE en windows).
func setHome(t *testing.T, dir string) {
	t.Helper()
	t.Setenv("HOME", dir)
	t.Setenv("USERPROFILE", dir)
}

func TestReadKeyFromEnvFile(t *testing.T) {
	home := t.TempDir()
	setHome(t, home)
	dir := filepath.Join(home, ".omniroute")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	body := "# comentario\nOMNIROUTE_API_KEY=\"secreto-123\"\nOTRA=x\n"
	if err := os.WriteFile(filepath.Join(dir, ".env"), []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
	if got := ReadKey(); got != "secreto-123" {
		t.Errorf("ReadKey() = %q, esperaba secreto-123 (sin comillas)", got)
	}
}

func TestReadKeyMissingFileReturnsEmpty(t *testing.T) {
	setHome(t, t.TempDir()) // sin ~/.omniroute/.env
	if got := ReadKey(); got != "" {
		t.Errorf("ReadKey() = %q, esperaba vacío", got)
	}
}

func TestConsultFailsWithoutKey(t *testing.T) {
	setHome(t, t.TempDir()) // sin key
	if _, err := Consult(context.Background(), "hola", "combo", 10); err == nil {
		t.Error("esperaba error cuando falta OMNIROUTE_API_KEY")
	}
}

func TestConsultReturnsJoinedText(t *testing.T) {
	withKeyAndServer(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("x-api-key") != "k" {
			t.Errorf("falta x-api-key: %q", r.Header.Get("x-api-key"))
		}
		_, _ = w.Write([]byte(`{"content":[{"type":"text","text":"linea1"},{"type":"text","text":"linea2"},{"type":"other","text":"ignorar"}]}`))
	})
	got, err := Consult(context.Background(), "hola", "combo", 100)
	if err != nil {
		t.Fatalf("Consult: %v", err)
	}
	if got != "linea1\nlinea2" {
		t.Errorf("Consult() = %q, esperaba linea1\nlinea2", got)
	}
}

func TestConsultRejectsNonSuccess(t *testing.T) {
	withKeyAndServer(t, func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusBadGateway)
	})
	if _, err := Consult(context.Background(), "hola", "combo", 100); err == nil {
		t.Error("esperaba error con status 502")
	}
}

func TestConsultFailsWhenNoTextContent(t *testing.T) {
	withKeyAndServer(t, func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(`{"content":[{"type":"image","text":""}]}`))
	})
	if _, err := Consult(context.Background(), "hola", "combo", 100); err == nil {
		t.Error("esperaba error cuando no hay contenido de texto")
	}
}
