package spec

import (
	"fmt"
	"os"
	"reflect"
	"strings"
	"testing"
)

const goodSpec = `---
slice: demo
created: 2026-10-24
ideas: []
---

# Spec

## Summary

Text.

## Why

Text.

## User stories

### Story 1: x

- **S1.** When x.
  Shows: R1

## Requirements

- **R1.** Thing.
  Check: unit

## Out of scope

- Nothing. Goes to: the backlog

## Decisions

- **SD1.** Choice.
  Why: a
  Not chosen: b
  Rule: R1

## Assumptions

- A.

## Open questions

None.
`

// errorsOf returns the errors Check finds, ignoring warnings and notes.
func errorsOf(text, slice string) []Problem {
	var out []Problem
	for _, p := range Check(Parse([]byte(text)), slice) {
		if p.Level == "error" {
			out = append(out, p)
		}
	}
	return out
}

// lineOf returns the line (from 1) of the nth line holding anchor. n -1 is
// the last such line. An empty anchor is the last line of the text.
func lineOf(t, anchor string, n int) int {
	lines := strings.Split(strings.TrimSuffix(t, "\n"), "\n")
	if anchor == "" {
		return len(lines)
	}
	found := []int{}
	for i, l := range lines {
		if strings.Contains(l, anchor) {
			found = append(found, i+1)
		}
	}
	if len(found) == 0 {
		panic("anchor not found: " + anchor)
	}
	if n < 0 {
		return found[len(found)-1]
	}
	return found[n]
}

func TestGoodSpecHasNoErrors(t *testing.T) {
	if got := errorsOf(goodSpec, "demo"); len(got) != 0 {
		t.Fatalf("errors: %v", got)
	}
}

