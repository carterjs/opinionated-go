package constantsattop

const retryLimit = 3

// Retries reports the limit.
func Retries() int {
	return retryLimit
}

const timeoutSeconds = 30 // want "constants belong at the top of the file"

// Timeout reports the timeout.
func Timeout() int {
	return timeoutSeconds
}
