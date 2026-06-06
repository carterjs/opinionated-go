package singlelettervars

import "fmt"

func TestLoopVar() {
	for i := 0; i < 10; i++ { // i is OK in for loop
		fmt.Println(i)
	}
}

// Short function - x is OK because function is <= 5 lines
func TestParam(x int) {
	fmt.Println(x)
}

// Longer function - parameters should be flagged (>7 lines)
func LongFunction(x int, y int) { // want "parameter .* is too short" "parameter .* is too short"
	fmt.Println(x)
	fmt.Println(y)
	fmt.Println(x + y)
	fmt.Println(x * y)
	fmt.Println(x - y)
	fmt.Println(x)
	fmt.Println(y)
}

func TestLocalVar() {
	y := 5 // want "variable .* is too short"
	fmt.Println(y)
}

func TestGoodNames() {
	value := 10
	fmt.Println(value)

	for idx := 0; idx < 5; idx++ {
		fmt.Println(idx)
	}
}