func TestErrorRows(t *testing.T) {
	type edit [2]string
	cases := []struct {
		name  string
		edits []edit
		want  func(t string) Problem
	}{
		{"no front matter", []edit{{"---\nslice: demo\ncreated: 2026-10-24\nideas: []\n---\n\n", ""}},
			func(string) Problem {
				return Problem{1, "error", "no front matter: the file must start with a line ---"}
			}},
		{"no ideas key", []edit{{"ideas: []\n", ""}},
			func(string) Problem { return Problem{1, "error", `front matter has no "ideas"`} }},
		{"wrong slice", []edit{{"slice: demo", "slice: other"}},
			func(t string) Problem {
				return Problem{lineOf(t, "slice: other", 0), "error", `front matter slice is "other", but the slice is demo`}
			}},
		{"bad date", []edit{{"created: 2026-10-24", "created: tomorrow"}},
			func(t string) Problem {
				return Problem{lineOf(t, "created:", 0), "error", `front matter created is "tomorrow", not a date like 2026-10-24`}
			}},
		{"impossible date", []edit{{"created: 2026-10-24", "created: 2026-13-45"}},
			func(t string) Problem {
				return Problem{lineOf(t, "created:", 0), "error", `front matter created is "2026-13-45", not a date like 2026-10-24`}
			}},
		{"ideas not a list", []edit{{"ideas: []", "ideas: none"}},
			func(t string) Problem {
				return Problem{lineOf(t, "ideas:", 0), "error", `front matter ideas is not a list, e.g. ideas: []`}
			}},
		{"missing section", []edit{{"## Why\n\nText.\n\n", ""}},
			func(t string) Problem {
				return Problem{lineOf(t, "## User stories", 0), "error", `missing section "## Why"`}
			}},
		{"missing last section", []edit{{"## Open questions\n\nNone.\n", ""}},
			func(t string) Problem {
				return Problem{lineOf(t, "", 0), "error", `missing section "## Open questions"`}
			}},
		{"unknown section", []edit{{"## Assumptions", "## Extra\n\nText.\n\n## Assumptions"}},
			func(t string) Problem {
				return Problem{lineOf(t, "## Extra", 0), "error", `unknown section "## Extra"`}
			}},
		{"section twice", []edit{{"None.\n", "None.\n\n## Open questions\n\nNone.\n"}},
			func(t string) Problem {
				return Problem{lineOf(t, "## Open questions", -1), "error",
					fmt.Sprintf(`section "## Open questions" appears twice (also line %d)`, lineOf(t, "## Open questions", 0))}
			}},
		{"sections out of order", []edit{{"## Summary\n\nText.\n\n## Why\n\nText.\n", "## Why\n\nText.\n\n## Summary\n\nText.\n"}},
			func(t string) Problem {
				return Problem{lineOf(t, "## Summary", 0), "error", `section "## Summary" must come before "## Why"`}
			}},
		{"not a marker", []edit{{"  Check: unit\n", "  Check: unit\n- **R2 x\n"}},
			func(t string) Problem {
				return Problem{lineOf(t, "- **R2 x", 0), "error", `"**R2 x" is not a marker; write it like **R7.**`}
			}},
		{"marker twice", []edit{{"  Check: unit\n", "  Check: unit\n- **R1.** Again.\n  Check: unit\n"}},
			func(t string) Problem {
				return Problem{lineOf(t, "**R1.** Again", 0), "error",
					fmt.Sprintf("R1 is used twice (also line %d)", lineOf(t, "**R1.** Thing", 0))}
			}},
		{"marker outside section", []edit{{"  Check: unit\n", "  Check: unit\n- **Q2.** y\n"}},
			func(t string) Problem {
				return Problem{lineOf(t, "**Q2.**", 0), "error", `Q2 is outside "## Open questions"`}
			}},
		{"no Check", []edit{{"  Check: unit\n", ""}},
			func(t string) Problem { return Problem{lineOf(t, "**R1.**", 0), "error", "R1 has no Check: line"} }},
		{"unknown check", []edit{{"Check: unit", "Check: unit, fuzz"}},
			func(t string) Problem {
				return Problem{lineOf(t, "Check:", 0), "error", `R1: unknown check "fuzz"; use unit, api, e2e, or manual`}
			}},
		{"no Measured", []edit{{"## User stories", "## Success criteria\n\n- **SC1.** Fast.\n\n## User stories"}},
			func(t string) Problem { return Problem{lineOf(t, "**SC1.**", 0), "error", "SC1 has no Measured: line"} }},
		{"no Shows", []edit{{"  Shows: R1\n", ""}},
			func(t string) Problem { return Problem{lineOf(t, "**S1.**", 0), "error", "S1 has no Shows: line"} }},
		{"shows missing R", []edit{{"Shows: R1", "Shows: R99"}},
			func(t string) Problem {
				return Problem{lineOf(t, "Shows:", 0), "error", "S1 shows R99, which does not exist"}
			}},
		{"shows removed R", []edit{{"  Check: unit\n", "  Check: unit\n- **R2.** (removed in R002: gone)\n"}, {"Shows: R1", "Shows: R1, R2"}},
			func(t string) Problem {
				return Problem{lineOf(t, "Shows:", 0), "error", "S1 shows R2, which was removed"}
			}},
		{"no Why", []edit{{"  Why: a\n", ""}},
			func(t string) Problem { return Problem{lineOf(t, "**SD1.**", 0), "error", "SD1 has no Why: line"} }},
		{"no Not chosen", []edit{{"  Not chosen: b\n", ""}},
			func(t string) Problem {
				return Problem{lineOf(t, "**SD1.**", 0), "error", "SD1 has no Not chosen: line"}
			}},
		{"no Rule", []edit{{"  Rule: R1\n", ""}},
			func(t string) Problem { return Problem{lineOf(t, "**SD1.**", 0), "error", "SD1 has no Rule: line"} }},
		{"Rule names missing R", []edit{{"Rule: R1", "Rule: R9"}},
			func(t string) Problem {
				return Problem{lineOf(t, "Rule:", 0), "error", "SD1: Rule: names R9, which does not exist"}
			}},
		{"Rule names removed R", []edit{{"  Check: unit\n", "  Check: unit\n- **R4.** (removed in R002: gone)\n"}, {"Rule: R1", "Rule: R4"}},
			func(t string) Problem {
				return Problem{lineOf(t, "Rule:", 0), "error", "SD1: Rule: names R4, which was removed"}
			}},
		{"no Goes to", []edit{{" Goes to: the backlog", ""}},
			func(t string) Problem {
				return Problem{lineOf(t, "- Nothing.", 0), "error", "out-of-scope item has no Goes to:"}
			}},
		{"bad Goes to", []edit{{"Goes to: the backlog", "Goes to: someday"}},
			func(t string) Problem {
				return Problem{lineOf(t, "Goes to:", 0), "error",
					`Goes to: "someday" is not a place; use the backlog, a later slice (<name>), or not planned`}
			}},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			text := goodSpec
			for _, e := range c.edits {
				if !strings.Contains(text, e[0]) {
					t.Fatalf("edit target not found: %q", e[0])
				}
				text = strings.Replace(text, e[0], e[1], 1)
			}
			got := errorsOf(text, "demo")
			want := []Problem{c.want(text)}
			if !reflect.DeepEqual(got, want) {
				t.Fatalf("got %v\nwant %v", got, want)
			}
		})
	}
}

func TestErrorGoodVariants(t *testing.T) {
	variants := [][2]string{
		{"ideas: []", "ideas: [docs/ideas/a.md]"},
		{"ideas: []", "ideas:\n  - docs/ideas/a.md\n  - docs/ideas/b.md"},
		{"ideas: []", "ideas:\n- docs/ideas/a.md"},
		{"Check: unit", "Check: unit, e2e"},
		{"Rule: R1", "Rule: none"},
		{"Goes to: the backlog", "Goes to: a later slice (x)"},
		{"Goes to: the backlog", "Goes to: not planned"},
		{"Why: a", "Why: a\n    more"},
	}
	for _, v := range variants {
		text := strings.Replace(goodSpec, v[0], v[1], 1)
		if got := errorsOf(text, "demo"); len(got) != 0 {
			t.Errorf("%q: errors %v", v[1], got)
		}
	}
	crlf := strings.ReplaceAll(goodSpec, "\n", "\r\n")
	if got := errorsOf(crlf, "demo"); len(got) != 0 {
		t.Errorf("crlf: errors %v", got)
	}
}

