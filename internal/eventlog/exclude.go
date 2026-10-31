package eventlog

import (
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/BertPratya/sliceflow/internal/git"
)

// excludeLines keep the lock file and the temp file of every slice out of
// git (R34).
var excludeLines = []string{
	"docs/slices/*/" + lockName,
	"docs/slices/*/" + logName + ".tmp",
}

// exclude makes sure the repo's .git/info/exclude has excludeLines. It adds
// only the lines that are missing, at the end, and never changes
// .gitignore. dir is any folder inside the repo or one of its worktrees.
// Two sliceflows at the same time may both add a line; a line twice does
// no harm.
func exclude(dir string) error {
	path, err := git.ExcludePath(dir)
	if err != nil {
		return err
	}
	content, err := os.ReadFile(path)
	if err != nil && !errors.Is(err, fs.ErrNotExist) {
		return err
	}

	have := map[string]bool{}
	for line := range strings.SplitSeq(string(content), "\n") {
		have[strings.TrimSpace(line)] = true
	}
	var add strings.Builder
	for _, line := range excludeLines {
		if !have[line] {
			add.WriteString(line + "\n")
		}
	}
	if add.Len() == 0 {
		return nil
	}
	text := add.String()
	if len(content) > 0 && content[len(content)-1] != '\n' {
		text = "\n" + text
	}

	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	f, err := os.OpenFile(path, os.O_WRONLY|os.O_APPEND|os.O_CREATE, 0o644)
	if err != nil {
		return err
	}
	if _, err := f.WriteString(text); err != nil {
		f.Close()
		return err
	}
	return f.Close()
}
