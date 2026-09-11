package config

import (
	"fmt"
	"os"
)

// Custom struct to store values of environment variables
type Config struct {
	DBString string
	Port     string
}

// Reads environment variables from a .env file and store them
// in a custom struct that can be used by other packages
func Load() (*Config, error) {
	// Load the DB connection string, error if it isnt set
	dbString := os.Getenv("GOOSE_DBSTRING")
	if dbString == "" {
		return nil, fmt.Errorf("GOOSE_DBSTRING is not set")
	}

	// Read the port from the .env, or default to 8080 if not set
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	// Return a tuple containing the loaded env vars in a Config object, and
	// nil for the error part since if this is reached, no errors were encountered
	return &Config{
		DBString: dbString,
		Port:     port,
	}, nil
}
