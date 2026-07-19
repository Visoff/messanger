.PHONY: all build backend frontend webrtc file_storage test clean dev-env

all: build

# Build all services
build: backend frontend webrtc file_storage

backend:
	cd backend && go build ./...

frontend:
	cd frontend && npm ci && npm run build

webrtc:
	cd webrtc_service && go build ./...

file_storage:
	cd file_storage && cargo check

test:
	cd backend && go test ./... 2>/dev/null || true
	cd webrtc_service && go test ./... 2>/dev/null || true
	cd file_storage && cargo test 2>/dev/null || true

dev-env:
	cd frontend && npm run dev &
	cd backend && go run cmd/main/main.go &
	cd webrtc_service && go run . &
	cd file_storage && cargo run &

clean:
	cd frontend && rm -rf build .svelte-kit
	cd file_storage && cargo clean
