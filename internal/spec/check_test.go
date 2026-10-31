package spec

import (
	"reflect"
	"regexp"
	"strings"
	"testing"
	"time"

	"github.com/BertPratya/sliceflow/templates"
)

// checkToday is the date the checks get as today.
var checkToday = time.Date(2026, 10, 31, 9, 0, 0, 0, time.UTC)

// validSpec is a small spec with no problems, for the slice "demo". It
// also passes R18-R22: S1 shows R1, and R1 has a Check: line. The line
// numbers in the tests below count from this text.
const validSpec = `---
slice: demo
created: 2026-10-29
---

# Spec

## Summary

Demo lets a user do a thing, so that it gets done.

## User stories

### Story 1: Do a thing

A user wants to do a thing.

- **S1.** **When** I run ` + "`demo`" + `,
  **Then** it works.
  Shows: R1

## Requirements

- **R1.** When the user runs demo, sliceflow works.
  Check: e2e

## Out of scope

None.

## Decisions

None.

## Open questions

None.
`

// Hints that many cases share.
const (
	hintNum       = "number them in order from the top, and update the Shows: and Rule: lines"
	hintPlace     = "replace it with real text"
	hintComment   = "delete it. A finished spec has no comments"
	msgPlace      = "placeholder left from the template: "
	msgComment    = "comment left from the template"
	hintStoryHead = `write it as e.g. "### Story 1: Start a new slice"`
)

// change replaces old with new in a spec's text.
type change struct{ old, new string }

// changed returns validSpec with the changes made, in order. Each old text
// must be found exactly once.
func changed(t *testing.T, changes ...change) string {
	t.Helper()
	s := validSpec
	for _, c := range changes {
		if n := strings.Count(s, c.old); n != 1 {
			t.Fatalf("%q found %d times, want 1", c.old, n)
		}
		s = strings.Replace(s, c.old, c.new, 1)
	}
	return s
}

// checkCase is one spec, and the problems it must give.
type checkCase struct {
	name    string
	changes []change
	want    []Problem
}

// runCases checks each case's spec with the slice name "demo".
func runCases(t *testing.T, cases []checkCase) {
	t.Helper()
	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			got := Check([]byte(changed(t, tt.changes...)), "demo", checkToday)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("got  %+v\nwant %+v", got, tt.want)
			}
		})
	}
}

// TestCheckValid checks that the valid spec gives no problems, also with
// \r\n line endings.
func TestCheckValid(t *testing.T) {
	for _, content := range []string{validSpec, strings.ReplaceAll(validSpec, "\n", "\r\n")} {
		if got := Check([]byte(content), "demo", checkToday); len(got) != 0 {
			t.Errorf("got %q, want no problems", got)
		}
	}
}

// TestCheckFrontMatter checks each message of R12.
func TestCheckFrontMatter(t *testing.T) {
	notKeyValue := `front matter line is not "<key>: <value>"`
	dateHint := `write it as e.g. "created: 2026-10-29"`
	runCases(t, []checkCase{
		{"no front matter",
			[]change{{"---\nslice: demo\ncreated: 2026-10-29\n---\n", ""}},
			[]Problem{{0, "no front matter", "start the file with the front matter of: sliceflow spec template"}}},
		{"not closed",
			[]change{{"created: 2026-10-29\n---\n", "created: 2026-10-29\n"}},
			[]Problem{{1, "front matter is not closed", `add a line "---" after the last key`}}},
		{"line with no colon",
			[]change{{"slice: demo\n", "slice: demo\njust text\n"}},
			[]Problem{{3, notKeyValue, dateHint}}},
		{"key with no value",
			[]change{{"created: 2026-10-29\n", "created:   \n"}},
			[]Problem{{3, notKeyValue, dateHint}}},
		{"key with no space after the colon",
			[]change{{"slice: demo\n", "slice:demo\n"}},
			[]Problem{{2, notKeyValue, dateHint}}},
		{"blank line",
			[]change{{"slice: demo\n", "slice: demo\n\n"}},
			[]Problem{{3, notKeyValue, dateHint}}},
		{"no slice",
			[]change{{"slice: demo\n", ""}},
			[]Problem{{0, "front matter has no slice", `add the line "slice: demo"`}}},
		{"no created: hint has today",
			[]change{{"created: 2026-10-29\n", ""}},
			[]Problem{{0, "front matter has no created", `add the line "created: 2026-10-31"`}}},
		{"key twice",
			[]change{{"slice: demo\n", "slice: demo\nslice: demo\n"}},
			[]Problem{{3, "front matter has slice twice", "keep one"}}},
		{"unknown key",
			[]change{{"slice: demo\n", "slice: demo\ntitle: Demo\n"}},
			[]Problem{{3, "unknown key title in the front matter",
				"remove it. The keys are slice and created; the title lives in the event log"}}},
		{"other slice",
			[]change{{"slice: demo", "slice: other"}},
			[]Problem{{2, "front matter says slice other, but this is the spec of demo", `change it to "slice: demo"`}}},
		{"created not a real day",
			[]change{{"created: 2026-10-29", "created: 2026-02-30"}},
			[]Problem{{3, "created 2026-02-30 is not a date as YYYY-MM-DD", dateHint}}},
		{"created with short year",
			[]change{{"created: 2026-10-29", "created: 26-10-29"}},
			[]Problem{{3, "created 26-10-29 is not a date as YYYY-MM-DD", dateHint}}},
		{"created with a time",
			[]change{{"created: 2026-10-29", "created: 2026-10-29T10:00"}},
			[]Problem{{3, "created 2026-10-29T10:00 is not a date as YYYY-MM-DD", dateHint}}},
	})
}

