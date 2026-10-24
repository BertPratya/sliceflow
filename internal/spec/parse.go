package spec

import (
	"regexp"
	"strings"
)

var (
	fenceOpen   = regexp.MustCompile("^\\s*(`{3,}|~{3,})")
	itemRe      = regexp.MustCompile(`^- \*\*([A-Z]+[0-9]+)\.\*\*(?:\s+(.*))?$`)
	markerLike  = regexp.MustCompile(`^- +\*\*(?:R|S|SD|SC|Q)[0-9 ]`)
	bulletRe    = regexp.MustCompile(`^- +(.*)$`)
	sectionRe   = regexp.MustCompile(`^##\s+(.+?)\s*$`)
	fieldRe     = regexp.MustCompile(`^([A-Za-z_][A-Za-z0-9_-]*):\s*(.*?)\s*$`)
	listItemRe  = regexp.MustCompile(`^\s*-\s+(.*?)\s*$`)
	labelRe     = regexp.MustCompile(`^(\s+)(Check|Shows|Measured|Why|Not chosen|Cost|Rule|Noticed|Replaces):\s*(.*?)\s*$`)
	removedRe   = regexp.MustCompile(`^\(removed in `)
	frontMarker = "---"
)

// scanned is one line, read for what to skip.
type scanned struct {
	raw    string // the line as written
	vis    string // comments and fenced lines blanked; inline code kept
	clean  string // comments, fenced lines, and inline code blanked
	fenced bool   // the line is in or on a fenced code block
}

// Parse reads the text of a spec. It never fails: what it cannot read, it
// leaves out, and the checks report it.
func Parse(text []byte) Doc {
	src := strings.ReplaceAll(string(text), "\r\n", "\n")
	src = strings.TrimPrefix(src, "\xef\xbb\xbf")
	lines := strings.Split(src, "\n")
	if len(lines) > 0 && lines[len(lines)-1] == "" {
		lines = lines[:len(lines)-1]
	}
	var d Doc
	d.LastLine = len(lines)

	sc := scanLines(lines, &d)

	i := 0
	if len(lines) > 0 && strings.TrimRight(lines[0], " \t") == frontMarker {
		// The front matter counts only when its closing --- is found.
		for j := 1; j < len(lines); j++ {
			if strings.TrimRight(lines[j], " \t") == frontMarker {
				d.HasFront = true
				for k := 1; k < j; k++ {
					readFront(&d, sc[k], k+1)
				}
				i = j + 1
				break
			}
		}
	}

	var cur *Section
	var curItem *Item
	var curLabel *Label
	var labelIndent int
	var text_ []string
	hasContent := false
	finish := func() {
		if cur != nil {
			cur.Text = strings.TrimSpace(strings.Join(text_, "\n"))
			cur.Empty = !hasContent
			d.Sections = append(d.Sections, *cur)
		}
		cur, curItem, curLabel, text_, hasContent = nil, nil, nil, nil, false
	}
	flushItem := func() {
		if curItem != nil {
			cur.Items = append(cur.Items, *curItem)
		}
		curItem, curLabel = nil, nil
	}
	for ; i < len(lines); i++ {
		n := i + 1
		clean := sc[i].clean
		if m := sectionRe.FindStringSubmatch(clean); m != nil {
			if cur != nil {
				flushItem()
			}
			finish()
			cur = &Section{Name: m[1], Line: n}
			continue
		}
		if cur == nil {
			continue
		}
		text_ = append(text_, strings.TrimRight(clean, " \t"))
		if sc[i].fenced || strings.TrimSpace(sc[i].vis) != "" {
			hasContent = true
		}
		vis := strings.TrimRight(sc[i].vis, " \t")
		if m := itemRe.FindStringSubmatch(clean); m != nil {
			flushItem()
			tail := ""
			if end := len("- **") + len(m[1]) + len(".**"); end <= len(vis) {
				tail = strings.TrimSpace(vis[end:])
			}
			curItem = &Item{ID: m[1], Line: n, Tail: tail, Removed: removedRe.MatchString(tail)}
			continue
		}
		if markerLike.MatchString(clean) {
			cur.BadMarkers = append(cur.BadMarkers, Mark{n, strings.TrimSpace(clean)})
			continue
		}
		if cur.Name == "Out of scope" {
			if m := bulletRe.FindStringSubmatch(clean); m != nil {
				flushItem()
				curItem = &Item{Line: n, Tail: strings.TrimSpace(vis[2:])}
				continue
			}
		}
		if curItem == nil {
			continue
		}
		if strings.TrimSpace(vis) == "" {
			curLabel = nil
			continue
		}
		if m := labelRe.FindStringSubmatch(vis); m != nil && clean[:len(m[1])] == m[1] {
			curItem.Labels = append(curItem.Labels, Label{Name: m[2], Value: m[3], Line: n})
			curLabel = &curItem.Labels[len(curItem.Labels)-1]
			labelIndent = len(m[1])
			continue
		}
		if curLabel != nil && indentOf(vis) > labelIndent {
			if curLabel.Value == "" {
				curLabel.Value = strings.TrimSpace(vis)
			} else {
				curLabel.Value += " " + strings.TrimSpace(vis)
			}
			continue
		}
		curLabel = nil
	}
	if cur != nil {
		flushItem()
	}
	finish()
	return d
}

