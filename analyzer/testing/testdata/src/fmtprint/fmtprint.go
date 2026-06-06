package fmtprint_test

import (
	"fmt"
	"testing"
)

func TestSomething(t *testing.T) {
	fmt.Print("hello") // want "use t.Log or t.Logf"
	fmt.Printf("hello %s\n", "world") // want "use t.Log or t.Logf"
	fmt.Println("done") // want "use t.Log or t.Logf"

	t.Log("hello")
	t.Logf("hello %s", "world")
}
