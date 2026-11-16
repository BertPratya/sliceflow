package git

import (
	"context"
	"errors"
	"strings"
)

// Config returns the value of the git setting name, as git sees it from
// dir, e.g. sliceflow.worktree.dir. A value set for the repo wins over a
// global one. It returns "" when the setting is not set, and when it is
// set to "". Any other failure, e.g. a broken config file, is returned as
// it is.
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
