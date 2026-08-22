package singlelettervars

import (
	"fmt"
	"testing"
)

func TestLoopVar() {
	for i := 0; i < 10; i++ { // i is OK in for loop
		fmt.Println(i)
	}
}

func TestParam(x int) { // OK: x is used within 5 lines
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
	fmt.Println(x) // line 1
	fmt.Println(x) // line 2
	fmt.Println(x) // line 3
	fmt.Println(x) // line 4
	fmt.Println(x) // line 5
	fmt.Println(x) // line 6 - x used beyond 5 lines
}

func TestUnderscoreParam(_ int) { // OK: underscore is allowed
	fmt.Println("nothing")
}

// n is the conventional name for a count and is allowed at any span.
func TestCountingVariable(values []string) int {
	n := 0
	for _, value := range values {
		if value == "" {
			continue
		}
		n++
	}
	fmt.Println(n)
	return n
}

// A name read exactly once stays readable however far the reference sits.
func TestSingleUseAcrossManyLines() {
	x := 10
	fmt.Println("one")
	fmt.Println("two")
	fmt.Println("three")
	fmt.Println("four")
	fmt.Println("five")
	fmt.Println("six")
	fmt.Println(x) // OK: x is used once
}

// A testing parameter keeps its idiomatic name outside _test.go files too.
func TestHelperOutsideTestFile(t *testing.T, values []string) {
	t.Helper()
	for _, value := range values {
		if value == "" {
			t.Error("empty value")
		}
	}
	t.Log("checked")
	t.Log("done")
}
