# Changelog

All notable changes to `zpcli` should be documented in this file.

## Unreleased

## v0.1.0

### Added

- Unified `site` command group for site management operations
- Stable MCP tool set: `search_videos`, `get_video_detail`, `list_sites`, `add_site`, `remove_site`, `validate_sites`, and `health_check`
- `version` command with build metadata output
- Opt-in structured logging with curl-style verbosity flags: `-v` and `-vv`
- Linux release artifact build targets and automated GitHub Release workflow
- Shared search, detail, site-management, and health services
- Global `--json` output mode for major CLI commands
- Atomic config write behavior and basic config validation
- Baseline architecture and MCP contract documentation

### Changed

- Site management now uses `zpcli site ...` as the only supported CLI entry point
- MCP documentation now focuses on stable tool names and includes detailed input/output guidance
- Search result rendering no longer shows the `score` column
- CLI and MCP now rely on shared service logic instead of command-only implementations

### Removed

- Legacy CLI site-management command aliases at the root level
- Legacy MCP tool aliases `search` and `get_detail`
- Unused Docker-related files and `TODO.md`

### Fixed

- Search output layout cleanup after removing the score column
- Release process now has reproducible Linux artifact output and checksums

### Upgrade Notes

- Use `zpcli site add`, `zpcli site ls`, `zpcli site rm`, `zpcli site validate`, and `zpcli site health`
- Use MCP tool names `search_videos` and `get_video_detail`
- Use `zpcli version` to confirm the installed binary version before debugging plugin integration
