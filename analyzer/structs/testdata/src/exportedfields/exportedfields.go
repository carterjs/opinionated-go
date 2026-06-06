package exportedfields

// Bad: Exported field
type Config struct {
	Name string // want "struct with methods should not have exported fields"
	port int
}

func (c *Config) GetPort() int {
	return c.port
}

// Good: No exported fields
type GoodConfig struct {
	name string
	port int
}

func (c *GoodConfig) Name() string {
	return c.name
}

// Bad: Exported field (even without methods)
type SimpleData struct {
	Value int // want "struct with methods should not have exported fields"
	name  string
}

// Bad: Multiple exported fields
type MultiField struct {
	ID   int    // want "struct with methods should not have exported fields"
	Name string // want "struct with methods should not have exported fields"
	value bool
}

// Good: All unexported fields
type PrivateData struct {
	id    int
	name  string
	value bool
}
