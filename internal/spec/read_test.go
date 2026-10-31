package spec

import (
	"fmt"
	"os"
	"reflect"
	"strings"
	"testing"

	"github.com/BertPratya/sliceflow/templates"
)

// text joins lines with \n, and ends the last one with \n too.
func text(lines ...string) []byte {
	return []byte(strings.Join(lines, "\n") + "\n")
}

// codeLines returns the numbers of the lines that are in code.
func codeLines(d doc) []int {
	var nums []int
	for _, l := range d.lines {
		if l.inCode {
			nums = append(nums, l.num)
		}
	}
	return nums
}

// commentLines returns the numbers of the lines that are partly in a
// comment.
func commentLines(d doc) []int {
	var nums []int
	for _, l := range d.lines {
		if l.inComment {
			nums = append(nums, l.num)
		}
	}
	return nums
}

// headingTexts returns the text of each heading.
func headingTexts(d doc) []string {
	var texts []string
	for _, h := range d.headings {
		texts = append(texts, h.text)
	}
	return texts
}

// itemIDs returns the id of each item.
func itemIDs(d doc) []string {
	var ids []string
	for _, it := range d.items {
		ids = append(ids, it.id)
	}
	return ids
}

// TestReadLines checks how the text is split into lines (R9).
func TestReadLines(t *testing.T) {
	tests := []struct {
		name    string
		content string
		want    []string
	}{
		{"LF", "a\nb\n", []string{"a", "b"}},
		{"CRLF", "a\r\nb\r\n", []string{"a", "b"}},
		{"mixed endings", "a\r\nb\nc\r\n", []string{"a", "b", "c"}},
		{"no newline at the end", "a\nb", []string{"a", "b"}},
		{"blank lines are kept", "a\n\n\nb\n", []string{"a", "", "", "b"}},
		{"a \\r inside a line stays", "a\rb\n", []string{"a\rb"}},
		{"byte order mark is dropped", "\uFEFF---\n", []string{"---"}},
		{"empty file", "", nil},
		{"only a newline", "\n", []string{""}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			d := read([]byte(tt.content))
			var got []string
			for i, l := range d.lines {
				if l.num != i+1 {
					t.Errorf("line %d has num %d, want %d", i, l.num, i+1)
				}
				got = append(got, l.text)
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("lines = %q, want %q", got, tt.want)
			}
		})
	}
}

// TestReadFrontMatter checks where the front matter and the body are.
func TestReadFrontMatter(t *testing.T) {
	tests := []struct {
		name     string
		content  []byte
		want     frontMatter
		wantBody int
		headings []string
	}{
		{
			name:     "closed",
			content:  text("---", "slice: a", "created: 2026-10-31", "---", "# Spec"),
			want:     frontMatter{found: true, closed: true, start: 1, end: 4},
			wantBody: 5,
			headings: []string{"Spec"},
		},
		{
			name:     "closed with CRLF",
			content:  []byte("---\r\nslice: a\r\n---\r\n# Spec\r\n"),
			want:     frontMatter{found: true, closed: true, start: 1, end: 3},
			wantBody: 4,
			headings: []string{"Spec"},
		},
		{
			name:     "closed right away",
			content:  text("---", "---"),
			want:     frontMatter{found: true, closed: true, start: 1, end: 2},
			wantBody: 3,
		},
		{
			name:     "not closed: the body starts at line 2",
			content:  text("---", "slice: a", "# Spec"),
			want:     frontMatter{found: true, start: 1},
			wantBody: 2,
			headings: []string{"Spec"},
		},
		{
			name:     "missing: the body starts at line 1",
			content:  text("# Spec", "---", "slice: a", "---"),
			want:     frontMatter{},
			wantBody: 1,
			headings: []string{"Spec"},
		},
		{
			name:     "a --- with a space is not front matter",
			content:  text("--- ", "slice: a", "---"),
			want:     frontMatter{},
			wantBody: 1,
		},
		{
			name:     "only a --- line",
			content:  text("---"),
			want:     frontMatter{found: true, start: 1},
			wantBody: 2,
		},
		{
			name:     "empty file",
			content:  nil,
			want:     frontMatter{},
			wantBody: 1,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			d := read(tt.content)
			if d.front != tt.want {
				t.Errorf("front = %+v, want %+v", d.front, tt.want)
			}
			if d.body != tt.wantBody {
				t.Errorf("body = %d, want %d", d.body, tt.wantBody)
			}
			if got := headingTexts(d); !reflect.DeepEqual(got, tt.headings) {
				t.Errorf("headings = %q, want %q", got, tt.headings)
			}
		})
	}
}

