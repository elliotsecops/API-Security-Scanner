package main

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"net/http"
	"sort"
	"strings"
	"sync"
	"time"
)

// Config represents the overall configuration
type Config struct {
	APIEndpoints      []APIEndpoint `yaml:"api_endpoints"`
	Auth              Auth          `yaml:"auth"`
	InjectionPayloads []string      `yaml:"injection_payloads"`
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

func (e AuthError) Error() string       { return e.message }
func (e HTTPMethodError) Error() string { return e.message }
func (e InjectionError) Error() string  { return e.message }

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

// runTests runs all security tests concurrently and returns a slice of EndpointResult
func runTests(config *Config) []EndpointResult {
	var wg sync.WaitGroup
	results := make([]EndpointResult, len(config.APIEndpoints))

	for i, endpoint := range config.APIEndpoints {
		wg.Add(3)
		results[i] = EndpointResult{URL: endpoint.URL, Score: 100}

		go func(e APIEndpoint, i int) {
			defer wg.Done()
			if err := testAuth(e, config.Auth); err != nil {
				results[i].Results = append(results[i].Results, TestResult{TestName: "Auth Test", Passed: false, Message: err.Error()})
				results[i].Score -= 30
			} else {
				results[i].Results = append(results[i].Results, TestResult{TestName: "Auth Test", Passed: true, Message: "Auth Test Passed"})
			}
		}(endpoint, i)

		go func(e APIEndpoint, i int) {
			defer wg.Done()
			if err := testHTTPMethod(e); err != nil {
				results[i].Results = append(results[i].Results, TestResult{TestName: "HTTP Method Test", Passed: false, Message: err.Error()})
				results[i].Score -= 20
			} else {
				results[i].Results = append(results[i].Results, TestResult{TestName: "HTTP Method Test", Passed: true, Message: "HTTP Method Test Passed"})
			}
		}(endpoint, i)

		go func(e APIEndpoint, i int) {
			defer wg.Done()
			if err := testInjection(e, config.InjectionPayloads); err != nil {
				results[i].Results = append(results[i].Results, TestResult{TestName: "Injection Test", Passed: false, Message: err.Error()})
				results[i].Score -= 50
			} else {
				results[i].Results = append(results[i].Results, TestResult{TestName: "Injection Test", Passed: true, Message: "Injection Test Passed"})
			}
		}(endpoint, i)
	}

	wg.Wait()
	return results
}

func testAuth(endpoint APIEndpoint, auth Auth) error {
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

	switch resp.StatusCode {
	case http.StatusOK, http.StatusCreated, http.StatusAccepted:
		return nil
	case http.StatusUnauthorized:
		return AuthError{"authentication failed: incorrect credentials"}
	case http.StatusForbidden:
		return AuthError{"authentication failed: access forbidden"}
	default:
		return fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}
}

func testHTTPMethod(endpoint APIEndpoint) error {
	client := &http.Client{Timeout: 10 * time.Second}
	req, err := http.NewRequest(endpoint.Method, endpoint.URL, bytes.NewBufferString(endpoint.Body))
	if err != nil {
		return fmt.Errorf("failed to create request: %v", err)
	}

	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("request failed: %v", err)
	}
	defer resp.Body.Close()

	switch resp.StatusCode {
	case http.StatusOK, http.StatusCreated, http.StatusAccepted, http.StatusUnauthorized:
		return nil // Consider 401 as "expected" for protected endpoints
	default:
		return HTTPMethodError{fmt.Sprintf("unexpected status code: %d", resp.StatusCode)}
	}
}

func testInjection(endpoint APIEndpoint, payloads []string) error {
	client := &http.Client{Timeout: 10 * time.Second}

	// First, send a request with no payload to get a baseline response
	baselineReq, err := http.NewRequest(endpoint.Method, endpoint.URL, bytes.NewBufferString(endpoint.Body))
	if err != nil {
		return fmt.Errorf("failed to create baseline request: %v", err)
	}

	baselineResp, err := client.Do(baselineReq)
	if err != nil {
		return fmt.Errorf("baseline request failed: %v", err)
	}
	defer baselineResp.Body.Close()

	baselineBody, err := ioutil.ReadAll(baselineResp.Body)
	if err != nil {
		return fmt.Errorf("failed to read baseline response body: %v", err)
	}

	for _, payload := range payloads {
		reqBody := fmt.Sprintf(endpoint.Body, payload)
		req, err := http.NewRequest(endpoint.Method, endpoint.URL, bytes.NewBufferString(reqBody))
		if err != nil {
			return fmt.Errorf("failed to create request: %v", err)
		}

		resp, err := client.Do(req)
		if err != nil {
			return fmt.Errorf("request failed: %v", err)
		}
		defer resp.Body.Close()

		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return fmt.Errorf("failed to read response body: %v", err)
		}

		// Check for indicators of successful SQL injection
		if indicatorsOfSQLInjection(string(body), string(baselineBody)) {
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

func generateDetailedReport(results []EndpointResult) {
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
			fmt.Printf("  Details: %s\n", formatTestMessage(testResult.Message))
		}

		fmt.Println("Risk Assessment:")
		fmt.Println(generateRiskAssessment(result))
		fmt.Println("------------------------")
	}

	fmt.Println("\nOverall Security Assessment:")
	fmt.Println(generateOverallAssessment(results))
}

func formatTestMessage(message string) string {
	return strings.TrimSpace(strings.TrimPrefix(message, "Test Failed for http://127.0.0.1:5000/post:"))
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
