package exportedfields

// Bad: one exported field, reported once against the type
type Config struct { // want "struct .Config. has methods and should not have exported fields \\(Name\\)"
	Name string
	port int
}

func (config *Config) Port() int {
	return config.port
}

// Good: no exported fields
type GoodConfig struct {
	name string
	port int
}

func (config *GoodConfig) Name() string {
	return config.name
}

// Good: a plain record with no methods is exported fields and nothing else
type SimpleData struct {
	Value int
	name  string
}

// Bad: three exported fields, still one finding for the struct
type MultiField struct { // want "struct .MultiField. has methods and should not have exported fields \\(ID, Name, Size\\)"
	ID         int
	Name, Size string
	value      bool
}

// Label gives MultiField behaviour.
func (field *MultiField) Label() string {
	return field.Name
}

// Good: all unexported fields
type PrivateData struct {
	id    int
	name  string
	value bool
}

// Good: an anonymous struct has no methods to route access through
var anonymous = struct {
	Field string
}{Field: "value"}
