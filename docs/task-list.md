# Initial Task List

This roadmap captures the first implementation work needed to move the app from scaffold to a usable private inventory system. The order is intentional: build a durable inventory source of truth before investing in external publishing integrations.

## Phase 1: App Foundation

1. ~~Define the MVP data model.~~
   - ~~Include item, item photo, inventory status, listing status, connected account, and sale.~~
   - ~~Decide the first required item fields: title, description, category, size, condition, price, status, and notes.~~

2. ~~Choose persistence.~~
   - ~~Start with SQLite for local development and early private deployment.~~
   - ~~Keep persistence behind a repository interface so Postgres or another database can replace it later.~~
   - ~~Add migrations and database setup documentation.~~

3. ~~Structure the backend API.~~
   - ~~Add package boundaries for handlers, services, repositories, and storage.~~
   - ~~Add a consistent error response shape.~~
   - ~~Add a request validation pattern.~~
   - ~~Keep `/healthz` as the operational health endpoint.~~

4. ~~Structure the frontend app.~~
   - ~~Add routing.~~
   - ~~Add an app shell and navigation.~~
   - ~~Add an API client service.~~
   - ~~Establish shared UI and state conventions.~~

## Phase 2: Inventory MVP

5. Create the item API.
   - `POST /items`
   - `GET /items`
   - `GET /items/{id}`
   - `PATCH /items/{id}`
   - Archive behavior for removed items.

6. Build the inventory list UI.
   - Show items in a card or table layout.
   - Filter by draft, ready to list, listed, sold, and archived.
   - Search by title, category, and notes.

7. Build the item capture and edit UI.
   - Optimize the form for mobile use.
   - Support saving drafts.
   - Validate required item fields.

8. Implement the status workflow.
   - Use initial statuses: `draft`, `ready_to_list`, `listed`, `sold`, and `archived`.
   - Add backend validation for valid status transitions.
   - Add frontend controls for changing item status.

## Phase 3: Photos And Media

9. Design photo storage.
   - Start with local filesystem storage for development.
   - Define a future object storage interface for S3-compatible storage.

10. Create the photo upload API.
    - Upload photos for an item.
    - List item photos.
    - Delete and reorder photos.

11. Build the photo capture UI.
    - Use camera and file inputs that work well on phones.
    - Preview, remove, and reorder photos.
    - Mark one photo as the primary photo.

## Phase 4: Seller Workflow

12. Build a listing preparation view.
    - Review title, description, price, photos, and destination readiness before publishing.

13. Build the inventory dashboard.
    - Show counts by status.
    - Show recently updated items.
    - Show items needing attention.

14. Add sale tracking.
    - Mark items sold.
    - Store sale date, sale price, and optional platform or account.
    - Hide sold items from publish queues by default.

## Phase 5: Platform Integration Readiness

15. Define the integration adapter interface.
    - Include capabilities, validation, publish result, and unsupported feature reporting.
    - Keep adapters platform-compliant and based on official APIs or permitted workflows.

16. Add a manual export adapter.
    - Provide a safe copy or export workflow before automating platform publishing.
    - Use this to deliver value without platform API risk.

17. Add the connected account model.
    - Store account metadata, platform, status, permissions, and last validation time.
    - Do not store real credentials until authentication and secret handling are designed.

18. Run the first platform research spike.
    - Pick one target platform.
    - Document official API support, terms, rate limits, auth requirements, and publish feasibility.

## Phase 6: Security And Operations

19. Add authentication.
    - Start with a simple private household or team login.
    - Leave room for seller, partner, and admin roles later.

20. Add authorization.
    - Protect item, media, and connected account APIs.
    - Define basic role permissions.

21. Add configuration management.
    - Add `.env.example`.
    - Document required environment variables.
    - Add config loading in Go.

22. Add a deployment baseline.
    - Containerize the API and web app.
    - Add production build docs.
    - Decide the initial hosting target.

23. Add backup and export.
    - Export inventory and photo metadata.
    - Document restore behavior.
    - Support the core goal of keeping inventory portable if an external account becomes unavailable.

## Phase 7: Quality And Documentation

24. Add API contract documentation.
    - Add OpenAPI generation or a maintained OpenAPI spec.
    - Use the contract to guide frontend API client work.

25. Expand Backstage-ready docs.
    - Add architecture decisions, runbooks, local development notes, and feature roadmap docs.

26. Expand the testing strategy.
    - Add backend handler and service tests.
    - Add frontend component and service tests.
    - Add an end-to-end smoke test once core flows exist.

## Recommended Starting Point

Start with these tasks:

1. ~~Define and document the MVP data model and statuses.~~
2. Add SQLite persistence and the item CRUD API.
3. Build the inventory list and item capture/edit UI.
