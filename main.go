package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"api-security-scanner/auth"
	"api-security-scanner/config"
	"api-security-scanner/history"
	"api-security-scanner/logging"
	"api-security-scanner/metrics"
	"api-security-scanner/scanner"
	"api-security-scanner/siem"
	"api-security-scanner/tenant"
	"api-security-scanner/types"
)

func main() {
	// Parse command line flags
	configFile := flag.String("config", "config.yaml", "Configuration file path")
	scanMode := flag.Bool("scan", false, "Run scan immediately")
	dashboardMode := flag.Bool("dashboard", false, "Start monitoring dashboard")
	tenantID := flag.String("tenant", "default", "Tenant ID for multi-tenant mode")
	outputFormat := flag.String("output", "json", "Output format (json, html, text)")
	historicalMode := flag.Bool("historical", false, "Show historical comparison")
	trendMode := flag.Bool("trend", false, "Show trend analysis")
	version := flag.Bool("version", false, "Show version information")

	flag.Parse()

	if *version {
		fmt.Println("API Security Scanner - Phase 4 Enterprise Edition")
		fmt.Println("Version 4.0.0")
		fmt.Println("Features: Multi-tenant support, SIEM integration, Advanced authentication, Performance monitoring")
		os.Exit(0)
	}

	// Initialize logging
	logging.SetGlobalLevel(logging.INFO)
	logging.SetGlobalFormat("json")

	logging.Info("Starting API Security Scanner - Phase 4 Enterprise Edition", map[string]interface{}{
		"version":    "4.0.0",
		"config":     *configFile,
		"tenant_id":  *tenantID,
		"scan_mode":  *scanMode,
		"dashboard": *dashboardMode,
	})

	// Load configuration
	appConfig, err := config.Load(*configFile)
	if err != nil {
		logging.Error("Failed to load configuration", map[string]interface{}{
			"error": err.Error(),
			"file":  *configFile,
		})
		log.Fatalf("Failed to load configuration: %v", err)
	}

	// Note: Tenant manager initialization removed for simplicity

	// Create tenant configuration from app config
	tenantConfig := &tenant.Tenant{
		ID:          appConfig.Tenant.ID,
		Name:        appConfig.Tenant.Name,
		Description: appConfig.Tenant.Description,
		IsActive:    appConfig.Tenant.IsActive,
		// Note: Settings would need to be converted but we'll keep it simple for now
	}

	// Initialize metrics collector
	metricsCollector := metrics.NewMetricsCollector(*appConfig.Metrics)

	// Initialize SIEM client if enabled
	var siemClient *siem.SIEMClient
	if appConfig.SIEM != nil && appConfig.SIEM.Enabled {
		// Convert config.SIEM to tenant.SIEMConfig
		siemConfig := &tenant.SIEMConfig{
			Enabled:     appConfig.SIEM.Enabled,
			Type:        tenant.SIEMType(appConfig.SIEM.Type),
			Config:      make(map[string]string),
			Format:      tenant.SIEMFormat(appConfig.SIEM.Format),
			EndpointURL: appConfig.SIEM.EndpointURL,
			AuthToken:   appConfig.SIEM.AuthToken,
		}
		// Convert interface{} config to string config
		for k, v := range appConfig.SIEM.Config {
			if str, ok := v.(string); ok {
				siemConfig.Config[k] = str
			}
		}
		siemClient, err = siem.NewSIEMClient(*siemConfig)
		if err != nil {
			logging.Error("Failed to initialize SIEM client", map[string]interface{}{
				"error": err.Error(),
			})
		} else {
			logging.Info("SIEM client initialized", map[string]interface{}{
				"type": appConfig.SIEM.Type,
			})
		}
	}

	// Initialize advanced authentication if enabled
	var authManager *auth.AdvancedAuthManager
	if appConfig.Auth.Enabled {
		// Create auth config from application config
		authConfig := &auth.AdvancedAuthConfig{
			Enabled: true,
			Type:    auth.AuthType(appConfig.Auth.Type),
		}
		authManager, err = auth.NewAdvancedAuthManager(authConfig)
		if err != nil {
			logging.Error("Failed to initialize authentication manager", map[string]interface{}{
				"error": err.Error(),
			})
		} else {
			logging.Info("Authentication manager initialized", map[string]interface{}{
				"type": appConfig.Auth.Type,
			})
		}
	}

	// Initialize historical data manager
	historyManager, err := history.NewHistoryManager(history.HistoricalData{
		Enabled:         true,
		StoragePath:     "./history",
		RetentionDays:   30,
		ComparePrevious: true,
		TrendAnalysis:   true,
	})
	if err != nil {
		logging.Error("Failed to initialize history manager", map[string]interface{}{
			"error": err.Error(),
		})
		log.Fatalf("Failed to initialize history manager: %v", err)
	}

	// Start dashboard server if requested
	if *dashboardMode {
		startDashboard(appConfig, metricsCollector)
	}

	// Run scan if requested
	if *scanMode {
		runScan(appConfig, tenantConfig, metricsCollector, siemClient, historyManager, authManager)
	}

	// Show historical data if requested
	if *historicalMode {
		showHistoricalComparison(historyManager, *outputFormat)
	}

	// Show trend analysis if requested
	if *trendMode {
		showTrendAnalysis(historyManager, *outputFormat)
	}

	// If no specific mode requested, start the full application
	if !*scanMode && !*dashboardMode && !*historicalMode && !*trendMode {
		startFullApplication(appConfig, tenantConfig, metricsCollector, siemClient, historyManager, authManager)
	}
}

