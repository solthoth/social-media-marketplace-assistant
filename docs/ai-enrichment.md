# AI Item Enrichment

AI item enrichment helps the seller move from a draft item with a title and photos to a more complete inventory record. The feature is intended to reduce manual entry while keeping the seller in control of the final listing details.

## Goals

- Generate item details from the item title and photos.
- Use OpenAI vision first, while keeping the provider behind an adapter so Azure OpenAI, AWS, or another model provider can replace it later.
- Run enrichment asynchronously so the edit form remains usable while generation is queued or processing.
- Fill only empty fields in the first implementation.
- Keep enrichment job history for auditability and future review.
- Avoid pricing suggestions in the first implementation.

## Non-Goals

- Automatically publishing generated listings.
- Replacing the seller review step.
- Scraping marketplaces or external sites for comparable pricing.
- Generating or changing original purchase price or selling price.

AI-assisted pricing should be revisited as a separate feature after the first enrichment workflow is stable.

## User Flow

1. The seller creates a draft item with a title.
2. The seller uploads one or more item photos from the edit page.
3. The seller starts AI enrichment.
4. The backend creates an enrichment job and processes it asynchronously.
5. The AI provider analyzes the title and photo content.
6. The backend stores the generated suggestion in job history.
7. The frontend applies completed suggestions only to currently empty fields.
8. The seller reviews and edits the final item details before listing.

## Backend Design

Add an `internal/enrichment` package for item enrichment orchestration:

- validate item readiness for enrichment.
- create and store enrichment jobs.
- collect item and photo inputs.
- call an AI provider through an interface.
- persist completed suggestions or failed job state.
- apply suggestions to empty item fields.

Add an `internal/ai` package for provider adapters:

- `fake`: deterministic provider for local tests and CI.
- `openai`: OpenAI Responses API provider for production use.
- future provider adapters can implement the same interface.

Suggested provider interface:

```go
type ItemDetailProvider interface {
    GenerateItemDetails(ctx context.Context, input ItemDetailInput) (ItemDetailSuggestion, error)
}
```

Suggested input shape:

```go
type ItemDetailInput struct {
    ItemID              string
    Title               string
    ExistingDescription string
    ExistingCategory    string
    ExistingSize        string
    ExistingCondition   string
    ExistingNotes       string
    Photos              []ItemPhotoInput
}
```

Suggested output shape:

```go
type ItemDetailSuggestion struct {
    Description string
    Category    string
    Size        string
    Condition   string
    Notes       string
}
```

## Persistence

Add an `item_enrichment_jobs` table:

- `id`
- `item_id`
- `status`
- `provider`
- `model`
- `requested_at`
- `started_at`
- `completed_at`
- `applied_at`
- `error_message`
- `input_snapshot_json`
- `suggestion_json`

Suggested statuses:

- `queued`
- `processing`
- `completed`
- `failed`

Use `applied_at` instead of a separate `applied` status so completed jobs remain easy to query while still recording when a suggestion was applied.

## API Design

Suggested endpoints:

- `POST /items/{id}/enrichment-jobs`
- `GET /items/{id}/enrichment-jobs`
- `GET /items/{id}/enrichment-jobs/{jobId}`
- `POST /items/{id}/enrichment-jobs/{jobId}/apply`

The apply endpoint should fill only empty item fields in the first implementation and return the updated item, job, and applied field list.

Example response:

```json
{
  "item": {},
  "job": {},
  "applied_fields": ["description", "category", "condition"]
}
```

## OpenAI Provider

The OpenAI adapter uses the Responses API with image inputs. The backend provides image inputs as base64 data URLs because local item photos are private and not publicly reachable.

Use structured output from the provider where practical so the service can validate and persist predictable suggestion fields. The prompt should instruct the model to leave uncertain fields empty rather than guessing, especially for size and condition.

## Configuration

Proposed environment variables:

- `AI_ENRICHMENT_ENABLED`: enables enrichment routes and worker behavior. Defaults to `false` until the feature is ready.
- `AI_PROVIDER`: `fake` or `openai`. Tests and CI should use `fake`.
- `AI_MODEL`: model name for the selected provider.
- `OPENAI_API_KEY`: required only when `AI_PROVIDER=openai`.

## Async Processing

The first implementation can use an in-process worker:

- API creates a queued job.
- Worker claims queued jobs and marks them processing.
- Worker calls the configured provider.
- Worker stores completed suggestions or failed state.
- Frontend polls the job endpoint until the job completes or fails.

The service boundary should not depend on in-process execution so a future queue-backed worker can replace it without changing the HTTP API.

## Testing Strategy

Backend tests:

- service rejects enrichment without a title.
- service rejects enrichment without photos.
- service creates queued jobs.
- fake provider completes jobs.
- provider failure marks jobs failed.
- apply fills empty fields only.
- apply does not overwrite existing fields.

Repository tests:

- create job.
- list job history.
- fetch one job.
- update job status.
- store input snapshots and suggestions.

Handler and integration tests:

- create enrichment job.
- list and fetch jobs.
- process job with fake provider.
- apply suggestions and verify item fields.

Frontend tests:

- AI panel is disabled until title and photos exist.
- starting a job calls the enrichment API.
- polling renders queued, processing, completed, and failed states.
- completed suggestions fill only empty form fields.

End-to-end tests:

- create item.
- upload photo.
- trigger fake enrichment.
- wait for suggestions.
- confirm empty fields are populated.