// TestCheckHeadings checks each message of R13.
func TestCheckHeadings(t *testing.T) {
	hintOrder := "put the sections in this order: Summary, User stories, Requirements, Out of scope, Decisions, Open questions"
	missingHint := func(before string) string {
		return `add it after "## ` + before + `". With nothing to say, write None`
	}
	runCases(t, []checkCase{
		{"first heading is not # Spec",
			[]change{{"# Spec\n", "# Specification\n"}},
			[]Problem{{6, `the first heading is not "# Spec"`, `add the line "# Spec" after the front matter`}}},
		{"first heading is a section",
			[]change{{"# Spec\n\n", ""}},
			[]Problem{{6, `the first heading is not "# Spec"`, `add the line "# Spec" after the front matter`}}},
		{"missing Summary",
			[]change{{"## Summary\n\nDemo lets a user do a thing, so that it gets done.\n\n", ""}},
			[]Problem{{0, `missing section "## Summary"`, `add it after "# Spec"`}}},
		{"missing Out of scope",
			[]change{{"## Out of scope\n\nNone.\n\n", ""}},
			[]Problem{{0, `missing section "## Out of scope"`, missingHint("Requirements")}}},
		{"missing Open questions",
			[]change{{"\n## Open questions\n\nNone.\n", ""}},
			[]Problem{{0, `missing section "## Open questions"`, missingHint("Decisions")}}},
		{"unknown section",
			[]change{{"## Out of scope\n", "## Notes\n\nSome notes.\n\n## Out of scope\n"}},
			[]Problem{{27, `unknown section "## Notes"`,
				"the sections are: Summary, User stories, Requirements, Out of scope, Decisions, Open questions"}}},
		{"section twice",
			[]change{{"## Open questions\n", "## Decisions\n\nNone.\n\n## Open questions\n"}},
			[]Problem{{35, `section "## Decisions" appears twice`, "merge them into one"}}},
		{"two neighbours swapped: the lower one is out of order",
			[]change{{"## Decisions\n\nNone.\n\n## Open questions\n\nNone.\n", "## Open questions\n\nNone.\n\n## Decisions\n\nNone.\n"}},
			[]Problem{{35, `section "## Decisions" is out of order`, hintOrder}}},
		{"last section moved to the top: only it is out of order",
			[]change{
				{"\n## Open questions\n\nNone.\n", ""},
				{"## Summary\n", "## Open questions\n\nNone.\n\n## Summary\n"},
			},
			[]Problem{{8, `section "## Open questions" is out of order`, hintOrder}}},
		{"first section moved to the bottom: only it is out of order",
			[]change{
				{"## Summary\n\nDemo lets a user do a thing, so that it gets done.\n\n", ""},
				{"## Open questions\n\nNone.\n", "## Open questions\n\nNone.\n\n## Summary\n\nDemo lets a user do a thing.\n"},
			},
			[]Problem{{35, `section "## Summary" is out of order`, hintOrder}}},
		{"empty section that may be None",
			[]change{{"## Out of scope\n\nNone.\n", "## Out of scope\n\n"}},
			[]Problem{{27, `section "## Out of scope" is empty`, "write its text, or None"}}},
		{"empty section that may not be None",
			[]change{{"Demo lets a user do a thing, so that it gets done.\n\n", ""}},
			[]Problem{{8, `section "## Summary" is empty`, "write its text"}}},
		{"cannot be None",
			[]change{{"Demo lets a user do a thing, so that it gets done.", "None."}},
			[]Problem{{8, `section "## Summary" cannot be None`,
				"a spec needs a summary, at least one story, and at least one requirement"}}},
		{"None. in code is text",
			[]change{{"Demo lets a user do a thing, so that it gets done.", "```\nNone.\n```"}},
			nil},
		{"None. with more text is text",
			[]change{{"Demo lets a user do a thing, so that it gets done.", "None.\nMore."}},
			nil},
		{"text only in inline code is text",
			[]change{{"Demo lets a user do a thing, so that it gets done.", "`demo`"}},
			nil},
	})
}

