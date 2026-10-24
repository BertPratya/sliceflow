// Package template holds the document templates built into sliceflow.
package template

import (
	_ "embed"
	"strings"
	"time"
)

//go:embed spec.md
var specTemplate string

// Spec returns the spec template with its front matter filled for a new
// slice. created is written in UTC as YYYY-MM-DD.
func Spec(name string, created time.Time) []byte {
	// Drop the template's own front matter: it ends at the second "---" line.
	body := specTemplate
	if strings.HasPrefix(body, "---\n") {
		if i := strings.Index(body[4:], "\n---\n"); i >= 0 {
			body = body[4+i+len("\n---\n"):]
		}
	}
	front := "---\nslice: " + name + "\ncreated: " + created.UTC().Format("2006-01-02") + "\nideas: []\n---\n"
	return []byte(front + body)
}
