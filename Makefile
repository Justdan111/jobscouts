# Convenience targets. Backend on :8080, frontend on :3000.
.PHONY: backend frontend dev install

install:
	cd frontend && npm install

backend:
	cd backend && go run ./cmd/server

frontend:
	cd frontend && npm run dev

# Run both (needs two terminals, or use `make -j2 dev`)
dev:
	@echo "Run 'make backend' in one terminal and 'make frontend' in another."
