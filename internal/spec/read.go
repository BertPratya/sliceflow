// Package spec reads a spec's text and checks it. It never runs git, reads
// files, or prints.
package spec

import (
	"regexp"
	"strconv"
	"strings"
)

// doc is a spec as the reader sees it (R9-R11, PD11).
type doc struct {
	lines    []line // every line of the file, lines[0] is line 1
	front    frontMatter
	body     int       // number of the first line after the front matter
	headings []heading // headings in the body, outside code and comments
	items    []item    // items in the body, in file order
}

// frontMatter says where the front matter is (R12).
type frontMatter struct {
	found  bool // line 1 is "---"
	closed bool // a second "---" line closes it
	start  int  // number of the first "---" line, 0 when not found
	end    int  // number of the closing "---" line, 0 when not closed
}

// line is one line of the file. All four texts have the same length in
// bytes, so a position in one is the same position in the others.
type line struct {
	num       int    // line number, from 1
	text      string // the line as written, without \n or a trailing \r
	masked    string // text with code (inline, or the whole line in a code block) as spaces; comments stay
	bare      string // masked, with comments as spaces too: markers are found here
	clean     string // text with comments as spaces; code stays: values are read here
	inCode    bool   // the line is a fence, or inside a code block (R10)
	inComment bool   // some part of the line is inside an HTML comment
}

// heading is a Markdown heading, e.g. "## Summary".
type heading struct {
	level int    // number of #s, 1 to 6
	text  string // the text after the #s, trimmed
	line  int
}

// item is a requirement, scenario, decision, or question (R11).
type item struct {
	id        string // e.g. "SD3"
	kind      string // "R", "S", "SD", or "Q"
	num       int    // e.g. 3
	line      int    // number of the item's first line
	end       int    // number of the item's last line
	text      string // the item's own text, after the marker, up to its first label, blank line, or code
	removed   bool   // text is "(removed in R<3 digits>: <why>)"
	section   string // text of the ## heading above it, or ""
	storyLine int    // line of the "### Story <n>:" heading it is under, or 0
	labels    []label
}

// label is a labeled line of an item, e.g. "  Check: e2e" (R11).
type label struct {
	name  string // one of knownLabels
	value string // the text after "<name>:", and its continuation lines, joined with one space and trimmed
	line  int
}

// knownLabels are the labels the reader finds. Any other "<word>:" is text.
var knownLabels = []string{"Check", "Shows", "Why", "Not chosen", "Cost", "Rule", "Replaces"}

var (
	// itemRe is the start of an item, e.g. "- **SD3.**".
	// The number has no leading zero.
	itemRe = regexp.MustCompile(`^- \*\*(SD|R|S|Q)([1-9][0-9]{0,8})\.\*\*`)
	// labelRe is a labeled line: exactly 2 spaces, then a known label.
	labelRe = regexp.MustCompile(`^  (` + strings.Join(knownLabels, "|") + `):`)
	// removedRe is the text of a removed item.
	removedRe = regexp.MustCompile(`^\(removed in R[0-9]{3}: \S.*\)$`)
	// storyRe is the start of a story heading's text, e.g. "Story 2:".
	storyRe = regexp.MustCompile(`^Story ([0-9]{1,9}):`)
)

// read reads a spec (R9-R11). Lines are split on \n, and a \r at the end
// of a line is removed. The empty piece after the last \n is not a line.
//
// Assumed: a UTF-8 byte order mark at the start of the file is dropped.
// Assumed: the front matter is read as plain lines: no code, no comments.
// Assumed, as in Markdown (CommonMark): a fence is a line that, after its
// leading spaces, starts with 3 or more backticks, with no backtick in the
// rest of the line. The opening fence may have text after its run, e.g.
// ```md. The block closes at the next fence whose run is at least as long
// as the opening one, with only spaces after it. So a ```` block may hold
// a ``` line.
// Assumed: inside an HTML comment, a ``` line does not open a code block;
// inside a code block, "<!--" does not open a comment. Whichever opens
// first wins, as in Markdown.
// Assumed: a code block or a comment that is never closed runs to the
// end of the file.
func read(content []byte) doc {
	text := strings.TrimPrefix(string(content), "\uFEFF")
	pieces := strings.Split(text, "\n")
	if pieces[len(pieces)-1] == "" {
		pieces = pieces[:len(pieces)-1]
	}
	var d doc
	for i, piece := range pieces {
		piece = strings.TrimSuffix(piece, "\r")
		d.lines = append(d.lines, line{num: i + 1, text: piece})
	}
	d.front, d.body = readFrontMatter(d.lines)
	for i := range d.body - 1 {
		l := &d.lines[i]
		l.masked, l.bare, l.clean = l.text, l.text, l.text
	}
	markContent(d.lines[d.body-1:])
	d.headings, d.items = readItems(d.lines[d.body-1:])
	return d
}

