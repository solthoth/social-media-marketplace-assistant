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

