# Architecture

## Monorepo Structure

```text
apps/web
services/api
docs
```

## Frontend

The Angular app owns the seller-facing user experience. It should stay optimized for fast capture, simple review, and clear inventory state.

Initial responsibilities:

- Item capture UI.
- Inventory dashboard.
- Listing status visibility.
- API health visibility during early development.

## Backend

The Go API owns core business data and integration orchestration.

Initial responsibilities:

- Health endpoint.
- Future item, media, account, and publishing endpoints.
- Integration boundary for platform-specific publishers.

Initial package boundaries:

- `internal/domain`: core entities and value objects.
- `internal/config`: environment-backed application configuration.
- `internal/httpserver`: HTTP routing and response helpers.
- `internal/items`: item service boundary.
- `internal/enrichment`: asynchronous AI-assisted item detail generation.
- `internal/ai`: provider adapters for AI model integrations.
- `internal/storage/sqlite`: SQLite connection and migrations.
- `internal/storage/localphotos`: future local filesystem photo storage adapter.

## Persistence

The first persistence target is SQLite using a pure Go driver. This keeps local development and CI simple while the app is private and early-stage.

Persistence should stay behind repository interfaces so a future Postgres migration does not force handlers or frontend contracts to change.

## Photo Storage

Photo bytes are stored outside SQLite. The MVP storage target is the local filesystem, configured by `PHOTO_STORAGE_PATH` and defaulting to `data/photos`. The backend should expose photo content through API routes so the frontend does not depend on local paths or future cloud storage details.

Photo storage should sit behind an internal interface so Azure, AWS, or another object/block storage provider can replace the local adapter later. See [photo-storage.md](photo-storage.md) for the detailed design.

## AI Enrichment

AI enrichment helps populate missing item details from the item title and photos. The first provider target is OpenAI vision through a provider adapter. The backend should own enrichment jobs, persistence, prompt construction, and provider calls while the frontend starts jobs, polls status, and applies completed suggestions to empty fields. See [ai-enrichment.md](ai-enrichment.md) for the detailed design.

## Integration Boundary

Platform integrations should be treated as adapters behind internal interfaces. Each adapter should document:

- Supported capabilities.
- Required permissions.
- Rate limits and operational constraints.
- Terms or policy considerations.
- Failure and retry behavior.

## Future Data Model

Likely core entities:

- User
- Item
- ItemPhoto
- InventoryStatus
- ConnectedAccount
- Listing
- ListingAttempt
- Sale