func startDashboard(appConfig *config.Config, metricsCollector *metrics.MetricsCollector) {
	logging.Info("Starting dashboard server", map[string]interface{}{
		"port": appConfig.Metrics.Port,
	})

	dashboardServer := metrics.NewDashboardServer(metricsCollector, *appConfig.Metrics)
	if err := dashboardServer.Start(); err != nil {
		logging.Error("Failed to start dashboard server", map[string]interface{}{
			"error": err.Error(),
		})
		log.Fatalf("Failed to start dashboard server: %v", err)
	}

	// Wait for interrupt signal
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan

	logging.Info("Shutting down dashboard server", nil)
	dashboardServer.Stop()
}

func runScan(appConfig *config.Config, tenantConfig *tenant.Tenant, metricsCollector *metrics.MetricsCollector, siemClient *siem.SIEMClient, historyManager *history.HistoryManager, authManager *auth.AdvancedAuthManager) {
	logging.Info("Starting security scan", map[string]interface{}{
		"tenant_id": tenantConfig.ID,
		"endpoints": len(appConfig.Scanner.APIEndpoints),
	})

	// Create scanner instance - RunTests is the main function
	results := scanner.RunTests(appConfig.Scanner)

	// Generate scan ID
	scanID := fmt.Sprintf("scan_%d", time.Now().Unix())

	// Start metrics collection
	metricsCollector.StartScan(scanID, tenantConfig.ID, len(appConfig.Scanner.APIEndpoints))

	// Start resource monitoring
	resourceMonitor := metrics.NewResourceMonitor(scanID, tenantConfig.ID, metricsCollector)
	resourceMonitor.Start()
	defer resourceMonitor.Stop()

	// Results are already generated from RunTests above

	// Record metrics for each endpoint
	for _, result := range results {
		metricsCollector.RecordEndpointTest(scanID, result.URL, time.Duration(100), result.Results)
	}

	// End metrics collection
	metricsCollector.EndScan(scanID)

	// Save results to history
	if err := historyManager.SaveScanResults(results); err != nil {
		logging.Error("Failed to save scan results to history", map[string]interface{}{
			"error": err.Error(),
		})
	}

	// Send results to SIEM if enabled
	if siemClient != nil {
		events := siem.ConvertScanResultsToEvents(tenantConfig.ID, results)
		if err := siemClient.SendBatchEvents(events); err != nil {
			logging.Error("Failed to send events to SIEM", map[string]interface{}{
				"error": err.Error(),
			})
		} else {
			logging.Info("Scan results sent to SIEM", map[string]interface{}{
				"events_count": len(events),
			})
		}
	}

	// Generate output report - results are already available
	fmt.Printf("Scan completed successfully. Processed %d endpoints.\n", len(results))

	logging.Info("Scan completed successfully", map[string]interface{}{
		"scan_id":     scanID,
		"endpoints":   len(results),
		"vulnerabilities": countVulnerabilities(results),
	})
}

