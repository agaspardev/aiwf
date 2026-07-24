package omniroute

import (
	"context"
	"net/http"
	"time"
)

// Status es el estado de configuración de omniroute (chequeo no destructivo).
type Status struct {
	KeyPresent bool // hay OMNIROUTE_API_KEY en ~/.omniroute/.env
	ServerUp   bool // el server local responde HTTP
}

// Configured indica si omniroute está listo para usarse.
func (s Status) Configured() bool { return s.KeyPresent && s.ServerUp }

// Seams inyectables para test.
var (
	keyReader = ReadKey

	// pingServer devuelve true si el server local responde algo por HTTP. Cualquier
	// respuesta (incluida 401/404) prueba que el proceso está vivo; solo un error de
	// conexión cuenta como "abajo". Timeout corto para no colgar la CLI ni molestar
	// al server del usuario.
	pingServer = func(ctx context.Context) bool {
		ctx, cancel := context.WithTimeout(ctx, 2*time.Second)
		defer cancel()
		req, err := http.NewRequestWithContext(ctx, http.MethodGet, DefaultURL+"/", nil)
		if err != nil {
			return false
		}
		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			return false
		}
		resp.Body.Close()
		return true
	}
)

// CheckStatus calcula el estado de configuración sin escribir nada (solo lee la key y
// hace un ping al server).
func CheckStatus(ctx context.Context) Status {
	return Status{
		KeyPresent: keyReader() != "",
		ServerUp:   pingServer(ctx),
	}
}

// RunDoctor delega en `omniroute doctor` (read-only), heredando stdout/stderr. Reusa
// el seam runCommand de install.go.
func RunDoctor(ctx context.Context) error {
	return runCommand(ctx, binaryName, "doctor")
}

// Guidance arma los pasos accionables según qué falta. Nunca incluye la key en el
// texto: aiwf guía, no manipula credenciales.
func Guidance(s Status) []string {
	if s.Configured() {
		return []string{"omniroute configurado (server arriba y API key presente)."}
	}
	var steps []string
	if !s.ServerUp {
		steps = append(steps,
			"Server no responde en "+DefaultURL+": arrancá omniroute (ej. `omniroute` o `omniroute setup` en la primera vez).")
	}
	if !s.KeyPresent {
		steps = append(steps,
			"Falta la API key: generala en el Dashboard ("+DefaultURL+") → Endpoints y colocala como OMNIROUTE_API_KEY en ~/.omniroute/.env.")
	}
	return steps
}
