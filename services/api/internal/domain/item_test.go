package domain

import (
	"testing"

	"github.com/stretchr/testify/suite"
)

type ItemSuite struct {
	suite.Suite
}

func TestItemSuite(t *testing.T) {
	suite.Run(t, new(ItemSuite))
}

func (s *ItemSuite) TestNewDraftItemUsesFlexibleCategoryAndMoney() {
	item := NewDraftItem(NewItemInput{
		Title:                      "Vintage denim jacket",
		Description:                "Medium wash denim jacket",
		Category:                   "90s outerwear",
		Size:                       "M",
		Condition:                  "Good",
		OriginalPurchasePriceCents: 1200,
		SellingPriceCents:          2500,
		Currency:                   "USD",
		Notes:                      "Check left sleeve before listing",
	})

	s.Equal("90s outerwear", item.Category)
	s.Equal(int64(1200), item.OriginalPurchasePrice.AmountCents)
	s.Equal(int64(2500), item.SellingPrice.AmountCents)
	s.Equal("USD", item.OriginalPurchasePrice.Currency)
	s.Equal("USD", item.SellingPrice.Currency)
	s.Equal(InventoryStatusDraft, item.Status)
}

func (s *ItemSuite) TestInventoryStatusValidation() {
	valid := []InventoryStatus{
		InventoryStatusDraft,
		InventoryStatusReadyToList,
		InventoryStatusListed,
		InventoryStatusSold,
		InventoryStatusArchived,
	}

	for _, status := range valid {
		s.True(status.IsValid(), "expected %q to be valid", status)
	}

	s.False(InventoryStatus("deleted").IsValid())
}

func (s *ItemSuite) TestInventoryStatusTransitions() {
	tests := []struct {
		name    string
		current InventoryStatus
		next    InventoryStatus
		want    bool
	}{
		{name: "same status is allowed", current: InventoryStatusDraft, next: InventoryStatusDraft, want: true},
		{name: "draft can become ready", current: InventoryStatusDraft, next: InventoryStatusReadyToList, want: true},
		{name: "draft can be archived", current: InventoryStatusDraft, next: InventoryStatusArchived, want: true},
		{name: "draft cannot become listed", current: InventoryStatusDraft, next: InventoryStatusListed, want: false},
		{name: "ready can return to draft", current: InventoryStatusReadyToList, next: InventoryStatusDraft, want: true},
		{name: "ready can become listed", current: InventoryStatusReadyToList, next: InventoryStatusListed, want: true},
		{name: "listed can return to ready", current: InventoryStatusListed, next: InventoryStatusReadyToList, want: true},
		{name: "listed can become sold", current: InventoryStatusListed, next: InventoryStatusSold, want: true},
		{name: "listed can be archived", current: InventoryStatusListed, next: InventoryStatusArchived, want: true},
		{name: "sold can return to listed", current: InventoryStatusSold, next: InventoryStatusListed, want: true},
		{name: "sold can be archived", current: InventoryStatusSold, next: InventoryStatusArchived, want: true},
		{name: "sold cannot become ready", current: InventoryStatusSold, next: InventoryStatusReadyToList, want: false},
		{name: "archived can be restored to draft", current: InventoryStatusArchived, next: InventoryStatusDraft, want: true},
		{name: "archived cannot become listed", current: InventoryStatusArchived, next: InventoryStatusListed, want: false},
		{name: "invalid next status is rejected", current: InventoryStatusDraft, next: InventoryStatus("deleted"), want: false},
	}

	for _, test := range tests {
		s.Run(test.name, func() {
			s.Equal(test.want, test.current.CanTransitionTo(test.next))
		})
	}
}

func (s *ItemSuite) TestMoneyValidation() {
	s.NoError((Money{AmountCents: 0, Currency: "USD"}).Validate())
	s.Error((Money{AmountCents: -1, Currency: "USD"}).Validate())
	s.Error((Money{AmountCents: 100, Currency: "usd"}).Validate())
	s.Error((Money{AmountCents: 100, Currency: "EUR"}).Validate())
}
