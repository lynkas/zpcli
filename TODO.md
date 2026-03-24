# ZPCLI Maturation Roadmap

## Goal

Turn `zpcli` into a mature, production-ready tool platform that:

- remains a great standalone CLI for human users
- exposes stable MCP interfaces for OpenClaw and other AI clients
- separates transport, business logic, and storage concerns
- can be operated, tested, and evolved without large rewrites

This file is intended to be the working execution plan for the project.

## Guiding Principles

- CLI-first: the binary should always remain useful on its own
- MCP-native: AI integration should be a first-class interface, not an afterthought
- Single source of truth: core logic should live in shared services, not in command handlers
- Backward-aware: avoid breaking current CLI usage unless the migration path is clear
- Observable by default: failures should be diagnosable
- Safe to operate: mutating actions should be explicit, auditable, and reversible where possible

## Current State Summary

Today the project already has:

- a working Cobra-based CLI
- JSON file storage for configured sites
- search and detail retrieval logic
- a basic MCP server over stdio and SSE

Current limitations:

- business logic is embedded inside CLI command handlers
- MCP behavior is not yet treated as a stable contract
- storage is simple and not hardened for migrations, locking, or recovery
- there is no health, audit, diagnostics, or compatibility strategy
- transport, formatting, and domain logic are tightly coupled

## Target Architecture

The project should evolve toward these layers:

1. Core domain and service layer
- search, detail, site management, health checks, validation, ranking
- transport-agnostic interfaces and models

2. Adapters / transport layer
- CLI adapter
- MCP stdio adapter
- MCP SSE/HTTP adapter

3. Storage and configuration layer
- site config repository
- schema versioning and migrations
- safe writes and file locking
- optional secrets handling

4. Operations layer
- structured logs
- health and doctor commands
- audit trail
- metrics hooks
- compatibility and release discipline

## Execution Plan

### Phase 0: Project Baseline

Objective:
Create a clean baseline before larger refactors begin.

Tasks:

- Document the current commands, flags, and behaviors.
- Record the current MCP tools, inputs, and outputs.
- Capture sample config files and expected search/detail responses.
- Define supported Go version and runtime assumptions.
- Decide the minimum backward-compatibility promise for the first stable milestone.

Deliverables:

- `README` update describing current CLI and MCP usage
- sample config fixture
- sample MCP transcript fixture
- architecture note describing current code paths

Acceptance criteria:

- A new contributor can understand how the tool currently works without reading all source files.
- Existing CLI behavior is documented before refactoring begins.
- Existing MCP behavior is documented before any schema changes are made.

### Phase 1: Extract a Shared Core

Objective:
Move business logic out of `cmd/` into a shared service layer.

Progress notes:

- 2026-03-23: extracted shared search, detail, and site-management services
- 2026-03-23: added reusable CLI/MCP text rendering helpers so MCP no longer depends on command handlers for search and detail

Tasks:

- Create internal packages for domain models and services.
- Extract search logic from `cmd/search.go` into a dedicated service.
- Extract detail lookup logic from `cmd/detail.go` into a dedicated service.
- Extract site add/remove/list logic from command handlers into services.
- Move endpoint normalization into a reusable component.
- Introduce context-aware service methods for cancellation and timeouts.
- Define typed result models shared across CLI and MCP.
- Define typed domain errors shared across CLI and MCP.

Suggested package direction:

- `internal/domain`
- `internal/service`
- `internal/transport/cli`
- `internal/transport/mcp`
- `internal/store`

Deliverables:

- command handlers become thin adapters
- reusable service interfaces
- shared DTO/domain models

Acceptance criteria:

- CLI commands call shared services instead of implementing business logic inline.
- MCP tool handlers call the same services as CLI commands.
- Search/detail/site management behavior remains functionally equivalent to current behavior.

### Phase 2: Stabilize the MCP Contract

Objective:
Turn MCP from a convenience feature into a supported product interface for OpenClaw.

Progress notes:

- 2026-03-23: added stable MCP tool aliases for search and detail
- 2026-03-23: added `validate_sites` and `health_check` MCP tools

Tasks:

- Inventory all existing MCP tools and decide which names are stable.
- Standardize tool naming and descriptions.
- Define consistent input schemas for every tool.
- Define consistent output structures for every tool.
- Introduce structured error categories for MCP responses.
- Separate human-readable output from machine-friendly output.
- Add version metadata for the MCP server.
- Add capability documentation for supported tools.
- Decide whether mutating tools require confirmation or explicit flags.
- Add request timeouts and defensive validation for MCP calls.

Recommended MCP tool set for the first mature milestone:

- `search_videos`
- `get_video_detail`
- `list_sites`
- `add_site`
- `remove_site`
- `validate_sites`
- `health_check`

Deliverables:

- MCP contract document
- stable tool schemas
- compatibility policy for tool changes

Acceptance criteria:

- OpenClaw can rely on tool names and schemas without frequent breakage.
- CLI-only formatting is not leaked into MCP payload design.
- Errors returned by MCP are actionable and consistent.

### Phase 3: Harden Storage and Configuration

Objective:
Make local state reliable enough for long-term use and automation.

Tasks:

- Add a formal config schema version.
- Implement migrations between schema versions.
- Use atomic write strategy with temporary files and rename.
- Add file locking to reduce corruption during concurrent writes.
- Validate site records before saving.
- Improve error messages for invalid config paths and malformed data.
- Add import/export commands or APIs.
- Consider separating secrets from the main config if needed later.
- Add config backup before destructive migrations.

Possible future config fields:

- site label
- endpoint type
- health metadata
- tags
- disabled/enabled state
- notes
- authentication or custom headers

Deliverables:

- versioned config model
- migration code
- config validation logic
- backup/recovery strategy

Acceptance criteria:

- Config survives interrupted writes without silent corruption.
- Older config files can be migrated forward safely.
- Concurrent CLI/MCP mutations do not trivially destroy data.

### Phase 4: Add Operational Maturity

Objective:
Make the project diagnosable and manageable in real use.

Progress notes:

- 2026-03-23: added CLI `health` and `validate` commands backed by a shared health service

Tasks:

- Add structured logging with levels.
- Add request IDs or operation IDs.
- Add `zpcli health` command.
- Add `zpcli doctor` command for environment and config diagnostics.
- Add `zpcli config validate`.
- Add latency and failure metrics hooks.
- Add audit logs for mutating operations.
- Add configurable timeouts and retry policy.
- Add a dry-run mode where applicable.
- Add site health scoring and validation tooling.

Deliverables:

- health and doctor commands
- logging strategy
- diagnostics documentation

Acceptance criteria:

- Common failures can be diagnosed from logs and commands.
- Operators can validate configuration before using the MCP server.
- Mutations can be traced after the fact.

### Phase 5: Improve UX for Both Humans and AI Clients

Objective:
Make the tool pleasant and predictable for both shell users and OpenClaw.

Tasks:

- Improve command help text and examples.
- Add consistent exit codes for common failure types.
- Standardize CLI output formatting.
- Add machine-readable output option such as `--json`.
- Add explicit quiet/verbose modes.
- Add clearer series/site identifiers and display labels.
- Decide on a stable user-facing vocabulary.
- Improve episode lookup ergonomics.
- Add site validation during `add` if safe and useful.

Deliverables:

- CLI UX guidelines
- JSON output support
- improved help text

Acceptance criteria:

- Human users can script the CLI reliably.
- AI clients do not need to parse table-like text when structured output is available.
- Error messaging is clear enough to act on quickly.

### Phase 6: Testing Strategy

Objective:
Build confidence so changes can be made without fear.

Tasks:

- Add unit tests for domain and service logic.
- Add fixture-driven tests for search/detail parsing.
- Add storage tests for load/save/migration cases.
- Add concurrency-oriented tests for write safety.
- Add MCP handler tests with request/response fixtures.
- Add CLI smoke tests for key commands.
- Add golden tests for human-readable output where useful.
- Add end-to-end tests for MCP stdio mode.
- Add regression tests for failure score behavior and ranking.

Deliverables:

- test matrix
- fixture directories
- CI test command

Acceptance criteria:

- Core services have focused unit coverage.
- MCP behavior is protected by fixtures.
- Key user flows are exercised in automated tests.

### Phase 7: Release Discipline and Compatibility

Objective:
Make the project safe to adopt as a dependency in OpenClaw workflows.

Tasks:

- Define versioning strategy.
- Decide what counts as a breaking change for CLI and MCP separately.
- Add changelog process.
- Add release checklist.
- Add compatibility notes for config migrations.
- Add example OpenClaw integration docs.
- Add packaged binary and container guidance.

Deliverables:

- release policy
- changelog template
- integration documentation

Acceptance criteria:

- Users can upgrade with confidence.
- MCP consumers can understand compatibility expectations.
- Releases are repeatable and documented.

## Suggested Implementation Order

This is the recommended practical order of work:

1. Phase 0 baseline documentation
2. Phase 1 shared core extraction
3. Phase 2 MCP contract stabilization
4. Phase 3 storage hardening
5. Phase 6 tests, started in parallel and expanded continuously
6. Phase 4 operational maturity
7. Phase 5 UX polish
8. Phase 7 release discipline

Notes:

- Testing should begin as soon as Phase 1 starts, not only after refactors are complete.
- MCP contract decisions should happen before large OpenClaw-facing integrations are published.
- Storage hardening should happen before encouraging concurrent CLI and MCP usage.

## Immediate Next Tasks

These are the best next actions from the current repository state:

- Create a short architecture note mapping current files to future layers.
- Extract endpoint normalization into a service/helper package with tests.
- Extract search logic into a shared service with typed inputs/outputs.
- Extract detail logic into a shared service with typed inputs/outputs.
- Refactor `cmd/mcp.go` to call shared services instead of command functions.
- Add the first MCP contract document with current and target tool schemas.
- Add a `--json` mode for at least one read-only command to establish the pattern.

## Definition of Done for the First Mature Milestone

The project can be considered "mature enough for serious OpenClaw integration" when:

- CLI and MCP both use the same shared service layer
- MCP tool names and schemas are intentionally versioned and documented
- config storage is versioned, validated, and safely written
- key flows are covered by automated tests
- `health`, `doctor`, and config validation commands exist
- structured logs and actionable errors are available
- release and compatibility expectations are documented

## Risks and Watchouts

- Refactoring too much at once without preserving behavior
- Keeping human-readable formatting coupled to machine-facing APIs
- Allowing MCP schemas to drift casually over time
- Expanding config shape without migration planning
- Underestimating the need for diagnostics once OpenClaw starts depending on the tool
- Treating SSE and stdio as separate products instead of two transports over one contract

## Working Rules for Execution

When implementing tasks from this file:

- Prefer small, reviewable refactors
- Add tests when extracting behavior
- Preserve current CLI behavior unless the change is intentional and documented
- Treat MCP changes as contract changes, not internal refactors
- After every code change, create a commit and push it once remote push is configured
- Update this file as milestones are completed or split

## Status Tracking Template

Use this checklist style when executing the roadmap:

- [ ] Phase 0 complete
- [ ] Phase 1 complete
- [ ] Phase 2 complete
- [ ] Phase 3 complete
- [ ] Phase 4 complete
- [ ] Phase 5 complete
- [ ] Phase 6 complete
- [ ] Phase 7 complete

For sub-work, append dated notes under the relevant phase.