// TestCheckLeftovers checks each message of R14.
func TestCheckLeftovers(t *testing.T) {
	summary := "Demo lets a user do a thing, so that it gets done."
	runCases(t, []checkCase{
		{"placeholder is cut after 40 characters (S5)",
			[]change{{summary, "{{Two or three sentences: what this feature lets someone do, and why.}}"}},
			[]Problem{{10, msgPlace + "{{Two or three sentences: what this feat…", hintPlace}}},
		{"short placeholder is shown up to its }}",
			[]change{{summary, "Demo lets {{who}} do {{what}}, so it is done."}},
			[]Problem{
				{10, msgPlace + "{{who}}", hintPlace},
				{10, msgPlace + "{{what}}", hintPlace},
			}},
		{"placeholder of exactly 40 characters is not cut",
			[]change{{summary, "{{When <trigger>, <system> <response>.}}"}},
			[]Problem{{10, msgPlace + "{{When <trigger>, <system> <response>.}}", hintPlace}}},
		{"placeholder cut after 40 runes, not bytes",
			[]change{{summary, "Demo {{" + strings.Repeat("é", 50) + "}}"}},
			[]Problem{{10, msgPlace + "{{" + strings.Repeat("é", 38) + "…", hintPlace}}},
		{"placeholder with no }} runs to the end of the line",
			[]change{{summary, "Demo {{lets a user"}},
			[]Problem{{10, msgPlace + "{{lets a user", hintPlace}}},
		{"comment",
			[]change{{summary, summary + "\n<!-- a note -->"}},
			[]Problem{{11, msgComment, hintComment}}},
		{"comment over lines is one problem",
			[]change{{summary, summary + "\n<!-- a note\n     on two lines -->"}},
			[]Problem{{11, msgComment, hintComment}}},
		{"placeholder inside a comment",
			[]change{{summary, summary + " <!-- {{x}} -->"}},
			[]Problem{
				{10, msgComment, hintComment},
				{10, msgPlace + "{{x}}", hintPlace},
			}},
		{"placeholder in the front matter",
			[]change{{"slice: demo\n", "slice: demo\nslice: {{slice}}\n"}},
			[]Problem{
				{3, "front matter has slice twice", "keep one"},
				{3, msgPlace + "{{slice}}", hintPlace},
			}},
		{"inline code is content",
			[]change{{summary, summary + " See `{{x}}` and `<!--`."}},
			nil},
		{"code block is content",
			[]change{{summary, summary + "\n\n```\n{{x}}\n<!-- y -->\n```"}},
			nil},
	})
}

