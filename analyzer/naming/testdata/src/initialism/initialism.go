package initialism

type Id string // want "initialism .* should be .*"

type ID string

func FetchUrl() string { // want "initialism .* should be .*"
	return ""
}

func FetchURL() string {
	return ""
}

// UserId embeds a miscased initialism as a trailing camelCase word.
type UserId string // want "initialism .* should be .*"

// UserID is correct.
type UserID string

// Full words that merely begin with an initialism's letters must not be
// flagged: "Identifier", "Identity", "Idle", and "Idempotent" all start with
// "Id" but are not the initialism.
func GetByIdentifier() string {
	return ""
}

type Identity string

func ExpireIdleAgents() {}

func AddIsIdempotent() {}
