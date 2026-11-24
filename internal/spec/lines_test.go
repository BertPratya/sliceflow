package spec

import (
	"reflect"
	"testing"
)

// marks returns, for each line, "code", or the line's Text in quotes.
func marks(lines []Line) []string {
	var out []string
	for _, l := range lines {
		if l.Code {
			out = append(out, "code")
		} else {
			out = append(out, `"`+l.Text+`"`)
		}
	}
	return out
}

// Each case reads a few lines and checks how each one is marked, and on
// which lines a comment starts.
func TestReadLines(t *testing.T) {
	tests := []struct {
		about        string
		text         string
		wantMarks    []string
		wantComments []int
	}{
		{"plain lines, the last with no line end", "a\nb",
			[]string{`"a"`, `"b"`}, nil},
		{"a code block, its fences included",
			"a\n```\n## Story 1: {{x}}\n<!-- note -->\n```\nb\n",
			[]string{`"a"`, "code", "code", "code", "code", `"b"`}, nil},
		{"an indented code block",
			"- **Example:**\n  ```\n  x\n  ```\n",
			[]string{`"- **Example:**"`, "code", "code", "code"}, nil},
		{"three backticks in the middle of a line start no block",
			"a ``` b\n## Story 2: {{short title}}\n",
			[]string{`"a ` + "```" + ` b"`, `"## Story 2: {{short title}}"`}, nil},
		{"a comment on one line, with text around it",
			"a <!-- x --> b\n",
			[]string{`"a  b"`}, []int{1}},
		{"a comment over several lines",
			"a <!-- x\ny\nz --> b\nc\n",
			[]string{`"a "`, `""`, `" b"`, `"c"`}, []int{1}},
		{"inside a comment, ``` starts no block, and the comment ends at -->",
			"<!--\n```\n## Story 1: x\n``` -->\n## Story 1: real\n",
			[]string{`""`, `""`, `""`, `""`, `"## Story 1: real"`}, []int{1}},
		{"inside a code block, <!-- is plain text",
			"```\n<!--\n```\nafter\n",
			[]string{"code", "code", "code", `"after"`}, nil},
		{"a comment marker in backticks is code, not a comment",
			"use `<!--` here\n",
			[]string{"\"use `<!--` here\""}, nil},
		{"two comments on one line are two comments",
			"<!-- a --> b <!-- c -->\n",
			[]string{`" b "`}, []int{1, 1}},
		{"\\r\\n line ends and a byte order mark",
			"\uFEFFa\r\nb\r\n",
			[]string{`"a"`, `"b"`}, nil},
	}
	for _, tt := range tests {
		t.Run(tt.about, func(t *testing.T) {
			lines, comments := readLines([]byte(tt.text))
			if got := marks(lines); !reflect.DeepEqual(got, tt.wantMarks) {
				t.Errorf("lines are %v, want %v", got, tt.wantMarks)
			}
			if !reflect.DeepEqual(comments, tt.wantComments) {
				t.Errorf("comments start on %v, want %v", comments, tt.wantComments)
			}
			for i, l := range lines {
				if l.Number != i+1 {
					t.Errorf("line %d has Number %d", i+1, l.Number)
				}
			}
		})
	}
}

// Plain blanks out code spans: a run of backticks up to the next run of
// the same length on the line. A run with no partner stays plain text.
func TestLinePlain(t *testing.T) {
	tests := []struct {
		text, want string
	}{
		{"no code", "no code"},
		{"a `{{x}}` b", "a         b"},
		{"a ``x ` y`` b", "a           b"},
		{"a `` b ` c", "a `` b ` c"},
		{"`a` and `b`", "    and    "},
	}
	for _, tt := range tests {
		lines, _ := readLines([]byte(tt.text))
		if got := lines[0].Plain(); got != tt.want {
			t.Errorf("Plain of %q is %q, want %q", tt.text, got, tt.want)
		}
	}
}
