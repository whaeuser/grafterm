# Grafterm [![CircleCI][circleci-image]][circleci-url] [![Go Report Card][go-reportcard-image]][go-reportcard-url]

Visualize metrics dashboards on the terminal, like a simplified and minimalist version of [Grafana] for terminal.

![grafterm red dashboard](/img/grafterm-red-compressed.gif)

## Features

- Multiple widgets (graph, singlestat, gauge).
- Multiple datasources usage.
- User stored datasources.
- Override dashboard datasource ID to different datasource ID configured by the user.
- Custom dashboards based on JSON configuration files.
- Extensible metrics datasource implementation (Prometheus and Graphite included).
- Templating of variables.
- Auto time interval adjustment for queries.
- Auto unit formatting on widgets.
- Fixed and adaptive grid.
- Color customization on widgets.
- Configurable autorefresh.
- Single binary and easy usage/deployment.
- **Enhanced error handling with graceful timeout management**.
- **Multiple build options (Docker-free, Docker, simple scripts)**.

## Installation

### Option 1: Download Binaries
Download the pre-built binaries from [releases]

### Option 2: Build from Source

#### Quick Start (Recommended)
```bash
# Clone the repository
git clone https://github.com/slok/grafterm.git
cd grafterm

# Install Go automatically (macOS) or use alternatives
./install-go.sh

# Build the binary
./build.sh

# Run grafterm
./bin/grafterm -c ./dashboard-examples/go.json
```

#### Alternative Build Methods

**Without Go Installation (Docker):**
```bash
# Build using Docker
./build-docker.sh
```

**Using Simple Makefile:**
```bash
# Build for current platform
make -f Makefile.simple build

# Build for multiple platforms
make -f Makefile.simple build-all

# Install to GOPATH/bin
make -f Makefile.simple install
```

**Manual Go Installation:**
```bash
# Install Go with Homebrew (macOS)
brew install go

# Or download from https://golang.org/dl/
# Then build:
go build -o bin/grafterm ./cmd/grafterm
```

## Development

### Setup Development Environment
```bash
# Install Go
./install-go.sh

# Fix dependencies
./fix-deps.sh

# Run tests
./test.sh
./test-integration.sh

# Build and run
./build.sh
./bin/grafterm -c ./dashboard-examples/go.json
```

### Build Commands
| Command | Description |
|---------|-------------|
| `./build.sh` | Build binary for current platform |
| `make -f Makefile.simple build` | Build using simple Makefile |
| `make -f Makefile.simple build-all` | Build for multiple platforms |
| `./build-docker.sh` | Build with Docker (no Go required) |
| `./test.sh` | Run unit tests |
| `./test-integration.sh` | Run integration tests |
| `make -f Makefile.simple test` | Run unit tests |
| `make -f Makefile.simple test-integration` | Run integration tests |

### Troubleshooting
```bash
# If Go is not installed
./no-go-help.sh  # Shows alternatives

# If dependencies are missing
./fix-deps.sh     # Fixes missing dependencies

# If build fails
./fix-build.sh    # Fixes common build issues
```

## Running options

Exit with `q` or `Esc`

### Simple

```bash
grafterm -c ./mydashboard.json
```

### Relative time

```bash
grafterm -c ./mydashboard.json -d 48h
```

### Refresh interval

```bash
grafterm -c ./mydashboard.json -r 2s
```

### Debugging

When grafterm doesn't show anything may be that has errors getting metrics or similar. There is available a `--debug` flag that will write a log on `grafterm.log` (this path can be override with `--log-path` flag)

**Note:** The application now includes enhanced error handling and timeout management. Network issues or slow responses won't crash the application - it will gracefully degrade and log timeout errors for debugging.

Read the log

```bash
tail -f ./grafterm.log
```

And run grafterm in debug mode.

```bash
grafterm -c ./mydashboard.json  -d 48h -r 2s --debug
```

### Fixed time

Setting a fixed time range to visualize the metrics using duration notation. In this example is start at `now-22h` and end at `now-20h`

```bash
grafterm -c ./mydashboard.json -s 22h -e 20h
```

Setting a fixed time range to visualize the metrics using timestamp [ISO 8601] notation.

```bash
grafterm -c ./mydashboard.json -s 2019-05-12T12:32:11+02:00 -e 2019-05-12T12:35:11+02:00
```

### Replacing dashboard variables

```bash
grafterm -c ./mydashboard.json -v env=prod -v job=envoy
```

### Replacing dashboard datasource configuration

