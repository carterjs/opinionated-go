package testnameloop

import "testing"

type testCase struct {
	name string
	skip bool
	// ... other fields
}

func TestBad_NameInIf(t *testing.T) {
	tests := []testCase{
		{name: "case1"},
		{name: "case2"},
	}
	for _, test := range tests {
		if test.name == "case1" { // want "avoid checking test.name in conditionals"
			t.Skip()
		}
		t.Run(test.name, func(t *testing.T) {
			// test body
		})
	}
}

func TestBad_NameInSwitch(t *testing.T) {
	tests := []testCase{
		{name: "case1"},
		{name: "case2"},
	}
	for _, test := range tests {
		switch test.name { // want "avoid switching on test.name"
		case "case1":
			t.Skip()
		}
		t.Run(test.name, func(t *testing.T) {
			// test body
		})
	}
}

func TestGood_SkipFieldHandling(t *testing.T) {
	tests := []testCase{
		{name: "case1", skip: true},
		{name: "case2", skip: false},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if test.skip {
				t.Skip()
			}
			// test body
		})
	}
}

func TestGood_ErrorHandling(t *testing.T) {
	tests := []testCase{
		{name: "case1"},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			// Error handling inside test is fine
			if err := someFunc(); err != nil {
				t.Fatalf("got error: %v", err)
			}
		})
	}
}

func someFunc() error {
	return nil
}