func TestErrorSkipsCode(t *testing.T) {
	text := strings.Replace(goodSpec, "Text.\n\n## Why", "```\n- **R5 bad\nname: {{slice}}\n```\n\nUse `<!--` here and `- **R6 x`.\n\n## Why", 1)
	if got := errorsOf(text, "demo"); len(got) != 0 {
		t.Fatalf("errors: %v", got)
	}
}

func TestErrorPlaceholderValue(t *testing.T) {
	text := goodSpec
	for _, e := range [][2]string{
		{"Check: unit", "Check: {{unit, api, e2e, or manual}}"},
		{"Shows: R1", "Shows: {{R3}}"},
		{"Rule: R1", "Rule: {{R11}}"},
		{"Goes to: the backlog", "Goes to: {{a later slice}}"},
		{"created: 2026-10-24", "created: {{date}}"},
		{"slice: demo", "slice: {{slice}}"},
		{"ideas: []", "ideas: {{ideas}}"},
	} {
		text = strings.Replace(text, e[0], e[1], 1)
	}
	if got := errorsOf(text, "demo"); len(got) != 0 {
		t.Fatalf("errors: %v", got)
	}
}

func TestProblemsSorted(t *testing.T) {
	text := strings.Replace(goodSpec, "  Check: unit\n", "", 1)
	text = strings.Replace(text, "ideas: []\n", "", 1)
	got := Check(Parse([]byte(text)), "demo")
	for i := 1; i < len(got); i++ {
		if got[i].Line < got[i-1].Line {
			t.Fatalf("not sorted: %v", got)
		}
	}
	if len(got) != 2 {
		t.Fatalf("got %v", got)
	}
}

func TestErrorsFresh(t *testing.T) {
	b, err := os.ReadFile("testdata/fresh.md")
	if err != nil {
		t.Fatal(err)
	}
	if got := errorsOf(string(b), "login-google"); len(got) != 0 {
		t.Fatalf("errors: %v", got)
	}
}

func TestErrors0001(t *testing.T) {
	b, err := os.ReadFile("testdata/0001-spec.md")
	if err != nil {
		t.Fatal(err)
	}
	if got := errorsOf(string(b), "0001-slice-and-events"); len(got) != 0 {
		t.Fatalf("errors: %v", got)
	}
}

// F1: a marker above the first "##" heading is checked too.
func TestErrorMarkerBeforeFirstSection(t *testing.T) {
	text := strings.Replace(goodSpec, "# Spec\n", "# Spec\n\n- **R9.** x\n- **R2 x\n", 1)
	got := errorsOf(text, "demo")
	want := []Problem{
		{lineOf(text, "- **R9.**", 0), "error", `R9 is outside "## Requirements"`},
		{lineOf(text, "- **R9.**", 0), "error", `R9 has no Check: line`},
		{lineOf(text, "- **R2 x", 0), "error", `"**R2 x" is not a marker; write it like **R7.**`},
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("got %v\nwant %v", got, want)
	}
}

// F2: the user's text is shown as written, with no Go escapes.
func TestErrorMessagesKeepUserText(t *testing.T) {
	text := strings.Replace(goodSpec, "- Nothing. Goes to: the backlog", `- Nothing. Goes to: "x\y"`, 1)
	got := errorsOf(text, "demo")
	want := []Problem{{lineOf(text, "Goes to:", 0), "error",
		`Goes to: "` + `"x\y"` + `" is not a place; use the backlog, a later slice (<name>), or not planned`}}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("got %v\nwant %v", got, want)
	}
}

// F3: the inline code of a bad marker stays in the message.
func TestErrorBadMarkerKeepsInlineCode(t *testing.T) {
	text := strings.Replace(goodSpec, "- **R1.** Thing.", "- **R2 uses `x` here\n- **R1.** Thing.", 1)
	got := errorsOf(text, "demo")
	want := []Problem{{lineOf(text, "- **R2 uses", 0), "error", "\"**R2 uses `x` here\" is not a marker; write it like **R7.**"}}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("got %v\nwant %v", got, want)
	}
}

// task-04-2/F1: a requirement above the first "##" heading counts for Shows:.
func TestErrorShowsRequirementBeforeFirstSection(t *testing.T) {
	text := strings.Replace(goodSpec, "# Spec\n", "# Spec\n\n- **R9.** x\n", 1)
	text = strings.Replace(text, "Shows: R1", "Shows: R9", 1)
	got := errorsOf(text, "demo")
	want := []Problem{
		{lineOf(text, "- **R9.**", 0), "error", `R9 is outside "## Requirements"`},
		{lineOf(text, "- **R9.**", 0), "error", `R9 has no Check: line`},
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("got %v\nwant %v", got, want)
	}
}
