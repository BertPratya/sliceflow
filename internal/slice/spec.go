package slice

import (
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/BertPratya/sliceflow/templates"
)

// specEdit is one change that specFile makes to the template.
type specEdit struct {
	name string         // names the edit in the error
	find *regexp.Regexp // what to change; it must match exactly once
	with string         // the new text; $ is not special
}

// templates/default/spec.md and specEdits must change together.
// specEdits is the changes that turn the spec template into the built-in
// spec of a new slice (R7), in the order they run.
var specEdits = []specEdit{
	{
		name: "remove the title line",
		find: regexp.MustCompile(`title: \{\{title\}\}\n`),
		with: "",
	},
	{
		name: "remove the title rule",
		find: regexp.MustCompile(`     - The title of the slice lives only in this front matter\. Don't\n       repeat it in a heading or in other files\.\n`),
		with: "",
	},
	{
		name: "change the list of filled fields",
		find: regexp.MustCompile(`slice, title, and date`),
		with: "slice and date",
	},
}

// SpecTemplate returns the built-in spec template, as "spec template"
// prints it (0002 R28, PD13): the template with the edits of specEdits,
// and {{slice}} and {{date}} left as they are. It has \n line endings only.
func SpecTemplate() (string, error) {
	return editSpec(templates.Spec)
}

// editSpec runs the edits of specEdits on the template text tmpl, after it
// turns \r\n into \n. Every edit must match exactly once; if not, the
// error names the edit, so a changed template is noticed.
func editSpec(tmpl string) (string, error) {
	text := strings.ReplaceAll(tmpl, "\r\n", "\n")
	for _, e := range specEdits {
		if n := len(e.find.FindAllStringIndex(text, -1)); n != 1 {
			return "", fmt.Errorf("spec template: %q matched %d times, want 1", e.name, n)
		}
		text = e.find.ReplaceAllLiteralString(text, e.with)
	}
	return text, nil
}

// specFile makes the spec of a new slice from the template text tmpl (R7).
// It runs the same edits as SpecTemplate, so it drops the title. Then it
// fills slice and created (the date of created, in UTC). The result has \n
// line endings only.
func specFile(tmpl, name string, created time.Time) (string, error) {
	text, err := editSpec(tmpl)
	if err != nil {
		return "", err
	}

	date := created.UTC().Format("2006-01-02")
	for _, f := range []struct{ marker, value string }{
		{"{{slice}}", name},
		{"{{date}}", date},
	} {
		if n := strings.Count(text, f.marker); n != 1 {
			return "", fmt.Errorf("spec template: %q found %d times, want 1", f.marker, n)
		}
		text = strings.Replace(text, f.marker, f.value, 1)
	}
	return text, nil
}
