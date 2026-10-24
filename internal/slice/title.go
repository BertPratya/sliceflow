package slice

import (
	"context"
	"fmt"

	"github.com/BertPratya/sliceflow/internal/eventlog"
	"github.com/BertPratya/sliceflow/internal/state"
)

// SetTitle saves slice.titled (Flow 6). It works in any phase. The commit
// holds only the log; the title never goes into a git argument.
func SetTitle(ctx context.Context, r Repo, name, title string) error {
	wt, err := openSlice(ctx, r, name)
	if err != nil {
		return err
	}
	_, _, err = saveEvent(ctx, r, name, wt, saveOp{
		Event: "slice.titled",
		Prepare: func(st state.State) (prepared, error) {
			if st.Phase == "unknown" {
				return prepared{}, &Error{Message: fmt.Sprintf("slice %s has no slice.created event, so its state is unknown. Nothing was changed.", name)}
			}
			if title == "" {
				return prepared{}, &Error{Message: "the title must not be empty. Nothing was changed."}
			}
			current := st.Title
			if current == "" {
				current = name
			}
			if title == current {
				return prepared{}, &Error{Message: fmt.Sprintf("the title of %s is already \"%s\". Nothing was changed.", name, title)}
			}
			return prepared{Data: eventlog.Titled{Title: title}, Message: name + ": title set"}, nil
		},
	})
	return err
}
