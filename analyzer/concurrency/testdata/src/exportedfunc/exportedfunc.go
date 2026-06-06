package exportedfunc

func ExportedWithFunc(fn func()) {} // want "prefer an interface with a method"

func unexportedWithFunc(fn func()) {}

func ExportedWithInterface(i interface{}) {}