// TestReadFrontMatterLinesStay checks that the front matter lines keep
// their text in masked, bare, and clean, so R12 and R14 can read them.
func TestReadFrontMatterLinesStay(t *testing.T) {
	d := read(text("---", "slice: {{slice}}", "---"))
	l := d.lines[1]
	if l.masked != l.text || l.bare != l.text || l.clean != l.text || l.inCode || l.inComment {
		t.Errorf("line 2 = %+v, want its text everywhere and no flags", l)
	}
}

// TestReadCodeBlocks checks that code blocks are content (R10).
func TestReadCodeBlocks(t *testing.T) {
	tests := []struct {
		name     string
		content  []byte
		code     []int
		headings []string
		items    []string
	}{
		{
			name:     "a heading and an item inside a fence are not read",
			content:  text("```", "# Not", "- **R1.** x", "```", "# Yes"),
			code:     []int{1, 2, 3, 4},
			headings: []string{"Yes"},
		},
		{
			name:    "a fence indented 2 spaces under an item",
			content: text("- **S1.** x", "  ```", "## Not", "- **R1.** y", "  ```", "  Shows: R1"),
			code:    []int{2, 3, 4, 5},
			items:   []string{"S1"},
		},
		{
			name:     "a fence with a language",
			content:  text("```go", "# Not", "```", "# Yes"),
			code:     []int{1, 2, 3},
			headings: []string{"Yes"},
		},
		{
			name:     "a fence that is never closed runs to the end",
			content:  text("# Yes", "```", "# Not", "- **R1.** x"),
			code:     []int{2, 3, 4},
			headings: []string{"Yes"},
		},
		{
			name:     "a comment marker inside code opens no comment",
			content:  text("```", "<!--", "```", "# Yes"),
			code:     []int{1, 2, 3},
			headings: []string{"Yes"},
		},
		{
			name:     "inline code with ``` is not a fence (spec 0002, line 290)",
			content:  text("  ```` ``` ```` up to the next such line, both included. Text inside a", "# Yes"),
			headings: []string{"Yes"},
		},
		{
			name:     "``` x ``` on its own line is not a fence",
			content:  text("``` x ```", "# Yes"),
			headings: []string{"Yes"},
		},
		{
			name:     "a ```` block holds a ``` fence",
			content:  text("````md", "```", "# Not", "```", "````", "# Yes"),
			code:     []int{1, 2, 3, 4, 5},
			headings: []string{"Yes"},
		},
		{
			name:     "a fence with text after it does not close a block",
			content:  text("```", "```go", "# Not", "```", "# Yes"),
			code:     []int{1, 2, 3, 4},
			headings: []string{"Yes"},
		},
		{
			name:     "a longer run closes a block",
			content:  text("```", "# Not", "`````  ", "# Yes"),
			code:     []int{1, 2, 3},
			headings: []string{"Yes"},
		},
		{
			name:    "front matter lines inside code stay content",
			content: text("```", "---", "slice: a", "---", "```"),
			code:    []int{1, 2, 3, 4, 5},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			d := read(tt.content)
			if got := codeLines(d); !reflect.DeepEqual(got, tt.code) {
				t.Errorf("code lines = %v, want %v", got, tt.code)
			}
			if got := headingTexts(d); !reflect.DeepEqual(got, tt.headings) {
				t.Errorf("headings = %q, want %q", got, tt.headings)
			}
			if got := itemIDs(d); !reflect.DeepEqual(got, tt.items) {
				t.Errorf("items = %q, want %q", got, tt.items)
			}
			if got := commentLines(d); got != nil {
				t.Errorf("comment lines = %v, want none", got)
			}
		})
	}
}

