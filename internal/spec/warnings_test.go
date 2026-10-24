package spec

import (
	"os"
	"reflect"
	"sort"
	"strings"
	"testing"
)

// softOf returns the warnings and notes Check finds, ignoring errors.
func softOf(text, slice string) []Problem {
	var out []Problem
	for _, p := range Check(Parse([]byte(text)), slice) {
		if p.Level != LevelError {
			out = append(out, p)
		}
	}
	return out
}

func TestWarningRows(t *testing.T) {
	const oq = "## Open questions\n\nNone.\n"
	cases := []struct {
		name       string
		from, to   string
		anchor     string // the line the problem is on
		anchorN    int
		wantLevel  string
		wantMsg    string
		wantNoSoft bool
	}{
		{"two questions", oq, "## Open questions\n\n- **Q1.** a?\n- **Q2.** b?\n", "## Open questions", 0,
			LevelWarning, "2 open questions left (Q1, Q2)", false},
		{"one question", oq, "## Open questions\n\n- **Q1.** a?\n", "## Open questions", 0,
			LevelWarning, "1 open question left (Q1)", false},
		{"no marker", oq, "## Open questions\n\nMaybe.\n", "## Open questions", 0,
			LevelWarning, `"## Open questions" is not None.`, false},
		{"placeholder", "## Summary\n\nText.\n", "## Summary\n\n{{short title}}\n", "{{short title}}", 0,
			LevelWarning, "placeholder left: {{short title}}", false},
		{"placeholder no close", "## Summary\n\nText.\n", "## Summary\n\nsee {{ the rest here\n", "{{ the rest", 0,
			LevelWarning, "placeholder left: {{ the rest here", false},
		{"comment", "## Summary\n\nText.\n", "## Summary\n\nText.\n<!-- a note -->\n", "<!--", 0,
			LevelWarning, "template comment left", false},
		{"empty Why", "## Why\n\nText.\n", "## Why\n\n", "## Why", 0,
			LevelWarning, `"## Why" is empty; write None. if there is nothing to say`, false},
		{"empty Key concepts", "## Requirements\n", "## Key concepts\n\n## Requirements\n", "", 0, "", "", true},
		{"empty Success criteria", "## User stories\n", "## Success criteria\n\n## User stories\n", "", 0, "", "", true},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			text := strings.Replace(goodSpec, c.from, c.to, 1)
			if text == goodSpec {
				t.Fatal("edit did not apply")
			}
			got := softOf(text, "demo")
			if c.wantNoSoft {
				if len(got) != 0 {
					t.Fatalf("got %v", got)
				}
				return
			}
			want := []Problem{{lineOf(text, c.anchor, c.anchorN), c.wantLevel, c.wantMsg}}
			if !reflect.DeepEqual(got, want) {
				t.Fatalf("got %v\nwant %v", got, want)
			}
		})
	}
}

func TestWarningSkips(t *testing.T) {
	// Code, inline code, and comments hide markers and placeholders (R18).
	text := strings.Replace(goodSpec, "## Summary\n\nText.\n",
		"## Summary\n\n```\nname: {{slice}}\n<!-- x -->\n```\n\nUse `<!--` and `{{y}}` here.\n", 1)
	if got := softOf(text, "demo"); len(got) != 0 {
		t.Fatalf("got %v", got)
	}
	text = strings.Replace(goodSpec, "## Summary\n\nText.\n",
		"## Summary\n\nText.\n<!-- {{inside}} **Q1.** -->\n", 1)
	got := softOf(text, "demo")
	want := []Problem{{lineOf(text, "<!--", 0), LevelWarning, "template comment left"}}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("got %v\nwant %v", got, want)
	}
}

func TestNoteNoScenario(t *testing.T) {
	text := strings.Replace(goodSpec, "  Check: unit\n",
		"  Check: unit\n- **R4.** (removed in R002: why)\n- **R8.** Other thing.\n  Check: unit\n", 1)
	got := softOf(text, "demo")
	want := []Problem{{lineOf(text, "**R8.**", 0), LevelNote, "no scenario shows R8"}}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("got %v\nwant %v", got, want)
	}
}

func TestWarnings0001(t *testing.T) {
	b, err := os.ReadFile("testdata/0001-spec.md")
	if err != nil {
		t.Fatal(err)
	}
	got := softOf(string(b), "0001-slice-and-events")
	var notes []string
	for _, p := range got {
		if p.Level == LevelWarning {
			t.Errorf("warning: %v", p)
		}
		notes = append(notes, p.Message)
	}
	// The notes come in line order, which is not the number order here.
	sort.Strings(notes)
	want := []string{"no scenario shows R11", "no scenario shows R40", "no scenario shows R43"}
	if !reflect.DeepEqual(notes, want) {
		t.Fatalf("got %v, want %v", notes, want)
	}
}

func TestCheckOrder(t *testing.T) {
	text := strings.Replace(goodSpec, "  Check: unit\n",
		"  Check: unit\n- **R2.** {{x}}\n- **R3.** Unshown.\n  Check: unit\n", 1)
	text = strings.Replace(text, "## Open questions\n\nNone.\n", "## Open questions\n\n- **Q1.** a?\n", 1)
	got := Check(Parse([]byte(text)), "demo")
	r2, r3 := lineOf(text, "**R2.**", 0), lineOf(text, "**R3.**", 0)
	want := []Problem{
		{r2, LevelError, "R2 has no Check: line"},
		{r2, LevelWarning, "placeholder left: {{x}}"},
		{r2, LevelNote, "no scenario shows R2"},
		{r3, LevelNote, "no scenario shows R3"},
		{lineOf(text, "## Open questions", 0), LevelWarning, "1 open question left (Q1)"},
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("got %v\nwant %v", got, want)
	}
}
