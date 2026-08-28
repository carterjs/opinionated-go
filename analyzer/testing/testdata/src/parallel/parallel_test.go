package parallel

import "testing"

func TestParse(t *testing.T) {
	t.Parallel()
	t.Run("returns error when input is empty", func(t *testing.T) {
		t.Parallel()
	})
}

func TestFormat(t *testing.T) { // want "does not call t.Parallel"
	t.Run("returns the formatted document", func(t *testing.T) { // want "subtest does not call t.Parallel"
		_ = Format(nil)
	})
}
