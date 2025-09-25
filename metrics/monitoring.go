package metrics

import (
	"context"
	"fmt"
	"net/http"
	"runtime"
	"sync"
	"time"

	"github.com/shirou/gopsutil/v3/cpu"
	"github.com/shirou/gopsutil/v3/net"

	"api-security-scanner/logging"
)

// Monitor represents a system monitor
type Monitor struct {
	interval     time.Duration
	running      bool
	stopChan     chan struct{}
	mu           sync.RWMutex
	cpuStats     []CPUStat
	memStats     []MemStat
	netStats     []NetStat
	lastCPUTime  time.Time
	lastCPUUsage float64
}

// CPUStat represents CPU usage statistics
type CPUStat struct {
	Timestamp time.Time
	Usage     float64
	Cores     int
}

// MemStat represents memory usage statistics
type MemStat struct {
	Timestamp time.Time
	Alloc     uint64
	TotalAlloc uint64
	Sys       uint64
	NumGC     uint32
	Goroutines int
}

// NetStat represents network usage statistics
type NetStat struct {
	Timestamp time.Time
	BytesSent uint64
	BytesRecv uint64
}

// ResourceMonitor monitors system resources
type ResourceMonitor struct {
	monitor     *Monitor
	collector   *MetricsCollector
	scanID      string
	tenantID    string
	ctx         context.Context
	cancel      context.CancelFunc
}

// NewMonitor creates a new system monitor
func NewMonitor(interval time.Duration) *Monitor {
	return &Monitor{
		interval:  interval,
		stopChan:  make(chan struct{}),
		cpuStats:  make([]CPUStat, 0),
		memStats:  make([]MemStat, 0),
		netStats:  make([]NetStat, 0),
	}
}

// Start starts the monitoring
func (m *Monitor) Start() {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.running {
		return
	}

	m.running = true
	m.lastCPUTime = time.Now()

	go m.run()

	logging.Info("System monitor started", map[string]interface{}{
		"interval": m.interval,
	})
}

// Stop stops the monitoring
func (m *Monitor) Stop() {
	m.mu.Lock()
	defer m.mu.Unlock()

	if !m.running {
		return
	}

	m.running = false
	close(m.stopChan)

	logging.Info("System monitor stopped", nil)
}

// run runs the monitoring loop
func (m *Monitor) run() {
	ticker := time.NewTicker(m.interval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			m.collectStats()
		case <-m.stopChan:
			return
		}
	}
}

// collectStats collects system statistics
func (m *Monitor) collectStats() {
	// Collect CPU stats
	cpuUsage := m.getCPUUsage()
	if cpuUsage >= 0 {
		m.mu.Lock()
		m.cpuStats = append(m.cpuStats, CPUStat{
			Timestamp: time.Now(),
			Usage:     cpuUsage,
			Cores:     runtime.NumCPU(),
		})
		m.mu.Unlock()
	}

	// Collect Go runtime stats
	var runtimeMemStats runtime.MemStats
	runtime.ReadMemStats(&runtimeMemStats)
	
	m.mu.Lock()
	m.memStats = append(m.memStats, MemStat{
		Timestamp:   time.Now(),
		Alloc:       runtimeMemStats.Alloc,
		TotalAlloc:  runtimeMemStats.TotalAlloc,
		Sys:         runtimeMemStats.Sys,
		NumGC:       runtimeMemStats.NumGC,
		Goroutines:  runtime.NumGoroutine(),
	})
	m.mu.Unlock()

	// Collect network stats using gopsutil
	netIO, err := net.IOCounters(false)
	if err != nil {
		logging.Error("Failed to get network stats", map[string]interface{}{"error": err})
	} else if len(netIO) > 0 {
		netStat := netIO[0] // Use first network interface
		m.mu.Lock()
		m.netStats = append(m.netStats, NetStat{
			Timestamp: time.Now(),
			BytesSent: netStat.BytesSent,
			BytesRecv: netStat.BytesRecv,
		})
		m.mu.Unlock()
	}
}

// getCPUUsage calculates CPU usage using gopsutil
func (m *Monitor) getCPUUsage() float64 {
	percentages, err := cpu.Percent(time.Second, false)
	if err != nil || len(percentages) == 0 {
		logging.Error("Failed to get CPU usage", map[string]interface{}{"error": err})
		return 0.0
	}
	
	return percentages[0]
}

// GetCPUStats returns CPU statistics
func (m *Monitor) GetCPUStats() []CPUStat {
	m.mu.RLock()
	defer m.mu.RUnlock()

	// Return a copy of the stats
	stats := make([]CPUStat, len(m.cpuStats))
	copy(stats, m.cpuStats)
	return stats
}

