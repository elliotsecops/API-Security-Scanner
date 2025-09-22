package siem

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log/syslog"
	"net/http"
	"strings"
	"time"

	"api-security-scanner/logging"
	"api-security-scanner/tenant"
	"api-security-scanner/types"
)

// SIEMEvent represents a security event for SIEM integration
type SIEMEvent struct {
	Timestamp      time.Time              `json:"timestamp"`
	EventType      string                 `json:"event_type"`
	Severity       string                 `json:"severity"`
	TenantID       string                 `json:"tenant_id"`
	SourceIP       string                 `json:"source_ip"`
	TargetURL      string                 `json:"target_url"`
	Method         string                 `json:"method"`
	Vulnerability  string                 `json:"vulnerability"`
	Description    string                 `json:"description"`
	RawData        map[string]interface{} `json:"raw_data"`
	Tags           []string               `json:"tags"`
}

// SIEMClient represents a SIEM integration client
type SIEMClient struct {
	config   tenant.SIEMConfig
	client   *http.Client
	syslogWriter *syslog.Writer
}

// NewSIEMClient creates a new SIEM client
func NewSIEMClient(config tenant.SIEMConfig) (*SIEMClient, error) {
	if !config.Enabled {
		return nil, fmt.Errorf("SIEM integration is not enabled")
	}

	client := &SIEMClient{
		config: config,
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
	}

	// Initialize syslog writer if needed
	if config.Type == tenant.SIEMTypeSyslog {
		writer, err := syslog.New(syslog.LOG_INFO|syslog.LOG_DAEMON, "api-security-scanner")
		if err != nil {
			return nil, fmt.Errorf("failed to create syslog writer: %v", err)
		}
		client.syslogWriter = writer
	}

	return client, nil
}

// SendEvent sends a security event to SIEM
func (c *SIEMClient) SendEvent(event *SIEMEvent) error {
	if !c.config.Enabled {
		return nil
	}

	switch c.config.Type {
	case tenant.SIEMTypeSplunk:
		return c.sendToSplunk(event)
	case tenant.SIEMTypeELK:
		return c.sendToELK(event)
	case tenant.SIEMTypeQRadar:
		return c.sendToQRadar(event)
	case tenant.SIEMTypeArcSight:
		return c.sendToArcSight(event)
	case tenant.SIEMTypeSyslog:
		return c.sendToSyslog(event)
	default:
		return fmt.Errorf("unsupported SIEM type: %s", c.config.Type)
	}
}

// SendBatchEvents sends multiple events to SIEM
func (c *SIEMClient) SendBatchEvents(events []*SIEMEvent) error {
	if !c.config.Enabled {
		return nil
	}

	for _, event := range events {
		if err := c.SendEvent(event); err != nil {
			logging.Error("Failed to send SIEM event", map[string]interface{}{
				"event_type": event.EventType,
				"error":      err.Error(),
			})
			// Continue sending other events
		}
	}

	return nil
}

// Close closes the SIEM client
func (c *SIEMClient) Close() error {
	if c.syslogWriter != nil {
		return c.syslogWriter.Close()
	}
	return nil
}

// sendToSplunk sends event to Splunk
func (c *SIEMClient) sendToSplunk(event *SIEMEvent) error {
	payload := map[string]interface{}{
		"time":       event.Timestamp.UTC().Format(time.RFC3339),
		"host":       c.getSourceIP(event),
		"source":     "api-security-scanner",
		"sourcetype": "api:security:scan",
		"event":      event,
	}

	data, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal Splunk event: %v", err)
	}

	req, err := http.NewRequest("POST", c.config.EndpointURL, bytes.NewBuffer(data))
	if err != nil {
		return fmt.Errorf("failed to create Splunk request: %v", err)
	}

	// Set Splunk headers
	req.Header.Set("Authorization", "Bearer "+c.config.AuthToken)
	req.Header.Set("Content-Type", "application/json")

	for key, value := range c.config.Config {
		req.Header.Set(key, value)
	}

	resp, err := c.client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send to Splunk: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		return fmt.Errorf("Splunk returned status code: %d", resp.StatusCode)
	}

	return nil
}

