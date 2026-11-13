package git

import (
	"context"
	"errors"
	"strings"
)

// Config returns the value of the git setting name, as git sees it from
// dir, e.g. sliceflow.worktree.dir (R18). A value set for the repo wins
// over a global one. It returns "" when the setting is not set, and when
// it is set to "".
//
// Steps:
//  1. Run git in dir: config --get <name>.
//  2. If it fails with exit code 1, return "": that is git's answer
//     "not set".
//  3. Any other error: return "" and the error as it is, e.g. exit code
//     128 for a broken config file.
//  4. Remove the line break at the end, and return the value.
func Config(ctx context.Context, dir, name string) (string, error) {
	out, err := run(ctx, dir, "config", "--get", name)
	if err != nil {
		var gitErr *Error
		if errors.As(err, &gitErr) && gitErr.ExitCode == 1 {
			return "", nil
		}
		return "", err
	}
	return strings.TrimSuffix(out, "\n"), nil
}
