package errornotlast

func BadReturns(x int) (error, string) { // want "error should be the last return value"
	return nil, ""
}

func GoodReturns() (string, error) {
	return "", nil
}
