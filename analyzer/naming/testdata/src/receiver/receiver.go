package receiver

// Store is a data store.
type Store struct{} // want "receiver name .s. is too short for Store \\(3 methods\\)"

// Get retrieves a value.
func (s *Store) Get() string {
	return ""
}

// Put stores a value.
func (s *Store) Put(value string) {}

// Delete removes a value.
func (s *Store) Delete() {}

// Service is a business logic service.
type Service struct{}

// Process processes a request.
func (svc *Service) Process() {}

// Short is a type with a short name.
type S struct{}

// Get on S is fine.
func (s *S) Get() {}

// Sealed-interface marker methods (unnamed receivers) should not be flagged or crash.
type InsertPatch struct{}

func (InsertPatch) patch() {}
