package scanner

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sort"
	"strings"
	"sync"
	"time"

	"api-security-scanner/logging"
	"api-security-scanner/ratelimit"
	"api-security-scanner/discovery"
	"api-security-scanner/history"
	"api-security-scanner/types"
)

// MaxResponseBodySize limits response reading to prevent OOM
const MaxResponseBodySize = 10 * 1024 * 1024 // 10MB

// Config represents the overall configuration
type Config struct {
	APIEndpoints       []types.APIEndpoint    `yaml:"api_endpoints"`
	Auth               Auth                   `yaml:"auth"`
	InjectionPayloads  []string               `yaml:"injection_payloads"`
	RateLimiting       RateLimiting           `yaml:"rate_limiting"`
	XSSPayloads        []string               `yaml:"xss_payloads"`
	Headers            map[string]string      `yaml:"headers"`
	// Phase 3 features
	NoSQLPayloads      []string               `yaml:"nosql_payloads"`
	OpenAPISpec        string                `yaml:"openapi_spec"`
	APIDiscovery       discovery.DiscoveryConfig `yaml:"api_discovery"`
	HistoricalData     history.HistoricalData  `yaml:"historical_data"`
}

// RateLimiting represents rate limiting configuration
type RateLimiting struct {
	RequestsPerSecond      int `yaml:"requests_per_second"`
	MaxConcurrentRequests  int `yaml:"max_concurrent_requests"`
}

// Auth represents authentication credentials
type Auth struct {
	Username string `yaml:"username"`
	Password string `yaml:"password"`
}

// Custom error types
type AuthError struct{ message string }
type HTTPMethodError struct{ message string }
type InjectionError struct{ message string }
type XSSError struct{ message string }
type HeaderSecurityError struct{ message string }
type AuthBypassError struct{ message string }
type ParameterTamperingError struct{ message string }
type NoSQLInjectionError struct{ message string }

func (e AuthError) Error() string              { return e.message }
func (e HTTPMethodError) Error() string        { return e.message }
func (e InjectionError) Error() string         { return e.message }
func (e XSSError) Error() string               { return e.message }
func (e HeaderSecurityError) Error() string    { return e.message }
func (e AuthBypassError) Error() string        { return e.message }
func (e ParameterTamperingError) Error() string { return e.message }
func (e NoSQLInjectionError) Error() string    { return e.message }

// testRunner holds shared state for test execution
type testRunner struct {
	client      *http.Client
	config      *Config
	limiter     *ratelimit.RateLimiter
	resultMutex sync.Mutex
}

// newTestRunner creates a testRunner with a shared HTTP client
func newTestRunner(config *Config) *testRunner {
	return &testRunner{
		client: &http.Client{
			Timeout: 10 * time.Second,
			Transport: &http.Transport{
				MaxIdleConns:        100,
				MaxIdleConnsPerHost: 10,
				IdleConnTimeout:     90 * time.Second,
			},
		},
		config: config,
	}
}

// doRequest performs an HTTP request and drains the body for connection reuse
func (tr *testRunner) doRequest(req *http.Request) (*http.Response, error) {
	resp, err := tr.client.Do(req)
	if err != nil {
		return nil, err
	}
	return resp, nil
}

// readBody safely reads response body up to MaxResponseBodySize
func readBody(resp *http.Response) ([]byte, error) {
	defer resp.Body.Close()
	limitedReader := io.LimitReader(resp.Body, MaxResponseBodySize)
	return io.ReadAll(limitedReader)
}

// drainBody reads and discards the response body for connection reuse
func drainBody(resp *http.Response) {
	if resp == nil || resp.Body == nil {
		return
	}
	io.Copy(io.Discard, io.LimitReader(resp.Body, 1024))
	resp.Body.Close()
}

// baselineCache caches baseline responses per endpoint to avoid duplicate requests
type baselineCache struct {
	mu      sync.RWMutex
	entries map[string]baselineEntry
}

type baselineEntry struct {
	statusCode int
	body       []byte
	err        error
}

func newBaselineCache() *baselineCache {
	return &baselineCache{entries: make(map[string]baselineEntry)}
}

func (bc *baselineCache) get(key string) (baselineEntry, bool) {
	bc.mu.RLock()
	defer bc.mu.RUnlock()
	entry, ok := bc.entries[key]
	return entry, ok
}

func (bc *baselineCache) set(key string, entry baselineEntry) {
	bc.mu.Lock()
	defer bc.mu.Unlock()
	bc.entries[key] = entry
}

// RunTests runs all security tests concurrently and returns a slice of EndpointResult
func RunTests(config *Config) []types.EndpointResult {
	logging.Info("Starting security tests", map[string]interface{}{
		"endpoints_count": len(config.APIEndpoints),
	})

	// Apply default values for rate limiting if not specified
	requestsPerSecond := config.RateLimiting.RequestsPerSecond
	if requestsPerSecond <= 0 {
		requestsPerSecond = 10
	}

	maxConcurrentRequests := config.RateLimiting.MaxConcurrentRequests
	if maxConcurrentRequests <= 0 {
		maxConcurrentRequests = 5
	}

	// Create rate limiter
	rateLimiter := ratelimit.NewRateLimiter(requestsPerSecond, maxConcurrentRequests)

	tr := newTestRunner(config)
	tr.limiter = rateLimiter
	bc := newBaselineCache()

	results := make([]types.EndpointResult, len(config.APIEndpoints))
	var wg sync.WaitGroup

	// Worker pool: limit concurrent endpoints being tested
	endpointSem := make(chan struct{}, maxConcurrentRequests)

	for i, endpoint := range config.APIEndpoints {
		wg.Add(1)
		results[i] = types.EndpointResult{URL: endpoint.URL, Score: 100}

		go func(e types.APIEndpoint, idx int) {
			defer wg.Done()
			endpointSem <- struct{}{}
			defer func() { <-endpointSem }()

			logging.Debug("Testing endpoint", map[string]interface{}{
				"url":    e.URL,
				"method": e.Method,
				"index":  idx,
			})

			tr.runAllTests(e, idx, &results[idx], bc)
		}(endpoint, i)
	}

	wg.Wait()

	logging.Info("Security tests completed", map[string]interface{}{
		"endpoints_count": len(results),
	})

	return results
}