func showHistoricalComparison(historyManager *history.HistoryManager, outputFormat string) {
	comparison, err := historyManager.CompareWithPrevious([]types.EndpointResult{})
	if err != nil {
		logging.Error("Failed to get historical comparison", map[string]interface{}{
			"error": err.Error(),
		})
		log.Fatalf("Failed to get historical comparison: %v", err)
	}

	switch outputFormat {
	case "json":
		history.GenerateHistoricalComparisonJSON(comparison)
	case "html":
		history.GenerateHistoricalComparisonHTML(comparison)
	case "text":
		history.GenerateHistoricalComparisonText(comparison)
	default:
		log.Fatalf("Unsupported output format: %s", outputFormat)
	}
}

func showTrendAnalysis(historyManager *history.HistoryManager, outputFormat string) {
	trendData, err := historyManager.GenerateTrendAnalysis()
	if err != nil {
		logging.Error("Failed to get trend analysis", map[string]interface{}{
			"error": err.Error(),
		})
		log.Fatalf("Failed to get trend analysis: %v", err)
	}

	switch outputFormat {
	case "json":
		history.GenerateTrendAnalysisJSON(trendData)
	case "html":
		history.GenerateTrendAnalysisHTML(trendData)
	case "text":
		history.GenerateTrendAnalysisText(trendData)
	default:
		log.Fatalf("Unsupported output format: %s", outputFormat)
	}
}

func startFullApplication(appConfig *config.Config, tenantConfig *tenant.Tenant, metricsCollector *metrics.MetricsCollector, siemClient *siem.SIEMClient, historyManager *history.HistoryManager, authManager *auth.AdvancedAuthManager) {
	logging.Info("Starting full application", map[string]interface{}{
		"dashboard_port": appConfig.Metrics.Port,
		"server_port":    appConfig.Server.Port,
	})

	// Start dashboard server
	dashboardServer := metrics.NewDashboardServer(metricsCollector, *appConfig.Metrics)
	if err := dashboardServer.Start(); err != nil {
		logging.Error("Failed to start dashboard server", map[string]interface{}{
			"error": err.Error(),
		})
		log.Fatalf("Failed to start dashboard server: %v", err)
	}
	defer dashboardServer.Stop()

	// Start system monitoring
	systemMonitor := metrics.NewMonitor(5 * time.Second)
	systemMonitor.Start()
	defer systemMonitor.Stop()

	// Create context for graceful shutdown
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Start periodic scanning
	go func() {
		ticker := time.NewTicker(1 * time.Hour)
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				logging.Info("Starting periodic scan", nil)
				scanID := fmt.Sprintf("periodic_%d", time.Now().Unix())

				metricsCollector.StartScan(scanID, tenantConfig.ID, len(appConfig.Scanner.APIEndpoints))
				resourceMonitor := metrics.NewResourceMonitor(scanID, tenantConfig.ID, metricsCollector)
				resourceMonitor.Start()

				results := scanner.RunTests(appConfig.Scanner)

				if len(results) == 0 {
					logging.Error("Periodic scan failed", map[string]interface{}{
						"error": "No results returned",
					})
				} else {
					for _, result := range results {
						metricsCollector.RecordEndpointTest(scanID, result.URL, time.Duration(100), result.Results)
					}

					metricsCollector.EndScan(scanID)
					historyManager.SaveScanResults(results)

					if siemClient != nil {
						events := siem.ConvertScanResultsToEvents(tenantConfig.ID, results)
						siemClient.SendBatchEvents(events)
					}
				}

				resourceMonitor.Stop()

			case <-ctx.Done():
				return
			}
		}
	}()

	// Wait for interrupt signal
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan

	logging.Info("Shutting down application", nil)
	cancel()

	// Cleanup old data
	metricsCollector.CleanupOldData()
}

func countVulnerabilities(results []types.EndpointResult) int {
	count := 0
	for _, result := range results {
		for _, test := range result.Results {
			if !test.Passed {
				count++
			}
		}
	}
	return count
}
