package doclength

// Parse reads the document and returns it.
func Parse(data []byte) error { return nil }

// Format renders the document to a string. The returned string is freshly
// allocated, so the caller owns it and may hold it past the next call. Blocks
// only on the writer it was handed at construction. This last sentence exists
// purely to carry the comment past the ceiling, which it now does.
func Format(data []byte) string { return "" } // want "doc comment on Format runs"

// Store holds documents.
type Store struct{}