// GetMemStats returns memory statistics
func (m *Monitor) GetMemStats() []MemStat {
	m.mu.RLock()
	defer m.mu.RUnlock()

	// Return a copy of the stats
	stats := make([]MemStat, len(m.memStats))
	copy(stats, m.memStats)
	return stats
}

// GetNetStats returns network statistics
func (m *Monitor) GetNetStats() []NetStat {
	m.mu.RLock()
	defer m.mu.RUnlock()

	// Return a copy of the stats
	stats := make([]NetStat, len(m.netStats))
	copy(stats, m.netStats)
	return stats
}

// GetLatestStats returns the latest statistics
func (m *Monitor) GetLatestStats() (CPUStat, MemStat, NetStat) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	var latestCPU CPUStat
	var latestMem MemStat
	var latestNet NetStat

	if len(m.cpuStats) > 0 {
		latestCPU = m.cpuStats[len(m.cpuStats)-1]
	}

	if len(m.memStats) > 0 {
		latestMem = m.memStats[len(m.memStats)-1]
	}

	if len(m.netStats) > 0 {
		latestNet = m.netStats[len(m.netStats)-1]
	}

	return latestCPU, latestMem, latestNet
}

// NewResourceMonitor creates a new resource monitor for a specific scan
func NewResourceMonitor(scanID, tenantID string, collector *MetricsCollector) *ResourceMonitor {
	ctx, cancel := context.WithCancel(context.Background())

	return &ResourceMonitor{
		monitor:   NewMonitor(5 * time.Second),
		collector: collector,
		scanID:    scanID,
		tenantID:  tenantID,
		ctx:       ctx,
		cancel:    cancel,
	}
}

// Start starts the resource monitoring
func (rm *ResourceMonitor) Start() {
	rm.monitor.Start()

	// Record initial resource usage right away to ensure at least one sample exists
	rm.reportCurrentStats()

	// Start periodic reporting
	go rm.reportStats()

	logging.Info("Resource monitor started", map[string]interface{}{
		"scan_id":  rm.scanID,
		"tenant_id": rm.tenantID,
	})
}

// Stop stops the resource monitoring
func (rm *ResourceMonitor) Stop() {
	// Record final resource usage before stopping
	rm.reportCurrentStats()
	
	rm.monitor.Stop()
	rm.cancel()

	logging.Info("Resource monitor stopped", map[string]interface{}{
		"scan_id":  rm.scanID,
		"tenant_id": rm.tenantID,
	})
}

// reportStats periodically reports statistics to the metrics collector
func (rm *ResourceMonitor) reportStats() {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			rm.reportCurrentStats()
		case <-rm.ctx.Done():
			return
		}
	}
}

// reportCurrentStats reports current statistics to the metrics collector
func (rm *ResourceMonitor) reportCurrentStats() {
	// Get current system stats
	cpuPercent, err := cpu.Percent(time.Second, false)
	cpuUsage := 0.0
	if err == nil && len(cpuPercent) > 0 {
		cpuUsage = cpuPercent[0]
	}

	// Get network stats
	netIO, err := net.IOCounters(false)
	networkBytes := int64(0)
	if err == nil && len(netIO) > 0 {
		networkBytes = int64(netIO[0].BytesSent + netIO[0].BytesRecv)
	}

	// Get Go runtime stats
	var runtimeMemStats runtime.MemStats
	runtime.ReadMemStats(&runtimeMemStats)
	
	// Convert to MB for memory
	runtimeMemMB := float64(runtimeMemStats.Alloc) / 1024 / 1024

	// Record resource usage in metrics collector
	rm.collector.RecordResourceUsage(
		rm.scanID,
		cpuUsage,
		runtimeMemMB, // Use runtime memory instead of system memory for consistency with original design
		runtime.NumGoroutine(),
		networkBytes,
		0, // Disk I/O not implemented
	)
}

// HealthChecker checks the health of the system
type HealthChecker struct {
	checks map[string]HealthCheck
}

// HealthCheck represents a health check
type HealthCheck struct {
	Name        string
	Description string
	CheckFunc   func() HealthStatus
	Interval    time.Duration
	Timeout     time.Duration
}

// HealthStatus represents the status of a health check
type HealthStatus struct {
	Status  string            `json:"status"` // healthy, degraded, unhealthy
	Message string            `json:"message"`
	Details map[string]interface{} `json:"details"`
	Timestamp time.Time       `json:"timestamp"`
}