// readFrontMatter finds the front matter (R12), and returns it with the
// number of the first body line.
//
// Assumed: a "---" line must be exactly "---". When the front matter is
// not closed, the body starts at line 2. With no front matter, the body
// starts at line 1.
func readFrontMatter(lines []line) (frontMatter, int) {
	if len(lines) == 0 || lines[0].text != "---" {
		return frontMatter{}, 1
	}
	for _, l := range lines[1:] {
		if l.text == "---" {
			return frontMatter{found: true, closed: true, start: 1, end: l.num}, l.num + 1
		}
	}
	return frontMatter{found: true, start: 1}, 2
}

// markContent fills masked, bare, clean, inCode, and inComment for the
// body lines (R10, PD11). A code block and a comment may run over many
// lines, so the state goes from one line to the next.
func markContent(lines []line) {
	open := 0 // backticks of the fence that opened the block, 0 when not in one
	inComment := false
	for i := range lines {
		l := &lines[i]
		if open > 0 {
			markCode(l)
			if n, rest, ok := fence(l.text); ok && n >= open && strings.Trim(rest, " ") == "" {
				open = 0 // the fence that closes the block
			}
			continue
		}
		if n, _, ok := fence(l.text); ok && !inComment {
			markCode(l) // the fence that opens a block
			open = n
			continue
		}
		inComment = markLine(l, inComment)
	}
}

// markCode marks a line of a code block: all of it is content.
func markCode(l *line) {
	l.inCode = true
	l.masked = spaces(len(l.text))
	l.bare = l.masked
	l.clean = l.text
}

// fence reads a fence line (R10): after its leading spaces, a run of 3
// or more backticks, and no backtick in the rest of the line. n is the
// length of the run, and rest is the text after it. ok is false when the
// line is not a fence, e.g. "``` x ```" is inline code.
func fence(text string) (n int, rest string, ok bool) {
	t := strings.TrimLeft(text, " ")
	n = backticks(t, 0)
	if n < 3 || strings.Contains(t[n:], "`") {
		return 0, "", false
	}
	return n, t[n:], true
}

// markLine fills masked, bare, clean, and inComment for one line that is
// not in a code block. inComment says whether the line starts inside a
// comment. It returns whether the line ends inside a comment.
//
// Inline code is a run of N backticks up to the next run of exactly N
// backticks on the same line. A run with no partner is plain text.
// A comment runs from "<!--" to the next "-->".
func markLine(l *line, inComment bool) bool {
	text := l.text
	masked := []byte(text)
	bare := []byte(text)
	clean := []byte(text)
	l.inComment = inComment
	i := 0
	for i < len(text) {
		switch {
		case inComment:
			end := len(text)
			if j := strings.Index(text[i:], "-->"); j >= 0 {
				end = i + j + len("-->")
				inComment = false
			}
			blank(bare, i, end)
			blank(clean, i, end)
			i = end
		case strings.HasPrefix(text[i:], "<!--"):
			inComment = true
			l.inComment = true
			blank(bare, i, i+len("<!--"))
			blank(clean, i, i+len("<!--"))
			i += len("<!--")
		case text[i] == '`':
			n := backticks(text, i)
			partner := closingRun(text, i+n, n)
			if partner < 0 {
				i += n // no partner: plain text
				continue
			}
			end := partner + n
			blank(masked, i, end)
			blank(bare, i, end)
			i = end
		default:
			i++
		}
	}
	l.masked, l.bare, l.clean = string(masked), string(bare), string(clean)
	return inComment
}

// backticks returns the length of the run of backticks at text[i].
func backticks(text string, i int) int {
	n := 0
	for i+n < len(text) && text[i+n] == '`' {
		n++
	}
	return n
}

// closingRun returns where the next run of exactly n backticks starts,
// looking from text[from]. It returns -1 when there is none.
func closingRun(text string, from, n int) int {
	for i := from; i < len(text); {
		if text[i] != '`' {
			i++
			continue
		}
		run := backticks(text, i)
		if run == n {
			return i
		}
		i += run
	}
	return -1
}

// blank sets b[from:to] to spaces.
func blank(b []byte, from, to int) {
	for i := from; i < to; i++ {
		b[i] = ' '
	}
}

