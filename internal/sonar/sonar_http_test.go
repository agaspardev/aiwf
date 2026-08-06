package sonar

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestGateStatusReturnsProjectStatus(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Authorization") != "Bearer tok" {
			t.Errorf("falta el bearer token: %q", r.Header.Get("Authorization"))
		}
		_, _ = w.Write([]byte(`{"projectStatus":{"status":"OK"}}`))
	}))
	defer server.Close()

	cfg := Config{Host: server.URL, ProjectKey: "p", Token: "tok"}
	status, err := cfg.GateStatus(context.Background())
	if err != nil {
		t.Fatalf("GateStatus: %v", err)
	}
	if status != "OK" {
		t.Errorf("status=%q, esperaba OK", status)
	}
}

func TestGateStatusRejectsNonSuccess(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
	}))
	defer server.Close()

	cfg := Config{Host: server.URL, ProjectKey: "p"}
	if _, err := cfg.GateStatus(context.Background()); err == nil {
		t.Error("esperaba error con status 401")
	}
}

func TestIssuesParsesTotalAndList(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(`{"total":2,"issues":[
			{"severity":"BLOCKER","message":"m1","component":"c1","line":10},
			{"severity":"CRITICAL","message":"m2","component":"c2","line":20}]}`))
	}))
	defer server.Close()

	cfg := Config{Host: server.URL, ProjectKey: "p"}
	total, issues, err := cfg.Issues(context.Background())
	if err != nil {
		t.Fatalf("Issues: %v", err)
	}
	if total != 2 || len(issues) != 2 {
		t.Fatalf("total=%d len=%d, esperaba 2/2", total, len(issues))
	}
	if issues[0].Severity != "BLOCKER" || issues[1].Line != 20 {
		t.Errorf("issues mal parseados: %+v", issues)
	}
}

func TestIssuesRejectsNonSuccess(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer server.Close()

	cfg := Config{Host: server.URL, ProjectKey: "p"}
	if _, _, err := cfg.Issues(context.Background()); err == nil {
		t.Error("esperaba error con status 500")
	}
}

func TestReachableTrueOnOK(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/system/status" {
			t.Errorf("path inesperado: %s", r.URL.Path)
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	if !Reachable(context.Background(), server.URL) {
		t.Error("esperaba Reachable=true")
	}
}

func TestReachableFalseWhenDown(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusServiceUnavailable)
	}))
	server.Close() // servidor caído: la conexión falla

	if Reachable(context.Background(), server.URL) {
		t.Error("esperaba Reachable=false con servidor caído")
	}
}
