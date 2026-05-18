package domain

import "testing"

func TestNewDraftItemUsesFlexibleCategoryAndMoney(t *testing.T) {
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

	if item.Category != "90s outerwear" {
		t.Fatalf("expected flexible category text, got %q", item.Category)
	}
	if item.Price.AmountCents != 2500 {
		t.Fatalf("expected price in cents, got %d", item.Price.AmountCents)
	}
	if item.Price.Currency != "USD" {
		t.Fatalf("expected USD currency, got %q", item.Price.Currency)
	}
	if item.Status != InventoryStatusDraft {
		t.Fatalf("expected draft status, got %q", item.Status)
	}
}

func TestInventoryStatusValidation(t *testing.T) {
	valid := []InventoryStatus{
		InventoryStatusDraft,
		InventoryStatusReadyToList,
		InventoryStatusListed,
		InventoryStatusSold,
		InventoryStatusArchived,
	}

	for _, status := range valid {
		if !status.IsValid() {
			t.Fatalf("expected %q to be valid", status)
		}
	}

	if InventoryStatus("deleted").IsValid() {
		t.Fatal("expected unknown status to be invalid")
	}
}

func TestMoneyValidation(t *testing.T) {
	if err := (Money{AmountCents: 0, Currency: "USD"}).Validate(); err != nil {
		t.Fatalf("expected zero-dollar price to be valid: %v", err)
	}

	if err := (Money{AmountCents: -1, Currency: "USD"}).Validate(); err == nil {
		t.Fatal("expected negative price to be invalid")
	}

	if err := (Money{AmountCents: 100, Currency: "usd"}).Validate(); err == nil {
		t.Fatal("expected lowercase currency to be invalid")
	}
}
