package anyinapi

// Bad: Using any in exported function
func ProcessData(input any) error { // want "avoid any/interface"
	return nil
}

// Bad: Using interface{} in exported function
func HandleValue(v interface{}) { // want "avoid any/interface"
}

// Good: Using concrete type
func ProcessString(input string) error {
	return nil
}

// Good: Using specific interface
type Reader interface {
	Read(p []byte) (n int, err error)
}

func ProcessReader(r Reader) error {
	return nil
}

// Good: Unexported function with any is fine
func processAny(input any) error {
	return nil
}

// Bad: Multiple any parameters
func MergeValues(a any, b any) any { // want "avoid any/interface" "avoid any/interface"
	return nil
}

// Good: Map with interface{} values is OK (not in parameter types)
var data = map[string]interface{}{
	"key": "value",
}
