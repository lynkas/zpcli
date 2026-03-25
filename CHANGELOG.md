# Changelog

All notable changes to `zpcli` should be documented in this file.

## Unreleased

### Added

- Shared search, detail, site-management, and health services
- MCP stable aliases and health / validation tools
- CLI `health`, `validate`, and `doctor` commands
- Global `--json` output mode for major CLI commands
- Atomic config write behavior and basic config validation
- Baseline architecture, MCP, and sample fixture documentation

### Changed

- CLI and MCP now rely on shared service logic instead of command-only implementations

### Known Limitations

- `go test` and `gofmt` could not be run in the current shell environment when these changes were made
- MCP still returns text-oriented content rather than fully structured result payloads
