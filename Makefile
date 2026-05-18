.PHONY: install-web run-web build-web test-web run-api fmt-check vet-api test-api test-api-integration test verify

install-web:
	npm install

fmt-check:
	npm run format:check

run-web:
	npm --workspace apps/web start

build-web:
	npm --workspace apps/web run build

test-web:
	npm --workspace apps/web run test

run-api:
	go run ./services/api/cmd/api

vet-api:
	go vet ./services/api/...

test-api:
	go test ./services/api/...

test-api-integration:
	go test -tags=integration ./services/api/internal/integration

test: test-api test-web test-api-integration

verify:
	npm run verify
