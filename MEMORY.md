# Project Memory

## Stable Facts

- Moon Bridge is a Go 1.25 protocol conversion and model routing proxy.
- The main service lives under `cmd/moonbridge` and `internal`.
- The public-facing API includes OpenAI Responses-compatible endpoints such as `/v1/responses`.
- Upstream protocol support includes Anthropic Messages, Google Gemini, and OpenAI Chat Completions.
- Cloudflare Worker/TinyGo support is managed through `cmd/cloudflare` and `package.json`.

## Working Preferences

- Keep changes narrow and aligned with existing package patterns.
- Prefer package-local tests for behavior changes.
- Treat config, routing, protocol conversion, persistence, and streaming behavior as higher-risk areas.
- Do not assume provider secrets or local runtime config are available.

## Decisions

- `AGENTS.md`, `MEMORY.md`, and `RULES.md` are repository files and should not be ignored.

## 近期修复

- 2026-05-23: 修复 OpenAI Responses adapter — Chat Completions 嵌套工具格式不识别导致 DeepSeek 400，function_call.arguments JSON 对象反序列化失败。修改 types.go, adapter.go, adapter_test.go。


- 2026-05-23: 修复 OpenAI Responses adapter — Chat Completions 嵌套工具格式不识别导致 DeepSeek 400，function_call.arguments JSON 对象反序列化失败。修改 types.go, adapter.go, adapter_test.go。
- 2026-05-23: 添加 /v1/chat/completions 兼容入口 — 将 Chat Completions 请求转换为 Responses 格式后委托给 handleResponses，解决 sub2api 频繁覆盖 openai_responses_supported 导致 502 的问题。新增 chat_compat.go，修改 server.go。
