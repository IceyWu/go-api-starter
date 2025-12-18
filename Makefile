.PHONY: build run test clean swagger dev prod

# Detect OS
ifeq ($(OS),Windows_NT)
    SHELL := powershell.exe
    .SHELLFLAGS := -NoProfile -Command
    RM = Remove-Item -Recurse -Force -ErrorAction SilentlyContinue
    LOAD_ENV_DEV = Get-Content .env.dev -ErrorAction SilentlyContinue | ForEach-Object { if($$_ -match '^([^#][^=]*)=(.*)$$') { [Environment]::SetEnvironmentVariable($$matches[1], $$matches[2]) } };
    LOAD_ENV_PROD = Get-Content .env.prod -ErrorAction SilentlyContinue | ForEach-Object { if($$_ -match '^([^#][^=]*)=(.*)$$') { [Environment]::SetEnvironmentVariable($$matches[1], $$matches[2]) } };
else
    RM = rm -rf
    LOAD_ENV_DEV = set -a && [ -f .env.dev ] && . ./.env.dev; set +a;
    LOAD_ENV_PROD = set -a && [ -f .env.prod ] && . ./.env.prod; set +a;
endif

# Build the application
build:
	go build -o bin/server ./cmd/server

# Run development
dev:
	$(LOAD_ENV_DEV) go run ./cmd/server

# Run production
prod:
	$(LOAD_ENV_PROD) go run ./cmd/server

# Run tests
test:
	go test -v ./...

# Clean build artifacts
clean:
	$(RM) bin/
	$(RM) *.db
	$(RM) server.exe
	$(RM) server

# Generate swagger docs
swagger:
	swag init -g cmd/server/main.go -o docs

# Install dependencies
deps:
	go mod tidy
	go mod download

# Format code
fmt:
	go fmt ./...

# Lint code
lint:
	golangci-lint run

# Build for Linux production
build-linux:
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags="-s -w" -o bin/server ./cmd/server

# Build for Windows production
build-windows:
	set CGO_ENABLED=0&& set GOOS=windows&& set GOARCH=amd64&& go build -ldflags="-s -w" -o bin/server.exe ./cmd/server

# Install pre-commit hooks
hooks:
	pre-commit install

# Run pre-commit on all files
hooks-run:
	pre-commit run --all-files

# Update dependencies
update:
	go get -u ./...
	go mod tidy
