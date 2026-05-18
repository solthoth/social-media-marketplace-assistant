# Development

## Local Backend

Run:

```sh
make run-api
```

Environment variables:

- `PORT`: API port. Defaults to `8080`.

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

## Testing

Run all tests:

```sh
make test
```

Run backend tests only:

```sh
make test-api
```

Run frontend tests only:

```sh
make test-web
```

## Verification

Run the local equivalent of the CI checks:

```sh
make verify
```

or:

```sh
npm run verify
```
