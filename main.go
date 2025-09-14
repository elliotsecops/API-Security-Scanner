package main

import (
	"log"

	"api-security-scanner/config"
	"api-security-scanner/scanner"
)

func main() {
	// Load configuration from the YAML file
	cfg, err := config.Load("config.yaml")
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	// Debug logging
	log.Printf("Loaded configuration: %+v", cfg)
	for _, endpoint := range cfg.APIEndpoints {
		log.Printf("Endpoint: %s, Method: %s", endpoint.URL, endpoint.Method)
	}

	// Run the security tests
	results := scanner.RunTests(cfg)

	// Generate detailed report
	scanner.GenerateDetailedReport(results)
}
