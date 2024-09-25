package main

import (
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
)

func TestIntegration(t *testing.T) {
	// Create a temporary config file for testing
	configContent := `
api_endpoints:
  - url: "http://example.com/api/v1/resource"
    method: "GET"
  - url: "http://example.com/api/v1/resource"
    method: "POST"
    body: '{"key": "value"}'

auth:
  username: "admin"
  password: "password"

injection_payloads:
  - "' OR '1'='1"
  - "'; DROP TABLE users;--"
`
	configFile, err := ioutil.TempFile("", "config.yaml")
	if err != nil {
		t.Fatalf("Failed to create temp config file: %v", err)
	}
	defer os.Remove(configFile.Name())

	if _, err := configFile.Write([]byte(configContent)); err != nil {
		t.Fatalf("Failed to write to temp config file: %v", err)
	}
	configFile.Close()

	// Create a mock server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		username, password, ok := r.BasicAuth()
		if !ok || username != "admin" || password != "password" {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	// Update the config file with the mock server URL
	configContent = `
api_endpoints:
  - url: "` + server.URL + `"
    method: "GET"
  - url: "` + server.URL + `"
    method: "POST"
    body: '{"key": "value"}'

auth:
  username: "admin"
  password: "password"

injection_payloads:
  - "' OR '1'='1"
  - "'; DROP TABLE users;--"
`
	if err := ioutil.WriteFile(configFile.Name(), []byte(configContent), 0644); err != nil {
		t.Fatalf("Failed to update temp config file: %v", err)
	}

	// Run the tests
	main()
}
