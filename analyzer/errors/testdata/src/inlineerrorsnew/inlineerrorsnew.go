package inlineerrorsnew

import "errors"

var ErrGood = errors.New("good error")

func BadReturn() error {
	return errors.New("bad") // want "errors.New should be assigned to a package-level var, not returned inline"
}

func GoodReturn() error {
	return ErrGood
}

func BadAssign() error {
	var x error
	x = errors.New("bad") // want "errors.New should be assigned to a package-level var, not inline"
	return x
}

func GoodAssign() error {
	x := ErrGood
	return x
}

func OtherCalls() {
	_ = errors.Join(nil) // good
}

func BadMultipleReturns() (string, error) {
	return "", errors.New("bad") // want "errors.New should be assigned to a package-level var, not returned inline"
}

func GoodMultipleReturns() (string, error) {
	return "", ErrGood
}
