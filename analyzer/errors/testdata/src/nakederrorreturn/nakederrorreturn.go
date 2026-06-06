package nakederrorreturn

import "fmt"

func BadReturn(x int) error {
	err := fmt.Errorf("something went wrong")
	return err // want "return error without wrapping: use fmt.Errorf with %w"
}

func GoodReturn(x int) error {
	err := fmt.Errorf("something went wrong")
	return fmt.Errorf("failed: %w", err)
}

func MultipleReturns() (string, error) {
	return "", nil // good
}

func BadMultipleReturns() (string, error) {
	err := fmt.Errorf("bad")
	return "", err // want "return error without wrapping: use fmt.Errorf with %w"
}

func NoError() int {
	return 42
}

func NamedErrReturn(x int) (result int, err error) {
	return 0, fmt.Errorf("error") // good - not a naked return
}