// TestCheckNumbers checks each message of R15.
func TestCheckNumbers(t *testing.T) {
	r1 := "  Check: e2e\n"
	s1 := "  Shows: R1\n"
	runCases(t, []checkCase{
		{"requirement twice",
			[]change{{r1, r1 + "- **R1.** sliceflow also works.\n  Check: e2e\n"}},
			[]Problem{{26, "R1 appears twice (first on line 24)", hintNum}}},
		{"requirement missing",
			[]change{{r1, r1 + "- **R3.** sliceflow also works.\n  Check: e2e\n"}},
			[]Problem{{26, "R3 comes after R1: expected R2", hintNum}}},
		{"one gap is one problem",
			[]change{{r1, r1 + "- **R3.** sliceflow also works.\n  Check: e2e\n- **R4.** And more.\n  Check: e2e\n"}},
			[]Problem{{26, "R3 comes after R1: expected R2", hintNum}}},
		{"two items swapped give three problems",
			[]change{{r1, r1 + "- **R3.** c.\n  Check: e2e\n- **R2.** b.\n  Check: e2e\n- **R4.** d.\n  Check: e2e\n- **R5.** e.\n  Check: e2e\n"}},
			[]Problem{
				{26, "R3 comes after R1: expected R2", hintNum},
				{28, "R2 comes after R3: expected R4", hintNum},
				{30, "R4 comes after R2: expected R3", hintNum},
			}},
		{"removed item keeps its number",
			[]change{{r1, r1 + "- **R2.** (removed in R001: merged into R1)\n- **R3.** sliceflow also works.\n  Check: e2e\n"}},
			nil},
		{"scenario missing",
			[]change{{s1, s1 + "- **S3.** **When** I stop,\n  **Then** it stops.\n  Shows: R1\n"}},
			[]Problem{{21, "S3 comes after S1: expected S2", hintNum}}},
		{"first decision is not 1",
			[]change{{"## Decisions\n\nNone.\n", "## Decisions\n\n- **SD2.** Demo is small.\n  Why: a\n  Not chosen: b\n  Cost: None\n  Rule: none\n"}},
			[]Problem{{33, "SD2 comes first: expected SD1", hintNum}}},
		{"first question is not 1, and removed",
			[]change{{"## Open questions\n\nNone.\n", "## Open questions\n\n- **Q2.** (removed in R001: answered in R1)\n"}},
			[]Problem{{37, "Q2 comes first: expected Q1", hintNum}}},
		{"story missing",
			[]change{{s1, s1 + "\n### Story 3: Do more\n\nA user wants more.\n\n- **S2.** **When** I go on,\n  **Then** it goes on.\n  Shows: R1\n"}},
			[]Problem{{22, "Story 3 comes after Story 1: expected Story 2", hintNum}}},
		{"story twice",
			[]change{{s1, s1 + "\n### Story 1: Do more\n\nA user wants more.\n\n- **S2.** **When** I go on,\n  **Then** it goes on.\n  Shows: R1\n"}},
			[]Problem{{22, "Story 1 appears twice (first on line 14)", hintNum}}},
		{"first story is not 1",
			[]change{{"### Story 1:", "### Story 2:"}},
			[]Problem{{14, "Story 2 comes first: expected Story 1", hintNum}}},
	})
}

// TestCheckSections checks the message of R16.
func TestCheckSections(t *testing.T) {
	runCases(t, []checkCase{
		{"requirement in Decisions",
			[]change{{"## Decisions\n\nNone.\n", "## Decisions\n\n- **R2.** sliceflow does more.\n  Check: e2e\n"}},
			[]Problem{{33, "R2 is in the wrong section", `move it to "## Requirements"`}}},
		{"scenario in Requirements",
			[]change{{"  Check: e2e\n", "  Check: e2e\n- **S2.** **When** I stop,\n  **Then** it stops.\n  Shows: R1\n"}},
			[]Problem{{26, "S2 is in the wrong section", `move it to "## User stories"`}}},
		{"decision in Out of scope",
			[]change{{"## Out of scope\n\nNone.\n", "## Out of scope\n\n- **SD1.** Demo is small.\n  Why: a\n  Not chosen: b\n  Cost: None\n  Rule: none\n"}},
			[]Problem{{29, "SD1 is in the wrong section", `move it to "## Decisions"`}}},
		{"question in Summary",
			[]change{{"so that it gets done.\n", "so that it gets done.\n\n- **Q1.** (removed in R001: answered)\n- **Q2.** Is it done?\n"}},
			[]Problem{
				{13, "Q2 is in the wrong section", `move it to "## Open questions"`},
				{13, "Q2 is not answered", "write the answer into the spec, and remove Q2. When no question is left, write None"},
			}},
		{"removed item is not checked",
			[]change{{"## Decisions\n\nNone.\n", "## Decisions\n\n- **R2.** (removed in R001: merged into R1)\n"}},
			nil},
		{"home section missing: only R13's problems",
			[]change{{"## Requirements\n\n- **R1.** When the user runs demo, sliceflow works.\n  Check: e2e\n",
				"## Requirement\n\n- **R1.** When the user runs demo, sliceflow works.\n  Check: e2e\n- **R2.** And stops.\n  Check: e2e\n"}},
			[]Problem{
				{0, `missing section "## Requirements"`, `add it after "## User stories". With nothing to say, write None`},
				{22, `unknown section "## Requirement"`,
					"the sections are: Summary, User stories, Requirements, Out of scope, Decisions, Open questions"},
			}},
	})
}