// NewHealthChecker creates a new health checker
func NewHealthChecker() *HealthChecker {
	return &HealthChecker{
		checks: make(map[string]HealthCheck),
	}
}

// AddCheck adds a health check
func (hc *HealthChecker) AddCheck(check HealthCheck) {
	hc.checks[check.Name] = check
}

// RunChecks runs all health checks
func (hc *HealthChecker) RunChecks() map[string]HealthStatus {
	results := make(map[string]HealthStatus)

	for name, check := range hc.checks {
		status := hc.runCheck(check)
		results[name] = status
	}

	return results
}

// runCheck runs a single health check
func (hc *HealthChecker) runCheck(check HealthCheck) HealthStatus {
	ctx, cancel := context.WithTimeout(context.Background(), check.Timeout)
	defer cancel()

	resultChan := make(chan HealthStatus)
	go func() {
		resultChan <- check.CheckFunc()
	}()

	select {
	case status := <-resultChan:
		return status
	case <-ctx.Done():
		return HealthStatus{
			Status:    "unhealthy",
			Message:   "Health check timed out",
			Details:   make(map[string]interface{}),
			Timestamp: time.Now(),
		}
	}
}

// GetOverallStatus returns the overall health status
func (hc *HealthChecker) GetOverallStatus() (string, map[string]HealthStatus) {
	results := hc.RunChecks()

	unhealthyCount := 0
	degradedCount := 0

	for _, status := range results {
		switch status.Status {
		case "unhealthy":
			unhealthyCount++
		case "degraded":
			degradedCount++
		}
	}

	overallStatus := "healthy"
	if unhealthyCount > 0 {
		overallStatus = "unhealthy"
	} else if degradedCount > 0 {
		overallStatus = "degraded"
	}

	return overallStatus, results
}

// CreateDefaultHealthChecks creates default health checks
func CreateDefaultHealthChecks() []HealthCheck {
	return []HealthCheck{
		{
			Name:        "memory_usage",
			Description: "Check memory usage",
			CheckFunc:   checkMemoryUsage,
			Interval:    30 * time.Second,
			Timeout:     5 * time.Second,
		},
		{
			Name:        "goroutine_count",
			Description: "Check goroutine count",
			CheckFunc:   checkGoroutineCount,
			Interval:    30 * time.Second,
			Timeout:     5 * time.Second,
		},
		{
			Name:        "network_connectivity",
			Description: "Check network connectivity",
			CheckFunc:   checkNetworkConnectivity,
			Interval:    60 * time.Second,
			Timeout:     10 * time.Second,
		},
	}
}

// Default health check functions
func checkMemoryUsage() HealthStatus {
	var memStats runtime.MemStats
	runtime.ReadMemStats(&memStats)

	memMB := float64(memStats.Alloc) / 1024 / 1024

	status := "healthy"
	message := fmt.Sprintf("Memory usage: %.2f MB", memMB)

	if memMB > 1000 { // 1GB threshold
		status = "degraded"
		message = fmt.Sprintf("High memory usage: %.2f MB", memMB)
	}

	if memMB > 2000 { // 2GB threshold
		status = "unhealthy"
		message = fmt.Sprintf("Critical memory usage: %.2f MB", memMB)
	}

	return HealthStatus{
		Status:    status,
		Message:   message,
		Details: map[string]interface{}{
			"memory_mb": memMB,
			"threshold_high": 1000.0,
			"threshold_critical": 2000.0,
		},
		Timestamp: time.Now(),
	}
}

func checkGoroutineCount() HealthStatus {
	count := runtime.NumGoroutine()

	status := "healthy"
	message := fmt.Sprintf("Goroutine count: %d", count)

	if count > 1000 {
		status = "degraded"
		message = fmt.Sprintf("High goroutine count: %d", count)
	}

	if count > 5000 {
		status = "unhealthy"
		message = fmt.Sprintf("Critical goroutine count: %d", count)
	}

	return HealthStatus{
		Status:    status,
		Message:   message,
		Details: map[string]interface{}{
			"goroutine_count": count,
			"threshold_high": 1000,
			"threshold_critical": 5000,
		},
		Timestamp: time.Now(),
	}
}

func checkNetworkConnectivity() HealthStatus {
	// Simple connectivity check to Google DNS
	client := http.Client{
		Timeout: 5 * time.Second,
	}

	resp, err := client.Get("http://8.8.8.8:53")
	if err != nil {
		return HealthStatus{
			Status:    "unhealthy",
			Message:   fmt.Sprintf("Network connectivity failed: %v", err),
			Details:   make(map[string]interface{}),
			Timestamp: time.Now(),
		}
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		return HealthStatus{
			Status:    "degraded",
			Message:   fmt.Sprintf("Network connectivity issue: HTTP %d", resp.StatusCode),
			Details: map[string]interface{}{
				"status_code": resp.StatusCode,
			},
			Timestamp: time.Now(),
		}
	}

	return HealthStatus{
		Status:    "healthy",
		Message:   "Network connectivity OK",
		Details: map[string]interface{}{
			"status_code": resp.StatusCode,
		},
		Timestamp: time.Now(),
	}
}

