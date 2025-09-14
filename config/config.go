package config

import (
	"api-security-scanner/scanner"
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

	return &config, nil
}
