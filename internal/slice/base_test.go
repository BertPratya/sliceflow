package slice

import (
	"testing"

	"github.com/BertPratya/sliceflow/internal/gittest"
)

func TestBaseBranch(t *testing.T) {
	tests := []struct {
		name     string
		branches []string // made in the repo, besides main
		config   string   // the value of sliceflow.base, or "" for not set
		from     string   // the --from flag, or "" for not given
		want     string   // the base branch, or "" when an error is wanted
		wantErr  string   // the exact error text, or "" for no error
	}{
		{name: "default is main", want: "main"},
		{name: "git config", branches: []string{"develop"}, config: "develop", want: "develop"},
		{name: "--from", branches: []string{"hotfix"}, from: "hotfix", want: "hotfix"},

		{name: "S3: --from wins over git config", branches: []string{"develop", "hotfix"}, config: "develop", from: "hotfix", want: "hotfix"},

		{
			name:    "S8: git config branch does not exist",
			config:  "develop",
			wantErr: "base branch develop does not exist (from git config sliceflow.base)",
		},
		{
			name:    "--from branch does not exist",
			from:    "nope",
			wantErr: "base branch nope does not exist (from --from)",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := gittest.NewRepo(t)
			for _, b := range tt.branches {
				gittest.Run(t, repo, "branch", b)
			}
			if tt.config != "" {
				gittest.Run(t, repo, "config", "sliceflow.base", tt.config)
			}

			got, err := baseBranch(repo, tt.from)

			gotErr := ""
			if err != nil {
				gotErr = err.Error()
			}
			if got != tt.want || gotErr != tt.wantErr {
				t.Errorf("baseBranch(from %q) = %q, %q, want %q, %q", tt.from, got, gotErr, tt.want, tt.wantErr)
			}
		})
	}
}

// TestBaseBranchDefaultMissing is apart from the table, because it needs a
// repository without a branch main.
func TestBaseBranchDefaultMissing(t *testing.T) {
	repo := gittest.NewRepo(t)
	gittest.Run(t, repo, "branch", "-m", "main", "trunk")

	_, err := baseBranch(repo, "")

	want := "base branch main does not exist (the default)"
	if err == nil || err.Error() != want {
		t.Errorf("baseBranch(no main) error = %v, want %q", err, want)
	}
}
