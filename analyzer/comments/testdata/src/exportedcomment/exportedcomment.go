package exportedcomment

// Good is a properly documented function.
func Good() {}

func NoComment() {}

// wrong comment that doesn't start with the function name // want "exported symbol.*must have a comment"
func Wrong() {}

// GoodFunction is a multi-line comment
// that explains the function in detail
// without any dividers or empty lines.
func GoodFunction() {}

// BrokenFunction has a divider // want "should not have empty lines"
//
// in the middle of the comment
func BrokenFunction() {}

// DividerFunction uses dashes // want "should not use repeated"
// --- as a divider
// more text here
func DividerFunction() {}

// GoodType represents something.
type GoodType struct{}

type NoCommentType struct{}
