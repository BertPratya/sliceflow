package cli

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"maps"
	"path/filepath"
	"testing"

	"github.com/BertPratya/sliceflow/internal/slice"
)

// readLine reads args with checkLine, and fails the test now when the
// command line has a problem or asks for help.
func readLine(t *testing.T, args ...string) line {
	t.Helper()
	found, err := checkLine(commands, args)
	if err != nil || found.wantsHelp {
		t.Fatalf("checkLine %q: problem %v, help %v; want neither", args, err, found.wantsHelp)
	}
	return found
}

// TestCreateOptions checks R71: options before or after the name,
// --option=<text>, the word after an option is its value, and the last
// one counts. An option that was not given has no key, so `--title=`
// (given, "") is told from no --title.
func TestCreateOptions(t *testing.T) {
	tests := []struct {
		name        string
		args        []string
		wantName    string
		wantOptions map[string]string
	}{
		{"no options",
			[]string{"slice", "create", "login"},
			"login", map[string]string{}},
		{"options after the name",
			[]string{"slice", "create", "login", "--title", "Log in", "--description", "A form.", "--from", "dev"},
			"login", map[string]string{"title": "Log in", "description": "A form.", "from": "dev"}},
		{"options before the name",
			[]string{"slice", "create", "--title", "Log in", "--from", "dev", "login"},
			"login", map[string]string{"title": "Log in", "from": "dev"}},
		{"option with =",
			[]string{"slice", "create", "login", "--title=Log in", "--description=A form.", "--from=dev"},
			"login", map[string]string{"title": "Log in", "description": "A form.", "from": "dev"}},
		{"empty value with =",
			[]string{"slice", "create", "login", "--title="},
			"login", map[string]string{"title": ""}},
		{"value with = inside",
			[]string{"slice", "create", "login", "--title=a=b"},
			"login", map[string]string{"title": "a=b"}},
		{"value that starts with -",
			[]string{"slice", "create", "login", "--title", "-x"},
			"login", map[string]string{"title": "-x"}},
		{"value that is --help",
			[]string{"slice", "create", "login", "--title", "--help"},
			"login", map[string]string{"title": "--help"}},
		{"value that is -h",
			[]string{"slice", "create", "--description", "-h", "login"},
			"login", map[string]string{"description": "-h"}},
		{"value that is a command word",
			[]string{"slice", "create", "--title", "create", "slice"},
			"slice", map[string]string{"title": "create"}},
		{"last one counts",
			[]string{"slice", "create", "login", "--title", "A", "--title=B", "--from", "x", "--from", "y"},
			"login", map[string]string{"title": "B", "from": "y"}},
		{"name not checked here",
			[]string{"slice", "create", "Bad Name!", "--title", "a\tb"},
			"Bad Name!", map[string]string{"title": "a\tb"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			found := readLine(t, tt.args...)
			if found.words != "slice create" {
				t.Fatalf("checkLine found %q, want %q", found.words, "slice create")
			}
			if found.call.name != tt.wantName {
				t.Errorf("name = %q, want %q", found.call.name, tt.wantName)
			}
			if !maps.Equal(found.call.options, tt.wantOptions) {
				t.Errorf("options = %q, want %q", found.call.options, tt.wantOptions)
			}
		})
	}
}

// TestStatusAndList checks that status gets the given name, and list
// none.
func TestStatusAndList(t *testing.T) {
	found := readLine(t, "slice", "status", "Login")
	if found.words != "slice status" || found.call.name != "Login" {
		t.Errorf("checkLine found %q with name %q, want %q with %q", found.words, found.call.name, "slice status", "Login")
	}

	found = readLine(t, "slice", "list")
	if found.words != "slice list" || found.call.name != "" {
		t.Errorf("checkLine found %q with name %q, want %q with none", found.words, found.call.name, "slice list")
	}
}

// wrongName is an error of another package that is wrong use, as the
// create flow's errors for R20 and R29 will be.
type wrongName struct{}

func (wrongName) Error() string {
	return "Error: invalid slice name \"x y\". Nothing was changed.\nA second line."
}
func (wrongName) WrongUse() bool { return true }

// notWrong has a WrongUse method that returns false.
type notWrong struct{}

func (notWrong) Error() string  { return "Error: it failed." }
func (notWrong) WrongUse() bool { return false }

