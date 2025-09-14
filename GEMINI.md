# Project Overview

This project is a Go-based API security scanner. It is designed to test API endpoints for common security vulnerabilities. The scanner reads a configuration file (`config.yaml`) to define the target API endpoints, authentication credentials, and payloads for injection attacks.

The scanner performs the following tests:
- **Authentication Tests:** Verifies that endpoints are properly secured and that authentication credentials are required.
- **HTTP Method Tests:** Checks for improper handling of HTTP methods.
- **SQL Injection Tests:** Attempts to inject malicious SQL queries to identify potential vulnerabilities.

After running the tests, the scanner generates a detailed report that includes an overall security score, a breakdown of passed and failed tests, a risk assessment, and an overall security assessment.

# Building and Running

## Prerequisites
- Go 1.16 or higher

## Running the Scanner
To run the scanner, use the following command:
```bash
go run main.go scanner.go
```

## Testing
To run the tests, use the following command:
```bash
go test
```

# Development Conventions

The project follows standard Go conventions. The code is organized into two main files: `main.go` and `scanner.go`. `main.go` handles the main application logic, while `scanner.go` contains the core scanning functionality.

The project uses the `gopkg.in/yaml.v2` library for parsing the `config.yaml` file.

The tests are located in `main_test.go` and `scanner_test.go` and can be run using the standard `go test` command.