// TestReadCodeLineIsMasked checks that a code line is all spaces in
// masked and bare, and keeps its text in clean.
func TestReadCodeLineIsMasked(t *testing.T) {
	d := read(text("```", "{{x}} <!--", "```"))
	l := d.lines[1]
	if l.masked != "          " || l.bare != "          " {
		t.Errorf("masked = %q, bare = %q, want only spaces", l.masked, l.bare)
	}
	if l.clean != l.text {
		t.Errorf("clean = %q, want %q", l.clean, l.text)
	}
}

// TestReadInlineCode checks that inline code is content (R10): it is
// spaces in masked and bare, and stays in clean.
func TestReadInlineCode(t *testing.T) {
	tests := []struct {
		name       string
		line       string
		wantMasked string
	}{
		{"one backtick", "a `b` c", "a     c"},
		{"two backticks with one inside", "a ``b`c`` d", "a         d"},
		{"two code spans", "`a` b `c`", "    b    "},
		{"an unmatched backtick is text", "a `b c", "a `b c"},
		{"a run with no run of the same length is text", "a ``b` c", "a ``b` c"},
		{"an unmatched run, then a pair", "`` a `b`", "`` a    "},
		{"a placeholder in inline code", "x `{{y}}` z", "x         z"},
		{"a comment marker in inline code", "`<!--` x", "       x"},
		{"four backticks around three (spec 0002, line 290)", "  ```` ``` ```` up to the next such line, both included. Text inside a", "                up to the next such line, both included. Text inside a"},
		{"multi-byte text keeps its length in bytes", "`é` x", "     x"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			d := read(text(tt.line))
			l := d.lines[0]
			if l.masked != tt.wantMasked {
				t.Errorf("masked = %q, want %q", l.masked, tt.wantMasked)
			}
			if l.bare != tt.wantMasked {
				t.Errorf("bare = %q, want %q", l.bare, tt.wantMasked)
			}
			if l.clean != l.text {
				t.Errorf("clean = %q, want %q", l.clean, l.text)
			}
			if l.inComment || l.inCode {
				t.Errorf("inComment = %v, inCode = %v, want both false", l.inComment, l.inCode)
			}
		})
	}
}

// TestReadMarkersInInlineCode checks that no heading, item, or labeled
// line is read inside inline code (R10).
func TestReadMarkersInInlineCode(t *testing.T) {
	d := read(text("`# Not`", "`- **R1.** x`", "- **R2.** x", "  `Check: e2e`"))
	if d.headings != nil {
		t.Errorf("headings = %+v, want none", d.headings)
	}
	if got := itemIDs(d); !reflect.DeepEqual(got, []string{"R2"}) {
		t.Fatalf("items = %q, want [R2]", got)
	}
	if d.items[0].labels != nil {
		t.Errorf("labels = %+v, want none", d.items[0].labels)
	}
}

