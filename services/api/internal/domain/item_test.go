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
		Title:       "Vintage denim jacket",
		Description: "Medium wash denim jacket",
		Category:    "90s outerwear",
		Size:        "M",
		Condition:   "Good",
		PriceCents:  2500,
		Currency:    "USD",
		Notes:       "Check left sleeve before listing",
	})

	s.Equal("90s outerwear", item.Category)
	s.Equal(int64(2500), item.Price.AmountCents)
	s.Equal("USD", item.Price.Currency)
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

func (s *ItemSuite) TestMoneyValidation() {
	s.NoError((Money{AmountCents: 0, Currency: "USD"}).Validate())
	s.Error((Money{AmountCents: -1, Currency: "USD"}).Validate())
	s.Error((Money{AmountCents: 100, Currency: "usd"}).Validate())
}
