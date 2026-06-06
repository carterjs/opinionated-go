package singlelettervars

import "fmt"

func TestLoopVar() {
	for i := 0; i < 10; i++ { // i is OK in for loop
		fmt.Println(i)
	}
}

func TestParam(x int) { // want "parameter .* is too short"
	fmt.Println(x)
}

func TestLocalVar() {
	y := 5
	fmt.Println(y)
}

// Variable used across many lines - should flag (spans 7 lines)
func TestLongSpan() {
	x := 10 // want "variable .* is too short and spans"
	fmt.Println(x)
	fmt.Println(x)
	fmt.Println(x)
	fmt.Println(x)
	fmt.Println(x)
	fmt.Println(x)
}

func TestGoodNames() {
	value := 10
	fmt.Println(value)

	for idx := 0; idx < 5; idx++ {
		fmt.Println(idx)
	}
}
