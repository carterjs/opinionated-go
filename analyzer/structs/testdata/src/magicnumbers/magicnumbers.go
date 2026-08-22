package magicnumbers

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

func empty() int {
	return 0
}

func step(count int) int {
	return count + 1
}

func ratio() float64 {
	return 0.75 // want "magic number 0.75"
}
