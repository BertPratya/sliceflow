// Command sliceflow runs a spec-driven workflow with AI coding agents.
package main

import (
	"os"

	"github.com/BertPratya/sliceflow/internal/cli"
)

func main() {
	os.Exit(cli.Run(os.Args[1:], os.Stdout, os.Stderr))
}