Replace dashbaord `prometheus` datasource with user datasource `thanos-prometheus` (check [Datasources](#datasources) section):

```bash
grafterm -c ./mydashboard.json -a "prometheus=thanos-prometheus"
```

Replace dashboard `prometheus` datasource with user datasource `thanos-prometheus` available on `/tmp/my-datasources.json` user datasource configuration file:

```bash
grafterm -c ./mydashboard.json -a "prometheus=thanos-prometheus" -u /tmp/my-datasources.json
```

## Error Handling & Reliability

The application has been enhanced with robust error handling:

- **Timeout Management**: All external API calls include proper timeouts (2-5 seconds)
- **Graceful Degradation**: Network timeouts don't crash the application
- **Context Propagation**: Proper context usage throughout the call chain
- **Error Logging**: Enhanced logging for debugging timeout issues
- **Widget Resilience**: Individual widget timeouts don't affect other widgets

### Common Issues and Solutions

| Issue | Solution |
|-------|----------|
| Build fails with "go: command not found" | Run `./install-go.sh` or `./no-go-help.sh` |
| Missing dependencies | Run `./fix-deps.sh` |
| Import conflicts | Run `./fix-build.sh` |
| Network timeouts | Check logs with `--debug` flag |
| Build errors | Check CRUSH.md for detailed guidelines |

## Dashboard

Check [this][cfg-md] section that explains how a dashboard is configured. Also check [dashboard examples][dashboard-examples]

## Datasources

Datasources are the way grafterm knows how to retrieve the metrics for the dashboard.

check available types and how to configure in [this][cfg-md] section.

**If you want support for a new datasource type, open an issue or send a PR**

### Overriding dashboard datasources

Dashboard referenced datasources on the queries can be override.

#### User datasource

Grafterm dashboards can have default datasources but the user can override these datasources using a datasources config file. This file has the same format as the dashboard configuration file but will ignore anything other than the `datasources` block. Example:

```json
{
  "version": "v1",
  "datasources": {
    "prometheus": {
      "prometheus": { "address": "http://127.0.0.1:9090" }
    },
    "localprom": {
      "prometheus": { "address": "http://127.0.0.1:9091" }
    },
    "thanos": {
      "prometheus": { "address": "http://127.0.0.1:9092" }
    },
    "m3db": {
      "prometheus": { "address": "http://127.0.0.1:9093" }
    },
    "victoriametrics": {
      "prometheus": { "address": "http://127.0.0.1:8428" }
    },
    "wikimedia": {
      "graphite": { "address": "https://graphite.wikimedia.org" }
    }
  }
}
```

If the dashboard has defined a datasource configuration with the ID `my-ds` reference, and the user datasources has this same datasource ID, grafterm will use the user defined one when the queries in the dashboard reference this ID.

The user datasources location can be configured with this priority (from highest to lowest):

- If `--user-datasources` explicit flag is used, it will use this.
- If `GRAFTERM_USER_DATASOURCES` env var is set, it will use this.
- As a fallback location will check `{USER_HOME}/grafterm/datasources.json` exists.

#### Alias

Apart from overriding the dashboard datasources IDs that match with the user datasources, the user can force an alias with the form `dashboard-ds-id=user-ds-id`.

For example, the dashboard uses a datasource named `prometheus-2b`, and we want to use our local prometheus configured on the user datasources as `localprom`, we could use the alias flag like this: `-a "prometheus-2b=localprom"`, now every query the dashboard widgets make to `prometheus-2b` will be made to `localprom`.

## Kudos

This project would not be possible without the effort of many people and projects but specially [Grafana] for the inspiration, ideas and the project itself, and [Termdash] for the rendering of all those fancy graphs on the terminal.

## Contributing

Contributions are welcome! Please check [CRUSH.md](./CRUSH.md) for development guidelines, build instructions, and code style standards.

### Development Setup
1. Fork the repository
2. Clone your fork: `git clone https://github.com/yourusername/grafterm.git`
3. Setup development environment: `./install-go.sh`
4. Create a feature branch: `git checkout -b feature-name`
5. Make your changes and run tests: `./test.sh`
6. Build and test: `./build.sh && ./bin/grafterm -c ./dashboard-examples/go.json`
7. Commit your changes: `git commit -m "Add feature"`
8. Push to the branch: `git push origin feature-name`
9. Create a Pull Request

### Build System
The project uses a flexible build system with multiple options:
- **Simple scripts**: `./build.sh`, `./test.sh`
- **Makefile**: `make -f Makefile.simple`
- **Docker support**: `./build-docker.sh`
- **Dependency management**: `./fix-deps.sh`

See [CRUSH.md](./CRUSH.md) for detailed build instructions and troubleshooting.

[circleci-image]: https://img.shields.io/circleci/project/github/slok/grafterm/master.svg
[circleci-url]: https://circleci.com/gh/slok/grafterm
[go-reportcard-image]: https://goreportcard.com/badge/github.com/slok/grafterm
[go-reportcard-url]: https://goreportcard.com/report/github.com/slok/grafterm
[grafana]: https://grafana.com/
[termdash]: https://github.com/mum4k/termdash
[releases]: https://github.com/slok/grafterm/releases
[cfg-md]: /docs/cfg.md
[dashboard-examples]: /dashboard-examples
[iso 8601]: https://en.wikipedia.org/wiki/ISO_8601
[prometheus]: http://prometheus.io
