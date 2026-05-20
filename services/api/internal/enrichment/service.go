package enrichment

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/solthoth/social-media-marketplace-assistant/services/api/internal/domain"
)

var (
	ErrItemNotReady   = errors.New("item not ready for enrichment")
	ErrJobNotFound    = errors.New("enrichment job not found")
	ErrJobNotComplete = errors.New("enrichment job is not complete")
)

type JobStatus string

const (
	JobStatusQueued     JobStatus = "queued"
	JobStatusProcessing JobStatus = "processing"
	JobStatusCompleted  JobStatus = "completed"
	JobStatusFailed     JobStatus = "failed"
)

type ItemRepository interface {
	Get(ctx context.Context, id string) (domain.Item, error)
	Update(ctx context.Context, item domain.Item) (domain.Item, error)
}

type PhotoRepository interface {
	ListPhotos(ctx context.Context, itemID string) ([]domain.ItemPhoto, error)
}

type JobRepository interface {
	CreateJob(ctx context.Context, job Job) (Job, error)
	ListJobs(ctx context.Context, itemID string) ([]Job, error)
	GetJob(ctx context.Context, itemID string, jobID string) (Job, error)
	GetJobByID(ctx context.Context, jobID string) (Job, error)
	UpdateJob(ctx context.Context, job Job) (Job, error)
}

type ItemDetailProvider interface {
	GenerateItemDetails(ctx context.Context, input ItemDetailInput) (ItemDetailSuggestion, error)
}

type ProviderConfig struct {
	Provider string
	Model    string
}

type Service struct {
	items    ItemRepository
	photos   PhotoRepository
	jobs     JobRepository
	provider ItemDetailProvider
	config   ProviderConfig
}

func NewService(items ItemRepository, photos PhotoRepository, jobs JobRepository, provider ItemDetailProvider, config ProviderConfig) Service {
	return Service{
		items:    items,
		photos:   photos,
		jobs:     jobs,
		provider: provider,
		config:   config,
	}
}

type Job struct {
	ID            string
	ItemID        string
	Status        JobStatus
	Provider      string
	Model         string
	RequestedAt   time.Time
	StartedAt     *time.Time
	CompletedAt   *time.Time
	AppliedAt     *time.Time
	ErrorMessage  string
	InputSnapshot ItemDetailInput
	Suggestion    ItemDetailSuggestion
}

type ItemDetailInput struct {
	ItemID              string           `json:"item_id"`
	Title               string           `json:"title"`
	ExistingDescription string           `json:"existing_description"`
	ExistingCategory    string           `json:"existing_category"`
	ExistingSize        string           `json:"existing_size"`
	ExistingCondition   string           `json:"existing_condition"`
	ExistingNotes       string           `json:"existing_notes"`
	Photos              []ItemPhotoInput `json:"photos"`
}

type ItemPhotoInput struct {
	ID       string `json:"id"`
	Filename string `json:"filename"`
	MimeType string `json:"mime_type"`
}

type ItemDetailSuggestion struct {
	Description string `json:"description"`
	Category    string `json:"category"`
	Size        string `json:"size"`
	Condition   string `json:"condition"`
	Notes       string `json:"notes"`
}

type ApplyResult struct {
	Item          domain.Item
	Job           Job
	AppliedFields []string
}

func (s Service) CreateJob(ctx context.Context, itemID string) (Job, error) {
	input, err := s.itemDetailInput(ctx, itemID)
	if err != nil {
		return Job{}, err
	}
	if strings.TrimSpace(input.Title) == "" || len(input.Photos) == 0 {
		return Job{}, ErrItemNotReady
	}

	now := time.Now().UTC()
	job := Job{
		ID:            uuid.NewString(),
		ItemID:        itemID,
		Status:        JobStatusQueued,
		Provider:      strings.TrimSpace(s.config.Provider),
		Model:         strings.TrimSpace(s.config.Model),
		RequestedAt:   now,
		InputSnapshot: input,
	}
	return s.jobs.CreateJob(ctx, job)
}

func (s Service) ListJobs(ctx context.Context, itemID string) ([]Job, error) {
	if _, err := s.items.Get(ctx, itemID); err != nil {
		return nil, err
	}
	return s.jobs.ListJobs(ctx, itemID)
}

func (s Service) GetJob(ctx context.Context, itemID string, jobID string) (Job, error) {
	if _, err := s.items.Get(ctx, itemID); err != nil {
		return Job{}, err
	}
	return s.jobs.GetJob(ctx, itemID, jobID)
}

