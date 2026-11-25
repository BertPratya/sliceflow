// Package spec reads a slice's spec and finds the problems in its shape:
// the problems that would stop the later steps of sliceflow. It reads
// text only, and knows nothing of files or git.
package spec

import "strings"

// Line is one line of a spec, as the checks see it.
type Line struct {
	Number int  // counted from 1
	Code   bool // inside a code block, or one of its ``` lines
	// Text is the line without its line end and without the text of any
	// HTML comment in it. It is "" for a code line, and for a line that
	// is wholly inside a comment.
	Text string
	// Spans are the code spans of Text, each from the start of its first
	// run of backticks to the end of its last, as [start, end) offsets.
	Spans [][2]int
}

// Plain returns Text with every code span blanked out by spaces, so a
// search in it finds only text outside backticks, at the same offsets
// as in Text.
func (l Line) Plain() string {
	if len(l.Spans) == 0 {
		return l.Text
	}
	b := []byte(l.Text)
	for _, span := range l.Spans {
		for i := span[0]; i < span[1]; i++ {
			b[i] = ' '
		}
	}
	return string(b)
}

// readLines splits data into lines and marks each one. Lines end with
// "\n" or "\r\n", and a UTF-8 byte order mark at the start is skipped.
// It also returns the number of the line where each HTML comment
// outside backticks starts, once per comment, in order.
//
// Reading from the top, whichever starts first wins: a code block runs
// from a line that starts with ```, after any spaces, to the next such
// line, and inside it "<!--" is plain text; a comment runs from "<!--"
// outside backticks to the first "-->", and inside it a ``` line does
// not start a code block. Backticks mark code from a run of backticks to
// the next run of the same length on the same line; a run with no such
// partner is plain text.
func readLines(data []byte) ([]Line, []int) {
	text := strings.TrimPrefix(string(data), "\uFEFF")
	raw := strings.Split(text, "\n")
	if raw[len(raw)-1] == "" {
		raw = raw[:len(raw)-1]
	}

	var lines []Line
	var comments []int
	inCode, inComment := false, false
	for i, r := range raw {
		r = strings.TrimSuffix(r, "\r")
		n := i + 1
		if inCode {
			if isFence(r) {
				inCode = false
			}
			lines = append(lines, Line{Number: n, Code: true})
			continue
		}
		if !inComment && isFence(r) {
			inCode = true
			lines = append(lines, Line{Number: n, Code: true})
			continue
		}
		visible, started, still := stripComments(r, inComment)
		for range started {
			comments = append(comments, n)
		}
		inComment = still
		lines = append(lines, Line{Number: n, Text: visible, Spans: codeSpans(visible)})
	}
	return lines, comments
}

// isFence reports whether line starts or ends a code block: it starts
// with ```, after any spaces.
func isFence(line string) bool {
	return strings.HasPrefix(strings.TrimLeft(line, " "), "```")
}

// stripComments returns line without the text of its HTML comments,
// "<!--" and "-->" included. inComment says whether the line starts
// inside a comment. It also returns how many comments start on the line
// outside backticks, and whether the line ends inside a comment.
func stripComments(line string, inComment bool) (visible string, started int, stillIn bool) {
	var b strings.Builder
	i := 0
	for i < len(line) {
		if inComment {
			end := strings.Index(line[i:], "-->")
			if end < 0 {
				return b.String(), started, true
			}
			i += end + len("-->")
			inComment = false
			continue
		}
		switch {
		case line[i] == '`':
			n := runLength(line, i)
			end := i + n
			if partner := findRun(line, end, n); partner >= 0 {
				end = partner + n
			}
			b.WriteString(line[i:end])
			i = end
		case strings.HasPrefix(line[i:], "<!--"):
			started++
			inComment = true
			i += len("<!--")
		default:
			b.WriteByte(line[i])
			i++
		}
	}
	return b.String(), started, inComment
}

// codeSpans returns the code spans of text, as Line.Spans holds them.
func codeSpans(text string) [][2]int {
	var spans [][2]int
	i := 0
	for i < len(text) {
		if text[i] != '`' {
			i++
			continue
		}
		n := runLength(text, i)
		partner := findRun(text, i+n, n)
		if partner < 0 {
			i += n
			continue
		}
		spans = append(spans, [2]int{i, partner + n})
		i = partner + n
	}
	return spans
}

// runLength returns how many backticks follow one another in s from i.
func runLength(s string, i int) int {
	n := 0
	for i+n < len(s) && s[i+n] == '`' {
		n++
	}
	return n
}

// findRun returns where the first run of exactly n backticks in s starts,
// at from or after it, or -1 when there is none.
func findRun(s string, from, n int) int {
	for j := from; j < len(s); {
		if s[j] != '`' {
			j++
			continue
		}
		m := runLength(s, j)
		if m == n {
			return j
		}
		j += m
	}
	return -1
}
