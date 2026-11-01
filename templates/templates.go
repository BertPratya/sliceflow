// Package templates holds the templates that are built into sliceflow.
package templates

import _ "embed"

// Spec is the text of templates/default/spec.md, built into the program
// (PD5). The slice package makes the spec of a new slice from it.
//
//go:embed default/spec.md
var Spec string
