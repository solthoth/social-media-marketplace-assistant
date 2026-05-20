package sqlite

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"time"

	"github.com/solthoth/social-media-marketplace-assistant/services/api/internal/enrichment"
)

type EnrichmentRepository struct {
	db *sql.DB
}

func NewEnrichmentRepository(db *sql.DB) EnrichmentRepository {
	return EnrichmentRepository{db: db}
}

func (r EnrichmentRepository) CreateJob(ctx context.Context, job enrichment.Job) (enrichment.Job, error) {
	inputSnapshot, err := json.Marshal(job.InputSnapshot)
	if err != nil {
		return enrichment.Job{}, err
	}
	suggestion, err := json.Marshal(job.Suggestion)
	if err != nil {
		return enrichment.Job{}, err
	}

	_, err = r.db.ExecContext(ctx, `
insert into item_enrichment_jobs (
  id, item_id, status, provider, model, requested_at, started_at, completed_at, applied_at, error_message, input_snapshot_json, suggestion_json
) values (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		job.ID,
		job.ItemID,
		job.Status,
		job.Provider,
		job.Model,
		formatTime(job.RequestedAt),
		formatOptionalTime(job.StartedAt),
		formatOptionalTime(job.CompletedAt),
		formatOptionalTime(job.AppliedAt),
		job.ErrorMessage,
		string(inputSnapshot),
		string(suggestion),
	)
	if err != nil {
		return enrichment.Job{}, err
	}
	return job, nil
}

func (r EnrichmentRepository) ListJobs(ctx context.Context, itemID string) ([]enrichment.Job, error) {
	rows, err := r.db.QueryContext(ctx, `
select id, item_id, status, provider, model, requested_at, started_at, completed_at, applied_at, error_message, input_snapshot_json, suggestion_json
from item_enrichment_jobs
where item_id = ?
order by requested_at desc`, itemID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	result := []enrichment.Job{}
	for rows.Next() {
		job, err := scanEnrichmentJob(rows)
		if err != nil {
			return nil, err
		}
		result = append(result, job)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return result, nil
}

func (r EnrichmentRepository) GetJob(ctx context.Context, itemID string, jobID string) (enrichment.Job, error) {
	row := r.db.QueryRowContext(ctx, `
select id, item_id, status, provider, model, requested_at, started_at, completed_at, applied_at, error_message, input_snapshot_json, suggestion_json
from item_enrichment_jobs
where item_id = ? and id = ?`, itemID, jobID)
	return scanEnrichmentJobRow(row)
}

func (r EnrichmentRepository) GetJobByID(ctx context.Context, jobID string) (enrichment.Job, error) {
	row := r.db.QueryRowContext(ctx, `
select id, item_id, status, provider, model, requested_at, started_at, completed_at, applied_at, error_message, input_snapshot_json, suggestion_json
from item_enrichment_jobs
where id = ?`, jobID)
	return scanEnrichmentJobRow(row)
}

func (r EnrichmentRepository) UpdateJob(ctx context.Context, job enrichment.Job) (enrichment.Job, error) {
	inputSnapshot, err := json.Marshal(job.InputSnapshot)
	if err != nil {
		return enrichment.Job{}, err
	}
	suggestion, err := json.Marshal(job.Suggestion)
	if err != nil {
		return enrichment.Job{}, err
	}

	result, err := r.db.ExecContext(ctx, `
update item_enrichment_jobs
set status = ?,
    provider = ?,
    model = ?,
    started_at = ?,
    completed_at = ?,
    applied_at = ?,
    error_message = ?,
    input_snapshot_json = ?,
    suggestion_json = ?
where id = ?`,
		job.Status,
		job.Provider,
		job.Model,
		formatOptionalTime(job.StartedAt),
		formatOptionalTime(job.CompletedAt),
		formatOptionalTime(job.AppliedAt),
		job.ErrorMessage,
		string(inputSnapshot),
		string(suggestion),
		job.ID,
	)
	if err != nil {
		return enrichment.Job{}, err
	}
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return enrichment.Job{}, err
	}
	if rowsAffected == 0 {
		return enrichment.Job{}, enrichment.ErrJobNotFound
	}
	return job, nil
}

func scanEnrichmentJobRow(row *sql.Row) (enrichment.Job, error) {
	job, err := scanEnrichmentJob(row)
	if errors.Is(err, sql.ErrNoRows) {
		return enrichment.Job{}, enrichment.ErrJobNotFound
	}
	return job, err
}

type enrichmentJobScanner interface {
	Scan(dest ...any) error
}

func scanEnrichmentJob(scanner enrichmentJobScanner) (enrichment.Job, error) {
	var job enrichment.Job
	var status string
	var requestedAt string
	var startedAt sql.NullString
	var completedAt sql.NullString
	var appliedAt sql.NullString
	var inputSnapshot string
	var suggestion string
	err := scanner.Scan(
		&job.ID,
		&job.ItemID,
		&status,
		&job.Provider,
		&job.Model,
		&requestedAt,
		&startedAt,
		&completedAt,
		&appliedAt,
		&job.ErrorMessage,
		&inputSnapshot,
		&suggestion,
	)
	if err != nil {
		return enrichment.Job{}, err
	}
	job.Status = enrichment.JobStatus(status)
	job.RequestedAt, err = parseTime(requestedAt)
	if err != nil {
		return enrichment.Job{}, err
	}
	if job.StartedAt, err = parseOptionalTime(startedAt); err != nil {
		return enrichment.Job{}, err
	}
	if job.CompletedAt, err = parseOptionalTime(completedAt); err != nil {
		return enrichment.Job{}, err
	}
	if job.AppliedAt, err = parseOptionalTime(appliedAt); err != nil {
		return enrichment.Job{}, err
	}
	if err := json.Unmarshal([]byte(inputSnapshot), &job.InputSnapshot); err != nil {
		return enrichment.Job{}, err
	}
	if err := json.Unmarshal([]byte(suggestion), &job.Suggestion); err != nil {
		return enrichment.Job{}, err
	}
	return job, nil
}

func formatOptionalTime(value *time.Time) any {
	if value == nil {
		return nil
	}
	return formatTime(*value)
}

func parseOptionalTime(value sql.NullString) (*time.Time, error) {
	if !value.Valid {
		return nil, nil
	}
	parsed, err := parseTime(value.String)
	if err != nil {
		return nil, err
	}
	return &parsed, nil
}
