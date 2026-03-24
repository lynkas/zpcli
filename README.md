# ZPCLI

`zpcli` is a CLI-first video CMS query tool with MCP support for AI clients such as OpenClaw.

## What It Does

- manage configured site endpoints
- search videos across configured sources
- fetch detailed video metadata and episode links
- expose the same capabilities over MCP (stdio or SSE)

## Commands

### Site Management

- `zpcli add <domain>`
- `zpcli add <seriesId> <domain>`
- `zpcli ls`
- `zpcli rm <id>`
- `zpcli validate`
- `zpcli health`

### Search and Detail

- `zpcli search <keyword>`
- `zpcli search --series <n> --page <n> --sort <time|overlap> <keyword>`
- `zpcli detail <siteId> <vodId>`
- `zpcli <siteId> <vodId> [episode]`

### MCP

- `zpcli mcp`
- `zpcli mcp --port 8080`

## Configuration

By default, site configuration is stored at:

- macOS: `$HOME/Library/Application Support/zpcli/sites.json`

You can override the config path with:

- `ZPCLI_CONFIG=/custom/path/sites.json`

Current config schema:

```json
{
  "version": 1,
  "series": [
    {
      "domains": [
        {
          "url": "example.com",
          "failure_score": 0
        }
      ]
    }
  ]
}
```

## MCP Tools

Current tool set includes:

- `search`
- `search_videos`
- `get_detail`
- `get_video_detail`
- `list_sites`
- `add_site`
- `remove_site`
- `validate_sites`
- `health_check`

For tool details, see [`/Users/cat/zpcli/docs/MCP_CONTRACT.md`](/Users/cat/zpcli/docs/MCP_CONTRACT.md).

## Notes

- The project is being refactored toward shared services so CLI and MCP use the same core logic.
- `go test` and `gofmt` are expected project workflows, but availability depends on the local shell environment.