func (s Service) ProcessJob(ctx context.Context, jobID string) (Job, error) {
	job, err := s.jobs.GetJobByID(ctx, jobID)
	if err != nil {
		return Job{}, err
	}
	if job.Status == JobStatusCompleted {
		return job, nil
	}

	startedAt := time.Now().UTC()
	job.Status = JobStatusProcessing
	job.StartedAt = &startedAt
	if job, err = s.jobs.UpdateJob(ctx, job); err != nil {
		return Job{}, err
	}

	suggestion, err := s.provider.GenerateItemDetails(ctx, job.InputSnapshot)
	completedAt := time.Now().UTC()
	job.CompletedAt = &completedAt
	if err != nil {
		job.Status = JobStatusFailed
		job.ErrorMessage = err.Error()
		updated, updateErr := s.jobs.UpdateJob(ctx, job)
		if updateErr != nil {
			return Job{}, updateErr
		}
		return updated, err
	}
	job.Status = JobStatusCompleted
	job.Suggestion = trimSuggestion(suggestion)
	job.ErrorMessage = ""
	return s.jobs.UpdateJob(ctx, job)
}

func (s Service) ApplyJob(ctx context.Context, itemID string, jobID string) (ApplyResult, error) {
	job, err := s.jobs.GetJob(ctx, itemID, jobID)
	if err != nil {
		return ApplyResult{}, err
	}
	if job.Status != JobStatusCompleted {
		return ApplyResult{}, ErrJobNotComplete
	}

	item, err := s.items.Get(ctx, itemID)
	if err != nil {
		return ApplyResult{}, err
	}
	applied := applySuggestionToEmptyFields(&item, job.Suggestion)
	item.UpdatedAt = time.Now().UTC()

	updatedItem, err := s.items.Update(ctx, item)
	if err != nil {
		return ApplyResult{}, err
	}
	now := time.Now().UTC()
	job.AppliedAt = &now
	updatedJob, err := s.jobs.UpdateJob(ctx, job)
	if err != nil {
		return ApplyResult{}, err
	}
	return ApplyResult{Item: updatedItem, Job: updatedJob, AppliedFields: applied}, nil
}

func (s Service) itemDetailInput(ctx context.Context, itemID string) (ItemDetailInput, error) {
	item, err := s.items.Get(ctx, itemID)
	if err != nil {
		return ItemDetailInput{}, err
	}
	itemPhotos, err := s.photos.ListPhotos(ctx, itemID)
	if err != nil {
		return ItemDetailInput{}, err
	}
	photoInputs := make([]ItemPhotoInput, 0, len(itemPhotos))
	for _, photo := range itemPhotos {
		photoInputs = append(photoInputs, ItemPhotoInput{
			ID:       photo.ID,
			Filename: photo.Filename,
			MimeType: photo.MimeType,
		})
	}
	return ItemDetailInput{
		ItemID:              item.ID,
		Title:               item.Title,
		ExistingDescription: item.Description,
		ExistingCategory:    item.Category,
		ExistingSize:        item.Size,
		ExistingCondition:   item.Condition,
		ExistingNotes:       item.Notes,
		Photos:              photoInputs,
	}, nil
}

func applySuggestionToEmptyFields(item *domain.Item, suggestion ItemDetailSuggestion) []string {
	applied := []string{}
	if strings.TrimSpace(item.Description) == "" && strings.TrimSpace(suggestion.Description) != "" {
		item.Description = strings.TrimSpace(suggestion.Description)
		applied = append(applied, "description")
	}
	if strings.TrimSpace(item.Category) == "" && strings.TrimSpace(suggestion.Category) != "" {
		item.Category = strings.TrimSpace(suggestion.Category)
		applied = append(applied, "category")
	}
	if strings.TrimSpace(item.Size) == "" && strings.TrimSpace(suggestion.Size) != "" {
		item.Size = strings.TrimSpace(suggestion.Size)
		applied = append(applied, "size")
	}
	if strings.TrimSpace(item.Condition) == "" && strings.TrimSpace(suggestion.Condition) != "" {
		item.Condition = strings.TrimSpace(suggestion.Condition)
		applied = append(applied, "condition")
	}
	if strings.TrimSpace(item.Notes) == "" && strings.TrimSpace(suggestion.Notes) != "" {
		item.Notes = strings.TrimSpace(suggestion.Notes)
		applied = append(applied, "notes")
	}
	return applied
}

func trimSuggestion(suggestion ItemDetailSuggestion) ItemDetailSuggestion {
	return ItemDetailSuggestion{
		Description: strings.TrimSpace(suggestion.Description),
		Category:    strings.TrimSpace(suggestion.Category),
		Size:        strings.TrimSpace(suggestion.Size),
		Condition:   strings.TrimSpace(suggestion.Condition),
		Notes:       strings.TrimSpace(suggestion.Notes),
	}
}