// TestReadComments checks that nothing is read inside an HTML comment.
func TestReadComments(t *testing.T) {
	tests := []struct {
		name     string
		content  []byte
		comment  []int
		headings []string
		items    []string
	}{
		{
			name: "a comment over several lines with a heading and an item",
			content: text(
				"# Spec",
				"<!-- start",
				"# Not",
				"- **R1.** x",
				"  Check: e2e",
				"end -->",
				"## Requirements",
				"- **R2.** y",
			),
			comment:  []int{2, 3, 4, 5, 6},
			headings: []string{"Spec", "Requirements"},
			items:    []string{"R2"},
		},
		{
			name:     "a comment on one line",
			content:  text("<!-- # Not -->", "# Yes"),
			comment:  []int{1},
			headings: []string{"Yes"},
		},
		{
			name:    "an item with a comment after its text is read",
			content: text("- **R1.** x <!-- note -->"),
			comment: []int{1},
			items:   []string{"R1"},
		},
		{
			name:     "a fence inside a comment opens no code block",
			content:  text("<!--", "```", "# Not", "```", "-->", "# Yes"),
			comment:  []int{1, 2, 3, 4, 5},
			headings: []string{"Yes"},
		},
		{
			name:     "a comment that is never closed runs to the end",
			content:  text("# Yes", "<!-- x", "# Not", "- **R1.** y"),
			comment:  []int{2, 3, 4},
			headings: []string{"Yes"},
		},
		{
			name:     "a --> with no <!-- is text",
			content:  text("-->", "# Yes"),
			headings: []string{"Yes"},
		},
		{
			name:     "two comments on one line",
			content:  text("<!-- a --> b <!-- c", "# Not", "-->", "# Yes"),
			comment:  []int{1, 2, 3},
			headings: []string{"Yes"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			d := read(tt.content)
			if got := commentLines(d); !reflect.DeepEqual(got, tt.comment) {
				t.Errorf("comment lines = %v, want %v", got, tt.comment)
			}
			if got := headingTexts(d); !reflect.DeepEqual(got, tt.headings) {
				t.Errorf("headings = %q, want %q", got, tt.headings)
			}
			if got := itemIDs(d); !reflect.DeepEqual(got, tt.items) {
				t.Errorf("items = %q, want %q", got, tt.items)
			}
			if got := codeLines(d); got != nil {
				t.Errorf("code lines = %v, want none", got)
			}
		})
	}
}

// TestReadCommentTexts checks masked, bare, and clean on a line with a
// comment: masked keeps the comment (R14 reads it), bare and clean drop it.
func TestReadCommentTexts(t *testing.T) {
	d := read(text("a <!-- {{b}} --> `c`"))
	l := d.lines[0]
	want := struct{ masked, bare, clean string }{
		masked: "a <!-- {{b}} -->    ",
		bare:   "a                   ",
		clean:  "a                `c`",
	}
	got := struct{ masked, bare, clean string }{l.masked, l.bare, l.clean}
	if got != want {
		t.Errorf("got %+v, want %+v", got, want)
	}
}

// TestReadHeadings checks which lines are headings, and their level and
// text.
func TestReadHeadings(t *testing.T) {
	tests := []struct {
		line string
		want []heading // nil: not a heading
	}{
		{"# Spec", []heading{{level: 1, text: "Spec", line: 1}}},
		{"## Summary  ", []heading{{level: 2, text: "Summary", line: 1}}},
		{"###### Six", []heading{{level: 6, text: "Six", line: 1}}},
		{"#", []heading{{level: 1, text: "", line: 1}}},
		{"## A `b`", []heading{{level: 2, text: "A `b`", line: 1}}},
		{"## A <!-- b -->", []heading{{level: 2, text: "A", line: 1}}},
		{"####### Seven", nil},
		{"#Spec", nil},
		{" # Spec", nil},
		{"`# Spec`", nil},
	}
	for _, tt := range tests {
		t.Run(tt.line, func(t *testing.T) {
			got := read(text(tt.line)).headings
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("headings = %+v, want %+v", got, tt.want)
			}
		})
	}
}

// TestReadItemStart checks which lines start an item (R11).
func TestReadItemStart(t *testing.T) {
	tests := []struct {
		line string
		want []string // ids; nil: not an item
	}{
		{"- **R3.** x", []string{"R3"}},
		{"- **S1.** **When** I go", []string{"S1"}},
		{"- **SD10.** x", []string{"SD10"}},
		{"- **Q1.**", []string{"Q1"}},
		{"  - **R3.** x", nil},
		{"* **R3.** x", nil},
		{"- **X3.** x", nil},
		{"- **R3** x", nil},
		{"- **R.** x", nil},
		{"- **r3.** x", nil},
		{"- **R10.** x", []string{"R10"}},
		{"- **R01.** x", nil},
		{"- **R0.** x", nil},
	}
	for _, tt := range tests {
		t.Run(tt.line, func(t *testing.T) {
			got := itemIDs(read(text(tt.line)))
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("items = %q, want %q", got, tt.want)
			}
		})
	}
}

