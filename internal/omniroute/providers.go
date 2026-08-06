package omniroute

import (
	"context"
	"fmt"
	"strings"
)

// ProviderInfo contiene el resumen de una conexión a provider.
type ProviderInfo struct {
	ID         string `json:"id"`
	Provider   string `json:"provider"`
	Name       string `json:"name"`
	Active     bool   `json:"active"`
	TestStatus string `json:"test_status"`
	Backoff    int    `json:"backoff_level"`
	LastError  string `json:"last_error,omitempty"`
}

type providerListResponse struct {
	Connections []struct {
		ID         string `json:"id"`
		Provider   string `json:"provider"`
		Name       string `json:"name"`
		IsActive   bool   `json:"isActive"`
		TestStatus string `json:"testStatus"`
		Backoff    int    `json:"backoffLevel"`
		LastError  string `json:"lastError,omitempty"`
	} `json:"connections"`
}

// ListProviders consulta los providers configurados en OmniRoute.
func ListProviders(ctx context.Context, baseURL, apiKey string) ([]ProviderInfo, error) {
	var payload providerListResponse
	if err := httpGetJSON(ctx, baseURL+"/api/providers", apiKey, &payload); err != nil {
		return nil, fmt.Errorf("listing providers: %w", err)
	}
	out := make([]ProviderInfo, 0, len(payload.Connections))
	for _, c := range payload.Connections {
		out = append(out, ProviderInfo{
			ID:         c.ID,
			Provider:   c.Provider,
			Name:       c.Name,
			Active:     c.IsActive,
			TestStatus: c.TestStatus,
			Backoff:    c.Backoff,
			LastError:  c.LastError,
		})
	}
	return out, nil
}

// ProviderSummary agrupa providers por tipo.
type ProviderSummary struct {
	Type            string
	Active          int
	Inactive        int
	ConnectionCount int
}

// SummarizeProviders agrupa los providers por tipo y cuenta estados.
func SummarizeProviders(providers []ProviderInfo) []ProviderSummary {
	byType := make(map[string]*ProviderSummary)
	var keys []string
	for _, p := range providers {
		ps, ok := byType[p.Provider]
		if !ok {
			ps = &ProviderSummary{Type: p.Provider}
			keys = append(keys, p.Provider)
		}
		ps.ConnectionCount++
		if p.Active {
			ps.Active++
		} else {
			ps.Inactive++
		}
		byType[p.Provider] = ps
	}
	out := make([]ProviderSummary, 0, len(keys))
	for _, k := range keys {
		out = append(out, *byType[k])
	}
	return out
}

// PrintProvidersTable imprime los providers en formato tabla CLI.
func PrintProvidersTable(providers []ProviderInfo) string {
	var b strings.Builder
	b.WriteString(fmt.Sprintf("%-20s %-22s %-8s %-10s %s\n", "Provider", "Name", "Active", "Status", "Backoff"))
	b.WriteString(strings.Repeat("-", 80) + "\n")
	for _, p := range providers {
		active := "✓"
		if !p.Active {
			active = "✗"
		}
		status := p.TestStatus
		if p.LastError != "" {
			status = "error"
		}
		b.WriteString(fmt.Sprintf("%-20s %-22s %-8s %-10s %d\n",
			p.Provider, truncate(p.Name, 20), active, status, p.Backoff))
	}
	return b.String()
}

func truncate(s string, n int) string {
	if n <= 0 {
		return ""
	}
	runes := []rune(s)
	if len(runes) <= n {
		return s
	}
	return string(runes[:n-1]) + "…"
}
