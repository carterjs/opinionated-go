package panicinnonmain

func BadPanic() {
	panic("something bad") // want "panic is not allowed in library code; return an error instead"
}

func BadPanicError(err error) {
	if err != nil {
		panic(err) // want "panic is not allowed in library code; return an error instead"
	}
}

func GoodError() error {
	return nil
}

func BadPanicRecover() {
	defer func() {
		panic("recovered but panics again") // want "panic is not allowed in library code; return an error instead"
	}()
}

func GoodOtherFunction() {
	_ = func() {} // ok
}
