package domain

import (
	"errors"
	"strings"
	"time"
)

type InventoryStatus string

const (
	InventoryStatusDraft       InventoryStatus = "draft"
	InventoryStatusReadyToList InventoryStatus = "ready_to_list"
	InventoryStatusListed      InventoryStatus = "listed"
	InventoryStatusSold        InventoryStatus = "sold"
	InventoryStatusArchived    InventoryStatus = "archived"
)

type ListingStatus string

const (
	ListingStatusDraft     ListingStatus = "draft"
	ListingStatusPublished ListingStatus = "published"
	ListingStatusFailed    ListingStatus = "failed"
	ListingStatusRemoved   ListingStatus = "removed"
)

type Currency string

const (
	CurrencyUSD Currency = "USD"
)

func (c Currency) IsValid() bool {
	return c == CurrencyUSD
}

type Money struct {
	AmountCents int64
	Currency    string
}

type Item struct {
	ID                    string
	Title                 string
	Description           string
	Category              string
	Size                  string
	Condition             string
	OriginalPurchasePrice Money
	SellingPrice          Money
	Status                InventoryStatus
	Notes                 string
	CreatedAt             time.Time
	UpdatedAt             time.Time
}

type ItemPhoto struct {
	ID        string
	ItemID    string
	StorageID string
	Filename  string
	MimeType  string
	SortOrder int
	IsPrimary bool
	CreatedAt time.Time
}

type ConnectedAccount struct {
	ID              string
	Platform        string
	DisplayName     string
	Status          string
	PermissionScope string
	LastValidatedAt *time.Time
	CreatedAt       time.Time
	UpdatedAt       time.Time
}

type Listing struct {
	ID                 string
	ItemID             string
	ConnectedAccountID string
	ExternalID         string
	Status             ListingStatus
	PublishedAt        *time.Time
	CreatedAt          time.Time
	UpdatedAt          time.Time
}

type ListingAttempt struct {
	ID        string
	ListingID string
	Status    ListingStatus
	Message   string
	CreatedAt time.Time
}

type Sale struct {
	ID        string
	ItemID    string
	SalePrice Money
	SoldAt    time.Time
	Platform  string
	AccountID string
	CreatedAt time.Time
}

type NewItemInput struct {
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

func NewDraftItem(input NewItemInput) Item {
	now := time.Now().UTC()
	return Item{
		Title:       strings.TrimSpace(input.Title),
		Description: strings.TrimSpace(input.Description),
		Category:    strings.TrimSpace(input.Category),
		Size:        strings.TrimSpace(input.Size),
		Condition:   strings.TrimSpace(input.Condition),
		OriginalPurchasePrice: Money{
			AmountCents: input.OriginalPurchasePriceCents,
			Currency:    strings.ToUpper(strings.TrimSpace(input.Currency)),
		},
		SellingPrice: Money{
			AmountCents: input.SellingPriceCents,
			Currency:    strings.ToUpper(strings.TrimSpace(input.Currency)),
		},
		Status:    InventoryStatusDraft,
		Notes:     strings.TrimSpace(input.Notes),
		CreatedAt: now,
		UpdatedAt: now,
	}
}

func (s InventoryStatus) IsValid() bool {
	switch s {
	case InventoryStatusDraft,
		InventoryStatusReadyToList,
		InventoryStatusListed,
		InventoryStatusSold,
		InventoryStatusArchived:
		return true
	default:
		return false
	}
}

func (m Money) Validate() error {
	if m.AmountCents < 0 {
		return errors.New("amount cents must be greater than or equal to zero")
	}
	if len(m.Currency) != 3 || m.Currency != strings.ToUpper(m.Currency) {
		return errors.New("currency must be a three-letter uppercase code")
	}
	if !Currency(m.Currency).IsValid() {
		return errors.New("currency must be one of the supported values")
	}
	return nil
}
