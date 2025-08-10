# Grafterm Development Guidelines

## Build & Test Commands

### 🔧 Go Installation
**Falls Go nicht installiert ist:**
- `./install-go.sh` - Installiere Go automatisch für macOS
- `./no-go-help.sh` - Zeigt Alternativen ohne lokale Go-Installation
- `brew install go` - Installiere Go mit Homebrew
- Manuell: Download von https://golang.org/dl/

### 🔧 Dependency Management
- `./fix-deps.sh` - Behebt fehlende Abhängigkeiten und aktualisiert go.sum
- `go mod tidy` - Bereinigt Abhängigkeiten
- `go mod download` - Läd Abhängigkeiten herunter
- `go mod verify` - Verifiziert Abhängigkeiten

### 🚀 Simple Build Commands (Go erforderlich)
- `./build.sh` - Build binary for current platform
- `make -f Makefile.simple build` - Build using simple Makefile
- `make -f Makefile.simple build-all` - Build for multiple platforms
- `make -f Makefile.simple install` - Install to GOPATH/bin

### 🐳 Docker Build Commands (kein Go erforderlich)
- `./build-docker.sh` - Build mit Docker
- `docker run -v $(pwd):/src -w /src golang:latest go build ./cmd/grafterm` - Direkter Docker Build

### 🧪 Testing Commands
- `./test.sh` - Run unit tests
- `./test-integration.sh` - Run integration tests
- `make -f Makefile.simple test` - Run unit tests
- `make -f Makefile.simple test-integration` - Run integration tests
- `make -f Makefile.simple test-all` - Run all tests
- `make -f Makefile.simple test-single` - Run single test (interactive)
- `make -f Makefile.simple test-package` - Test specific package (interactive)
- `go test ./... -v -run TestName` - Run single test directly
- `go test ./internal/service/metric/prometheus -v` - Test specific package

### 🛠️ Development Commands
- `make -f Makefile.simple deps` - Download dependencies
- `make -f Makefile.simple fmt` - Format code
- `make -f Makefile.simple lint` - Run linter (if installed)
- `make -f Makefile.simple clean` - Clean build artifacts
- `make -f Makefile.simple run` - Build and run application

### 📦 Original Docker Commands (Still Available)
- `make test` - Run integration tests (Docker)
- `make unit-test` - Run unit tests (Docker)
- `make integration-test` - Run integration tests (Docker)
- `make build` - Build development Docker image
- `make build-binary` - Build production binary (Docker)

## Code Style Guidelines

### Go Conventions
- Use standard Go formatting: `go fmt ./...`
- Follow Go naming conventions: CamelCase for exported, camelCase for private
- Interface names should be `-er` suffixes when possible (e.g., `Gatherer`)
- Package names: lowercase, single words, no underscores

### Error Handling
- Always check and return errors explicitly
- Use typed errors for expected error conditions
- Wrap errors with context using `fmt.Errorf`: `return fmt.Errorf("failed to gather metrics: %w", err)`
- Handle context timeouts specifically: `if ctx.Err() == context.DeadlineExceeded`
- Handle context cancellation: `if ctx.Err() == context.Canceled`
- Don't ignore errors - handle or return them
- Use timeouts for all external operations (2-5 seconds recommended)

### Imports
- Group imports: standard, third-party, local
- Use absolute paths for internal imports
- Keep imports minimal and remove unused imports

### Types & Interfaces
- Prefer interface types as return values (contravariant)
- Define interfaces on the consumer side
- Use concrete types as function parameters when possible

### Testing
- Use testify for assertions
- Follow naming: `TestXxx` for unit tests
- Use `-tags='integration'` for integration tests
- Mock external dependencies using mockgen

### Logging
- Use zerolog for all logging
- Log at appropriate levels (Info, Warn, Error)
- Use structured logging when helpful

### Timeout Handling
- Use `context.WithTimeout` for all external API calls
- Set reasonable timeouts (2-5 seconds for widgets, 3-5 seconds for app sync)
- Check for `context.DeadlineExceeded` and `context.Canceled` specifically
- Gracefully handle timeouts without crashing the application
- Log timeout errors appropriately for debugging