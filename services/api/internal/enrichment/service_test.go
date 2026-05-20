package enrichment

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/solthoth/social-media-marketplace-assistant/services/api/internal/domain"
	"github.com/solthoth/social-media-marketplace-assistant/services/api/internal/items"
	"github.com/solthoth/social-media-marketplace-assistant/services/api/internal/photos"
	"github.com/stretchr/testify/suite"
)

type ServiceSuite struct {
	suite.Suite
	items      *memoryItemRepository
	photos     *memoryPhotoRepository
	jobs       *memoryJobRepository
	provider   *fakeProvider
	service    Service
	existingID string
}

func TestServiceSuite(t *testing.T) {
	suite.Run(t, new(ServiceSuite))
}

func (s *ServiceSuite) SetupTest() {
	s.items = newMemoryItemRepository()
	s.photos = newMemoryPhotoRepository()
	s.jobs = newMemoryJobRepository()
	s.provider = &fakeProvider{
		suggestion: ItemDetailSuggestion{
			Description: "A soft denim jacket with visible button closures.",
			Category:    "Clothing",
			Size:        "M",
			Condition:   "Good",
			Notes:       "Review sleeve cuffs before listing.",
		},
	}
	s.service = NewService(s.items, s.photos, s.jobs, s.provider, ProviderConfig{
		Provider: "fake",
		Model:    "fake-vision",
	})
	s.existingID = "item-1"
	s.items.items[s.existingID] = domain.Item{
		ID:           s.existingID,
		Title:        "Denim jacket",
		SellingPrice: domain.Money{Currency: string(domain.CurrencyUSD)},
		Status:       domain.InventoryStatusDraft,
		CreatedAt:    time.Now().UTC(),
		UpdatedAt:    time.Now().UTC(),
	}
	s.photos.photos[s.existingID] = []domain.ItemPhoto{
		{ID: "photo-1", ItemID: s.existingID, StorageID: "items/item-1/photos/photo-1", Filename: "front.png", MimeType: "image/png"},
	}
}

func (s *ServiceSuite) TestCreateJobRequiresTitleAndPhotos() {
	s.items.items["missing-title"] = domain.Item{ID: "missing-title"}
	_, err := s.service.CreateJob(context.Background(), "missing-title")
	s.ErrorIs(err, ErrItemNotReady)

	s.items.items["missing-photos"] = domain.Item{ID: "missing-photos", Title: "Scarf"}
	_, err = s.service.CreateJob(context.Background(), "missing-photos")
	s.ErrorIs(err, ErrItemNotReady)
}

func (s *ServiceSuite) TestCreateJobStoresQueuedJobWithSnapshot() {
	job, err := s.service.CreateJob(context.Background(), s.existingID)

	s.Require().NoError(err)
	s.NotEmpty(job.ID)
	s.Equal(s.existingID, job.ItemID)
	s.Equal(JobStatusQueued, job.Status)
	s.Equal("fake", job.Provider)
	s.Equal("fake-vision", job.Model)
	s.Equal("Denim jacket", job.InputSnapshot.Title)
	s.Len(job.InputSnapshot.Photos, 1)
}

func (s *ServiceSuite) TestProcessJobCompletesWithProviderSuggestion() {
	job, err := s.service.CreateJob(context.Background(), s.existingID)
	s.Require().NoError(err)

	processed, err := s.service.ProcessJob(context.Background(), job.ID)

	s.Require().NoError(err)
	s.Equal(JobStatusCompleted, processed.Status)
	s.NotNil(processed.StartedAt)
	s.NotNil(processed.CompletedAt)
	s.Equal("Clothing", processed.Suggestion.Category)
	s.Equal(s.existingID, s.provider.lastInput.ItemID)
	s.Len(s.provider.lastInput.Photos, 1)
}

func (s *ServiceSuite) TestCreateJobAddsPhotoDataURLsWhenStorageIsConfigured() {
	storage := &memoryPhotoStorage{content: map[string][]byte{
		"items/item-1/photos/photo-1": pngBytes,
	}}
	s.service = NewServiceWithPhotoContent(s.items, s.photos, s.jobs, s.provider, storage, ProviderConfig{
		Provider: "fake",
		Model:    "fake-vision",
	})

	job, err := s.service.CreateJob(context.Background(), s.existingID)

	s.Require().NoError(err)
	s.Equal("data:image/png;base64,iVBORw0KGgo=", job.InputSnapshot.Photos[0].DataURL)
}

