package getenvoutsidemain

import "os"

// Bad: os.Getenv in regular package
func GetDatabaseURL() string {
	return os.Getenv("DATABASE_URL") // want "os.Getenv only allowed"
}

// Bad: os.LookupEnv in regular package
func GetAPIKey() (string, bool) {
	return os.LookupEnv("API_KEY") // want "os.Getenv only allowed"
}

// Good: Not calling getenv
func GetStaticValue() string {
	return "static"
}

// Good: Using a config struct instead
type Config struct {
	databaseURL string
}

func (c *Config) DatabaseURL() string {
	return c.databaseURL
}

// Bad: Getenv inside a helper function
func loadEnv(key string) string {
	return os.Getenv(key) // want "os.Getenv only allowed"
}
