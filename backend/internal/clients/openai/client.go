package openai

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strings"
)

type Client struct {
	apiKey  string
	baseURL string
	model   string
	hc      *http.Client
}

func NewClient(apiKey, baseURL, model string, hc *http.Client) *Client {
	return &Client{apiKey: apiKey, baseURL: baseURL, model: model, hc: hc}
}

func (c *Client) Generate(ctx context.Context, prompt string) (string, error) {
	if c.apiKey == "" {
		return "", errors.New("OPENAI_API_KEY is not set")
	}

	body := OpenAIResponsesRequest{
		Model: c.model,
		Input: prompt,
	}

	b, _ := json.Marshal(body)

	req, _ := http.NewRequestWithContext(ctx, "POST", c.baseURL+"/responses", bytes.NewBuffer(b))
	req.Header.Set("Authorization", "Bearer "+c.apiKey)
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.hc.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	raw, _ := io.ReadAll(resp.Body)
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return "", errors.New("openai api error: " + resp.Status + " body: " + string(raw))
	}

	var out OpenAIResponsesResponse
	if err := json.Unmarshal(raw, &out); err != nil {
		return "", err
	}

	text := strings.TrimSpace(extractText(out))
	if text == "" {
		return "", errors.New("empty text from openai")
	}

	return text, nil
}
