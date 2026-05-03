APP_NAME ?= rbc-2026-api
API_IMAGE ?= $(APP_NAME):local
BASE_URL ?= http://localhost:9999

.PHONY: build docker-build docker-up load-test-local

build:
	go build -trimpath -ldflags="-s -w" -o bin/api ./cmd/api

docker-build:
	docker build --platform linux/amd64 -t $(API_IMAGE) .

docker-up:
	API_IMAGE=$(API_IMAGE) docker compose up -d

load-test-local:
	BASE_URL=$(BASE_URL) k6 run test/load-test.js