// runAllTests runs all security tests for a single endpoint sequentially
// This reduces goroutine overhead and simplifies synchronization
func (tr *testRunner) runAllTests(endpoint types.APIEndpoint, idx int, result *types.EndpointResult, bc *baselineCache) {
	tests := []struct {
		name   string
		fn     func(types.APIEndpoint, *baselineCache) error
		weight int
	}{
		{"Auth Test", tr.testAuth, 30},
		{"HTTP Method Test", tr.testHTTPMethod, 20},
		{"Injection Test", tr.testInjection, 50},
		{"XSS Test", tr.testXSS, 40},
		{"Header Security Test", tr.testHeaderSecurity, 25},
		{"Auth Bypass Test", tr.testAuthBypass, 35},
		{"Parameter Tampering Test", tr.testParameterTampering, 30},
		{"NoSQL Injection Test", tr.testNoSQLInjection, 45},
	}

	for _, test := range tests {
		tr.limiter.Wait()
		if err := test.fn(endpoint, bc); err != nil {
			tr.resultMutex.Lock()
			result.Results = append(result.Results, types.TestResult{
				TestName: test.name,
				Passed:   false,
				Message:  err.Error(),
			})
			result.Score -= test.weight
			tr.resultMutex.Unlock()
			logging.Warn(test.name+" failed", map[string]interface{}{
				"url":   endpoint.URL,
				"error": err.Error(),
			})
		} else {
			tr.resultMutex.Lock()
			result.Results = append(result.Results, types.TestResult{
				TestName: test.name,
				Passed:   true,
				Message:  test.name + " Passed",
			})
			tr.resultMutex.Unlock()
			logging.Debug(test.name+" passed", map[string]interface{}{
				"url": endpoint.URL,
			})
		}
		tr.limiter.Done()
	}
}

func (tr *testRunner) testAuth(endpoint types.APIEndpoint, _ *baselineCache) error {
	logging.Debug("Testing authentication", map[string]interface{}{
		"url":    endpoint.URL,
		"method": endpoint.Method,
	})

	req, err := http.NewRequest(endpoint.Method, endpoint.URL, bytes.NewBufferString(endpoint.Body))
	if err != nil {
		logging.Error("Failed to create request", map[string]interface{}{
			"url":   endpoint.URL,
			"error": err.Error(),
		})
		return fmt.Errorf("failed to create request: %v", err)
	}

	req.SetBasicAuth(tr.config.Auth.Username, tr.config.Auth.Password)

	resp, err := tr.doRequest(req)
	if err != nil {
		logging.Error("Request failed", map[string]interface{}{
			"url":   endpoint.URL,
			"error": err.Error(),
		})
		return fmt.Errorf("request failed: %v", err)
	}
	drainBody(resp)

	switch resp.StatusCode {
	case http.StatusOK, http.StatusCreated, http.StatusAccepted:
		return nil
	case http.StatusUnauthorized:
		return AuthError{"authentication failed: incorrect credentials"}
	case http.StatusForbidden:
		return AuthError{"authentication failed: access forbidden"}
	default:
		logging.Warn("Unexpected status code", map[string]interface{}{
			"url":    endpoint.URL,
			"status": resp.StatusCode,
		})
		return fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}
}

func (tr *testRunner) testHTTPMethod(endpoint types.APIEndpoint, _ *baselineCache) error {
	req, err := http.NewRequest(endpoint.Method, endpoint.URL, bytes.NewBufferString(endpoint.Body))
	if err != nil {
		return fmt.Errorf("failed to create request: %v", err)
	}
	req.SetBasicAuth(tr.config.Auth.Username, tr.config.Auth.Password)

	resp, err := tr.doRequest(req)
	if err != nil {
		return fmt.Errorf("request failed: %v", err)
	}
	drainBody(resp)

	switch resp.StatusCode {
	case http.StatusOK, http.StatusCreated, http.StatusAccepted:
		return nil
	case http.StatusMethodNotAllowed, http.StatusNotFound:
		return HTTPMethodError{fmt.Sprintf("disallowed method returned status: %d", resp.StatusCode)}
	default:
		return fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}
}

func (tr *testRunner) getBaseline(endpoint types.APIEndpoint, bc *baselineCache) (baselineEntry, error) {
	key := endpoint.URL + "|" + endpoint.Method + "|" + endpoint.Body
	if entry, ok := bc.get(key); ok {
		return entry, nil
	}

	req, err := http.NewRequest(endpoint.Method, endpoint.URL, bytes.NewBufferString(endpoint.Body))
	if err != nil {
		return baselineEntry{}, fmt.Errorf("failed to create baseline request: %v", err)
	}
	req.SetBasicAuth(tr.config.Auth.Username, tr.config.Auth.Password)

	resp, err := tr.doRequest(req)
	if err != nil {
		return baselineEntry{}, fmt.Errorf("baseline request failed: %v", err)
	}

	body, err := readBody(resp)
	entry := baselineEntry{
		statusCode: resp.StatusCode,
		body:       body,
		err:        err,
	}
	bc.set(key, entry)
	return entry, nil
}

