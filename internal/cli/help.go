package cli

import (
	"fmt"
	"strings"
	"text/tabwriter"
)

// help builds the usage text of g from its commands, so the help always
// lists exactly the commands that g can route to.
func (g group) help() string {
	var b strings.Builder

	fmt.Fprintf(&b, "%s\n\n", g.about)
	fmt.Fprintf(&b, "Usage:\n  %s <command> [arguments]\n\n", g.path)

	b.WriteString("Commands:\n")
	// The tabwriter lines up the second column: each \t ends a column.
	tw := tabwriter.NewWriter(&b, 0, 0, 4, ' ', 0)
	for _, c := range g.commands {
		usage := strings.TrimSpace(c.name + " " + c.args)
		fmt.Fprintf(tw, "  %s\t%s\n", usage, c.about)
	}
	// Flush can only fail when writing fails, and a strings.Builder
	// never fails to write.
	_ = tw.Flush()

	b.WriteString("\nFlags:\n")
	b.WriteString("  -h, --help    Show this help\n\n")
	fmt.Fprintf(&b, "Run \"%s <command> --help\" for more about a command.\n", g.path)

	return b.String()
}
