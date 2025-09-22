# Contributing to API Security Scanner

Thank you for your interest in contributing to the API Security Scanner! This document provides guidelines and instructions for contributors.

## 📋 Table of Contents

1. [Code of Conduct](#code-of-conduct)
2. [Getting Started](#getting-started)
3. [Development Setup](#development-setup)
4. [Coding Standards](#coding-standards)
5. [Testing Guidelines](#testing-guidelines)
6. [Pull Request Process](#pull-request-process)
7. [Issue Reporting](#issue-reporting)
8. [Documentation Guidelines](#documentation-guidelines)
9. [Release Process](#release-process)
10. [Community Guidelines](#community-guidelines)

## 🤝 Code of Conduct

### Our Pledge

We as members, contributors, and leaders pledge to make participation in our community a harassment-free experience for everyone, regardless of age, body size, visible or invisible disability, ethnicity, sex characteristics, gender identity and expression, level of experience, education, socio-economic status, nationality, personal appearance, race, religion, or sexual identity and orientation.

### Our Standards

Examples of behavior that contributes to creating a positive environment include:

- Using welcoming and inclusive language
- Being respectful of differing viewpoints and experiences
- Gracefully accepting constructive criticism
- Focusing on what is best for the community
- Showing empathy towards other community members

### Unacceptable Behavior

Examples of unacceptable behavior include:

- Trolling, insulting/derogatory comments, and personal or political attacks
- Public or private harassment
- Publishing others' private information, such as a physical or electronic address, without explicit permission
- Other conduct which could reasonably be considered inappropriate in a professional setting

### Enforcement

Instances of abusive, harassing, or otherwise unacceptable behavior may be reported by contacting the project team at security@elliotsecops.com. All complaints will be reviewed and investigated and will result in a response that is deemed necessary and appropriate to the circumstances. The project team is obligated to maintain confidentiality with regard to the reporter of an incident.

## 🚀 Getting Started

### Prerequisites

- Go 1.24 or later
- Git
- Basic understanding of Go programming
- Familiarity with REST APIs and security concepts

### Fork and Clone

1. **Fork the repository**
   ```bash
   # Fork on GitHub, then clone your fork
   git clone https://github.com/your-username/API-Security-Scanner.git
   cd API-Security-Scanner
   ```

2. **Add upstream remote**
   ```bash
   git remote add upstream https://github.com/elliotsecops/API-Security-Scanner.git
   ```

3. **Create a development branch**
   ```bash
   git checkout -b feature/amazing-feature
   ```

## ⚙️ Development Setup

### Environment Setup

```bash
# Clone the repository
git clone https://github.com/elliotsecops/API-Security-Scanner.git
cd API-Security-Scanner

# Install Go dependencies
go mod tidy

# Install development dependencies
go get github.com/golang-jwt/jwt/v5
go get golang.org/x/oauth2
go get github.com/sirupsen/logrus
go get github.com/gorilla/websocket

# Install development tools
go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
go install github.com/swaggo/swag/cmd/swag@latest
```

### Project Structure

```
API-Security-Scanner/
├── main.go                 # Application entry point
├── config/                 # Configuration management
│   ├── config.go
│   └── types.go
├── scanner/                # Core scanner functionality
│   ├── scanner.go
│   ├── tests.go
│   └── types.go
├── tenant/                 # Multi-tenant management
│   ├── tenant.go
│   └── manager.go
├── siem/                   # SIEM integration
│   ├── siem.go
│   ├── syslog.go
│   └── types.go
├── auth/                   # Authentication
│   ├── auth.go
│   └── advanced.go
├── metrics/                # Performance metrics
│   ├── metrics.go
│   ├── dashboard.go
│   └── types.go
├── server/                 # HTTP server
│   ├── server.go
│   ├── handlers.go
│   └── middleware.go
├── logging/                # Logging system
│   └── logging.go
├── utils/                  # Utility functions
│   ├── http.go
│   ├── validation.go
│   └── crypto.go
├── static/                 # Static files for dashboard
├── tests/                  # Test files
│   ├── integration/
│   ├── unit/
│   └── e2e/
├── docs/                   # Documentation
├── examples/               # Example configurations
├── scripts/                # Helper scripts
└── CONTRIBUTING.md         # This file
```

### Building and Testing

```bash
# Build the application
go build -o api-security-scanner

# Run tests
go test ./...

# Run tests with coverage
go test -cover ./...

# Run specific test
go test -run TestFunctionName ./scanner

# Run benchmarks
go test -bench=. ./...

# Run integration tests
go test -tags=integration ./...

# Run end-to-end tests
go test -tags=e2e ./...
```

### Development Tools

```bash
# Lint code
golangci-lint run

# Format code
go fmt ./...

# Vet code
go vet ./...

# Generate documentation
swag init

# Check for vulnerabilities
go list -json -deps ./... | nancy sleuth
```

## 📏 Coding Standards

### Go Coding Standards

1. **Follow Go Conventions**
   - Use `go fmt` for formatting
   - Follow [Effective Go](https://golang.org/doc/effective_go.html)
   - Use [Go Code Review Comments](https://github.com/golang/go/wiki/CodeReviewComments)

2. **Naming Conventions**
   ```go
   // Package names: short, concise, lowercase
   package scanner

   // Function names: mixedCase, public functions start with capital
   func ScanEndpoint(url string) (*ScanResult, error) {
       // ...
   }

   // Variable names: mixedCase, descriptive
   func processResults(results []ScanResult) error {
       var totalVulnerabilities int
       // ...
   }

   // Constants: UPPER_SNAKE_CASE
   const DefaultTimeout = 30 * time.Second

   // Struct names: PascalCase
   type ScanConfig struct {
       Endpoints []Endpoint `json:"endpoints"`
       Timeout   time.Duration `json:"timeout"`
   }
   ```

3. **Error Handling**
   ```go
   // Good error handling
   func ScanAPI(config *ScanConfig) (*ScanResult, error) {
       if config == nil {
           return nil, fmt.Errorf("config cannot be nil")
       }

       result, err := executeScan(config)
       if err != nil {
           return nil, fmt.Errorf("failed to execute scan: %w", err)
       }

       return result, nil
   }

   // Wrap errors with context
   func validateConfig(config *ScanConfig) error {
       if len(config.Endpoints) == 0 {
           return fmt.Errorf("no endpoints configured")
       }
       return nil
   }
   ```

4. **Documentation**
   ```go
   // Package scanner provides API security scanning functionality
   package scanner

   // ScanResult represents the result of a security scan
   type ScanResult struct {
       URL            string                 `json:"url"`             // Target URL
       Score          int                    `json:"score"`           // Security score (0-100)
       Vulnerabilities []Vulnerability      `json:"vulnerabilities"` // Found vulnerabilities
       Timestamp      time.Time              `json:"timestamp"`      // Scan timestamp
   }

   // NewScanner creates a new security scanner with the given configuration
   func NewScanner(config *Config) (*Scanner, error) {
       // Implementation
   }

   // Scan performs a security scan on the configured endpoints
   // It returns a ScanResult containing the security assessment and any vulnerabilities found
   func (s *Scanner) Scan(ctx context.Context) (*ScanResult, error) {
       // Implementation
   }
   ```

### Security Standards

1. **Input Validation**
   ```go
   func validateEndpoint(url string) error {
       if url == "" {
           return fmt.Errorf("URL cannot be empty")
       }

       parsedURL, err := url.Parse(url)
       if err != nil {
           return fmt.Errorf("invalid URL format: %w", err)
       }

       if parsedURL.Scheme != "http" && parsedURL.Scheme != "https" {
           return fmt.Errorf("only HTTP and HTTPS protocols are supported")
       }

       return nil
   }
   ```

2. **Safe HTTP Requests**
   ```go
   func makeSafeRequest(url string, timeout time.Duration) (*http.Response, error) {
       client := &http.Client{
           Timeout: timeout,
           Transport: &http.Transport{
               TLSClientConfig: &tls.Config{
                   InsecureSkipVerify: false, // Always verify certificates
               },
           },
       }

       req, err := http.NewRequest("GET", url, nil)
       if err != nil {
           return nil, err
       }

       // Set security headers
       req.Header.Set("User-Agent", "API-Security-Scanner/4.0")
       req.Header.Set("Accept", "application/json")

       return client.Do(req)
   }
   ```

3. **Secure Configuration**
   ```go
   type Config struct {
       APIKey         string `json:"-"`                // Don't expose in JSON
       DatabaseURL    string `json:"database_url"`    // Database connection
       LogLevel       string `json:"log_level"`       // Logging level

       // Sensitive fields should not be logged
       sensitiveFields map[string]bool `json:"-"`
   }

   func (c *Config) Sanitize() *Config {
       // Return a copy with sensitive fields removed
       sanitized := *c
       sanitized.APIKey = "***REDACTED***"
       return &sanitized
   }
   ```

### Testing Standards

1. **Unit Tests**
   ```go
   package scanner

   import (
       "testing"
       "github.com/stretchr/testify/assert"
       "github.com/stretchr/testify/require"
   )

   func TestValidateEndpoint(t *testing.T) {
       tests := []struct {
           name    string
           url     string
           wantErr bool
           errMsg  string
       }{
           {
               name:    "valid HTTPS URL",
               url:     "https://api.example.com/users",
               wantErr: false,
           },
           {
               name:    "invalid URL",
               url:     "not-a-url",
               wantErr: true,
               errMsg:  "invalid URL format",
           },
           {
               name:    "empty URL",
               url:     "",
               wantErr: true,
               errMsg:  "URL cannot be empty",
           },
       }

       for _, tt := range tests {
           t.Run(tt.name, func(t *testing.T) {
               err := validateEndpoint(tt.url)

               if tt.wantErr {
                   require.Error(t, err)
                   assert.Contains(t, err.Error(), tt.errMsg)
               } else {
                   require.NoError(t, err)
               }
           })
       }
   }
   ```

2. **Integration Tests**
   ```go
   package integration

   import (
       "context"
       "net/http"
       "net/http/httptest"
       "testing"

       "github.com/elliotsecops/API-Security-Scanner/scanner"
   )

   func TestScannerIntegration(t *testing.T) {
       // Create test server
       server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
           w.Header().Set("Content-Type", "application/json")
           w.WriteHeader(http.StatusOK)
           w.Write([]byte(`{"status": "ok"}`))
       }))
       defer server.Close()

       // Create scanner config
       config := &scanner.Config{
           Endpoints: []scanner.Endpoint{
               {
                   URL:    server.URL,
                   Method: "GET",
               },
           },
       }

       // Create scanner
       s, err := scanner.NewScanner(config)
       require.NoError(t, err)

       // Run scan
       result, err := s.Scan(context.Background())
       require.NoError(t, err)

       // Validate results
       assert.NotNil(t, result)
       assert.Greater(t, result.Score, 0)
   }
   ```

3. **Benchmark Tests**
   ```go
   func BenchmarkScanner(b *testing.B) {
       config := &scanner.Config{
           Endpoints: []scanner.Endpoint{
               {
                   URL:    "https://httpbin.org/get",
                   Method: "GET",
               },
           },
       }

       s, err := scanner.NewScanner(config)
       require.NoError(b, err)

       b.ResetTimer()
       for i := 0; i < b.N; i++ {
           _, err := s.Scan(context.Background())
           require.NoError(b, err)
       }
   }
   ```

## 🧪 Testing Guidelines

### Test Organization

```
tests/
├── unit/                    # Unit tests
│   ├── scanner_test.go
│   ├── config_test.go
│   └── auth_test.go
├── integration/             # Integration tests
│   ├── scanner_integration_test.go
│   ├── siem_integration_test.go
│   └── api_integration_test.go
├── e2e/                     # End-to-end tests
│   ├── full_scan_test.go
│   └── ui_test.go
└── testdata/               # Test data files
    ├── config_valid.yaml
    ├── config_invalid.yaml
    └── mock_responses.json
```

### Test Coverage

- Aim for **80%+ code coverage**
- Critical paths should have **95%+ coverage**
- Security-related code must have **100% test coverage**

### Test Data Management

```go
// Use test fixtures for consistent test data
func TestConfigValidation(t *testing.T) {
   tests := loadTestFixtures("config_test_fixtures.json")

   for _, test := range tests {
       t.Run(test.Name, func(t *testing.T) {
           // Test logic
       })
   }
}

// Use test servers for HTTP testing
func createTestServer() *httptest.Server {
   return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
       switch r.URL.Path {
       case "/health":
           w.WriteHeader(http.StatusOK)
           w.Write([]byte(`{"status": "healthy"}`))
       default:
           w.WriteHeader(http.StatusNotFound)
       }
   }))
}
```

## 🔄 Pull Request Process

### PR Workflow

1. **Create Feature Branch**
   ```bash
   git checkout -b feature/your-feature-name
   ```

2. **Make Changes**
   - Follow coding standards
   - Write tests for new functionality
   - Update documentation
   - Ensure all tests pass

3. **Commit Changes**
   ```bash
   git add .
   git commit -m "Add amazing feature: description of changes"
   ```

4. **Push to Fork**
   ```bash
   git push origin feature/your-feature-name
   ```

5. **Create Pull Request**
   - Use descriptive title
   - Provide detailed description
   - Link to relevant issues
   - Add screenshots if applicable

6. **Address Feedback**
   - Respond to review comments
   - Make requested changes
   - Keep PR history clean

### Pull Request Template

```markdown
## Description
<!-- Describe your changes in detail -->

## Type of Change
<!-- What type of change is this? -->
- [ ] Bug fix (non-breaking change which fixes an issue)
- [ ] New feature (non-breaking change which adds functionality)
- [ ] Breaking change (fix or feature that would cause existing functionality to not work as expected)
- [ ] Documentation update
- [ ] Refactoring (no functional changes)
- [ ] Performance improvement
- [ ] Security improvement

## How Has This Been Tested?
<!-- Please describe in detail how you tested your changes -->
- [ ] Unit tests written and passing
- [ ] Integration tests written and passing
- [ ] Manual testing performed
- [ ] Edge cases considered and tested

## Test Environment
- OS: [e.g. Ubuntu 20.04]
- Go version: [e.g. 1.24.0]
- Browser: [e.g. Chrome 90.0]

## Checklist:
- [ ] My code follows the style guidelines of this project
- [ ] I have performed a self-review of my own code
- [ ] I have commented my code, particularly in hard-to-understand areas
- [ ] I have made corresponding changes to the documentation
- [ ] My changes generate no new warnings
- [ ] I have added tests that prove my fix is effective or that my feature works
- [ ] New and existing unit tests pass locally with my changes
- [ ] Any dependent changes have been merged and published in downstream modules

## Related Issues
<!-- Link to related GitHub issues -->
Closes #123
Related to #456
```

### PR Review Process

1. **Automated Checks**
   - Code formatting
   - Test coverage
   - Linting
   - Security scanning

2. **Peer Review**
   - At least one maintainer approval required
   - Focus on code quality and security
   - Verify documentation is updated

3. **CI/CD Pipeline**
   - Run all tests
   - Build artifacts
   - Security scanning
   - Performance benchmarks

## 🐛 Issue Reporting

### Bug Reports

1. **Use the Bug Report Template**
   ```markdown
   ## Bug Description
   <!-- Clear and concise description of the bug -->

   ## Steps to Reproduce
   1. Step one
   2. Step two
   3. See error

   ## Expected Behavior
   <!-- What you expected to happen -->

   ## Actual Behavior
   <!-- What actually happened -->

   ## Environment
   - OS: [e.g. Ubuntu 20.04]
   - Go version: [e.g. 1.24.0]
   - Scanner version: [e.g. 4.0.0]
   - Configuration: [attach config file if possible]

   ## Additional Context
   <!-- Any other context about the problem -->
   ```

2. **Provide Enough Information**
   - Error messages and stack traces
   - Configuration files (redact sensitive data)
   - Log files
   - Steps to reproduce
   - Expected vs actual behavior

3. **Search Existing Issues**
   - Check if the bug has already been reported
   - Add to existing issues if found

### Feature Requests

1. **Use the Feature Request Template**
   ```markdown
   ## Feature Description
   <!-- Clear and concise description of the feature -->

   ## Problem Statement
   <!-- What problem does this feature solve? -->

   ## Proposed Solution
   <!-- Describe the solution you'd like -->

   ## Alternatives Considered
   <!-- Describe any alternative solutions or features you've considered -->

   ## Additional Context
   <!-- Any other context or screenshots about the feature request -->
   ```

2. **Provide Context**
   - Use case description
   - User personas affected
   - Business value
   - Technical considerations

### Security Vulnerability Reporting

**Do not report security vulnerabilities publicly.**

Contact the security team directly:
- Email: security@elliotsecops.com
- PGP Key: Available on request

Include:
- Vulnerability description
- Proof of concept
- Potential impact
- Suggested remediation

## 📚 Documentation Guidelines

### Code Documentation

1. **Package Documentation**
   ```go
   // Package scanner provides comprehensive API security scanning capabilities
   // including SQL injection, XSS, and authentication bypass detection.
   //
   // Features:
   //   - Multi-tenant support
   //   - SIEM integration
   //   - Advanced authentication methods
   //   - Real-time monitoring
   package scanner
   ```

2. **Type Documentation**
   ```go
   // Scanner represents a security scanner with configuration and state
   type Scanner struct {
       config     *Config           // Scanner configuration
       httpClient *http.Client      // HTTP client for requests
       metrics    *MetricsCollector // Metrics collector
       logger     *logrus.Logger    // Logger instance
   }

   // ScanResult contains the results of a security scan
   type ScanResult struct {
       ID             string            `json:"id"`              // Unique scan identifier
       TargetURL      string            `json:"target_url"`      // Target endpoint URL
       Score          int               `json:"score"`           // Security score (0-100)
       Vulnerabilities []Vulnerability  `json:"vulnerabilities"` // Found vulnerabilities
       Timestamp      time.Time         `json:"timestamp"`       // Scan timestamp
       Metadata       map[string]string `json:"metadata"`        // Additional metadata
   }
   ```

3. **Function Documentation**
   ```go
   // NewScanner creates and initializes a new security scanner with the provided configuration.
   // It validates the configuration, sets up the HTTP client, and initializes metrics collection.
   //
   // Parameters:
   //   - config: Scanner configuration containing endpoints, authentication, and settings
   //
   // Returns:
   //   - *Scanner: Initialized scanner instance
   //   - error: Error if configuration is invalid or initialization fails
   //
   // Example:
   //   config := &Config{Endpoints: []Endpoint{{URL: "https://api.example.com"}}}
   //   scanner, err := NewScanner(config)
   //   if err != nil {
   //       log.Fatal(err)
   //   }
   func NewScanner(config *Config) (*Scanner, error) {
       // Implementation
   }
   ```

### User Documentation

1. **README Updates**
   - Update installation instructions
   - Add new features to feature list
   - Update configuration examples
   - Update version information

2. **API Documentation**
   - Document new endpoints
   - Update request/response examples
   - Add authentication requirements
   - Include error handling examples

3. **Configuration Documentation**
   - Add new configuration options
   - Update example configurations
   - Document environment variables
   - Include migration guides

### Documentation Structure

```
docs/
├── user-guide/              # User documentation
│   ├── installation.md
│   ├── configuration.md
│   └── troubleshooting.md
├── developer-guide/          # Developer documentation
│   ├── architecture.md
│   ├── contributing.md
│   └── testing.md
├── api/                      # API documentation
│   ├── rest-api.md
│   ├── websocket-api.md
│   └── authentication.md
├── examples/                 # Example configurations and scripts
└── release-notes/            # Release notes and changelogs
```

## 📦 Release Process

### Version Management

- Follow [Semantic Versioning](https://semver.org/)
- Use `MAJOR.MINOR.PATCH` format
- Pre-release versions use `-alpha.1`, `-beta.1`, `-rc.1` suffixes

### Release Checklist

1. **Pre-Release**
   - [ ] All tests passing
   - [ ] Code coverage requirements met
   - [ ] Documentation updated
   - [ ] CHANGELOG updated
   - [ ] Version incremented
   - [ ] Release notes prepared

2. **Build and Test**
   - [ ] Build for all platforms
   - [ ] Run integration tests
   - [ ] Perform security scan
   - [ ] Test upgrade process

3. **Release**
   - [ ] Create Git tag
   - [ ] Build release artifacts
   - [ ] Create GitHub release
   - [ ] Update documentation
   - [ ] Publish to package repositories

4. **Post-Release**
   - [ ] Monitor for issues
   - [ ] Update project boards
   - [ ] Plan next release
   - [ ] Communicate release to users

### Release Commands

```bash
# Create release branch
git checkout -b release/v4.1.0

# Update version in files
# - main.go
# - config.go
# - README.md

# Commit version changes
git commit -m "Bump version to v4.1.0"

# Create tag
git tag -a v4.1.0 -m "Release v4.1.0"

# Push changes
git push origin main
git push origin v4.1.0

# Build release artifacts
./scripts/build-release.sh

# Create GitHub release
gh release create v4.1.0 \
  --title "Release v4.1.0" \
  --notes-file RELEASE_NOTES.md \
  dist/*.tar.gz dist/*.zip
```

## 🌍 Community Guidelines

### Communication Channels

- **GitHub Issues**: Bug reports and feature requests
- **GitHub Discussions**: General discussions and Q&A
- **Email**: security@elliotsecops.com (security vulnerabilities only)
- **Discord**: Community chat (coming soon)

### Community Expectations

1. **Be Respectful**
   - Treat others with respect and professionalism
   - Assume good intent
   - Be inclusive and welcoming

2. **Be Helpful**
   - Help newcomers get started
   - Share knowledge and experience
   - Provide constructive feedback

3. **Be Patient**
   - Response times may vary
   - Maintainers have limited time
   - Complex issues take time to resolve

### Contribution Recognition

- Contributors are recognized in release notes
- Top contributors are featured in the README
- Opportunities for project maintainer roles
- Professional references and recommendations

## 🏆 Recognition

### Contributor Levels

1. **Contributor**: Submitted PRs that were merged
2. **Active Contributor**: Multiple merged PRs over time
3. **Core Contributor**: Regular contributions to core features
4. **Maintainer**: Trusted contributor with commit access

### Benefits

- GitHub badges and recognition
- Opportunities for speaking engagements
- Professional networking
- Resume/CV enhancement
- Learning and skill development

---

Thank you for contributing to the API Security Scanner! Your contributions help make this project better for everyone.

If you have any questions about contributing, please don't hesitate to reach out to our team or open an issue on GitHub.