// TestReadItems checks each item's kind, number, lines, section, and
// story in a small spec.
func TestReadItems(t *testing.T) {
	content := text(
		"---", // 1
		"slice: a",
		"created: 2026-10-31",
		"---",
		"# Spec", // 5
		"",
		"## User stories",
		"",
		"### Story 1: Go",
		"", // 10
		"A user goes.",
		"",
		"- **S1.** **When** I go,",
		"  Shows: R1",
		"- **S2.** (removed in R002: merged into S1)", // 15
		"",
		"### Story 2: Stop",
		"",
		"- **S3.** **When** I stop,",
		"  Shows: R2", // 20
		"",
		"## Requirements",
		"",
		"### Starting",
		"", // 25
		"- **R1.** sliceflow goes.",
		"  Check: e2e",
		"- **R2.** sliceflow stops.",
		"  Check: unit, e2e",
		"", // 30
		"## Decisions",
		"",
		"- **SD1.** Go first.",
		"  Why: it is fast.",
		"", // 35
		"## Open questions",
		"",
		"- **Q12.** Why?",
	)
	type short struct {
		id, kind   string
		num        int
		line, end  int
		removed    bool
		section    string
		storyLine  int
		labelCount int
	}
	want := []short{
		{"S1", "S", 1, 13, 14, false, "User stories", 9, 1},
		{"S2", "S", 2, 15, 16, true, "User stories", 9, 0},
		{"S3", "S", 3, 19, 21, false, "User stories", 17, 1},
		{"R1", "R", 1, 26, 27, false, "Requirements", 0, 1},
		{"R2", "R", 2, 28, 30, false, "Requirements", 0, 1},
		{"SD1", "SD", 1, 33, 35, false, "Decisions", 0, 1},
		{"Q12", "Q", 12, 38, 38, false, "Open questions", 0, 0},
	}
	var got []short
	for _, it := range read(content).items {
		got = append(got, short{it.id, it.kind, it.num, it.line, it.end, it.removed, it.section, it.storyLine, len(it.labels)})
	}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("items:\n got  %+v\n want %+v", got, want)
	}
}

// TestReadStoryAndSection checks each item's section (R13: up to the
// next ## heading) and the line of the story it is under.
func TestReadStoryAndSection(t *testing.T) {
	content := text(
		"- **S1.** before any heading", // 1
		"## User stories",
		"### Story 3: Go",
		"#### Detail",
		"- **S2.** under a #### inside a story", // 5
		"### Not a story",
		"- **S3.** under a ### that is not a story",
		"### Story 1: A",
		"- **S4.** under the first Story 1",
		"### Story 1: B", // 10
		"- **S5.** under the second Story 1",
		"### Story 0: x",
		"- **S6.** under Story 0",
		"# Spec",
		"- **S7.** under a # heading: same section, no story", // 15
		"## Requirements",
		"- **R1.** x",
	)
	type where struct {
		section   string
		storyLine int
	}
	want := []where{
		{"", 0},
		{"User stories", 3},
		{"User stories", 0},
		{"User stories", 8},
		{"User stories", 10},
		{"User stories", 12},
		{"User stories", 0},
		{"Requirements", 0},
	}
	var got []where
	for _, it := range read(content).items {
		got = append(got, where{it.section, it.storyLine})
	}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("got %+v, want %+v", got, want)
	}
}

// TestReadRemoved checks which items are removed (R11).
func TestReadRemoved(t *testing.T) {
	tests := []struct {
		name    string
		content []byte
		want    bool
	}{
		{"removed", text("- **R4.** (removed in R002: merged into R3)"), true},
		{"removed, why has brackets", text("- **R4.** (removed in R002: merged (see R3))"), true},
		{"removed, over two lines", text("- **R4.** (removed in R002: merged", "  into R3)"), true},
		{"removed, with an old label left", text("- **R4.** (removed in R002: merged)", "  Check: e2e"), true},
		{"removed, why in inline code", text("- **R4.** (removed in R002: `x` went)"), true},
		{"two digits", text("- **R4.** (removed in R02: x)"), false},
		{"no why", text("- **R4.** (removed in R002: )"), false},
		{"no space after the colon", text("- **R4.** (removed in R002:x)"), false},
		{"text after it", text("- **R4.** (removed in R002: x) more"), false},
		{"no brackets", text("- **R4.** removed in R002: x"), false},
		{"normal item", text("- **R4.** sliceflow goes."), false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			items := read(tt.content).items
			if len(items) != 1 {
				t.Fatalf("got %d items, want 1", len(items))
			}
			if items[0].removed != tt.want {
				t.Errorf("removed = %v, want %v (text %q)", items[0].removed, tt.want, items[0].text)
			}
		})
	}
}

