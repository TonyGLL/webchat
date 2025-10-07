package config

import (
	"fmt"
	"os"

	"github.com/joho/godotenv"
)

// LoadConfig loads the environment variables from the given file path.
func LoadConfig(path string) error {
	// If no path is provided, it will look for a .env file in the current directory
	err := godotenv.Load(path)
	if err != nil {
		return fmt.Errorf("error loading .env file from path %s: %w", path, err)
	}
	return nil
}

// Getenv retrieves the value of the environment variable named by the key.
// It returns the value, which will be empty if the variable is not present.
func Getenv(key string) string {
	return os.Getenv(key)
}