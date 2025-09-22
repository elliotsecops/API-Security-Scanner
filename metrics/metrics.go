package metrics

import (
	"encoding/json"
	"fmt"
	"sort"
	"sync"
	"time"

	"api-security-scanner/types"
	"api-security-scanner/logging"
)

// MetricType represents different types of metrics
type MetricType string

const (
	MetricTypeCounter   MetricType = "counter"
	MetricTypeGauge     MetricType = "gauge"
	MetricTypeHistogram MetricType = "histogram"
	MetricTypeTimer     MetricType = "timer"
)

// Metric represents a single metric
type Metric struct {
	Name      string                 `json:"name"`
	Type      MetricType             `json:"type"`
	Value     float64                `json:"value"`
	Timestamp time.Time              `json:"timestamp"`
	Tags      map[string]string      `json:"tags,omitempty"`
	Labels    map[string]interface{} `json:"labels,omitempty"`
}

// ScanMetrics represents metrics collected during a scan
type ScanMetrics struct {
	ScanID          string        `json:"scan_id"`
	TenantID        string        `json:"tenant_id"`
	StartTime       time.Time     `json:"start_time"`
	EndTime         time.Time     `json:"end_time,omitempty"`
	Duration        time.Duration `json:"duration,omitempty"`
	TotalEndpoints  int           `json:"total_endpoints"`
	EndpointsTested int           `json:"endpoints_tested"`
	Vulnerabilities VulnMetrics   `json:"vulnerabilities"`
	Performance     PerfMetrics   `json:"performance"`
	ResourceUsage   ResourceUsage `json:"resource_usage"`
}

// VulnMetrics represents vulnerability metrics
type VulnMetrics struct {
	Total           int                    `json:"total"`
	Critical        int                    `json:"critical"`
	High            int                    `json:"high"`
	Medium          int                    `json:"medium"`
	Low             int                    `json:"low"`
	ByType          map[string]int         `json:"by_type"`
	ByEndpoint      map[string]int         `json:"by_endpoint"`
	TrendData       []VulnTrendPoint       `json:"trend_data,omitempty"`
}

// PerfMetrics represents performance metrics
type PerfMetrics struct {
	AvgResponseTime    time.Duration       `json:"avg_response_time"`
	MaxResponseTime    time.Duration       `json:"max_response_time"`
	MinResponseTime    time.Duration       `json:"min_response_time"`
	Throughput         float64             `json:"throughput"` // requests per second
	ErrorRate          float64             `json:"error_rate"` // percentage
	ResponseTimeDist   []ResponseTimePoint `json:"response_time_distribution,omitempty"`
}

// ResourceUsage represents system resource usage
type ResourceUsage struct {
	CPUUsage      float64 `json:"cpu_usage"`      // percentage
	MemoryUsage   float64 `json:"memory_usage"`   // MB
	Goroutines    int     `json:"goroutines"`
	NetworkBytes  int64   `json:"network_bytes"`
	DiskIO        int64   `json:"disk_io"`
}

// VulnTrendPoint represents a point in vulnerability trend data
type VulnTrendPoint struct {
	Timestamp time.Time `json:"timestamp"`
	Count     int       `json:"count"`
	Severity  string    `json:"severity"`
}

// ResponseTimePoint represents a point in response time distribution
type ResponseTimePoint struct {
	Bucket    string        `json:"bucket"`
	Count     int           `json:"count"`
	Duration  time.Duration `json:"duration"`
	Percentile float64       `json:"percentile"`
}

// Dashboard represents monitoring dashboard configuration
type Dashboard struct {
	Enabled       bool                   `yaml:"enabled"`
	Port          int                    `yaml:"port"`
	UpdateInterval time.Duration          `yaml:"update_interval"`
	RetentionDays int                    `yaml:"retention_days"`
	Charts        []DashboardChart       `yaml:"charts"`
	Alerts        []AlertConfig          `yaml:"alerts"`
}

