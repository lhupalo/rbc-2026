BASE_URL ?= http://localhost:9999
IMAGE    ?= lhupalo/rbc-2026:latest

.PHONY: build docker-build docker-push docker-up load-test-local show-results

build:
	go build -trimpath -ldflags="-s -w" -o bin/api ./cmd/api

docker-build:
	docker build --platform linux/amd64 -t $(IMAGE) .

docker-push: docker-build
	docker push $(IMAGE)

docker-up:
	docker compose up -d

load-test-local:
	BASE_URL=$(BASE_URL) k6 run test/load-test.js
	@$(MAKE) --no-print-directory show-results

show-results:
	@echo ""
	@echo "========== RESULTADOS =========="
	@python3 scripts/show-results.py results.json
	@echo "================================"
