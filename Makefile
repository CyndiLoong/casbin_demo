.PHONY: all build run test clean docker-up docker-down backend-run frontend-run install-deps

all: build

install-deps:
	cd backend && go mod tidy
	cd frontend && npm install

build-backend:
	cd backend && go build -o server.exe ./cmd/server

build-frontend:
	cd frontend && npm run build

build: build-backend build-frontend

backend-run:
	cd backend && go run ./cmd/server

frontend-run:
	cd frontend && npm run dev

run: backend-run

test-backend:
	cd backend && go test -v ./...

test-api:
	powershell -ExecutionPolicy Bypass -File ./scripts/test-api.ps1

docker-up:
	docker-compose up -d --build

docker-down:
	docker-compose down -v

docker-logs:
	docker-compose logs -f

clean:
	cd backend && if exist server.exe del server.exe
	cd frontend && if exist dist rmdir /s /q dist
