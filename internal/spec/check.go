package spec

import "sort"

// Problem levels.
const (
	LevelError   = "error"
	LevelWarning = "warning"
	LevelNote    = "note"
)

// Problem is one thing found in a spec.
type Problem struct {
	Line           int
	Level, Message string
}

// Check returns the problems of a spec, for the slice with this name. They
// are sorted by line, then errors before warnings before notes.
func Check(d Doc, slice string) []Problem {
	var ps []Problem
	ps = append(ps, frontErrors(d, slice)...)
	ps = append(ps, sectionErrors(d)...)
	ps = append(ps, itemErrors(d)...)
	ps = append(ps, outOfScopeErrors(d)...)
	ps = append(ps, warnings(d)...)
	sortProblems(ps)
	return ps
}

func levelRank(l string) int {
	switch l {
	case LevelError:
		return 0
	case LevelWarning:
		return 1
	}
	return 2
}

func sortProblems(ps []Problem) {
	sort.SliceStable(ps, func(i, j int) bool {
		if ps[i].Line != ps[j].Line {
			return ps[i].Line < ps[j].Line
		}
		return levelRank(ps[i].Level) < levelRank(ps[j].Level)
	})
}

func errorAt(line int, msg string) Problem {
	return Problem{Line: line, Level: LevelError, Message: msg}
}
