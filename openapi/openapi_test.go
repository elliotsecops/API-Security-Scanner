package openapi

import (
	"os"
	"path/filepath"
	"testing"

	"api-security-scanner/types"
)

const sampleSpec = `openapi: 3.0.0
info:
  title: Sample API
  version: 1.0.0
servers:
  - url: http://api.example.com
paths:
  /users:
    get:
      responses:
        '200':
          description: OK
    post:
      requestBody:
        required: false
        content:
          application/json:
            schema:
              type: object
      responses:
        '201':
          description: Created
  /reports:
    get:
      parameters:
        - name: filter
          in: query
          schema:
            type: string
      responses:
        '200':
          description: OK
`

func writeTempSpec(t *testing.T, content string) string {
	t.Helper()
	dir := t.TempDir()
	path := filepath.Join(dir, "spec.yaml")
	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		t.Fatalf("failed to write temp spec: %v", err)
	}
	return path
}

func newIntegration(t *testing.T) *OpenAPIIntegration {
	t.Helper()
	specPath := writeTempSpec(t, sampleSpec)
	integ, err := NewOpenAPIIntegration(specPath)
	if err != nil {
		t.Fatalf("failed to create integration: %v", err)
	}
	return integ
}

func TestGenerateEndpointsFromSpec(t *testing.T) {
	integ := newIntegration(t)

	endpoints := integ.GenerateEndpointsFromSpec()
	want := []types.APIEndpoint{
		{URL: "http://api.example.com/users", Method: "GET"},
		{URL: "http://api.example.com/users", Method: "POST"},
		{URL: "http://api.example.com/reports", Method: "GET"},
	}

	if len(endpoints) != len(want) {
		t.Fatalf("expected %d endpoints, got %d", len(want), len(endpoints))
	}

	for _, expected := range want {
		if !containsEndpoint(endpoints, expected) {
			t.Fatalf("missing expected endpoint %+v in %#v", expected, endpoints)
		}
	}
}

func TestValidateEndpointAgainstSpec(t *testing.T) {
	integ := newIntegration(t)

	if err := integ.ValidateEndpointAgainstSpec(types.APIEndpoint{URL: "http://api.example.com/users", Method: "GET"}); err != nil {
		t.Fatalf("expected GET /users to be valid, got %v", err)
	}

	err := integ.ValidateEndpointAgainstSpec(types.APIEndpoint{URL: "http://api.example.com/users", Method: "DELETE"})
	if err == nil {
		t.Fatal("expected DELETE /users to be invalid")
	}

	err = integ.ValidateEndpointAgainstSpec(types.APIEndpoint{URL: "http://api.example.com/unknown", Method: "GET"})
	if err == nil {
		t.Fatal("expected GET /unknown to be invalid")
	}
}

func TestGenerateTestCasesFromSpec(t *testing.T) {
	integ := newIntegration(t)

	base := types.APIEndpoint{URL: "http://api.example.com/reports", Method: "GET"}
	tests := integ.GenerateTestCasesFromSpec(base)
	if len(tests) != 1 {
		t.Fatalf("expected 1 test for query parameters, got %d", len(tests))
	}
	if tests[0].URL == base.URL {
		t.Fatalf("expected query injection to modify URL, got %s", tests[0].URL)
	}

	post := types.APIEndpoint{URL: "http://api.example.com/users", Method: "POST"}
	postTests := integ.GenerateTestCasesFromSpec(post)
	if len(postTests) != 2 {
		t.Fatalf("expected 2 request body variations, got %d", len(postTests))
	}

	seenMalformed := false
	seenInjection := false
	for _, tc := range postTests {
		switch tc.Body {
		case "{":
			seenMalformed = true
		case `{"test": "' OR '1'='1"}`:
			seenInjection = true
		default:
			t.Fatalf("unexpected generated body %q", tc.Body)
		}
	}
	if !seenMalformed || !seenInjection {
		t.Fatal("expected both malformed and injection bodies to be generated")
	}
}

func TestExtractPathFromURL(t *testing.T) {
	integ := newIntegration(t)

	if got := integ.extractPathFromURL("http://api.example.com/users"); got != "/users" {
		t.Fatalf("expected /users, got %q", got)
	}

	if got := integ.extractPathFromURL("https://other.example.com/service"); got != "/service" {
		t.Fatalf("expected /service for foreign host, got %q", got)
	}
}

func containsEndpoint(list []types.APIEndpoint, target types.APIEndpoint) bool {
	for _, ep := range list {
		if ep.URL == target.URL && ep.Method == target.Method {
			return true
		}
	}
	return false
}
