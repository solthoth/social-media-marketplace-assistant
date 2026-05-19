# Development

## Local Backend

Run:

```sh
make run-api
```

Environment variables:

- `PORT`: API port. Defaults to `8080`.
- `DATABASE_PATH`: SQLite database path. Defaults to `data/app.db`.
- `PHOTO_STORAGE_PATH`: local item photo storage path. Defaults to `data/photos`.
- `AI_ENRICHMENT_ENABLED`: enables AI enrichment routes and worker behavior. Defaults to disabled until implemented.
- `AI_PROVIDER`: AI provider adapter. Use `fake` for tests and local deterministic development; use `openai` for OpenAI.
- `AI_MODEL`: model name for the configured AI provider.
- `OPENAI_API_KEY`: required only when `AI_PROVIDER=openai`.

Health endpoint:

```sh
curl http://localhost:8080/healthz
```

## Local Frontend

Install dependencies:

```sh
make install-web
```

Run:

```sh
make run-web
```

The Angular dev server should be available at `http://localhost:4200`.

The frontend calls API routes through the `/api` path. The Angular development server proxies `/api` to the Go API at `http://localhost:8080`.

The inventory page at `/items` loads item records from `GET /api/items`, supports status filtering, and filters the loaded result set by title, category, or notes.

Use `/items/new` to capture a new draft item. Use `/items/{id}/edit` to edit an existing item from the inventory list. The form stores original purchase price and selling price as decimal dollars in the UI and sends integer cents to the API. Existing items move through the status workflow from the edit form.

Photo capture is available from `/items/{id}/edit` after an item exists. The photo UI supports phone camera capture and local image file selection, previews thumbnail content through the API proxy, and lets sellers remove, reorder, and choose the primary photo.

The frontend uses Angular Material components with the Rose/Red prebuilt theme. Prefer Material form fields, buttons, toolbar, and card components for new mobile-facing UI controls so spacing, touch targets, and validation states stay consistent.

## Docker Compose

Build and run the web and API containers:

```sh
docker compose up --build
```

The web app is available at `http://localhost:4200`, and the API is available at `http://localhost:8080`. The web container proxies `/api` requests to the API container.

## Git Hooks

Install `pre-commit`:

```sh
pipx install pre-commit
```

or:

```sh
python -m pip install pre-commit
```

Install the repository hooks:

```sh
pre-commit install --hook-type pre-commit --hook-type commit-msg
```

The `pre-commit` hook runs the same core quality gates as GitHub Actions. The `commit-msg` hook verifies conventional-changelog commit messages.

## Dependency Updates

Dependabot is configured in `.github/dependabot.yml` to check npm workspaces, Go modules, GitHub Actions, Dockerfiles, and Docker Compose weekly on Monday mornings Pacific time. Dependency update pull requests use conventional commit prefixes so commitlint can validate them.

## Testing

Run all tests:

```sh
make test
```

Run backend tests only:

```sh
make test-api
```

Run backend integration tests only:

```sh
make test-api-integration
```

Run frontend tests only:

```sh
make test-web
```

Run end-to-end tests:

```sh
npm run e2e:test
```

Playwright starts the API and frontend dev server for the E2E suite. These tests are intentionally not part of the pre-commit hook.

## Verification

Run the local equivalent of the CI checks:

```sh
make verify
```

or:

```sh
npm run verify
```

## Data Model

See [data-model.md](data-model.md) for the current MVP entities, statuses, and persistence assumptions.

## Photo Storage

See [photo-storage.md](photo-storage.md) for local photo storage, planned image variants, and the future cloud storage adapter boundary.

The photo upload API stores supported images in the configured photo storage path and serves them through API content routes. Supported upload formats are JPEG, PNG, and WebP.

## AI Enrichment

See [ai-enrichment.md](ai-enrichment.md) for the planned asynchronous workflow that generates missing item details from title and photos. CI and local tests should use the fake provider so runs stay deterministic and do not require external API calls.

## API

See [api.md](api.md) for the current API endpoints and response shapes.
