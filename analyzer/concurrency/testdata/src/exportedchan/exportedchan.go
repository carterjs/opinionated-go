package exportedchan

func ExportedWithChannel(ch chan struct{}) {} // want "prefer wrapping coordination primitives"

func unexportedWithChannel(ch chan struct{}) {}

func ExportedWithoutChannel() {}
