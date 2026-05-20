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
	"time"

	"github.com/solthoth/social-media-marketplace-assistant/services/api/internal/ai"
	"github.com/solthoth/social-media-marketplace-assistant/services/api/internal/enrichment"
	"github.com/solthoth/social-media-marketplace-assistant/services/api/internal/httpserver"
	"github.com/solthoth/social-media-marketplace-assistant/services/api/internal/items"
	"github.com/solthoth/social-media-marketplace-assistant/services/api/internal/photos"
	"github.com/solthoth/social-media-marketplace-assistant/services/api/internal/storage/localphotos"
	"github.com/solthoth/social-media-marketplace-assistant/services/api/internal/storage/sqlite"
	"github.com/stretchr/testify/suite"
)

type EnrichmentAPISuite struct {
	suite.Suite
	db                *sql.DB
	router            http.Handler
	enrichmentService enrichment.Service
}

func TestEnrichmentAPISuite(t *testing.T) {
	suite.Run(t, new(EnrichmentAPISuite))
}

func (s *EnrichmentAPISuite) SetupTest() {
	db, err := sqlite.Open(context.Background(), filepath.Join(s.T().TempDir(), "integration.db"))
	s.Require().NoError(err)

	itemRepository := sqlite.NewItemRepository(db)
	itemService := items.NewService(itemRepository)
	photoRepository := sqlite.NewPhotoRepository(db)
	photoStorage := localphotos.NewStorage(filepath.Join(s.T().TempDir(), "photos"))
	photoService := photos.NewService(itemRepository, photoRepository, photoStorage)
	enrichmentRepository := sqlite.NewEnrichmentRepository(db)
	s.enrichmentService = enrichment.NewService(
		itemRepository,
		photoRepository,
		enrichmentRepository,
		ai.FakeProvider{},
		enrichment.ProviderConfig{Provider: "fake", Model: "fake-vision"},
	)

	s.db = db
	s.router = httpserver.NewRouter(httpserver.RouterDependencies{
		ItemService:       &itemService,
		PhotoService:      &photoService,
		EnrichmentService: &s.enrichmentService,
	})
}

func (s *EnrichmentAPISuite) TearDownTest() {
	s.Require().NoError(s.db.Close())
}

func (s *EnrichmentAPISuite) TestEnrichmentLifecycleThroughAPI() {
	itemID := s.createItem("Denim jacket")
	s.uploadPhoto(itemID, "front.png")

	create := s.request(http.MethodPost, "/items/"+itemID+"/enrichment-jobs", nil, "application/json")
	s.Equal(http.StatusCreated, create.Code)
	var created map[string]any
	s.Require().NoError(json.NewDecoder(create.Body).Decode(&created))
	jobID := created["id"].(string)
	s.Equal("queued", created["status"])

	s.Eventually(func() bool {
		processed, err := s.enrichmentService.GetJob(context.Background(), itemID, jobID)
		return err == nil && processed.Status == enrichment.JobStatusCompleted
	}, time.Second, 10*time.Millisecond)

	apply := s.request(http.MethodPost, "/items/"+itemID+"/enrichment-jobs/"+jobID+"/apply", nil, "application/json")
	s.Require().Equal(http.StatusOK, apply.Code)

	var applied map[string]any
	s.Require().NoError(json.NewDecoder(apply.Body).Decode(&applied))
	item := applied["item"].(map[string]any)
	s.Equal("Uncategorized", item["category"])
	s.NotEmpty(item["description"])

	list := s.request(http.MethodGet, "/items/"+itemID+"/enrichment-jobs", nil, "application/json")
	s.Equal(http.StatusOK, list.Code)
	var listed map[string]any
	s.Require().NoError(json.NewDecoder(list.Body).Decode(&listed))
	s.Len(listed["jobs"], 1)
}

func (s *EnrichmentAPISuite) request(method string, target string, body io.Reader, contentType string) *httptest.ResponseRecorder {
	request := httptest.NewRequest(method, target, body)
	request.Header.Set("Content-Type", contentType)
	response := httptest.NewRecorder()
	s.router.ServeHTTP(response, request)
	return response
}

func (s *EnrichmentAPISuite) createItem(title string) string {
	create := s.request(http.MethodPost, "/items", bytes.NewBufferString(`{"title":"`+title+`"}`), "application/json")
	s.Equal(http.StatusCreated, create.Code)

	var item map[string]any
	s.Require().NoError(json.NewDecoder(create.Body).Decode(&item))
	itemID := item["id"].(string)
	s.NotEmpty(itemID)
	return itemID
}

func (s *EnrichmentAPISuite) uploadPhoto(itemID string, filename string) {
	upload := s.multipartRequest(http.MethodPost, "/items/"+itemID+"/photos", filename, pngBytes)
	s.Equal(http.StatusCreated, upload.Code)
}

func (s *EnrichmentAPISuite) multipartRequest(method string, target string, filename string, content []byte) *httptest.ResponseRecorder {
	var body bytes.Buffer
	writer := multipart.NewWriter(&body)
	part, err := writer.CreateFormFile("photo", filename)
	s.Require().NoError(err)
	_, err = part.Write(content)
	s.Require().NoError(err)
	s.Require().NoError(writer.Close())
	return s.request(method, target, &body, writer.FormDataContentType())
}