// TestReport checks R76 for errors from a command's work: wrong use
// gives 2, also when wrapped, and any other error 1. The error's text is
// printed as it is, with a line break.
func TestReport(t *testing.T) {
	tests := []struct {
		name     string
		err      error
		wantCode int
		wantText string
	}{
		{"no error", nil, 0, ""},
		{"wrong use", wrongName{}, 2, "Error: invalid slice name \"x y\". Nothing was changed.\nA second line.\n"},
		{"wrapped wrong use", fmt.Errorf("wrapped: %w", wrongName{}), 2, "wrapped: Error: invalid slice name \"x y\". Nothing was changed.\nA second line.\n"},
		{"WrongUse false", notWrong{}, 1, "Error: it failed.\n"},
		{"plain error", errors.New("Error: git failed."), 1, "Error: git failed.\n"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var stderr bytes.Buffer
			if code := report(&stderr, tt.err); code != tt.wantCode {
				t.Errorf("exit code = %d, want %d", code, tt.wantCode)
			}
			if stderr.String() != tt.wantText {
				t.Errorf("stderr = %q, want %q", stderr.String(), tt.wantText)
			}
		})
	}
}

// TestPlaceholders checks that each command whose work is not built yet
// fails with exit code 1. TEMPORARY: remove a case when its step is
// built.
func TestPlaceholders(t *testing.T) {
	tests := []struct {
		args []string
		want string
	}{
		{[]string{"slice", "status", "x"}, "Error: sliceflow slice status is not built yet.\n"},
		{[]string{"slice", "list"}, "Error: sliceflow slice list is not built yet.\n"},
	}
	for _, tt := range tests {
		var stdout, stderr bytes.Buffer
		code := Run(context.Background(), tt.args, &stdout, &stderr)
		if code != 1 || stderr.String() != tt.want || stdout.Len() != 0 {
			t.Errorf("Run %q: code %d, stdout %q, stderr %q; want 1, nothing, %q",
				tt.args, code, stdout.String(), stderr.String(), tt.want)
		}
	}
}

// TestPrintCreated checks R5, R8, and R22: the lines of a slice that
// was made, in their order, with both notes.
func TestPrintCreated(t *testing.T) {
	worktree := filepath.Join("home", "you", ".sliceflow", "worktrees", "shop", "card-payment")
	spec := filepath.Join(worktree, "docs", "slices", "card-payment", "spec.md")
	c := slice.Created{
		Name:         "card-payment",
		Lowered:      true,
		Base:         "main",
		BaseCommit:   "3f9a1c2e8b7d4a6f0c1e5d9b2a7f3c8e4d6b1a09",
		Worktree:     worktree,
		SpecPath:     spec,
		IgnoredFiles: []string{"spec.md", ".events.jsonl"},
	}
	want := "Created slice card-payment.\n" +
		"  Note:     \"Card-Payment\" was made lowercase.\n" +
		"  Note:     spec.md matches a .gitignore rule, so sliceflow added it anyway.\n" +
		"  Note:     .events.jsonl matches a .gitignore rule, so sliceflow added it anyway.\n" +
		"  Branch:   slice/card-payment (from main, at 3f9a1c2)\n" +
		"  Worktree: " + worktree + "\n" +
		"Next: write the spec in " + spec + "\n"

	var out bytes.Buffer
	if err := printCreated(&out, "Card-Payment", c); err != nil {
		t.Fatalf("printCreated: %v", err)
	}
	if out.String() != want {
		t.Errorf("printCreated printed:\n%s\nwant:\n%s", out.String(), want)
	}
}

// TestPrintCreatedNoNotes checks R8 and R22: with no lowercase and no
// ignored file, there is no Note line.
func TestPrintCreatedNoNotes(t *testing.T) {
	c := slice.Created{
		Name:       "login",
		Base:       "dev",
		BaseCommit: "8c41d07aa",
		Worktree:   "wt",
		SpecPath:   "spec",
	}
	want := "Created slice login.\n" +
		"  Branch:   slice/login (from dev, at 8c41d07)\n" +
		"  Worktree: wt\n" +
		"Next: write the spec in spec\n"
	var out bytes.Buffer
	if err := printCreated(&out, "login", c); err != nil {
		t.Fatalf("printCreated: %v", err)
	}
	if out.String() != want {
		t.Errorf("printCreated printed:\n%s\nwant:\n%s", out.String(), want)
	}
}
