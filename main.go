package main

import (
	"flag"
	"os"

	"api-security-scanner/config"
	"api-security-scanner/logging"
	"api-security-scanner/scanner"
)

func main() {
	// Command line flags
	configFile := flag.String("config", "config.yaml", "Path to configuration file")
	outputFormat := flag.String("output", "text", "Output format (text, json, html, csv, xml)")
	validateOnly := flag.Bool("validate", false, "Validate configuration only, don't run tests")
	logLevel := flag.String("log-level", "info", "Log level (debug, info, warn, error)")
	logFormat := flag.String("log-format", "text", "Log format (text, json)")
	flag.Parse()

	// Set up logging
	setupLogging(*logLevel, *logFormat)

	// Load configuration from the YAML file
	cfg, err := config.Load(*configFile)
	if err != nil {
		logging.Error("Failed to load configuration", map[string]interface{}{
			"error": err.Error(),
		})
		os.Exit(1)
	}

	// Apply default values for rate limiting if not specified
	requestsPerSecond := cfg.RateLimiting.RequestsPerSecond
	if requestsPerSecond <= 0 {
		requestsPerSecond = 10
	}
	
	maxConcurrentRequests := cfg.RateLimiting.MaxConcurrentRequests
	if maxConcurrentRequests <= 0 {
		maxConcurrentRequests = 5
	}

	// Log configuration details
	logging.Info("Loaded configuration", map[string]interface{}{
		"endpoints_count": len(cfg.APIEndpoints),
		"has_auth":        cfg.Auth.Username != "" && cfg.Auth.Password != "",
		"payloads_count":  len(cfg.InjectionPayloads),
		"rate_limiting": map[string]interface{}{
			"configured": map[string]interface{}{
				"requests_per_second":     cfg.RateLimiting.RequestsPerSecond,
				"max_concurrent_requests": cfg.RateLimiting.MaxConcurrentRequests,
			},
			"effective": map[string]interface{}{
				"requests_per_second":     requestsPerSecond,
				"max_concurrent_requests": maxConcurrentRequests,
			},
		},
	})

	for i, endpoint := range cfg.APIEndpoints {
		logging.Debug("Endpoint details", map[string]interface{}{
			"index":  i,
			"url":    endpoint.URL,
			"method": endpoint.Method,
		})
	}

	// Test endpoint reachability
	issues := config.TestConfigReachability(cfg)
	if len(issues) > 0 {
		logging.Warn("Configuration reachability issues detected", map[string]interface{}{
			"issues": issues,
		})
	}

	// If validate-only flag is set, exit after validation
	if *validateOnly {
		logging.Info("Configuration is valid", nil)
		return
	}

	// Run the security tests
	logging.Info("Starting security tests", map[string]interface{}{
		"endpoints_count": len(cfg.APIEndpoints),
	})
	results := scanner.RunTests(cfg)

	// Generate report based on selected format
	switch *outputFormat {
	case "json":
		scanner.GenerateJSONReport(results)
	case "html":
		scanner.GenerateHTMLReport(results)
	case "csv":
		scanner.GenerateCSVReport(results)
	case "xml":
		scanner.GenerateXMLReport(results)
	default:
		scanner.GenerateDetailedReport(results)
	}
}

func setupLogging(levelStr, format string) {
	// Convert level string to LogLevel
	var level logging.LogLevel
	switch levelStr {
	case "debug":
		level = logging.DEBUG
	case "info":
		level = logging.INFO
	case "warn":
		level = logging.WARN
	case "error":
		level = logging.ERROR
	default:
		level = logging.INFO
	}

	// Set global logger configuration
	logging.SetGlobalLevel(level)
	logging.SetGlobalFormat(format)
}
