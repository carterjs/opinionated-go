package nakederrorreturn

import "fmt"

// FakeThing is a test double whose exported methods are not this package's
// public API; naked returns here are exempt.
type FakeThing struct{}

func (f *FakeThing) Do() error {
	err := fmt.Errorf("boom")
	return err
}
