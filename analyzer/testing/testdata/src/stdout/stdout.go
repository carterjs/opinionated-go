package stdout_test

import (
	"io"
	"os"
	"testing"
)

func TestSomething(t *testing.T) {
	var w io.Writer
	w = os.Stdout // want "use t.Output()"
	w = os.Stderr // want "use t.Output()"
	_ = w

	out := t.Output()
	_ = out
}
