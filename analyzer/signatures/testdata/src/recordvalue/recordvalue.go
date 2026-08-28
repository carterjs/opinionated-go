package recordvalue

import "time"

type User struct {
	Name string
}

type Cache struct{}

// Seen returns when the key was last read; time.Time has value semantics.
func (cache *Cache) Seen(key string) (time.Time, bool) { return time.Time{}, false }

// TTL returns how long the key lives.
func (cache *Cache) TTL(key string) (time.Duration, bool) { return 0, false }

// Token returns the raw token.
func (cache *Cache) Token(key string) (string, bool) { return "", false }

// User returns the user, spelled correctly.
func (cache *Cache) User(key string) (*User, error) { return nil, nil }

func RecordByValue(id string) (User, bool) { return User{}, false } // want "recordvalue.User is a record"
