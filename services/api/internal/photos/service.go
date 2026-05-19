package photos

import (
	"bytes"
	"context"
	"errors"
	"io"
	"net/http"
	"path/filepath"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/solthoth/social-media-marketplace-assistant/services/api/internal/domain"
	"github.com/solthoth/social-media-marketplace-assistant/services/api/internal/items"
)

const maxPhotoBytes = 10 * 1024 * 1024

var (
	ErrInvalidPhoto  = errors.New("invalid photo")
	ErrPhotoNotFound = errors.New("photo not found")
)

type ItemRepository interface {
	Get(ctx context.Context, id string) (domain.Item, error)
}

type Repository interface {
	CreatePhoto(ctx context.Context, photo domain.ItemPhoto) (domain.ItemPhoto, error)
	ListPhotos(ctx context.Context, itemID string) ([]domain.ItemPhoto, error)
	GetPhoto(ctx context.Context, itemID string, photoID string) (domain.ItemPhoto, error)
	DeletePhoto(ctx context.Context, itemID string, photoID string) error
	ReorderPhotos(ctx context.Context, itemID string, photoIDs []string) ([]domain.ItemPhoto, error)
}

type Storage interface {
	Save(ctx context.Context, object StorageObject, content io.Reader) (StoredObject, error)
	Open(ctx context.Context, storageID string, variant domain.PhotoVariant) (io.ReadCloser, ObjectInfo, error)
	Delete(ctx context.Context, storageID string) error
}

type StorageObject struct {
	StorageID   string
	Variant     domain.PhotoVariant
	Extension   string
	ContentType string
}

type StoredObject struct {
	StorageID string
	SizeBytes int64
}

type ObjectInfo struct {
	ContentType string
	SizeBytes   int64
}

type Service struct {
	items      ItemRepository
	repository Repository
	storage    Storage
}

func NewService(items ItemRepository, repository Repository, storage Storage) Service {
	return Service{
		items:      items,
		repository: repository,
		storage:    storage,
	}
}

type UploadPhotoInput struct {
	Filename string
	Content  io.Reader
}

func (s Service) UploadPhoto(ctx context.Context, itemID string, input UploadPhotoInput) (domain.ItemPhoto, error) {
	if _, err := s.items.Get(ctx, itemID); err != nil {
		return domain.ItemPhoto{}, err
	}
	if input.Content == nil {
		return domain.ItemPhoto{}, ErrInvalidPhoto
	}

	body, err := readLimited(input.Content)
	if err != nil {
		return domain.ItemPhoto{}, err
	}
	mimeType := http.DetectContentType(body)
	extension, ok := extensionForMimeType(mimeType)
	if !ok {
		return domain.ItemPhoto{}, ErrInvalidPhoto
	}

	existing, err := s.repository.ListPhotos(ctx, itemID)
	if err != nil {
		return domain.ItemPhoto{}, err
	}

	photoID := uuid.NewString()
	storageID := "items/" + itemID + "/photos/" + photoID
	if _, err := s.storage.Save(ctx, StorageObject{
		StorageID:   storageID,
		Variant:     domain.PhotoVariantOriginal,
		Extension:   extension,
		ContentType: mimeType,
	}, bytes.NewReader(body)); err != nil {
		return domain.ItemPhoto{}, err
	}

	photo := domain.ItemPhoto{
		ID:        photoID,
		ItemID:    itemID,
		StorageID: storageID,
		Filename:  sanitizeFilename(input.Filename, extension),
		MimeType:  mimeType,
		SortOrder: len(existing),
		IsPrimary: len(existing) == 0,
		CreatedAt: time.Now().UTC(),
	}
	return s.repository.CreatePhoto(ctx, photo)
}

func (s Service) ListPhotos(ctx context.Context, itemID string) ([]domain.ItemPhoto, error) {
	if _, err := s.items.Get(ctx, itemID); err != nil {
		return nil, err
	}
	return s.repository.ListPhotos(ctx, itemID)
}

func (s Service) OpenPhoto(ctx context.Context, itemID string, photoID string, variant domain.PhotoVariant) (io.ReadCloser, ObjectInfo, error) {
	if !variant.IsValid() {
		return nil, ObjectInfo{}, ErrInvalidPhoto
	}
	photo, err := s.repository.GetPhoto(ctx, itemID, photoID)
	if err != nil {
		return nil, ObjectInfo{}, err
	}
	return s.storage.Open(ctx, photo.StorageID, variant)
}

func (s Service) DeletePhoto(ctx context.Context, itemID string, photoID string) error {
	photo, err := s.repository.GetPhoto(ctx, itemID, photoID)
	if err != nil {
		return err
	}
	if err := s.storage.Delete(ctx, photo.StorageID); err != nil {
		return err
	}
	return s.repository.DeletePhoto(ctx, itemID, photoID)
}

func (s Service) ReorderPhotos(ctx context.Context, itemID string, photoIDs []string) ([]domain.ItemPhoto, error) {
	if _, err := s.items.Get(ctx, itemID); err != nil {
		return nil, err
	}
	if len(photoIDs) == 0 {
		return nil, ErrInvalidPhoto
	}
	return s.repository.ReorderPhotos(ctx, itemID, photoIDs)
}

func readLimited(content io.Reader) ([]byte, error) {
	body, err := io.ReadAll(io.LimitReader(content, maxPhotoBytes+1))
	if err != nil {
		return nil, err
	}
	if len(body) == 0 || len(body) > maxPhotoBytes {
		return nil, ErrInvalidPhoto
	}
	return body, nil
}

func extensionForMimeType(mimeType string) (string, bool) {
	switch mimeType {
	case "image/jpeg":
		return ".jpg", true
	case "image/png":
		return ".png", true
	case "image/webp":
		return ".webp", true
	default:
		return "", false
	}
}

func sanitizeFilename(filename string, extension string) string {
	name := strings.TrimSpace(filepath.Base(filename))
	if name == "." || name == "/" || name == "" {
		return "photo" + extension
	}
	return name
}

func IsItemNotFound(err error) bool {
	return errors.Is(err, items.ErrItemNotFound)
}
