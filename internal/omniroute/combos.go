package omniroute

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
)

// ComboDefinition contiene la identidad y los aliases efectivos de un combo.
type ComboDefinition struct {
	Name   string
	Models []string
}

type comboListResponse struct {
	Combos []struct {
		Name   string `json:"name"`
		Models []struct {
			Model string `json:"model"`
		} `json:"models"`
	} `json:"combos"`
}

// ListCombos consulta el estado efectivo de OmniRoute sin exponer credenciales.
func ListCombos(ctx context.Context, baseURL, apiKey string) ([]ComboDefinition, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, baseURL+"/api/combos", nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+apiKey)
	req.Header.Set("x-api-key", apiKey)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("omniroute combos: status %d", resp.StatusCode)
	}

	var payload comboListResponse
	if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
		return nil, fmt.Errorf("decodificando combos de omniroute: %w", err)
	}
	combos := make([]ComboDefinition, 0, len(payload.Combos))
	for _, combo := range payload.Combos {
		models := make([]string, 0, len(combo.Models))
		for _, model := range combo.Models {
			models = append(models, model.Model)
		}
		combos = append(combos, ComboDefinition{Name: combo.Name, Models: models})
	}
	return combos, nil
}
