package spec

import (
	"fmt"
	"strings"
)

func warningAt(line int, msg string) Problem {
	return Problem{Line: line, Level: LevelWarning, Message: msg}
}

func noteAt(line int, msg string) Problem {
	return Problem{Line: line, Level: LevelNote, Message: msg}
}

// warnings returns R16's warnings and R17's notes.
func warnings(d Doc) []Problem {
	var ps []Problem
	for _, m := range d.Placeholders {
		ps = append(ps, warningAt(m.Line, "placeholder left: "+m.Text))
	}
	for _, m := range d.Comments {
		ps = append(ps, warningAt(m.Line, "template comment left"))
	}
	for _, s := range d.Sections {
		if s.Empty {
			if s.Name == "Open questions" {
				ps = append(ps, warningAt(s.Line, openQuestions(s)))
			}
			if s.Name != "Success criteria" && s.Name != "Key concepts" {
				ps = append(ps, warningAt(s.Line, fmt.Sprintf("\"## %s\" is empty; write None. if there is nothing to say", s.Name)))
			}
			continue
		}
		if s.Name == "Open questions" && s.Text != "None." {
			ps = append(ps, warningAt(s.Line, openQuestions(s)))
		}
	}
	return append(ps, scenarioNotes(d)...)
}

func openQuestions(s Section) string {
	var ids []string
	for _, it := range s.Items {
		if idLetters(it.ID) == "Q" {
			ids = append(ids, it.ID)
		}
	}
	switch len(ids) {
	case 0:
		return "\"## Open questions\" is not None."
	case 1:
		return "1 open question left (" + ids[0] + ")"
	}
	return fmt.Sprintf("%d open questions left (%s)", len(ids), strings.Join(ids, ", "))
}

// scenarioNotes gives a note for each requirement that no Shows: line names.
func scenarioNotes(d Doc) []Problem {
	shown := map[string]bool{}
	var all []Section
	all = append(append(all, d.Preamble), d.Sections...)
	for _, s := range all {
		for _, it := range s.Items {
			for _, l := range findLabels(it, "Shows") {
				for _, name := range splitList(l.Value) {
					shown[name] = true
				}
			}
		}
	}
	var ps []Problem
	for _, s := range all {
		for _, it := range s.Items {
			if idLetters(it.ID) == "R" && !it.Removed && !shown[it.ID] {
				ps = append(ps, noteAt(it.Line, "no scenario shows "+it.ID))
			}
		}
	}
	return ps
}
