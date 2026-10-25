package cli

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"io"

	"github.com/BertPratya/sliceflow/internal/template"
)

const templateShowUsage = `Usage: sliceflow template show <template>

Prints a built-in template, with all its rule comments. Known templates:
spec. It works in any folder.
`

func init() {
	register(command{Name: "template show", Usage: templateShowUsage, Run: runTemplateShow})
}

func runTemplateShow(ctx context.Context, args []string, stdout, stderr io.Writer) int {
	fs := flag.NewFlagSet("template show", flag.ContinueOnError)
	fs.SetOutput(io.Discard)
	pos, err := parseFlags(fs, templateShowUsage, args, stdout)
	if errors.Is(err, errHelp) {
		return 0
	}
	if err != nil {
		return usageError(stderr, err.Error())
	}
	if len(pos) == 0 {
		return usageError(stderr, "missing template name")
	}
	if len(pos) > 1 {
		return usageError(stderr, "too many arguments")
	}
	if pos[0] != "spec" {
		return fail(stderr, fmt.Errorf("unknown template %q. Known templates: spec.", pos[0]))
	}
	stdout.Write(template.Raw())
	return 0
}
