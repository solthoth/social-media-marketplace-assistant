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

## Testing

Run backend tests:

```sh
make test-api
```

