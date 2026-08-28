package paramcount

func SixGrouped(a, b, c, d, e, f string) {} // want "function has 6 parameters"

func FiveMixed(a string, b int, c bool, d, e string) {} // want "function has 5 parameters"

func FourIsFine(a string, b int, c bool, d string) {}

func sixUnexported(a, b, c, d, e, f string) {}