// TestCheckStories checks each message of R17.
func TestCheckStories(t *testing.T) {
	s1 := "  Shows: R1\n"
	sentence := "A user wants to do a thing.\n\n"
	// story2 adds S2 to the first story, then Story 2 with S3.
	story2 := "- **S2.** **When** I stop,\n  **Then** it stops.\n  Shows: R1\n" +
		"\n### Story 2: B\n\nA user wants more.\n\n- **S3.** **When** I go on,\n  **Then** it goes on.\n  Shows: R1\n"
	runCases(t, []checkCase{
		{"heading is not a story",
			[]change{{s1, s1 + "\n### Notes\n\nSome notes.\n"}},
			[]Problem{{22, `heading under "## User stories" is not "### Story <n>: <title>"`, hintStoryHead}}},
		{"story with no title",
			[]change{{"### Story 1: Do a thing", "### Story 1:"}},
			[]Problem{{14, `heading under "## User stories" is not "### Story <n>: <title>"`, hintStoryHead}}},
		{"no sentence",
			[]change{{sentence, ""}},
			[]Problem{{14, "Story 1 has no sentence", "add one sentence under the heading: who wants to do what"}}},
		{"sentence only in a comment",
			[]change{{sentence, "<!-- A user -->\n\n"}},
			[]Problem{
				{14, "Story 1 has no sentence", "add one sentence under the heading: who wants to do what"},
				{16, msgComment, hintComment},
			}},
		{"no scenarios",
			[]change{{s1, s1 + "\n### Story 2: Do more\n\nA user wants more.\n"}},
			[]Problem{{22, "Story 2 has no scenarios", `add a scenario under it, e.g. "- **S1.** **When** ..."`}}},
		{"no sentence and no scenarios",
			[]change{{s1, s1 + "\n### Story 2: Do more\n"}},
			[]Problem{
				{22, "Story 2 has no sentence", "add one sentence under the heading: who wants to do what"},
				{22, "Story 2 has no scenarios", `add a scenario under it, e.g. "- **S1.** **When** ..."`},
			}},
		{"scenario not under a story",
			[]change{
				{"- **S1.** **When** I run", "- **S2.** **When** I run"},
				{"## User stories\n\n", "## User stories\n\n- **S1.** **When** I start,\n  **Then** it starts.\n  Shows: R1\n\n"},
			},
			[]Problem{{14, "S1 is not under a story", `add a heading "### Story <n>: <title>" above it`}}},
		{"removed scenario not under a story",
			[]change{
				{"- **S1.** **When** I run", "- **S2.** **When** I run"},
				{"## User stories\n\n", "## User stories\n\n- **S1.** (removed in R001: merged into S2)\n\n"},
			},
			nil},
		{"heading line is not the sentence",
			[]change{{sentence, "#### Sub\n\n"}},
			[]Problem{{14, "Story 1 has no sentence", "add one sentence under the heading: who wants to do what"}}},
		{"story heading with no colon: one problem",
			[]change{{"### Story 1: Do a thing", "### Story 1 A"}, {s1, s1 + story2}},
			[]Problem{{14, `heading under "## User stories" is not "### Story <n>: <title>"`, hintStoryHead}}},
		{"story heading in lower case: one problem",
			[]change{{"### Story 1: Do a thing", "### story 1: A"}, {s1, s1 + story2}},
			[]Problem{{14, `heading under "## User stories" is not "### Story <n>: <title>"`, hintStoryHead}}},
		{"heading with no number holds its scenarios",
			[]change{{s1, s1 + "\n### Notes\n\n- **S2.** **When** I stop,\n  **Then** it stops.\n  Shows: R1\n"}},
			[]Problem{{22, `heading under "## User stories" is not "### Story <n>: <title>"`, hintStoryHead}}},
		{"#### heading stays in the story",
			[]change{{s1, s1 + "\n#### More\n\n- **S2.** **When** I stop,\n  **Then** it stops.\n  Shows: R1\n"}},
			nil},
	})
}

