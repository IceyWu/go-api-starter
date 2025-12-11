.PHONY: build run test clean swagger

# Build the application
build:
	go build -o bin/server ./cmd/server

# Run the application (development)
run:
	@if exist .env.dev (for /f "tokens=*" %%a in (.env.dev) do @set "%%a") & go run ./cmd/server
	@if not exist .env.dev set APP_ENV=development & go run ./cmd/server

# Run the application (production mode locally)
run-prod:
	@if exist .env.prod (for /f "tokens=*" %%a in (.env.prod) do @set "%%a") & go run ./cmd/server
	@if not exist .env.prod set APP_ENV=production & go run ./cmd/server

# Run dev (cross-platform using PowerShell)
dev:
	powershell -Command "Get-Content .env.dev | ForEach-Object { if($$_ -match '^([^#][^=]*)=(.*)$$') { [Environment]::SetEnvironmentVariable($$matches[1], $$matches[2]) } }; go run ./cmd/server"

# Run prod (cross-platform using PowerShell)
prod:
	powershell -Command "Get-Content .env.prod | ForEach-Object { if($$_ -match '^([^#][^=]*)=(.*)$$') { [Environment]::SetEnvironmentVariable($$matches[1], $$matches[2]) } }; go run ./cmd/server"

# Run tests
test:
	go test -v ./...

# Clean build artifacts
clean:
	rm -rf bin/
	rm -f *.db
	rm -f server.exe

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

# Build for production
build-prod:
	CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o bin/server ./cmd/server

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

# Sync from template repository
sync:
	npx syn --repo https://github.com/IceyWu/go-api-starter --branch main
