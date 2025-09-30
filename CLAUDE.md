# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

Grafterm is a terminal-based metrics dashboard visualization tool similar to a minimalist Grafana for the terminal. It fetches metrics from various datasources (Prometheus, Graphite, InfluxDB) and renders them in the terminal using widgets (graphs, gauges, singlestats).

**Language**: Go
**Main Frameworks**:
- [Termdash](https://github.com/mum4k/termdash) for terminal rendering
- Prometheus client library for metrics gathering

## Common Commands

### Building
- `./build.sh` - Build binary for current platform (requires Go)
- `make -f Makefile.simple build` - Build using simple Makefile
- `make -f Makefile.simple build-all` - Build for multiple platforms
- `./build-docker.sh` - Build with Docker (no Go required)

### Testing
- `./test.sh` - Run unit tests
- `./test-integration.sh` - Run integration tests
- `go test ./... -v` - Run all tests with verbose output
- `go test ./internal/service/metric/prometheus -v` - Test specific package
- `go test ./... -v -run TestName` - Run single test

### Development
- `./install-go.sh` - Install Go automatically (macOS)
- `./fix-deps.sh` - Fix missing dependencies
- `./fix-build.sh` - Fix common build issues
- `go mod tidy` - Clean up dependencies
- `./bin/grafterm -c ./dashboard-examples/go.json` - Run with example dashboard
- `./bin/grafterm -c ./dashboard.json --debug` - Run with debug logging

### Running
- `grafterm -c ./dashboard.json` - Run with dashboard config
- `grafterm -c ./dashboard.json -d 48h` - Set relative time range
- `grafterm -c ./dashboard.json -r 2s` - Set refresh interval
- Exit with `q` or `Esc`

## Code Architecture

### Core Components

1. **cmd/grafterm** - Application entry point
   - `main.go`: Initializes app, loads configuration, sets up datasources
   - `flags.go`: Command-line flag definitions

2. **internal/controller** - Business logic layer
   - Translates view requests to model operations
   - Coordinates metric gathering via `metric.Gatherer` interface
   - Methods: `GetSingleMetric`, `GetSingleInstantMetric`, `GetRangeMetrics`

3. **internal/model** - Core data models
   - `dashboard.go`: Dashboard configuration structures
   - `datasource.go`: Datasource definitions (Prometheus, Graphite, InfluxDB)
   - `metric.go`: Metric data structures (`Metric`, `MetricSeries`)

4. **internal/service** - Service layer
   - **configuration/**: Dashboard JSON loading and parsing
   - **metric/**: Metric gathering from various datasources
     - `datasource/`: Datasource registry and routing
     - `prometheus/`: Prometheus client implementation
     - `graphite/`: Graphite client implementation
     - `influxdb/`: InfluxDB client implementation
     - `middleware/`: Logging middleware for metric operations
     - `prometheus/enhanced.go`: Enhanced Prometheus gatherer with timeout/retry logic
   - **unit/**: Unit formatting and time utilities
   - **log/**: Logging using zerolog

5. **internal/view** - Presentation layer
   - `app.go`: Main application loop with refresh/sync logic
   - **page/**: Dashboard page rendering and widget coordination
     - **widget/**: Widget implementations (gauge, graph, singlestat)
   - **render/**: Terminal rendering abstraction
     - **termdash/**: Termdash-specific rendering implementation
   - **template/**: Go template support for dashboard variables
   - **variable/**: Dashboard variable system (constants, intervals)

### Data Flow

```
main.go
  → App.Run() (view/app.go)
    → Syncer.Sync() (view/page/dashboard.go)
      → Widget rendering (view/page/widget/*)
        → Controller methods (controller/controller.go)
          → Gatherer.Gather*() (service/metric/datasource/)
            → Datasource clients (prometheus/graphite/influxdb)
```

### Key Patterns

- **Interface-based design**: `metric.Gatherer`, `render.Renderer`, `controller.Controller`
- **Middleware pattern**: Metric gathering wrapped with logging middleware
- **Context propagation**: All external calls use context for timeouts/cancellation
- **Sync loop**: App runs periodic sync operations to refresh dashboard widgets

## Error Handling & Timeouts

The codebase has enhanced error handling with timeout management:

- All external API calls use `context.WithTimeout` (2-5 seconds typical)
- Widget-level timeouts don't crash the application - graceful degradation
- Check for `context.DeadlineExceeded` and `context.Canceled` specifically
- Enhanced Prometheus gatherer (`enhanced.go`) includes retry logic and dynamic timeout scaling
- App-level sync has 5-second timeout (`view/app.go:94`)
- Always wrap errors with context: `fmt.Errorf("failed to X: %w", err)`

## Code Style

- Follow standard Go conventions: `go fmt ./...`
- Naming: CamelCase for exported, camelCase for private
- Interface names use `-er` suffix (`Gatherer`, `Renderer`, `Controller`)
- Import groups: standard → third-party → local
- Always return errors explicitly, never ignore them
- Use testify for test assertions
- Use zerolog for structured logging
- Package names: lowercase, single words, no underscores

## Dashboard Configuration

Dashboards are JSON files with three main sections:
- `datasources`: Define metric sources (Prometheus/Graphite/InfluxDB endpoints)
- `variables`: Template variables (constants, intervals)
- `dashboard.widgets`: Widget definitions (graph, gauge, singlestat)

See `docs/cfg.md` and `dashboard-examples/` for configuration details.

## Testing

- Unit tests live next to implementation files (`*_test.go`)
- Integration tests use `-tags='integration'` build tag
- Mocks generated with mockgen in `internal/mocks/`
- Test Prometheus package when changing metric gathering logic
- Run integration tests before major releases