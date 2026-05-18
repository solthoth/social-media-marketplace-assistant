//go:build integration

package integration

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"testing"

	"github.com/solthoth/social-media-marketplace-assistant/services/api/internal/httpserver"
	"github.com/solthoth/social-media-marketplace-assistant/services/api/internal/items"
	"github.com/solthoth/social-media-marketplace-assistant/services/api/internal/storage/sqlite"
	"github.com/stretchr/testify/suite"
)

type ItemsAPISuite struct {
	suite.Suite
	db     *sql.DB
	router http.Handler
}

func TestItemsAPISuite(t *testing.T) {
	suite.Run(t, new(ItemsAPISuite))
}

func (s *ItemsAPISuite) SetupTest() {
	db, err := sqlite.Open(context.Background(), filepath.Join(s.T().TempDir(), "integration.db"))
	s.Require().NoError(err)

	repository := sqlite.NewItemRepository(db)
	service := items.NewService(repository)

	s.db = db
	s.router = httpserver.NewRouter(httpserver.RouterDependencies{ItemService: &service})
}

func (s *ItemsAPISuite) TearDownTest() {
	s.Require().NoError(s.db.Close())
}

func (s *ItemsAPISuite) TestItemLifecyclePersistsThroughAPI() {
	create := s.request(http.MethodPost, "/items", bytes.NewBufferString(`{
		"title": "Ceramic bowl",
		"category": "Kitchen",
		"original_purchase_price_cents": 700,
		"selling_price_cents": 1800
	}`))
	s.Equal(http.StatusCreated, create.Code)

	var created map[string]any
	s.Require().NoError(json.NewDecoder(create.Body).Decode(&created))
	id := created["id"].(string)
	s.NotEmpty(id)

	update := s.request(http.MethodPatch, "/items/"+id, bytes.NewBufferString(`{
		"description": "Blue handmade bowl",
		"selling_price_cents": 2200,
		"status": "ready_to_list"
	}`))
	s.Equal(http.StatusOK, update.Code)

	var updated map[string]any
	s.Require().NoError(json.NewDecoder(update.Body).Decode(&updated))
	s.Equal(float64(700), updated["original_purchase_price_cents"])
	s.Equal(float64(2200), updated["selling_price_cents"])

	deleteResponse := s.request(http.MethodDelete, "/items/"+id, nil)
	s.Equal(http.StatusNoContent, deleteResponse.Code)

	get := s.request(http.MethodGet, "/items/"+id, nil)
	s.Equal(http.StatusOK, get.Code)

	var archived map[string]any
	s.Require().NoError(json.NewDecoder(get.Body).Decode(&archived))
	s.Equal("archived", archived["status"])
}

func (s *ItemsAPISuite) request(method string, target string, body *bytes.Buffer) *httptest.ResponseRecorder {
	var requestBody io.Reader
	if body != nil {
		requestBody = body
	}
	request := httptest.NewRequest(method, target, requestBody)
	request.Header.Set("Content-Type", "application/json")
	response := httptest.NewRecorder()
	s.router.ServeHTTP(response, request)
	return response
}
