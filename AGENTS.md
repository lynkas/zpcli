# AGENTS.md

## Project Overview
- `zpcli` is a Go CLI-first video CMS query tool with MCP support.
- Keep CLI and MCP behavior aligned by pushing shared logic into reusable services instead of duplicating command logic.
- Prefer small, targeted changes that preserve existing command shapes and output expectations unless the task explicitly changes them.

## Repository Layout
- `cmd/`: Cobra command handlers and presentation/output helpers.
- `internal/domain/`: shared models.
- `internal/service/`: shared business logic used by CLI and MCP.
- `store/`: config persistence, schema normalization, and save/load behavior.
- `docs/`: architecture and MCP contract notes.
- `data/`: sample fixtures.

## Working Rules
- Prefer implementing behavior in `internal/service/` or `store/` when logic is shared.
- Keep `cmd/` handlers thin; they should orchestrate input/output and call shared logic.
- Preserve JSON output behavior when touching commands that support `--json`.
- Do not introduce unrelated refactors while fixing a focused issue.
- Follow existing Go formatting and naming patterns; run `gofmt` on changed Go files when possible.

## Validation
- For Go code changes, prefer targeted validation first, then broader checks:
  - `go test ./...`
  - `gofmt -w <changed files>`
- If a change affects CLI or MCP contracts, update the corresponding docs in `README.md` or `docs/`.

## Storage and Config Notes
- Config schema is versioned in `store/store.go`; preserve normalization and validation behavior.
- Be careful with save-path logic and atomic writes because config data lives outside the repo by default.
- Respect `ZPCLI_CONFIG` overrides when testing config-related behavior.

## Contributor Tips
- Review `README.md` for the supported command surface before changing command behavior.
- Review `TODO.md` for current architectural direction before making structural changes.
- Add tests near the changed logic when adjacent tests already exist.
