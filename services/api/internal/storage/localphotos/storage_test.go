package localphotos

import (
	"bytes"
	"context"
	"io"
	"testing"

	"github.com/solthoth/social-media-marketplace-assistant/services/api/internal/domain"
	"github.com/solthoth/social-media-marketplace-assistant/services/api/internal/photos"
	"github.com/stretchr/testify/suite"
)

type StorageSuite struct {
	suite.Suite
	storage Storage
}

func TestStorageSuite(t *testing.T) {
	suite.Run(t, new(StorageSuite))
}

func (s *StorageSuite) SetupTest() {
	s.storage = NewStorage(s.T().TempDir())
}

func (s *StorageSuite) TestSaveOpenAndDeleteObject() {
	stored, err := s.storage.Save(context.Background(), photos.StorageObject{
		StorageID:   "items/item-1/photos/photo-1",
		Variant:     domain.PhotoVariantOriginal,
		Extension:   ".png",
		ContentType: "image/png",
	}, bytes.NewBufferString("image-bytes"))
	s.Require().NoError(err)
	s.Equal("items/item-1/photos/photo-1", stored.StorageID)
	s.Equal(int64(11), stored.SizeBytes)

	content, info, err := s.storage.Open(context.Background(), stored.StorageID, domain.PhotoVariantThumbnail)
	s.Require().NoError(err)
	defer content.Close()

	body, err := io.ReadAll(content)
	s.Require().NoError(err)
	s.Equal("image-bytes", string(body))
	s.Equal("image/png", info.ContentType)

	s.Require().NoError(s.storage.Delete(context.Background(), stored.StorageID))
	_, _, err = s.storage.Open(context.Background(), stored.StorageID, domain.PhotoVariantOriginal)
	s.ErrorIs(err, photos.ErrPhotoNotFound)
}

func (s *StorageSuite) TestRejectsUnsafeStorageIDs() {
	unsafeIDs := []string{
		"../outside",
		"/absolute/path",
		"items/../../outside",
		"items\\item-1\\photos\\photo-1",
		"items//item-1/photos/photo-1",
	}

	for _, storageID := range unsafeIDs {
		s.Run(storageID, func() {
			_, err := s.storage.Save(context.Background(), photos.StorageObject{
				StorageID:   storageID,
				Variant:     domain.PhotoVariantOriginal,
				Extension:   ".png",
				ContentType: "image/png",
			}, bytes.NewBufferString("image-bytes"))
			s.ErrorIs(err, photos.ErrInvalidPhoto)

			_, _, err = s.storage.Open(context.Background(), storageID, domain.PhotoVariantOriginal)
			s.ErrorIs(err, photos.ErrInvalidPhoto)

			s.ErrorIs(s.storage.Delete(context.Background(), storageID), photos.ErrInvalidPhoto)
		})
	}
}

func (s *StorageSuite) TestRejectsUnsafeExtensions() {
	_, err := s.storage.Save(context.Background(), photos.StorageObject{
		StorageID:   "items/item-1/photos/photo-1",
		Variant:     domain.PhotoVariantOriginal,
		Extension:   "../png",
		ContentType: "image/png",
	}, bytes.NewBufferString("image-bytes"))

	s.ErrorIs(err, photos.ErrInvalidPhoto)
}
