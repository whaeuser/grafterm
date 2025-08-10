# Grafterm Development Guidelines

## Build & Test Commands
- `make test` - Run integration tests
- `make unit-test` - Run unit tests  
- `make integration-test` - Run integration tests
- `go test ./... -v -run TestName` - Run single test (requires Go installation)
- `go test ./internal/service/metric/prometheus -v` - Test specific package (requires Go)
- `make build` - Build development Docker image
- `make build-binary` - Build production binary
- `make deps` - Vendor dependencies

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