package initfunc

var GlobalState string

func init() { // want "init functions should be avoided"
	GlobalState = "initialized"
}

func Initialize() {
	GlobalState = "initialized"
}
