package omniroute

import (
	"context"
	"encoding/json"
	"fmt"
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

// Detailed extiende Status con info enriquecida desde APIs de OmniRoute.
type Detailed struct {
	Status
	MCPOnline       bool   `json:"mcp_online"`
	MCPTransport    string `json:"mcp_transport"`
	CompressionOn   bool   `json:"compression_on"`
	CompressionMode string `json:"compression_mode"`
	ProvidersActive int    `json:"providers_active"`
	ProviderCount   int    `json:"provider_count"`
	CacheHits       int    `json:"cache_hits"`
	CacheMisses     int    `json:"cache_misses"`
	CacheSize       int    `json:"cache_size"`
	TotalCalls24h   int    `json:"total_calls_24h"`
}

// Seams inyectables para test.
var (
	keyReader = ReadKey

	// pingServer devuelve true si el server local responde algo por HTTP.
	pingServer = func(ctx context.Context) bool {
		ctx, cancel := context.WithTimeout(ctx, 2*time.Second)
		defer cancel()
		req, err := http.NewRequestWithContext(ctx, http.MethodGet, baseURL+"/", nil)
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

	// httpGetJSON es un seam para test: hace GET y decodifica JSON en dst.
	httpGetJSON = func(ctx context.Context, url, apiKey string, dst any) error {
		req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
		if err != nil {
			return err
		}
		req.Header.Set("Authorization", "Bearer "+apiKey)
		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			return err
		}
		defer resp.Body.Close()
		if resp.StatusCode != http.StatusOK {
			return fmt.Errorf("http %d", resp.StatusCode)
		}
		return json.NewDecoder(resp.Body).Decode(dst)
	}
)

// CheckStatus calcula el estado de configuración sin escribir nada.
func CheckStatus(ctx context.Context) Status {
	return Status{
		KeyPresent: keyReader() != "",
		ServerUp:   pingServer(ctx),
	}
}

// CheckDetailedStatus devuelve el estado enriquecido consultando APIs de OmniRoute.
func CheckDetailedStatus(ctx context.Context) Detailed {
	d := Detailed{Status: CheckStatus(ctx)}
	if !d.Configured() {
		return d
	}
	apiKey := keyReader()

	// MCP status
	var mcp struct {
		Online    bool   `json:"online"`
		Transport string `json:"transport"`
		Activity  struct {
			TotalCalls24h int `json:"totalCalls24h"`
		} `json:"activity"`
	}
	if err := httpGetJSON(ctx, baseURL+"/api/mcp/status", apiKey, &mcp); err == nil {
		d.MCPOnline = mcp.Online
		d.MCPTransport = mcp.Transport
		d.TotalCalls24h = mcp.Activity.TotalCalls24h
	}

	// Compression status
	var comp struct {
		Enabled     bool   `json:"enabled"`
		DefaultMode string `json:"defaultMode"`
	}
	if err := httpGetJSON(ctx, baseURL+"/api/settings/compression", apiKey, &comp); err == nil {
		d.CompressionOn = comp.Enabled
		d.CompressionMode = comp.DefaultMode
	}

	// Providers
	var provs struct {
		Connections []any `json:"connections"`
	}
	if err := httpGetJSON(ctx, baseURL+"/api/providers", apiKey, &provs); err == nil {
		d.ProviderCount = len(provs.Connections)
		for _, c := range provs.Connections {
			if m, ok := c.(map[string]any); ok {
				if active, ok := m["isActive"].(bool); ok && active {
					d.ProvidersActive++
				}
			}
		}
	}

	// Cache stats
	var cache struct {
		Hits   int `json:"hits"`
		Misses int `json:"misses"`
		Size   int `json:"size"`
	}
	if err := httpGetJSON(ctx, baseURL+"/api/cache/stats", apiKey, &cache); err == nil {
		d.CacheHits = cache.Hits
		d.CacheMisses = cache.Misses
		d.CacheSize = cache.Size
	}

	return d
}

// RunDoctor delega en `omniroute doctor` (read-only), heredando stdout/stderr.
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
