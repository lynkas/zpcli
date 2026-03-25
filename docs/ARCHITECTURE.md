# Architecture

## Current Layout

### Entry Point

- [`/Users/cat/zpcli/main.go`](/Users/cat/zpcli/main.go)

### CLI Transport

- [`/Users/cat/zpcli/cmd/root.go`](/Users/cat/zpcli/cmd/root.go)
- [`/Users/cat/zpcli/cmd/search.go`](/Users/cat/zpcli/cmd/search.go)
- [`/Users/cat/zpcli/cmd/detail.go`](/Users/cat/zpcli/cmd/detail.go)
- [`/Users/cat/zpcli/cmd/add.go`](/Users/cat/zpcli/cmd/add.go)
- [`/Users/cat/zpcli/cmd/ls.go`](/Users/cat/zpcli/cmd/ls.go)
- [`/Users/cat/zpcli/cmd/rm.go`](/Users/cat/zpcli/cmd/rm.go)
- [`/Users/cat/zpcli/cmd/health.go`](/Users/cat/zpcli/cmd/health.go)
- [`/Users/cat/zpcli/cmd/validate.go`](/Users/cat/zpcli/cmd/validate.go)
- [`/Users/cat/zpcli/cmd/render.go`](/Users/cat/zpcli/cmd/render.go)

### MCP Transport

- [`/Users/cat/zpcli/cmd/mcp.go`](/Users/cat/zpcli/cmd/mcp.go)

### Shared Domain Models

- [`/Users/cat/zpcli/internal/domain/search.go`](/Users/cat/zpcli/internal/domain/search.go)
- [`/Users/cat/zpcli/internal/domain/detail.go`](/Users/cat/zpcli/internal/domain/detail.go)
- [`/Users/cat/zpcli/internal/domain/site.go`](/Users/cat/zpcli/internal/domain/site.go)
- [`/Users/cat/zpcli/internal/domain/health.go`](/Users/cat/zpcli/internal/domain/health.go)

### Shared Services

- [`/Users/cat/zpcli/internal/service/endpoint.go`](/Users/cat/zpcli/internal/service/endpoint.go)
- [`/Users/cat/zpcli/internal/service/search.go`](/Users/cat/zpcli/internal/service/search.go)
- [`/Users/cat/zpcli/internal/service/detail.go`](/Users/cat/zpcli/internal/service/detail.go)
- [`/Users/cat/zpcli/internal/service/sites.go`](/Users/cat/zpcli/internal/service/sites.go)
- [`/Users/cat/zpcli/internal/service/health.go`](/Users/cat/zpcli/internal/service/health.go)

### Storage

- [`/Users/cat/zpcli/store/store.go`](/Users/cat/zpcli/store/store.go)

Current storage protections:

- version normalization on load
- basic config validation before save
- atomic temp-file write and rename on save

## Design Direction

The project is moving toward:

1. shared services for all business logic
2. thin CLI and MCP transport adapters
3. hardened storage and config lifecycle
4. stable MCP contracts for OpenClaw integration

## Current Shared Flow

### Search

1. CLI or MCP loads store data
2. search service ranks candidate domains
3. search service queries endpoints concurrently
4. caller updates failure scores
5. CLI/MCP rendering writes text output

### Detail

1. CLI or MCP loads store data
2. detail service resolves site id and fetches remote detail
3. detail service parses players and episodes into shared models
4. caller updates failure scores on transport-visible failures
5. CLI/MCP rendering writes text output

### Site Management

1. CLI or MCP loads store data
2. site service performs add/list/remove operations
3. store persists results

## Known Gaps

- no structured `--json` output yet
- no migration framework beyond version field
- MCP still returns text content rather than structured objects
- no automated tests beyond endpoint normalization
