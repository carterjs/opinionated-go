package constantsattop

const retryLimit = 3

// Retries reports the limit.
func Retries() int {
	return retryLimit
}

const timeoutSeconds = 30 // want "constants belong at the top of the file, above the functions that use them \\(2 blocks are below\\)"

// Timeout reports the timeout.
func Timeout() int {
	return timeoutSeconds
}

const backoffSeconds = 5

// Backoff reports the backoff.
func Backoff() int {
	return backoffSeconds
}
