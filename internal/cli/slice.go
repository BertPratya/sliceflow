package cli

import (
	"context"
	"fmt"
	"io"
	"os"

	"github.com/BertPratya/sliceflow/internal/slice"
)

// runSliceCreate makes a new slice from the name and the options in in,
// in the user's checkout of the current folder. When it works, it prints
// the branch it was made on and where the spec is, and returns 0. Otherwise
// it prints what went wrong, and returns 2 for a title, a description,
// or a name that cannot be used, or 1 for any other failure.
func runSliceCreate(ctx context.Context, in input, stdout, stderr io.Writer) int {
	dir, err := os.Getwd()
	if err != nil {
		printFailure(stderr, "find the current folder", err, nothingChanged)
		return 1
	}

	given := in.args["name"]
	created, err := slice.Create(ctx, dir, slice.CreateOptions{
		Name:        given,
		Title:       in.options["title"],
		Description: in.options["description"],
		From:        in.options["from"],
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

// shortCommit returns the first seven characters of a commit ID, the
// short form people know from git.
func shortCommit(commit string) string {
	if len(commit) < 7 {
		return commit
	}
	return commit[:7]
}