// DashboardChart represents a chart configuration
type DashboardChart struct {
	Title       string   `yaml:"title"`
	Type       string   `yaml:"type"` // line, bar, pie, gauge
	Metrics    []string `yaml:"metrics"`
	TimeRange  string   `yaml:"time_range"` // 1h, 24h, 7d, 30d
	Refresh    int      `yaml:"refresh"`    // seconds
}

// AlertConfig represents alert configuration
type AlertConfig struct {
	Name          string  `yaml:"name"`
	Metric        string  `yaml:"metric"`
	Condition     string  `yaml:"condition"` // >, <, >=, <=, ==
	Threshold     float64 `yaml:"threshold"`
	Duration      string  `yaml:"duration"`  // how long condition must be met
	Notifications []string `yaml:"notifications"` // email, slack, webhook
}

// MetricsCollector collects and manages metrics
type MetricsCollector struct {
	mutex        sync.RWMutex
	metrics      map[string][]Metric
	scanMetrics  map[string]*ScanMetrics
	history      []ScanMetrics
	config       Dashboard
	collectors   []MetricCollector
}

// MetricCollector interface for collecting different types of metrics
type MetricCollector interface {
	Collect() ([]Metric, error)
	Name() string
}

// NewMetricsCollector creates a new metrics collector
func NewMetricsCollector(config Dashboard) *MetricsCollector {
	return &MetricsCollector{
		metrics:     make(map[string][]Metric),
		scanMetrics: make(map[string]*ScanMetrics),
		history:     make([]ScanMetrics, 0),
		config:      config,
		collectors:  []MetricCollector{},
	}
}

// StartScan starts collecting metrics for a scan
func (mc *MetricsCollector) StartScan(scanID, tenantID string, endpointCount int) {
	mc.mutex.Lock()
	defer mc.mutex.Unlock()

	scanMetrics := &ScanMetrics{
		ScanID:         scanID,
		TenantID:       tenantID,
		StartTime:      time.Now(),
		TotalEndpoints: endpointCount,
		Vulnerabilities: VulnMetrics{
			ByType:     make(map[string]int),
			ByEndpoint: make(map[string]int),
		},
	}

	mc.scanMetrics[scanID] = scanMetrics
}

// RecordEndpointTest records metrics for an endpoint test
func (mc *MetricsCollector) RecordEndpointTest(scanID, endpointURL string, responseTime time.Duration, results []types.TestResult) {
	mc.mutex.Lock()
	defer mc.mutex.Unlock()

	scan, exists := mc.scanMetrics[scanID]
	if !exists {
		return
	}

	scan.EndpointsTested++

	// Update response time metrics
	if scan.Performance.MaxResponseTime < responseTime {
		scan.Performance.MaxResponseTime = responseTime
	}
	if scan.Performance.MinResponseTime == 0 || scan.Performance.MinResponseTime > responseTime {
		scan.Performance.MinResponseTime = responseTime
	}

	// Calculate average response time
	if scan.EndpointsTested > 0 {
		totalTime := scan.Performance.AvgResponseTime * time.Duration(scan.EndpointsTested-1)
		scan.Performance.AvgResponseTime = (totalTime + responseTime) / time.Duration(scan.EndpointsTested)
	}

	// Count vulnerabilities
	for _, result := range results {
		if !result.Passed {
			scan.Vulnerabilities.Total++
			scan.Vulnerabilities.ByEndpoint[endpointURL]++

			switch result.TestName {
			case "Injection Test", "NoSQL Injection Test":
				scan.Vulnerabilities.Critical++
			case "XSS Test", "Auth Bypass Test":
				scan.Vulnerabilities.High++
			case "Parameter Tampering Test":
				scan.Vulnerabilities.Medium++
			default:
				scan.Vulnerabilities.Low++
			}

			scan.Vulnerabilities.ByType[result.TestName]++
		}
	}
}

