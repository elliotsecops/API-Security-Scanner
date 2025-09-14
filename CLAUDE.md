# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

This is a Go-based API Security Scanner tool designed to perform security testing on API endpoints. The tool conducts three main security tests: authentication testing, HTTP method validation, and SQL injection detection.

## Development Commands

### Building and Running
- `go run main.go` - Run the main security scanner
- `go run automate.go` - Run automation script for Git operations and CI setup
- `go build` - Build the application
- `go build -o api-security-scanner main.go` - Build with custom output name

### Testing
- `go test ./...` - Run all tests
- `go test -v ./...` - Run tests with verbose output
- `go test -run TestIntegration ./...` - Run integration tests
- `go test -run TestPerformAuthTest ./...` - Run specific auth tests

### Code Quality
- `go fmt ./...` - Format all Go code
- `go vet ./...` - Check for potential issues
- `go mod tidy` - Clean up dependencies

## Architecture

### Core Components

1. **main.go** - Entry point with configuration loading and test orchestration
2. **scanner.go** - Core security testing logic with test implementations
3. **automate.go** - Git automation and CI/CD workflow setup
4. **config.yaml** - Configuration file defining API endpoints and test parameters

### Security Testing Implementation

The scanner performs three types of security tests concurrently using goroutines:

- **Authentication Testing** (`testAuth`): Validates basic auth credentials and handles HTTP 401/403 responses
- **HTTP Method Testing** (`testHTTPMethod`): Validates that endpoints respond appropriately to different HTTP methods
- **SQL Injection Testing** (`testInjection`): Tests for SQL injection vulnerabilities using payload-based testing

### Configuration Structure

```yaml
api_endpoints:  # List of endpoints to test
  - url: "endpoint_url"
    method: "HTTP_METHOD"
    body: "request_body"
auth:           # Basic auth credentials
  username: "username"
  password: "password"
injection_payloads:  # SQL injection test payloads
  - "malicious_payload"
```

### Test Result Scoring

Each endpoint starts with a score of 100/100. Points are deducted for failed tests:
- Authentication failures: -30 points
- HTTP method failures: -20 points  
- Injection vulnerabilities: -50 points

### Error Handling

The project uses custom error types for different test failures:
- `AuthError` - Authentication-related failures
- `HTTPMethodError` - HTTP method validation failures
- `InjectionError` - SQL injection detection

## Key Files and Their Purposes

- **config.yaml**: Main configuration with endpoints and test parameters
- **scanner.go**: Contains all security testing logic and report generation
- **main_test.go**: Integration tests with mock servers
- **scanner_test.go**: Unit tests for individual security functions
- **automate.go**: Git automation and GitHub Actions workflow creation

## Testing Approach

The project uses httptest for HTTP testing with mock servers. Tests validate:
- Authentication success/failure scenarios
- HTTP method compliance
- SQL injection detection accuracy
- Report generation functionality

## Important Notes

- The tool is designed for defensive security testing only
- Configuration should target your own API endpoints for testing
- Test payloads are defensive in nature, designed to detect vulnerabilities rather than exploit them
- All HTTP requests include proper timeouts (10 seconds) to prevent hanging