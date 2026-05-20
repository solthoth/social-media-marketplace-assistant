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
	"github.com/solthoth/social-media-marketplace-assistant/services/api/internal/enrichment"
	"github.com/solthoth/social-media-marketplace-assistant/services/api/internal/items"
	"github.com/solthoth/social-media-marketplace-assistant/services/api/internal/photos"
	"github.com/stretchr/testify/suite"
)

type EnrichmentHandlerSuite struct {
	suite.Suite
	router        http.Handler
	itemRepo      *handlerMemoryRepository
	photoRepo     *handlerPhotoRepository
	enrichService enrichment.Service
}

func TestEnrichmentHandlerSuite(t *testing.T) {
	suite.Run(t, new(EnrichmentHandlerSuite))
}

func (s *EnrichmentHandlerSuite) SetupTest() {
	s.itemRepo = newHandlerMemoryRepository()
	s.photoRepo = newHandlerPhotoRepository()
	itemService := items.NewService(s.itemRepo)
	photoService := photos.NewService(s.itemRepo, s.photoRepo, newHandlerPhotoStorage())
	enrichmentRepo := newHandlerEnrichmentRepository()
	s.enrichService = enrichment.NewService(
		s.itemRepo,
		s.photoRepo,
		enrichmentRepo,
		handlerEnrichmentProvider{},
		enrichment.ProviderConfig{Provider: "fake", Model: "fake-vision"},
	)
	s.router = NewRouter(RouterDependencies{
		ItemService:       &itemService,
		PhotoService:      &photoService,
		EnrichmentService: &s.enrichService,
	})
}

func (s *EnrichmentHandlerSuite) TestCreateListGetAndApplyEnrichmentJob() {
	item := s.createItem("Denim jacket")
	photo := domain.ItemPhoto{
		ID: "photo-1", ItemID: item.ID, StorageID: "items/" + item.ID + "/photos/photo-1",
		Filename: "front.png", MimeType: "image/png", IsPrimary: true,
	}
	_, err := s.photoRepo.CreatePhoto(context.Background(), photo)
	s.Require().NoError(err)

	create := s.request(http.MethodPost, "/items/"+item.ID+"/enrichment-jobs", nil)
	s.Equal(http.StatusCreated, create.Code)
	var created enrichmentJobResponse
	s.Require().NoError(json.NewDecoder(create.Body).Decode(&created))
	s.NotEmpty(created.ID)
	s.Equal("queued", created.Status)

	processed, err := s.enrichService.ProcessJob(context.Background(), created.ID)
	s.Require().NoError(err)
	s.Equal(enrichment.JobStatusCompleted, processed.Status)

	list := s.request(http.MethodGet, "/items/"+item.ID+"/enrichment-jobs", nil)
	s.Equal(http.StatusOK, list.Code)
	var listed listEnrichmentJobsResponse
	s.Require().NoError(json.NewDecoder(list.Body).Decode(&listed))
	s.Len(listed.Jobs, 1)

	get := s.request(http.MethodGet, "/items/"+item.ID+"/enrichment-jobs/"+created.ID, nil)
	s.Equal(http.StatusOK, get.Code)
	var fetched enrichmentJobResponse
	s.Require().NoError(json.NewDecoder(get.Body).Decode(&fetched))
	s.Equal("Clothing", fetched.Suggestion.Category)

	apply := s.request(http.MethodPost, "/items/"+item.ID+"/enrichment-jobs/"+created.ID+"/apply", nil)
	s.Equal(http.StatusOK, apply.Code)
	var applied applyEnrichmentJobResponse
	s.Require().NoError(json.NewDecoder(apply.Body).Decode(&applied))
	s.Equal("Clothing", applied.Item.Category)
	s.Contains(applied.AppliedFields, "description")
}

func (s *EnrichmentHandlerSuite) TestCreateRejectsItemWithoutPhotos() {
	item := s.createItem("Denim jacket")

	response := s.request(http.MethodPost, "/items/"+item.ID+"/enrichment-jobs", nil)

	s.Equal(http.StatusBadRequest, response.Code)
}

func (s *EnrichmentHandlerSuite) createItem(title string) itemResponse {
	response := s.request(http.MethodPost, "/items", bytes.NewBufferString(`{"title":"`+title+`"}`))
	s.Equal(http.StatusCreated, response.Code)
	var item itemResponse
	s.Require().NoError(json.NewDecoder(response.Body).Decode(&item))
	return item
}

func (s *EnrichmentHandlerSuite) request(method string, target string, body *bytes.Buffer) *httptest.ResponseRecorder {
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

type handlerEnrichmentProvider struct{}

func (p handlerEnrichmentProvider) GenerateItemDetails(ctx context.Context, input enrichment.ItemDetailInput) (enrichment.ItemDetailSuggestion, error) {
	return enrichment.ItemDetailSuggestion{
		Description: "A denim jacket photographed from the front.",
		Category:    "Clothing",
		Size:        "M",
		Condition:   "Good",
		Notes:       "Confirm measurements before listing.",
	}, nil
}

type handlerEnrichmentRepository struct {
	mu   sync.RWMutex
	jobs map[string]enrichment.Job
}

func newHandlerEnrichmentRepository() *handlerEnrichmentRepository {
	return &handlerEnrichmentRepository{jobs: map[string]enrichment.Job{}}
}

func (r *handlerEnrichmentRepository) CreateJob(ctx context.Context, job enrichment.Job) (enrichment.Job, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.jobs[job.ID] = job
	return job, nil
}

func (r *handlerEnrichmentRepository) ListJobs(ctx context.Context, itemID string) ([]enrichment.Job, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	result := []enrichment.Job{}
	for _, job := range r.jobs {
		if job.ItemID == itemID {
			result = append(result, job)
		}
	}
	slices.SortFunc(result, func(a enrichment.Job, b enrichment.Job) int {
		return b.RequestedAt.Compare(a.RequestedAt)
	})
	return result, nil
}

func (r *handlerEnrichmentRepository) GetJob(ctx context.Context, itemID string, jobID string) (enrichment.Job, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	job, ok := r.jobs[jobID]
	if !ok || job.ItemID != itemID {
		return enrichment.Job{}, enrichment.ErrJobNotFound
	}
	return job, nil
}

func (r *handlerEnrichmentRepository) GetJobByID(ctx context.Context, jobID string) (enrichment.Job, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	job, ok := r.jobs[jobID]
	if !ok {
		return enrichment.Job{}, enrichment.ErrJobNotFound
	}
	return job, nil
}

func (r *handlerEnrichmentRepository) UpdateJob(ctx context.Context, job enrichment.Job) (enrichment.Job, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, ok := r.jobs[job.ID]; !ok {
		return enrichment.Job{}, enrichment.ErrJobNotFound
	}
	r.jobs[job.ID] = job
	return job, nil
}