// RecordResourceUsage records system resource usage
func (mc *MetricsCollector) RecordResourceUsage(scanID string, cpu, memory float64, goroutines int, network, disk int64) {
	mc.mutex.Lock()
	defer mc.mutex.Unlock()

	scan, exists := mc.scanMetrics[scanID]
	if !exists {
		return
	}

	scan.ResourceUsage = ResourceUsage{
		CPUUsage:     cpu,
		MemoryUsage:  memory,
		Goroutines:   goroutines,
		NetworkBytes: network,
		DiskIO:       disk,
	}
}

// EndScan completes scan metrics collection
func (mc *MetricsCollector) EndScan(scanID string) {
	mc.mutex.Lock()
	defer mc.mutex.Unlock()

	scan, exists := mc.scanMetrics[scanID]
	if !exists {
		return
	}

	scan.EndTime = time.Now()
	scan.Duration = scan.EndTime.Sub(scan.StartTime)

	// Calculate throughput
	if scan.Duration > 0 {
		scan.Performance.Throughput = float64(scan.EndpointsTested) / scan.Duration.Seconds()
	}

	// Calculate error rate
	if scan.EndpointsTested > 0 {
		scan.Performance.ErrorRate = float64(scan.Vulnerabilities.Total) / float64(scan.EndpointsTested) * 100
	}

	// Move to history
	mc.history = append(mc.history, *scan)
	delete(mc.scanMetrics, scanID)

	// Maintain history size
	maxHistory := mc.config.RetentionDays * 24 // hourly scans
	if len(mc.history) > maxHistory {
		mc.history = mc.history[1:]
	}

	logging.Info("Scan metrics collected", map[string]interface{}{
		"scan_id":          scanID,
		"duration":        scan.Duration,
		"endpoints_tested": scan.EndpointsTested,
		"vulnerabilities":  scan.Vulnerabilities.Total,
	})
}

// GetScanMetrics retrieves metrics for a specific scan
func (mc *MetricsCollector) GetScanMetrics(scanID string) (*ScanMetrics, error) {
	mc.mutex.RLock()
	defer mc.mutex.RUnlock()

	for _, scan := range mc.history {
		if scan.ScanID == scanID {
			return &scan, nil
		}
	}

	return nil, fmt.Errorf("scan not found: %s", scanID)
}

// GetTenantMetrics retrieves metrics for a specific tenant
func (mc *MetricsCollector) GetTenantMetrics(tenantID string, timeRange time.Duration) *TenantMetrics {
	mc.mutex.RLock()
	defer mc.mutex.RUnlock()

	cutoff := time.Now().Add(-timeRange)
	var tenantScans []ScanMetrics

	for _, scan := range mc.history {
		if scan.TenantID == tenantID && scan.StartTime.After(cutoff) {
			tenantScans = append(tenantScans, scan)
		}
	}

	return mc.aggregateTenantMetrics(tenantScans)
}

// GetSystemMetrics retrieves system-wide metrics
func (mc *MetricsCollector) GetSystemMetrics(timeRange time.Duration) *SystemMetrics {
	mc.mutex.RLock()
	defer mc.mutex.RUnlock()

	cutoff := time.Now().Add(-timeRange)
	var recentScans []ScanMetrics

	for _, scan := range mc.history {
		if scan.StartTime.After(cutoff) {
			recentScans = append(recentScans, scan)
		}
	}

	return mc.aggregateSystemMetrics(recentScans)
}

// TenantMetrics represents aggregated metrics for a tenant
type TenantMetrics struct {
	TenantID         string            `json:"tenant_id"`
	TimeRange        time.Duration     `json:"time_range"`
	TotalScans       int               `json:"total_scans"`
	TotalEndpoints   int               `json:"total_endpoints"`
	Vulnerabilities  VulnMetrics       `json:"vulnerabilities"`
	Performance      PerfMetrics       `json:"performance"`
	TopVulnerableEndpoints []EndpointScore `json:"top_vulnerable_endpoints"`
	TrendData        []VulnTrendPoint  `json:"trend_data"`
}