func indentOf(s string) int {
	return len(s) - len(strings.TrimLeft(s, " \t"))
}

// readFront reads one line of the front matter.
func readFront(d *Doc, l scanned, n int) {
	if m := listItemRe.FindStringSubmatch(l.vis); m != nil && len(d.Front) > 0 {
		f := &d.Front[len(d.Front)-1]
		f.ListItems = append(f.ListItems, m[1])
		return
	}
	if m := fieldRe.FindStringSubmatch(l.vis); m != nil {
		d.Front = append(d.Front, Field{Key: m[1], Value: m[2], Line: n})
	}
}

// scanLines finds fenced blocks, inline code, comments, and placeholders.
func scanLines(lines []string, d *Doc) []scanned {
	out := make([]scanned, len(lines))
	inComment := false
	fence := ""  // the fence that opened, e.g. "```"
	codeRun := 0 // backticks of an inline code that goes on from the line before
	for i, raw := range lines {
		n := i + 1
		s := scanned{raw: raw}
		if fence != "" {
			s.vis, s.clean, s.fenced = blank(raw), blank(raw), true
			t := strings.TrimSpace(raw)
			if len(t) >= len(fence) && strings.Trim(t, fence[:1]) == "" {
				fence = ""
			}
			out[i] = s
			continue
		}
		if !inComment && codeRun == 0 {
			if m := fenceOpen.FindStringSubmatch(raw); m != nil && !(m[1][0] == '`' && strings.Contains(raw[len(m[0]):], "`")) {
				fence = m[1]
				s.vis, s.clean, s.fenced = blank(raw), blank(raw), true
				out[i] = s
				continue
			}
		}
		vis := []byte(raw)
		clean := []byte(raw)
		for j := 0; j < len(raw); {
			if inComment {
				k := strings.Index(raw[j:], "-->")
				end := len(raw)
				if k >= 0 {
					end = j + k + 3
					inComment = false
				}
				blankRange(vis, j, end)
				blankRange(clean, j, end)
				j = end
				continue
			}
			if codeRun > 0 {
				end := len(raw)
				if e := closingRun(raw, j, codeRun); e >= 0 {
					end = e
					codeRun = 0
				}
				blankRange(clean, j, end)
				j = end
				continue
			}
			c := raw[j]
			if c == '`' {
				k := j
				for k < len(raw) && raw[k] == '`' {
					k++
				}
				run := raw[j:k]
				if e := closingRun(raw, k, len(run)); e >= 0 {
					blankRange(clean, j, e)
					j = e
				} else if spansLines(lines, i, len(run)) {
					blankRange(clean, j, len(raw))
					codeRun = len(run)
					j = len(raw)
				} else {
					j = k
				}
				continue
			}
			if strings.HasPrefix(raw[j:], "<!--") {
				d.Comments = append(d.Comments, Mark{n, strings.TrimSpace(raw[j:])})
				inComment = true
				blankRange(vis, j, j+4)
				blankRange(clean, j, j+4)
				j += 4
				continue
			}
			j++
		}
		s.vis, s.clean = string(vis), string(clean)
		for j := 0; ; {
			k := strings.Index(s.clean[j:], "{{")
			if k < 0 {
				break
			}
			k += j
			end := len(s.clean)
			if e := strings.Index(s.clean[k:], "}}"); e >= 0 {
				end = k + e + 2
			}
			d.Placeholders = append(d.Placeholders, Mark{n, strings.TrimSpace(s.clean[k:end])})
			j = end
		}
		out[i] = s
	}
	return out
}

// closingRun finds the end of a run of exactly n backticks at or after from.
// It returns the index after the run, or -1.
func closingRun(s string, from, n int) int {
	for j := from; j < len(s); {
		if s[j] != '`' {
			j++
			continue
		}
		k := j
		for k < len(s) && s[k] == '`' {
			k++
		}
		if k-j == n {
			return k
		}
		j = k
	}
	return -1
}

// spansLines tells whether an inline code with a run of n backticks, open at
// the end of line i, closes on a later line of the same paragraph.
func spansLines(lines []string, i, n int) bool {
	for j := i + 1; j < len(lines) && strings.TrimSpace(lines[j]) != ""; j++ {
		if startsBlock(lines[j]) {
			return false
		}
		if closingRun(lines[j], 0, n) >= 0 {
			return true
		}
	}
	return false
}

// startsBlock tells whether a line starts a new block: a list item, a heading,
// or a fence. Inline code cannot run over it.
func startsBlock(line string) bool {
	t := strings.TrimLeft(line, " \t")
	return strings.HasPrefix(t, "- ") || strings.HasPrefix(t, "#") ||
		strings.HasPrefix(t, "```") || strings.HasPrefix(t, "~~~")
}

func blank(s string) string {
	return strings.Repeat(" ", len(s))
}

func blankRange(b []byte, from, to int) {
	for i := from; i < to && i < len(b); i++ {
		b[i] = ' '
	}
}
