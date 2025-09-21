package main

import (
	"flag"
	"os"

	"api-security-scanner/config"
	"api-security-scanner/logging"
	"api-security-scanner/scanner"
	"api-security-scanner/openapi"
	"api-security-scanner/discovery"
	"api-security-scanner/history"
	"api-security-scanner/types"
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

	// Handle OpenAPI specification integration
	var openAPIIntegration *openapi.OpenAPIIntegration
	if cfg.OpenAPISpec != "" {
		logging.Info("Loading OpenAPI specification", map[string]interface{}{
			"spec_path": cfg.OpenAPISpec,
		})

		var err error
		openAPIIntegration, err = openapi.NewOpenAPIIntegration(cfg.OpenAPISpec)
		if err != nil {
			logging.Error("Failed to load OpenAPI specification", map[string]interface{}{
				"error": err.Error(),
			})
			os.Exit(1)
		}

		// Generate endpoints from OpenAPI spec if no endpoints are configured
		if len(cfg.APIEndpoints) == 0 {
			logging.Info("No endpoints configured, generating from OpenAPI spec", nil)
			cfg.APIEndpoints = openAPIIntegration.GenerateEndpointsFromSpec()
		}

		// Log OpenAPI spec info
		specInfo := openAPIIntegration.GetSpecInfo()
		logging.Info("OpenAPI specification loaded", specInfo)
	}

	// Handle API discovery if enabled
	if cfg.APIDiscovery.Enabled && len(cfg.APIEndpoints) > 0 {
		logging.Info("Starting API discovery", map[string]interface{}{
			"max_depth": cfg.APIDiscovery.MaxDepth,
			"follow_links": cfg.APIDiscovery.FollowLinks,
		})

		apiDiscovery := discovery.NewAPIDiscovery(cfg.APIDiscovery)
		var discoveredEndpoints []types.APIEndpoint

		// Discover endpoints from each configured base URL
		for _, endpoint := range cfg.APIEndpoints {
			// Use the endpoint URL as a base for discovery
			baseURL := endpoint.URL
			if endpoints, err := apiDiscovery.DiscoverEndpoints(baseURL); err == nil {
				discoveredEndpoints = append(discoveredEndpoints, endpoints...)
			}
		}

		// Add discovered endpoints to the configuration
		if len(discoveredEndpoints) > 0 {
			logging.Info("Discovered additional endpoints", map[string]interface{}{
				"discovered_count": len(discoveredEndpoints),
			})
			cfg.APIEndpoints = append(cfg.APIEndpoints, discoveredEndpoints...)
		}
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

	// Initialize history manager if historical data is enabled
	var historyManager *history.HistoryManager
	if cfg.HistoricalData.Enabled {
		var err error
		historyManager, err = history.NewHistoryManager(cfg.HistoricalData)
		if err != nil {
			logging.Error("Failed to initialize history manager", map[string]interface{}{
				"error": err.Error(),
			})
			os.Exit(1)
		}
		logging.Info("History manager initialized", map[string]interface{}{
			"storage_path": cfg.HistoricalData.StoragePath,
			"retention_days": cfg.HistoricalData.RetentionDays,
		})
	}

	// Run the security tests
	logging.Info("Starting security tests", map[string]interface{}{
		"endpoints_count": len(cfg.APIEndpoints),
	})
	results := scanner.RunTests(cfg)

	// Save results to history if enabled
	if historyManager != nil {
		if err := historyManager.SaveScanResults(results); err != nil {
			logging.Error("Failed to save scan results to history", map[string]interface{}{
				"error": err.Error(),
			})
		}
	}

	// Generate historical comparison if enabled
	var comparison *history.ComparisonResult
	if historyManager != nil && cfg.HistoricalData.ComparePrevious {
		var err error
		comparison, err = historyManager.CompareWithPrevious(results)
		if err != nil {
			logging.Warn("Failed to generate historical comparison", map[string]interface{}{
				"error": err.Error(),
			})
		}
	}

	// Generate trend analysis if enabled
	var trendData *history.TrendData
	if historyManager != nil && cfg.HistoricalData.TrendAnalysis {
		var err error
		trendData, err = historyManager.GenerateTrendAnalysis()
		if err != nil {
			logging.Warn("Failed to generate trend analysis", map[string]interface{}{
				"error": err.Error(),
			})
		}
	}

	// Generate report based on selected format
	switch *outputFormat {
	case "json":
		scanner.GenerateJSONReport(results)
		if comparison != nil {
			history.GenerateHistoricalComparisonJSON(comparison)
		}
		if trendData != nil {
			history.GenerateTrendAnalysisJSON(trendData)
		}
	case "html":
		scanner.GenerateHTMLReport(results)
		if comparison != nil {
			history.GenerateHistoricalComparisonHTML(comparison)
		}
		if trendData != nil {
			history.GenerateTrendAnalysisHTML(trendData)
		}
	case "csv":
		scanner.GenerateCSVReport(results)
	case "xml":
		scanner.GenerateXMLReport(results)
	default:
		scanner.GenerateDetailedReport(results)
		if comparison != nil {
			history.GenerateHistoricalComparisonText(comparison)
		}
		if trendData != nil {
			history.GenerateTrendAnalysisText(trendData)
		}
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
