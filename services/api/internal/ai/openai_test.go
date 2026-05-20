package ai

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"testing"

	"github.com/solthoth/social-media-marketplace-assistant/services/api/internal/enrichment"
	"github.com/stretchr/testify/suite"
)

type OpenAIProviderSuite struct {
	suite.Suite
}

func TestOpenAIProviderSuite(t *testing.T) {
	suite.Run(t, new(OpenAIProviderSuite))
}

func (s *OpenAIProviderSuite) TestGenerateItemDetailsSendsVisionRequestAndParsesSuggestion() {
	var requestBody map[string]any
	client := &http.Client{Transport: roundTripFunc(func(r *http.Request) (*http.Response, error) {
		s.Equal("/v1/responses", r.URL.Path)
		s.Equal("Bearer test-key", r.Header.Get("Authorization"))
		s.Require().NoError(json.NewDecoder(r.Body).Decode(&requestBody))
		return jsonResponse(http.StatusOK, `{
			"id": "resp_123",
			"output": [{
				"type": "message",
				"content": [{
					"type": "output_text",
					"text": "{\"description\":\"Faded denim jacket\",\"category\":\"Clothing\",\"size\":\"M\",\"condition\":\"Good\",\"notes\":\"Confirm measurements.\"}"
				}]
			}]
		}`), nil
	})}

	provider := NewOpenAIProvider(OpenAIProviderConfig{
		APIKey:     "test-key",
		Model:      "gpt-4.1-mini",
		BaseURL:    "https://api.test/v1",
		HTTPClient: client,
	})

	suggestion, err := provider.GenerateItemDetails(context.Background(), enrichment.ItemDetailInput{
		Title: "Denim jacket",
		Photos: []enrichment.ItemPhotoInput{
			{ID: "photo-1", Filename: "front.png", MimeType: "image/png", DataURL: "data:image/png;base64,abc123"},
		},
	})

	s.Require().NoError(err)
	s.Equal("Faded denim jacket", suggestion.Description)
	s.Equal("Clothing", suggestion.Category)

	s.Equal("gpt-4.1-mini", requestBody["model"])
	s.Equal(false, requestBody["store"])
	input := requestBody["input"].([]any)
	message := input[0].(map[string]any)
	content := message["content"].([]any)
	image := content[1].(map[string]any)
	s.Equal("input_image", image["type"])
	s.Equal("data:image/png;base64,abc123", image["image_url"])
	text := requestBody["text"].(map[string]any)
	format := text["format"].(map[string]any)
	s.Equal("json_schema", format["type"])
}

func (s *OpenAIProviderSuite) TestGenerateItemDetailsRequiresAPIKeyAndPhotoData() {
	provider := NewOpenAIProvider(OpenAIProviderConfig{Model: "gpt-4.1-mini"})

	_, err := provider.GenerateItemDetails(context.Background(), enrichment.ItemDetailInput{
		Title:  "Denim jacket",
		Photos: []enrichment.ItemPhotoInput{{ID: "photo-1"}},
	})

	s.ErrorIs(err, ErrOpenAIProviderNotConfigured)

	provider = NewOpenAIProvider(OpenAIProviderConfig{APIKey: "test-key", Model: "gpt-4.1-mini"})
	_, err = provider.GenerateItemDetails(context.Background(), enrichment.ItemDetailInput{
		Title:  "Denim jacket",
		Photos: []enrichment.ItemPhotoInput{{ID: "photo-1"}},
	})

	s.ErrorIs(err, ErrOpenAIImageMissing)
}

func (s *OpenAIProviderSuite) TestGenerateItemDetailsReturnsProviderErrors() {
	client := &http.Client{Transport: roundTripFunc(func(r *http.Request) (*http.Response, error) {
		return textResponse(http.StatusTooManyRequests, "rate limited"), nil
	})}

	provider := NewOpenAIProvider(OpenAIProviderConfig{
		APIKey:     "test-key",
		Model:      "gpt-4.1-mini",
		BaseURL:    "https://api.test/v1",
		HTTPClient: client,
	})

	_, err := provider.GenerateItemDetails(context.Background(), enrichment.ItemDetailInput{
		Title:  "Denim jacket",
		Photos: []enrichment.ItemPhotoInput{{ID: "photo-1", DataURL: "data:image/png;base64,abc123"}},
	})

	s.Error(err)
}

type roundTripFunc func(*http.Request) (*http.Response, error)

func (f roundTripFunc) RoundTrip(request *http.Request) (*http.Response, error) {
	return f(request)
}

func jsonResponse(status int, body string) *http.Response {
	response := textResponse(status, body)
	response.Header.Set("Content-Type", "application/json")
	return response
}

func textResponse(status int, body string) *http.Response {
	return &http.Response{
		StatusCode: status,
		Header:     http.Header{},
		Body:       io.NopCloser(bytes.NewBufferString(body)),
	}
}
