package namedreturnvalues

func BadNamed() (result int, err error) { // want "named return values are banned; return values must be unnamed" "named return values are banned; return values must be unnamed"
	return 0, nil
}

func BadPartiallyNamed() (x int, y error) { // want "named return values are banned; return values must be unnamed" "named return values are banned; return values must be unnamed"
	return 0, nil
}

func BadMultipleNamed() (a int, b string, c error) { // want "named return values are banned; return values must be unnamed" "named return values are banned; return values must be unnamed" "named return values are banned; return values must be unnamed"
	return 0, "", nil
}

func GoodUnnamed() (int, error) {
	return 0, nil
}

func GoodNoResults() {
}

func GoodSingleUnnamed() int {
	return 42
}
