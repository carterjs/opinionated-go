package casenaming

import "testing"

func TestParse(t *testing.T) {
	tests := []struct {
		name  string
		input string
	}{
		{name: "returns the parsed document", input: "a"},
		{name: "returns error when input is empty", input: ""},
		{name: "Returns error when input is short", input: "b"},     // want "starts uppercase"
		{name: "returns error when input is long.", input: "c"},     // want "ends with a period"
		{name: "returns_error_when_input_is_null", input: "d"},      // want "uses underscores"
		{name: "returnsErrorWhenInputIsBad", input: "e"},            // want "is a single word"
		{name: "success", input: "f"},                               // want "is a single word"
		{name: "happy path", input: "g"},                            // want "names an outcome class"
		{name: "should return error when input is odd", input: "h"}, // want "opens with"
		{name: "returns the parsed document", input: "i"},           // want "already used in this table"
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			_ = test.input
		})
	}
}
