package cli

import (
	"errors"
	"flag"
	"fmt"
	"io"
	"strings"
)

// errHelp is returned by parseFlags after it printed the usage. The command
// returns 0 on it.
var errHelp = errors.New("help requested")

// parseFlags reads flags that come before or after the positional arguments,
// and sets them in fs. It returns the positional arguments. "--" ends the
// flags. -h and --help print usage on stdout and return errHelp. Wrong use
// returns an error with R38's words, e.g. "unknown flag --colour".
func parseFlags(fs *flag.FlagSet, usage string, args []string, stdout io.Writer) ([]string, error) {
	var pos []string
	for i := 0; i < len(args); i++ {
		a := args[i]
		if a == "--" {
			pos = append(pos, args[i+1:]...)
			break
		}
		if len(a) < 2 || a[0] != '-' {
			pos = append(pos, a)
			continue
		}
		text := flagText(a)
		name := strings.TrimLeft(text, "-")
		value, hasValue := "", false
		if j := strings.Index(a, "="); j >= 0 {
			value, hasValue = a[j+1:], true
		}
		f := fs.Lookup(name)
		if f == nil {
			if name == "h" || name == "help" {
				fmt.Fprint(stdout, usage)
				return nil, errHelp
			}
			return nil, errors.New("unknown flag " + text)
		}
		isBool := false
		if b, ok := f.Value.(interface{ IsBoolFlag() bool }); ok {
			isBool = b.IsBoolFlag()
		}
		switch {
		case hasValue:
		case isBool:
			value = "true"
		case i+1 < len(args):
			i++
			value = args[i]
		default:
			return nil, errors.New("flag needs a value: " + text)
		}
		if err := fs.Set(name, value); err != nil {
			return nil, fmt.Errorf("bad value for %s: %v", text, err)
		}
	}
	return pos, nil
}