// TestReadLabels checks the labeled lines of an item, and their values
// (R11).
func TestReadLabels(t *testing.T) {
	tests := []struct {
		name    string
		content []byte
		want    []label
	}{
		{
			name:    "one label",
			content: text("- **R1.** x", "  Check: e2e"),
			want:    []label{{"Check", "e2e", 2}},
		},
		{
			name:    "a 4-space continuation",
			content: text("- **SD1.** x", "  Not chosen: a, because", "    b is slow.", "  Cost: None"),
			want:    []label{{"Not chosen", "a, because b is slow.", 2}, {"Cost", "None", 4}},
		},
		{
			name:    "a deeper continuation",
			content: text("- **SD1.** x", "  Why: a", "        b"),
			want:    []label{{"Why", "a b", 2}},
		},
		{
			name:    "every known label",
			content: text("- **SD1.** x", "  Check: a", "  Shows: b", "  Why: c", "  Not chosen: d", "  Cost: e", "  Rule: f", "  Replaces: g"),
			want: []label{
				{"Check", "a", 2}, {"Shows", "b", 3}, {"Why", "c", 4}, {"Not chosen", "d", 5},
				{"Cost", "e", 6}, {"Rule", "f", 7}, {"Replaces", "g", 8},
			},
		},
		{
			name:    "an empty value",
			content: text("- **SD1.** x", "  Rule:"),
			want:    []label{{"Rule", "", 2}},
		},
		{
			name:    "inline code stays in the value, a comment does not",
			content: text("- **SD1.** x", "  Rule: `R1` <!-- note -->"),
			want:    []label{{"Rule", "`R1`", 2}},
		},
		{
			name:    "a blank line ends the value",
			content: text("- **SD1.** x", "  Why: a", "", "    b"),
			want:    []label{{"Why", "a", 2}},
		},
		{
			name:    "a code block ends the value",
			content: text("- **S1.** x", "  Shows: R1", "    ```", "    y", "    ```"),
			want:    []label{{"Shows", "R1", 2}},
		},
		{"3 spaces is not a label", text("- **R1.** x", "   Check: e2e"), nil},
		{"4 spaces is not a label", text("- **R1.** x", "    Check: e2e"), nil},
		{"no spaces is not a label", text("- **R1.** x", "Check: e2e"), nil},
		{"an unknown label is text", text("- **R1.** x", "  Note: e2e"), nil},
		{"a label must have its colon", text("- **R1.** x", "  Check e2e"), nil},
		{"a label inside a comment", text("- **R1.** x", "  <!-- Check: e2e -->"), nil},
		{"a label inside a code block", text("- **R1.** x", "  ```", "  Check: e2e", "  ```"), nil},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			items := read(tt.content).items
			if len(items) != 1 {
				t.Fatalf("got %d items, want 1", len(items))
			}
			if !reflect.DeepEqual(items[0].labels, tt.want) {
				t.Errorf("labels = %+v, want %+v", items[0].labels, tt.want)
			}
		})
	}
}

// TestReadLabelsBelongToTheirItem checks that an item ends at the next
// item or heading, so its labels stay its own (R11).
func TestReadLabelsBelongToTheirItem(t *testing.T) {
	d := read(text("- **R1.** x", "- **R2.** y", "  Check: e2e", "### Next", "  Check: unit"))
	if len(d.items) != 2 {
		t.Fatalf("got %d items, want 2", len(d.items))
	}
	if d.items[0].labels != nil {
		t.Errorf("R1 labels = %+v, want none", d.items[0].labels)
	}
	want := []label{{"Check", "e2e", 3}}
	if !reflect.DeepEqual(d.items[1].labels, want) {
		t.Errorf("R2 labels = %+v, want %+v", d.items[1].labels, want)
	}
	if d.items[1].end != 3 {
		t.Errorf("R2 end = %d, want 3", d.items[1].end)
	}
}

