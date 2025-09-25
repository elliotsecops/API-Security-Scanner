package metrics

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"sync"
	"time"

	"golang.org/x/net/websocket"

	"api-security-scanner/logging"
)

// DashboardServer serves the monitoring dashboard
type DashboardServer struct {
	server      *http.Server
	collector   *MetricsCollector
	guiPath     string
	mutex       sync.RWMutex
	clients     map[chan []byte]bool
	credentials *CredentialManager
}

// NewDashboardServer creates a new dashboard server
func NewDashboardServer(collector *MetricsCollector, config Dashboard) *DashboardServer {
	// Check if GUI build exists
	guiPath := "./gui/build"
	if _, err := os.Stat(guiPath); os.IsNotExist(err) {
		// Fall back to legacy HTML dashboard if GUI not built
		guiPath = ""
		logging.Info("GUI build not found, using legacy dashboard", nil)
	} else {
		logging.Info("Using React GUI", map[string]interface{}{"path": guiPath})
	}

	credManager, err := NewCredentialManager("")
	if err != nil {
		logging.Error("Failed to initialize dashboard credential store", map[string]interface{}{
			"error": err.Error(),
		})
		if fallback, fallbackErr := NewInMemoryCredentialManager(); fallbackErr == nil {
			credManager = fallback
		} else {
			logging.Error("Fallback credential store failed", map[string]interface{}{
				"error": fallbackErr.Error(),
			})
		}
	}

	server := &DashboardServer{
		collector:   collector,
		guiPath:     guiPath,
		clients:     make(map[chan []byte]bool),
		credentials: credManager,
	}

	server.server = &http.Server{
		Addr:    fmt.Sprintf(":%d", config.Port),
		Handler: server,
	}

	return server
}

// Start starts the dashboard server
func (ds *DashboardServer) Start() error {
	logging.Info("Starting dashboard server", map[string]interface{}{
		"port": ds.server.Addr,
	})

	go func() {
		if err := ds.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logging.Error("Dashboard server error", map[string]interface{}{
				"error": err.Error(),
			})
		}
	}()

	return nil
}

// Stop stops the dashboard server
func (ds *DashboardServer) Stop() error {
	if ds.server != nil {
		return ds.server.Close()
	}
	return nil
}

// ServeHTTP handles HTTP requests
func (ds *DashboardServer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// Serve API endpoints
	switch r.URL.Path {
	case "/api/metrics":
		ds.serveMetrics(w, r)
		return
	case "/api/metrics/websocket":
		ds.handleWebSocket(w, r)
		return
	case "/api/system":
		ds.serveSystemMetrics(w, r)
		return
	case "/api/tenant":
		ds.serveTenantMetrics(w, r)
		return
	case "/api/export":
		ds.serveExport(w, r)
		return
	case "/api/scans":
		ds.serveScans(w, r)
		return
	case "/api/auth/login":
		ds.serveAuthLogin(w, r)
		return
	case "/api/auth/change-password":
		ds.serveAuthUpdate(w, r)
		return
	case "/api/tenants":
		ds.serveTenants(w, r)
		return
	case "/ws":
		ds.handleWebSocket(w, r)
		return
	case "/metrics":
		ds.servePrometheusMetrics(w, r)
		return
	}

	// Serve GUI static files if available
	if ds.guiPath != "" {
		// Try to serve static file
		filePath := filepath.Join(ds.guiPath, r.URL.Path)
		if _, err := os.Stat(filePath); err == nil {
			http.ServeFile(w, r, filePath)
			return
		}

		// Try to serve index.html for SPA routing
		if r.URL.Path != "/" && r.URL.Path != "/favicon.ico" {
			indexFile := filepath.Join(ds.guiPath, "index.html")
			if _, err := os.Stat(indexFile); err == nil {
				http.ServeFile(w, r, indexFile)
				return
			}
		}

		// Serve index.html for root path
		if r.URL.Path == "/" {
			indexFile := filepath.Join(ds.guiPath, "index.html")
			http.ServeFile(w, r, indexFile)
			return
		}
	}

	// Fall back to legacy dashboard if GUI not available
	if r.URL.Path == "/" {
		ds.serveDashboard(w, r)
	} else {
		http.NotFound(w, r)
	}
}

