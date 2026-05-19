//go:build integration

package integration

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"testing"

	"github.com/solthoth/social-media-marketplace-assistant/services/api/internal/httpserver"
	"github.com/solthoth/social-media-marketplace-assistant/services/api/internal/items"
	"github.com/solthoth/social-media-marketplace-assistant/services/api/internal/photos"
	"github.com/solthoth/social-media-marketplace-assistant/services/api/internal/storage/localphotos"
	"github.com/solthoth/social-media-marketplace-assistant/services/api/internal/storage/sqlite"
	"github.com/stretchr/testify/suite"
)

type PhotosAPISuite struct {
	suite.Suite
	db     *sql.DB
	router http.Handler
}

func TestPhotosAPISuite(t *testing.T) {
	suite.Run(t, new(PhotosAPISuite))
}

func (s *PhotosAPISuite) SetupTest() {
	db, err := sqlite.Open(context.Background(), filepath.Join(s.T().TempDir(), "integration.db"))
	s.Require().NoError(err)

	itemRepository := sqlite.NewItemRepository(db)
	itemService := items.NewService(itemRepository)
	photoRepository := sqlite.NewPhotoRepository(db)
	photoStorage := localphotos.NewStorage(filepath.Join(s.T().TempDir(), "photos"))
	photoService := photos.NewService(itemRepository, photoRepository, photoStorage)

	s.db = db
	s.router = httpserver.NewRouter(httpserver.RouterDependencies{ItemService: &itemService, PhotoService: &photoService})
}

func (s *PhotosAPISuite) TearDownTest() {
	s.Require().NoError(s.db.Close())
}

func (s *PhotosAPISuite) TestPhotoLifecycleThroughAPI() {
	itemID := s.createItem("Ceramic bowl")

	upload := s.multipartRequest(http.MethodPost, "/items/"+itemID+"/photos", "front.png", pngBytes)
	s.Equal(http.StatusCreated, upload.Code)

	var photo map[string]any
	s.Require().NoError(json.NewDecoder(upload.Body).Decode(&photo))
	photoID := photo["id"].(string)
	s.Equal("front.png", photo["filename"])
	s.Equal("image/png", photo["mime_type"])
	s.Equal(float64(0), photo["sort_order"])
	s.Equal(true, photo["is_primary"])

	list := s.request(http.MethodGet, "/items/"+itemID+"/photos", nil, "application/json")
	s.Equal(http.StatusOK, list.Code)
	var listed listPhotosIntegrationResponse
	s.Require().NoError(json.NewDecoder(list.Body).Decode(&listed))
	s.Len(listed.Photos, 1)
	s.Equal(photoID, listed.Photos[0].ID)
	s.Contains(listed.Photos[0].ContentURLs, "thumbnail")

	content := s.request(http.MethodGet, "/items/"+itemID+"/photos/"+photoID+"/content?variant=medium", nil, "application/json")
	s.Equal(http.StatusOK, content.Code)
	s.Equal("image/png", content.Header().Get("Content-Type"))
	s.Equal(pngBytes, content.Body.Bytes())

	deleteResponse := s.request(http.MethodDelete, "/items/"+itemID+"/photos/"+photoID, nil, "application/json")
	s.Equal(http.StatusNoContent, deleteResponse.Code)

	missingContent := s.request(http.MethodGet, "/items/"+itemID+"/photos/"+photoID+"/content", nil, "application/json")
	s.Equal(http.StatusNotFound, missingContent.Code)
}

func (s *PhotosAPISuite) TestPhotoReorderPersistsThroughAPI() {
	itemID := s.createItem("Leather boots")
	front := s.uploadPhoto(itemID, "front.png")
	back := s.uploadPhoto(itemID, "back.png")

	reorder := s.request(
		http.MethodPatch,
		"/items/"+itemID+"/photos/order",
		bytes.NewBufferString(`{"photo_ids":["`+back.ID+`","`+front.ID+`"]}`),
		"application/json",
	)
	s.Equal(http.StatusOK, reorder.Code)

	var reordered listPhotosIntegrationResponse
	s.Require().NoError(json.NewDecoder(reorder.Body).Decode(&reordered))
	s.Equal([]string{back.ID, front.ID}, integrationPhotoIDs(reordered.Photos))
	s.Equal(0, reordered.Photos[0].SortOrder)
	s.Equal(1, reordered.Photos[1].SortOrder)

	list := s.request(http.MethodGet, "/items/"+itemID+"/photos", nil, "application/json")
	s.Equal(http.StatusOK, list.Code)

	var listed listPhotosIntegrationResponse
	s.Require().NoError(json.NewDecoder(list.Body).Decode(&listed))
	s.Equal([]string{back.ID, front.ID}, integrationPhotoIDs(listed.Photos))
}

