package ai

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/solthoth/social-media-marketplace-assistant/services/api/internal/enrichment"
)

var (
	ErrOpenAIProviderNotConfigured = errors.New("openai provider is not configured")
	ErrOpenAIImageMissing          = errors.New("openai image input is missing")
)

const defaultOpenAIBaseURL = "https://api.openai.com/v1"

type OpenAIProviderConfig struct {
	APIKey     string
	Model      string
	BaseURL    string
	HTTPClient *http.Client
}

type OpenAIProvider struct {
	apiKey string
	model  string
	url    string
	client *http.Client
}

func NewOpenAIProvider(config OpenAIProviderConfig) OpenAIProvider {
	baseURL := strings.TrimRight(config.BaseURL, "/")
	if baseURL == "" {
		baseURL = defaultOpenAIBaseURL
	}
	client := config.HTTPClient
	if client == nil {
		client = &http.Client{Timeout: 60 * time.Second}
	}
	return OpenAIProvider{
		apiKey: strings.TrimSpace(config.APIKey),
		model:  strings.TrimSpace(config.Model),
		url:    baseURL + "/responses",
		client: client,
	}
}

func (p OpenAIProvider) GenerateItemDetails(ctx context.Context, input enrichment.ItemDetailInput) (enrichment.ItemDetailSuggestion, error) {
	if p.apiKey == "" || p.model == "" {
		return enrichment.ItemDetailSuggestion{}, ErrOpenAIProviderNotConfigured
	}
	payload, err := newOpenAIResponsesRequest(p.model, input)
	if err != nil {
		return enrichment.ItemDetailSuggestion{}, err
	}
	body, err := json.Marshal(payload)
	if err != nil {
		return enrichment.ItemDetailSuggestion{}, err
	}
	request, err := http.NewRequestWithContext(ctx, http.MethodPost, p.url, bytes.NewReader(body))
	if err != nil {
		return enrichment.ItemDetailSuggestion{}, err
	}
	request.Header.Set("Authorization", "Bearer "+p.apiKey)
	request.Header.Set("Content-Type", "application/json")

	response, err := p.client.Do(request)
	if err != nil {
		return enrichment.ItemDetailSuggestion{}, err
	}
	defer response.Body.Close()
	if response.StatusCode < 200 || response.StatusCode >= 300 {
		errorBody, _ := io.ReadAll(io.LimitReader(response.Body, 4096))
		return enrichment.ItemDetailSuggestion{}, fmt.Errorf("openai response failed: status %d: %s", response.StatusCode, strings.TrimSpace(string(errorBody)))
	}

	var parsed openAIResponsesResponse
	if err := json.NewDecoder(response.Body).Decode(&parsed); err != nil {
		return enrichment.ItemDetailSuggestion{}, err
	}
	outputText := parsed.OutputText()
	if strings.TrimSpace(outputText) == "" {
		return enrichment.ItemDetailSuggestion{}, errors.New("openai response did not include output text")
	}
	var suggestion enrichment.ItemDetailSuggestion
	if err := json.Unmarshal([]byte(outputText), &suggestion); err != nil {
		return enrichment.ItemDetailSuggestion{}, err
	}
	return suggestion, nil
}

func newOpenAIResponsesRequest(model string, input enrichment.ItemDetailInput) (map[string]any, error) {
	content := []map[string]any{
		{
			"type": "input_text",
			"text": itemDetailPrompt(input),
		},
	}
	for _, photo := range input.Photos {
		if strings.TrimSpace(photo.DataURL) == "" {
			return nil, ErrOpenAIImageMissing
		}
		content = append(content, map[string]any{
			"type":      "input_image",
			"image_url": photo.DataURL,
			"detail":    "low",
		})
	}

	return map[string]any{
		"model":        model,
		"store":        false,
		"instructions": "You help a private resale seller draft inventory details from a title and item photos. Return only structured JSON. Leave uncertain fields empty instead of guessing.",
		"input": []map[string]any{
			{
				"role":    "user",
				"content": content,
			},
		},
		"text": map[string]any{
			"format": map[string]any{
				"type":        "json_schema",
				"name":        "item_detail_suggestion",
				"description": "Suggested non-pricing item details for a resale inventory item.",
				"strict":      true,
				"schema": map[string]any{
					"type":                 "object",
					"additionalProperties": false,
					"required":             []string{"description", "category", "size", "condition", "notes"},
					"properties": map[string]any{
						"description": map[string]any{"type": "string"},
						"category":    map[string]any{"type": "string"},
						"size":        map[string]any{"type": "string"},
						"condition":   map[string]any{"type": "string"},
						"notes":       map[string]any{"type": "string"},
					},
				},
			},
		},
	}, nil
}

func itemDetailPrompt(input enrichment.ItemDetailInput) string {
	var builder strings.Builder
	builder.WriteString("Draft missing non-pricing details for this resale inventory item.\n")
	builder.WriteString("Title: " + input.Title + "\n")
	builder.WriteString("Existing description: " + input.ExistingDescription + "\n")
	builder.WriteString("Existing category: " + input.ExistingCategory + "\n")
	builder.WriteString("Existing size: " + input.ExistingSize + "\n")
	builder.WriteString("Existing condition: " + input.ExistingCondition + "\n")
	builder.WriteString("Existing notes: " + input.ExistingNotes + "\n")
	builder.WriteString("Do not suggest prices. Use concise seller-friendly language.")
	return builder.String()
}

type openAIResponsesResponse struct {
	Output     []openAIOutputItem `json:"output"`
	TextOutput string             `json:"output_text"`
}

func (r openAIResponsesResponse) OutputText() string {
	if strings.TrimSpace(r.TextOutput) != "" {
		return r.TextOutput
	}
	for _, output := range r.Output {
		for _, content := range output.Content {
			if content.Type == "output_text" && strings.TrimSpace(content.Text) != "" {
				return content.Text
			}
		}
	}
	return ""
}

type openAIOutputItem struct {
	Type    string                `json:"type"`
	Content []openAIOutputContent `json:"content"`
}

type openAIOutputContent struct {
	Type string `json:"type"`
	Text string `json:"text"`
}
