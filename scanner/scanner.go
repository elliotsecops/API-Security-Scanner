package scanner

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"sort"
	"strings"
	"sync"
	"time"

	"api-security-scanner/logging"
	"api-security-scanner/ratelimit"
)

// Config represents the overall configuration
type Config struct {
	APIEndpoints      []APIEndpoint `yaml:"api_endpoints"`
	Auth              Auth          `yaml:"auth"`
	InjectionPayloads []string      `yaml:"injection_payloads"`
	RateLimiting      RateLimiting  `yaml:"rate_limiting"`
	XSSPayloads       []string      `yaml:"xss_payloads"`
	Headers           map[string]string `yaml:"headers"`
}

// RateLimiting represents rate limiting configuration
type RateLimiting struct {
	RequestsPerSecond      int `yaml:"requests_per_second"`
	MaxConcurrentRequests  int `yaml:"max_concurrent_requests"`
}

// APIEndpoint represents a single API endpoint configuration
type APIEndpoint struct {
	URL    string `yaml:"url"`
	Method string `yaml:"method"`
	Body   string `yaml:"body"`
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

func (e AuthError) Error() string              { return e.message }
func (e HTTPMethodError) Error() string        { return e.message }
func (e InjectionError) Error() string         { return e.message }
func (e XSSError) Error() string               { return e.message }
func (e HeaderSecurityError) Error() string    { return e.message }
func (e AuthBypassError) Error() string        { return e.message }
func (e ParameterTamperingError) Error() string { return e.message }

// EndpointResult represents the results of tests for a single endpoint
type EndpointResult struct {
	URL     string
	Score   int
	Results []TestResult
}

// TestResult represents the result of a single test
type TestResult struct {
	TestName string
	Passed   bool
	Message  string
}

// RunTests runs all security tests concurrently and returns a slice of EndpointResult
func RunTests(config *Config) []EndpointResult {
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

	var wg sync.WaitGroup
	results := make([]EndpointResult, len(config.APIEndpoints))

	for i, endpoint := range config.APIEndpoints {
		wg.Add(7) // Updated to include Phase 2 tests
		results[i] = EndpointResult{URL: endpoint.URL, Score: 100}

		logging.Debug("Testing endpoint", map[string]interface{}{
			"url":    endpoint.URL,
			"method": endpoint.Method,
			"index":  i,
		})

		go func(e APIEndpoint, i int) {
			defer wg.Done()
			// Wait for rate limiter
			rateLimiter.Wait()
			defer rateLimiter.Done()
			
			if err := testAuth(e, config.Auth); err != nil {
				results[i].Results = append(results[i].Results, TestResult{TestName: "Auth Test", Passed: false, Message: err.Error()})
				results[i].Score -= 30
				logging.Warn("Auth test failed", map[string]interface{}{
					"url":   e.URL,
					"error": err.Error(),
				})
			} else {
				results[i].Results = append(results[i].Results, TestResult{TestName: "Auth Test", Passed: true, Message: "Auth Test Passed"})
				logging.Debug("Auth test passed", map[string]interface{}{
					"url": e.URL,
				})
			}
		}(endpoint, i)

		go func(e APIEndpoint, i int) {
			defer wg.Done()
			// Wait for rate limiter
			rateLimiter.Wait()
			defer rateLimiter.Done()
			
			if err := testHTTPMethod(e, config.Auth); err != nil {
				results[i].Results = append(results[i].Results, TestResult{TestName: "HTTP Method Test", Passed: false, Message: err.Error()})
				results[i].Score -= 20
				logging.Warn("HTTP method test failed", map[string]interface{}{
					"url":   e.URL,
					"error": err.Error(),
				})
			} else {
				results[i].Results = append(results[i].Results, TestResult{TestName: "HTTP Method Test", Passed: true, Message: "HTTP Method Test Passed"})
				logging.Debug("HTTP method test passed", map[string]interface{}{
					"url": e.URL,
				})
			}
		}(endpoint, i)

		go func(e APIEndpoint, i int) {
			defer wg.Done()
			// Wait for rate limiter
			rateLimiter.Wait()
			defer rateLimiter.Done()
			
			if err := testInjection(e, config.Auth, config.InjectionPayloads); err != nil {
				results[i].Results = append(results[i].Results, TestResult{TestName: "Injection Test", Passed: false, Message: err.Error()})
				results[i].Score -= 50
				logging.Warn("Injection test failed", map[string]interface{}{
					"url":   e.URL,
					"error": err.Error(),
				})
			} else {
				results[i].Results = append(results[i].Results, TestResult{TestName: "Injection Test", Passed: true, Message: "Injection Test Passed"})
				logging.Debug("Injection test passed", map[string]interface{}{
					"url": e.URL,
				})
			}
		}(endpoint, i)

		// Phase 2: XSS vulnerability detection
		go func(e APIEndpoint, i int) {
			defer wg.Done()
			// Wait for rate limiter
			rateLimiter.Wait()
			defer rateLimiter.Done()
			
			if err := testXSS(e, config.Auth, config.XSSPayloads); err != nil {
				results[i].Results = append(results[i].Results, TestResult{TestName: "XSS Test", Passed: false, Message: err.Error()})
				results[i].Score -= 40
				logging.Warn("XSS test failed", map[string]interface{}{
					"url":   e.URL,
					"error": err.Error(),
				})
			} else {
				results[i].Results = append(results[i].Results, TestResult{TestName: "XSS Test", Passed: true, Message: "XSS Test Passed"})
				logging.Debug("XSS test passed", map[string]interface{}{
					"url": e.URL,
				})
			}
		}(endpoint, i)

		// Phase 2: Header security analysis
		go func(e APIEndpoint, i int) {
			defer wg.Done()
			// Wait for rate limiter
			rateLimiter.Wait()
			defer rateLimiter.Done()
			
			if err := testHeaderSecurity(e, config.Auth, config.Headers); err != nil {
				results[i].Results = append(results[i].Results, TestResult{TestName: "Header Security Test", Passed: false, Message: err.Error()})
				results[i].Score -= 25
				logging.Warn("Header security test failed", map[string]interface{}{
					"url":   e.URL,
					"error": err.Error(),
				})
			} else {
				results[i].Results = append(results[i].Results, TestResult{TestName: "Header Security Test", Passed: true, Message: "Header Security Test Passed"})
				logging.Debug("Header security test passed", map[string]interface{}{
					"url": e.URL,
				})
			}
		}(endpoint, i)

		// Phase 2: Authentication bypass testing
		go func(e APIEndpoint, i int) {
			defer wg.Done()
			// Wait for rate limiter
			rateLimiter.Wait()
			defer rateLimiter.Done()
			
			if err := testAuthBypass(e, config.Auth); err != nil {
				results[i].Results = append(results[i].Results, TestResult{TestName: "Auth Bypass Test", Passed: false, Message: err.Error()})
				results[i].Score -= 35
				logging.Warn("Auth bypass test failed", map[string]interface{}{
					"url":   e.URL,
					"error": err.Error(),
				})
			} else {
				results[i].Results = append(results[i].Results, TestResult{TestName: "Auth Bypass Test", Passed: true, Message: "Auth Bypass Test Passed"})
				logging.Debug("Auth bypass test passed", map[string]interface{}{
					"url": e.URL,
				})
			}
		}(endpoint, i)

		// Phase 2: Parameter tampering detection
		go func(e APIEndpoint, i int) {
			defer wg.Done()
			// Wait for rate limiter
			rateLimiter.Wait()
			defer rateLimiter.Done()
			
			if err := testParameterTampering(e, config.Auth); err != nil {
				results[i].Results = append(results[i].Results, TestResult{TestName: "Parameter Tampering Test", Passed: false, Message: err.Error()})
				results[i].Score -= 30
				logging.Warn("Parameter tampering test failed", map[string]interface{}{
					"url":   e.URL,
					"error": err.Error(),
				})
			} else {
				results[i].Results = append(results[i].Results, TestResult{TestName: "Parameter Tampering Test", Passed: true, Message: "Parameter Tampering Test Passed"})
				logging.Debug("Parameter tampering test passed", map[string]interface{}{
					"url": e.URL,
				})
			}
		}(endpoint, i)
	}

	wg.Wait()

	logging.Info("Security tests completed", map[string]interface{}{
		"endpoints_count": len(results),
	})

	return results
}

func testAuth(endpoint APIEndpoint, auth Auth) error {
	logging.Debug("Testing authentication", map[string]interface{}{
		"url":    endpoint.URL,
		"method": endpoint.Method,
	})

	client := &http.Client{Timeout: 10 * time.Second}
	req, err := http.NewRequest(endpoint.Method, endpoint.URL, bytes.NewBufferString(endpoint.Body))
	if err != nil {
		logging.Error("Failed to create request", map[string]interface{}{
			"url":   endpoint.URL,
			"error": err.Error(),
		})
		return fmt.Errorf("failed to create request: %v", err)
	}

	req.SetBasicAuth(auth.Username, auth.Password)

	resp, err := client.Do(req)
	if err != nil {
		logging.Error("Request failed", map[string]interface{}{
			"url":   endpoint.URL,
			"error": err.Error(),
		})
		return fmt.Errorf("request failed: %v", err)
	}
	defer resp.Body.Close()

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

func testHTTPMethod(endpoint APIEndpoint, auth Auth) error {
	client := &http.Client{Timeout: 10 * time.Second}
	req, err := http.NewRequest(endpoint.Method, endpoint.URL, bytes.NewBufferString(endpoint.Body))
	if err != nil {
		return fmt.Errorf("failed to create request: %v", err)
	}
	req.SetBasicAuth(auth.Username, auth.Password)

	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("request failed: %v", err)
	}
	defer resp.Body.Close()

	// A 401 or 403 is an auth failure, not an HTTP method failure.
	// The auth test will catch these. For this test, we only care about other statuses.
	switch resp.StatusCode {
	case http.StatusOK, http.StatusCreated, http.StatusAccepted:
		return nil // Correct method used
	case http.StatusMethodNotAllowed, http.StatusNotFound:
		return HTTPMethodError{fmt.Sprintf("disallowed method returned status: %d", resp.StatusCode)}
	default:
		// Any other error is unexpected.
		return fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}
}

func testInjection(endpoint APIEndpoint, auth Auth, payloads []string) error {
	logging.Debug("Testing injection", map[string]interface{}{
		"url":           endpoint.URL,
		"method":        endpoint.Method,
		"payloads_count": len(payloads),
	})

	client := &http.Client{Timeout: 10 * time.Second}

	// First, send a request with no payload to get a baseline response
	baselineReq, err := http.NewRequest(endpoint.Method, endpoint.URL, bytes.NewBufferString(endpoint.Body))
	if err != nil {
		logging.Error("Failed to create baseline request", map[string]interface{}{
			"url":   endpoint.URL,
			"error": err.Error(),
		})
		return fmt.Errorf("failed to create baseline request: %v", err)
	}
	baselineReq.SetBasicAuth(auth.Username, auth.Password)

	baselineResp, err := client.Do(baselineReq)
	if err != nil {
		logging.Error("Baseline request failed", map[string]interface{}{
			"url":   endpoint.URL,
			"error": err.Error(),
		})
		return fmt.Errorf("baseline request failed: %v", err)
	}
	defer baselineResp.Body.Close()

	// If baseline is unauthorized, we can't continue the injection test.
	if baselineResp.StatusCode == http.StatusUnauthorized || baselineResp.StatusCode == http.StatusForbidden {
		logging.Warn("Cannot perform injection test", map[string]interface{}{
			"url":    endpoint.URL,
			"status": baselineResp.StatusCode,
		})
		return fmt.Errorf("cannot perform injection test: baseline request failed with status %d", baselineResp.StatusCode)
	}

	baselineBody, err := io.ReadAll(baselineResp.Body)
	if err != nil {
		logging.Error("Failed to read baseline response body", map[string]interface{}{
			"url":   endpoint.URL,
			"error": err.Error(),
		})
		return fmt.Errorf("failed to read baseline response body: %v", err)
	}

	for i, payload := range payloads {
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
		req.SetBasicAuth(auth.Username, auth.Password)

		resp, err := client.Do(req)
		if err != nil {
			logging.Error("Request failed", map[string]interface{}{
				"url":     endpoint.URL,
				"payload": payload,
				"error":   err.Error(),
			})
			return fmt.Errorf("request failed: %v", err)
		}
		defer resp.Body.Close()

		body, err := io.ReadAll(resp.Body)
		if err != nil {
			logging.Error("Failed to read response body", map[string]interface{}{
				"url":     endpoint.URL,
				"payload": payload,
				"error":   err.Error(),
			})
			return fmt.Errorf("failed to read response body: %v", err)
		}

		// Check for indicators of successful SQL injection
		if indicatorsOfSQLInjection(string(body), string(baselineBody)) {
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
	// List of common SQL error messages
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

	// Check if the response contains any SQL error messages
	for _, errorMsg := range sqlErrorMessages {
		if strings.Contains(responseBody, errorMsg) {
			return true
		}
	}

	// Check for significant differences in response length
	if len(responseBody) > len(baselineBody)*2 || len(responseBody) < len(baselineBody)/2 {
		return true
	}

	// Check for changes in response structure
	if strings.Count(responseBody, "{") != strings.Count(baselineBody, "{") ||
		strings.Count(responseBody, "}") != strings.Count(baselineBody, "}") {
		return true
	}

	return false
}

// testXSS tests for cross-site scripting vulnerabilities
func testXSS(endpoint APIEndpoint, auth Auth, payloads []string) error {
	logging.Debug("Testing XSS", map[string]interface{}{
		"url":           endpoint.URL,
		"method":        endpoint.Method,
		"payloads_count": len(payloads),
	})

	client := &http.Client{Timeout: 10 * time.Second}

	// First, send a request with no payload to get a baseline response
	baselineReq, err := http.NewRequest(endpoint.Method, endpoint.URL, bytes.NewBufferString(endpoint.Body))
	if err != nil {
		logging.Error("Failed to create baseline request", map[string]interface{}{
			"url":   endpoint.URL,
			"error": err.Error(),
		})
		return fmt.Errorf("failed to create baseline request: %v", err)
	}
	baselineReq.SetBasicAuth(auth.Username, auth.Password)

	baselineResp, err := client.Do(baselineReq)
	if err != nil {
		logging.Error("Baseline request failed", map[string]interface{}{
			"url":   endpoint.URL,
			"error": err.Error(),
		})
		return fmt.Errorf("baseline request failed: %v", err)
	}
	defer baselineResp.Body.Close()

	// If baseline is unauthorized, we can't continue the XSS test.
	if baselineResp.StatusCode == http.StatusUnauthorized || baselineResp.StatusCode == http.StatusForbidden {
		logging.Warn("Cannot perform XSS test", map[string]interface{}{
			"url":    endpoint.URL,
			"status": baselineResp.StatusCode,
		})
		return fmt.Errorf("cannot perform XSS test: baseline request failed with status %d", baselineResp.StatusCode)
	}

	baselineBody, err := io.ReadAll(baselineResp.Body)
	if err != nil {
		logging.Error("Failed to read baseline response body", map[string]interface{}{
			"url":   endpoint.URL,
			"error": err.Error(),
		})
		return fmt.Errorf("failed to read baseline response body: %v", err)
	}

	for i, payload := range payloads {
		logging.Debug("Testing XSS payload", map[string]interface{}{
			"url":     endpoint.URL,
			"payload": payload,
			"index":   i,
		})

		// Inject payload into the body
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
		req.SetBasicAuth(auth.Username, auth.Password)

		resp, err := client.Do(req)
		if err != nil {
			logging.Error("Request failed", map[string]interface{}{
				"url":     endpoint.URL,
				"payload": payload,
				"error":   err.Error(),
			})
			return fmt.Errorf("request failed: %v", err)
		}
		defer resp.Body.Close()

		body, err := io.ReadAll(resp.Body)
		if err != nil {
			logging.Error("Failed to read response body", map[string]interface{}{
				"url":     endpoint.URL,
				"payload": payload,
				"error":   err.Error(),
			})
			return fmt.Errorf("failed to read response body: %v", err)
		}

		// Check for indicators of successful XSS
		if indicatorsOfXSS(string(body), string(baselineBody), payload) {
			logging.Warn("Potential XSS detected", map[string]interface{}{
				"url":     endpoint.URL,
				"payload": payload,
			})
			return XSSError{fmt.Sprintf("potential XSS detected with payload: %s", payload)}
		}
	}
	return nil
}

// indicatorsOfXSS checks for indicators of successful XSS
func indicatorsOfXSS(responseBody, baselineBody, payload string) bool {
	// Check if the payload appears in the response without proper sanitization
	if strings.Contains(responseBody, payload) && !strings.Contains(baselineBody, payload) {
		// Check if the payload appears in a script context or as HTML
		scriptContext := strings.Contains(responseBody, fmt.Sprintf("<script>%s</script>", payload)) ||
			strings.Contains(responseBody, fmt.Sprintf("onload=\"%s\"", payload)) ||
			strings.Contains(responseBody, fmt.Sprintf("onerror=\"%s\"", payload)) ||
			strings.Contains(responseBody, fmt.Sprintf("onclick=\"%s\"", payload))
		
		if scriptContext {
			return true
		}
		
		// Check if payload appears in HTML tags
		if strings.Contains(responseBody, fmt.Sprintf("<%s>", payload)) ||
			strings.Contains(responseBody, fmt.Sprintf(">%s<", payload)) {
			return true
		}
	}
	
	return false
}

// testHeaderSecurity analyzes security headers
func testHeaderSecurity(endpoint APIEndpoint, auth Auth, customHeaders map[string]string) error {
	logging.Debug("Testing header security", map[string]interface{}{
		"url": endpoint.URL,
		"method": endpoint.Method,
	})

	client := &http.Client{Timeout: 10 * time.Second}
	req, err := http.NewRequest(endpoint.Method, endpoint.URL, bytes.NewBufferString(endpoint.Body))
	if err != nil {
		logging.Error("Failed to create request", map[string]interface{}{
			"url":   endpoint.URL,
			"error": err.Error(),
		})
		return fmt.Errorf("failed to create request: %v", err)
	}
	req.SetBasicAuth(auth.Username, auth.Password)

	// Add custom headers
	for key, value := range customHeaders {
		req.Header.Set(key, value)
	}

	resp, err := client.Do(req)
	if err != nil {
		logging.Error("Request failed", map[string]interface{}{
			"url":   endpoint.URL,
			"error": err.Error(),
		})
		return fmt.Errorf("request failed: %v", err)
	}
	defer resp.Body.Close()

	// Analyze security headers
	issues := []string{}

	// Check for missing security headers
	securityHeaders := map[string]string{
		"X-Content-Type-Options": "nosniff",
		"X-Frame-Options":        "DENY or SAMEORIGIN",
		"X-XSS-Protection":       "1; mode=block",
		"Strict-Transport-Security": "max-age=31536000; includeSubDomains",
		"Content-Security-Policy": "policy directives",
	}

	for header, recommended := range securityHeaders {
		if resp.Header.Get(header) == "" {
			issues = append(issues, fmt.Sprintf("Missing recommended security header: %s (recommended value: %s)", header, recommended))
		}
	}

	// Check for insecure headers
	insecureHeaders := []string{
		"X-Powered-By",
		"Server",
	}

	for _, header := range insecureHeaders {
		if resp.Header.Get(header) != "" {
			issues = append(issues, fmt.Sprintf("Insecure information disclosure header: %s (%s)", header, resp.Header.Get(header)))
		}
	}

	// Check CORS settings
	accessControlAllowOrigin := resp.Header.Get("Access-Control-Allow-Origin")
	if accessControlAllowOrigin == "*" {
		issues = append(issues, "Insecure CORS policy: Access-Control-Allow-Origin set to wildcard (*)")
	}

	// Check cookie security
	setCookie := resp.Header.Values("Set-Cookie")
	for _, cookie := range setCookie {
		if !strings.Contains(cookie, "Secure") {
			issues = append(issues, "Cookie missing Secure attribute: " + cookie)
		}
		if !strings.Contains(cookie, "HttpOnly") {
			issues = append(issues, "Cookie missing HttpOnly attribute: " + cookie)
		}
		if !strings.Contains(cookie, "SameSite") {
			issues = append(issues, "Cookie missing SameSite attribute: " + cookie)
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

// testAuthBypass tests for authentication bypass vulnerabilities
func testAuthBypass(endpoint APIEndpoint, auth Auth) error {
	logging.Debug("Testing authentication bypass", map[string]interface{}{
		"url": endpoint.URL,
		"method": endpoint.Method,
		"has_auth": auth.Username != "" && auth.Password != "",
	})

	client := &http.Client{Timeout: 10 * time.Second}

	// Test 1: Request without authentication
	req1, err := http.NewRequest(endpoint.Method, endpoint.URL, bytes.NewBufferString(endpoint.Body))
	if err != nil {
		logging.Error("Failed to create request without auth", map[string]interface{}{
			"url":   endpoint.URL,
			"error": err.Error(),
		})
		return fmt.Errorf("failed to create request without auth: %v", err)
	}

	resp1, err := client.Do(req1)
	if err != nil {
		logging.Error("Request without auth failed", map[string]interface{}{
			"url":   endpoint.URL,
			"error": err.Error(),
		})
		return fmt.Errorf("request without auth failed: %v", err)
	}
	defer resp1.Body.Close()

	// If we can access the endpoint without auth when it should require it, that's a bypass
	if resp1.StatusCode == http.StatusOK || resp1.StatusCode == http.StatusCreated || resp1.StatusCode == http.StatusAccepted {
		logging.Warn("Authentication bypass detected", map[string]interface{}{
			"url": endpoint.URL,
			"status_without_auth": resp1.StatusCode,
		})
		return AuthBypassError{fmt.Sprintf("authentication bypass detected: endpoint accessible without authentication (status: %d)", resp1.StatusCode)}
	}

	// Test 2: Request with modified authentication tokens
	if auth.Username != "" && auth.Password != "" {
		// Test with invalid credentials
		req2, err := http.NewRequest(endpoint.Method, endpoint.URL, bytes.NewBufferString(endpoint.Body))
		if err != nil {
			logging.Error("Failed to create request with invalid auth", map[string]interface{}{
				"url":   endpoint.URL,
				"error": err.Error(),
			})
			return fmt.Errorf("failed to create request with invalid auth: %v", err)
		}
		req2.SetBasicAuth("invalid_user", "invalid_pass")

		resp2, err := client.Do(req2)
		if err != nil {
			logging.Error("Request with invalid auth failed", map[string]interface{}{
				"url":   endpoint.URL,
				"error": err.Error(),
			})
			return fmt.Errorf("request with invalid auth failed: %v", err)
		}
		defer resp2.Body.Close()

		// If invalid credentials still grant access, that's a bypass
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
		"X-Forwarded-For": "127.0.0.1",
		"X-Original-URL":  endpoint.URL,
		"X-Rewrite-URL":   endpoint.URL,
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
	
	// Add bypass headers
	for key, value := range bypassHeaders {
		req3.Header.Set(key, value)
	}

	resp3, err := client.Do(req3)
	if err != nil {
		logging.Error("Request with bypass headers failed", map[string]interface{}{
			"url":   endpoint.URL,
			"error": err.Error(),
		})
		return fmt.Errorf("request with bypass headers failed: %v", err)
	}
	defer resp3.Body.Close()

	// If adding bypass headers grants access, that's a bypass
	if resp3.StatusCode == http.StatusOK || resp3.StatusCode == http.StatusCreated || resp3.StatusCode == http.StatusAccepted {
		logging.Warn("Authentication bypass with headers", map[string]interface{}{
			"url": endpoint.URL,
			"status_with_bypass_headers": resp3.StatusCode,
		})
		return AuthBypassError{fmt.Sprintf("authentication bypass detected: endpoint accessible with bypass headers (status: %d)", resp3.StatusCode)}
	}

	return nil
}

func GenerateDetailedReport(results []EndpointResult) {
	fmt.Println("\nAPI Security Scan Detailed Report")
	fmt.Println("==================================")

	for _, result := range results {
		fmt.Printf("\nEndpoint: %s\n", result.URL)
		fmt.Printf("Overall Score: %d/100\n", result.Score)
		fmt.Println("Test Results:")

		// Sort test results for consistent output
		sort.Slice(result.Results, func(i, j int) bool {
			return result.Results[i].TestName < result.Results[j].TestName
		})

		for _, testResult := range result.Results {
			status := "PASSED"
			if !testResult.Passed {
				status = "FAILED"
			}
			fmt.Printf("- %s: %s\n", testResult.TestName, status)
			fmt.Printf("  Details: %s\n", formatTestMessage(testResult.Message, result.URL))
		}

		fmt.Println("Risk Assessment:")
		fmt.Println(generateRiskAssessment(result))
		fmt.Println("------------------------")
	}

	fmt.Println("\nOverall Security Assessment:")
	fmt.Println(generateOverallAssessment(results))
}

func formatTestMessage(message string, url string) string {
	prefix := fmt.Sprintf("Test Failed for %s:", url)
	return strings.TrimSpace(strings.TrimPrefix(message, prefix))
}

func generateRiskAssessment(result EndpointResult) string {
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
			}
		}
	}

	if len(risks) == 0 {
		return "No significant risks detected."
	}
	return strings.Join(risks, "\n")
}

func generateOverallAssessment(results []EndpointResult) string {
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
	averageScore := totalScore / len(results)

	assessment := fmt.Sprintf("Average Security Score: %d/100\n", averageScore)
	assessment += fmt.Sprintf("Critical Vulnerabilities Detected: %d\n\n", criticalVulnerabilities)

	if averageScore >= 90 {
		assessment += "Overall security posture is strong, but continuous monitoring is recommended."
	} else if averageScore >= 70 {
		assessment += "Moderate security risks detected. Address identified vulnerabilities promptly."
	} else {
		assessment += "Significant security risks identified. Immediate action is required to improve API security."
	}

	return assessment
}

// GenerateJSONReport generates a JSON formatted report
func GenerateJSONReport(results []EndpointResult) {
	fmt.Println("{")
	fmt.Printf("  \"scan_results\": [")
	for i, result := range results {
		if i > 0 {
			fmt.Println(",")
		}
		fmt.Printf("\n    {")
		fmt.Printf("\n      \"endpoint\": \"%s\",", result.URL)
		fmt.Printf("\n      \"score\": %d,", result.Score)
		fmt.Printf("\n      \"tests\": [")
		for j, testResult := range result.Results {
			if j > 0 {
				fmt.Printf(",")
			}
			fmt.Printf("\n        {")
			fmt.Printf("\n          \"name\": \"%s\",", testResult.TestName)
			fmt.Printf("\n          \"passed\": %t,", testResult.Passed)
			fmt.Printf("\n          \"message\": \"%s\"", testResult.Message)
			fmt.Printf("\n        }")
		}
		fmt.Printf("\n      ],")
		fmt.Printf("\n      \"risk_assessment\": \"%s\"", generateRiskAssessment(result))
		fmt.Printf("\n    }")
	}
	fmt.Println("\n  ],")
	fmt.Printf("  \"overall_assessment\": \"%s\"\n", generateOverallAssessment(results))
	fmt.Println("}")
}

// GenerateHTMLReport generates an HTML formatted report
func GenerateHTMLReport(results []EndpointResult) {
	fmt.Println("<!DOCTYPE html>")
	fmt.Println("<html>")
	fmt.Println("<head>")
	fmt.Println("  <title>API Security Scan Report</title>")
	fmt.Println("  <style>")
	fmt.Println("    body { font-family: Arial, sans-serif; margin: 20px; }")
	fmt.Println("    .header { background-color: #f0f0f0; padding: 10px; border-radius: 5px; }")
	fmt.Println("    .endpoint { margin: 20px 0; padding: 15px; border: 1px solid #ccc; border-radius: 5px; }")
	fmt.Println("    .passed { color: green; }")
	fmt.Println("    .failed { color: red; }")
	fmt.Println("    .score-high { color: green; font-weight: bold; }")
	fmt.Println("    .score-medium { color: orange; font-weight: bold; }")
	fmt.Println("    .score-low { color: red; font-weight: bold; }")
	fmt.Println("  </style>")
	fmt.Println("</head>")
	fmt.Println("<body>")
	fmt.Println("  <h1>API Security Scan Detailed Report</h1>")
	
	for _, result := range results {
		fmt.Printf("  <div class=\"endpoint\">\n")
		fmt.Printf("    <h2>Endpoint: %s</h2>\n", result.URL)
		
		// Score with color coding
		scoreClass := "score-low"
		if result.Score >= 90 {
			scoreClass = "score-high"
		} else if result.Score >= 70 {
			scoreClass = "score-medium"
		}
		fmt.Printf("    <p><strong>Overall Score:</strong> <span class=\"%s\">%d/100</span></p>\n", scoreClass, result.Score)
		
		fmt.Println("    <h3>Test Results:</h3>")
		fmt.Println("    <ul>")
		for _, testResult := range result.Results {
			statusClass := "passed"
			statusText := "PASSED"
			if !testResult.Passed {
				statusClass = "failed"
				statusText = "FAILED"
			}
			fmt.Printf("      <li><strong>%s:</strong> <span class=\"%s\">%s</span> - %s</li>\n", 
				testResult.TestName, statusClass, statusText, testResult.Message)
		}
		fmt.Println("    </ul>")
		
		fmt.Println("    <h3>Risk Assessment:</h3>")
		fmt.Printf("    <p>%s</p>\n", generateRiskAssessment(result))
		fmt.Println("  </div>")
	}
	
	fmt.Println("  <div class=\"endpoint\">")
	fmt.Println("    <h2>Overall Security Assessment</h2>")
	fmt.Printf("    <p>%s</p>\n", generateOverallAssessment(results))
	fmt.Println("  </div>")
	
	fmt.Println("</body>")
	fmt.Println("</html>")
}

// GenerateCSVReport generates a CSV formatted report
func GenerateCSVReport(results []EndpointResult) {
	// CSV header
	fmt.Println("Endpoint,Score,Test Name,Passed,Message,Risk Assessment")
	
	for _, result := range results {
		// Escape quotes in fields
		endpoint := strings.ReplaceAll(result.URL, "\"", "\"\"")
		riskAssessment := strings.ReplaceAll(generateRiskAssessment(result), "\"", "\"\"")
		
		for _, testResult := range result.Results {
			testName := strings.ReplaceAll(testResult.TestName, "\"", "\"\"")
			message := strings.ReplaceAll(testResult.Message, "\"", "\"\"")
			passed := "true"
			if !testResult.Passed {
				passed = "false"
			}
			
			fmt.Printf("\"%s\",%d,\"%s\",%s,\"%s\",\"%s\"\n", 
				endpoint, result.Score, testName, passed, message, riskAssessment)
		}
	}
	
	// Add overall assessment
	overall := strings.ReplaceAll(generateOverallAssessment(results), "\"", "\"\"")
	fmt.Printf("\"OVERALL\",,\"\",\"\",\"\",\"%s\"\n", overall)
}

// GenerateXMLReport generates an XML formatted report
func GenerateXMLReport(results []EndpointResult) {
	fmt.Println("<?xml version=\"1.0\" encoding=\"UTF-8\"?>")
	fmt.Println("<api_security_scan>")
	fmt.Println("  <scan_results>")
	
	for _, result := range results {
		fmt.Println("    <endpoint>")
		fmt.Printf("      <url>%s</url>\n", result.URL)
		fmt.Printf("      <score>%d</score>\n", result.Score)
		fmt.Println("      <tests>")
		
		for _, testResult := range result.Results {
			fmt.Println("        <test>")
			fmt.Printf("          <name>%s</name>\n", testResult.TestName)
			fmt.Printf("          <passed>%t</passed>\n", testResult.Passed)
			fmt.Printf("          <message>%s</message>\n", testResult.Message)
			fmt.Println("        </test>")
		}
		
		fmt.Println("      </tests>")
		fmt.Printf("      <risk_assessment>%s</risk_assessment>\n", generateRiskAssessment(result))
		fmt.Println("    </endpoint>")
	}
	
	fmt.Println("  </scan_results>")
	fmt.Printf("  <overall_assessment>%s</overall_assessment>\n", generateOverallAssessment(results))
	fmt.Println("</api_security_scan>")
}

// testParameterTampering tests for parameter manipulation vulnerabilities
func testParameterTampering(endpoint APIEndpoint, auth Auth) error {
	logging.Debug("Testing parameter tampering", map[string]interface{}{
		"url": endpoint.URL,
		"method": endpoint.Method,
	})

	client := &http.Client{Timeout: 10 * time.Second}

	// Test 1: Modify numeric parameters in the body
	if strings.Contains(endpoint.Body, "\"key\":") {
		// Try replacing values with different numeric values
		modifiedBody := strings.Replace(endpoint.Body, "\"value\"", "\"12345\"", -1)
		
		req, err := http.NewRequest(endpoint.Method, endpoint.URL, bytes.NewBufferString(modifiedBody))
		if err != nil {
			logging.Error("Failed to create request with modified parameters", map[string]interface{}{
				"url":   endpoint.URL,
				"error": err.Error(),
			})
			return fmt.Errorf("failed to create request with modified parameters: %v", err)
		}
		req.SetBasicAuth(auth.Username, auth.Password)

		resp, err := client.Do(req)
		if err != nil {
			logging.Error("Request with modified parameters failed", map[string]interface{}{
				"url":   endpoint.URL,
				"error": err.Error(),
			})
			return fmt.Errorf("request with modified parameters failed: %v", err)
		}
		defer resp.Body.Close()

		// If changing parameters still grants the same access, that might indicate a vulnerability
		if resp.StatusCode == http.StatusOK || resp.StatusCode == http.StatusCreated || resp.StatusCode == http.StatusAccepted {
			// This is normal behavior, not necessarily a vulnerability
			logging.Debug("Parameter modification test completed", map[string]interface{}{
				"url": endpoint.URL,
				"status": resp.StatusCode,
			})
		}
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
		req.SetBasicAuth(auth.Username, auth.Password)

		resp, err := client.Do(req)
		if err != nil {
			logging.Error("Request with extra parameters failed", map[string]interface{}{
				"url":   endpoint.URL,
				"error": err.Error(),
			})
			return fmt.Errorf("request with extra parameters failed: %v", err)
		}
		defer resp.Body.Close()

		// If adding extra parameters still grants access, that might indicate poor input validation
		logging.Debug("Extra parameter test completed", map[string]interface{}{
			"url": endpoint.URL,
			"status_with_extra_params": resp.StatusCode,
		})
	}

	// Test 3: Test for IDOR (Insecure Direct Object Reference) by trying to access different resource IDs
	// This is a simplified test - in a real implementation, this would be more sophisticated
	if strings.Contains(endpoint.URL, "/") {
		// Try to access a different resource by modifying the URL
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
			req.SetBasicAuth(auth.Username, auth.Password)

			resp, err := client.Do(req)
			if err != nil {
				logging.Error("Request with modified URL failed", map[string]interface{}{
					"url":   modifiedURL,
					"error": err.Error(),
				})
				return fmt.Errorf("request with modified URL failed: %v", err)
			}
			defer resp.Body.Close()

			// If we can access a different resource, that might indicate an IDOR vulnerability
			if resp.StatusCode == http.StatusOK || resp.StatusCode == http.StatusCreated || resp.StatusCode == http.StatusAccepted {
				logging.Warn("Potential IDOR detected", map[string]interface{}{
					"original_url": endpoint.URL,
					"modified_url": modifiedURL,
					"status": resp.StatusCode,
				})
				return ParameterTamperingError{fmt.Sprintf("potential IDOR detected: able to access %s (status: %d)", modifiedURL, resp.StatusCode)}
			}
		}
	}

	return nil
}
