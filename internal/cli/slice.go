package cli

import (
	"context"
	"fmt"
	"io"
	"os"

	"github.com/BertPratya/sliceflow/internal/slice"
)

// runSliceCreate makes a new slice from the name and the options in in,
// in the project of the current folder. When it works, it prints where
// the branch, the worktree, and the spec are, and returns 0.
func runSliceCreate(ctx context.Context, in input, stdout, stderr io.Writer) int {
	dir, err := os.Getwd()
	if err != nil {
		fmt.Fprintf(stderr, "Error: %v\n", err)
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
		fmt.Fprintf(stderr, "Error: %v\n", err)
		return 1
	}

	fmt.Fprintf(stdout, createdLine, created.Name)
	if given != created.Name {
		fmt.Fprintf(stdout, madeLowercaseNote, given)
	}
	for _, file := range created.Ignored {
		fmt.Fprintf(stdout, ignoredFileNote, file)
	}
	fmt.Fprintf(stdout, branchLine, created.Name, created.Base, shortCommit(created.BaseCommit))
	fmt.Fprintf(stdout, worktreeLine, created.Worktree)
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
