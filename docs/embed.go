// Package docs holds the pages of the documentation that the ant command prints itself.
package docs

import _ "embed"

// Check is the page of ant check, whose ## Codes section ant check -explain prints from.
//
//go:embed tools/check.md
var Check string
