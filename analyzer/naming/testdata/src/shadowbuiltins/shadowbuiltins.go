package shadowbuiltins

// Buffer is a thing with methods.
type Buffer struct {
	items []int
}

// Len is a method, so it cannot shadow the built-in.
func (buffer *Buffer) Len() int {
	return len(buffer.items)
}

func (buffer *Buffer) copy() []int {
	return append([]int(nil), buffer.items...)
}

func (buffer *Buffer) max() int {
	return len(buffer.items)
}

func len(values []int) int { // want "function .len. shadows Go built-in"
	return 0
}

func copy(values []int) int { // want "function .copy. shadows Go built-in"
	return 0
}

func size(values []int) int {
	return 0
}
