package ai

import (
	"context"

	"github.com/solthoth/social-media-marketplace-assistant/services/api/internal/enrichment"
)

type FakeProvider struct{}

func (p FakeProvider) GenerateItemDetails(ctx context.Context, input enrichment.ItemDetailInput) (enrichment.ItemDetailSuggestion, error) {
	return enrichment.ItemDetailSuggestion{
		Description: "AI-generated draft details for " + input.Title + ".",
		Category:    "Uncategorized",
		Condition:   "Review needed",
		Notes:       "Review generated details before listing.",
	}, nil
}
