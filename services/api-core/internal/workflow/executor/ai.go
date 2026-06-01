package executor

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"
	"time"
)

type aiConfig struct {
	Prompt string `json:"prompt"`
	Model  string `json:"model"`
}

type GeminiResponse struct {
	Candidates []struct {
		Content struct {
			Parts []struct {
				Text string `json:"text"`
			} `json:"parts"`
		} `json:"content"`
	} `json:"candidates"`
}

func (e *Executor) executeAI(config []byte, input []byte) ([]byte, error) {

	var cfg aiConfig

	if err := json.Unmarshal(config, &cfg); err != nil {
		return nil, err
	}

	if cfg.Prompt == "" {
		return nil, errors.New("prompt is required")
	}

	prompt := interpolateString(cfg.Prompt, input)

	model := os.Getenv("GEMINI_MODEL")
	if model == "" {
		model = "gemini-2.5-flash"
	}

	apiKey := os.Getenv("GEMINI_API_KEY")
	if apiKey == "" {
		return nil, errors.New("GEMINI_API_KEY not configured")
	}

	reqBody := map[string]interface{}{
		"contents": []map[string]interface{}{
			{
				"parts": []map[string]interface{}{
					{
						"text": prompt,
					},
				},
			},
		},
	}

	bodyBytes, _ := json.Marshal(reqBody)

	url := fmt.Sprintf(
		"https://generativelanguage.googleapis.com/v1beta/models/%s:generateContent?key=%s",
		model,
		apiKey,
	)

	req, err := http.NewRequest(
		http.MethodPost,
		url,
		bytes.NewReader(bodyBytes),
	)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{
		Timeout: 60 * time.Second,
	}

	res, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	var response GeminiResponse

	if err := json.NewDecoder(res.Body).Decode(&response); err != nil {
		return nil, err
	}

	if len(response.Candidates) == 0 {
		return nil, errors.New("no response from gemini")
	}

	text := response.Candidates[0].Content.Parts[0].Text

	output := map[string]interface{}{
		"text": text,
	}

	return json.Marshal(output)
}