// AlertManager manages alerts based on metrics
type AlertManager struct {
	rules      []AlertRule
	notifiers  []AlertNotifier
	mu         sync.RWMutex
}

// AlertRule represents an alert rule
type AlertRule struct {
	Name        string
	Description string
	Metric      string
	Condition   string
	Threshold   float64
	Duration    time.Duration
	Enabled     bool
}

// Alert represents an alert
type Alert struct {
	RuleName    string    `json:"rule_name"`
	Message     string    `json:"message"`
	Severity    string    `json:"severity"`
	Value       float64   `json:"value"`
	Timestamp   time.Time `json:"timestamp"`
}

// AlertNotifier interface for alert notifications
type AlertNotifier interface {
	Notify(alert Alert) error
	Name() string
}

// NewAlertManager creates a new alert manager
func NewAlertManager() *AlertManager {
	return &AlertManager{
		rules:     make([]AlertRule, 0),
		notifiers: make([]AlertNotifier, 0),
	}
}

// AddRule adds an alert rule
func (am *AlertManager) AddRule(rule AlertRule) {
	am.mu.Lock()
	defer am.mu.Unlock()

	am.rules = append(am.rules, rule)
}

// AddNotifier adds an alert notifier
func (am *AlertManager) AddNotifier(notifier AlertNotifier) {
	am.mu.Lock()
	defer am.mu.Unlock()

	am.notifiers = append(am.notifiers, notifier)
}

// EvaluateRules evaluates all alert rules against current metrics
func (am *AlertManager) EvaluateRules(metrics *SystemMetrics) []Alert {
	am.mu.RLock()
	defer am.mu.RUnlock()

	var alerts []Alert

	for _, rule := range am.rules {
		if !rule.Enabled {
			continue
		}

		alert := am.evaluateRule(rule, metrics)
		if alert != nil {
			alerts = append(alerts, *alert)
		}
	}

	return alerts
}

// evaluateRule evaluates a single alert rule
func (am *AlertManager) evaluateRule(rule AlertRule, metrics *SystemMetrics) *Alert {
	var value float64
	var found bool

	// Find the metric value
	switch rule.Metric {
	case "cpu_usage":
		if metrics.ResourceUsage.CPUUsage > 0 {
			value = metrics.ResourceUsage.CPUUsage
			found = true
		}
	case "memory_usage":
		if metrics.ResourceUsage.MemoryUsage > 0 {
			value = metrics.ResourceUsage.MemoryUsage
			found = true
		}
	case "vulnerability_count":
		value = float64(metrics.Vulnerabilities.Total)
		found = true
	case "error_rate":
		value = metrics.Performance.ErrorRate
		found = true
	}

	if !found {
		return nil
	}

	// Check condition
	var triggered bool
	switch rule.Condition {
	case ">":
		triggered = value > rule.Threshold
	case ">=":
		triggered = value >= rule.Threshold
	case "<":
		triggered = value < rule.Threshold
	case "<=":
		triggered = value <= rule.Threshold
	case "==":
		triggered = value == rule.Threshold
	}

	if !triggered {
		return nil
	}

	// Determine severity
	severity := "medium"
	if rule.Threshold > 90 || rule.Metric == "vulnerability_count" {
		severity = "critical"
	} else if rule.Threshold > 70 {
		severity = "high"
	}

	return &Alert{
		RuleName:  rule.Name,
		Message:   fmt.Sprintf("%s: %.2f %s %.2f", rule.Description, value, rule.Condition, rule.Threshold),
		Severity:  severity,
		Value:     value,
		Timestamp: time.Now(),
	}
}

// SendAlerts sends alerts via all configured notifiers
func (am *AlertManager) SendAlerts(alerts []Alert) {
	am.mu.RLock()
	defer am.mu.RUnlock()

	for _, alert := range alerts {
		for _, notifier := range am.notifiers {
			if err := notifier.Notify(alert); err != nil {
				logging.Error("Failed to send alert", map[string]interface{}{
					"notifier": notifier.Name(),
					"alert":    alert.RuleName,
					"error":    err.Error(),
				})
			}
		}
	}
}