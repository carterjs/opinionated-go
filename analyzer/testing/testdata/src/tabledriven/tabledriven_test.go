package tabledriven

import "testing"

func TestParse(t *testing.T) {
	tests := []struct{ name string }{{name: "returns error when input is empty"}}
	for _, test := range tests {
		_ = test.name
	}
}

func TestFormat(t *testing.T) { // want "is not table-driven"
	_ = Format(nil)
}