// spaces returns n spaces.
func spaces(n int) string {
	return strings.Repeat(" ", n)
}

// readItems finds the headings and the items in the body lines (R11).
// An item runs up to the next item or heading.
//
// Assumed: a heading is 1 to 6 #s at the start of a line, then a space,
// a tab, or the end of the line.
// A section runs from its ## heading to the next ## heading (R13).
// Assumed: an item's story is reset by a # or ## heading, and by a ###
// heading that is not "### Story <n>:". A #### heading keeps it.
func readItems(lines []line) ([]heading, []item) {
	var headings []heading
	var items []item
	section := ""
	storyLine := 0
	open := -1 // index in items of the item still open, or -1
	closeItem := func(end int) {
		if open >= 0 {
			items[open].end = end
			finishItem(&items[open], lines)
			open = -1
		}
	}
	for _, l := range lines {
		if h, ok := readHeading(l); ok {
			closeItem(l.num - 1)
			headings = append(headings, h)
			switch h.level {
			case 1:
				storyLine = 0
			case 2:
				section, storyLine = h.text, 0
			case 3:
				storyLine = 0
				if _, ok := storyNumber(h.text); ok {
					storyLine = h.line
				}
			}
			continue
		}
		m := itemRe.FindStringSubmatch(l.bare)
		if m == nil {
			continue
		}
		closeItem(l.num - 1)
		num, _ := strconv.Atoi(m[2]) // at most 9 digits: always fits
		items = append(items, item{
			id:        m[1] + m[2],
			kind:      m[1],
			num:       num,
			line:      l.num,
			section:   section,
			storyLine: storyLine,
		})
		open = len(items) - 1
	}
	if len(lines) > 0 {
		closeItem(lines[len(lines)-1].num)
	}
	return headings, items
}

// readHeading reads a heading from a line. ok is false when the line is
// not a heading.
func readHeading(l line) (h heading, ok bool) {
	level := 0
	for level < len(l.bare) && l.bare[level] == '#' {
		level++
	}
	if level == 0 || level > 6 {
		return heading{}, false
	}
	if level < len(l.bare) && l.bare[level] != ' ' && l.bare[level] != '\t' {
		return heading{}, false
	}
	return heading{level: level, text: strings.TrimSpace(l.clean[level:]), line: l.num}, true
}

// storyNumber returns n for a heading text "Story <n>: ...". ok is false
// when the text does not start that way.
func storyNumber(text string) (n int, ok bool) {
	m := storyRe.FindStringSubmatch(text)
	if m == nil {
		return 0, false
	}
	n, _ = strconv.Atoi(m[1]) // at most 9 digits: always fits
	return n, true
}

// finishItem reads an item's text, whether it is removed, and its labels,
// from its lines. it.line and it.end must be set.
//
// Assumed: the item's text is the rest of its first line, and the next
// lines up to its first labeled line, blank line, or code, joined with
// one space. A removed item may go on over more lines this way.
// Assumed: a label's value goes on over the next lines that start with 4
// spaces or more, up to a blank line or code.
func finishItem(it *item, lines []line) {
	first := lines[0].num
	at := func(num int) line { return lines[num-first] }

	start := at(it.line)
	marker := len(itemRe.FindString(start.bare))
	parts := []string{strings.TrimSpace(start.clean[marker:])}
	for num := it.line + 1; num <= it.end; num++ {
		l := at(num)
		if l.inCode || labelRe.MatchString(l.bare) || strings.TrimSpace(l.clean) == "" {
			break
		}
		parts = append(parts, strings.TrimSpace(l.clean))
	}
	it.text = strings.TrimSpace(strings.Join(parts, " "))
	it.removed = removedRe.MatchString(it.text)

	num := it.line + 1
	for num <= it.end {
		l := at(num)
		num++
		m := labelRe.FindStringSubmatch(l.bare) // a code line's bare is blank
		if m == nil {
			continue
		}
		parts := []string{strings.TrimSpace(l.clean[len(m[0]):])}
		for num <= it.end && isContinuation(at(num)) {
			parts = append(parts, strings.TrimSpace(at(num).clean))
			num++
		}
		it.labels = append(it.labels, label{
			name:  m[1],
			value: strings.TrimSpace(strings.Join(parts, " ")),
			line:  l.num,
		})
	}
}

// isContinuation says whether a line goes on with the label above it: it
// starts with 4 spaces or more, and has text outside comments.
func isContinuation(l line) bool {
	return !l.inCode && strings.HasPrefix(l.text, "    ") && strings.TrimSpace(l.clean) != ""
}
