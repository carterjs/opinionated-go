package nosleep_test

import (
	"testing"
	"time"
)

func TestWithSleep(t *testing.T) {
	time.Sleep(100 * time.Millisecond) // want "avoid time.Sleep in tests"
}

func TestWithWait(t *testing.T) {
	// Proper way: use test utilities or channels
	done := make(chan struct{})
	go func() {
		defer close(done)
		// do work
	}()
	<-done
}
