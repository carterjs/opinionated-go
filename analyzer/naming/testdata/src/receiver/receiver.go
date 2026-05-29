package receiver

// Store is a data store.
type Store struct{}

// Get retrieves a value.
func (s *Store) Get() string { // want "receiver name .* is too short"
	return ""
}

// Service is a business logic service.
type Service struct{}

// Process processes a request.
func (svc *Service) Process() {}

// Short is a type with a short name.
type S struct{}

// Get on S is fine.
func (s *S) Get() {}