// SystemMetrics represents system-wide metrics
type SystemMetrics struct {
	ActiveTenants    int               `json:"active_tenants"`
	TotalScans       int               `json:"total_scans"`
	TotalEndpoints   int               `json:"total_endpoints"`
	Vulnerabilities  VulnMetrics       `json:"vulnerabilities"`
	Performance      PerfMetrics       `json:"performance"`
	ResourceUsage    ResourceUsage     `json:"resource_usage"`
	TenantMetrics    map[string]*TenantMetrics `json:"tenant_metrics"`
}

// EndpointScore represents endpoint vulnerability scoring
type EndpointScore struct {
	URL          string  `json:"url"`
	VulnerabilityCount int  `json:"vulnerability_count"`
	Score        float64 `json:"score"`
	LastScan     time.Time `json:"last_scan"`
}

// Helper methods for aggregation
func (mc *MetricsCollector) aggregateTenantMetrics(scans []ScanMetrics) *TenantMetrics {
	if len(scans) == 0 {
		return nil
	}

	tenantID := scans[0].TenantID
	metrics := &TenantMetrics{
		TenantID:    tenantID,
		TimeRange:   time.Since(scans[len(scans)-1].StartTime),
		TotalScans:  len(scans),
		Vulnerabilities: VulnMetrics{
			ByType:     make(map[string]int),
			ByEndpoint: make(map[string]int),
		},
	}

	endpointScores := make(map[string]*EndpointScore)

	for _, scan := range scans {
		metrics.TotalEndpoints += scan.TotalEndpoints

		// Aggregate vulnerabilities
		metrics.Vulnerabilities.Total += scan.Vulnerabilities.Total
		metrics.Vulnerabilities.Critical += scan.Vulnerabilities.Critical
		metrics.Vulnerabilities.High += scan.Vulnerabilities.High
		metrics.Vulnerabilities.Medium += scan.Vulnerabilities.Medium
		metrics.Vulnerabilities.Low += scan.Vulnerabilities.Low

		// Aggregate by type
		for vulnType, count := range scan.Vulnerabilities.ByType {
			metrics.Vulnerabilities.ByType[vulnType] += count
		}

		// Aggregate by endpoint
		for endpoint, count := range scan.Vulnerabilities.ByEndpoint {
			metrics.Vulnerabilities.ByEndpoint[endpoint] += count

			// Track endpoint scores
			if _, exists := endpointScores[endpoint]; !exists {
				endpointScores[endpoint] = &EndpointScore{
					URL: endpoint,
					LastScan: scan.StartTime,
				}
			}
			endpointScores[endpoint].VulnerabilityCount += count
		}
	}

	// Calculate endpoint scores
	for _, score := range endpointScores {
		score.Score = float64(score.VulnerabilityCount) / float64(len(scans))
	}

	// Sort endpoints by vulnerability count
	for _, score := range endpointScores {
		metrics.TopVulnerableEndpoints = append(metrics.TopVulnerableEndpoints, *score)
	}
	sort.Slice(metrics.TopVulnerableEndpoints, func(i, j int) bool {
		return metrics.TopVulnerableEndpoints[i].VulnerabilityCount >
			   metrics.TopVulnerableEndpoints[j].VulnerabilityCount
	})

	// Limit to top 10
	if len(metrics.TopVulnerableEndpoints) > 10 {
		metrics.TopVulnerableEndpoints = metrics.TopVulnerableEndpoints[:10]
	}

	return metrics
}

