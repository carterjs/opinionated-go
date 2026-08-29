package absence

import "time"

type User struct {
	Name string
}

type Store struct{}

// User returns the user with the given ID.
func (store *Store) User(id string) (*User, error) { return nil, nil }

// Users returns every user.
func (store *Store) Users() ([]User, error) { return nil, nil }

// TTL returns how long the key lives.
func (store *Store) TTL(key string) (time.Duration, bool) { return 0, false }

// Seen returns when the key was last read.
func (store *Store) Seen(key string) (time.Time, bool) { return time.Time{}, false }

// MatchString reports whether the pattern matches, which is the answer, not an ok.
func MatchString(pattern, value string) (bool, error) { return false, nil }

func BothSpellings(id string) (*User, bool) { return nil, false } // want "is a pointer alongside an ok"

// OkAndError is no longer flagged: whether ok and error may coexist is a
// judgment call this analyzer does not have a well-defined rule for.
func OkAndError(id string) (string, bool, error) { return "", false, nil }

// CollectionOk and MapOk are no longer flagged: a nil slice or map already
// reads as absent to every caller, so pairing one with an ok is not the kind
// of ambiguity this analyzer is after.
func CollectionOk(id string) ([]User, bool) { return nil, false }

func MapOk(id string) (map[string]User, bool) { return nil, false }

func PointerToSlice(id string) (*[]User, error) { return nil, nil } // want "a nil slice is an empty slice"

func PrimitivePointer(id string) (*string, error) { return nil, nil } // want "is a pointer to a primitive"
