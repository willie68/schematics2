SHELL := /bin/sh

.PHONY: help backend-run backend-tidy frontend-install frontend-run

help:
	@echo "Targets:"
	@echo "  backend-tidy     - go mod tidy im Backend"
	@echo "  backend-run      - Backend starten"
	@echo "  frontend-install - npm install im Frontend"
	@echo "  frontend-run     - Frontend dev server starten"

backend-tidy:
	cd backend && go mod tidy

backend-run:
	cd backend && go run cmd/api/main.go

frontend-install:
	cd frontend && npm install

frontend-run:
	cd frontend && npm run dev
