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

func BothSpellings(id string) (*User, bool) { return nil, false } // want "a record spells absence with nil"

func OkAndError(id string) (string, bool, error) { return "", false, nil } // want "a signature returns ok or error, never both"

func CollectionOk(id string) ([]User, bool) { return nil, false } // want "a collection has no absent state"

func MapOk(id string) (map[string]User, bool) { return nil, false } // want "a collection has no absent state"

func PointerToSlice(id string) (*[]User, error) { return nil, nil } // want "a nil slice is an empty slice"
