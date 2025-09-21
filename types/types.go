package types

// APIEndpoint represents a single API endpoint configuration
type APIEndpoint struct {
	URL    string `yaml:"url"`
	Method string `yaml:"method"`
	Body   string `yaml:"body,omitempty"`
}

// TestResult represents the result of a single security test
type TestResult struct {
	TestName string `json:"test_name"`
	Passed   bool   `json:"passed"`
	Message  string `json:"message"`
}

// EndpointResult represents the result of testing a single endpoint
type EndpointResult struct {
	URL     string        `json:"url"`
	Score   int           `json:"score"`
	Results []TestResult  `json:"results"`
}