func (tr *testRunner) testInjection(endpoint types.APIEndpoint, bc *baselineCache) error {
	logging.Debug("Testing injection", map[string]interface{}{
		"url":            endpoint.URL,
		"method":         endpoint.Method,
		"payloads_count": len(tr.config.InjectionPayloads),
	})

	baseline, err := tr.getBaseline(endpoint, bc)
	if err != nil {
		return err
	}

	if baseline.statusCode == http.StatusUnauthorized || baseline.statusCode == http.StatusForbidden {
		logging.Warn("Cannot perform injection test", map[string]interface{}{
			"url":    endpoint.URL,
			"status": baseline.statusCode,
		})
		return fmt.Errorf("cannot perform injection test: baseline request failed with status %d", baseline.statusCode)
	}

	baselineBody := string(baseline.body)

	for i, payload := range tr.config.InjectionPayloads {
		logging.Debug("Testing injection payload", map[string]interface{}{
			"url":     endpoint.URL,
			"payload": payload,
			"index":   i,
		})

		reqBody := fmt.Sprintf(endpoint.Body, payload)
		req, err := http.NewRequest(endpoint.Method, endpoint.URL, bytes.NewBufferString(reqBody))
		if err != nil {
			logging.Error("Failed to create request", map[string]interface{}{
				"url":     endpoint.URL,
				"payload": payload,
				"error":   err.Error(),
			})
			return fmt.Errorf("failed to create request: %v", err)
		}
		req.SetBasicAuth(tr.config.Auth.Username, tr.config.Auth.Password)

		resp, err := tr.doRequest(req)
		if err != nil {
			logging.Error("Request failed", map[string]interface{}{
				"url":     endpoint.URL,
				"payload": payload,
				"error":   err.Error(),
			})
			return fmt.Errorf("request failed: %v", err)
		}

		body, err := readBody(resp)
		if err != nil {
			logging.Error("Failed to read response body", map[string]interface{}{
				"url":     endpoint.URL,
				"payload": payload,
				"error":   err.Error(),
			})
			return fmt.Errorf("failed to read response body: %v", err)
		}

		if indicatorsOfSQLInjection(string(body), baselineBody) {
			logging.Warn("Potential SQL injection detected", map[string]interface{}{
				"url":     endpoint.URL,
				"payload": payload,
			})
			return InjectionError{fmt.Sprintf("potential SQL injection detected with payload: %s", payload)}
		}
	}
	return nil
}

func indicatorsOfSQLInjection(responseBody, baselineBody string) bool {
	sqlErrorMessages := []string{
		"SQL syntax",
		"mysql_fetch_array",
		"ORA-01756",
		"SQLite3::SQLException",
		"PostgreSQL ERROR",
		"Incorrect syntax near",
		"SQLSTATE[",
		"JDBC Driver",
		"Microsoft SQL Server",
		"You have an error in your SQL syntax",
	}

	for _, errorMsg := range sqlErrorMessages {
		if strings.Contains(responseBody, errorMsg) {
			return true
		}
	}

	if len(responseBody) > len(baselineBody)*2 || len(responseBody) < len(baselineBody)/2 {
		return true
	}

	if strings.Count(responseBody, "{") != strings.Count(baselineBody, "{") ||
		strings.Count(responseBody, "}") != strings.Count(baselineBody, "}") {
		return true
	}

	return false
}

func (tr *testRunner) testXSS(endpoint types.APIEndpoint, bc *baselineCache) error {
	logging.Debug("Testing XSS", map[string]interface{}{
		"url":            endpoint.URL,
		"method":         endpoint.Method,
		"payloads_count": len(tr.config.XSSPayloads),
	})

	baseline, err := tr.getBaseline(endpoint, bc)
	if err != nil {
		return err
	}

	if baseline.statusCode == http.StatusUnauthorized || baseline.statusCode == http.StatusForbidden {
		logging.Warn("Cannot perform XSS test", map[string]interface{}{
			"url":    endpoint.URL,
			"status": baseline.statusCode,
		})
		return fmt.Errorf("cannot perform XSS test: baseline request failed with status %d", baseline.statusCode)
	}

	baselineBody := string(baseline.body)

	for i, payload := range tr.config.XSSPayloads {
		logging.Debug("Testing XSS payload", map[string]interface{}{
			"url":     endpoint.URL,
			"payload": payload,
			"index":   i,
		})

		reqBody := strings.Replace(endpoint.Body, "\"value\"", fmt.Sprintf("\"%s\"", payload), -1)
		req, err := http.NewRequest(endpoint.Method, endpoint.URL, bytes.NewBufferString(reqBody))
		if err != nil {
			logging.Error("Failed to create request", map[string]interface{}{
				"url":     endpoint.URL,
				"payload": payload,
				"error":   err.Error(),
			})
			return fmt.Errorf("failed to create request: %v", err)
		}
		req.SetBasicAuth(tr.config.Auth.Username, tr.config.Auth.Password)

		resp, err := tr.doRequest(req)
		if err != nil {
			logging.Error("Request failed", map[string]interface{}{
				"url":     endpoint.URL,
				"payload": payload,
				"error":   err.Error(),
			})
			return fmt.Errorf("request failed: %v", err)
		}

		body, err := readBody(resp)
		if err != nil {
			logging.Error("Failed to read response body", map[string]interface{}{
				"url":     endpoint.URL,
				"payload": payload,
				"error":   err.Error(),
			})
			return fmt.Errorf("failed to read response body: %v", err)
		}

		if indicatorsOfXSS(string(body), baselineBody, payload) {
			logging.Warn("Potential XSS detected", map[string]interface{}{
				"url":     endpoint.URL,
				"payload": payload,
			})
			return XSSError{fmt.Sprintf("potential XSS detected with payload: %s", payload)}
		}
	}
	return nil
}

