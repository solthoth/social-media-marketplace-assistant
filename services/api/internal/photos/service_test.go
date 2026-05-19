package photos

import (
	"bytes"
	"context"
	"io"
	"slices"
	"testing"
	"time"

	"github.com/solthoth/social-media-marketplace-assistant/services/api/internal/domain"
	"github.com/solthoth/social-media-marketplace-assistant/services/api/internal/items"
	"github.com/stretchr/testify/suite"
)

type ServiceSuite struct {
	suite.Suite
	itemRepository  *memoryItemRepository
	photoRepository *memoryPhotoRepository
	storage         *memoryStorage
	service         Service
}

func TestServiceSuite(t *testing.T) {
	suite.Run(t, new(ServiceSuite))
}

func (s *ServiceSuite) SetupTest() {
	s.itemRepository = newMemoryItemRepository()
	s.photoRepository = newMemoryPhotoRepository()
	s.storage = newMemoryStorage()
	s.service = NewService(s.itemRepository, s.photoRepository, s.storage)
	s.itemRepository.items["item-1"] = domain.Item{ID: "item-1", Title: "Denim jacket"}
}

func (s *ServiceSuite) TestUploadPhotoStoresOriginalAndCreatesMetadata() {
	photo, err := s.service.UploadPhoto(context.Background(), "item-1", UploadPhotoInput{
		Filename: " jacket.png ",
		Content:  bytes.NewReader(pngBytes),
	})

	s.Require().NoError(err)
	s.NotEmpty(photo.ID)
	s.Equal("item-1", photo.ItemID)
	s.Equal("jacket.png", photo.Filename)
	s.Equal("image/png", photo.MimeType)
	s.Equal(0, photo.SortOrder)
	s.True(photo.IsPrimary)
	s.Equal("items/item-1/photos/"+photo.ID, photo.StorageID)
	s.Equal(pngBytes, s.storage.objects[photo.StorageID][domain.PhotoVariantOriginal])
}

func (s *ServiceSuite) TestUploadPhotoRejectsUnsupportedContent() {
	_, err := s.service.UploadPhoto(context.Background(), "item-1", UploadPhotoInput{
		Filename: "notes.txt",
		Content:  bytes.NewBufferString("not an image"),
	})

	s.ErrorIs(err, ErrInvalidPhoto)
}

func (s *ServiceSuite) TestListDeleteAndReorderPhotos() {
	first, err := s.service.UploadPhoto(context.Background(), "item-1", UploadPhotoInput{
		Filename: "first.png",
		Content:  bytes.NewReader(pngBytes),
	})
	s.Require().NoError(err)
	second, err := s.service.UploadPhoto(context.Background(), "item-1", UploadPhotoInput{
		Filename: "second.png",
		Content:  bytes.NewReader(pngBytes),
	})
	s.Require().NoError(err)

	list, err := s.service.ListPhotos(context.Background(), "item-1")
	s.Require().NoError(err)
	s.Equal([]string{first.ID, second.ID}, photoIDs(list))
	s.False(second.IsPrimary)

	reordered, err := s.service.ReorderPhotos(context.Background(), "item-1", []string{second.ID, first.ID})
	s.Require().NoError(err)
	s.Equal([]string{second.ID, first.ID}, photoIDs(reordered))
	s.Equal(0, reordered[0].SortOrder)
	s.Equal(1, reordered[1].SortOrder)

	s.Require().NoError(s.service.DeletePhoto(context.Background(), "item-1", first.ID))
	list, err = s.service.ListPhotos(context.Background(), "item-1")
	s.Require().NoError(err)
	s.Equal([]string{second.ID}, photoIDs(list))
}

func (s *ServiceSuite) TestOpenPhotoReturnsStoredContent() {
	uploaded, err := s.service.UploadPhoto(context.Background(), "item-1", UploadPhotoInput{
		Filename: "jacket.png",
		Content:  bytes.NewReader(pngBytes),
	})
	s.Require().NoError(err)

	content, info, err := s.service.OpenPhoto(context.Background(), "item-1", uploaded.ID, domain.PhotoVariantThumbnail)
	s.Require().NoError(err)
	defer content.Close()

	body, err := io.ReadAll(content)
	s.Require().NoError(err)
	s.Equal(pngBytes, body)
	s.Equal("image/png", info.ContentType)
}

func photoIDs(photos []domain.ItemPhoto) []string {
	result := make([]string, 0, len(photos))
	for _, photo := range photos {
		result = append(result, photo.ID)
	}
	return result
}

