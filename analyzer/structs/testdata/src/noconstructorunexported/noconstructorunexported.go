package noconstructorunexported

// Bad: Has unexported fields
type Database struct { // want "struct has unexported fields but no constructor"
	host     string
	port     int
	user     string
	password string
}

// Good: All exported fields
type Logger struct {
	Level string
	Name  string
}

func NewLogger(name string) *Logger {
	return &Logger{
		Level: "info",
		Name:  name,
	}
}

// Good: All exported fields
type PublicData struct {
	Name  string
	Value int
}

// Bad: Has unexported fields
type Cache struct { // want "struct has unexported fields but no constructor"
	mu    interface{}
	items map[string]interface{}
	ttl   int
}

// Good: Has constructor for unexported fields
type Handler struct {
	logger interface{}
	store  interface{}
}

func NewHandler() *Handler {
	return &Handler{}
}

// Good: Empty struct
type Empty struct {
}

// Bad: Has unexported fields
type Mixed struct { // want "struct has unexported fields but no constructor"
	Name   string
	secret string
}
