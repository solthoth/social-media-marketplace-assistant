package httpserver

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"slices"
	"sync"
	"testing"

	"github.com/solthoth/social-media-marketplace-assistant/services/api/internal/domain"
	"github.com/solthoth/social-media-marketplace-assistant/services/api/internal/items"
	"github.com/solthoth/social-media-marketplace-assistant/services/api/internal/photos"
	"github.com/stretchr/testify/suite"
)

type PhotosHandlerSuite struct {
	suite.Suite
	router http.Handler
}

func TestPhotosHandlerSuite(t *testing.T) {
	suite.Run(t, new(PhotosHandlerSuite))
}

func (s *PhotosHandlerSuite) SetupTest() {
	itemRepository := newHandlerMemoryRepository()
	itemService := items.NewService(itemRepository)
	photoService := photos.NewService(itemRepository, newHandlerPhotoRepository(), newHandlerPhotoStorage())
	s.router = NewRouter(RouterDependencies{ItemService: &itemService, PhotoService: &photoService})
}

func (s *PhotosHandlerSuite) TestUploadListReadReorderAndDeletePhotos() {
	createResponse := s.jsonRequest(http.MethodPost, "/items", bytes.NewBufferString(`{"title":"Denim jacket"}`))
	s.Equal(http.StatusCreated, createResponse.Code)
	var item itemResponse
	s.Require().NoError(json.NewDecoder(createResponse.Body).Decode(&item))

	frontResponse := s.multipartRequest(http.MethodPost, "/items/"+item.ID+"/photos", "front.png", pngBytes)
	s.Equal(http.StatusCreated, frontResponse.Code)
	var front photoResponse
	s.Require().NoError(json.NewDecoder(frontResponse.Body).Decode(&front))
	s.NotEmpty(front.ID)
	s.Equal("front.png", front.Filename)
	s.Equal("image/png", front.MimeType)
	s.True(front.IsPrimary)
	s.Contains(front.ContentURLs, "thumbnail")

	backResponse := s.multipartRequest(http.MethodPost, "/items/"+item.ID+"/photos", "back.png", pngBytes)
	s.Equal(http.StatusCreated, backResponse.Code)
	var back photoResponse
	s.Require().NoError(json.NewDecoder(backResponse.Body).Decode(&back))

	listResponse := s.jsonRequest(http.MethodGet, "/items/"+item.ID+"/photos", nil)
	s.Equal(http.StatusOK, listResponse.Code)
	var list listPhotosResponse
	s.Require().NoError(json.NewDecoder(listResponse.Body).Decode(&list))
	s.Equal([]string{front.ID, back.ID}, photoResponseIDs(list.Photos))

	contentResponse := s.jsonRequest(http.MethodGet, "/items/"+item.ID+"/photos/"+front.ID+"/content?variant=thumbnail", nil)
	s.Equal(http.StatusOK, contentResponse.Code)
	s.Equal("image/png", contentResponse.Header().Get("Content-Type"))
	s.Equal(pngBytes, contentResponse.Body.Bytes())

	reorderResponse := s.jsonRequest(
		http.MethodPatch,
		"/items/"+item.ID+"/photos/order",
		bytes.NewBufferString(`{"photo_ids":["`+back.ID+`","`+front.ID+`"]}`),
	)
	s.Equal(http.StatusOK, reorderResponse.Code)
	var reordered listPhotosResponse
	s.Require().NoError(json.NewDecoder(reorderResponse.Body).Decode(&reordered))
	s.Equal([]string{back.ID, front.ID}, photoResponseIDs(reordered.Photos))

	primaryResponse := s.jsonRequest(http.MethodPatch, "/items/"+item.ID+"/photos/"+back.ID+"/primary", nil)
	s.Equal(http.StatusOK, primaryResponse.Code)
	var primary listPhotosResponse
	s.Require().NoError(json.NewDecoder(primaryResponse.Body).Decode(&primary))
	s.True(primary.Photos[0].IsPrimary)
	s.False(primary.Photos[1].IsPrimary)

	deleteResponse := s.jsonRequest(http.MethodDelete, "/items/"+item.ID+"/photos/"+front.ID, nil)
	s.Equal(http.StatusNoContent, deleteResponse.Code)
}

func (s *PhotosHandlerSuite) TestUploadRejectsMissingItem() {
	response := s.multipartRequest(http.MethodPost, "/items/missing/photos", "front.png", pngBytes)

	s.Equal(http.StatusNotFound, response.Code)
}

func (s *PhotosHandlerSuite) jsonRequest(method string, target string, body io.Reader) *httptest.ResponseRecorder {
	request := httptest.NewRequest(method, target, body)
	request.Header.Set("Content-Type", "application/json")
	response := httptest.NewRecorder()
	s.router.ServeHTTP(response, request)
	return response
}

