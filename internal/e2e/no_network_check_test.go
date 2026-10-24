package e2e

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/BertPratya/sliceflow/internal/testrepo"
)

// gitSubcommand returns the git subcommand that a GIT_TRACE line runs, or ""
// when the line does not start a git command. It handles the lines
// "trace: built-in: git <args>", "trace: exec: git <args>" and
// "trace: run_command: git <args>", with or without quotes, and the helper
// form "git-<name>".
func gitSubcommand(line string) string {
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
			return ""
		}
		prog := filepath.Base(fields[0])
		prog = strings.TrimSuffix(prog, ".exe")
		if strings.HasPrefix(prog, "git-") {
			return strings.TrimPrefix(prog, "git-")
		}
		if prog != "git" {
			// A helper such as "remote-curl" or "git-remote-https" run by name.
			return prog
		}
		args := fields[1:]
		for k := 0; k < len(args); k++ {
			a := args[k]
			switch {
			case a == "-c" || a == "-C" || a == "--git-dir" || a == "--work-tree" || a == "--namespace":
				k++
			case strings.HasPrefix(a, "-"):
			default:
				return a
			}
		}
		return ""
	}
	return ""
}

func forbiddenGitCommand(sub string) bool {
	switch sub {
	case "fetch", "pull", "push":
		return true
	}
	return strings.HasPrefix(sub, "remote-")
}

func TestGitSubcommand(t *testing.T) {
	cases := []struct {
		line string
		sub  string
		bad  bool
	}{
		{"12:00:00.000000 git.c:460 trace: built-in: git fetch origin", "fetch", true},
		{"12:00:00.000000 git.c:460 trace: built-in: git -c a=b pull", "pull", true},
		{"12:00:00.000000 git.c:460 trace: built-in: git push origin main", "push", true},
		{"12:00:00.000000 run-command.c:657 trace: run_command: 'git-remote-https' 'origin' 'https://x'", "remote-https", true},
		{"12:00:00.000000 run-command.c:657 trace: run_command: git remote-curl origin", "remote-curl", true},
		{"12:00:00.000000 git.c:460 trace: exec: git-remote-https origin", "remote-https", true},
		{"12:00:00.000000 git.c:460 trace: built-in: git remote add origin /tmp/push/fetch", "remote", false},
		{"12:00:00.000000 git.c:460 trace: built-in: git -C /tmp/push status", "status", false},
		{"12:00:00.000000 git.c:460 trace: built-in: git commit -m fetch", "commit", false},
		{"12:00:00.000000 something else", "", false},
	}
	for _, c := range cases {
		sub := gitSubcommand(c.line)
		if sub != c.sub || forbiddenGitCommand(sub) != c.bad {
			t.Errorf("%q: got %q (forbidden %v), want %q (%v)", c.line, sub, forbiddenGitCommand(sub), c.sub, c.bad)
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
	lines := strings.Split(string(b), "\n")
	seen := 0
	for _, line := range lines {
		sub := gitSubcommand(line)
		if sub != "" {
			seen++
		}
		if forbiddenGitCommand(sub) {
			t.Errorf("forbidden git command in trace: %s", line)
		}
	}
	if seen == 0 {
		t.Fatalf("the trace has no git command; GIT_TRACE does not work?\n%s", b)
	}
}
