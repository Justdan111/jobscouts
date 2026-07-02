package llm

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"time"
)

// Client is a thin wrapper over the Anthropic Messages API. If no API key is
// configured it reports Enabled()==false and callers fall back gracefully.
type Client struct {
	key   string
	model string
	http  *http.Client
}

func New(key, model string) *Client {
	return &Client{key: key, model: model, http: &http.Client{Timeout: 60 * time.Second}}
}

func (c *Client) Enabled() bool { return c.key != "" }

// Complete sends a single user message and returns the concatenated text.
func (c *Client) Complete(ctx context.Context, prompt string, maxTokens int) (string, error) {
	if !c.Enabled() {
		return "", errors.New("anthropic key not configured")
	}
	body, _ := json.Marshal(map[string]any{
		"model":      c.model,
		"max_tokens": maxTokens,
		"messages":   []map[string]string{{"role": "user", "content": prompt}},
	})
	req, _ := http.NewRequestWithContext(ctx, http.MethodPost,
		"https://api.anthropic.com/v1/messages", bytes.NewReader(body))
	req.Header.Set("content-type", "application/json")
	req.Header.Set("x-api-key", c.key)
	req.Header.Set("anthropic-version", "2023-06-01")

	res, err := c.http.Do(req)
	if err != nil {
		return "", err
	}
	defer res.Body.Close()
	if res.StatusCode != 200 {
		return "", fmt.Errorf("anthropic status %d", res.StatusCode)
	}
	var out struct {
		Content []struct {
			Type, Text string
		} `json:"content"`
	}
	if err := json.NewDecoder(res.Body).Decode(&out); err != nil {
		return "", err
	}
	var sb bytes.Buffer
	for _, b := range out.Content {
		if b.Type == "text" {
			sb.WriteString(b.Text)
		}
	}
	return sb.String(), nil
}
