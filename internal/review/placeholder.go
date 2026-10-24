package review

import (
	"context"
	"os"
)

const placeholderText = "# Review\n\n" +
	"## Summary\n\nNo review was done: placeholder reviewer.\n\n" +
	"## Checked\n\nNone.\n\n" +
	"## Tests\n\nNone.\n\n" +
	"## Previous findings\n\nNone.\n\n" +
	"## Findings\n\nNone.\n\n" +
	"## Requirements check\n\nNone.\n\n" +
	"## Needs human decision\n\nNone.\n"

type placeholder struct{}

// Placeholder returns the reviewer that reads nothing and always passes.
func Placeholder() Reviewer { return placeholder{} }

func (placeholder) Name() string { return "placeholder" }

// Review appends the placeholder text after the front matter in req.File.
func (placeholder) Review(ctx context.Context, req Request) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	f, err := os.OpenFile(req.File, os.O_WRONLY|os.O_APPEND, 0o644)
	if err != nil {
		return err
	}
	if _, err := f.WriteString(placeholderText); err != nil {
		f.Close()
		return err
	}
	return f.Close()
}