// TestReadTemplate reads the built-in template: the examples in its
// comments are not read, the items after them are.
func TestReadTemplate(t *testing.T) {
	d := read([]byte(templates.Spec))
	if !d.front.closed {
		t.Errorf("front = %+v, want closed", d.front)
	}
	wantHeadings := []string{
		"Spec", "Summary", "User stories", "Story 1: {{short title}}", "Requirements",
		"{{group name}}", "Out of scope", "Decisions", "Open questions",
	}
	if got := headingTexts(d); !reflect.DeepEqual(got, wantHeadings) {
		t.Errorf("headings = %q, want %q", got, wantHeadings)
	}
	if got, want := itemIDs(d), []string{"S1", "S2", "R1", "SD1", "Q1"}; !reflect.DeepEqual(got, want) {
		t.Errorf("items = %q, want %q", got, want)
	}
	var names []string
	for _, it := range d.items {
		for _, l := range it.labels {
			names = append(names, it.id+" "+l.name)
		}
	}
	want := []string{
		"S1 Shows", "S2 Shows", "R1 Check",
		"SD1 Why", "SD1 Not chosen", "SD1 Cost", "SD1 Rule", "SD1 Replaces",
	}
	if !reflect.DeepEqual(names, want) {
		t.Errorf("labels = %q, want %q", names, want)
	}
}

// TestReadHeadingLines checks that headings after the front matter keep
// their line number in the file.
func TestReadHeadingLines(t *testing.T) {
	d := read(text("---", "slice: a", "---", "# Spec", "", "## Summary"))
	want := []heading{{level: 1, text: "Spec", line: 4}, {level: 2, text: "Summary", line: 6}}
	if !reflect.DeepEqual(d.headings, want) {
		t.Errorf("headings = %+v, want %+v", d.headings, want)
	}
}

// TestReadItemText checks an item's text over two lines: it stops at the
// first labeled line.
func TestReadItemText(t *testing.T) {
	d := read(text("- **R1.** sliceflow goes,", "  and then stops.", "  Check: e2e", "  more text"))
	if len(d.items) != 1 {
		t.Fatalf("got %d items, want 1", len(d.items))
	}
	if got, want := d.items[0].text, "sliceflow goes, and then stops."; got != want {
		t.Errorf("text = %q, want %q", got, want)
	}
}

// TestReadSpec0002 reads a real spec: its code blocks and inline code hold
// fences, headings, and items that must not be read.
func TestReadSpec0002(t *testing.T) {
	content, err := os.ReadFile("../../docs/specs/0002-spec-and-state.md")
	if err != nil {
		t.Fatal(err)
	}
	d := read(content)

	var sections []string
	for _, h := range d.headings {
		if h.level == 2 {
			sections = append(sections, h.text)
		}
	}
	wantSections := []string{"Summary", "User stories", "Requirements", "Out of scope", "Decisions", "Open questions"}
	if !reflect.DeepEqual(sections, wantSections) {
		t.Errorf("sections = %q, want %q", sections, wantSections)
	}

	ids := map[string][]string{}
	for _, it := range d.items {
		ids[it.kind] = append(ids[it.kind], it.id)
	}
	for _, k := range []struct {
		kind  string
		count int
	}{{"R", 36}, {"S", 19}, {"SD", 9}, {"Q", 0}} {
		var want []string
		for n := 1; n <= k.count; n++ {
			want = append(want, fmt.Sprintf("%s%d", k.kind, n))
		}
		if !reflect.DeepEqual(ids[k.kind], want) {
			t.Errorf("%s items = %q, want %q", k.kind, ids[k.kind], want)
		}
	}
}
