# Development

## Local Backend

Run:

```sh
make run-api
```

Environment variables:

- `PORT`: API port. Defaults to `8080`.
- `DATABASE_PATH`: SQLite database path. Defaults to `data/app.db`.

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

Use `/items/new` to capture a new draft item. Use `/items/{id}/edit` to edit an existing item from the inventory list. The form stores original purchase price and selling price as decimal dollars in the UI and sends integer cents to the API.

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

## API

See [api.md](api.md) for the current API endpoints and response shapes.
