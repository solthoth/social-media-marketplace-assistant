package sqlite

import (
	"context"
	"database/sql"
	"testing"
	"time"

	"github.com/solthoth/social-media-marketplace-assistant/services/api/internal/domain"
	"github.com/solthoth/social-media-marketplace-assistant/services/api/internal/items"
	"github.com/solthoth/social-media-marketplace-assistant/services/api/internal/photos"
	"github.com/stretchr/testify/suite"
	_ "modernc.org/sqlite"
)

type PhotoRepositorySuite struct {
	suite.Suite
	db         *sql.DB
	items      ItemRepository
	repository PhotoRepository
}

func TestPhotoRepositorySuite(t *testing.T) {
	suite.Run(t, new(PhotoRepositorySuite))
}

func (s *PhotoRepositorySuite) SetupTest() {
	db, err := sql.Open("sqlite", ":memory:")
	s.Require().NoError(err)
	s.Require().NoError(ApplyMigrations(context.Background(), db))
	s.db = db
	s.items = NewItemRepository(db)
	s.repository = NewPhotoRepository(db)

	_, err = s.items.Create(context.Background(), domain.Item{
		ID: "item-1", Title: "Denim jacket",
		OriginalPurchasePrice: domain.Money{Currency: "USD"},
		SellingPrice:          domain.Money{Currency: "USD"},
		Status:                domain.InventoryStatusDraft,
		CreatedAt:             time.Now().UTC(),
		UpdatedAt:             time.Now().UTC(),
	})
	s.Require().NoError(err)
}

func (s *PhotoRepositorySuite) TearDownTest() {
	s.Require().NoError(s.db.Close())
}

func (s *PhotoRepositorySuite) TestCreateListGetReorderAndDeletePhotos() {
	first := domain.ItemPhoto{
		ID: "photo-1", ItemID: "item-1", StorageID: "items/item-1/photos/photo-1",
		Filename: "front.png", MimeType: "image/png", SortOrder: 0, IsPrimary: true, CreatedAt: time.Now().UTC(),
	}
	second := domain.ItemPhoto{
		ID: "photo-2", ItemID: "item-1", StorageID: "items/item-1/photos/photo-2",
		Filename: "back.png", MimeType: "image/png", SortOrder: 1, CreatedAt: time.Now().UTC(),
	}

	created, err := s.repository.CreatePhoto(context.Background(), first)
	s.Require().NoError(err)
	s.Equal(first.ID, created.ID)
	_, err = s.repository.CreatePhoto(context.Background(), second)
	s.Require().NoError(err)

	fetched, err := s.repository.GetPhoto(context.Background(), "item-1", first.ID)
	s.Require().NoError(err)
	s.True(fetched.IsPrimary)

	reordered, err := s.repository.ReorderPhotos(context.Background(), "item-1", []string{second.ID, first.ID})
	s.Require().NoError(err)
	s.Equal([]string{second.ID, first.ID}, photoRepositoryIDs(reordered))
	s.Equal(0, reordered[0].SortOrder)

	s.Require().NoError(s.repository.DeletePhoto(context.Background(), "item-1", first.ID))
	list, err := s.repository.ListPhotos(context.Background(), "item-1")
	s.Require().NoError(err)
	s.Equal([]string{second.ID}, photoRepositoryIDs(list))

	_, err = s.repository.GetPhoto(context.Background(), "item-1", first.ID)
	s.ErrorIs(err, photos.ErrPhotoNotFound)
}

func (s *PhotoRepositorySuite) TestReorderRejectsMissingPhotoIDs() {
	_, err := s.repository.ReorderPhotos(context.Background(), "item-1", []string{"missing"})

	s.ErrorIs(err, photos.ErrInvalidPhoto)
}

func photoRepositoryIDs(photos []domain.ItemPhoto) []string {
	result := make([]string, 0, len(photos))
	for _, photo := range photos {
		result = append(result, photo.ID)
	}
	return result
}

var _ = items.ErrItemNotFound
