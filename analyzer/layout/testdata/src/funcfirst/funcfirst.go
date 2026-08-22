package funcfirst

func helper() bool { // want "funcfirst.go must open with its type declarations; move helper below them"
	return true
}

// Thing is the subject of this package.
type Thing struct {
	name string
}
