package e2e

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/BertPratya/sliceflow/internal/testrepo"
)

// gitCommand returns the git subcommand that a GIT_TRACE line runs, and the
// words after it, or "" when the line does not start a git command. It
// handles the lines "trace: built-in: git <args>", "trace: exec: git <args>"
// and "trace: run_command: git <args>", with or without quotes, and the
// helper form "git-<name>".
func gitCommand(line string) (string, []string) {
	for _, kind := range []string{"trace: built-in: ", "trace: exec: ", "trace: run_command: "} {
		i := strings.Index(line, kind)
		if i < 0 {
			continue
		}
		fields := strings.Fields(line[i+len(kind):])
		for j := range fields {
			fields[j] = strings.Trim(fields[j], `'"`)
		}
		if len(fields) == 0 {
			return "", nil
		}
		prog := filepath.Base(fields[0])
		prog = strings.TrimSuffix(prog, ".exe")
		if strings.HasPrefix(prog, "git-") {
			return strings.TrimPrefix(prog, "git-"), fields[1:]
		}
		if prog != "git" {
			// A helper such as "remote-curl" run by name.
			return prog, fields[1:]
		}
		args := fields[1:]
		for k := 0; k < len(args); k++ {
			a := args[k]
			switch {
			case a == "-c" || a == "-C" || a == "--git-dir" || a == "--work-tree" || a == "--namespace":
				k++
			case strings.HasPrefix(a, "-"):
			default:
				return a, args[k+1:]
			}
		}
		return "", nil
	}
	return "", nil
}

// forbiddenGitCommand reports whether a git command can connect to a remote.
// Plain "git remote" and "git remote -v" only read the config, so they stay
// allowed. "git remote show|update|prune|set-head" connect.
func forbiddenGitCommand(sub string, rest []string) bool {
	switch sub {
	case "fetch", "pull", "push", "ls-remote",
		"upload-pack", "receive-pack", "fetch-pack", "send-pack":
		return true
	case "remote":
		for _, a := range rest {
			if strings.HasPrefix(a, "-") {
				continue
			}
			switch a {
			case "show", "update", "prune", "set-head":
				return true
			}
			return false
		}
		return false
	}
	return strings.HasPrefix(sub, "remote-")
}

func TestGitSubcommand(t *testing.T) {
	const b = "12:00:00.000000 git.c:460 trace: built-in: "
	cases := []struct {
		line string
		sub  string
		bad  bool
	}{
		{b + "git fetch origin", "fetch", true},
		{b + "git -c a=b pull", "pull", true},
		{b + "git push origin main", "push", true},
		{"12:00:00.000000 run-command.c:657 trace: run_command: 'git-remote-https' 'origin' 'https://x'", "remote-https", true},
		{"12:00:00.000000 run-command.c:657 trace: run_command: git remote-curl origin", "remote-curl", true},
		{"12:00:00.000000 git.c:460 trace: exec: git-remote-https origin", "remote-https", true},
		{b + "git ls-remote origin", "ls-remote", true},
		{"12:00:00.000000 git.c:460 trace: exec: git-upload-pack /tmp/r", "upload-pack", true},
		{b + "git upload-pack /tmp/r", "upload-pack", true},
		{b + "git receive-pack /tmp/r", "receive-pack", true},
		{b + "git fetch-pack /tmp/r", "fetch-pack", true},
		{b + "git send-pack /tmp/r", "send-pack", true},
		{b + "git remote show origin", "remote", true},
		{b + "git remote update", "remote", true},
		{b + "git remote prune origin", "remote", true},
		{b + "git remote set-head origin -a", "remote", true},
		{b + "git -C /tmp/r remote show origin", "remote", true},
		{b + "git remote", "remote", false},
		{b + "git remote -v", "remote", false},
		{b + "git remote add origin /tmp/push/fetch", "remote", false},
		{b + "git remote add show /tmp/r", "remote", false},
		{b + "git -C /tmp/push status", "status", false},
		{b + "git commit -m fetch", "commit", false},
		{b + "git ls-files", "ls-files", false},
		{"12:00:00.000000 something else", "", false},
	}
	for _, c := range cases {
		sub, rest := gitCommand(c.line)
		if sub != c.sub || forbiddenGitCommand(sub, rest) != c.bad {
			t.Errorf("%q: got %q (forbidden %v), want %q (%v)", c.line, sub, forbiddenGitCommand(sub, rest), c.sub, c.bad)
		}
	}
}

// R40: no command runs git fetch, git pull, git push, or a remote helper,
// also when a remote is set up.
func TestNoNetwork(t *testing.T) {
	dir := testrepo.New(t)
	missing := filepath.Join(t.TempDir(), "no-such-remote")
	testrepo.Git(t, dir, "remote", "add", "origin", missing)

	trace := filepath.Join(t.TempDir(), "trace.txt")
	t.Setenv("GIT_TRACE", trace)

	for _, args := range [][]string{
		{"slice", "create", "a"},
		{"slice", "create", "b", "--json"},
		{"slice", "status", "a"},
		{"slice", "list", "--all"},
	} {
		stdout, stderr, code := run(t, dir, args...)
		if code != 0 {
			t.Fatalf("%v: exit %d\nstdout: %s\nstderr: %s", args, code, stdout, stderr)
		}
	}

	b, err := os.ReadFile(trace)
	if err != nil {
		t.Fatal(err)
	}
	// The test's own setup also writes git lines into the trace, so the guard
	// needs lines only the binary writes: "slice create" runs git worktree,
	// and "slice list" runs git for-each-ref. If git does not trace for the
	// binary, a fetch would hide, so the test must fail.
	seen := map[string]bool{}
	for _, line := range strings.Split(string(b), "\n") {
		sub, rest := gitCommand(line)
		seen[sub] = true
		if forbiddenGitCommand(sub, rest) {
			t.Errorf("forbidden git command in trace: %s", line)
		}
	}
	for _, want := range []string{"worktree", "for-each-ref"} {
		if !seen[want] {
			t.Errorf("the trace has no git %s line from the binary; GIT_TRACE does not work?\n%s", want, b)
		}
	}
}

// R40 of 0001: the commands of 0002 run no fetch, pull, or push either.
func TestNoNetworkNewCommands(t *testing.T) {
	dir := draftSlice(t, s15Text())
	testrepo.Git(t, dir, "remote", "add", "origin", filepath.Join(t.TempDir(), "no-such-remote"))

	trace := filepath.Join(t.TempDir(), "trace.txt")
	t.Setenv("GIT_TRACE", trace)

	for _, args := range [][]string{
		{"spec", "validate"},
		{"spec", "submit"},
		{"spec", "review"},
		{"slice", "title", specName, "Sign in"},
		{"template", "show", "spec"},
	} {
		stdout, stderr, code := run(t, wtOf(dir), args...)
		if code != 0 {
			t.Fatalf("%v: exit %d\nstdout: %s\nstderr: %s", args, code, stdout, stderr)
		}
	}

	b, err := os.ReadFile(trace)
	if err != nil {
		t.Fatal(err)
	}
	seen := map[string]bool{}
	for _, line := range strings.Split(string(b), "\n") {
		sub, rest := gitCommand(line)
		seen[sub] = true
		if forbiddenGitCommand(sub, rest) {
			t.Errorf("forbidden git command in trace: %s", line)
		}
	}
	// Only the binary ran after GIT_TRACE was set, so its commits must show.
	if !seen["commit"] {
		t.Errorf("the trace has no git commit line from the binary; GIT_TRACE does not work?\n%s", b)
	}
}