func indicatorsOfXSS(responseBody, baselineBody, payload string) bool {
	if strings.Contains(responseBody, payload) && !strings.Contains(baselineBody, payload) {
		scriptContext := strings.Contains(responseBody, fmt.Sprintf("<script>%s</script>", payload)) ||
			strings.Contains(responseBody, fmt.Sprintf("onload=\"%s\"", payload)) ||
			strings.Contains(responseBody, fmt.Sprintf("onerror=\"%s\"", payload)) ||
			strings.Contains(responseBody, fmt.Sprintf("onclick=\"%s\"", payload))

		if scriptContext {
			return true
		}

		if strings.Contains(responseBody, fmt.Sprintf("<%s>", payload)) ||
			strings.Contains(responseBody, fmt.Sprintf(">%s<", payload)) {
			return true
		}
	}

	return false
}

func (tr *testRunner) testHeaderSecurity(endpoint types.APIEndpoint, _ *baselineCache) error {
	logging.Debug("Testing header security", map[string]interface{}{
		"url":    endpoint.URL,
		"method": endpoint.Method,
	})

	req, err := http.NewRequest(endpoint.Method, endpoint.URL, bytes.NewBufferString(endpoint.Body))
	if err != nil {
		logging.Error("Failed to create request", map[string]interface{}{
			"url":   endpoint.URL,
			"error": err.Error(),
		})
		return fmt.Errorf("failed to create request: %v", err)
	}
	req.SetBasicAuth(tr.config.Auth.Username, tr.config.Auth.Password)

	for key, value := range tr.config.Headers {
		req.Header.Set(key, value)
	}

	resp, err := tr.doRequest(req)
	if err != nil {
		logging.Error("Request failed", map[string]interface{}{
			"url":   endpoint.URL,
			"error": err.Error(),
		})
		return fmt.Errorf("request failed: %v", err)
	}
	drainBody(resp)

	issues := make([]string, 0, 10)

	securityHeaders := map[string]string{
		"X-Content-Type-Options":    "nosniff",
		"X-Frame-Options":           "DENY or SAMEORIGIN",
		"X-XSS-Protection":          "1; mode=block",
		"Strict-Transport-Security": "max-age=31536000; includeSubDomains",
		"Content-Security-Policy":   "policy directives",
	}

	for header, recommended := range securityHeaders {
		if resp.Header.Get(header) == "" {
			issues = append(issues, fmt.Sprintf("Missing recommended security header: %s (recommended value: %s)", header, recommended))
		}
	}

	insecureHeaders := []string{
		"X-Powered-By",
		"Server",
	}

	for _, header := range insecureHeaders {
		if val := resp.Header.Get(header); val != "" {
			issues = append(issues, fmt.Sprintf("Insecure information disclosure header: %s (%s)", header, val))
		}
	}

	if resp.Header.Get("Access-Control-Allow-Origin") == "*" {
		issues = append(issues, "Insecure CORS policy: Access-Control-Allow-Origin set to wildcard (*)")
	}

	for _, cookie := range resp.Header.Values("Set-Cookie") {
		if !strings.Contains(cookie, "Secure") {
			issues = append(issues, "Cookie missing Secure attribute: "+cookie)
		}
		if !strings.Contains(cookie, "HttpOnly") {
			issues = append(issues, "Cookie missing HttpOnly attribute: "+cookie)
		}
		if !strings.Contains(cookie, "SameSite") {
			issues = append(issues, "Cookie missing SameSite attribute: "+cookie)
		}
	}

	if len(issues) > 0 {
		logging.Warn("Header security issues detected", map[string]interface{}{
			"url":    endpoint.URL,
			"issues": issues,
		})
		return HeaderSecurityError{fmt.Sprintf("header security issues detected: %s", strings.Join(issues, "; "))}
	}

	return nil
}

