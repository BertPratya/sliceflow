package project

import (
	"path/filepath"
	"runtime"
	"testing"
)

// The tests build every path from a temporary folder, so they are full
// paths on every system. On Windows, "/srv/wt" is not a full path.

// Test 1: each kind of setting gives the folder that R17 or R18 asks for.
func TestWorktreePath(t *testing.T) {
	base := t.TempDir()
	root := filepath.Join(base, "code", "shop")
	home := setHome(t, filepath.Join(base, "home"))

	// On Windows, "\" is a folder separator, so `~\wt` is in the home
	// folder. On other systems, it is a folder named `~\wt` (R18).
	backslashWant := filepath.Join(root, `~\wt`, "shop", "login")
	if runtime.GOOS == "windows" {
		backslashWant = filepath.Join(home, "wt", "shop", "login")
	}

	tests := []struct {
		about   string
		setting string
		want    string
	}{
		{"not set", "", filepath.Join(home, ".sliceflow", "worktrees", "shop", "login")},
		{"a full path", filepath.Join(base, "srv", "wt"), filepath.Join(base, "srv", "wt", "shop", "login")},
		{"a full path with . and ..", base + "/srv/./wt/../worktrees", filepath.Join(base, "srv", "worktrees", "shop", "login")},
		{"next to the project", "../wt", filepath.Join(base, "code", "wt", "shop", "login")},
		{"inside the project", "srv/wt", filepath.Join(root, "srv", "wt", "shop", "login")},
		{"only ~", "~", filepath.Join(home, "shop", "login")},
		{"~/ at the start", "~/wt", filepath.Join(home, "wt", "shop", "login")},
		{`~\ at the start`, `~\wt`, backslashWant},
		{"~ then a name", "~bob/wt", filepath.Join(root, "~bob", "wt", "shop", "login")},
		{"~ not at the start", "backup/~old", filepath.Join(root, "backup", "~old", "shop", "login")},
	}
	for _, tt := range tests {
		t.Run(tt.about, func(t *testing.T) {
			got, err := WorktreePath(root, tt.setting, "login")
			if err != nil {
				t.Fatalf("WorktreePath(%q) returned the error %v, want none", tt.setting, err)
			}
			if got != tt.want {
				t.Errorf("WorktreePath(%q) returned %q, want %q", tt.setting, got, tt.want)
			}
		})
	}
}

// Test 2: when the home folder is unknown, a setting that needs it gives
// an error, with "" as the path.
func TestWorktreePathWithNoHome(t *testing.T) {
	root := filepath.Join(t.TempDir(), "shop")
	setHome(t, "")

	for _, setting := range []string{"", "~", "~/wt"} {
		got, err := WorktreePath(root, setting, "login")
		if err == nil {
			t.Errorf("WorktreePath(%q) returned no error, want one", setting)
		}
		if got != "" {
			t.Errorf("WorktreePath(%q) returned %q with its error, want \"\"", setting, got)
		}
	}
}

// Test 3: when the home folder is unknown, a setting that does not need
// it still works.
func TestWorktreePathWithNoHomeAndAFullPath(t *testing.T) {
	base := t.TempDir()
	root := filepath.Join(base, "shop")
	setHome(t, "")
	setting := filepath.Join(base, "srv", "wt")

	got, err := WorktreePath(root, setting, "login")
	if err != nil {
		t.Fatalf("WorktreePath(%q) returned the error %v, want none", setting, err)
	}
	if want := filepath.Join(setting, "shop", "login"); got != want {
		t.Errorf("WorktreePath(%q) returned %q, want %q", setting, got, want)
	}
}

// setHome makes os.UserHomeDir return home for the rest of the test, and
// returns home. It sets HOME for Linux and macOS, and USERPROFILE for
// Windows. An empty home makes os.UserHomeDir fail.
func setHome(t *testing.T, home string) string {
	t.Helper()
	t.Setenv("HOME", home)
	t.Setenv("USERPROFILE", home)
	return home
}
