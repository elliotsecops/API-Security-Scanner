package siem

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log/syslog"
	"net/http"
	"strings"
	"sync"
	"time"

	"api-security-scanner/logging"
	"api-security-scanner/tenant"
	"api-security-scanner/types"
)

// SIEMEvent represents a security event for SIEM integration
type SIEMEvent struct {
	Timestamp     time.Time              `json:"timestamp"`
	EventType     string                 `json:"event_type"`
	Severity      string                 `json:"severity"`
	TenantID      string                 `json:"tenant_id"`
	SourceIP      string                 `json:"source_ip"`
	TargetURL     string                 `json:"target_url"`
	Method        string                 `json:"method"`
	Vulnerability string                 `json:"vulnerability"`
	Description   string                 `json:"description"`
	RawData       map[string]interface{} `json:"raw_data"`
	Tags          []string               `json:"tags"`
}

// SIEMClient represents a SIEM integration client
type SIEMClient struct {
	config       tenant.SIEMConfig
	client       *http.Client
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
			Transport: &http.Transport{
				MaxIdleConns:        20,
				MaxIdleConnsPerHost: 5,
				IdleConnTimeout:     90 * time.Second,
			},
		},
	}

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

// SendBatchEvents sends multiple events to SIEM using concurrent workers
func (c *SIEMClient) SendBatchEvents(events []*SIEMEvent) error {
	if !c.config.Enabled {
		return nil
	}

	var wg sync.WaitGroup
	errChan := make(chan error, len(events))
	sem := make(chan struct{}, 5) // Limit concurrent sends

	for _, event := range events {
		wg.Add(1)
		go func(e *SIEMEvent) {
			defer wg.Done()
			sem <- struct{}{}
			defer func() { <-sem }()
			if err := c.SendEvent(e); err != nil {
				errChan <- err
				logging.Error("Failed to send SIEM event", map[string]interface{}{
					"event_type": e.EventType,
					"error":      err.Error(),
				})
			}
		}(event)
	}

	wg.Wait()
	close(errChan)

	var errs []string
	for err := range errChan {
		errs = append(errs, err.Error())
	}
	if len(errs) > 0 {
		return fmt.Errorf("some events failed to send: %s", strings.Join(errs, "; "))
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

// drainBody reads and discards response body for connection reuse
func drainBody(resp *http.Response) {
	if resp == nil || resp.Body == nil {
		return
	}
	io.Copy(io.Discard, io.LimitReader(resp.Body, 1024))
	resp.Body.Close()
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

	req.Header.Set("Authorization", "Bearer "+c.config.AuthToken)
	req.Header.Set("Content-Type", "application/json")

	for key, value := range c.config.Config {
		req.Header.Set(key, value)
	}

	resp, err := c.client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send to Splunk: %v", err)
	}
	drainBody(resp)

	if resp.StatusCode >= 400 {
		return fmt.Errorf("Splunk returned status code: %d", resp.StatusCode)
	}

	return nil
}

// sendToELK sends event to ELK
func (c *SIEMClient) sendToELK(event *SIEMEvent) error {
	data, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("failed to marshal ELK event: %v", err)
	}

	req, err := http.NewRequest("POST", c.config.EndpointURL, bytes.NewBuffer(data))
	if err != nil {
		return fmt.Errorf("failed to create ELK request: %v", err)
	}

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
	drainBody(resp)

	if resp.StatusCode >= 400 {
		return fmt.Errorf("ELK returned status code: %d", resp.StatusCode)
	}

	return nil
}

// sendToQRadar sends event to IBM QRadar
func (c *SIEMClient) sendToQRadar(event *SIEMEvent) error {
	cefMessage := c.formatCEF(event)

	req, err := http.NewRequest("POST", c.config.EndpointURL, strings.NewReader(cefMessage))
	if err != nil {
		return fmt.Errorf("failed to create QRadar request: %v", err)
	}

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
	drainBody(resp)

	if resp.StatusCode >= 400 {
		return fmt.Errorf("QRadar returned status code: %d", resp.StatusCode)
	}

	return nil
}

// sendToArcSight sends event to HP ArcSight
func (c *SIEMClient) sendToArcSight(event *SIEMEvent) error {
	leefMessage := c.formatLEEF(event)

	req, err := http.NewRequest("POST", c.config.EndpointURL, strings.NewReader(leefMessage))
	if err != nil {
		return fmt.Errorf("failed to create ArcSight request: %v", err)
	}

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
	drainBody(resp)

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

// formatCEF formats event as Common Event Format
func (c *SIEMClient) formatCEF(event *SIEMEvent) string {
	cefHeader := fmt.Sprintf("CEF:%s|%s|%s|%s|%s|%s|%s|",
		"0", "API-Security-Scanner", "Security-Scanner", "1.0",
		event.EventType, event.Vulnerability, mapSeverityToCEF(event.Severity))

	extensions := fmt.Sprintf("cs1=%s cs1Label=TenantID cs2=%s cs2Label=SourceIP dhost=%s requestMethod=%s msg=%s",
		event.TenantID, c.getSourceIP(event), event.TargetURL, event.Method, event.Description)

	return cefHeader + extensions
}

// formatLEEF formats event as LEEF format
func (c *SIEMClient) formatLEEF(event *SIEMEvent) string {
	leefHeader := fmt.Sprintf("LEEF:1.0|API-Security-Scanner|Security-Scanner|1.0|%s|%s|%s|",
		event.EventType, event.Vulnerability, mapSeverityToLEEF(event.Severity))

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
	return "127.0.0.1"
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
		Timestamp:     time.Now(),
		EventType:     "scan_completed",
		Severity:      "low",
		TenantID:      tenantID,
		TargetURL:     targetURL,
		Method:        "SCAN",
		Vulnerability: scanType,
		Description:   fmt.Sprintf("Completed %s scan for %d endpoints", scanType, endpointCount),
		RawData:       make(map[string]interface{}),
		Tags:          []string{"scan", "api-security", "completed"},
	}
}

// CreateAuthEvent creates a SIEM event for authentication events
func CreateAuthEvent(tenantID, authType, result, sourceIP string) *SIEMEvent {
	severity := "medium"
	if result == "success" {
		severity = "low"
	}
	return &SIEMEvent{
		Timestamp:     time.Now(),
		EventType:     "authentication",
		Severity:      severity,
		TenantID:      tenantID,
		SourceIP:      sourceIP,
		Vulnerability: authType,
		Description:   fmt.Sprintf("Authentication %s for %s", result, authType),
		RawData:       make(map[string]interface{}),
		Tags:          []string{"authentication", "security"},
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
					"GET",
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
