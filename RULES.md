# Project Rules

## Development

- Make the smallest maintainable change that solves the request.
- Prefer existing code patterns over introducing new abstractions.
- Avoid unrelated refactors, formatting churn, and broad cleanup.
- Preserve user changes in the working tree.
- Use `apply_patch` for manual edits.

## Exploration

- Use `rg` for search.
- Scope file reads and command output to the relevant package or symbol.
- Byte-cap commands with unknown or potentially large output.
- Avoid dumping generated, binary, minified, database, or large JSON files unless required.

## Go Code

- Favor simple control flow, clear names, strong types, and small functions.
- Keep behavior close to the package that owns it.
- Add tests next to touched code when changing behavior.
- Avoid single-use helpers unless they clearly reduce complexity.

## Validation

- Documentation-only changes do not require tests.
- For behavior changes, run the narrowest useful `go test` target.
- Use `go test ./...` only for cross-cutting changes or when local package tests are insufficient.
- Do not run full builds or broad test suites by default.
