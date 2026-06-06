package sentinelnotatpackagelevel

import "errors"

var ErrGood = errors.New("good")

func BadCreate() error {
	var err error
	err = errors.New("bad") // want "sentinel errors must be declared at package level with var, not assigned in functions"
	return err
}

func BadCreateMultiple() error {
	var e1, e2 error
	e1 = errors.New("bad1") // want "sentinel errors must be declared at package level with var, not assigned in functions"
	e2 = errors.New("bad2") // want "sentinel errors must be declared at package level with var, not assigned in functions"
	if true {
		return e1
	}
	return e2
}

func GoodUsePackageLevel() error {
	return ErrGood
}

func GoodReuseError(existing error) error {
	err := existing // good - not errors.New()
	return err
}

func GoodMultipleAssignment() (error, error) {
	e1, e2 := ErrGood, ErrGood // good - not errors.New()
	return e1, e2
}
