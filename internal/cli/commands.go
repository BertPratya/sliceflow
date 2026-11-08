package cli

import (
	"io"

	"github.com/spf13/cobra"
)

// The help texts, word for word. Each command keeps its text in Long,
// and printHelp prints Long as it is. Cobra's own help layout is never
// used (TD1).
const (
	// guideHelp is the guide of S41, for `sliceflow` and `sliceflow
	// slice` (R67).
	guideHelp = `sliceflow runs a spec-driven workflow with AI coding agents.

Usage:
  sliceflow <command> [options]

Commands:
  slice create <name>  Make a new slice
  slice status <name>  Show one slice
  slice list           Show all slices

For help with a command, run: sliceflow <command> --help
`

	// createHelp is the help of `slice create`, as in S44 (R68).
	createHelp = `Make a new slice: a branch, a worktree, a spec, and an event log.

Usage:
  sliceflow slice create <name> [options]

Options:
  --title <text>        A short title for the slice. Default: the name.
  --description <text>  One or two sentences about the slice.
  --from <branch>       The branch to start from. Default: main.
  --help                Show this help.
`

	// statusHelp is the help of `slice status`, as in S45 (R69).
	statusHelp = `Show one slice: its title, description, start, branch, worktree, and events.

Usage:
  sliceflow slice status <name> [options]

Options:
  --help  Show this help.
`

	// listHelp is the help of `slice list`, as in R69.
	listHelp = `Show all slices in a table: name, title, and description.

Usage:
  sliceflow slice list [options]

Options:
  --help  Show this help.
`
)

// takesNameKey is the key of a command's Annotations that marks a
// command that takes one slice name: `slice create` and `slice status`.
const takesNameKey = "sliceflow-takes-name"

// takesName reports whether cmd takes one slice name.
func takesName(cmd *cobra.Command) bool {
	return cmd.Annotations[takesNameKey] == "yes"
}

// usageLine returns the usage line of cmd, as in R73 and R75, e.g.
// "sliceflow slice create <name> [options]".
func usageLine(cmd *cobra.Command) string {
	// UseLine is the path of the parent, then Use. Each command sets
	// DisableFlagsInUseLine, so Cobra adds no " [flags]".
	return cmd.UseLine()
}

// newCommands builds the commands of sliceflow, and returns the first
// one, `sliceflow`. Each command calls its work in f. Everything that
// Cobra prints on its own is turned off or replaced (TD1):
//   - its "Error:" line and its usage on an error (SilenceErrors,
//     SilenceUsage);
//   - its "Did you mean" suggestions (DisableSuggestions);
//   - its `completion` command (CompletionOptions);
//   - its `help` command (a hidden command in its place);
//   - its help and usage layout (SetHelpFunc, SetUsageFunc);
//   - its flag errors (SetFlagErrorFunc, see flagError).
//
// The children of a command use these settings too.
//
// checkLine stops every wrong command line before Cobra runs, and run
// prints the help, not Cobra. So only SilenceErrors and SilenceUsage
// matter today: they stop Cobra's text when a command's work fails.
// DisableSuggestions, CompletionOptions, SetHelpCommand, SetHelpFunc,
// SetUsageFunc, SetFlagErrorFunc, and MousetrapHelpText (in init) are
// a second guard, in case Cobra's behaviour changes in a later release.
// No test can reach them today: with any one of them removed, all the
// tests still pass.
func newCommands(f flows, stdout, stderr io.Writer) *cobra.Command {
	root := &cobra.Command{
		Use:                   "sliceflow",
		Long:                  guideHelp,
		Args:                  cobra.ArbitraryArgs,
		SilenceErrors:         true,
		SilenceUsage:          true,
		DisableSuggestions:    true,
		DisableFlagsInUseLine: true,
		CompletionOptions:     cobra.CompletionOptions{DisableDefaultCmd: true},
	}
	root.SetOut(stdout)
	root.SetErr(stderr)
	root.SetHelpCommand(&cobra.Command{Use: "help", Hidden: true})
	root.SetHelpFunc(func(cmd *cobra.Command, _ []string) {
		printHelp(cmd.OutOrStdout(), cmd)
	})
	root.SetUsageFunc(func(*cobra.Command) error { return nil })
	root.SetFlagErrorFunc(flagError)

	sliceCmd := &cobra.Command{
		Use:                   "slice",
		Long:                  guideHelp,
		Args:                  cobra.ArbitraryArgs,
		DisableFlagsInUseLine: true,
	}
	root.AddCommand(sliceCmd)
	sliceCmd.AddCommand(newCreate(f), newStatus(f), newList(f))
	return root
}

// newCreate builds `slice create <name>`, with --title, --description,
// and --from (S44). It gives the values to f.create.
func newCreate(f flows) *cobra.Command {
	var in createInput
	create := &cobra.Command{
		Use:                   "create <name> [options]",
		Long:                  createHelp,
		Annotations:           map[string]string{takesNameKey: "yes"},
		Args:                  cobra.ArbitraryArgs,
		DisableFlagsInUseLine: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) != 1 {
				// checkLine lets only one name through, so this is a
				// safety net.
				return missingName(cmd)
			}
			in.name = args[0]
			in.fromGiven = cmd.Flags().Changed("from")
			return f.create(cmd.Context(), in, cmd.OutOrStdout(), cmd.ErrOrStderr())
		},
	}
	// pflag keeps the last value when an option is given more than once
	// (R71).
	create.Flags().StringVar(&in.title, "title", "", "")
	create.Flags().StringVar(&in.description, "description", "", "")
	create.Flags().StringVar(&in.from, "from", "main", "")
	return create
}

// newStatus builds `slice status <name>` (S45). It gives the name to
// f.status.
func newStatus(f flows) *cobra.Command {
	return &cobra.Command{
		Use:                   "status <name> [options]",
		Long:                  statusHelp,
		Annotations:           map[string]string{takesNameKey: "yes"},
		Args:                  cobra.ArbitraryArgs,
		DisableFlagsInUseLine: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) != 1 {
				return missingName(cmd)
			}
			in := statusInput{name: args[0]}
			return f.status(cmd.Context(), in, cmd.OutOrStdout(), cmd.ErrOrStderr())
		},
	}
}

// newList builds `slice list` (R69). It calls f.list.
func newList(f flows) *cobra.Command {
	return &cobra.Command{
		Use:                   "list [options]",
		Long:                  listHelp,
		Args:                  cobra.ArbitraryArgs,
		DisableFlagsInUseLine: true,
		RunE: func(cmd *cobra.Command, _ []string) error {
			return f.list(cmd.Context(), cmd.OutOrStdout(), cmd.ErrOrStderr())
		},
	}
}
