package slicecmd

import (
	"context"
	"fmt"
	"io"
	"os"

	"github.com/BertPratya/sliceflow/internal/cli/command"
	"github.com/BertPratya/sliceflow/internal/slice"
)

// runSliceCreate makes a new slice from the name and the options in in,
// in the user's checkout of the current folder. When it works, it prints
// the branch it was made on and where the spec is, and returns 0. Otherwise
// it prints what went wrong, and returns 2 for a title, a description,
// or a name that cannot be used, or 1 for any other failure.
func runSliceCreate(ctx context.Context, in command.Input, stdout, stderr io.Writer) int {
	dir, err := os.Getwd()
	if err != nil {
		command.PrintFailure(stderr, "find the current folder", err, command.NothingChanged)
		return 1
	}

	given := in.Args["name"]
	created, err := slice.Create(ctx, dir, slice.CreateOptions{
		Name:        given,
		Title:       in.Options["title"],
		Description: in.Options["description"],
		From:        in.Options["from"],
	})
	if err != nil {
		return printCreateError(stderr, err)
	}

	fmt.Fprintf(stdout, createdLine, created.Name, created.Base)
	if given != created.Name {
		fmt.Fprintf(stdout, madeLowercaseNote, given)
	}
	for _, file := range created.Ignored {
		fmt.Fprintf(stdout, ignoredFileNote, file)
	}
	fmt.Fprintf(stdout, nextLine, created.Spec)
	return 0
}
