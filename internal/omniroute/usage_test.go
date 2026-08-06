package omniroute

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestGetUsageParsesPipeDelimitedLines(t *testing.T) {
	body := strings.Join([]string{
		"2026-08-06T10:00:00 | gpt | openai | acc | 120 | 800 | ok",
		"2026-08-06T10:01:00 | gemini | google | acc | 80 | 300 | ok",
		"linea malformada sin suficientes campos",
		"2026-08-06T10:02:00 | groq | groq | acc | notanumber | 100 | ok",
	}, "\n")
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(body))
	}))
	defer server.Close()

	r, err := GetUsage(context.Background(), server.URL, "k")
	if err != nil {
		t.Fatalf("GetUsage: %v", err)
	}
	// 3 líneas con >=7 campos cuentan como request; la malformada se saltea.
	if r.TotalRequests != 3 {
		t.Errorf("TotalRequests=%d, esperaba 3", r.TotalRequests)
	}
	// tokens sumables: 120 + 80 (el "notanumber" no suma pero la línea sí cuenta).
	if r.TotalTokens != 200 {
		t.Errorf("TotalTokens=%d, esperaba 200", r.TotalTokens)
	}
}

func TestGetUsageRejectsNonSuccessStatus(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer server.Close()

	if _, err := GetUsage(context.Background(), server.URL, "k"); err == nil {
		t.Error("esperaba error con status 500")
	}
}

func TestPrintUsageIncluyeAvgSoloConRequests(t *testing.T) {
	since := time.Date(2026, 8, 1, 0, 0, 0, 0, time.UTC)

	withReqs := PrintUsage(&UsageReport{Period: "p", TotalRequests: 4, TotalTokens: 400}, since)
	if !strings.Contains(withReqs, "Avg tokens/req:  100") {
		t.Errorf("esperaba avg 100 en:\n%s", withReqs)
	}

	zero := PrintUsage(&UsageReport{Period: "p", TotalRequests: 0}, since)
	if strings.Contains(zero, "Avg tokens/req") {
		t.Errorf("no esperaba avg con 0 requests:\n%s", zero)
	}
	if !strings.Contains(zero, "Since:           2026-08-01") {
		t.Errorf("esperaba fecha since formateada en:\n%s", zero)
	}
}

func TestActionString(t *testing.T) {
	cases := map[Action]string{Install: "install", Skip: "skip", Action(99): "unknown"}
	for a, want := range cases {
		if got := a.String(); got != want {
			t.Errorf("Action(%d).String()=%q, esperaba %q", a, got, want)
		}
	}
}
