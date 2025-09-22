package metrics

import (
	"encoding/json"
	"fmt"
	"html/template"
	"net/http"
	"time"
	"sync"

	"api-security-scanner/logging"
)

// DashboardServer serves the monitoring dashboard
type DashboardServer struct {
	server     *http.Server
	collector  *MetricsCollector
	templates  *template.Template
	mutex      sync.RWMutex
	clients    map[chan []byte]bool
}

// NewDashboardServer creates a new dashboard server
func NewDashboardServer(collector *MetricsCollector, config Dashboard) *DashboardServer {
	templates := template.Must(template.New("dashboard").Parse(dashboardHTML))

	server := &DashboardServer{
		collector: collector,
		templates: templates,
		clients:   make(map[chan []byte]bool),
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
	switch r.URL.Path {
	case "/":
		ds.serveDashboard(w, r)
	case "/api/metrics":
		ds.serveMetrics(w, r)
	case "/api/system":
		ds.serveSystemMetrics(w, r)
	case "/api/tenant":
		ds.serveTenantMetrics(w, r)
	case "/api/export":
		ds.serveExport(w, r)
	case "/ws":
		ds.handleWebSocket(w, r)
	default:
		http.NotFound(w, r)
	}
}

// serveDashboard serves the main dashboard page
func (ds *DashboardServer) serveDashboard(w http.ResponseWriter, r *http.Request) {
	data := struct {
		Title   string
		Version string
	}{
		Title:   "API Security Scanner Dashboard",
		Version: "4.0.0",
	}

	if err := ds.templates.ExecuteTemplate(w, "dashboard", data); err != nil {
		logging.Error("Failed to render dashboard", map[string]interface{}{
			"error": err.Error(),
		})
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}
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
	// WebSocket implementation would go here
	// This is a simplified version for demonstration
	w.Header().Set("Content-Type", "text/plain")
	w.Write([]byte("WebSocket not implemented yet"))
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