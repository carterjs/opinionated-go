package nakederrorreturn

import "fmt"

// BadReturn hands the error back across the package boundary untouched.
func BadReturn() error { // want "BadReturn returns an error without wrapping"
	err := fmt.Errorf("something went wrong")
	return err
}

// GoodReturn says what it was doing.
func GoodReturn() error {
	err := fmt.Errorf("something went wrong")
	return fmt.Errorf("failed: %w", err)
}

// BadMultipleReturns is reported for the error result only.
func BadMultipleReturns() (string, error) { // want "BadMultipleReturns returns an error without wrapping in 2 places"
	err := fmt.Errorf("bad")
	if err != nil {
		return "", err
	}
	return "", err
}

// MultipleReturns returns no error to wrap.
func MultipleReturns() (string, error) {
	return "", nil
}

// NoError has nothing to report.
func NoError() int {
	return 42
}

// NamedErrReturn wraps at the call site.
func NamedErrReturn() (result int, err error) {
	return 0, fmt.Errorf("error")
}

// Delegate is exported, and a closure inside it is still on the boundary.
func Delegate() error { // want "Delegate returns an error without wrapping"
	run := func() error {
		err := fmt.Errorf("inner")
		return err
	}
	return run()
}

// helper is unexported: its caller is a few lines away and adds the context.
func helper() error {
	err := fmt.Errorf("something went wrong")
	return err
}

type store struct{}

// Load is exported even though its receiver is not.
func (s *store) Load() error { // want "Load returns an error without wrapping"
	err := helper()
	return err
}

func (s *store) load() error {
	err := helper()
	return err
}