// serveDashboard serves the main dashboard page
func (ds *DashboardServer) serveDashboard(w http.ResponseWriter, r *http.Request) {
	// Simple HTML fallback if GUI not available
	w.Header().Set("Content-Type", "text/html")
	w.Write([]byte(`<!DOCTYPE html>
<html>
<head>
    <title>API Security Scanner</title>
    <style>
        body { font-family: Arial, sans-serif; margin: 40px; }
        .card { background: #f5f5f5; padding: 20px; margin: 20px 0; border-radius: 8px; }
        .metric { font-size: 24px; font-weight: bold; color: #667eea; }
    </style>
</head>
<body>
    <h1>API Security Scanner Dashboard</h1>
    <div class="card">
        <div class="metric" id="totalScans">-</div>
        <div>Total Scans</div>
    </div>
    <div class="card">
        <div class="metric" id="activeTenants">-</div>
        <div>Active Tenants</div>
    </div>
    <script>
        fetch('/api/system')
            .then(response => response.json())
            .then(data => {
                document.getElementById('totalScans').textContent = data.total_scans || 0;
                document.getElementById('activeTenants').textContent = data.active_tenants || 0;
            });
    </script>
</body>
</html>`))
}

// serveMetrics serves metrics data
func (ds *DashboardServer) serveMetrics(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	timeRange := ds.parseTimeRange(r.URL.Query().Get("range"))
	if timeRange == 0 {
		timeRange = 24 * time.Hour
	}

	systemMetrics := ds.collector.GetSystemMetrics(timeRange)
	if systemMetrics == nil {
		w.Write([]byte("{}"))
		return
	}

	data, err := json.MarshalIndent(systemMetrics, "", "  ")
	if err != nil {
		logging.Error("Failed to marshal metrics", map[string]interface{}{
			"error": err.Error(),
		})
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	w.Write(data)
}

// serveSystemMetrics serves system-wide metrics
func (ds *DashboardServer) serveSystemMetrics(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	timeRange := ds.parseTimeRange(r.URL.Query().Get("range"))
	if timeRange == 0 {
		timeRange = 24 * time.Hour
	}

	metrics := ds.collector.GetSystemMetrics(timeRange)
	if metrics == nil {
		w.Write([]byte("{}"))
		return
	}

	data, err := json.MarshalIndent(metrics, "", "  ")
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	w.Write(data)
}

// serveTenantMetrics serves tenant-specific metrics
func (ds *DashboardServer) serveTenantMetrics(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	tenantID := r.URL.Query().Get("tenant_id")
	if tenantID == "" {
		http.Error(w, "Tenant ID is required", http.StatusBadRequest)
		return
	}

	timeRange := ds.parseTimeRange(r.URL.Query().Get("range"))
	if timeRange == 0 {
		timeRange = 24 * time.Hour
	}

	metrics := ds.collector.GetTenantMetrics(tenantID, timeRange)
	if metrics == nil {
		w.Write([]byte("{}"))
		return
	}

	data, err := json.MarshalIndent(metrics, "", "  ")
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	w.Write(data)
}

// serveExport serves exported metrics data
func (ds *DashboardServer) serveExport(w http.ResponseWriter, r *http.Request) {
	format := r.URL.Query().Get("format")
	if format == "" {
		format = "json"
	}

	data, err := ds.collector.ExportMetrics(format)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	switch format {
	case "json":
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("Content-Disposition", "attachment; filename=metrics.json")
	case "prometheus":
		w.Header().Set("Content-Type", "text/plain; version=0.0.4")
		w.Header().Set("Content-Disposition", "attachment; filename=metrics.prometheus")
	}

	w.Write(data)
}

// handleWebSocket handles WebSocket connections for real-time updates
func (ds *DashboardServer) handleWebSocket(w http.ResponseWriter, r *http.Request) {
	tenantID := r.URL.Query().Get("tenant")
	if tenantID == "" {
		tenantID = "default"
	}

	timeRange := ds.parseTimeRange(r.URL.Query().Get("range"))
	if timeRange <= 0 {
		timeRange = 24 * time.Hour
	}

	updateInterval := ds.collector.config.UpdateInterval
	if updateInterval <= 0 {
		updateInterval = 30 * time.Second
	}

	sender := websocket.Handler(func(conn *websocket.Conn) {
		defer conn.Close()

		sendSnapshot := func() error {
			payload := map[string]interface{}{
				"type":      "metrics",
				"timestamp": time.Now().UTC().Format(time.RFC3339),
				"tenant_id": tenantID,
				"data": map[string]interface{}{
					"tenant": ds.collector.GetTenantMetrics(tenantID, timeRange),
					"system": ds.collector.GetSystemMetrics(timeRange),
				},
			}

			if err := websocket.JSON.Send(conn, payload); err != nil {
				logging.Error("WebSocket send error", map[string]interface{}{
					"error": err.Error(),
				})
				return err
			}

			return nil
		}

		if err := sendSnapshot(); err != nil {
			return
		}

		ticker := time.NewTicker(updateInterval)
		defer ticker.Stop()

		for range ticker.C {
			if err := sendSnapshot(); err != nil {
				return
			}
		}
	})

	sender.ServeHTTP(w, r)
}

// parseTimeRange parses time range from string
func (ds *DashboardServer) parseTimeRange(rangeStr string) time.Duration {
	switch rangeStr {
	case "1h":
		return time.Hour
	case "24h":
		return 24 * time.Hour
	case "7d":
		return 7 * 24 * time.Hour
	case "30d":
		return 30 * 24 * time.Hour
	default:
		return 24 * time.Hour
	}
}

// serveScans serves scan data for GUI
func (ds *DashboardServer) serveScans(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	// Mock scan data for demonstration
	scans := []map[string]interface{}{
		{
			"id":                       "scan_001",
			"name":                     "Daily Security Scan",
			"tenant_id":                "tenant-001",
			"started_at":               "2024-01-15T10:00:00Z",
			"completed_at":             "2024-01-15T10:15:00Z",
			"average_score":            85.5,
			"risk_level":               "medium",
			"total_vulnerabilities":    12,
			"critical_vulnerabilities": 2,
			"high_vulnerabilities":     5,
			"medium_vulnerabilities":   3,
			"low_vulnerabilities":      2,
			"duration":                 900,
			"endpoints": []map[string]interface{}{
				{
					"url":             "https://api.example.com/users",
					"method":          "GET",
					"score":           90,
					"vulnerabilities": 3,
					"status":          "completed",
				},
			},
		},
	}

	data, err := json.MarshalIndent(scans, "", "  ")
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	w.Write(data)
}

// serveAuthLogin handles authentication
func (ds *DashboardServer) serveAuthLogin(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	if ds.credentials == nil {
		http.Error(w, "Credential store unavailable", http.StatusInternalServerError)
		return
	}

	var credentials struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}

	if err := json.NewDecoder(r.Body).Decode(&credentials); err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	if !ds.credentials.Authenticate(credentials.Username, credentials.Password) {
		http.Error(w, "Invalid credentials", http.StatusUnauthorized)
		return
	}

	response := map[string]interface{}{
		"token": fmt.Sprintf("mock-jwt-token-%d", time.Now().Unix()),
		"user": map[string]interface{}{
			"id":       "1",
			"username": ds.credentials.Username(),
			"role":     "administrator",
		},
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// serveAuthUpdate allows rotating the dashboard credentials.
func (ds *DashboardServer) serveAuthUpdate(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" && r.Method != "PUT" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	if ds.credentials == nil {
		http.Error(w, "Credential store unavailable", http.StatusInternalServerError)
		return
	}

	var payload struct {
		CurrentUsername string `json:"current_username"`
		CurrentPassword string `json:"current_password"`
		NewUsername     string `json:"new_username"`
		NewPassword     string `json:"new_password"`
		ConfirmPassword string `json:"confirm_password"`
	}

	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	if payload.NewPassword != payload.ConfirmPassword {
		http.Error(w, "New password confirmation does not match", http.StatusBadRequest)
		return
	}

	if !ds.credentials.Authenticate(payload.CurrentUsername, payload.CurrentPassword) {
		http.Error(w, "Current credentials are incorrect", http.StatusUnauthorized)
		return
	}

	if err := ds.credentials.Update(payload.NewUsername, payload.NewPassword); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	logging.Info("Dashboard credentials updated", map[string]interface{}{
		"username":   ds.credentials.Username(),
		"updated_at": ds.credentials.LastUpdated().Format(time.RFC3339),
	})

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"message":    "Credentials updated successfully",
		"username":   ds.credentials.Username(),
		"updated_at": ds.credentials.LastUpdated().Format(time.RFC3339),
	})
}

