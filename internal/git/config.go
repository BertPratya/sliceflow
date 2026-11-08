package git

import (
	"context"
	"errors"
)

// ConfigValue reads the git setting key in the folder dir, with git
// config --get. git picks the value, so a value set for the project wins
// over a global one (R18).
//
// set is false when the setting is not set: git exits with 1 then. Any
// other failure is returned as it is. The value is what git printed,
// without the line break at its end, so an empty value is "" with set
// true.
func ConfigValue(ctx context.Context, dir, key string) (value string, set bool, err error) {
	out, err := run(ctx, dir, "config", "--get", key)
	var gitErr *Error
	if errors.As(err, &gitErr) && gitErr.ExitCode == 1 {
		return "", false, nil
	}
	if err != nil {
		return "", false, err
	}
	return trimNewline(out), true, nil
}
