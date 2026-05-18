# Social Media Marketplace Assistant

A monorepo for a private inventory and listing assistant. The app is intended to help non-technical sellers capture item photos and details once, manage inventory from a central source, and publish eligible listings to connected social media or marketplace accounts.

The first design choices are:

- Frontend: Angular and TypeScript
- Backend: Go
- Repository shape: monorepo with independently deployable apps and services

This project should use official platform integrations where available and keep a durable local inventory record so unsold items can be republished or migrated if a connected account is unavailable.

## Repository Layout

```text
.
├── apps/
│   └── web/              # Angular frontend
├── docs/                 # Project documentation for humans and future Backstage use
├── services/
│   └── api/              # Go backend API
├── go.work               # Go workspace
├── Makefile              # Common local commands
└── package.json          # Root npm workspace commands
```

## Prerequisites

- Go 1.26+
- Node.js 24+
- npm 11+
- Python 3 with `pre-commit`

Install `pre-commit` with either:

```sh
pipx install pre-commit
```

or:

```sh
python -m pip install pre-commit
```

## Getting Started

Install frontend dependencies:

```sh
npm install
```

Install local Git hooks:

```sh
pre-commit install --hook-type pre-commit --hook-type commit-msg
```

Run the backend API:

```sh
make run-api
```

The API listens on `http://localhost:8080` by default.

Check backend health:

```sh
curl http://localhost:8080/healthz
```

Run the frontend:

```sh
make run-web
```

The Angular development server listens on `http://localhost:4200` by default.

## Useful Commands

```sh
make verify         # Run the same core checks as CI
make test           # Run backend and frontend tests
make test-api       # Run Go tests
make test-web       # Run Angular tests
make run-api        # Start the Go API
make install-web    # Install npm workspace dependencies
make run-web        # Start Angular dev server
make build-web      # Build Angular app
```

The equivalent npm command for CI-style local verification is:

```sh
npm run verify
```

## Documentation

Start with [docs/context.md](docs/context.md) for the durable project context that future assistant sessions should read first.
