// Command sliceflow runs a spec-driven workflow with AI coding agents.
//
// main only starts the program: it turns Ctrl+C into a cancelled
// context (TD5), catches a crash (TD4), and exits with the code that
// internal/cli returns.
package main

import (
	"context"
	"fmt"
	"io"
	"os"
	"os/signal"
	"runtime/debug"

	"github.com/BertPratya/sliceflow/internal/cli"
)

func main() {
	// os.Exit does not run deferred calls, so the work runs in start,
	// whose deferred calls are done when it returns.
	os.Exit(catchCrash(os.Stderr, start))
}

// start runs the command line, and returns its exit code. Ctrl+C
// cancels the context that the command gets (TD5). It stays caught
// until start returns, so a second Ctrl+C does not stop the clean-up
// of the first.
func start() int {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()
	return cli.Run(ctx, os.Args[1:], os.Stdout, os.Stderr)
}

// catchCrash calls work, and returns its exit code. When work crashes
// (panics), catchCrash prints to errOut what Go prints for a crash: the
// line "panic: <value>", a blank line, and the stack of the goroutine.
// Then it returns 1, not the 2 that Go uses, because 2 means a wrong
// command line (TD4, SD17, R76).
//
// Only a crash in the goroutine of work is caught. Go cannot catch a
// crash in another goroutine; it still ends the program with 2.
func catchCrash(errOut io.Writer, work func() int) (code int) {
	defer func() {
		value := recover()
		if value == nil {
			return
		}
		// The error of Fprintf is not checked: the exit code already
		// says that the program failed, and there is no other place to
		// write to.
		fmt.Fprintf(errOut, "panic: %s\n\n%s", panicText(value), debug.Stack())
		code = 1
	}()
	return work()
}

// panicText writes the value of a crash the way Go does for the common
// kinds: the text of an error, the String of a value that has one, and
// any other value with %v.
func panicText(value any) string {
	switch v := value.(type) {
	case error:
		return v.Error()
	case fmt.Stringer:
		return v.String()
	default:
		return fmt.Sprintf("%v", v)
	}
}
