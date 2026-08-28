package main

import "os"

func main() { // want "main has 11 statements"
	a := 1
	b := 2
	c := 3
	d := 4
	e := 5
	f := 6
	g := 7
	h := 8
	i := 9
	j := 10
	_ = a + b + c + d + e + f + g + h + i + j
}

func other() {
	_ = os.Args
}
