package exportedcomment

import _ "embed"

// LegacyEmbedFile uses legacy syntax.
// go:embed testdata.txt // want "use //go:embed"
var LegacyEmbedFile string

// ModernEmbedFile uses the correct syntax.
//go:embed testdata.txt
var ModernEmbedFile string
