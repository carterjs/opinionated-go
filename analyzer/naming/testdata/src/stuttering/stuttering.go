package stuttering

// Bad: repeats package name
type StutteringType struct{} // want "is a stuttering name"

// Bad: repeats package name
func StutteringFunction() {} // want "is a stuttering name"

// Good: doesn't repeat package name
type Type struct{}

// Good: doesn't repeat package name
func Function() {}

// Good: unexported, no warning
type stutteringType struct{}

// Good: unexported, no warning
func stutteringFunction() {}

// Good: method, no warning
func (t Type) StutteringType() {}

// Good: exact match to the package name is the eponymous-type idiom, no warning
type Stuttering struct{}

// Good: continues the same word (an agent noun of the package name) rather
// than tacking on a new capitalized word, no warning
type Stutterer struct{}

// Good: same reasoning for a function name
func Stutteringly() {}
