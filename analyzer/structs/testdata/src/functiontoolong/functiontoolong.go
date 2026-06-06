package functiontoolong

// Good: Function under 60 lines
func ShortFunction(x int) int {
	a := x + 1
	b := a + 2
	c := b + 3
	return c
}

// Bad: Function over 60 lines (more than 60 lines)
func TooLongFunction(x int) int { // want "function too long"
	result := x + 1
	result += 2
	result += 3
	result += 4
	result += 5
	result += 6
	result += 7
	result += 8
	result += 9
	result += 10
	result += 11
	result += 12
	result += 13
	result += 14
	result += 15
	result += 16
	result += 17
	result += 18
	result += 19
	result += 20
	result += 21
	result += 22
	result += 23
	result += 24
	result += 25
	result += 26
	result += 27
	result += 28
	result += 29
	result += 30
	result += 31
	result += 32
	result += 33
	result += 34
	result += 35
	result += 36
	result += 37
	result += 38
	result += 39
	result += 40
	result += 41
	result += 42
	result += 43
	result += 44
	result += 45
	result += 46
	result += 47
	result += 48
	result += 49
	result += 50
	result += 51
	result += 52
	result += 53
	result += 54
	result += 55
	result += 56
	result += 57
	result += 58
	result += 59
	return result
}

// Good: Function under 60 lines
func AnotherShortFunction() int {
	x := 1
	y := 2
	z := 3
	return x + y + z
}