func (tr *testRunner) testAuthBypass(endpoint types.APIEndpoint, _ *baselineCache) error {
	logging.Debug("Testing authentication bypass", map[string]interface{}{
		"url":      endpoint.URL,
		"method":   endpoint.Method,
		"has_auth": tr.config.Auth.Username != "" && tr.config.Auth.Password != "",
	})

	// Test 1: Request without authentication
	req1, err := http.NewRequest(endpoint.Method, endpoint.URL, bytes.NewBufferString(endpoint.Body))
	if err != nil {
		logging.Error("Failed to create request without auth", map[string]interface{}{
			"url":   endpoint.URL,
			"error": err.Error(),
		})
		return fmt.Errorf("failed to create request without auth: %v", err)
	}

	resp1, err := tr.doRequest(req1)
	if err != nil {
		logging.Error("Request without auth failed", map[string]interface{}{
			"url":   endpoint.URL,
			"error": err.Error(),
		})
		return fmt.Errorf("request without auth failed: %v", err)
	}
	drainBody(resp1)

	if resp1.StatusCode == http.StatusOK || resp1.StatusCode == http.StatusCreated || resp1.StatusCode == http.StatusAccepted {
		logging.Warn("Authentication bypass detected", map[string]interface{}{
			"url":                 endpoint.URL,
			"status_without_auth": resp1.StatusCode,
		})
		return AuthBypassError{fmt.Sprintf("authentication bypass detected: endpoint accessible without authentication (status: %d)", resp1.StatusCode)}
	}

	// Test 2: Request with invalid credentials
	if tr.config.Auth.Username != "" && tr.config.Auth.Password != "" {
		req2, err := http.NewRequest(endpoint.Method, endpoint.URL, bytes.NewBufferString(endpoint.Body))
		if err != nil {
			logging.Error("Failed to create request with invalid auth", map[string]interface{}{
				"url":   endpoint.URL,
				"error": err.Error(),
			})
			return fmt.Errorf("failed to create request with invalid auth: %v", err)
		}
		req2.SetBasicAuth("invalid_user", "invalid_pass")

		resp2, err := tr.doRequest(req2)
		if err != nil {
			logging.Error("Request with invalid auth failed", map[string]interface{}{
				"url":   endpoint.URL,
				"error": err.Error(),
			})
			return fmt.Errorf("request with invalid auth failed: %v", err)
		}
		drainBody(resp2)

		if resp2.StatusCode == http.StatusOK || resp2.StatusCode == http.StatusCreated || resp2.StatusCode == http.StatusAccepted {
			logging.Warn("Authentication bypass with invalid credentials", map[string]interface{}{
				"url": endpoint.URL,
				"status_with_invalid_auth": resp2.StatusCode,
			})
			return AuthBypassError{fmt.Sprintf("authentication bypass detected: endpoint accessible with invalid credentials (status: %d)", resp2.StatusCode)}
		}
	}

	// Test 3: Check for common auth bypass headers
	bypassHeaders := map[string]string{
		"X-Forwarded-For":  "127.0.0.1",
		"X-Original-URL":   endpoint.URL,
		"X-Rewrite-URL":    endpoint.URL,
		"X-Originating-IP": "127.0.0.1",
	}

	req3, err := http.NewRequest(endpoint.Method, endpoint.URL, bytes.NewBufferString(endpoint.Body))
	if err != nil {
		logging.Error("Failed to create request with bypass headers", map[string]interface{}{
			"url":   endpoint.URL,
			"error": err.Error(),
		})
		return fmt.Errorf("failed to create request with bypass headers: %v", err)
	}

	for key, value := range bypassHeaders {
		req3.Header.Set(key, value)
	}

	resp3, err := tr.doRequest(req3)
	if err != nil {
		logging.Error("Request with bypass headers failed", map[string]interface{}{
			"url":   endpoint.URL,
			"error": err.Error(),
		})
		return fmt.Errorf("request with bypass headers failed: %v", err)
	}
	drainBody(resp3)

	if resp3.StatusCode == http.StatusOK || resp3.StatusCode == http.StatusCreated || resp3.StatusCode == http.StatusAccepted {
		logging.Warn("Authentication bypass with headers", map[string]interface{}{
			"url": endpoint.URL,
			"status_with_bypass_headers": resp3.StatusCode,
		})
		return AuthBypassError{fmt.Sprintf("authentication bypass detected: endpoint accessible with bypass headers (status: %d)", resp3.StatusCode)}
	}

	return nil
}

func (tr *testRunner) testParameterTampering(endpoint types.APIEndpoint, _ *baselineCache) error {
	logging.Debug("Testing parameter tampering", map[string]interface{}{
		"url":    endpoint.URL,
		"method": endpoint.Method,
	})

	// Test 1: Modify numeric parameters in the body
	if strings.Contains(endpoint.Body, "\"key\":") {
		modifiedBody := strings.Replace(endpoint.Body, "\"value\"", "\"12345\"", -1)

		req, err := http.NewRequest(endpoint.Method, endpoint.URL, bytes.NewBufferString(modifiedBody))
		if err != nil {
			logging.Error("Failed to create request with modified parameters", map[string]interface{}{
				"url":   endpoint.URL,
				"error": err.Error(),
			})
			return fmt.Errorf("failed to create request with modified parameters: %v", err)
		}
		req.SetBasicAuth(tr.config.Auth.Username, tr.config.Auth.Password)

		resp, err := tr.doRequest(req)
		if err != nil {
			logging.Error("Request with modified parameters failed", map[string]interface{}{
				"url":   endpoint.URL,
				"error": err.Error(),
			})
			return fmt.Errorf("request with modified parameters failed: %v", err)
		}
		drainBody(resp)

		logging.Debug("Parameter modification test completed", map[string]interface{}{
			"url":    endpoint.URL,
			"status": resp.StatusCode,
		})
	}

	// Test 2: Add extra parameters
	if endpoint.Body != "" {
		extraParamBody := strings.TrimRight(endpoint.Body, "}") + ", \"extra_param\": \"tampered_value\"}"

		req, err := http.NewRequest(endpoint.Method, endpoint.URL, bytes.NewBufferString(extraParamBody))
		if err != nil {
			logging.Error("Failed to create request with extra parameters", map[string]interface{}{
				"url":   endpoint.URL,
				"error": err.Error(),
			})
			return fmt.Errorf("failed to create request with extra parameters: %v", err)
		}
		req.SetBasicAuth(tr.config.Auth.Username, tr.config.Auth.Password)

		resp, err := tr.doRequest(req)
		if err != nil {
			logging.Error("Request with extra parameters failed", map[string]interface{}{
				"url":   endpoint.URL,
				"error": err.Error(),
			})
			return fmt.Errorf("request with extra parameters failed: %v", err)
		}
		drainBody(resp)

		logging.Debug("Extra parameter test completed", map[string]interface{}{
			"url":                      endpoint.URL,
			"status_with_extra_params": resp.StatusCode,
		})
	}

	// Test 3: Test for IDOR
	if strings.Contains(endpoint.URL, "/") {
		modifiedURL := strings.Replace(endpoint.URL, "1", "2", -1)
		if modifiedURL != endpoint.URL {
			req, err := http.NewRequest(endpoint.Method, modifiedURL, bytes.NewBufferString(endpoint.Body))
			if err != nil {
				logging.Error("Failed to create request with modified URL", map[string]interface{}{
					"url":   modifiedURL,
					"error": err.Error(),
				})
				return fmt.Errorf("failed to create request with modified URL: %v", err)
			}
			req.SetBasicAuth(tr.config.Auth.Username, tr.config.Auth.Password)

			resp, err := tr.doRequest(req)
			if err != nil {
				logging.Error("Request with modified URL failed", map[string]interface{}{
					"url":   modifiedURL,
					"error": err.Error(),
				})
				return fmt.Errorf("request with modified URL failed: %v", err)
			}
			drainBody(resp)

			if resp.StatusCode == http.StatusOK || resp.StatusCode == http.StatusCreated || resp.StatusCode == http.StatusAccepted {
				logging.Warn("Potential IDOR detected", map[string]interface{}{
					"original_url": endpoint.URL,
					"modified_url": modifiedURL,
					"status":       resp.StatusCode,
				})
				return ParameterTamperingError{fmt.Sprintf("potential IDOR detected: able to access %s (status: %d)", modifiedURL, resp.StatusCode)}
			}
		}
	}

	return nil
}

