package config

import (
	"api-security-scanner/scanner"
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

// Load loads the configuration from a YAML file
func Load(filename string) (*scanner.Config, error) {
	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, err
	}

	var config scanner.Config
	err = yaml.Unmarshal(data, &config)
	if err != nil {
		return nil, err
	}

	// Validate the configuration
	if err := Validate(&config); err != nil {
		return nil, fmt.Errorf("configuration validation failed: %w", err)
	}

	return &config, nil
}

// Validate checks if the configuration is valid
func Validate(config *scanner.Config) error {
	// Check if we have at least one API endpoint
	if len(config.APIEndpoints) == 0 {
		return fmt.Errorf("at least one API endpoint is required")
	}

	// Validate each endpoint
	for i, endpoint := range config.APIEndpoints {
		if endpoint.URL == "" {
			return fmt.Errorf("endpoint %d: URL is required", i)
		}
		if endpoint.Method == "" {
			return fmt.Errorf("endpoint %d: HTTP method is required", i)
		}
		// Validate HTTP method is one of the standard methods
		validMethods := map[string]bool{
			"GET": true, "POST": true, "PUT": true, "DELETE": true, 
			"PATCH": true, "HEAD": true, "OPTIONS": true, "TRACE": true,
		}
		if !validMethods[endpoint.Method] {
			return fmt.Errorf("endpoint %d: invalid HTTP method '%s'", i, endpoint.Method)
		}
	}

	// Check if we have at least one injection payload
	if len(config.InjectionPayloads) == 0 {
		return fmt.Errorf("at least one injection payload is required")
	}

	// If auth is provided, both username and password are required
	if (config.Auth.Username != "" && config.Auth.Password == "") || 
	   (config.Auth.Username == "" && config.Auth.Password != "") {
		return fmt.Errorf("both username and password are required for authentication, or neither")
	}

	return nil
}

// TestConfigReachability tests if the configured endpoints are reachable
func TestConfigReachability(config *scanner.Config) []string {
	var issues []string
	// In a real implementation, we would test endpoint reachability here
	// For now, we'll just return an empty slice
	return issues
}