// serveTenants serves tenant data
func (ds *DashboardServer) serveTenants(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	// Mock tenant data
	tenants := []map[string]interface{}{
		{
			"id":          "tenant-001",
			"name":        "Enterprise Corp",
			"description": "Main enterprise tenant",
			"is_active":   true,
			"settings": map[string]interface{}{
				"max_endpoints":  100,
				"scan_frequency": "daily",
			},
		},
		{
			"id":          "tenant-002",
			"name":        "Development Team",
			"description": "Development and testing tenant",
			"is_active":   true,
			"settings": map[string]interface{}{
				"max_endpoints":  50,
				"scan_frequency": "hourly",
			},
		},
	}

	data, err := json.MarshalIndent(tenants, "", "  ")
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	w.Write(data)
}

// servePrometheusMetrics serves metrics in Prometheus format
func (ds *DashboardServer) servePrometheusMetrics(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/plain; version=0.0.4")

	// Get metrics data
	systemMetrics := ds.collector.GetSystemMetrics(24 * time.Hour)

	// Create Prometheus-formatted output
	var prometheusData string

	// System-level metrics HELP and TYPE definitions (always present)
	prometheusData += fmt.Sprintf("# HELP api_scanner_total_scans Total number of scans\n")
	prometheusData += fmt.Sprintf("# TYPE api_scanner_total_scans counter\n")
	
	prometheusData += fmt.Sprintf("# HELP api_scanner_total_endpoints Total number of endpoints tested\n")
	prometheusData += fmt.Sprintf("# TYPE api_scanner_total_endpoints gauge\n")
	
	prometheusData += fmt.Sprintf("# HELP api_scanner_total_vulnerabilities Total number of vulnerabilities\n")
	prometheusData += fmt.Sprintf("# TYPE api_scanner_total_vulnerabilities gauge\n")
	
	prometheusData += fmt.Sprintf("# HELP api_scanner_critical_vulnerabilities Number of critical vulnerabilities\n")
	prometheusData += fmt.Sprintf("# TYPE api_scanner_critical_vulnerabilities gauge\n")
	
	prometheusData += fmt.Sprintf("# HELP api_scanner_high_vulnerabilities Number of high vulnerabilities\n")
	prometheusData += fmt.Sprintf("# TYPE api_scanner_high_vulnerabilities gauge\n")
	
	prometheusData += fmt.Sprintf("# HELP api_scanner_medium_vulnerabilities Number of medium vulnerabilities\n")
	prometheusData += fmt.Sprintf("# TYPE api_scanner_medium_vulnerabilities gauge\n")
	
	prometheusData += fmt.Sprintf("# HELP api_scanner_low_vulnerabilities Number of low vulnerabilities\n")
	prometheusData += fmt.Sprintf("# TYPE api_scanner_low_vulnerabilities gauge\n")
	
	prometheusData += fmt.Sprintf("# HELP api_scanner_active_tenants Number of active tenants\n")
	prometheusData += fmt.Sprintf("# TYPE api_scanner_active_tenants gauge\n")
	
	prometheusData += fmt.Sprintf("# HELP api_scanner_avg_response_time Average response time in milliseconds\n")
	prometheusData += fmt.Sprintf("# TYPE api_scanner_avg_response_time gauge\n")
	
	prometheusData += fmt.Sprintf("# HELP api_scanner_throughput Requests per second\n")
	prometheusData += fmt.Sprintf("# TYPE api_scanner_throughput gauge\n")
	
	prometheusData += fmt.Sprintf("# HELP api_scanner_error_rate Percentage of errors\n")
	prometheusData += fmt.Sprintf("# TYPE api_scanner_error_rate gauge\n")
	
	prometheusData += fmt.Sprintf("# HELP api_scanner_cpu_usage CPU usage percentage\n")
	prometheusData += fmt.Sprintf("# TYPE api_scanner_cpu_usage gauge\n")
	
	prometheusData += fmt.Sprintf("# HELP api_scanner_memory_usage Memory usage in MB\n")
	prometheusData += fmt.Sprintf("# TYPE api_scanner_memory_usage gauge\n")
	
	prometheusData += fmt.Sprintf("# HELP api_scanner_goroutines Number of goroutines\n")
	prometheusData += fmt.Sprintf("# TYPE api_scanner_goroutines gauge\n")
	
	prometheusData += fmt.Sprintf("# HELP api_scanner_tenant_total_scans Total scans for tenant\n")
	prometheusData += fmt.Sprintf("# TYPE api_scanner_tenant_total_scans gauge\n")
	
	prometheusData += fmt.Sprintf("# HELP api_scanner_tenant_critical_vulnerabilities Critical vulnerabilities for tenant\n")
	prometheusData += fmt.Sprintf("# TYPE api_scanner_tenant_critical_vulnerabilities gauge\n")
	
	prometheusData += fmt.Sprintf("# HELP api_scanner_tenant_high_vulnerabilities High vulnerabilities for tenant\n")
	prometheusData += fmt.Sprintf("# TYPE api_scanner_tenant_high_vulnerabilities gauge\n")

	// Add metric values if system metrics exist
	if systemMetrics != nil {
		prometheusData += fmt.Sprintf("api_scanner_total_scans %d\n", systemMetrics.TotalScans)
		prometheusData += fmt.Sprintf("api_scanner_total_endpoints %d\n", systemMetrics.TotalEndpoints)
		prometheusData += fmt.Sprintf("api_scanner_total_vulnerabilities %d\n", systemMetrics.Vulnerabilities.Total)
		prometheusData += fmt.Sprintf("api_scanner_critical_vulnerabilities %d\n", systemMetrics.Vulnerabilities.Critical)
		prometheusData += fmt.Sprintf("api_scanner_high_vulnerabilities %d\n", systemMetrics.Vulnerabilities.High)
		prometheusData += fmt.Sprintf("api_scanner_medium_vulnerabilities %d\n", systemMetrics.Vulnerabilities.Medium)
		prometheusData += fmt.Sprintf("api_scanner_low_vulnerabilities %d\n", systemMetrics.Vulnerabilities.Low)
		prometheusData += fmt.Sprintf("api_scanner_active_tenants %d\n", systemMetrics.ActiveTenants)

		// Performance metrics
		if systemMetrics.Performance.AvgResponseTime > 0 {
			prometheusData += fmt.Sprintf("api_scanner_avg_response_time %f\n", float64(systemMetrics.Performance.AvgResponseTime)/float64(time.Millisecond))
		} else {
			prometheusData += fmt.Sprintf("api_scanner_avg_response_time 0.0\n")
		}

		prometheusData += fmt.Sprintf("api_scanner_throughput %f\n", systemMetrics.Performance.Throughput)
		prometheusData += fmt.Sprintf("api_scanner_error_rate %f\n", systemMetrics.Performance.ErrorRate)

		// Resource usage metrics
		prometheusData += fmt.Sprintf("api_scanner_cpu_usage %f\n", systemMetrics.ResourceUsage.CPUUsage)
		prometheusData += fmt.Sprintf("api_scanner_memory_usage %f\n", systemMetrics.ResourceUsage.MemoryUsage)
		prometheusData += fmt.Sprintf("api_scanner_goroutines %d\n", systemMetrics.ResourceUsage.Goroutines)
	} else {
		// If no system metrics, still output default zero values
		prometheusData += fmt.Sprintf("api_scanner_total_scans 0\n")
		prometheusData += fmt.Sprintf("api_scanner_total_endpoints 0\n")
		prometheusData += fmt.Sprintf("api_scanner_total_vulnerabilities 0\n")
		prometheusData += fmt.Sprintf("api_scanner_critical_vulnerabilities 0\n")
		prometheusData += fmt.Sprintf("api_scanner_high_vulnerabilities 0\n")
		prometheusData += fmt.Sprintf("api_scanner_medium_vulnerabilities 0\n")
		prometheusData += fmt.Sprintf("api_scanner_low_vulnerabilities 0\n")
		prometheusData += fmt.Sprintf("api_scanner_active_tenants 0\n")
		prometheusData += fmt.Sprintf("api_scanner_avg_response_time 0.0\n")
		prometheusData += fmt.Sprintf("api_scanner_throughput 0.0\n")
		prometheusData += fmt.Sprintf("api_scanner_error_rate 0.0\n")
		prometheusData += fmt.Sprintf("api_scanner_cpu_usage 0.0\n")
		prometheusData += fmt.Sprintf("api_scanner_memory_usage 0.0\n")
		prometheusData += fmt.Sprintf("api_scanner_goroutines 0\n")
	}

	// Add metrics for each tenant
	if systemMetrics != nil && systemMetrics.TenantMetrics != nil {
		for tenantID, tenantMetrics := range systemMetrics.TenantMetrics {
			if tenantMetrics != nil {
				tenantLabel := fmt.Sprintf(`{tenant_id="%s"}`, tenantID)
				prometheusData += fmt.Sprintf("api_scanner_tenant_total_scans%s %d\n", tenantLabel, tenantMetrics.TotalScans)
				prometheusData += fmt.Sprintf("api_scanner_tenant_critical_vulnerabilities%s %d\n", tenantLabel, tenantMetrics.Vulnerabilities.Critical)
				prometheusData += fmt.Sprintf("api_scanner_tenant_high_vulnerabilities%s %d\n", tenantLabel, tenantMetrics.Vulnerabilities.High)
			}
		}
	} else {
		// Provide default tenant metrics with empty labels if no tenant data
		prometheusData += fmt.Sprintf("api_scanner_tenant_total_scans{tenant_id=\"default\"} 0\n")
		prometheusData += fmt.Sprintf("api_scanner_tenant_critical_vulnerabilities{tenant_id=\"default\"} 0\n")
		prometheusData += fmt.Sprintf("api_scanner_tenant_high_vulnerabilities{tenant_id=\"default\"} 0\n")
	}

	w.Write([]byte(prometheusData))
}

