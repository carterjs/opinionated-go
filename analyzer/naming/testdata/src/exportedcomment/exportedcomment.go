package exportedcomment

// Good is a properly documented function.
func Good() {}

func NoComment() {} // want "exported symbol.*must have a comment"

// wrong comment that doesn't start with the function name - want "exported symbol.*must have a comment"
func Wrong() {}

// GoodFunction is a multi-line comment
// that explains the function in detail
// without any dividers or empty lines.
func GoodFunction() {}

// BrokenFunction has a divider
//
// in the middle of the comment - want "should not have empty lines"
func BrokenFunction() {}

// DividerFunction uses dashes
// --- as a divider - want "should not use repeated"
// more text here
func DividerFunction() {}

// GoodType represents something.
type GoodType struct{}

type NoCommentType struct{} // want "exported symbol.*must have a comment"
