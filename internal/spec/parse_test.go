package spec

import (
	"reflect"
	"strings"
	"testing"
)

const basicSpec = `---
slice: demo
title: Demo
---

# Spec

## Summary

Text here.

## Requirements

- **R1.** When x, y shall z.
  Check: unit
- **R2.** Another one.
  Why: first part
    second part
  Check: e2e

## Out of scope

- Reminders. Goes to: the backlog
`

func TestParseSections(t *testing.T) {
	d := Parse([]byte(basicSpec))
	var names []string
	var lines []int
	for _, s := range d.Sections {
		names = append(names, s.Name)
		lines = append(lines, s.Line)
	}
	if !reflect.DeepEqual(names, []string{"Summary", "Requirements", "Out of scope"}) {
		t.Fatalf("names = %v", names)
	}
	if !reflect.DeepEqual(lines, []int{8, 12, 21}) {
		t.Fatalf("lines = %v", lines)
	}
	if d.Sections[0].Empty || d.Sections[0].Text != "Text here." {
		t.Fatalf("summary = %+v", d.Sections[0])
	}
	items := d.Sections[1].Items
	if len(items) != 2 {
		t.Fatalf("items = %+v", items)
	}
	if items[0].ID != "R1" || items[0].Line != 14 || items[0].Removed {
		t.Fatalf("item 0 = %+v", items[0])
	}
	if items[0].Tail != "When x, y shall z." {
		t.Fatalf("tail = %q", items[0].Tail)
	}
	if !reflect.DeepEqual(items[0].Labels, []Label{{"Check", "unit", 15}}) {
		t.Fatalf("labels 0 = %+v", items[0].Labels)
	}
	want := []Label{{"Why", "first part second part", 17}, {"Check", "e2e", 19}}
	if items[1].ID != "R2" || items[1].Line != 16 || !reflect.DeepEqual(items[1].Labels, want) {
		t.Fatalf("item 1 = %+v", items[1])
	}
	oos := d.Sections[2].Items
	if len(oos) != 1 || oos[0].ID != "" || oos[0].Tail != "Reminders. Goes to: the backlog" {
		t.Fatalf("out of scope = %+v", oos)
	}
	if d.LastLine != 23 {
		t.Fatalf("LastLine = %d", d.LastLine)
	}
}

func TestParseSkips(t *testing.T) {
	src := "## Requirements\n" + // 1
		"\n" + // 2
		"```\n" + // 3
		"- **R9.** no\n" + // 4
		"name: {{slice}}\n" + // 5
		"<!-- not a comment\n" + // 6
		"```\n" + // 7
		"\n" + // 8
		"Use `{{x}}` and `<!--` here.\n" + // 9
		"\n" + // 10
		"<!-- a comment\n" + // 11
		"- **S1.** no\n" + // 12
		"{{y}}\n" + // 13
		"-->\n" + // 14
		"- **R1.** yes\n" // 15
	d := Parse([]byte(src))
	if len(d.Placeholders) != 0 {
		t.Fatalf("placeholders = %+v", d.Placeholders)
	}
	if len(d.Comments) != 1 || d.Comments[0].Line != 11 {
		t.Fatalf("comments = %+v", d.Comments)
	}
	s := d.Sections[0]
	if len(s.Items) != 1 || s.Items[0].ID != "R1" || len(s.BadMarkers) != 0 {
		t.Fatalf("section = %+v", s)
	}
	// A real placeholder is still found.
	d = Parse([]byte("## A\n\nx {{real}} y\n"))
	if len(d.Placeholders) != 1 || d.Placeholders[0].Line != 3 || d.Placeholders[0].Text != "{{real}}" {
		t.Fatalf("placeholders = %+v", d.Placeholders)
	}
}

func TestParseCRLF(t *testing.T) {
	a := Parse([]byte(basicSpec))
	b := Parse([]byte(strings.ReplaceAll(basicSpec, "\n", "\r\n")))
	if !reflect.DeepEqual(a, b) {
		t.Fatalf("differ:\n%+v\n%+v", a, b)
	}
}

func TestParseRemoved(t *testing.T) {
	d := Parse([]byte("## Requirements\n\n- **R4.** (removed in R002: why)\n- **R5.** kept\n"))
	items := d.Sections[0].Items
	if len(items) != 2 || !items[0].Removed || items[0].ID != "R4" || items[1].Removed {
		t.Fatalf("items = %+v", items)
	}
}

func TestParseFrontMatter(t *testing.T) {
	cases := []struct {
		src   string
		value string
		list  []string
	}{
		{"---\nideas: []\n---\n", "[]", nil},
		{"---\nideas: [docs/ideas/a.md]\n---\n", "[docs/ideas/a.md]", nil},
		{"---\nideas:\n  - docs/ideas/a.md\n---\n", "", []string{"docs/ideas/a.md"}},
	}
	for _, c := range cases {
		d := Parse([]byte(c.src))
		if !d.HasFront || len(d.Front) != 1 {
			t.Fatalf("%q: %+v", c.src, d)
		}
		f := d.Front[0]
		if f.Key != "ideas" || f.Value != c.value || f.Line != 2 || !reflect.DeepEqual(f.ListItems, c.list) {
			t.Fatalf("%q: field %+v", c.src, f)
		}
	}
	d := Parse([]byte("# Spec\n\n---\nslice: x\n---\n"))
	if d.HasFront || len(d.Front) != 0 {
		t.Fatalf("no front: %+v", d)
	}
	d = Parse([]byte("---\nslice: {{slice}}\n---\n"))
	if len(d.Placeholders) != 1 || d.Placeholders[0].Line != 2 {
		t.Fatalf("front placeholder: %+v", d.Placeholders)
	}
}
