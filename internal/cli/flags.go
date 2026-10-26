package cli

import (
	"errors"
	"slices"
	"strings"
)

// flagSpec names the flags a command accepts, without the dashes.
type flagSpec struct {
	values []string // flags that take a value, like "from"
	bools  []string // flags with no value, like "json"
}

// flags is the result of parseFlags.
type flags struct {
	values map[string]string // value flags that were given
	bools  map[string]bool   // bool flags that were given
	help   bool              // -h or --help was given
}

// parseFlags splits args into flags and positional words. Flags may come
// before or after the positional words. A value flag is written
// "--name value" or "--name=value". A bool flag is written "--name". The
// word "--" ends the flags: every word after it is positional.
//
// A value flag takes the next word as its value, unless that word starts
// with "-" (a lone "-" is fine). So "--from --json" is "flag needs a
// value: --from". To give a value that starts with a dash, use the equals
// form: "--from=-x".
//
// The text of a returned error is what is wrong, e.g. "unknown flag --x",
// so the caller can pass it to wrongUse. Even after an error, the whole
// list is read, so f.help is right.
func parseFlags(args []string, spec flagSpec) (flags, []string, error) {
	f := flags{values: map[string]string{}, bools: map[string]bool{}}
	var positional []string

	// A parse error does not stop the loop, so that -h later in the words
	// still sets help. The first error is the one we return.
	var firstErr error
	fail := func(msg string) {
		if firstErr == nil {
			firstErr = errors.New(msg)
		}
	}

	for i := 0; i < len(args); i++ {
		word := args[i]

		if word == "--" {
			positional = append(positional, args[i+1:]...)
			break
		}
		if !strings.HasPrefix(word, "-") || word == "-" {
			positional = append(positional, word)
			continue
		}
		if word == "-h" || word == "--help" {
			f.help = true
			continue
		}

		// Split "--name=value" into the name and the value.
		name, value, hasValue := strings.Cut(word, "=")
		if !strings.HasPrefix(name, "--") {
			fail("unknown flag " + name)
			continue
		}
		bare := strings.TrimPrefix(name, "--")

		switch {
		case slices.Contains(spec.bools, bare):
			// "--json=x" is reported as an unknown flag: the wrong-use
			// texts have no better fit for a bool flag with a value.
			if hasValue {
				fail("unknown flag " + name)
				continue
			}
			f.bools[bare] = true
		case slices.Contains(spec.values, bare):
			if !hasValue {
				if i+1 >= len(args) || (strings.HasPrefix(args[i+1], "-") && args[i+1] != "-") {
					fail("flag needs a value: " + name)
					continue
				}
				i++
				value = args[i]
			}
			f.values[bare] = value
		default:
			fail("unknown flag " + name)
		}
	}

	return f, positional, firstErr
}
