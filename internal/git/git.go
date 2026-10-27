package git

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"slices"
	"strings"
)

// ErrNotRepo means the folder is not inside a git repository.
var ErrNotRepo = errors.New("not inside a git repository")

// Root returns the top folder of the git repository that dir is in. It
// returns ErrNotRepo when dir is not inside a repository.
func Root(dir string) (string, error) {
	args := []string{"rev-parse", "--show-toplevel"}
	out, err := command(dir, args...).Output()
	if err != nil {
		var exitErr *exec.ExitError
		if errors.As(err, &exitErr) && strings.Contains(string(exitErr.Stderr), "not a git repository") {
			return "", ErrNotRepo
		}
		return "", gitError(args, err)
	}

	return strings.TrimSpace(string(out)), nil
}

// HasCommits reports whether the repository that dir is in has at least one
// commit. A new repository has none until its first commit.
func HasCommits(dir string) (bool, error) {
	// With --verify --quiet, git exits with 0 when HEAD points at a commit,
	// and with 1, printing nothing, when there is no commit yet.
	args := []string{"rev-parse", "--verify", "--quiet", "HEAD"}
	_, err := command(dir, args...).Output()
	if err == nil {
		return true, nil
	}

	var exitErr *exec.ExitError
	if errors.As(err, &exitErr) && exitErr.ExitCode() == 1 {
		return false, nil
	}
	return false, gitError(args, err)
}

// Config returns the value of the git setting key, as git config --get
// prints it. ok is false when the setting is not set.
func Config(dir, key string) (value string, ok bool, err error) {
	// git config --get exits with 1 when the key is not set.
	args := []string{"config", "--get", key}
	out, err := command(dir, args...).Output()
	if err != nil {
		var exitErr *exec.ExitError
		if errors.As(err, &exitErr) && exitErr.ExitCode() == 1 {
			return "", false, nil
		}
		return "", false, gitError(args, err)
	}
	return strings.TrimSpace(string(out)), true, nil
}

// BranchExists reports whether the repository that dir is in has a local
// branch with this name. A branch of a remote, like origin/main, does not
// count.
func BranchExists(dir, branch string) (bool, error) {
	args := []string{"rev-parse", "--verify", "--quiet", "refs/heads/" + branch}
	_, err := command(dir, args...).Output()
	if err == nil {
		return true, nil
	}

	var exitErr *exec.ExitError
	if errors.As(err, &exitErr) && exitErr.ExitCode() == 1 {
		return false, nil
	}
	return false, gitError(args, err)
}

// MainRoot returns the top folder of the main checkout of the repository
// that dir is in (PD6). Inside a linked worktree, Root gives the worktree;
// MainRoot still gives the main checkout. The path has / between the parts.
func MainRoot(dir string) (string, error) {
	args := []string{"rev-parse", "--path-format=absolute", "--git-common-dir"}
	out, err := command(dir, args...).Output()
	if err != nil {
		return "", gitError(args, err)
	}
	// A repo made with --separate-git-dir, or a submodule, is not handled (PD6).
	// The common git folder is <main checkout>/.git, so its parent is the
	// main checkout.
	gitDir := filepath.FromSlash(strings.TrimSpace(string(out)))
	return filepath.ToSlash(filepath.Dir(gitDir)), nil
}

// RemoteBranches returns the remote-tracking branches named exactly
// <remote>/<branch>, as git last saw them, sorted. For branch "slice/x" it
// may return ["origin/slice/x"]. A branch like origin/feature/slice/x does
// not count. It never talks to the remote.
func RemoteBranches(dir, branch string) ([]string, error) {
	args := []string{"remote"}
	out, err := command(dir, args...).Output()
	if err != nil {
		return nil, gitError(args, err)
	}

	var found []string
	// A remote name can have a "/" in it, so ask git about each remote and
	// check the exact reference, instead of cutting names apart.
	for remote := range strings.SplitSeq(string(out), "\n") {
		remote = strings.TrimSpace(remote)
		if remote == "" {
			continue
		}
		// With --verify --quiet, git exits with 0 when the reference exists,
		// and with 1, printing nothing, when it does not.
		args := []string{"show-ref", "--verify", "--quiet", "refs/remotes/" + remote + "/" + branch}
		_, err := command(dir, args...).Output()
		if err == nil {
			found = append(found, remote+"/"+branch)
			continue
		}
		var exitErr *exec.ExitError
		if errors.As(err, &exitErr) && exitErr.ExitCode() == 1 {
			continue
		}
		return nil, gitError(args, err)
	}
	slices.Sort(found)
	return found, nil
}

// PathExists reports whether path is on the commit rev, e.g. a branch name.
// It works for files and for folders. A file at the path counts too; this
// is fine because the caller checks the name first.
func PathExists(dir, rev, path string) (bool, error) {
	args := []string{"ls-tree", "--name-only", rev, "--", path}
	out, err := command(dir, args...).Output()
	if err != nil {
		return false, gitError(args, err)
	}
	return strings.TrimSpace(string(out)) != "", nil
}

