package items

import (
	"context"
	"testing"

	"github.com/solthoth/social-media-marketplace-assistant/services/api/internal/domain"
	"github.com/stretchr/testify/suite"
)

type ServiceSuite struct {
	suite.Suite
	repository *memoryRepository
	service    Service
}

func TestServiceSuite(t *testing.T) {
	suite.Run(t, new(ServiceSuite))
}

func (s *ServiceSuite) SetupTest() {
	s.repository = newMemoryRepository()
	s.service = NewService(s.repository)
}

func (s *ServiceSuite) TestCreateItemDefaultsToDraftAndUSD() {
	item, err := s.service.CreateItem(context.Background(), CreateItemInput{
		Title:                      " Leather boots ",
		Category:                   "Shoes",
		OriginalPurchasePriceCents: 2400,
		SellingPriceCents:          4200,
	})

	s.Require().NoError(err)
	s.NotEmpty(item.ID)
	s.Equal("Leather boots", item.Title)
	s.Equal("Shoes", item.Category)
	s.Equal(int64(2400), item.OriginalPurchasePrice.AmountCents)
	s.Equal(int64(4200), item.SellingPrice.AmountCents)
	s.Equal("USD", item.SellingPrice.Currency)
	s.Equal(domain.InventoryStatusDraft, item.Status)
}

func (s *ServiceSuite) TestCreateItemRequiresTitle() {
	_, err := s.service.CreateItem(context.Background(), CreateItemInput{})

	s.ErrorIs(err, ErrInvalidItem)
}

func (s *ServiceSuite) TestCreateItemRejectsUnsupportedCurrency() {
	_, err := s.service.CreateItem(context.Background(), CreateItemInput{
		Title:    "Foreign currency draft",
		Currency: "EUR",
	})

	s.ErrorIs(err, ErrInvalidItem)
}

func (s *ServiceSuite) TestCreateItemDefaultsPricesToZero() {
	item, err := s.service.CreateItem(context.Background(), CreateItemInput{
		Title: "No price draft",
	})

	s.Require().NoError(err)
	s.Equal(int64(0), item.OriginalPurchasePrice.AmountCents)
	s.Equal(int64(0), item.SellingPrice.AmountCents)
}

func (s *ServiceSuite) TestListGetUpdateAndArchiveItem() {
	created, err := s.service.CreateItem(context.Background(), CreateItemInput{
		Title:             "Bracelet",
		Description:       "Silver tone",
		Category:          "Jewelry",
		SellingPriceCents: 1500,
		Currency:          "USD",
	})
	s.Require().NoError(err)

	list, err := s.service.ListItems(context.Background(), ListItemsFilter{})
	s.Require().NoError(err)
	s.Len(list, 1)

	fetched, err := s.service.GetItem(context.Background(), created.ID)
	s.Require().NoError(err)
	s.Equal(created.ID, fetched.ID)

	title := "Vintage bracelet"
	originalPurchasePriceCents := int64(700)
	sellingPriceCents := int64(1800)
	status := domain.InventoryStatusReadyToList
	updated, err := s.service.UpdateItem(context.Background(), created.ID, UpdateItemInput{
		Title:                      &title,
		OriginalPurchasePriceCents: &originalPurchasePriceCents,
		SellingPriceCents:          &sellingPriceCents,
		Status:                     &status,
	})
	s.Require().NoError(err)
	s.Equal("Vintage bracelet", updated.Title)
	s.Equal(int64(700), updated.OriginalPurchasePrice.AmountCents)
	s.Equal(int64(1800), updated.SellingPrice.AmountCents)
	s.Equal(domain.InventoryStatusReadyToList, updated.Status)

	archived, err := s.service.ArchiveItem(context.Background(), created.ID)
	s.Require().NoError(err)
	s.Equal(domain.InventoryStatusArchived, archived.Status)
}
