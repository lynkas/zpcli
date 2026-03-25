# OpenClaw Integration

## Recommended Model

Use `zpcli` as:

- a standalone CLI for human operators
- an MCP server for OpenClaw and other AI clients

This keeps the core logic in one project while preserving terminal-native workflows.

## Preferred MCP Tool Names

Use these names in new OpenClaw integrations:

- `search_videos`
- `get_video_detail`
- `list_sites`
- `add_site`
- `remove_site`
- `validate_sites`
- `health_check`

## Startup Options

### Stdio

Recommended for local agent integration:

- `zpcli mcp`

### SSE

Recommended when running the tool as a long-lived service:

- `zpcli mcp --port 8080`

## Operational Guidance

- call `health_check` before relying on the tool in a long-running workflow
- call `validate_sites` after mutating site configuration
- prefer CLI `--json` for shell automation outside MCP
