package initialism

type Id string // want "initialism .* should be .*"

type ID string

func FetchUrl() string { // want "initialism .* should be .*"
	return ""
}

func FetchURL() string {
	return ""
}
