# Tool adapters

The workflow lives in `AGENTS.md`, `workflow/`, and `.agents/skills/`.
Adapters only point each tool at it. `flow-setup` copies the ones a
project needs, to the paths below.

| Tool | Copy | To | Does |
|---|---|---|---|
| Claude Code | `claude/CLAUDE.md` | `CLAUDE.md` | imports `AGENTS.md` |
| Claude Code | (symlinks) | `.claude/skills/flow-*` → `../../.agents/skills/flow-*` | Claude Code reads skills only from `.claude/skills/` |
| Claude Code | `claude/agents/` | `.claude/agents/` | roles as subagents (Cursor reads these too) |
| Claude Code | `claude/settings.json` | `.claude/settings.json` (merge) | asks the human before approve, reject, wontfix |
| Codex | `codex/agents/` | `.codex/agents/` | roles as subagents |
| Codex | `codex/rules/flow.rules` | `.codex/rules/flow.rules` | asks the human before approve, reject, wontfix |
| Cursor | `cursor/hooks.json`, `cursor/hooks/` | `.cursor/hooks.json`, `.cursor/hooks/` | blocks the agent from approve and reject |

Codex and Cursor read `AGENTS.md` and `.agents/skills/` without adapters.

## Notes

- **Windows:** git symlinks need extra setup. Without them, copy the skill
  folders into `.claude/skills/` instead (only needed for Claude Code).
- **Codex:** project agents and rules load only when the project is trusted.
  Rules are experimental. Agents inherit the parent's sandbox, so they cannot
  be made read-only per agent; `flow review submit` enforces it instead.
- **Cursor:** uses `deny`, not `ask`: the human types `flow approve` in their
  own terminal. The guard needs Node. Wontfix is not blocked here, because
  workers may wontfix medium and low findings; `flow` refuses high ones.
- **Same names everywhere** (`flow-worker`, …), so Cursor, which reads both
  `.claude/agents/` and `.codex/agents/`, sees each role once.
- **Not yet:** Stop hooks that run `flow check`. Add them with the CLI.
