// Package govuln adapta la salida de `govulncheck -json` (stream NDJSON de
// mensajes osv/finding) al modelo Finding del agregador. Determinista, sin red.
package govuln

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"

	"github.com/agaspardev/aiwf/internal/report"
)

// govulnMessage es un mensaje del stream: solo osv o finding nos interesan.
type govulnMessage struct {
	OSV *struct {
		ID      string `json:"id"`
		Summary string `json:"summary"`
	} `json:"osv"`
	Finding *struct {
		OSV   string `json:"osv"`
		Trace []struct {
			Module  string `json:"module"`
			Package string `json:"package"`
		} `json:"trace"`
	} `json:"finding"`
}

// ParseGovulncheck decodifica el stream y produce un Finding por vulnerabilidad
// EFECTIVAMENTE encontrada (dedup por id de osv). Los osv sin finding asociado se
// ignoran (definición sin traza no es un hallazgo).
func ParseGovulncheck(data []byte) ([]report.Finding, error) {
	dec := json.NewDecoder(bytes.NewReader(data))
	summaries := map[string]string{}
	locations := map[string]string{}
	order := []string{}
	seen := map[string]bool{}

	for {
		var msg govulnMessage
		if err := dec.Decode(&msg); err != nil {
			if err == io.EOF {
				break
			}
			return nil, fmt.Errorf("govulncheck stream inválido: %w", err)
		}
		switch {
		case msg.OSV != nil:
			summaries[msg.OSV.ID] = msg.OSV.Summary
		case msg.Finding != nil:
			id := msg.Finding.OSV
			if id == "" {
				continue
			}
			if !seen[id] {
				seen[id] = true
				order = append(order, id)
			}
			// primera traza con módulo/paquete gana como ubicación
			if locations[id] == "" && len(msg.Finding.Trace) > 0 {
				t := msg.Finding.Trace[0]
				loc := t.Module
				if t.Package != "" {
					loc = t.Package
				}
				locations[id] = loc
			}
		}
	}

	findings := make([]report.Finding, 0, len(order))
	for _, id := range order {
		findings = append(findings, report.Finding{
			Tool:     "govulncheck",
			RuleID:   id,
			Severity: report.SeverityHigh,
			Message:  summaries[id],
			File:     locations[id],
			Category: "vuln",
		})
	}
	return findings, nil
}