func (s *PhotosHandlerSuite) multipartRequest(method string, target string, filename string, content []byte) *httptest.ResponseRecorder {
	var body bytes.Buffer
	writer := multipart.NewWriter(&body)
	part, err := writer.CreateFormFile("photo", filename)
	s.Require().NoError(err)
	_, err = part.Write(content)
	s.Require().NoError(err)
	s.Require().NoError(writer.Close())

	request := httptest.NewRequest(method, target, &body)
	request.Header.Set("Content-Type", writer.FormDataContentType())
	response := httptest.NewRecorder()
	s.router.ServeHTTP(response, request)
	return response
}

func photoResponseIDs(photos []photoResponse) []string {
	result := make([]string, 0, len(photos))
	for _, photo := range photos {
		result = append(result, photo.ID)
	}
	return result
}

type handlerPhotoRepository struct {
	mu     sync.RWMutex
	photos map[string]domain.ItemPhoto
}

func newHandlerPhotoRepository() *handlerPhotoRepository {
	return &handlerPhotoRepository{photos: map[string]domain.ItemPhoto{}}
}

func (r *handlerPhotoRepository) CreatePhoto(ctx context.Context, photo domain.ItemPhoto) (domain.ItemPhoto, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.photos[photo.ID] = photo
	return photo, nil
}

func (r *handlerPhotoRepository) ListPhotos(ctx context.Context, itemID string) ([]domain.ItemPhoto, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
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

func (r *handlerPhotoRepository) GetPhoto(ctx context.Context, itemID string, photoID string) (domain.ItemPhoto, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	photo, ok := r.photos[photoID]
	if !ok || photo.ItemID != itemID {
		return domain.ItemPhoto{}, photos.ErrPhotoNotFound
	}
	return photo, nil
}

func (r *handlerPhotoRepository) DeletePhoto(ctx context.Context, itemID string, photoID string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	delete(r.photos, photoID)
	return nil
}

func (r *handlerPhotoRepository) ReorderPhotos(ctx context.Context, itemID string, photoIDs []string) ([]domain.ItemPhoto, error) {
	r.mu.Lock()
	for index, photoID := range photoIDs {
		photo := r.photos[photoID]
		photo.SortOrder = index
		r.photos[photoID] = photo
	}
	r.mu.Unlock()
	return r.ListPhotos(ctx, itemID)
}

func (r *handlerPhotoRepository) SetPrimaryPhoto(ctx context.Context, itemID string, photoID string) ([]domain.ItemPhoto, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	photo, ok := r.photos[photoID]
	if !ok || photo.ItemID != itemID {
		return nil, photos.ErrPhotoNotFound
	}
	for id, existing := range r.photos {
		if existing.ItemID == itemID {
			existing.IsPrimary = id == photoID
			r.photos[id] = existing
		}
	}
	return r.listPhotosLocked(itemID), nil
}

func (r *handlerPhotoRepository) listPhotosLocked(itemID string) []domain.ItemPhoto {
	result := []domain.ItemPhoto{}
	for _, photo := range r.photos {
		if photo.ItemID == itemID {
			result = append(result, photo)
		}
	}
	slices.SortFunc(result, func(a domain.ItemPhoto, b domain.ItemPhoto) int {
		return a.SortOrder - b.SortOrder
	})
	return result
}

type handlerPhotoStorage struct {
	mu      sync.RWMutex
	objects map[string][]byte
}

func newHandlerPhotoStorage() *handlerPhotoStorage {
	return &handlerPhotoStorage{objects: map[string][]byte{}}
}

func (s *handlerPhotoStorage) Save(ctx context.Context, object photos.StorageObject, content io.Reader) (photos.StoredObject, error) {
	body, err := io.ReadAll(content)
	if err != nil {
		return photos.StoredObject{}, err
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	s.objects[object.StorageID] = body
	return photos.StoredObject{StorageID: object.StorageID, SizeBytes: int64(len(body))}, nil
}

func (s *handlerPhotoStorage) Open(ctx context.Context, storageID string, variant domain.PhotoVariant) (io.ReadCloser, photos.ObjectInfo, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	body, ok := s.objects[storageID]
	if !ok {
		return nil, photos.ObjectInfo{}, photos.ErrPhotoNotFound
	}
	return io.NopCloser(bytes.NewReader(body)), photos.ObjectInfo{ContentType: "image/png", SizeBytes: int64(len(body))}, nil
}

func (s *handlerPhotoStorage) Delete(ctx context.Context, storageID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
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