// CommitID returns the full id of the commit that rev names, e.g. a branch.
func CommitID(dir, rev string) (string, error) {
	args := []string{"rev-parse", "--verify", rev + "^{commit}"}
	out, err := command(dir, args...).Output()
	if err != nil {
		return "", gitError(args, err)
	}
	return strings.TrimSpace(string(out)), nil
}

// CreateBranch makes the branch at the commit. The branch tracks nothing.
func CreateBranch(dir, branch, commit string) error {
	args := []string{"branch", "--no-track", branch, commit}
	if _, err := command(dir, args...).Output(); err != nil {
		return gitError(args, err)
	}
	return nil
}

// AddWorktree checks out the existing branch in a new worktree at path.
// path may have / between the parts; git gets it in the form of the system.
func AddWorktree(dir, path, branch string) error {
	args := []string{"worktree", "add", filepath.FromSlash(path), branch}
	if _, err := command(dir, args...).Output(); err != nil {
		return gitError(args, err)
	}
	return nil
}

// RemoveWorktree removes the worktree at path, with its files and staged
// changes (--force). path may have / between the parts.
func RemoveWorktree(dir, path string) error {
	args := []string{"worktree", "remove", "--force", filepath.FromSlash(path)}
	if _, err := command(dir, args...).Output(); err != nil {
		return gitError(args, err)
	}
	return nil
}

// PruneWorktrees makes git forget worktrees whose folder is gone.
func PruneWorktrees(dir string) error {
	args := []string{"worktree", "prune"}
	if _, err := command(dir, args...).Output(); err != nil {
		return gitError(args, err)
	}
	return nil
}

// DeleteBranch deletes the local branch, even if it is not merged.
func DeleteBranch(dir, branch string) error {
	args := []string{"branch", "-D", branch}
	if _, err := command(dir, args...).Output(); err != nil {
		return gitError(args, err)
	}
	return nil
}

// Add stages the files at paths. It uses --force, so a .gitignore of the
// user cannot keep them out.
func Add(dir string, paths ...string) error {
	args := append([]string{"add", "--force", "--"}, paths...)
	if _, err := command(dir, args...).Output(); err != nil {
		return gitError(args, err)
	}
	return nil
}

// Commit makes a normal commit of the staged files. It leaves the author,
// the hooks, and the signing to the user's git settings.
func Commit(dir, msg string) error {
	args := []string{"commit", "-m", msg}
	if _, err := command(dir, args...).Output(); err != nil {
		return gitError(args, err)
	}
	return nil
}

// command prepares git with args, to run in dir. LC_ALL=C makes git print
// its messages in English, so sliceflow can read them on every computer.
func command(dir string, args ...string) *exec.Cmd {
	cmd := exec.Command("git", args...)
	cmd.Dir = dir
	cmd.Env = append(os.Environ(), "LC_ALL=C")
	return cmd
}

// CommandError is the error of a git command that failed (PD4). It keeps
// all of git's message lines, so a caller can read more than the first.
type CommandError struct {
	Command  string   // the git command without options, e.g. "commit" or "worktree add"
	ExitCode int      // git's exit code, or -1 when git could not start
	Lines    []string // the lines git printed on stderr, not empty, trimmed
	Err      error    // the underlying error
}

// Error returns the first line git printed, or "git <command> failed (exit
// code <n>)" when it printed nothing, or "run git: ..." when git could not
// start.
func (e *CommandError) Error() string {
	if e.ExitCode == -1 {
		return fmt.Sprintf("run git: %v", e.Err)
	}
	if len(e.Lines) > 0 {
		return e.Lines[0]
	}
	return fmt.Sprintf("git %s failed (exit code %d)", e.Command, e.ExitCode)
}

// Unwrap returns the underlying error.
func (e *CommandError) Unwrap() error {
	return e.Err
}

// gitError turns an error from running git with args into a *CommandError.
func gitError(args []string, err error) error {
	ce := &CommandError{Command: commandName(args), ExitCode: -1, Err: err}
	var exitErr *exec.ExitError
	if errors.As(err, &exitErr) {
		// git ran and failed. If not, git could not run at all, e.g. it is
		// not installed, and the exit code stays -1.
		ce.ExitCode = exitErr.ExitCode()
		for line := range strings.SplitSeq(string(exitErr.Stderr), "\n") {
			if line = strings.TrimSpace(line); line != "" {
				ce.Lines = append(ce.Lines, line)
			}
		}
	}
	return ce
}

// commandName returns the git command of args without options and values:
// the first word, and the second too when the first is "worktree".
func commandName(args []string) string {
	if args[0] == "worktree" && len(args) > 1 {
		return args[0] + " " + args[1]
	}
	return args[0]
}
