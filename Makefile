.PHONY: dev fmt test proto

dev:
	./scripts/dev.sh

fmt:
	cd services/core-go && go fmt ./...
	cd apps/desktop && npm run lint --if-present
	cd services/ai-py && python -m ruff format .

test:
	cd services/core-go && go test ./...
	cd services/ai-py && python -c "import main"
	cd apps/desktop && npm run build

proto:
	@echo "proto generation is not configured yet; see README for guidance."
