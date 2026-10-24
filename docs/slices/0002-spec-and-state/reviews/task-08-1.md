---
review: task-08-1
slice: 0002-spec-and-state
kind: task
target: task 08
round: 1
reviewer: code-reviewer (claude opus)
---

# Review

## Summary

The task covers R35 and R36 as written, and the three tests are real.
`template show spec` prints the embedded file byte for byte, and every
user-facing text matches S34, S35, R35, R36, and R38 of
0001-slice-and-events word for word. No findings.

## Checked

- Commits: 553d4ab
- Files read: internal/cli/cli.go, internal/cli/template.go,
  internal/cli/cli_test.go, internal/cli/flags.go, internal/cli/output.go,
  internal/template/template.go, internal/template/template_test.go,
  internal/e2e/cli_groups_and_template_test.go,
  internal/e2e/cli_skeleton_test.go, cmd/sliceflow/main.go, .gitattributes,
  docs/slices/0002-spec-and-state/spec.md (R35 to R37, S33 to S36),
  docs/slices/0001-slice-and-events/spec.md (R38, R39), the task's brief
- Ran:
  go vet ./internal/cli/ ./internal/template/ (pass)
  go test ./internal/cli/ ./internal/template/ ./internal/e2e/ -run 'Template|Groups|Help|Raw|Usage|TopHelp' -count=1 (pass)
  built binary, `template show spec` in a folder outside git, piped to cmp against internal/template/spec.md (identical, 12173 bytes)
  `template show spec` with PATH set to a folder without git (prints all 12173 bytes, exit 0)
  `spec nope`, `template nope` (unknown command "spec nope" / "template nope", exit 2)
  `spec --json`, `template --json`, `template show --json`, `template show plan --json` (unknown flag --json, exit 2, stdout empty)
  `template show spec extra` (too many arguments, exit 2)
  `template show --help`, `template show -h` (own usage on stdout, exit 0)
  `template show -- spec` (prints the template, exit 0)
  `template show ""` (unknown template "", exit 1)
- Requirements: R35, R36
- Interfaces: `func Raw() []byte` in internal/template matches the task;
  `topUsage` has the five new lines exactly as the task's Interfaces
  names them; `template show` uses `register`, so tasks 09, 12, and 16
  can register `spec validate`, `spec submit`, `spec review`, and
  `slice title` the same way, and `isGroup` already routes `spec` and
  `slice` second words to the registry

## Tests

- `TestTemplateShow`: ok
- `TestTemplateShowUnknown`: ok
- `TestGroupsAlone`: ok

## Previous findings

None.

## Findings

None.

## Requirements check

None.

## Needs human decision

None.
