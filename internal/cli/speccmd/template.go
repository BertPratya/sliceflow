package speccmd

import (
	"context"
	"fmt"
	"io"

	"github.com/BertPratya/sliceflow/internal/cli/command"
	"github.com/BertPratya/sliceflow/templates"
)

// runTemplate prints the spec template built into sliceflow to stdout,
// exactly as it is stored, with its comments and placeholders, and
// returns 0. It needs no project and no git.
func runTemplate(ctx context.Context, in command.Input, stdout, stderr io.Writer) int {
	fmt.Fprint(stdout, templates.Spec)
	return 0
}
