package git_test

import (
	"context"
	"errors"
	"testing"

	"github.com/BertPratya/sliceflow/internal/git"
	"github.com/BertPratya/sliceflow/internal/gittest"
)

// Test 1: ConfigValue tells a setting that is not set from one that is
// set, also to "", and keeps spaces inside the value. Only the line
// break at the end is removed.
func TestConfigValue(t *testing.T) {
	r := gittest.New(t)

	value, set, err := git.ConfigValue(context.Background(), r.Dir, "test.key")
	if err != nil || set || value != "" {
		t.Errorf("not set: ConfigValue = %q, %v, %v, want \"\", false, no error", value, set, err)
	}
	for _, want := range []string{"", "a b", "~/worktrees"} {
		r.SetConfig("test.key", want)
		value, set, err := git.ConfigValue(context.Background(), r.Dir, "test.key")
		if err != nil || !set || value != want {
			t.Errorf("ConfigValue = %q, %v, %v, want %q, true, no error", value, set, err, want)
		}
	}
}

// Test 2: a value set for the project wins over a global one, and a
// global one is used when the project has none (R18).
func TestConfigValueProjectWins(t *testing.T) {
	r := gittest.New(t)
	gittest.SetGlobalConfig(t, "test.key", "global")

	if value, _, _ := git.ConfigValue(context.Background(), r.Dir, "test.key"); value != "global" {
		t.Errorf("global only: ConfigValue = %q, want %q", value, "global")
	}
	r.SetConfig("test.key", "project")
	if value, _, _ := git.ConfigValue(context.Background(), r.Dir, "test.key"); value != "project" {
		t.Errorf("project and global: ConfigValue = %q, want %q", value, "project")
	}
}

// Test 3: when git fails with a code that is not 1, here because the
// project's config file is broken, ConfigValue returns git's error, not
// "not set".
func TestConfigValueBrokenConfig(t *testing.T) {
	r := gittest.New(t)
	r.BreakConfig()

	value, set, err := git.ConfigValue(context.Background(), r.Dir, "test.key")
	var gitErr *git.Error
	if !errors.As(err, &gitErr) {
		t.Errorf("ConfigValue = %q, %v, %v (%T), want a *git.Error", value, set, err, err)
	}
}