// sendToELK sends event to ELK (Elasticsearch, Logstash, Kibana)
func (c *SIEMClient) sendToELK(event *SIEMEvent) error {
	data, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("failed to marshal ELK event: %v", err)
	}

	req, err := http.NewRequest("POST", c.config.EndpointURL, bytes.NewBuffer(data))
	if err != nil {
		return fmt.Errorf("failed to create ELK request: %v", err)
	}

	// Set ELK headers
	if c.config.AuthToken != "" {
		req.Header.Set("Authorization", "Bearer "+c.config.AuthToken)
	}
	req.Header.Set("Content-Type", "application/json")

	for key, value := range c.config.Config {
		req.Header.Set(key, value)
	}

	resp, err := c.client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send to ELK: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		return fmt.Errorf("ELK returned status code: %d", resp.StatusCode)
	}

	return nil
}

// sendToQRadar sends event to IBM QRadar
func (c *SIEMClient) sendToQRadar(event *SIEMEvent) error {
	// Convert to Common Event Format (CEF)
	cefMessage := c.formatCEF(event)

	req, err := http.NewRequest("POST", c.config.EndpointURL, strings.NewReader(cefMessage))
	if err != nil {
		return fmt.Errorf("failed to create QRadar request: %v", err)
	}

	// Set QRadar headers
	if c.config.AuthToken != "" {
		req.Header.Set("Authorization", "Bearer "+c.config.AuthToken)
	}
	req.Header.Set("Content-Type", "text/plain")

	for key, value := range c.config.Config {
		req.Header.Set(key, value)
	}

	resp, err := c.client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send to QRadar: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		return fmt.Errorf("QRadar returned status code: %d", resp.StatusCode)
	}

	return nil
}

// sendToArcSight sends event to HP ArcSight
func (c *SIEMClient) sendToArcSight(event *SIEMEvent) error {
	// Convert to LEEF format
	leefMessage := c.formatLEEF(event)

	req, err := http.NewRequest("POST", c.config.EndpointURL, strings.NewReader(leefMessage))
	if err != nil {
		return fmt.Errorf("failed to create ArcSight request: %v", err)
	}

	// Set ArcSight headers
	if c.config.AuthToken != "" {
		req.Header.Set("Authorization", "Bearer "+c.config.AuthToken)
	}
	req.Header.Set("Content-Type", "text/plain")

	for key, value := range c.config.Config {
		req.Header.Set(key, value)
	}

	resp, err := c.client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send to ArcSight: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		return fmt.Errorf("ArcSight returned status code: %d", resp.StatusCode)
	}

	return nil
}

// sendToSyslog sends event to syslog
func (c *SIEMClient) sendToSyslog(event *SIEMEvent) error {
	if c.syslogWriter == nil {
		return fmt.Errorf("syslog writer not initialized")
	}

	message := fmt.Sprintf("API Security Scan: %s - %s %s - %s - %s",
		event.TenantID,
		event.Method,
		event.TargetURL,
		event.Vulnerability,
		event.Description,
	)

	switch event.Severity {
	case "critical":
		return c.syslogWriter.Crit(message)
	case "high":
		return c.syslogWriter.Err(message)
	case "medium":
		return c.syslogWriter.Warning(message)
	case "low":
		return c.syslogWriter.Info(message)
	default:
		return c.syslogWriter.Info(message)
	}
}

// formatCEF formats event as Common Event Format (CEF)
func (c *SIEMClient) formatCEF(event *SIEMEvent) string {
	cefVersion := "0"
	DeviceVendor := "API-Security-Scanner"
	DeviceProduct := "Security-Scanner"
	DeviceVersion := "1.0"
	DeviceEventClassID := event.EventType
	Name := event.Vulnerability
	Severity := mapSeverityToCEF(event.Severity)

	// CEF:Version|Device Vendor|Device Product|Device Version|Device Event Class ID|Name|Severity|Extension
	cefHeader := fmt.Sprintf("CEF:%s|%s|%s|%s|%s|%s|%s|",
		cefVersion, DeviceVendor, DeviceProduct, DeviceVersion,
		DeviceEventClassID, Name, Severity)

	// Extension fields
	extensions := fmt.Sprintf("cs1=%s cs1Label=TenantID cs2=%s cs2Label=SourceIP dhost=%s requestMethod=%s msg=%s",
		event.TenantID, c.getSourceIP(event), event.TargetURL, event.Method, event.Description)

	return cefHeader + extensions
}

