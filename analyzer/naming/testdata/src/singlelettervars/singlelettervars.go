package singlelettervars

import "fmt"

func TestLoopVar() {
	for i := 0; i < 10; i++ { // i is OK in for loop
		fmt.Println(i)
	}
}

func TestParam(x int) { // OK: function is 3 lines (<=5), so single letter is allowed
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

// Small scope functions should allow single-letter params
func TestSmallScopeClosure() {
	data := []int{3, 1, 2}
	// This closure is <= 5 lines, so i, j are OK
	sort := func(i, j int) bool { return data[i] < data[j] } // OK: small scope
	_ = sort
}

func TestSmallScopeFunction(x int) bool { // OK: function is <= 5 lines
	return x > 0
}

func TestLargeScopeParam(x int) { // want "parameter .* is too short"
	fmt.Println(x)
	fmt.Println(x)
	fmt.Println(x)
	fmt.Println(x)
	fmt.Println(x)
	fmt.Println(x)
}

func TestUnderscoreParam(_ int) { // OK: underscore is allowed
	fmt.Println("nothing")
}
