
## 找个片看吗？找个片看吧！澳门首家片站上线啦！

# ZPCLI

## 你需要：

- 找到能用的采集站api的域名
- 配置采集站链接
- 开始搜索
- 点击返回的m3u8并观看

## 你甚至可以

- 跑一个mcp服务器，让agent帮你搜。

`zpcli` is a CLI-first video CMS query tool with MCP support for AI clients such as OpenClaw.

## What It Does

- manage configured site endpoints
- search videos across configured sources
- fetch detailed video metadata and episode links
- expose the same capabilities over MCP (stdio or SSE)

## Commands

### Site Management

- `zpcli site add <domain>`
- `zpcli site add <seriesId> <domain>`
- `zpcli site ls`
- `zpcli site rm <id>`
- `zpcli site validate`
- `zpcli site health`
- `zpcli doctor`

All major CLI commands now support:

- `zpcli --json ...`

### Search and Detail

- `zpcli search <keyword>`
- `zpcli search --series <n> --page <n> --sort <time|overlap> <keyword>`
- `zpcli detail <siteId> <vodId>`
- `zpcli <siteId> <vodId> [episode]`

### MCP

- `zpcli mcp`
- `zpcli mcp --port 8080`

## Build

Minimum Go version:

- Go `1.21` or newer

Local development build:

- `make build`

Show build metadata:

- `zpcli version`
- `zpcli --json version`

Linux release artifacts:

- `make release-artifacts`

Produced files:

- `dist/zpcli_linux_amd64`
- `dist/zpcli_linux_arm64`
- `dist/checksums.txt`

GitHub Releases are published automatically when a tag like `v0.1.0` is pushed.

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
- MCP `search_videos` supports paging with an optional `page` argument.
- For MCP clients, if page 1 misses an expected result, try later pages with the same keyword before changing the query.
- For MCP clients, do not expand a short core keyword with extra words when searching.
- Config saves now use atomic write semantics to reduce the chance of partial file corruption.
- `go test` and `gofmt` are expected project workflows, but availability depends on the local shell environment.
