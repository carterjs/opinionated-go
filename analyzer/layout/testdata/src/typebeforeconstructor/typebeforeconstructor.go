package typebeforeconstructor

const limit = 1

// Store holds things.
type Store struct {
	items []string
}

// NewStore builds a Store.
func NewStore() *Store {
	return &Store{}
}

// Cache holds things too.
type Cache struct {
	items []string
}

func helper() bool {
	return true
}

// NewCache is separated from Cache by helper.
func NewCache() *Cache { // want "constructor NewCache must be declared immediately after type Cache"
	return &Cache{}
}

// NewRegistry lives away from the type it builds.
func NewRegistry() *Registry { // want "constructor NewRegistry belongs beside type Registry in registry.go"
	return &Registry{}
}
