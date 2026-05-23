# Coding Agent Instructions

## Project Context

Moon Bridge is a Go 1.25 protocol conversion and model routing proxy. It exposes OpenAI Responses API endpoints and routes requests to upstream Anthropic Messages, Google Gemini, and OpenAI Chat Completions providers. There is also a small Cloudflare Worker/TinyGo toolchain managed through `package.json`.

## Companion Files

- Read `MEMORY.md` for stable project facts and decisions that should persist across sessions.
- Read `RULES.md` for repository-specific development rules before making code changes.

## Operating Principles

- Keep changes small, direct, and maintainable.
- Assume the user is a principal engineer; be concise and technically precise.
- Prefer existing patterns in the surrounding package over new abstractions.
- Avoid broad refactors, speculative helpers, and unrelated cleanup.
- Read the narrowest relevant code first, then patch the failing path.
- Preserve user changes in the working tree. Do not revert unrelated edits.

## Repository Conventions

- Use `rg` and scoped file reads for exploration.
- Keep command output byte-capped when output size is unknown.
- Use `apply_patch` for manual edits.
- Default to ASCII in new or edited files unless the file already uses non-ASCII text or the content requires it.
- Favor clear Go names, simple control flow, strong types, and package-local behavior.
- Keep tests close to the touched package when adding or changing behavior.

## Common Commands

Prefer the cheapest command that validates the changed surface:

```bash
go test ./internal/protocol/openai
go test ./internal/service/server
go test ./cmd/moonbridge
```

Use broader validation only when the risk justifies it:

```bash
go test ./...
```

Cloudflare Worker commands are available through pnpm/npm scripts when working on `cmd/cloudflare` or Worker assets:

```bash
pnpm build
pnpm dev
pnpm deploy
```

## Validation Guidance

- Skip validation for documentation-only or very low-risk edits, and say so plainly.
- For Go behavior changes, run the smallest relevant `go test` package target.
- Run `go test ./...` only for cross-cutting changes, public protocol behavior, routing, persistence, or config loading changes.
- Do not run full builds or broad test suites by default.

## Communication

- State the approach before non-trivial edits.
- During longer work, report only what was found, what changed, and any remaining risk.
- Final summaries should include files touched and validation run or skipped.
