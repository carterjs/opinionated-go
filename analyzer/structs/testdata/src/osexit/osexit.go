package osexit

import "os"

func Process() error {
	os.Exit(1) // want "os.Exit only allowed in main"
	return nil
}

func fail() {
	os.Exit(2) // want "os.Exit only allowed in main"
}
