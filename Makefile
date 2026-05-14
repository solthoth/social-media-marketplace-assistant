.PHONY: install-web run-web build-web run-api test-api test

install-web:
	npm install

run-web:
	npm --workspace apps/web start

build-web:
	npm --workspace apps/web run build

run-api:
	go run ./services/api/cmd/api

test-api:
	go test ./services/api/...

test: test-api