func (tr *testRunner) testNoSQLInjection(endpoint types.APIEndpoint, bc *baselineCache) error {
	logging.Debug("Testing NoSQL injection", map[string]interface{}{
		"url":            endpoint.URL,
		"method":         endpoint.Method,
		"payloads_count": len(tr.config.NoSQLPayloads),
	})

	payloads := tr.config.NoSQLPayloads
	if len(payloads) == 0 {
		payloads = []string{
			"{$ne: null}",
			"{$gt: ''}",
			"{$or: [1,1]}",
			"{$where: 'sleep(100)'}",
			"{$regex: '.*'}",
			"{$exists: true}",
			"{$in: [1,2,3]}",
		}
	}

	baseline, err := tr.getBaseline(endpoint, bc)
	if err != nil {
		return err
	}

	if baseline.statusCode == http.StatusUnauthorized || baseline.statusCode == http.StatusForbidden {
		logging.Warn("Cannot perform NoSQL injection test", map[string]interface{}{
			"url":    endpoint.URL,
			"status": baseline.statusCode,
		})
		return fmt.Errorf("cannot perform NoSQL injection test: baseline request failed with status %d", baseline.statusCode)
	}

	baselineBody := string(baseline.body)

	for i, payload := range payloads {
		logging.Debug("Testing NoSQL injection payload", map[string]interface{}{
			"url":     endpoint.URL,
			"payload": payload,
			"index":   i,
		})

		reqBody := strings.Replace(endpoint.Body, "\"value\"", fmt.Sprintf("\"%s\"", payload), -1)
		req, err := http.NewRequest(endpoint.Method, endpoint.URL, bytes.NewBufferString(reqBody))
		if err != nil {
			logging.Error("Failed to create request", map[string]interface{}{
				"url":     endpoint.URL,
				"payload": payload,
				"error":   err.Error(),
			})
			return fmt.Errorf("failed to create request: %v", err)
		}
		req.SetBasicAuth(tr.config.Auth.Username, tr.config.Auth.Password)

		resp, err := tr.doRequest(req)
		if err != nil {
			logging.Error("Request failed", map[string]interface{}{
				"url":     endpoint.URL,
				"payload": payload,
				"error":   err.Error(),
			})
			return fmt.Errorf("request failed: %v", err)
		}

		body, err := readBody(resp)
		if err != nil {
			logging.Error("Failed to read response body", map[string]interface{}{
				"url":     endpoint.URL,
				"payload": payload,
				"error":   err.Error(),
			})
			return fmt.Errorf("failed to read response body: %v", err)
		}

		if indicatorsOfNoSQLInjection(string(body), baselineBody, payload) {
			logging.Warn("Potential NoSQL injection detected", map[string]interface{}{
				"url":     endpoint.URL,
				"payload": payload,
			})
			return NoSQLInjectionError{fmt.Sprintf("potential NoSQL injection detected with payload: %s", payload)}
		}
	}
	return nil
}

func indicatorsOfNoSQLInjection(responseBody, baselineBody, payload string) bool {
	nosqlErrorMessages := []string{
		"MongoError",
		"MongoServerError",
		"MongoDB error",
		"Cannot create property",
		"Unexpected token",
		"SyntaxError",
		"CastError",
		"ValidationError",
		"MongoCursor",
		"MongoTimeoutError",
	}

	for _, errorMsg := range nosqlErrorMessages {
		if strings.Contains(responseBody, errorMsg) {
			return true
		}
	}

	if strings.Contains(responseBody, payload) && !strings.Contains(baselineBody, payload) {
		return true
	}

	if len(responseBody) > len(baselineBody)*2 {
		return true
	}

	if strings.Contains(responseBody, "{$") || strings.Contains(responseBody, "_id") {
		return true
	}

	responseLines := strings.Split(responseBody, "\n")
	baselineLines := strings.Split(baselineBody, "\n")
	if len(responseLines) > int(float64(len(baselineLines))*1.5) {
		return true
	}

	return false
}

func GenerateDetailedReport(results []types.EndpointResult) {
	var b strings.Builder
	b.WriteString("\nAPI Security Scan Detailed Report\n")
	b.WriteString("==================================\n")

	for _, result := range results {
		b.WriteString(fmt.Sprintf("\nEndpoint: %s\n", result.URL))
		b.WriteString(fmt.Sprintf("Overall Score: %d/100\n", result.Score))
		b.WriteString("Test Results:\n")

		sort.Slice(result.Results, func(i, j int) bool {
			return result.Results[i].TestName < result.Results[j].TestName
		})

		for _, testResult := range result.Results {
			status := "PASSED"
			if !testResult.Passed {
				status = "FAILED"
			}
			b.WriteString(fmt.Sprintf("- %s: %s\n", testResult.TestName, status))
			b.WriteString(fmt.Sprintf("  Details: %s\n", formatTestMessage(testResult.Message, result.URL)))
		}

		b.WriteString("Risk Assessment:\n")
		b.WriteString(generateRiskAssessment(result))
		b.WriteString("\n------------------------\n")
	}

	b.WriteString("\nOverall Security Assessment:\n")
	b.WriteString(generateOverallAssessment(results))
	b.WriteString("\n")
	fmt.Print(b.String())
}

