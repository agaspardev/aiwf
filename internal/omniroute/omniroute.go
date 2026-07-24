// Package omniroute habla con el proxy local de OmniRoute (API Anthropic-compatible):
// lectura de la API key y consultas de solo texto (usado por `aiwf gemini`), y provee
// el ciclo de vida de instalación (Detect/Decide/InstallCommand/Ensure — ver install.go).
package omniroute

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

// DefaultURL es el endpoint local por defecto de OmniRoute.
const DefaultURL = "http://127.0.0.1:20128"

// ReadKey devuelve OMNIROUTE_API_KEY desde ~/.omniroute/.env, o "" si no está.
func ReadKey() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return ""
	}
	f, err := os.Open(filepath.Join(home, ".omniroute", ".env"))
	if err != nil {
		return ""
	}
	defer f.Close()
	sc := bufio.NewScanner(f)
	for sc.Scan() {
		if v, ok := strings.CutPrefix(strings.TrimSpace(sc.Text()), "OMNIROUTE_API_KEY="); ok {
			return strings.Trim(strings.TrimSpace(v), `"'`)
		}
	}
	return ""
}

type message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type request struct {
	Model     string    `json:"model"`
	MaxTokens int       `json:"max_tokens"`
	Messages  []message `json:"messages"`
}

type response struct {
	Content []struct {
		Type string `json:"type"`
		Text string `json:"text"`
	} `json:"content"`
}

// Consult envía un prompt de solo texto al combo indicado y devuelve el texto de la
// respuesta. No expone claves ni habilita herramientas locales.
func Consult(ctx context.Context, prompt, combo string, maxTokens int) (string, error) {
	key := ReadKey()
	if key == "" {
		return "", fmt.Errorf("no se encontró OMNIROUTE_API_KEY en ~/.omniroute/.env")
	}
	body, err := json.Marshal(request{
		Model:     combo,
		MaxTokens: maxTokens,
		Messages:  []message{{Role: "user", Content: prompt}},
	})
	if err != nil {
		return "", err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, DefaultURL+"/v1/messages", bytes.NewReader(body))
	if err != nil {
		return "", err
	}
	req.Header.Set("Authorization", "Bearer "+key)
	req.Header.Set("x-api-key", key)
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("omniroute: status %d", resp.StatusCode)
	}
	var out response
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		return "", err
	}
	var texts []string
	for _, c := range out.Content {
		if c.Type == "text" && c.Text != "" {
			texts = append(texts, c.Text)
		}
	}
	if len(texts) == 0 {
		return "", fmt.Errorf("OmniRoute no devolvió contenido de texto")
	}
	return strings.Join(texts, "\n"), nil
}
