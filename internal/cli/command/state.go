package command

import (
	"fmt"
	"io"

	"github.com/BertPratya/sliceflow/internal/state"
)

// PrintStateWarning prints the warning for a state that cannot be worked
// out: which line of the log at logPath stopped it, its event, and why.
// It prints nothing when s is known. Print it after PrintBadLines.
func PrintStateWarning(w io.Writer, logPath string, s state.State) {
	u := s.Unknown
	if u == nil {
		return
	}
	if u.Reason == state.ReasonNotAllowed {
		fmt.Fprintf(w, notAllowedWarning, u.Line, logPath, u.Event, s.Position)
		return
	}
	fmt.Fprintf(w, unknownEventWarning, u.Line, logPath, u.Event)
}
