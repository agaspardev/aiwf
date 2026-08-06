package omniroute

import (
	"context"
	"os"
	"path/filepath"
	"testing"
)

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
