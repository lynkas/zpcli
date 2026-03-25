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

- bump the `VERSION` file
- update `CHANGELOG.md`
- update `README.md` if command surface changed
- update `docs/MCP_CONTRACT.md` if MCP tool behavior changed
- update release and changelog notes if a roadmap milestone moved
- confirm Linux artifacts build with `make release-artifacts`
- run `go test ./...` if Go is available
- run formatting if Go is available
- verify `zpcli health`, `zpcli validate`, and `zpcli doctor`
- verify `zpcli mcp` tool list output
- verify config load/save against a test config path

## Automated GitHub Release

- update `VERSION`, then create and push a matching tag such as `v0.1.0`
- GitHub Actions workflow `.github/workflows/release.yml` will:
  - run `go test ./...`
  - build Linux binaries for `amd64` and `arm64`
  - generate `dist/checksums.txt`
  - publish a GitHub Release with the Linux artifacts attached

## Artifact Naming

- `zpcli_linux_amd64`
- `zpcli_linux_arm64`
- `checksums.txt`

## Upgrade Notes

- use MCP tool names such as `search_videos` and `get_video_detail`
- document config migration changes before release