func formatTestMessage(message string, url string) string {
	prefix := fmt.Sprintf("Test Failed for %s:", url)
	return strings.TrimSpace(strings.TrimPrefix(message, prefix))
}

func generateRiskAssessment(result types.EndpointResult) string {
	var risks []string
	for _, testResult := range result.Results {
		if !testResult.Passed {
			switch testResult.TestName {
			case "Auth Test":
				risks = append(risks, "- Authentication vulnerabilities may allow unauthorized access.")
			case "HTTP Method Test":
				risks = append(risks, "- Improper HTTP method handling could lead to security bypasses.")
			case "Injection Test":
				risks = append(risks, "- SQL injection vulnerabilities pose a significant data breach risk.")
			case "XSS Test":
				risks = append(risks, "- Cross-site scripting vulnerabilities could allow malicious script execution.")
			case "Header Security Test":
				risks = append(risks, "- Insecure headers may expose sensitive information or lack security protections.")
			case "Auth Bypass Test":
				risks = append(risks, "- Authentication bypass vulnerabilities could allow unauthorized access to protected resources.")
			case "Parameter Tampering Test":
				risks = append(risks, "- Parameter tampering vulnerabilities could allow attackers to manipulate API requests.")
			case "NoSQL Injection Test":
				risks = append(risks, "- NoSQL injection vulnerabilities pose a significant data breach risk in NoSQL databases.")
			}
		}
	}

	if len(risks) == 0 {
		return "No significant risks detected."
	}
	return strings.Join(risks, "\n")
}

func generateOverallAssessment(results []types.EndpointResult) string {
	totalScore := 0
	criticalVulnerabilities := 0
	for _, result := range results {
		totalScore += result.Score
		for _, testResult := range result.Results {
			if !testResult.Passed && testResult.TestName == "Injection Test" {
				criticalVulnerabilities++
			}
		}
	}

	if len(results) == 0 {
		return "No endpoints tested."
	}

	averageScore := totalScore / len(results)

	var b strings.Builder
	b.WriteString(fmt.Sprintf("Average Security Score: %d/100\n", averageScore))
	b.WriteString(fmt.Sprintf("Critical Vulnerabilities Detected: %d\n\n", criticalVulnerabilities))

	if averageScore >= 90 {
		b.WriteString("Overall security posture is strong, but continuous monitoring is recommended.")
	} else if averageScore >= 70 {
		b.WriteString("Moderate security risks detected. Address identified vulnerabilities promptly.")
	} else {
		b.WriteString("Significant security risks identified. Immediate action is required to improve API security.")
	}

	return b.String()
}

// GenerateJSONReport generates a JSON formatted report
func GenerateJSONReport(results []types.EndpointResult) {
	output := struct {
		ScanResults       []jsonEndpointResult `json:"scan_results"`
		OverallAssessment string               `json:"overall_assessment"`
	}{
		ScanResults:       make([]jsonEndpointResult, len(results)),
		OverallAssessment: generateOverallAssessment(results),
	}

	for i, result := range results {
		output.ScanResults[i] = jsonEndpointResult{
			Endpoint:         result.URL,
			Score:            result.Score,
			Tests:            make([]jsonTestResult, len(result.Results)),
			RiskAssessment:   generateRiskAssessment(result),
		}
		for j, tr := range result.Results {
			output.ScanResults[i].Tests[j] = jsonTestResult{
				Name:    tr.TestName,
				Passed:  tr.Passed,
				Message: tr.Message,
			}
		}
	}

	data, _ := json.MarshalIndent(output, "", "  ")
	fmt.Println(string(data))
}

type jsonEndpointResult struct {
	Endpoint       string           `json:"endpoint"`
	Score          int              `json:"score"`
	Tests          []jsonTestResult `json:"tests"`
	RiskAssessment string           `json:"risk_assessment"`
}

type jsonTestResult struct {
	Name    string `json:"name"`
	Passed  bool   `json:"passed"`
	Message string `json:"message"`
}

// GenerateHTMLReport generates an HTML formatted report
func GenerateHTMLReport(results []types.EndpointResult) {
	var b strings.Builder
	b.WriteString("<!DOCTYPE html>\n<html>\n<head>\n")
	b.WriteString("  <title>API Security Scan Report</title>\n")
	b.WriteString(`  <style>
    body { font-family: Arial, sans-serif; margin: 20px; }
    .header { background-color: #f0f0f0; padding: 10px; border-radius: 5px; }
    .endpoint { margin: 20px 0; padding: 15px; border: 1px solid #ccc; border-radius: 5px; }
    .passed { color: green; } .failed { color: red; }
    .score-high { color: green; font-weight: bold; }
    .score-medium { color: orange; font-weight: bold; }
    .score-low { color: red; font-weight: bold; }
  </style>
`)
	b.WriteString("</head>\n<body>\n")
	b.WriteString("  <h1>API Security Scan Detailed Report</h1>\n")

	for _, result := range results {
		b.WriteString("  <div class=\"endpoint\">\n")
		b.WriteString(fmt.Sprintf("    <h2>Endpoint: %s</h2>\n", result.URL))

		scoreClass := "score-low"
		if result.Score >= 90 {
			scoreClass = "score-high"
		} else if result.Score >= 70 {
			scoreClass = "score-medium"
		}
		b.WriteString(fmt.Sprintf("    <p><strong>Overall Score:</strong> <span class=\"%s\">%d/100</span></p>\n", scoreClass, result.Score))

		b.WriteString("    <h3>Test Results:</h3>\n")
		b.WriteString("    <ul>\n")
		for _, testResult := range result.Results {
			statusClass := "passed"
			statusText := "PASSED"
			if !testResult.Passed {
				statusClass = "failed"
				statusText = "FAILED"
			}
			b.WriteString(fmt.Sprintf("      <li><strong>%s:</strong> <span class=\"%s\">%s</span> - %s</li>\n",
				testResult.TestName, statusClass, statusText, testResult.Message))
		}
		b.WriteString("    </ul>\n")

		b.WriteString("    <h3>Risk Assessment:</h3>\n")
		b.WriteString(fmt.Sprintf("    <p>%s</p>\n", generateRiskAssessment(result)))
		b.WriteString("  </div>\n")
	}

	b.WriteString("  <div class=\"endpoint\">\n")
	b.WriteString("    <h2>Overall Security Assessment</h2>\n")
	b.WriteString(fmt.Sprintf("    <p>%s</p>\n", generateOverallAssessment(results)))
	b.WriteString("  </div>\n")
	b.WriteString("</body>\n</html>\n")
	fmt.Print(b.String())
}