// TestCheckOrderOnOneLine checks that problems on one line keep the order
// R12-R22, and problems with no line come first (R2).
func TestCheckOrderOnOneLine(t *testing.T) {
	runCases(t, []checkCase{
		{"R12 before R14",
			[]change{{"slice: demo", "slice: {{slice}}"}},
			[]Problem{
				{2, "front matter says slice {{slice}}, but this is the spec of demo", `change it to "slice: demo"`},
				{2, msgPlace + "{{slice}}", hintPlace},
			}},
		{"R13 before R14, no line first",
			[]change{{"## Requirements\n", "## Requirements {{x}}\n"}},
			[]Problem{
				{0, `missing section "## Requirements"`, `add it after "## User stories". With nothing to say, write None`},
				{22, `unknown section "## Requirements {{x}}"`,
					"the sections are: Summary, User stories, Requirements, Out of scope, Decisions, Open questions"},
				{22, msgPlace + "{{x}}", hintPlace},
			}},
		{"R14, R15, R16 on one item",
			[]change{{"  Check: e2e\n", "  Check: e2e\n- **SD5.** {{x}}\n  Why: a\n  Not chosen: b\n  Cost: None\n  Rule: none\n"}},
			[]Problem{
				{26, msgPlace + "{{x}}", hintPlace},
				{26, "SD5 comes first: expected SD1", hintNum},
				{26, "SD5 is in the wrong section", `move it to "## Decisions"`},
			}},
		{"R15 before R17 on a story heading",
			[]change{{"### Story 1: Do a thing", "### Story 2:"}},
			[]Problem{
				{14, "Story 2 comes first: expected Story 1", hintNum},
				{14, `heading under "## User stories" is not "### Story <n>: <title>"`, hintStoryHead},
			}},
	})
}

// TestCheckEmpty checks an empty file.
func TestCheckEmpty(t *testing.T) {
	got := Check(nil, "demo", checkToday)
	want := []Problem{
		{0, "no front matter", "start the file with the front matter of: sliceflow spec template"},
		{0, `the first heading is not "# Spec"`, `add the line "# Spec" after the front matter`},
	}
	for i, s := range sections {
		hint := `add it after "# Spec"`
		if i > 0 {
			hint = `add it after "## ` + sections[i-1] + `". With nothing to say, write None`
		}
		want = append(want, Problem{0, `missing section "## ` + s + `"`, hint})
	}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("got  %+v\nwant %+v", got, want)
	}
}

// templateText returns the built-in template as a new slice "demo" gets
// it: without its title line, and with {{slice}} and {{date}} filled.
// (The slice package's SpecTemplate also edits two comment lines; that
// changes no problem here, only line numbers after line 26.)
func templateText(t *testing.T) string {
	t.Helper()
	s := strings.ReplaceAll(templates.Spec, "\r\n", "\n")
	for _, c := range []change{
		{"title: {{title}}\n", ""},
		{"{{slice}}", "demo"},
		{"{{date}}", "2026-10-29"},
	} {
		if n := strings.Count(s, c.old); n != 1 {
			t.Fatalf("template: %q found %d times, want 1", c.old, n)
		}
		s = strings.Replace(s, c.old, c.new, 1)
	}
	return s
}

