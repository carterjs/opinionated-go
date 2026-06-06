package inlineable

import "fmt"

// One-liner used exactly twice - should inline
func helper1() { // want "should be inlined"
	fmt.Println("hi")
}

// Short function (3 statements) used once - should inline
func helper2() { // want "should be inlined"
	x := 1
	y := 2
	_ = x + y
}

func useHelper1() {
	helper1()
	helper1()
}

func useHelper2() {
	helper2()
}

// Longer function with loop - don't inline
func helper3() {
	for i := 0; i < 10; i++ {
		fmt.Println(i)
		fmt.Println(i * 2)
		fmt.Println(i * 3)
	}
}

// Never used - not our concern
func unused() {
	fmt.Println("never called")
}
