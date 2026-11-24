package speccmd

import (
	"context"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"

	"github.com/BertPratya/sliceflow/internal/cli/command"
	"github.com/BertPratya/sliceflow/internal/cli/slicecmd"
	"github.com/BertPratya/sliceflow/internal/slice"
	"github.com/BertPratya/sliceflow/internal/spec"
)

// runValidate checks the spec of the slice named in in, as the file is
// now in the user's checkout of the current folder. It prints to stdout
// every problem, sorted by line, each with a hint, then how many there
// are, and returns 1; or that there are none, and returns 0. When the
// spec cannot be found or read, it prints why to stderr, and returns 2
// for a name that cannot be used, or 1 for any other failure.
func runValidate(ctx context.Context, in command.Input, stdout, stderr io.Writer) int {
	dir, err := os.Getwd()
	if err != nil {
		command.PrintFailure(stderr, "find the current folder", err, command.NothingChanged)
		return 1
	}
	found, err := slice.SpecPath(ctx, dir, in.Args["name"])
	if err != nil {
		return slicecmd.PrintSliceError(stderr, err)
	}

	data, err := os.ReadFile(found.Path)
	if errors.Is(err, fs.ErrNotExist) {
		fmt.Fprintf(stderr, noSpecLines, found.Name, found.Path)
		return 1
	}
	if err != nil {
		command.PrintFailure(stderr, "read "+found.Path, err, command.NothingChanged)
		return 1
	}

	problems := spec.Check(spec.Read(data))
	if len(problems) == 0 {
		fmt.Fprintf(stdout, noProblemsLine, found.Path)
		return 0
	}
	for _, p := range problems {
		fmt.Fprintf(stdout, problemLines, found.Path, p.Line, p.Message, p.Hint)
	}
	if len(problems) == 1 {
		fmt.Fprint(stdout, oneProblemLine)
	} else {
		fmt.Fprintf(stdout, problemsLine, len(problems))
	}
	return 1
}
