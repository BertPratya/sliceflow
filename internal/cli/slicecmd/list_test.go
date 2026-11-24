package slicecmd

import (
	"testing"

	"github.com/BertPratya/sliceflow/internal/slice"
)

// listTable pads the name and the title columns to their longest value,
// the header included, counted in characters, not bytes. The description
// is not padded, and no line ends with a space.
func TestListTable(t *testing.T) {
	tests := []struct {
		about  string
		slices []slice.ListedSlice
		want   string
	}{
		{"the header is the widest", []slice.ListedSlice{{Name: "a", Title: "b", Description: "c"}},
			"NAME  TITLE  DESCRIPTION\n" +
				"a     b      c\n"},
		{"a value is the widest, and one row has no description", []slice.ListedSlice{
			{Name: "card-payment", Title: "Pay with a card", Description: "Let a customer pay."},
			{Name: "old-search", Title: "Search by name"},
		},
			"NAME          TITLE            DESCRIPTION\n" +
				"card-payment  Pay with a card  Let a customer pay.\n" +
				"old-search    Search by name\n"},
		{"characters, not bytes", []slice.ListedSlice{
			{Name: "cafe", Title: "Café menu", Description: "x"},
			{Name: "tea", Title: "Tea", Description: "y"},
		},
			"NAME  TITLE      DESCRIPTION\n" +
				"cafe  Café menu  x\n" +
				"tea   Tea        y\n"},
	}
	for _, tt := range tests {
		t.Run(tt.about, func(t *testing.T) {
			if got := listTable(tt.slices); got != tt.want {
				t.Errorf("listTable gave:\n%s\nwant:\n%s", got, tt.want)
			}
		})
	}
}
