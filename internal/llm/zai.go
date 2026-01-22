package llm

import (
	"ai-browser-agent/internal/agent/promts"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"ai-browser-agent/internal/config"
	"ai-browser-agent/internal/core"
)

type ZaiClient struct {
	apiKey      string
	baseURL     string
	model       string
	maxTokens   int
	temperature float32
	client      *http.Client
}

func NewZai(cfg *config.Config) Client {
	return &ZaiClient{
		apiKey:      cfg.Env.ZaiAPIKey,
		baseURL:     cfg.Env.ZaiBaseURL,
		model:       cfg.LLM.Model,
		maxTokens:   cfg.LLM.MaxTokens,
		temperature: cfg.LLM.Temperature,
		client:      &http.Client{},
	}
}

func (z *ZaiClient) NextAction(fullPrompt string) (*core.Action, error) {
	messages := []map[string]string{
		{
			"role":    "system",
			"content": promts.SystemPrompt,
		},
		{
			"role":    "user",
			"content": fullPrompt, // goal + snapshot + history
		},
	}

	reqBody := map[string]interface{}{
		"model":       z.model,
		"messages":    messages,
		"max_tokens":  z.maxTokens,
		"temperature": z.temperature,
		"response_format": map[string]string{
			"type": "json_object",
		},
	}

	bodyBytes, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("marshal request: %w", err)
	}

	req, err := http.NewRequest("POST", z.baseURL+"/chat/completions", bytes.NewReader(bodyBytes))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", "Bearer "+z.apiKey)
	req.Header.Set("Content-Type", "application/json")

	resp, err := z.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("http request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("ZAI API error %d: %s", resp.StatusCode, string(body))
	}

	var apiResp struct {
		Choices []struct {
			Message struct {
				Content string `json:"content"`
			} `json:"message"`
		} `json:"choices"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&apiResp); err != nil {
		return nil, fmt.Errorf("decode response: %w", err)
	}

	if len(apiResp.Choices) == 0 {
		return nil, fmt.Errorf("no choices in response")
	}

	rawJSON := apiResp.Choices[0].Message.Content

	var action core.Action
	if err = json.Unmarshal([]byte(rawJSON), &action); err != nil {
		return nil, fmt.Errorf("unmarshal action JSON (%s): %w", rawJSON, err)
	}

	if action.Type == "" {
		return nil, fmt.Errorf("empty action type")
	}

	return &action, nil
}
