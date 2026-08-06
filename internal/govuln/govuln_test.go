package govuln

import "testing"

// Stream NDJSON representativo de `govulncheck -json`: definiciones osv +
// findings que las referencian (varios findings pueden apuntar al mismo osv).
const govulnStream = `
{"config": {"protocol_version": "v1.0.0"}}
{"progress": {"message": "Scanning..."}}
{"osv": {"id": "GO-2024-2687", "summary": "HTTP/2 flood in net/http", "affected": []}}
{"finding": {"osv": "GO-2024-2687", "fixed_version": "1.22.0", "trace": [{"module": "stdlib", "package": "net/http", "function": "Serve"}]}}
{"finding": {"osv": "GO-2024-2687", "fixed_version": "1.22.0", "trace": [{"module": "stdlib", "package": "net/http", "function": "ServeTLS"}]}}
{"osv": {"id": "GO-2023-0001", "summary": "unused vuln (no finding)", "affected": []}}
`

func TestParseGovulncheckDedupesByOSV(t *testing.T) {
	fs, err := ParseGovulncheck([]byte(govulnStream))
	if err != nil {
		t.Fatalf("ParseGovulncheck: %v", err)
	}
	// Dos findings referencian el mismo osv -> un Finding. El osv sin finding se ignora.
	if len(fs) != 1 {
		t.Fatalf("esperaba 1 finding deduplicado, got %d: %+v", len(fs), fs)
	}
	f := fs[0]
	if f.Tool != "govulncheck" || f.RuleID != "GO-2024-2687" || f.Category != "vuln" {
		t.Fatalf("finding mal mapeado: %+v", f)
	}
	if f.Message == "" || f.File == "" {
		t.Errorf("esperaba message y file (módulo/paquete) poblados: %+v", f)
	}
}

func TestParseGovulncheckNoFindings(t *testing.T) {
	fs, err := ParseGovulncheck([]byte(`{"config":{}}` + "\n" + `{"progress":{"message":"done"}}`))
	if err != nil {
		t.Fatalf("ParseGovulncheck: %v", err)
	}
	if len(fs) != 0 {
		t.Fatalf("sin findings esperaba 0, got %d", len(fs))
	}
}

func TestParseGovulncheckTruncatedErrors(t *testing.T) {
	if _, err := ParseGovulncheck([]byte(`{"osv": {"id"`)); err == nil {
		t.Error("stream truncado debería dar error")
	}
}