func (mc *MetricsCollector) aggregateSystemMetrics(scans []ScanMetrics) *SystemMetrics {
	if len(scans) == 0 {
		return nil
	}

	tenantMap := make(map[string]bool)
	tenantMetrics := make(map[string]*TenantMetrics)

	for _, scan := range scans {
		tenantMap[scan.TenantID] = true
	}

	metrics := &SystemMetrics{
		ActiveTenants: len(tenantMap),
		TotalScans:    len(scans),
		Vulnerabilities: VulnMetrics{
			ByType:     make(map[string]int),
			ByEndpoint: make(map[string]int),
		},
		TenantMetrics: tenantMetrics,
	}

	// Aggregate metrics across all scans
	for _, scan := range scans {
		metrics.TotalEndpoints += scan.TotalEndpoints

		// Aggregate vulnerabilities
		metrics.Vulnerabilities.Total += scan.Vulnerabilities.Total
		metrics.Vulnerabilities.Critical += scan.Vulnerabilities.Critical
		metrics.Vulnerabilities.High += scan.Vulnerabilities.High
		metrics.Vulnerabilities.Medium += scan.Vulnerabilities.Medium
		metrics.Vulnerabilities.Low += scan.Vulnerabilities.Low

		// Aggregate by type
		for vulnType, count := range scan.Vulnerabilities.ByType {
			metrics.Vulnerabilities.ByType[vulnType] += count
		}
	}

	// Aggregate tenant metrics
	for tenantID := range tenantMap {
		var tenantScans []ScanMetrics
		for _, scan := range scans {
			if scan.TenantID == tenantID {
				tenantScans = append(tenantScans, scan)
			}
		}
		tenantMetrics[tenantID] = mc.aggregateTenantMetrics(tenantScans)
	}

	return metrics
}

// ExportMetrics exports metrics in various formats
func (mc *MetricsCollector) ExportMetrics(format string) ([]byte, error) {
	mc.mutex.RLock()
	defer mc.mutex.RUnlock()

	switch format {
	case "json":
		return mc.exportJSON()
	case "prometheus":
		return mc.exportPrometheus()
	default:
		return nil, fmt.Errorf("unsupported export format: %s", format)
	}
}

func (mc *MetricsCollector) exportJSON() ([]byte, error) {
	data := struct {
		ScanMetrics  map[string]*ScanMetrics `json:"scan_metrics"`
		History      []ScanMetrics           `json:"history"`
		System       *SystemMetrics          `json:"system"`
	}{
		ScanMetrics: mc.scanMetrics,
		History:     mc.history,
		System:      mc.GetSystemMetrics(24 * time.Hour),
	}

	return json.MarshalIndent(data, "", "  ")
}

func (mc *MetricsCollector) exportPrometheus() ([]byte, error) {
	var prometheusData string

	// Export system metrics
	system := mc.GetSystemMetrics(24 * time.Hour)
	if system != nil {
		prometheusData += fmt.Sprintf("# HELP api_scanner_total_scans Total number of scans\n")
		prometheusData += fmt.Sprintf("# TYPE api_scanner_total_scans counter\n")
		prometheusData += fmt.Sprintf("api_scanner_total_scans %d\n", system.TotalScans)

		prometheusData += fmt.Sprintf("# HELP api_scanner_total_vulnerabilities Total number of vulnerabilities\n")
		prometheusData += fmt.Sprintf("# TYPE api_scanner_total_vulnerabilities counter\n")
		prometheusData += fmt.Sprintf("api_scanner_total_vulnerabilities %d\n", system.Vulnerabilities.Total)

		prometheusData += fmt.Sprintf("# HELP api_scanner_critical_vulnerabilities Number of critical vulnerabilities\n")
		prometheusData += fmt.Sprintf("# TYPE api_scanner_critical_vulnerabilities counter\n")
		prometheusData += fmt.Sprintf("api_scanner_critical_vulnerabilities %d\n", system.Vulnerabilities.Critical)
	}

	return []byte(prometheusData), nil
}

// CleanupOldData removes old metrics data
func (mc *MetricsCollector) CleanupOldData() {
	mc.mutex.Lock()
	defer mc.mutex.Unlock()

	if mc.config.RetentionDays <= 0 {
		return
	}

	cutoff := time.Now().AddDate(0, 0, -mc.config.RetentionDays)
	var newHistory []ScanMetrics

	for _, scan := range mc.history {
		if scan.StartTime.After(cutoff) {
			newHistory = append(newHistory, scan)
		}
	}

	mc.history = newHistory
	logging.Info("Cleaned up old metrics data", map[string]interface{}{
		"removed_count": len(mc.history) - len(newHistory),
		"retention_days": mc.config.RetentionDays,
	})
}