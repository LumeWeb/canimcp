# AGENTS.md

This file provides development guidelines and architectural documentation for
the canimcp project.

## Common Commands

### Building
```bash
# Build all packages
go build -v ./...
```

### Testing
```bash
# Run all tests with race detection and coverage
go test -v -race -coverprofile=coverage.out -covermode=atomic ./...

# Run tests for the root package only
go test -v -race .

# View coverage report
go tool cover -func=coverage.out
```

### Mock Generation

```bash
# Generate mocks for interfaces (uses .mockery.yaml; mockery is
# pre-installed at $HOME/go/bin/mockery; never reinstall it)
mockery
```

### Dependency Management
```bash
# Download dependencies
go mod download

# Verify dependencies
go mod verify

# Tidy dependencies
go mod tidy
```

## Project Overview

canimcp is a caniuse-style compatibility database and detector for MCP
hosts, transports, and wire capabilities. A static `Profile` declares which
features a host + transport combination supports; ordered `Detector`s resolve
runtime wire signals (clientInfo, headers, auth) onto that declaration, and
an unmatched request degrades to a generic profile for the transport.

## Architecture

### Package Structure
- **Module path**: `go.lumeweb.com/canimcp`
- **Root package** (`canimcp`): all functionality as a flat package;
  `doc.go` holds the package documentation, one file per host detector
  (`claude.go`, `grok.go`, ...)

### Design
- **Profile = static declaration + runtime overlay**: transport-mechanism
  features (source/sink/reachability) are derived from the transport and can
  never drift from it; genuine per-host capability features are declared per
  host profile.
- **Ordered detectors, first match wins**: each detector matches exactly one
  host from wire signals and is registered in a `DetectorRegistry` by
  priority; a fallback emits a generic profile for the transport.
- **Runtime overlay is explicit**: a static profile's `FeatureSet` is a
  shared map — callers that overlay runtime flags MUST `CloneFeatures` first.
- **Predicate abstraction**: `Predicate` (with `HostIs`, `TransportIs`, `Not`,
  `And`) covers gates that are not features, so conditional DSLs can gate on
  the profile without importing product code.
- **Boundary**: the core package must not import an MCP SDK, a terminal UI,
  or any product/deployment concern. SDK-specific evidence extraction belongs
  in an adapter package (e.g. `canimcp/mcpsdk`), never the core.

### Testing Conventions
- Tests are colocated next to source (`*_test.go`)
- Tests pin current behavior (characterization) — detector priorities,
  profile feature sets, and transport-derived mechanisms are
  regression-guarded; do not change semantics silently
- Do not add a README/board copyright header to source files; attribution
  lives only in the LICENSE file
- Generated mocks live in `mocks/` (mockery, testify templates)