// Dashboard HTML template
const dashboardHTML = `
<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>{{.Title}}</title>
    <style>
        body {
            font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, sans-serif;
            margin: 0;
            padding: 0;
            background-color: #f5f5f5;
        }
        .header {
            background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
            color: white;
            padding: 1rem;
            box-shadow: 0 2px 4px rgba(0,0,0,0.1);
        }
        .header h1 {
            margin: 0;
            font-size: 1.8rem;
        }
        .header .version {
            font-size: 0.9rem;
            opacity: 0.8;
        }
        .container {
            max-width: 1200px;
            margin: 0 auto;
            padding: 1rem;
        }
        .dashboard-grid {
            display: grid;
            grid-template-columns: repeat(auto-fit, minmax(300px, 1fr));
            gap: 1rem;
            margin-bottom: 2rem;
        }
        .card {
            background: white;
            border-radius: 8px;
            padding: 1.5rem;
            box-shadow: 0 2px 4px rgba(0,0,0,0.1);
        }
        .card h3 {
            margin: 0 0 1rem 0;
            color: #333;
            font-size: 1.2rem;
        }
        .metric-value {
            font-size: 2rem;
            font-weight: bold;
            color: #667eea;
        }
        .metric-label {
            color: #666;
            font-size: 0.9rem;
        }
        .metric-grid {
            display: grid;
            grid-template-columns: repeat(auto-fit, minmax(150px, 1fr));
            gap: 1rem;
        }
        .metric-item {
            text-align: center;
        }
        .progress-bar {
            width: 100%;
            height: 8px;
            background-color: #e0e0e0;
            border-radius: 4px;
            overflow: hidden;
            margin-top: 0.5rem;
        }
        .progress-fill {
            height: 100%;
            background-color: #667eea;
            transition: width 0.3s ease;
        }
        .severity-critical { color: #dc3545; }
        .severity-high { color: #fd7e14; }
        .severity-medium { color: #ffc107; }
        .severity-low { color: #28a745; }
        .controls {
            margin-bottom: 1rem;
            display: flex;
            gap: 1rem;
            align-items: center;
        }
        .controls select, .controls button {
            padding: 0.5rem;
            border: 1px solid #ddd;
            border-radius: 4px;
            background: white;
        }
        .controls button {
            background: #667eea;
            color: white;
            border: none;
            cursor: pointer;
        }
        .controls button:hover {
            background: #5a6fd8;
        }
        .chart-container {
            position: relative;
            height: 300px;
            margin-top: 1rem;
        }
        .loading {
            text-align: center;
            padding: 2rem;
            color: #666;
        }
        .error {
            background-color: #f8d7da;
            color: #721c24;
            padding: 1rem;
            border-radius: 4px;
            margin: 1rem 0;
        }
        .refresh-info {
            font-size: 0.8rem;
            color: #666;
            text-align: right;
            margin-top: 0.5rem;
        }
    </style>
</head>
<body>
    <div class="header">
        <h1>{{.Title}}</h1>
        <div class="version">Version {{.Version}}</div>
    </div>

    <div class="container">
        <div class="controls">
            <label>Time Range:</label>
            <select id="timeRange">
                <option value="1h">Last Hour</option>
                <option value="24h" selected>Last 24 Hours</option>
                <option value="7d">Last 7 Days</option>
                <option value="30d">Last 30 Days</option>
            </select>
            <button onclick="refreshData()">Refresh</button>
            <button onclick="exportData('json')">Export JSON</button>
            <button onclick="exportData('prometheus')">Export Prometheus</button>
            <div class="refresh-info">Auto-refresh in: <span id="countdown">30</span>s</div>
        </div>

        <div class="dashboard-grid">
            <!-- System Overview -->
            <div class="card">
                <h3>System Overview</h3>
                <div class="metric-grid">
                    <div class="metric-item">
                        <div class="metric-value" id="totalScans">-</div>
                        <div class="metric-label">Total Scans</div>
                    </div>
                    <div class="metric-item">
                        <div class="metric-value" id="activeTenants">-</div>
                        <div class="metric-label">Active Tenants</div>
                    </div>
                    <div class="metric-item">
                        <div class="metric-value" id="totalEndpoints">-</div>
                        <div class="metric-label">Total Endpoints</div>
                    </div>
                </div>
            </div>

            <!-- Vulnerability Summary -->
            <div class="card">
                <h3>Vulnerability Summary</h3>
                <div class="metric-grid">
                    <div class="metric-item">
                        <div class="metric-value severity-critical" id="criticalVulns">-</div>
                        <div class="metric-label">Critical</div>
                        <div class="progress-bar">
                            <div class="progress-fill severity-critical" id="criticalProgress"></div>
                        </div>
                    </div>
                    <div class="metric-item">
                        <div class="metric-value severity-high" id="highVulns">-</div>
                        <div class="metric-label">High</div>
                        <div class="progress-bar">
                            <div class="progress-fill severity-high" id="highProgress"></div>
                        </div>
                    </div>
                    <div class="metric-item">
                        <div class="metric-value severity-medium" id="mediumVulns">-</div>
                        <div class="metric-label">Medium</div>
                        <div class="progress-bar">
                            <div class="progress-fill severity-medium" id="mediumProgress"></div>
                        </div>
                    </div>
                    <div class="metric-item">
                        <div class="metric-value severity-low" id="lowVulns">-</div>
                        <div class="metric-label">Low</div>
                        <div class="progress-bar">
                            <div class="progress-fill severity-low" id="lowProgress"></div>
                        </div>
                    </div>
                </div>
            </div>

            <!-- Performance Metrics -->
            <div class="card">
                <h3>Performance Metrics</h3>
                <div class="metric-grid">
                    <div class="metric-item">
                        <div class="metric-value" id="avgResponseTime">-</div>
                        <div class="metric-label">Avg Response Time</div>
                    </div>
                    <div class="metric-item">
                        <div class="metric-value" id="throughput">-</div>
                        <div class="metric-label">Throughput (req/s)</div>
                    </div>
                    <div class="metric-item">
                        <div class="metric-value" id="errorRate">-</div>
                        <div class="metric-label">Error Rate (%)</div>
                    </div>
                </div>
            </div>

            <!-- Resource Usage -->
            <div class="card">
                <h3>Resource Usage</h3>
                <div class="metric-grid">
                    <div class="metric-item">
                        <div class="metric-value" id="cpuUsage">-</div>
                        <div class="metric-label">CPU Usage (%)</div>
                    </div>
                    <div class="metric-item">
                        <div class="metric-value" id="memoryUsage">-</div>
                        <div class="metric-label">Memory Usage (MB)</div>
                    </div>
                    <div class="metric-item">
                        <div class="metric-value" id="goroutines">-</div>
                        <div class="metric-label">Goroutines</div>
                    </div>
                </div>
            </div>
        </div>

        <!-- Top Vulnerable Endpoints -->
        <div class="card">
            <h3>Top Vulnerable Endpoints</h3>
            <div id="topEndpoints">Loading...</div>
        </div>

        <!-- Tenant Details -->
        <div class="card">
            <h3>Tenant Details</h3>
            <div class="controls">
                <label>Tenant:</label>
                <select id="tenantSelect">
                    <option value="">Select Tenant</option>
                </select>
                <button onclick="loadTenantMetrics()">Load</button>
            </div>
            <div id="tenantDetails">Select a tenant to view details</div>
        </div>
    </div>

    <script>
        let refreshInterval;
        let countdownInterval;
        let countdown = 30;

        function initialize() {
            refreshData();
            startAutoRefresh();
        }

        function startAutoRefresh() {
            refreshInterval = setInterval(() => {
                refreshData();
                countdown = 30;
            }, 30000);

            countdownInterval = setInterval(() => {
                countdown--;
                document.getElementById('countdown').textContent = countdown;
                if (countdown <= 0) {
                    countdown = 30;
                }
            }, 1000);
        }

        function refreshData() {
            const timeRange = document.getElementById('timeRange').value;
            loadSystemMetrics(timeRange);
            loadTenantOptions(timeRange);
        }

        async function loadSystemMetrics(timeRange) {
            try {
                const response = await fetch("/api/system?range=" + timeRange);
                const data = await response.json();

                document.getElementById('totalScans').textContent = data.total_scans || 0;
                document.getElementById('activeTenants').textContent = data.active_tenants || 0;
                document.getElementById('totalEndpoints').textContent = data.total_endpoints || 0;

                const vulns = data.vulnerabilities || {};
                const total = vulns.total || 0;

                document.getElementById('criticalVulns').textContent = vulns.critical || 0;
                document.getElementById('highVulns').textContent = vulns.high || 0;
                document.getElementById('mediumVulns').textContent = vulns.medium || 0;
                document.getElementById('lowVulns').textContent = vulns.low || 0;

                // Update progress bars
                updateProgressBar('criticalProgress', vulns.critical || 0, total);
                updateProgressBar('highProgress', vulns.high || 0, total);
                updateProgressBar('mediumProgress', vulns.medium || 0, total);
                updateProgressBar('lowProgress', vulns.low || 0, total);

                // Performance metrics
                const perf = data.performance || {};
                document.getElementById('avgResponseTime').textContent = formatDuration(perf.avg_response_time || 0);
                document.getElementById('throughput').textContent = (perf.throughput || 0).toFixed(2);
                document.getElementById('errorRate').textContent = (perf.error_rate || 0).toFixed(2);

                // Resource usage
                const resource = data.resource_usage || {};
                document.getElementById('cpuUsage').textContent = (resource.cpu_usage || 0).toFixed(1);
                document.getElementById('memoryUsage').textContent = (resource.memory_usage || 0).toFixed(1);
                document.getElementById('goroutines').textContent = resource.goroutines || 0;

                // Top endpoints
                loadTopEndpoints(data.vulnerabilities.by_endpoint || {});

            } catch (error) {
                console.error('Error loading system metrics:', error);
            }
        }

        function updateProgressBar(id, value, total) {
            const element = document.getElementById(id);
            const percentage = total > 0 ? (value / total) * 100 : 0;
            element.style.width = percentage + '%';
        }

        function loadTopEndpoints(endpoints) {
            const container = document.getElementById('topEndpoints');
            const sorted = Object.entries(endpoints)
                .sort(([,a], [,b]) => b - a)
                .slice(0, 10);

            if (sorted.length === 0) {
                container.innerHTML = '<p>No vulnerable endpoints found</p>';
                return;
            }

            let html = '<div class="metric-grid">';
            sorted.forEach(([endpoint, count]) => {
                html += '<div class="metric-item">' +
                    '<div class="metric-value">' + count + '</div>' +
                    '<div class="metric-label">' + endpoint.substring(0, 50) + (endpoint.length > 50 ? '...' : '') + '</div>' +
                    '</div>';
            });
            html += '</div>';
            container.innerHTML = html;
        }

        async function loadTenantOptions(timeRange) {
            try {
                const response = await fetch("/api/system?range=" + timeRange);
                const data = await response.json();

                const select = document.getElementById('tenantSelect');
                select.innerHTML = '<option value="">Select Tenant</option>';

                if (data.tenant_metrics) {
                    Object.keys(data.tenant_metrics).forEach(tenantId => {
                        const option = document.createElement('option');
                        option.value = tenantId;
                        option.textContent = tenantId;
                        select.appendChild(option);
                    });
                }
            } catch (error) {
                console.error('Error loading tenant options:', error);
            }
        }

        async function loadTenantMetrics() {
            const tenantId = document.getElementById('tenantSelect').value;
            const timeRange = document.getElementById('timeRange').value;

            if (!tenantId) return;

            try {
                const response = await fetch("/api/tenant?tenant_id=" + tenantId + "&range=" + timeRange);
                const data = await response.json();

                const container = document.getElementById('tenantDetails');
                container.innerHTML = '<div class="metric-grid">' +
                    '<div class="metric-item">' +
                        '<div class="metric-value">' + (data.total_scans || 0) + '</div>' +
                        '<div class="metric-label">Total Scans</div>' +
                    '</div>' +
                    '<div class="metric-item">' +
                        '<div class="metric-value">' + (data.total_endpoints || 0) + '</div>' +
                        '<div class="metric-label">Total Endpoints</div>' +
                    '</div>' +
                    '<div class="metric-item">' +
                        '<div class="metric-value severity-critical">' + (data.vulnerabilities.critical || 0) + '</div>' +
                        '<div class="metric-label">Critical</div>' +
                    '</div>' +
                    '<div class="metric-item">' +
                        '<div class="metric-value severity-high">' + (data.vulnerabilities.high || 0) + '</div>' +
                        '<div class="metric-label">High</div>' +
                    '</div>' +
                '</div>';
            } catch (error) {
                console.error('Error loading tenant metrics:', error);
                document.getElementById('tenantDetails').innerHTML = '<div class="error">Error loading tenant data</div>';
            }
        }

        async function exportData(format) {
            try {
                const response = await fetch("/api/export?format=" + format);
                const blob = await response.blob();
                const url = window.URL.createObjectURL(blob);
                const a = document.createElement('a');
                a.href = url;
                a.download = "metrics." + format;
                a.click();
                window.URL.revokeObjectURL(url);
            } catch (error) {
                console.error('Error exporting data:', error);
            }
        }

        function formatDuration(ms) {
            if (ms < 1000) return ms + 'ms';
            if (ms < 60000) return (ms / 1000).toFixed(1) + 's';
            return (ms / 60000).toFixed(1) + 'm';
        }

        // Event listeners
        document.getElementById('timeRange').addEventListener('change', refreshData);

        // Initialize on page load
        window.addEventListener('load', initialize);

        // Cleanup on page unload
        window.addEventListener('beforeunload', () => {
            if (refreshInterval) clearInterval(refreshInterval);
            if (countdownInterval) clearInterval(countdownInterval);
        });
    </script>
</body>
</html>
`
