package httpserver

import (
	"context"
	"errors"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/solthoth/social-media-marketplace-assistant/services/api/internal/enrichment"
	"github.com/solthoth/social-media-marketplace-assistant/services/api/internal/items"
)

type enrichmentApplication interface {
	CreateJob(ctx context.Context, itemID string) (enrichment.Job, error)
	ListJobs(ctx context.Context, itemID string) ([]enrichment.Job, error)
	GetJob(ctx context.Context, itemID string, jobID string) (enrichment.Job, error)
	ProcessJob(ctx context.Context, jobID string) (enrichment.Job, error)
	ApplyJob(ctx context.Context, itemID string, jobID string) (enrichment.ApplyResult, error)
}

type enrichmentRoutes struct {
	service enrichmentApplication
}

type enrichmentJobResponse struct {
	ID            string                          `json:"id"`
	ItemID        string                          `json:"item_id"`
	Status        string                          `json:"status"`
	Provider      string                          `json:"provider"`
	Model         string                          `json:"model"`
	RequestedAt   string                          `json:"requested_at"`
	StartedAt     *string                         `json:"started_at"`
	CompletedAt   *string                         `json:"completed_at"`
	AppliedAt     *string                         `json:"applied_at"`
	ErrorMessage  string                          `json:"error_message"`
	InputSnapshot enrichment.ItemDetailInput      `json:"input_snapshot"`
	Suggestion    enrichment.ItemDetailSuggestion `json:"suggestion"`
}

type listEnrichmentJobsResponse struct {
	Jobs []enrichmentJobResponse `json:"jobs"`
}

type applyEnrichmentJobResponse struct {
	Item          itemResponse          `json:"item"`
	Job           enrichmentJobResponse `json:"job"`
	AppliedFields []string              `json:"applied_fields"`
}

func registerEnrichmentRoutes(router *gin.Engine, service enrichmentApplication) {
	routes := enrichmentRoutes{service: service}
	router.POST("/items/:id/enrichment-jobs", routes.create)
	router.GET("/items/:id/enrichment-jobs", routes.list)
	router.GET("/items/:id/enrichment-jobs/:jobID", routes.get)
	router.POST("/items/:id/enrichment-jobs/:jobID/apply", routes.apply)
}

func (r enrichmentRoutes) create(c *gin.Context) {
	job, err := r.service.CreateJob(c.Request.Context(), c.Param("id"))
	if err != nil {
		writeEnrichmentServiceError(c, err)
		return
	}
	go func() {
		_, _ = r.service.ProcessJob(context.Background(), job.ID)
	}()
	c.JSON(http.StatusCreated, newEnrichmentJobResponse(job))
}

func (r enrichmentRoutes) list(c *gin.Context) {
	jobs, err := r.service.ListJobs(c.Request.Context(), c.Param("id"))
	if err != nil {
		writeEnrichmentServiceError(c, err)
		return
	}
	response := listEnrichmentJobsResponse{Jobs: make([]enrichmentJobResponse, 0, len(jobs))}
	for _, job := range jobs {
		response.Jobs = append(response.Jobs, newEnrichmentJobResponse(job))
	}
	c.JSON(http.StatusOK, response)
}

func (r enrichmentRoutes) get(c *gin.Context) {
	job, err := r.service.GetJob(c.Request.Context(), c.Param("id"), c.Param("jobID"))
	if err != nil {
		writeEnrichmentServiceError(c, err)
		return
	}
	c.JSON(http.StatusOK, newEnrichmentJobResponse(job))
}

func (r enrichmentRoutes) apply(c *gin.Context) {
	result, err := r.service.ApplyJob(c.Request.Context(), c.Param("id"), c.Param("jobID"))
	if err != nil {
		writeEnrichmentServiceError(c, err)
		return
	}
	c.JSON(http.StatusOK, applyEnrichmentJobResponse{
		Item:          newItemResponse(result.Item),
		Job:           newEnrichmentJobResponse(result.Job),
		AppliedFields: result.AppliedFields,
	})
}

func newEnrichmentJobResponse(job enrichment.Job) enrichmentJobResponse {
	return enrichmentJobResponse{
		ID:            job.ID,
		ItemID:        job.ItemID,
		Status:        string(job.Status),
		Provider:      job.Provider,
		Model:         job.Model,
		RequestedAt:   job.RequestedAt.Format(timeFormat),
		StartedAt:     formatOptionalResponseTime(job.StartedAt),
		CompletedAt:   formatOptionalResponseTime(job.CompletedAt),
		AppliedAt:     formatOptionalResponseTime(job.AppliedAt),
		ErrorMessage:  job.ErrorMessage,
		InputSnapshot: job.InputSnapshot,
		Suggestion:    job.Suggestion,
	}
}

func formatOptionalResponseTime(value *time.Time) *string {
	if value == nil {
		return nil
	}
	formatted := value.Format(timeFormat)
	return &formatted
}

func writeEnrichmentServiceError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, enrichment.ErrItemNotReady):
		writeError(c, NewAPIError(http.StatusBadRequest, "item_not_ready", "item needs a title and at least one photo before enrichment"))
	case errors.Is(err, enrichment.ErrJobNotComplete):
		writeError(c, NewAPIError(http.StatusBadRequest, "enrichment_not_complete", "enrichment job is not complete"))
	case errors.Is(err, enrichment.ErrJobNotFound):
		writeError(c, NewAPIError(http.StatusNotFound, "enrichment_job_not_found", "enrichment job was not found"))
	case errors.Is(err, items.ErrItemNotFound):
		writeError(c, NewAPIError(http.StatusNotFound, "item_not_found", "item was not found"))
	default:
		writeError(c, NewAPIError(http.StatusInternalServerError, "internal_error", "request failed"))
	}
}
