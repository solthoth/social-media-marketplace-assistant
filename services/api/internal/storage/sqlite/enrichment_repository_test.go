package sqlite

import (
	"context"
	"database/sql"
	"testing"
	"time"

	"github.com/solthoth/social-media-marketplace-assistant/services/api/internal/enrichment"
	"github.com/stretchr/testify/suite"
	_ "modernc.org/sqlite"
)

type EnrichmentRepositorySuite struct {
	suite.Suite
	db         *sql.DB
	repository EnrichmentRepository
}

func TestEnrichmentRepositorySuite(t *testing.T) {
	suite.Run(t, new(EnrichmentRepositorySuite))
}

func (s *EnrichmentRepositorySuite) SetupTest() {
	db, err := sql.Open("sqlite", ":memory:")
	s.Require().NoError(err)
	s.Require().NoError(ApplyMigrations(context.Background(), db))
	s.db = db
	s.repository = NewEnrichmentRepository(db)
}

func (s *EnrichmentRepositorySuite) TearDownTest() {
	s.Require().NoError(s.db.Close())
}

func (s *EnrichmentRepositorySuite) TestCreateListGetAndUpdateJob() {
	now := time.Now().UTC()
	job := enrichment.Job{
		ID:          "job-1",
		ItemID:      "item-1",
		Status:      enrichment.JobStatusQueued,
		Provider:    "fake",
		Model:       "fake-vision",
		RequestedAt: now,
		InputSnapshot: enrichment.ItemDetailInput{
			ItemID: "item-1",
			Title:  "Denim jacket",
			Photos: []enrichment.ItemPhotoInput{{ID: "photo-1", Filename: "front.png", MimeType: "image/png"}},
		},
	}

	created, err := s.repository.CreateJob(context.Background(), job)
	s.Require().NoError(err)
	s.Equal(job.ID, created.ID)

	list, err := s.repository.ListJobs(context.Background(), "item-1")
	s.Require().NoError(err)
	s.Len(list, 1)
	s.Equal("Denim jacket", list[0].InputSnapshot.Title)

	completedAt := now.Add(time.Second)
	created.Status = enrichment.JobStatusCompleted
	created.CompletedAt = &completedAt
	created.Suggestion = enrichment.ItemDetailSuggestion{Category: "Clothing"}
	updated, err := s.repository.UpdateJob(context.Background(), created)
	s.Require().NoError(err)
	s.Equal(enrichment.JobStatusCompleted, updated.Status)

	fetched, err := s.repository.GetJob(context.Background(), "item-1", job.ID)
	s.Require().NoError(err)
	s.Equal("Clothing", fetched.Suggestion.Category)

	byID, err := s.repository.GetJobByID(context.Background(), job.ID)
	s.Require().NoError(err)
	s.Equal(fetched.ID, byID.ID)
}

func (s *EnrichmentRepositorySuite) TestGetJobRejectsMissingJob() {
	_, err := s.repository.GetJob(context.Background(), "item-1", "missing")

	s.ErrorIs(err, enrichment.ErrJobNotFound)
}
