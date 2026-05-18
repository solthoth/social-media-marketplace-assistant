package items

import (
	"context"

	"github.com/solthoth/social-media-marketplace-assistant/services/api/internal/domain"
)

type Repository interface {
	Create(ctx context.Context, item domain.Item) (domain.Item, error)
}

type Service struct {
	repository Repository
}

func NewService(repository Repository) Service {
	return Service{
		repository: repository,
	}
}
