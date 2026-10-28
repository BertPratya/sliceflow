package slice

import (
	"strings"
	"testing"
	"time"

	"github.com/BertPratya/sliceflow/templates"
)

func TestSpecFile(t *testing.T) {
	// 23:30 in UTC-5 is already the next day in UTC.
	created := time.Date(2026, 10, 24, 23, 30, 0, 0, time.FixedZone("x", -5*3600))

	got, err := specFile(templates.Spec, "spec-submit", created)
	if err != nil {
		t.Fatal(err)
	}

	front := "---\nslice: spec-submit\ncreated: 2026-10-25\n---\n"
	if !strings.HasPrefix(got, front) {
		t.Errorf("front matter is wrong, the spec starts with:\n%s", got[:min(len(got), 200)])
	}
	for _, bad := range []string{"{{slice}}", "{{date}}", "{{title}}", "title:", "The title of the slice lives only", "slice, title"} {
		if strings.Contains(got, bad) {
			t.Errorf("spec still has %q", bad)
		}
	}
	if !strings.Contains(got, "sliceflow fills slice and date itself") {
		t.Error("spec lost the changed sentence 'slice and date'")
	}
	if strings.Contains(got, "\r") {
		t.Error("spec has a \\r")
	}
}

func TestSpecFileCRLF(t *testing.T) {
	crlf := strings.ReplaceAll(templates.Spec, "\n", "\r\n")
	got, err := specFile(crlf, "x", time.Now())
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(got, "\r") {
		t.Error("spec has a \\r")
	}
}

func TestSpecFileBrokenTemplate(t *testing.T) {
	good := templates.Spec
	breaks := map[string]string{
		"remove the title line":            "title: {{title}}\n",
		"remove the title rule":            "The title of the slice lives only",
		"change the list of filled fields": "slice, title, and date",
		"{{slice}}":                        "{{slice}}",
		"{{date}}":                         "{{date}}",
	}
	for want, cut := range breaks {
		t.Run(want, func(t *testing.T) {
			if !strings.Contains(good, cut) {
				t.Fatalf("test is out of date: template has no %q", cut)
			}
			broken := strings.Replace(good, cut, "", 1)
			_, err := specFile(broken, "x", time.Now())
			if err == nil || !strings.Contains(err.Error(), want) {
				t.Errorf("error = %v, want it to name %q", err, want)
			}
		})
	}
}
