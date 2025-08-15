package config

import (
	"bufio"
	"os"
	"strings"
)

// LoadEnvFile loads environment variables from a .env file
func LoadEnvFile(filename string) error {
	file, err := os.Open(filename)
	if err != nil {
		// If .env file doesn't exist, it's not an error
		return nil
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())

		// Skip empty lines and comments
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		// Parse key=value pairs
		if strings.Contains(line, "=") {
			parts := strings.SplitN(line, "=", 2)
			if len(parts) == 2 {
				key := strings.TrimSpace(parts[0])
				value := strings.TrimSpace(parts[1])

				// Remove quotes if present
				if len(value) >= 2 && (value[0] == '"' && value[len(value)-1] == '"') {
					value = value[1 : len(value)-1]
				}

				// Set environment variable if not already set
				if os.Getenv(key) == "" {
					os.Setenv(key, value)
				}
			}
		}
	}

	return scanner.Err()
}

// LoadDefaultEnvFile loads the default .env file in the current directory
func LoadDefaultEnvFile() error {
	return LoadEnvFile(".env")
}
