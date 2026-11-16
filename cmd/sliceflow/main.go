// Command sliceflow runs a spec-driven workflow with AI coding agents.
package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"runtime/debug"
	"syscall"

	"github.com/BertPratya/sliceflow/internal/cli"
)

// main runs sliceflow with the command line it was started with, and
// exits with the exit code from run.
func main() {
	os.Exit(run())
}

// run runs the command the user typed, and returns the exit code for the
// program: the command's own code, or 1 when sliceflow crashes. A crash
// is reported on stderr the way Go reports it, with the panic and the
// stack trace. Ctrl+C and SIGTERM cancel the context that the command
// gets, which stops a running git.
func run() (exitCode int) {
	defer func() {
		if r := recover(); r != nil {
			fmt.Fprintf(os.Stderr, "panic: %v\n\n%s", r, debug.Stack())
			exitCode = 1
		}
	}()

	ctx, stop := signal.NotifyContext(
		context.Background(),
		os.Interrupt,
		syscall.SIGTERM,
	)
	defer stop()

	return cli.Run(ctx, os.Args[1:], os.Stdout, os.Stderr)
}
