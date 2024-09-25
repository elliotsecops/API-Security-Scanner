package main

import (
	"io/ioutil"
	"log"

	"gopkg.in/yaml.v2"
)

func main() {
	// Load configuration from the YAML file
	config, err := loadConfig("config.yaml")
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	// Debug logging
	log.Printf("Loaded configuration: %+v", config)
	for _, endpoint := range config.APIEndpoints {
		log.Printf("Endpoint: %s, Method: %s", endpoint.URL, endpoint.Method)
	}

	// Run the security tests
	results := runTests(config)

	// Generate detailed report
	generateDetailedReport(results)
}

// loadConfig loads the configuration from a YAML file
func loadConfig(filename string) (*Config, error) {
	data, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, err
	}

	var config Config
	err = yaml.Unmarshal(data, &config)
	if err != nil {
		return nil, err
	}

	return &config, nil
}
