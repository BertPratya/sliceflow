package git

import (
	"context"
	"errors"
)

// FileAt returns the file at path, relative to the repo's top folder and
// written with /, as committed at the tip of the local branch. found is
// false when the branch does not exist, or has no file at path. Any other
// failure is returned as it is.
func FileAt(ctx context.Context, dir, branch, path string) (data []byte, found bool, err error) {
	object := "refs/heads/" + branch + ":" + path
	_, err = run(ctx, dir, "rev-parse", "--verify", "--quiet", object)
	var gitErr *Error
	if errors.As(err, &gitErr) && gitErr.ExitCode == 1 {
		return nil, false, nil
	}
	if err != nil {
		return nil, false, err
	}
	out, err := run(ctx, dir, "cat-file", "blob", object)
	if err != nil {
		return nil, false, err
	}
	return []byte(out), true, nil
}
