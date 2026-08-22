// Package exportedcomment tests comment format rules.
package exportedcomment

// Good is a properly documented function.
func Good() {}

func NoComment() {} // want "exported symbol.*must have a doc comment"

// wrong comment that doesn't start with the function name // want "exported symbol.*must have a comment starting"
func Wrong() {}

// GoodFunction is a multi-line comment
// that explains the function in detail
// without any dividers or empty lines.
func GoodFunction() {}

// ParagraphFunction separates paragraphs with an empty comment line.
//
// The empty line is fine; it is how godoc marks a paragraph break.
func ParagraphFunction() {}

// DividerFunction uses dashes // want "should not use repeated"
// --- as a divider
// more text here
func DividerFunction() {}

// GoodType represents something.
type GoodType struct{}

type NoCommentType struct{} // want "exported symbol.*must have a doc comment"

// ExportedConst is documented.
const ExportedConst = true

// GoodConstant is properly documented.
const GoodConstant = 1

var ExportedVar = 0 // want "exported symbol.*must have a doc comment"

// GoodVariable is properly documented.
var GoodVariable = ""

// +build ignore // want "use //go:build instead"
const _ = 0