type memoryItemRepository struct {
	items map[string]domain.Item
}

func newMemoryItemRepository() *memoryItemRepository {
	return &memoryItemRepository{items: map[string]domain.Item{}}
}

func (r *memoryItemRepository) Get(ctx context.Context, id string) (domain.Item, error) {
	item, ok := r.items[id]
	if !ok {
		return domain.Item{}, items.ErrItemNotFound
	}
	return item, nil
}

type memoryPhotoRepository struct {
	photos map[string]domain.ItemPhoto
}

func newMemoryPhotoRepository() *memoryPhotoRepository {
	return &memoryPhotoRepository{photos: map[string]domain.ItemPhoto{}}
}

func (r *memoryPhotoRepository) CreatePhoto(ctx context.Context, photo domain.ItemPhoto) (domain.ItemPhoto, error) {
	r.photos[photo.ID] = photo
	return photo, nil
}

func (r *memoryPhotoRepository) ListPhotos(ctx context.Context, itemID string) ([]domain.ItemPhoto, error) {
	result := []domain.ItemPhoto{}
	for _, photo := range r.photos {
		if photo.ItemID == itemID {
			result = append(result, photo)
		}
	}
	slices.SortFunc(result, func(a domain.ItemPhoto, b domain.ItemPhoto) int {
		return a.SortOrder - b.SortOrder
	})
	return result, nil
}

func (r *memoryPhotoRepository) GetPhoto(ctx context.Context, itemID string, photoID string) (domain.ItemPhoto, error) {
	photo, ok := r.photos[photoID]
	if !ok || photo.ItemID != itemID {
		return domain.ItemPhoto{}, ErrPhotoNotFound
	}
	return photo, nil
}

func (r *memoryPhotoRepository) DeletePhoto(ctx context.Context, itemID string, photoID string) error {
	delete(r.photos, photoID)
	return nil
}

func (r *memoryPhotoRepository) ReorderPhotos(ctx context.Context, itemID string, photoIDs []string) ([]domain.ItemPhoto, error) {
	for index, photoID := range photoIDs {
		photo := r.photos[photoID]
		photo.SortOrder = index
		r.photos[photoID] = photo
	}
	return r.ListPhotos(ctx, itemID)
}

type memoryStorage struct {
	objects map[string]map[domain.PhotoVariant][]byte
}

func newMemoryStorage() *memoryStorage {
	return &memoryStorage{objects: map[string]map[domain.PhotoVariant][]byte{}}
}

func (s *memoryStorage) Save(ctx context.Context, object StorageObject, content io.Reader) (StoredObject, error) {
	body, err := io.ReadAll(content)
	if err != nil {
		return StoredObject{}, err
	}
	if s.objects[object.StorageID] == nil {
		s.objects[object.StorageID] = map[domain.PhotoVariant][]byte{}
	}
	s.objects[object.StorageID][object.Variant] = body
	return StoredObject{StorageID: object.StorageID, SizeBytes: int64(len(body))}, nil
}

func (s *memoryStorage) Open(ctx context.Context, storageID string, variant domain.PhotoVariant) (io.ReadCloser, ObjectInfo, error) {
	body, ok := s.objects[storageID][variant]
	if !ok {
		body = s.objects[storageID][domain.PhotoVariantOriginal]
	}
	return io.NopCloser(bytes.NewReader(body)), ObjectInfo{ContentType: "image/png", SizeBytes: int64(len(body))}, nil
}

func (s *memoryStorage) Delete(ctx context.Context, storageID string) error {
	delete(s.objects, storageID)
	return nil
}

var pngBytes = []byte{
	0x89, 0x50, 0x4e, 0x47, 0x0d, 0x0a, 0x1a, 0x0a,
	0x00, 0x00, 0x00, 0x0d, 0x49, 0x48, 0x44, 0x52,
	0x00, 0x00, 0x00, 0x01, 0x00, 0x00, 0x00, 0x01,
	0x08, 0x02, 0x00, 0x00, 0x00, 0x90, 0x77, 0x53,
	0xde, 0x00, 0x00, 0x00, 0x0c, 0x49, 0x44, 0x41,
	0x54, 0x08, 0xd7, 0x63, 0xf8, 0xcf, 0xc0, 0x00,
	0x00, 0x03, 0x01, 0x01, 0x00, 0x18, 0xdd, 0x8d,
	0xb0, 0x00, 0x00, 0x00, 0x00, 0x49, 0x45, 0x4e,
	0x44, 0xae, 0x42, 0x60, 0x82,
}

var _ = time.Now
