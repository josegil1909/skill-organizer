.PHONY: all help dev build install test clean

all: help

help:
	@echo "AI Ecosystem Organizer — available commands:"
	@echo "  make dev      - Start backend (Go:8080) and frontend (Astro:4321) in parallel"
	@echo "  make install  - Install backend and frontend (Bun) dependencies"
	@echo "  make test     - Run backend unit tests"
	@echo "  make build    - Build the static frontend and the Go binary"
	@echo "  make clean    - Remove build artifacts (bin/, dist/)"

install:
	@echo "Installing Go dependencies in backend..."
	cd backend && go mod download
	@echo "Installing Bun dependencies in frontend..."
	cd frontend && bun install

test:
	@echo "Running backend tests..."
	cd backend && go test -v ./...

dev:
	@echo "Starting Go backend (port 8080) and Astro frontend (port 4321)..."
	@trap 'kill 0' EXIT; \
	(cd backend && go run main.go) & \
	(cd frontend && bun run dev) & \
	wait

build:
	@echo "Building frontend with Bun..."
	cd frontend && bun run build
	@echo "Compiling Go binary to bin/organizer-backend..."
	mkdir -p bin
	cd backend && go build -o ../bin/organizer-backend main.go
	@echo "Build succeeded: bin/organizer-backend"

clean:
	rm -rf bin/ frontend/dist/
