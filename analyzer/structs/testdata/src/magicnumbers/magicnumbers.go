package magicnumbers

import "strconv"

const (
	maxUploadBytes = 10 << 20
	retryLimit     = 3
)

// MaxBytes names its limit.
func MaxBytes() int {
	return maxUploadBytes
}

func maxBytes() int {
	return 10 << 20 // want "magic number 10" "magic number 20"
}

func retries() int {
	return 3 // want "magic number 3"
}

// The same number repeated in one function is reported once, at its first use.
func window(offset int) int {
	start := offset + 7 // want "magic number 7"
	end := start + 7
	return end + 7
}

func empty() int {
	return 0
}

func step(count int) int {
	return count + 1
}

func ratio() float64 {
	return 0.75 // want "magic number 0.75"
}

// parseDecimal's 10 is a base, not a value anyone chose; not a magic number.
func parseDecimal(s string) (int64, error) {
	return strconv.ParseInt(s, 10, 64)
}

// formatHex's 16 is a base too.
func formatHex(n int64) string {
	return strconv.FormatInt(n, 16)
}