// TestCheckTemplate checks the template as a new slice gets it: only the
// placeholders and comments left from it are problems, sorted by line,
// and its open question Q1. A label whose value is a placeholder gives
// no R18-R21 problem. The {{real output}} in the code block is content
// (R10).
func TestCheckTemplate(t *testing.T) {
	got := Check([]byte(templateText(t)), "demo", checkToday)
	place := func(line int, p string) Problem { return Problem{line, msgPlace + p, hintPlace} }
	comment := func(line int) Problem { return Problem{line, msgComment, hintComment} }
	want := []Problem{
		comment(6), // HOW TO FILL THIS TEMPLATE
		place(23, "{{...}}"),
		place(25, "{{...}}"),
		place(33, "{{ is"),
		place(34, "{{ left in it is not finished."),
		place(35, "{{"),
		comment(45), // Summary
		place(49, "{{Two or three sentences: what this feat…"),
		comment(53), // User stories
		place(101, "{{short title}}"),
		place(103, "{{One sentence: who wants to do what.}}"),
		place(105, "{{what the user does}}"),
		place(106, "{{what the user sees}}"),
		place(110, "{{the requirements this scenario shows, …"),
		place(111, "{{the starting state}}"),
		place(112, "{{what the user does}}"),
		place(113, "{{what the user sees}}"),
		place(114, "{{R3}}"),
		comment(118), // Requirements
		place(173, "{{group name}}"),
		place(175, "{{When <trigger>, <system> <response>.}}"),
		place(176, "{{unit, api, e2e, or manual}}"),
		comment(180), // Out of scope
		place(188, "{{What this slice does not do.}}"),
		comment(192), // Decisions
		place(219, "{{The choice.}}"),
		place(220, "{{the reason}}"),
		place(221, "{{the main other option, and why not}}"),
		place(222, "{{the downside we accept, or None}}"),
		place(223, "{{the requirement that states it, e.g. R…"),
		place(224, "{{an earlier decision, e.g. device-pairi…"),
		comment(228), // Open questions
		place(233, "{{A question for the human.}}"),
		{233, "Q1 is not answered", "write the answer into the spec, and remove Q1. When no question is left, write None"},
	}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("got:")
		for _, p := range got {
			t.Errorf("  %d: %s", p.Line, p.Message)
		}
		t.Errorf("want:")
		for _, p := range want {
			t.Errorf("  %d: %s", p.Line, p.Message)
		}
	}
}

// TestCheckTemplateRaw checks the template with {{slice}} and {{date}}
// left as they are: the front matter problems come first on their lines.
func TestCheckTemplateRaw(t *testing.T) {
	s := strings.Replace(strings.ReplaceAll(templates.Spec, "\r\n", "\n"), "title: {{title}}\n", "", 1)
	got := Check([]byte(s), "demo", checkToday)
	want := []Problem{
		{2, "front matter says slice {{slice}}, but this is the spec of demo", `change it to "slice: demo"`},
		{2, msgPlace + "{{slice}}", hintPlace},
		{3, "created {{date}} is not a date as YYYY-MM-DD", `write it as e.g. "created: 2026-10-29"`},
		{3, msgPlace + "{{date}}", hintPlace},
	}
	if len(got) < len(want) || !reflect.DeepEqual(got[:len(want)], want) {
		t.Errorf("got %q, want it to start with %q", got, want)
	}
}

// TestCheckTemplateFilled fills the template by hand, the way a user
// does: with every comment deleted and every placeholder replaced, it
// gives no problems.
func TestCheckTemplateFilled(t *testing.T) {
	s := regexp.MustCompile(`(?s)<!--.*?-->\n`).ReplaceAllString(templateText(t), "")
	s = strings.NewReplacer(
		"{{Two or three sentences: what this feature lets someone do, and why.}}", "Demo lets a user do a thing.",
		"{{short title}}", "Do a thing",
		"{{One sentence: who wants to do what.}}", "A user wants to do a thing.",
		"{{what the user does}}", "I run demo",
		"{{what the user sees}}", "it works",
		"{{real output}}", "ok",
		"{{the requirements this scenario shows, e.g. R1, R2}}", "R1",
		"{{the starting state}}", "demo is set up",
		"{{R3}}", "R1",
		"{{group name}}", "Running",
		"{{When <trigger>, <system> <response>.}}", "When the user runs demo, sliceflow works.",
		"{{unit, api, e2e, or manual}}", "e2e",
		"{{What this slice does not do.}}", "Running on a phone.",
		"{{The choice.}}", "Demo is small.",
		"{{the reason}}", "small is easy",
		"{{the main other option, and why not}}", "big, because it is slow",
		"{{the downside we accept, or None}}", "None",
		"{{the requirement that states it, e.g. R11, or none}}", "R1",
		"{{an earlier decision, e.g. device-pairing/SD3}}", "device-pairing/SD3",
		"- **Q1.** {{A question for the human.}}", "None.",
	).Replace(s)
	if strings.Contains(s, "{{") || strings.Contains(s, "<!--") {
		t.Fatalf("the template is not filled:\n%s", s)
	}
	if got := Check([]byte(s), "demo", checkToday); len(got) != 0 {
		t.Errorf("got %q, want no problems", got)
	}
}