func (s *PhotosAPISuite) TestSetPrimaryPhotoPersistsThroughAPI() {
	itemID := s.createItem("Canvas bag")
	front := s.uploadPhoto(itemID, "front.png")
	back := s.uploadPhoto(itemID, "back.png")

	primary := s.request(http.MethodPatch, "/items/"+itemID+"/photos/"+back.ID+"/primary", nil, "application/json")
	s.Equal(http.StatusOK, primary.Code)

	var updated listPhotosIntegrationResponse
	s.Require().NoError(json.NewDecoder(primary.Body).Decode(&updated))
	s.Equal([]string{front.ID, back.ID}, integrationPhotoIDs(updated.Photos))
	s.False(updated.Photos[0].IsPrimary)
	s.True(updated.Photos[1].IsPrimary)

	list := s.request(http.MethodGet, "/items/"+itemID+"/photos", nil, "application/json")
	s.Equal(http.StatusOK, list.Code)

	var listed listPhotosIntegrationResponse
	s.Require().NoError(json.NewDecoder(list.Body).Decode(&listed))
	s.False(listed.Photos[0].IsPrimary)
	s.True(listed.Photos[1].IsPrimary)
}

func (s *PhotosAPISuite) TestPhotoAPIValidationErrors() {
	missingItemUpload := s.multipartRequest(http.MethodPost, "/items/missing/photos", "front.png", pngBytes)
	s.Equal(http.StatusNotFound, missingItemUpload.Code)

	itemID := s.createItem("Invalid upload target")

	invalidUpload := s.multipartRequest(http.MethodPost, "/items/"+itemID+"/photos", "notes.txt", []byte("not an image"))
	s.Equal(http.StatusBadRequest, invalidUpload.Code)

	invalidReorder := s.request(
		http.MethodPatch,
		"/items/"+itemID+"/photos/order",
		bytes.NewBufferString(`{"photo_ids":["missing-photo"]}`),
		"application/json",
	)
	s.Equal(http.StatusBadRequest, invalidReorder.Code)

	invalidVariant := s.request(http.MethodGet, "/items/"+itemID+"/photos/missing/content?variant=poster", nil, "application/json")
	s.Equal(http.StatusBadRequest, invalidVariant.Code)
}

func (s *PhotosAPISuite) request(method string, target string, body io.Reader, contentType string) *httptest.ResponseRecorder {
	request := httptest.NewRequest(method, target, body)
	request.Header.Set("Content-Type", contentType)
	response := httptest.NewRecorder()
	s.router.ServeHTTP(response, request)
	return response
}

func (s *PhotosAPISuite) multipartRequest(method string, target string, filename string, content []byte) *httptest.ResponseRecorder {
	var body bytes.Buffer
	writer := multipart.NewWriter(&body)
	part, err := writer.CreateFormFile("photo", filename)
	s.Require().NoError(err)
	_, err = part.Write(content)
	s.Require().NoError(err)
	s.Require().NoError(writer.Close())
	return s.request(method, target, &body, writer.FormDataContentType())
}

func (s *PhotosAPISuite) createItem(title string) string {
	create := s.request(http.MethodPost, "/items", bytes.NewBufferString(`{"title":"`+title+`"}`), "application/json")
	s.Equal(http.StatusCreated, create.Code)

	var item map[string]any
	s.Require().NoError(json.NewDecoder(create.Body).Decode(&item))
	itemID := item["id"].(string)
	s.NotEmpty(itemID)
	return itemID
}

func (s *PhotosAPISuite) uploadPhoto(itemID string, filename string) photoIntegrationResponse {
	upload := s.multipartRequest(http.MethodPost, "/items/"+itemID+"/photos", filename, pngBytes)
	s.Equal(http.StatusCreated, upload.Code)

	var photo photoIntegrationResponse
	s.Require().NoError(json.NewDecoder(upload.Body).Decode(&photo))
	s.NotEmpty(photo.ID)
	return photo
}

func integrationPhotoIDs(photos []photoIntegrationResponse) []string {
	result := make([]string, 0, len(photos))
	for _, photo := range photos {
		result = append(result, photo.ID)
	}
	return result
}

type listPhotosIntegrationResponse struct {
	Photos []photoIntegrationResponse `json:"photos"`
}

type photoIntegrationResponse struct {
	ID          string            `json:"id"`
	ItemID      string            `json:"item_id"`
	Filename    string            `json:"filename"`
	MimeType    string            `json:"mime_type"`
	SortOrder   int               `json:"sort_order"`
	IsPrimary   bool              `json:"is_primary"`
	ContentURLs map[string]string `json:"content_urls"`
	CreatedAt   string            `json:"created_at"`
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
