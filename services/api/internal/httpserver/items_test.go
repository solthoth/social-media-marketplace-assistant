package httpserver

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"slices"
	"sync"
	"testing"

	"github.com/solthoth/social-media-marketplace-assistant/services/api/internal/domain"
	"github.com/solthoth/social-media-marketplace-assistant/services/api/internal/items"
	"github.com/stretchr/testify/suite"
)

type ItemsHandlerSuite struct {
	suite.Suite
	service *items.Service
	router  http.Handler
}

func TestItemsHandlerSuite(t *testing.T) {
	suite.Run(t, new(ItemsHandlerSuite))
}

func (s *ItemsHandlerSuite) SetupTest() {
	repository := newHandlerMemoryRepository()
	service := items.NewService(repository)
	s.service = &service
	s.router = NewRouter(RouterDependencies{ItemService: s.service})
}

func (s *ItemsHandlerSuite) TestCreateListGetPatchAndArchiveItem() {
	createBody := bytes.NewBufferString(`{
		"title": "Denim jacket",
		"description": "Medium wash",
		"category": "Clothing",
		"size": "M",
		"condition": "Good",
		"original_purchase_price_cents": 1800,
		"selling_price_cents": 3200,
		"currency": "USD",
		"notes": "Steam before photos"
	}`)

	createResponse := s.request(http.MethodPost, "/items", createBody)
	s.Equal(http.StatusCreated, createResponse.Code)

	var created itemResponse
	s.Require().NoError(json.NewDecoder(createResponse.Body).Decode(&created))
	s.NotEmpty(created.ID)
	s.Equal("Denim jacket", created.Title)
	s.Equal(int64(1800), created.OriginalPurchasePriceCents)
	s.Equal(int64(3200), created.SellingPriceCents)
	s.Equal("draft", created.Status)

	listResponse := s.request(http.MethodGet, "/items", nil)
	s.Equal(http.StatusOK, listResponse.Code)

	var list listItemsResponse
	s.Require().NoError(json.NewDecoder(listResponse.Body).Decode(&list))
	s.Len(list.Items, 1)

	getResponse := s.request(http.MethodGet, "/items/"+created.ID, nil)
	s.Equal(http.StatusOK, getResponse.Code)

	patchResponse := s.request(
		http.MethodPatch,
		"/items/"+created.ID,
		bytes.NewBufferString(`{"title":"Listed denim jacket","selling_price_cents":3600,"status":"ready_to_list"}`),
	)
	s.Equal(http.StatusOK, patchResponse.Code)

	var updated itemResponse
	s.Require().NoError(json.NewDecoder(patchResponse.Body).Decode(&updated))
	s.Equal("Listed denim jacket", updated.Title)
	s.Equal(int64(3600), updated.SellingPriceCents)
	s.Equal("ready_to_list", updated.Status)

	deleteResponse := s.request(http.MethodDelete, "/items/"+created.ID, nil)
	s.Equal(http.StatusNoContent, deleteResponse.Code)

	archivedResponse := s.request(http.MethodGet, "/items/"+created.ID, nil)
	s.Equal(http.StatusOK, archivedResponse.Code)

	var archived itemResponse
	s.Require().NoError(json.NewDecoder(archivedResponse.Body).Decode(&archived))
	s.Equal("archived", archived.Status)
}

func (s *ItemsHandlerSuite) TestCreateItemValidationError() {
	response := s.request(http.MethodPost, "/items", bytes.NewBufferString(`{"price_cents":100}`))

	s.Equal(http.StatusBadRequest, response.Code)

	var body errorResponse
	s.Require().NoError(json.NewDecoder(response.Body).Decode(&body))
	s.Equal("invalid_item", body.Error.Code)
}

func (s *ItemsHandlerSuite) TestPatchInvalidStatusTransitionReturnsBadRequest() {
	createResponse := s.request(
		http.MethodPost,
		"/items",
		bytes.NewBufferString(`{"title":"Denim jacket"}`),
	)
	s.Equal(http.StatusCreated, createResponse.Code)

	var created itemResponse
	s.Require().NoError(json.NewDecoder(createResponse.Body).Decode(&created))

	patchResponse := s.request(
		http.MethodPatch,
		"/items/"+created.ID,
		bytes.NewBufferString(`{"status":"listed"}`),
	)

	s.Equal(http.StatusBadRequest, patchResponse.Code)

	var body errorResponse
	s.Require().NoError(json.NewDecoder(patchResponse.Body).Decode(&body))
	s.Equal("invalid_status_transition", body.Error.Code)
}

func (s *ItemsHandlerSuite) TestGetMissingItemReturnsNotFound() {
	response := s.request(http.MethodGet, "/items/missing", nil)

	s.Equal(http.StatusNotFound, response.Code)
}

func (s *ItemsHandlerSuite) request(method string, target string, body *bytes.Buffer) *httptest.ResponseRecorder {
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

type handlerMemoryRepository struct {
	mu    sync.RWMutex
	items map[string]domain.Item
}

func newHandlerMemoryRepository() *handlerMemoryRepository {
	return &handlerMemoryRepository{
		items: map[string]domain.Item{},
	}
}

var _ items.Repository = (*handlerMemoryRepository)(nil)

func (r *handlerMemoryRepository) Create(ctx context.Context, item domain.Item) (domain.Item, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.items[item.ID] = item
	return item, nil
}

func (r *handlerMemoryRepository) List(ctx context.Context, filter items.ListItemsFilter) ([]domain.Item, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	result := make([]domain.Item, 0, len(r.items))
	for _, item := range r.items {
		if filter.Status != nil && item.Status != *filter.Status {
			continue
		}
		result = append(result, item)
	}
	slices.SortFunc(result, func(a domain.Item, b domain.Item) int {
		return a.CreatedAt.Compare(b.CreatedAt)
	})
	return result, nil
}

func (r *handlerMemoryRepository) Get(ctx context.Context, id string) (domain.Item, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	item, ok := r.items[id]
	if !ok {
		return domain.Item{}, items.ErrItemNotFound
	}
	return item, nil
}

func (r *handlerMemoryRepository) Update(ctx context.Context, item domain.Item) (domain.Item, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, ok := r.items[item.ID]; !ok {
		return domain.Item{}, items.ErrItemNotFound
	}
	r.items[item.ID] = item
	return item, nil
}
