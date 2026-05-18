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
		Title:      " Leather boots ",
		Category:   "Shoes",
		PriceCents: 4200,
	})

	s.Require().NoError(err)
	s.NotEmpty(item.ID)
	s.Equal("Leather boots", item.Title)
	s.Equal("Shoes", item.Category)
	s.Equal(int64(4200), item.Price.AmountCents)
	s.Equal("USD", item.Price.Currency)
	s.Equal(domain.InventoryStatusDraft, item.Status)
}

func (s *ServiceSuite) TestCreateItemRequiresTitle() {
	_, err := s.service.CreateItem(context.Background(), CreateItemInput{
		PriceCents: 100,
	})

	s.ErrorIs(err, ErrInvalidItem)
}

func (s *ServiceSuite) TestListGetUpdateAndArchiveItem() {
	created, err := s.service.CreateItem(context.Background(), CreateItemInput{
		Title:       "Bracelet",
		Description: "Silver tone",
		Category:    "Jewelry",
		PriceCents:  1500,
		Currency:    "USD",
	})
	s.Require().NoError(err)

	list, err := s.service.ListItems(context.Background(), ListItemsFilter{})
	s.Require().NoError(err)
	s.Len(list, 1)

	fetched, err := s.service.GetItem(context.Background(), created.ID)
	s.Require().NoError(err)
	s.Equal(created.ID, fetched.ID)

	title := "Vintage bracelet"
	status := domain.InventoryStatusReadyToList
	updated, err := s.service.UpdateItem(context.Background(), created.ID, UpdateItemInput{
		Title:  &title,
		Status: &status,
	})
	s.Require().NoError(err)
	s.Equal("Vintage bracelet", updated.Title)
	s.Equal(domain.InventoryStatusReadyToList, updated.Status)

	archived, err := s.service.ArchiveItem(context.Background(), created.ID)
	s.Require().NoError(err)
	s.Equal(domain.InventoryStatusArchived, archived.Status)
}