func (s *ServiceSuite) TestProcessJobStoresFailedStateWhenProviderFails() {
	s.provider.err = errors.New("provider unavailable")
	job, err := s.service.CreateJob(context.Background(), s.existingID)
	s.Require().NoError(err)

	processed, err := s.service.ProcessJob(context.Background(), job.ID)

	s.Error(err)
	s.Equal(JobStatusFailed, processed.Status)
	s.Contains(processed.ErrorMessage, "provider unavailable")
}

func (s *ServiceSuite) TestApplyJobFillsEmptyFieldsOnly() {
	item := s.items.items[s.existingID]
	item.Description = "Seller-written description"
	s.items.items[s.existingID] = item
	job, err := s.service.CreateJob(context.Background(), s.existingID)
	s.Require().NoError(err)
	processed, err := s.service.ProcessJob(context.Background(), job.ID)
	s.Require().NoError(err)

	result, err := s.service.ApplyJob(context.Background(), s.existingID, processed.ID)

	s.Require().NoError(err)
	s.Equal("Seller-written description", result.Item.Description)
	s.Equal("Clothing", result.Item.Category)
	s.Equal("M", result.Item.Size)
	s.Equal("Good", result.Item.Condition)
	s.Equal("Review sleeve cuffs before listing.", result.Item.Notes)
	s.Equal([]string{"category", "size", "condition", "notes"}, result.AppliedFields)
	s.NotNil(result.Job.AppliedAt)
}

func (s *ServiceSuite) TestApplyRejectsIncompleteJob() {
	job, err := s.service.CreateJob(context.Background(), s.existingID)
	s.Require().NoError(err)

	_, err = s.service.ApplyJob(context.Background(), s.existingID, job.ID)

	s.ErrorIs(err, ErrJobNotComplete)
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

func (r *memoryItemRepository) Update(ctx context.Context, item domain.Item) (domain.Item, error) {
	r.items[item.ID] = item
	return item, nil
}

type memoryPhotoRepository struct {
	photos map[string][]domain.ItemPhoto
}

func newMemoryPhotoRepository() *memoryPhotoRepository {
	return &memoryPhotoRepository{photos: map[string][]domain.ItemPhoto{}}
}

func (r *memoryPhotoRepository) ListPhotos(ctx context.Context, itemID string) ([]domain.ItemPhoto, error) {
	return r.photos[itemID], nil
}

type memoryPhotoStorage struct {
	content map[string][]byte
}

func (s *memoryPhotoStorage) OpenPhotoContent(ctx context.Context, photo domain.ItemPhoto) ([]byte, error) {
	return s.content[photo.StorageID], nil
}

var pngBytes = []byte{0x89, 0x50, 0x4e, 0x47, 0x0d, 0x0a, 0x1a, 0x0a}

type memoryJobRepository struct {
	jobs map[string]Job
}

func newMemoryJobRepository() *memoryJobRepository {
	return &memoryJobRepository{jobs: map[string]Job{}}
}

func (r *memoryJobRepository) CreateJob(ctx context.Context, job Job) (Job, error) {
	r.jobs[job.ID] = job
	return job, nil
}

func (r *memoryJobRepository) ListJobs(ctx context.Context, itemID string) ([]Job, error) {
	result := []Job{}
	for _, job := range r.jobs {
		if job.ItemID == itemID {
			result = append(result, job)
		}
	}
	return result, nil
}

func (r *memoryJobRepository) GetJob(ctx context.Context, itemID string, jobID string) (Job, error) {
	job, ok := r.jobs[jobID]
	if !ok || job.ItemID != itemID {
		return Job{}, ErrJobNotFound
	}
	return job, nil
}

func (r *memoryJobRepository) GetJobByID(ctx context.Context, jobID string) (Job, error) {
	job, ok := r.jobs[jobID]
	if !ok {
		return Job{}, ErrJobNotFound
	}
	return job, nil
}

func (r *memoryJobRepository) UpdateJob(ctx context.Context, job Job) (Job, error) {
	if _, ok := r.jobs[job.ID]; !ok {
		return Job{}, ErrJobNotFound
	}
	r.jobs[job.ID] = job
	return job, nil
}

type fakeProvider struct {
	suggestion ItemDetailSuggestion
	err        error
	lastInput  ItemDetailInput
}

func (p *fakeProvider) GenerateItemDetails(ctx context.Context, input ItemDetailInput) (ItemDetailSuggestion, error) {
	p.lastInput = input
	if p.err != nil {
		return ItemDetailSuggestion{}, p.err
	}
	return p.suggestion, nil
}

var _ = photos.ErrPhotoNotFound
