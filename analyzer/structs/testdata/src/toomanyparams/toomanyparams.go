package toomanyparams

// Good: 4 parameters
func GoodFunction(a string, b int, c bool, d float64) {}

// Good: uses config struct
type Config struct {
	A string
	B int
	C bool
	D float64
	E string
}

func ConfigFunction(a string, cfg Config) {}

// Bad: 5 parameters
func BadFunction(a string, b int, c bool, d float64, e string) {} // want "has 5 parameters"

// Bad: 6 parameters
func TooManyFunction(a string, b int, c bool, d float64, e string, f string) {} // want "has 6 parameters"

// Good: no parameters
func NoParams() {}

// Good: 1 parameter
func OneParam(a string) {}