// GenerateCSVReport generates a CSV formatted report
func GenerateCSVReport(results []types.EndpointResult) {
	var b strings.Builder
	b.WriteString("Endpoint,Score,Test Name,Passed,Message,Risk Assessment\n")

	for _, result := range results {
		endpoint := strings.ReplaceAll(result.URL, "\"", "\"\"")
		riskAssessment := strings.ReplaceAll(generateRiskAssessment(result), "\"", "\"\"")

		for _, testResult := range result.Results {
			testName := strings.ReplaceAll(testResult.TestName, "\"", "\"\"")
			message := strings.ReplaceAll(testResult.Message, "\"", "\"\"")
			passed := "true"
			if !testResult.Passed {
				passed = "false"
			}

			b.WriteString(fmt.Sprintf("\"%s\",%d,\"%s\",%s,\"%s\",\"%s\"\n",
				endpoint, result.Score, testName, passed, message, riskAssessment))
		}
	}

	overall := strings.ReplaceAll(generateOverallAssessment(results), "\"", "\"\"")
	b.WriteString(fmt.Sprintf("\"OVERALL\",,\"\",\"\",\"\",\"%s\"\n", overall))
	fmt.Print(b.String())
}

// GenerateXMLReport generates an XML formatted report
func GenerateXMLReport(results []types.EndpointResult) {
	var b strings.Builder
	b.WriteString("<?xml version=\"1.0\" encoding=\"UTF-8\"?>\n")
	b.WriteString("<api_security_scan>\n")
	b.WriteString("  <scan_results>\n")

	for _, result := range results {
		b.WriteString("    <endpoint>\n")
		b.WriteString(fmt.Sprintf("      <url>%s</url>\n", result.URL))
		b.WriteString(fmt.Sprintf("      <score>%d</score>\n", result.Score))
		b.WriteString("      <tests>\n")

		for _, testResult := range result.Results {
			b.WriteString("        <test>\n")
			b.WriteString(fmt.Sprintf("          <name>%s</name>\n", testResult.TestName))
			b.WriteString(fmt.Sprintf("          <passed>%t</passed>\n", testResult.Passed))
			b.WriteString(fmt.Sprintf("          <message>%s</message>\n", testResult.Message))
			b.WriteString("        </test>\n")
		}

		b.WriteString("      </tests>\n")
		b.WriteString(fmt.Sprintf("      <risk_assessment>%s</risk_assessment>\n", generateRiskAssessment(result)))
		b.WriteString("    </endpoint>\n")
	}

	b.WriteString("  </scan_results>\n")
	b.WriteString(fmt.Sprintf("  <overall_assessment>%s</overall_assessment>\n", generateOverallAssessment(results)))
	b.WriteString("</api_security_scan>\n")
	fmt.Print(b.String())
}

// Backward-compatible wrapper functions for tests
// These use the default HTTP transport to allow test mocking.

func newTestRunnerForTests(config *Config) *testRunner {
	tr := newTestRunner(config)
	tr.client = &http.Client{Timeout: 10 * time.Second}
	return tr
}

func testAuth(endpoint types.APIEndpoint, auth Auth) error {
	tr := newTestRunnerForTests(&Config{Auth: auth})
	return tr.testAuth(endpoint, nil)
}

func testHTTPMethod(endpoint types.APIEndpoint, auth Auth) error {
	tr := newTestRunnerForTests(&Config{Auth: auth})
	return tr.testHTTPMethod(endpoint, nil)
}

func testInjection(endpoint types.APIEndpoint, auth Auth, payloads []string) error {
	tr := newTestRunnerForTests(&Config{Auth: auth, InjectionPayloads: payloads})
	return tr.testInjection(endpoint, newBaselineCache())
}

func testXSS(endpoint types.APIEndpoint, auth Auth, payloads []string) error {
	tr := newTestRunnerForTests(&Config{Auth: auth, XSSPayloads: payloads})
	return tr.testXSS(endpoint, newBaselineCache())
}

func testHeaderSecurity(endpoint types.APIEndpoint, auth Auth, customHeaders map[string]string) error {
	tr := newTestRunnerForTests(&Config{Auth: auth, Headers: customHeaders})
	return tr.testHeaderSecurity(endpoint, nil)
}

func testAuthBypass(endpoint types.APIEndpoint, auth Auth) error {
	tr := newTestRunnerForTests(&Config{Auth: auth})
	return tr.testAuthBypass(endpoint, nil)
}

func testParameterTampering(endpoint types.APIEndpoint, auth Auth) error {
	tr := newTestRunnerForTests(&Config{Auth: auth})
	return tr.testParameterTampering(endpoint, nil)
}

func testNoSQLInjection(endpoint types.APIEndpoint, auth Auth, payloads []string) error {
	tr := newTestRunnerForTests(&Config{Auth: auth, NoSQLPayloads: payloads})
	return tr.testNoSQLInjection(endpoint, newBaselineCache())
}
