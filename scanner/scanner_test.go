package scanner

import (
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestRunTests(t *testing.T) {
	// Universal mock server for all test cases
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Auth check for all requests
		user, pass, ok := r.BasicAuth()
		if !ok || user != "admin" || pass != "password" {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		// Endpoint for HTTP method test
		if r.URL.Path == "/method-test" {
			if r.Method != "POST" {
				http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
				return
			}
		}

		// Endpoint for injection test
		if r.URL.Path == "/injection-test" {
			bodyBytes, _ := io.ReadAll(r.Body)
			body := string(bodyBytes)
			if strings.Contains(body, "' OR '1'='1") {
				http.Error(w, "You have an error in your SQL syntax", http.StatusInternalServerError)
				return
			}
		}

		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status": "ok"}`))
	}))
	defer server.Close()

	// Define test cases
	testCases := []struct {
		name                string
		config              *Config
		expectedScore       int
		expectedTestsFailed []string
	}{
		{
			name: "Successful Run - No Vulnerabilities",
			config: &Config{
				APIEndpoints: []APIEndpoint{{URL: server.URL + "/clean", Method: "GET"}},
				Auth:         Auth{Username: "admin", Password: "password"},
			},
			expectedScore: 100,
		},
		{
			name: "Authentication Failure",
			config: &Config{
				APIEndpoints: []APIEndpoint{{URL: server.URL + "/auth-fail", Method: "GET"}},
				Auth:         Auth{Username: "admin", Password: "wrongpassword"},
			},
			expectedScore:       0,
			expectedTestsFailed: []string{"Auth Test", "HTTP Method Test", "Injection Test"},
		},
		{
			name: "HTTP Method Failure",
			config: &Config{
				APIEndpoints: []APIEndpoint{{URL: server.URL + "/method-test", Method: "GET"}},
				Auth:         Auth{Username: "admin", Password: "password"},
			},
			expectedScore:       80,
			expectedTestsFailed: []string{"HTTP Method Test"},
		},
		{
			name: "SQL Injection Failure",
			config: &Config{
				APIEndpoints:      []APIEndpoint{{URL: server.URL + "/injection-test", Method: "POST", Body: `{"query": "%s"}`}},
				Auth:              Auth{Username: "admin", Password: "password"},
				InjectionPayloads: []string{"' OR '1'='1"},
			},
			expectedScore:       50,
			expectedTestsFailed: []string{"Injection Test"},
		},
		{
			name: "Unreachable Server",
			config: &Config{
				APIEndpoints: []APIEndpoint{{URL: "http://localhost:12345", Method: "GET"}},
				Auth:         Auth{Username: "admin", Password: "password"},
			},
			expectedScore:       0, // All tests should fail
			expectedTestsFailed: []string{"Auth Test", "HTTP Method Test", "Injection Test"},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			results := RunTests(tc.config)
			if len(results) != 1 {
				t.Fatalf("Expected 1 result, got %d", len(results))
			}
			result := results[0]

			if result.Score != tc.expectedScore {
				t.Errorf("Expected score %d, got %d", tc.expectedScore, result.Score)
			}

			failedTests := make(map[string]bool)
			for _, res := range result.Results {
				if !res.Passed {
					failedTests[res.TestName] = true
				}
			}

			if len(failedTests) != len(tc.expectedTestsFailed) {
				t.Errorf("Expected %d failed tests, but %d failed", len(tc.expectedTestsFailed), len(failedTests))
			}

			for _, testName := range tc.expectedTestsFailed {
				if !failedTests[testName] {
					t.Errorf("Expected test '%s' to fail, but it passed", testName)
				}
			}
		})
	}
}
