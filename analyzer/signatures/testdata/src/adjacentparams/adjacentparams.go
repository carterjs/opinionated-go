package adjacentparams

type Options struct {
	Overwrite bool
}

func Move(source, destination string) error { return nil } // want "adjacent parameters of type string are swappable"

func Copy(source string, destination string) error { return nil } // want "adjacent parameters of type string are swappable"

func MoveWithOptions(source string, options Options) error { return nil }

func Rename(path string, depth int) error { return nil }

func unexportedMove(source, destination string) error { return nil }
