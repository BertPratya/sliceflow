package cli

import (
	"fmt"
	"io"

	"github.com/BertPratya/sliceflow/internal/spec"
)

// printProblems prints one line per problem (R13): spec.md:<line>: <level>: <message>.
func printProblems(w io.Writer, ps []spec.Problem) {
	for _, p := range ps {
		fmt.Fprintf(w, "spec.md:%d: %s: %s\n", p.Line, p.Level, p.Message)
	}
}

// plural returns "<n> <word>", with an "s" on the word unless n is 1.
func plural(n int, word string) string {
	if n == 1 {
		return fmt.Sprintf("%d %s", n, word)
	}
	return fmt.Sprintf("%d %ss", n, word)
}