// formatLEEF formats event as LEEF format
func (c *SIEMClient) formatLEEF(event *SIEMEvent) string {
	// LEEF:Version|Vendor|Product|Device Version|Device Event Class ID|Name|Severity|Extensions
	leefHeader := fmt.Sprintf("LEEF:1.0|API-Security-Scanner|Security-Scanner|1.0|%s|%s|%s|",
		event.EventType, event.Vulnerability, mapSeverityToLEEF(event.Severity))

	// Extensions
	extensions := fmt.Sprintf("tenantID=%s src=%s dst=%s requestMethod=%s msg=%s",
		event.TenantID, c.getSourceIP(event), event.TargetURL, event.Method, event.Description)

	return leefHeader + extensions
}

// getSourceIP extracts source IP from event or returns default
func (c *SIEMClient) getSourceIP(event *SIEMEvent) string {
	if event.SourceIP != "" {
		return event.SourceIP
	}
	if ip, exists := event.RawData["source_ip"]; exists {
		if ipStr, ok := ip.(string); ok {
			return ipStr
		}
	}
	return "127.0.0.1" // Default fallback
}

// mapSeverityToCEF maps severity to CEF format
func mapSeverityToCEF(severity string) string {
	switch strings.ToLower(severity) {
	case "critical":
		return "10"
	case "high":
		return "8"
	case "medium":
		return "5"
	case "low":
		return "3"
	default:
		return "1"
	}
}

// mapSeverityToLEEF maps severity to LEEF format
func mapSeverityToLEEF(severity string) string {
	switch strings.ToLower(severity) {
	case "critical":
		return "10"
	case "high":
		return "8"
	case "medium":
		return "6"
	case "low":
		return "4"
	default:
		return "1"
	}
}

// CreateVulnerabilityEvent creates a SIEM event for vulnerability detection
func CreateVulnerabilityEvent(tenantID, vulnerability, description, targetURL, method string, severity string) *SIEMEvent {
	return &SIEMEvent{
		Timestamp:     time.Now(),
		EventType:     "vulnerability_detected",
		Severity:      severity,
		TenantID:      tenantID,
		TargetURL:     targetURL,
		Method:        method,
		Vulnerability: vulnerability,
		Description:   description,
		RawData:       make(map[string]interface{}),
		Tags:          []string{"vulnerability", "api-security", "automated-scan"},
	}
}

// CreateScanEvent creates a SIEM event for scan operations
func CreateScanEvent(tenantID, scanType, targetURL string, endpointCount int) *SIEMEvent {
	return &SIEMEvent{
		Timestamp:    time.Now(),
		EventType:    "scan_completed",
		Severity:     "low",
		TenantID:     tenantID,
		TargetURL:    targetURL,
		Method:       "SCAN",
		Vulnerability: scanType,
		Description:  fmt.Sprintf("Completed %s scan for %d endpoints", scanType, endpointCount),
		RawData:      make(map[string]interface{}),
		Tags:         []string{"scan", "api-security", "completed"},
	}
}

// CreateAuthEvent creates a SIEM event for authentication events
func CreateAuthEvent(tenantID, authType, result, sourceIP string) *SIEMEvent {
	return &SIEMEvent{
		Timestamp:    time.Now(),
		EventType:    "authentication",
		Severity:     func() string {
			if result == "success" {
				return "low"
			}
			return "medium"
		}(),
		TenantID:     tenantID,
		SourceIP:     sourceIP,
		Vulnerability: authType,
		Description:  fmt.Sprintf("Authentication %s for %s", result, authType),
		RawData:      make(map[string]interface{}),
		Tags:         []string{"authentication", "security"},
	}
}

// ConvertScanResultsToEvents converts scan results to SIEM events
func ConvertScanResultsToEvents(tenantID string, results []types.EndpointResult) []*SIEMEvent {
	var events []*SIEMEvent

	for _, result := range results {
		for _, testResult := range result.Results {
			if !testResult.Passed {
				severity := determineSeverity(testResult.TestName)
				event := CreateVulnerabilityEvent(
					tenantID,
					testResult.TestName,
					testResult.Message,
					result.URL,
					"GET", // This could be enhanced to capture the actual method
					severity,
				)
				events = append(events, event)
			}
		}
	}

	return events
}

// determineSeverity determines severity based on test name
func determineSeverity(testName string) string {
	switch testName {
	case "Injection Test", "NoSQL Injection Test":
		return "critical"
	case "XSS Test", "Auth Bypass Test":
		return "high"
	case "Parameter Tampering Test":
		return "medium"
	case "Header Security Test", "Auth Test", "HTTP Method Test":
		return "low"
	default:
		return "medium"
	}
}