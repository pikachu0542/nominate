package config

import (
	"fmt"
	"os"
	"strconv"
)

// Custom struct to store values of environment variables
type Config struct {
	DBString string
	Port     string
	Auth     AuthConfig
	DevFlags DevConfig
}

type AuthConfig struct {
	ClientID     string
	ClientSecret string
	JWTSecret    string
	State        string
	Host         string
}

type DevConfig struct {
	DisableActiveFilters bool
	ForceEboard          bool
	ForceChair           bool
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

	auth, err := loadAuthConfig()
	if err != nil {
		return nil, fmt.Errorf("Error loading auth config: %w", err)
	}

	dev, err := loadDevConfig()
	if err != nil {
		return nil, fmt.Errorf("Error loading dev config: %w", err)
	}

	// Return a tuple containing the loaded env vars in a Config object, and
	// nil for the error part since if this is reached, no errors were encountered
	return &Config{
		DBString: dbString,
		Port:     port,
		Auth:     *auth,
		DevFlags: *dev,
	}, nil
}

// Loads the environment variables needed for CSH auth
func loadAuthConfig() (*AuthConfig, error) {
	clientID := os.Getenv("NOMINATE_OIDC_ID")
	if clientID == "" {
		return nil, fmt.Errorf("NOMINATE_OIDC_ID is not set")
	}

	clientSecret := os.Getenv("NOMINATE_OIDC_SECRET")
	if clientSecret == "" {
		return nil, fmt.Errorf("NOMINATE_OIDC_SECRET is not set")
	}

	jwtSecret := os.Getenv("NOMINATE_JWT_TOKEN")
	if jwtSecret == "" {
		return nil, fmt.Errorf("NOMINATE_JWT_TOKEN is not set")
	}

	state := os.Getenv("NOMINATE_STATE")
	if state == "" {
		return nil, fmt.Errorf("NOMINATE_STATE is not set")
	}

	host := os.Getenv("NOMINATE_HOST")
	if host == "" {
		return nil, fmt.Errorf("NOMINATE_HOST is not set")
	}

	return &AuthConfig{
		ClientID:     clientID,
		ClientSecret: clientSecret,
		JWTSecret:    jwtSecret,
		State:        state,
		Host:         host,
	}, nil
}

// Loads environment variables used to set dev overrides
func loadDevConfig() (*DevConfig, error) {
	disableActiveFilters, err := parseBoolEnv("DEV_DISABLE_ACTIVE_FILTERS")
	if err != nil {
		return nil, err
	}

	forceEboard, err := parseBoolEnv("DEV_FORCE_IS_EBOARD")
	if err != nil {
		return nil, err
	}

	forceChair, err := parseBoolEnv("DEV_FORCE_IS_CHAIR")
	if err != nil {
		return nil, err
	}

	return &DevConfig{
		DisableActiveFilters: disableActiveFilters,
		ForceEboard:          forceEboard,
		ForceChair:           forceChair,
	}, nil
}

// Reads a string environment variable and tries top parse it to a boolean value
func parseBoolEnv(key string) (bool, error) {
	raw := os.Getenv(key)

	// Default to false if not set
	if raw == "" {
		return false, nil
	}

	// Try to parse to a bool, error if unable
	val, err := strconv.ParseBool(raw)
	if err != nil {
		return false, fmt.Errorf("Invalid boolean value %s: %w", key, err)
	}

	return val, nil
}
