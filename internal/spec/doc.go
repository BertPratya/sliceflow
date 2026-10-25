// Package spec reads the text of a spec.md into a Doc. It touches no files
// and no git. Everything inside fenced code blocks, `inline code`, and
// template comments is skipped.
package spec

// Doc is a spec read into parts. Every line number starts at 1.
type Doc struct {
	HasFront     bool
	Front        []Field
	Preamble     Section // the lines above the first "##" heading; Name "" and Line 0
	Sections     []Section
	Placeholders []Mark // every "{{" outside code and comments
	Comments     []Mark // each template comment, at its first line
	LastLine     int    // the number of lines in the text
}

// Field is a front matter key. ListItems holds the "- " lines under it.
type Field struct {
	Key, Value string
	Line       int
	ListItems  []string
}

// Section is a "##" section. Name has no "## ". Text is the section's text
// outside code and comments. Empty is true when Text has nothing in it.
type Section struct {
	Name       string
	Line       int
	Items      []Item
	BadMarkers []Mark
	Empty      bool
	Text       string
}

// Item is a marked item like "- **R7.** text". ID is "" for a plain bullet
// in the Out of scope section. Tail is the first line after the marker.
type Item struct {
	ID      string
	Line    int
	Removed bool
	Labels  []Label
	Tail    string
}

// Label is a labeled line of an item, like "Check: unit". Value holds its
// continued lines, joined by spaces.
type Label struct {
	Name, Value string
	Line        int
}

// Mark is a line number with the text found there.
type Mark struct {
	Line int
	Text string
}
