package slice_test

import (
	"context"
	"testing"

	"github.com/BertPratya/sliceflow/internal/slice"
	"github.com/BertPratya/sliceflow/internal/testrepo"
)

func TestBaseBranch(t *testing.T) {
	dir := testrepo.New(t)
	ctx := context.Background()
	r, err := slice.FindRepo(ctx, dir)
	if err != nil {
		t.Fatal(err)
	}
	check := func(from, wantB, wantS string) {
		t.Helper()
		b, s, err := slice.BaseBranch(ctx, r, from)
		if err != nil {
			t.Fatal(err)
		}
		if b != wantB || s != wantS {
			t.Errorf("from %q: got %q, %q; want %q, %q", from, b, s, wantB, wantS)
		}
	}
	check("", "main", "the default")
	testrepo.Git(t, dir, "config", "sliceflow.base", "develop")
	check("", "develop", "from git config sliceflow.base")
	check("hotfix", "hotfix", "from --from")
}
