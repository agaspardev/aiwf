package omniroute

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestListProviders(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/providers" {
			http.NotFound(w, r)
			return
		}
		if got := r.Header.Get("Authorization"); got != "Bearer test-key" {
			t.Fatalf("Authorization = %q", got)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"connections":[
			{"id":"c1","provider":"antigravity","name":"main","isActive":true,"testStatus":"active","backoffLevel":0},
			{"id":"c2","provider":"groq","name":"free","isActive":false,"testStatus":"active","backoffLevel":2,"lastError":"rate limited"}
		]}`))
	}))
	defer server.Close()

	providers, err := ListProviders(context.Background(), server.URL, "test-key")
	if err != nil {
		t.Fatalf("ListProviders: %v", err)
	}
	if len(providers) != 2 {
		t.Fatalf("got %d providers, want 2", len(providers))
	}
	if providers[0].Provider != "antigravity" || !providers[0].Active {
		t.Errorf("provider[0] = %+v", providers[0])
	}
	if providers[1].Provider != "groq" || providers[1].Active {
		t.Errorf("provider[1] = %+v", providers[1])
	}
	if providers[1].LastError != "rate limited" {
		t.Errorf("provider[1].LastError = %q", providers[1].LastError)
	}
}

func TestListProvidersRejectsNonSuccessStatus(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		http.Error(w, "nope", http.StatusBadGateway)
	}))
	defer server.Close()

	_, err := ListProviders(context.Background(), server.URL, "key")
	if err == nil || !strings.Contains(err.Error(), "http 502") {
		t.Fatalf("ListProviders error = %v", err)
	}
}

func TestSummarizeProviders(t *testing.T) {
	providers := []ProviderInfo{
		{Provider: "antigravity", Active: true},
		{Provider: "antigravity", Active: false},
		{Provider: "groq", Active: true},
		{Provider: "groq", Active: true},
		{Provider: "openrouter", Active: false},
	}
	summary := SummarizeProviders(providers)
	if len(summary) != 3 {
		t.Fatalf("SummarizeProviders returned %d groups, want 3", len(summary))
	}
	for _, s := range summary {
		switch s.Type {
		case "antigravity":
			if s.Active != 1 || s.Inactive != 1 {
				t.Errorf("antigravity: active=%d inactive=%d", s.Active, s.Inactive)
			}
		case "groq":
			if s.Active != 2 || s.Inactive != 0 {
				t.Errorf("groq: active=%d inactive=%d", s.Active, s.Inactive)
			}
		case "openrouter":
			if s.Active != 0 || s.Inactive != 1 {
				t.Errorf("openrouter: active=%d inactive=%d", s.Active, s.Inactive)
			}
		}
	}
}

func TestPrintProvidersTable(t *testing.T) {
	providers := []ProviderInfo{
		{Provider: "antigravity", Name: "main", Active: true, TestStatus: "active"},
		{Provider: "groq", Name: "free", Active: false, TestStatus: "error", LastError: "quota"},
	}
	table := PrintProvidersTable(providers)
	if !strings.Contains(table, "antigravity") || !strings.Contains(table, "groq") {
		t.Errorf("table missing provider names: %s", table)
	}
}
