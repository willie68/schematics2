.PHONY: build gencert build-web build-server build-all clean help

help:
	@echo "MCSPhotoIndex Build Targets:"
	@echo "  make build        - Build everything (web first, then server)"
	@echo "  make gencert      - Generate TLS certificate and key"
	@echo "  make build-web    - Build Frontend only"
	@echo "  make build-server - Build server binaries only (requires web build first)"
	@echo "  make build-all    - Build all binaries"
	@echo "  make clean        - Clean build artifacts"

build: gencert build-web build-server

gencert:
	@echo "Generating TLS certificate..."
	cd .. && go run ./cmd/gencert && cd scripts
	@echo "Certificate generation complete!"

build-web:
	@echo "Building Frontend..."
	cd ../../frontend && npm run build && cd ../../backend
	@echo "Frontend build complete!"

build-server: build-all

build-all: gencert build-web
	@echo "Building Go binaries..."
	cd .. && \
	go build -ldflags="-s -w" -o ./bin/mcsphotoindex-server ./cmd/server && \
	go build -ldflags="-s -w" -o ./bin/mcsphotoindex-command ./cmd/command && \
	go build -ldflags="-s -w" -o ./bin/mcsphotoindex-duplicates ./cmd/duplicates && \
	go build -ldflags="-s -w" -o ./bin/mcsphotoindex-container ./cmd/container && \
	go build -ldflags="-s -w" -o ./bin/mcsphotoindex-management ./cmd/management && \
	go build -ldflags="-s -w" -o ./bin/mcsphotoindex-imagecheck ./cmd/imagecheck && \
	cd scripts
	@echo "Build complete!"

clean:
	@echo "Cleaning build artifacts..."
	cd .. && rm -f ./bin/mcsphotoindex-* && rm -rf ../frontend/dist && cd scripts
	@echo "Clean complete!"
