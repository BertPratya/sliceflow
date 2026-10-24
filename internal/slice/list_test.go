package slice

import (
	"context"
	"reflect"
	"testing"
	"time"

	"github.com/BertPratya/sliceflow/internal/testrepo"
)

func listNames(l SliceList) []string {
	names := []string{}
	for _, s := range l.Slices {
		names = append(names, s.Slice)
	}
	return names
}

// TestListOrder: same created time sorts a to z; no created time is last.
func TestListOrder(t *testing.T) {
	r := newRepo(t)
	ctx := context.Background()
	for _, n := range []string{"z-slice", "m-slice", "a-slice"} {
		if _, err := Create(ctx, r, CreateOptions{Name: n}); err != nil {
			t.Fatal(err)
		}
	}
	// Older than the others: first, although its name is last.
	r.Now = func() time.Time { return fixedNow.Add(-time.Hour) }
	if _, err := Create(ctx, r, CreateOptions{Name: "zz-oldest"}); err != nil {
		t.Fatal(err)
	}
	// Branches with no log: no created time; last, by name.
	testrepo.Git(t, r.Dir, "branch", "slice/b-nolog")
	testrepo.Git(t, r.Dir, "branch", "slice/a-nolog")
	got, err := List(ctx, r, false)
	if err != nil {
		t.Fatal(err)
	}
	want := []string{"zz-oldest", "a-slice", "m-slice", "z-slice", "a-nolog", "b-nolog"}
	if !reflect.DeepEqual(listNames(got), want) {
		t.Errorf("order = %v, want %v", listNames(got), want)
	}
	if len(got.Warnings) != 0 {
		t.Errorf("Warnings = %+v", got.Warnings)
	}
}

func TestListMergedAndAll(t *testing.T) {
	r := newRepo(t)
	ctx := context.Background()
	var wts []string
	for _, n := range []string{"first-slice", "second-slice"} {
		c, err := Create(ctx, r, CreateOptions{Name: n})
		if err != nil {
			t.Fatal(err)
		}
		wts = append(wts, c.Worktree)
	}
	testrepo.Git(t, r.Dir, "merge", "--no-ff", "-m", "merge", "slice/first-slice")
	// first-slice keeps its branch: its folder on the base makes it merged.
	got, err := List(ctx, r, false)
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(listNames(got), []string{"second-slice"}) {
		t.Errorf("in progress = %v", listNames(got))
	}
	// A merged slice with no branch is only in --all.
	testrepo.Git(t, r.Dir, "worktree", "remove", "--force", wts[0])
	testrepo.Git(t, r.Dir, "branch", "-D", "slice/first-slice")
	got, err = List(ctx, r, true)
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(listNames(got), []string{"first-slice", "second-slice"}) {
		t.Errorf("all = %v", listNames(got))
	}
	got, err = List(ctx, r, false)
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(listNames(got), []string{"second-slice"}) {
		t.Errorf("in progress = %v", listNames(got))
	}
}

func TestListBaseMissing(t *testing.T) {
	r := newRepo(t)
	ctx := context.Background()
	if _, err := Create(ctx, r, CreateOptions{Name: "spec-submit"}); err != nil {
		t.Fatal(err)
	}
	testrepo.Git(t, r.Dir, "branch", "-m", "main", "master")
	got, err := List(ctx, r, true)
	if err != nil {
		t.Fatal(err)
	}
	wantW := []Warning{{Text: "base branch main does not exist (the default); merged slices are not shown"}}
	if !reflect.DeepEqual(got.Warnings, wantW) || !reflect.DeepEqual(listNames(got), []string{"spec-submit"}) {
		t.Errorf("got %+v", got)
	}
}

func TestListBadBranchName(t *testing.T) {
	r := newRepo(t)
	ctx := context.Background()
	if _, err := Create(ctx, r, CreateOptions{Name: "spec-submit"}); err != nil {
		t.Fatal(err)
	}
	testrepo.Git(t, r.Dir, "branch", "slice/My_Thing")
	testrepo.Git(t, r.Dir, "branch", "slice/team/login")
	got, err := List(ctx, r, false)
	if err != nil {
		t.Fatal(err)
	}
	wantW := []Warning{
		{Text: "branch slice/My_Thing: not a valid slice name; skipped"},
		{Text: "branch slice/team/login: not a valid slice name; skipped"},
	}
	if !reflect.DeepEqual(got.Warnings, wantW) {
		t.Errorf("Warnings = %+v", got.Warnings)
	}
	if !reflect.DeepEqual(listNames(got), []string{"spec-submit"}) || len(got.Slices[0].Warnings) != 0 {
		t.Errorf("got %+v", got.Slices)
	}
}

func TestListNoSlices(t *testing.T) {
	r := newRepo(t)
	got, err := List(context.Background(), r, true)
	if err != nil {
		t.Fatal(err)
	}
	if got.Slices == nil || got.Warnings == nil || len(got.Slices) != 0 || len(got.Warnings) != 0 {
		t.Errorf("got %#v", got)
	}
}
