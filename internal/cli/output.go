package cli

import (
	"fmt"
	"io"
	"strings"

	"github.com/BertPratya/sliceflow/internal/slice"
)

// usageError prints a wrong-use error (R38) and returns exit code 2.
func usageError(stderr io.Writer, msg string) int {
	fmt.Fprintf(stderr, "Error: %s. Run sliceflow --help.\n", msg)
	return 2
}

// fail prints an error and returns exit code 1. Lines after the first are
// indented by 2 spaces.
func fail(stderr io.Writer, err error) int {
	lines := strings.Split(err.Error(), "\n")
	fmt.Fprintf(stderr, "Error: %s\n", lines[0])
	for _, l := range lines[1:] {
		fmt.Fprintf(stderr, "  %s\n", l)
	}
	return 1
}

// printWarnings prints each warning on stderr (R28).
func printWarnings(stderr io.Writer, ws []slice.Warning) {
	for _, w := range ws {
		if w.Slice == "" {
			fmt.Fprintf(stderr, "Warning: %s\n", w.Text)
		} else {
			fmt.Fprintf(stderr, "Warning: %s: %s\n", w.Slice, w.Text)
		}
	}
}
