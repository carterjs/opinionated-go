package inlinelength

// Parse reads the document.
func Parse(data []byte) error {
	// A single line of why is fine.
	count := len(data)

	// This explanation runs past one line, so the why has outgrown what an // want "inline comment runs"
	// inline comment is for and belongs somewhere a reader can find it.
	_ = count

	return nil
}
