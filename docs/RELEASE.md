# Release Notes

## Versioning Direction

- Treat CLI behavior and MCP behavior as separate compatibility surfaces.
- Prefer additive changes for MCP tools.
- Avoid renaming stable MCP tools after publication.

## Breaking Change Guidance

The following should be treated as breaking:

- removing or renaming stable MCP tools
- changing required MCP input fields
- changing config schema without migration support
- changing CLI flags or output semantics without documentation

## Release Checklist

- update `CHANGELOG.md`
- update `README.md` if command surface changed
- update `docs/MCP_CONTRACT.md` if MCP tool behavior changed
- update `TODO.md` progress notes if a roadmap milestone moved
- run `go test ./...` if Go is available
- run formatting if Go is available
- verify `zpcli health`, `zpcli validate`, and `zpcli doctor`
- verify `zpcli mcp` tool list output
- verify config load/save against a test config path

## Upgrade Notes

- use MCP tool names such as `search_videos` and `get_video_detail`
- document config migration changes before release
