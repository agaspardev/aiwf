package omniroute

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"
)

// UsageReport contiene el resumen de uso de OmniRoute.
type UsageReport struct {
	TotalRequests int
	TotalTokens   int64
	TotalCostUSD  float64
	Period        string
}

// GetUsage consulta el reporte de uso de OmniRoute.
// El endpoint /api/usage/logs devuelve líneas pipe-delimitadas:
//
//	timestamp | model | provider | account | tokens | durationMs | status
func GetUsage(ctx context.Context, baseURL, apiKey string) (*UsageReport, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, baseURL+"/api/usage/logs?limit=5000", nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+apiKey)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("getting usage: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("usage endpoint: http %d", resp.StatusCode)
	}

	raw, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("reading usage: %w", err)
	}

	report := &UsageReport{Period: "last 5000 requests"}
	lines := strings.Split(strings.TrimSpace(string(raw)), "\n")
	for _, line := range lines {
		parts := strings.Split(line, " | ")
		if len(parts) < 7 {
			continue
		}
		report.TotalRequests++
		if tokens, err := strconv.ParseInt(strings.TrimSpace(parts[4]), 10, 64); err == nil {
			report.TotalTokens += tokens
		}
	}
	return report, nil
}

// PrintUsage imprime el reporte de uso en formato CLI.
func PrintUsage(r *UsageReport, since time.Time) string {
	var b strings.Builder
	b.WriteString(fmt.Sprintf("Period:          %s\n", r.Period))
	b.WriteString(fmt.Sprintf("Since:           %s\n", since.Format("2006-01-02")))
	b.WriteString(fmt.Sprintf("Total requests:  %d\n", r.TotalRequests))
	b.WriteString(fmt.Sprintf("Total tokens:    %d\n", r.TotalTokens))
	if r.TotalRequests > 0 {
		avgTokens := r.TotalTokens / int64(r.TotalRequests)
		b.WriteString(fmt.Sprintf("Avg tokens/req:  %d\n", avgTokens))
	}
	return b.String()
}
