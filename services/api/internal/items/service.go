package items

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/solthoth/social-media-marketplace-assistant/services/api/internal/domain"
)

var (
	ErrInvalidItem             = errors.New("invalid item")
	ErrInvalidStatusTransition = errors.New("invalid status transition")
	ErrItemNotFound            = errors.New("item not found")
)

type Repository interface {
	Create(ctx context.Context, item domain.Item) (domain.Item, error)
	List(ctx context.Context, filter ListItemsFilter) ([]domain.Item, error)
	Get(ctx context.Context, id string) (domain.Item, error)
	Update(ctx context.Context, item domain.Item) (domain.Item, error)
}

type Service struct {
	repository Repository
}

func NewService(repository Repository) Service {
	return Service{
		repository: repository,
	}
}

type CreateItemInput struct {
	Title                      string
	Description                string
	Category                   string
	Size                       string
	Condition                  string
	OriginalPurchasePriceCents int64
	SellingPriceCents          int64
	Currency                   string
	Notes                      string
}

type UpdateItemInput struct {
	Title                      *string
	Description                *string
	Category                   *string
	Size                       *string
	Condition                  *string
	OriginalPurchasePriceCents *int64
	SellingPriceCents          *int64
	Currency                   *string
	Status                     *domain.InventoryStatus
	Notes                      *string
}

type ListItemsFilter struct {
	Status *domain.InventoryStatus
}

func (s Service) CreateItem(ctx context.Context, input CreateItemInput) (domain.Item, error) {
	currency := strings.TrimSpace(input.Currency)
	if currency == "" {
		currency = "USD"
	}

	item := domain.NewDraftItem(domain.NewItemInput{
		Title:                      input.Title,
		Description:                input.Description,
		Category:                   input.Category,
		Size:                       input.Size,
		Condition:                  input.Condition,
		OriginalPurchasePriceCents: input.OriginalPurchasePriceCents,
		SellingPriceCents:          input.SellingPriceCents,
		Currency:                   currency,
		Notes:                      input.Notes,
	})
	item.ID = uuid.NewString()

	if err := validateItem(item); err != nil {
		return domain.Item{}, err
	}

	return s.repository.Create(ctx, item)
}

func (s Service) ListItems(ctx context.Context, filter ListItemsFilter) ([]domain.Item, error) {
	if filter.Status != nil && !filter.Status.IsValid() {
		return nil, ErrInvalidItem
	}
	return s.repository.List(ctx, filter)
}

func (s Service) GetItem(ctx context.Context, id string) (domain.Item, error) {
	if strings.TrimSpace(id) == "" {
		return domain.Item{}, ErrItemNotFound
	}
	return s.repository.Get(ctx, id)
}

func (s Service) UpdateItem(ctx context.Context, id string, input UpdateItemInput) (domain.Item, error) {
	item, err := s.GetItem(ctx, id)
	if err != nil {
		return domain.Item{}, err
	}

	if input.Title != nil {
		item.Title = strings.TrimSpace(*input.Title)
	}
	if input.Description != nil {
		item.Description = strings.TrimSpace(*input.Description)
	}
	if input.Category != nil {
		item.Category = strings.TrimSpace(*input.Category)
	}
	if input.Size != nil {
		item.Size = strings.TrimSpace(*input.Size)
	}
	if input.Condition != nil {
		item.Condition = strings.TrimSpace(*input.Condition)
	}
	if input.OriginalPurchasePriceCents != nil {
		item.OriginalPurchasePrice.AmountCents = *input.OriginalPurchasePriceCents
	}
	if input.SellingPriceCents != nil {
		item.SellingPrice.AmountCents = *input.SellingPriceCents
	}
	if input.Currency != nil {
		currency := strings.ToUpper(strings.TrimSpace(*input.Currency))
		item.OriginalPurchasePrice.Currency = currency
		item.SellingPrice.Currency = currency
	}
	if input.Status != nil {
		if !item.Status.CanTransitionTo(*input.Status) {
			return domain.Item{}, ErrInvalidStatusTransition
		}
		item.Status = *input.Status
	}
	if input.Notes != nil {
		item.Notes = strings.TrimSpace(*input.Notes)
	}
	item.UpdatedAt = time.Now().UTC()

	if err := validateItem(item); err != nil {
		return domain.Item{}, err
	}

	return s.repository.Update(ctx, item)
}

func (s Service) ArchiveItem(ctx context.Context, id string) (domain.Item, error) {
	status := domain.InventoryStatusArchived
	return s.UpdateItem(ctx, id, UpdateItemInput{
		Status: &status,
	})
}

func validateItem(item domain.Item) error {
	if strings.TrimSpace(item.Title) == "" {
		return ErrInvalidItem
	}
	if !item.Status.IsValid() {
		return ErrInvalidItem
	}
	if err := item.OriginalPurchasePrice.Validate(); err != nil {
		return ErrInvalidItem
	}
	if err := item.SellingPrice.Validate(); err != nil {
		return ErrInvalidItem
	}
	return nil
}
