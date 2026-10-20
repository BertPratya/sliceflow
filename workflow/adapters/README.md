# Tool adapters

The workflow lives in `AGENTS.md`, `workflow/`, and `templates/`. Adapters
only point each tool at it. They are copied to the paths below.

| Tool | Copy | To | Does |
|---|---|---|---|
| Claude Code | `claude/CLAUDE.md` | `CLAUDE.md` | imports `AGENTS.md` |
| Claude Code | `claude/agents/` | `.claude/agents/` | roles as subagents (Cursor reads these too) |
| Codex | `codex/agents/` | `.codex/agents/` | roles as subagents |

Codex and Cursor read `AGENTS.md` without adapters.

## Notes

- **Codex:** project agents load only when the project is trusted. Agents
  inherit the parent's sandbox, so they cannot be made read-only per agent;
  the role files tell reviewers to write only their review file.
- **Same names everywhere** (`flow-worker`, …), so Cursor, which reads both
  `.claude/agents/` and `.codex/agents/`, sees each role once.
