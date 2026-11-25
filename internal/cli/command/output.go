package command

// The printing helpers that every command uses: for a tool or the
// system that failed, and for a stop with Ctrl+C.

import (
	"context"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"strconv"
	"strings"

	"github.com/BertPratya/sliceflow/internal/git"
)

// PrintFailure prints that sliceflow could not do what, followed by
// ending, then why: git's own error output when git failed, or else the
// system's error text.
func PrintFailure(w io.Writer, what string, err error, ending string) {
	var gitErr *git.Error
	if errors.As(err, &gitErr) {
		fmt.Fprintf(w, gitFailedLine, what, ending)
		printGitText(w, gitErr.Stderr)
		return
	}
	fmt.Fprintf(w, systemFailedLine, what, ending)
	fmt.Fprint(w, systemSaid)
	fmt.Fprintf(w, "  %s\n", systemText(err))
}

// printGitText prints what git printed to its error output, word for
// word, with each line indented by two spaces. An empty line stays empty.
// When git printed nothing, it says so.
func printGitText(w io.Writer, stderr string) {
	text := strings.TrimSuffix(strings.ReplaceAll(stderr, "\r\n", "\n"), "\n")
	if text == "" {
		fmt.Fprint(w, gitSaidNothing)
		return
	}
	fmt.Fprint(w, gitSaid)
	for _, line := range strings.Split(text, "\n") {
		if line == "" {
			fmt.Fprintln(w)
			continue
		}
		fmt.Fprintf(w, "  %s\n", line)
	}
}

// systemText returns the system's own words for err. For a file error,
// that is the reason only, e.g. "permission denied": the first line
// already names the file.
func systemText(err error) string {
	var pathErr *fs.PathError
	if errors.As(err, &pathErr) {
		return pathErr.Err.Error()
	}
	return err.Error()
}

// IsStop reports whether err comes from the user stopping sliceflow,
// e.g. with Ctrl+C.
func IsStop(err error) bool {
	return errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded)
}

// PrintBadLines prints one warning that names the lines of the log at
// path that could not be read, the first 10 of them at most. It prints
// nothing when there are none.
func PrintBadLines(w io.Writer, path string, lines []int) {
	switch {
	case len(lines) == 0:
		return
	case len(lines) == 1:
		fmt.Fprintf(w, oneBadLineWarning, lines[0], path)
		return
	}
	shown := lines[:min(len(lines), 10)]
	numbers := make([]string, len(shown))
	for i, line := range shown {
		numbers[i] = strconv.Itoa(line)
	}
	list := strings.Join(numbers, ", ")
	if len(lines) > 10 {
		list += fmt.Sprintf(moreBadLines, len(lines)-10)
	}
	fmt.Fprintf(w, badLinesWarning, list, path)
}
