# Backlog

| Item | Status | From | Added | Closed by |
|---|---|---|---|---|
| User hooks on events: a pre hook before an event runs, and a post hook after it (e.g. slice.created), with a trust check before running a repo's hooks. Maybe slice 0013. | open | 0001-slice-and-events planning talk | 2026-10-23 | - |
| Cancel a slice: `slice.cancelled` event and a cancel command (design 2.7). Not in the recorded workflow yet. | open | 0001-slice-and-events planning talk | 2026-10-23 | - |
| Who caused an event: a `by` field, detected from the agent tool's environment (e.g. `CLAUDECODE`, `CODEX_THREAD_ID`, `CURSOR_AGENT`), with the session ID when the tool gives one (`CLAUDE_CODE_SESSION_ID`, `CODEX_THREAD_ID`), else `human`; maybe a `--by` override. Not proof: lead and subagents share one Claude session ID. With the Claude Code integration. | open | 0001-slice-and-events planning talk | 2026-10-24 | - |
| A `status.md` written from `.events.jsonl`, for reading the status on GitHub, if we still want one after the dashboard. 0001 shows the status with `sliceflow slice status` only. (Changes the design's 2026-10-18 decision: event log plus `status.md`.) | open | 0001-slice-and-events planning talk | 2026-10-24 | - |
| A local server (`sliceflow serve`, like Theia): an HTTP API and a live channel (WebSocket or SSE), a browser app for the dashboard and actions such as starting agents, on 127.0.0.1 with a token and Host/Origin checks. It calls the same `internal/` core as the CLI. WebSocket would be a new dependency. | open | 0001-slice-and-events planning talk | 2026-10-24 | - |
