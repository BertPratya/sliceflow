package git

import (
	"context"
	"strings"
)

// FolderInCommit reports whether the folder path exists in the commit
// (R23). path is a path from the project's top folder, with "/" between
// its parts and no "/" at its end, for example docs/slices/login. Only
// the commit counts: files that are not committed do not.
//
// Steps:
//  1. Ask git ls-tree --full-tree -z --end-of-options <commit> <path>.
//     --full-tree reads path from the project's top folder, from any
//     subfolder. --end-of-options makes git read every word after it as
//     a commit or a path, never as an option. Without a "/" at the end
//     of path, git lists the entry itself, not what is inside it.
//  2. Each entry is "<mode> <type> <hash>\t<path>", ended by a zero byte.
//     Return true when an entry has exactly this path, and the type
//     tree (a folder). A file with that name is not a folder.
func FolderInCommit(ctx context.Context, dir, commit, path string) (bool, error) {
	out, err := run(ctx, dir, "ls-tree", "--full-tree", "-z", "--end-of-options", commit, path)
	if err != nil {
		return false, err
	}
	for _, entry := range strings.Split(out, "\x00") {
		info, entryPath, ok := strings.Cut(entry, "\t")
		// A guard that no test can reach: git lists only the entry
		// with this exact path.
		if !ok || entryPath != path {
			continue
		}
		fields := strings.Fields(info)
		if len(fields) == 3 && fields[1] == "tree" {
			return true, nil
		}
	}
	return false, nil
}
