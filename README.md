# API Security Scanner

[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)
[![Go Version](https://img.shields.io/badge/Go-1.21+-blue.svg)](https://golang.org/dl/)
[![Build Status](https://img.shields.io/badge/Build-Passing-green.svg)](https://github.com/elliotsecops/API-Security-Scanner)

A comprehensive, professional-grade API security testing tool written in Go. Designed to help developers and security professionals identify vulnerabilities in their REST APIs through automated security testing.

## 🎯 Overview

The API Security Scanner performs automated security assessments of API endpoints by testing for common vulnerabilities including:

- **Authentication Bypass Detection**
- **SQL Injection Vulnerabilities**
- **HTTP Method Validation**
- **Security Header Analysis**
- **Parameter Tampering Detection**
- **Cross-Site Scripting (XSS) Vulnerabilities**
- **Authentication Bypass Testing**

Built with performance and reliability in mind, the scanner uses concurrent execution to efficiently test multiple endpoints simultaneously while providing detailed security reports.

## ✨ Key Features

### 🔒 Security Testing
- **Authentication Testing**: Validates basic auth credentials and identifies access control issues
- **SQL Injection Detection**: Comprehensive payload-based testing for SQL injection vulnerabilities
- **HTTP Method Validation**: Ensures proper HTTP method handling and prevents method-based attacks
- **XSS Vulnerability Detection**: Tests for cross-site scripting vulnerabilities using common payloads
- **Header Security Analysis**: Analyzes HTTP response headers for security issues
- **Authentication Bypass Testing**: Tests for authentication vulnerabilities
- **Parameter Tampering Detection**: Tests for parameter manipulation vulnerabilities
- **Concurrent Execution**: High-performance parallel testing of multiple endpoints
- **Detailed Reporting**: Comprehensive security assessments with risk analysis

### 🚀 Performance & Reliability
- **Fast Execution**: Concurrent testing with configurable rate limiting
- **Robust Error Handling**: Graceful handling of network timeouts and connection issues
- **Memory Efficient**: Optimized for large-scale API testing
- **Configurable Timeouts**: Prevents hanging requests with configurable timeouts

### 📊 Reporting & Output
- **Multiple Output Formats**: Text, JSON, HTML, CSV, and XML output formats
- **Risk Assessment**: Automated risk scoring and remediation recommendations
- **Structured Logging**: Configurable logging with multiple formats (text, JSON)
- **Score-based Metrics**: 100-point scoring system for security posture assessment

### ⚙️ Configuration & Management
- **Configuration Validation**: Schema validation with detailed error messages
- **Rate Limiting**: Configurable request rate and concurrency limits
- **Endpoint Reachability Testing**: Pre-flight validation of API endpoints

## 🛠️ Installation

### Prerequisites

- **Go 1.21 or higher** - [Download Go](https://golang.org/dl/)
- **Git** - [Download Git](https://git-scm.com/downloads)

### Quick Install

```bash
# Clone the repository
git clone https://github.com/elliotsecops/API-Security-Scanner.git
cd API-Security-Scanner

# Build the application
go build -o api-security-scanner

# Or run directly with Go
go run main.go
```

### Docker Installation (Optional)

```bash
# Build Docker image
docker build -t api-security-scanner .

# Run with Docker
docker run --rm -v $(pwd)/config.yaml:/app/config.yaml api-security-scanner
```

## 🚀 Usage

### Basic Usage

```bash
# Run with default configuration
./api-security-scanner

# Run with custom configuration file
./api-security-scanner -config /path/to/custom-config.yaml
```

### Command Line Options

| Option | Description | Default |
|--------|-------------|---------|
| `-config` | Path to configuration file | `config.yaml` |
| `-output` | Output format (text, json, html, csv, xml) | `text` |
| `-validate` | Validate configuration only, don't run tests | `false` |
| `-log-level` | Log level (debug, info, warn, error) | `info` |
| `-log-format` | Log format (text, json) | `text` |
| `-timeout` | Request timeout in seconds | `10` |
| `-verbose` | Enable verbose logging | `false` |

### Configuration

The scanner uses a YAML configuration file to define test parameters:

```yaml
# API endpoints to test
api_endpoints:
  - url: "https://api.example.com/v1/users"
    method: "GET"
  - url: "https://api.example.com/v1/data"
    method: "POST"
    body: '{"query": "value"}'

# Authentication credentials
auth:
  username: "admin"
  password: "securepassword"

# SQL injection test payloads
injection_payloads:
  - "' OR '1'='1"
  - "'; DROP TABLE users;--"
  - "1' OR '1'='1"
  - "admin'--"

# XSS test payloads
xss_payloads:
  - "<script>alert('XSS')</script>"
  - "'><script>alert('XSS')</script>"
  - "<img src=x onerror=alert('XSS')>"

# Rate limiting configuration
rate_limiting:
  requests_per_second: 10
  max_concurrent_requests: 5

# Custom headers
headers:
  "User-Agent": "API-Security-Scanner/1.0"
  "X-Scanner": "true"
```

## 📋 Configuration Reference

### API Endpoints Configuration

```yaml
api_endpoints:
  - url: "https://api.example.com/endpoint"  # Required: API endpoint URL
    method: "GET"                           # Required: HTTP method (GET, POST, PUT, DELETE, etc.)
    body: '{"param": "value"}'             # Optional: Request body for POST/PUT requests
```

### Authentication Configuration

```yaml
auth:
  username: "your_username"     # Required: Username for basic authentication
  password: "your_password"     # Required: Password for basic authentication
```

### Injection Payloads Configuration

```yaml
injection_payloads:
  - "' OR '1'='1"                    # Classic SQL injection
  - "'; DROP TABLE users;--"         # SQL DROP statement
  - "1' OR '1'='1"                   # Numeric SQL injection
  - "admin'--"                       # Comment-based SQL injection
```

### XSS Payloads Configuration

```yaml
xss_payloads:
  - "<script>alert('XSS')</script>"      # Basic script tag injection
  - "'><script>alert('XSS')</script>"    # Attribute breaking injection
  - "<img src=x onerror=alert('XSS')>"   # Image tag injection
  - "javascript:alert('XSS')"            # JavaScript URI injection
```

### Headers Configuration

```yaml
headers:
  "User-Agent": "API-Security-Scanner/2.0"
  "X-Test-Header": "test-value"
  "Accept": "application/json"
```

## 📊 Sample Output

### Successful Scan

```
API Security Scan Detailed Report
==================================

Endpoint: https://api.example.com/v1/users
Overall Score: 100/100
Test Results:
- Auth Test: PASSED
  Details: Authentication successful
- HTTP Method Test: PASSED
  Details: Method validation successful
- Injection Test: PASSED
  Details: No injection vulnerabilities detected
- XSS Test: PASSED
  Details: No XSS vulnerabilities detected
- Header Security Test: PASSED
  Details: All security headers present
- Auth Bypass Test: PASSED
  Details: Authentication properly enforced
- Parameter Tampering Test: PASSED
  Details: Parameter validation successful

Risk Assessment:
No significant risks detected.

Endpoint: https://api.example.com/v1/data
Overall Score: 25/100
Test Results:
- Auth Test: PASSED
  Details: Authentication successful
- HTTP Method Test: PASSED
  Details: Method validation successful
- Injection Test: FAILED
  Details: Potential SQL injection detected with payload: ' OR '1'='1
- XSS Test: FAILED
  Details: Potential XSS detected with payload: <script>alert('XSS')</script>
- Header Security Test: FAILED
  Details: Missing security headers: X-Frame-Options, X-Content-Type-Options
- Auth Bypass Test: FAILED
  Details: Endpoint accessible without authentication
- Parameter Tampering Test: PASSED
  Details: Parameter validation successful

Risk Assessment:
- SQL injection vulnerabilities pose a significant data breach risk.
- Cross-site scripting vulnerabilities could allow malicious script execution.
- Insecure headers may expose sensitive information or lack security protections.
- Authentication bypass vulnerabilities could allow unauthorized access to protected resources.

Overall Security Assessment:
Average Security Score: 62/100
Critical Vulnerabilities Detected: 2

Moderate security risks detected. Address identified vulnerabilities promptly.
```

## 🎯 Security Scoring

The scanner uses a 100-point scoring system:

- **Starting Score**: 100/100 for each endpoint
- **Authentication Failure**: -30 points
- **HTTP Method Failure**: -20 points
- **Injection Vulnerability**: -50 points
- **XSS Vulnerability**: -40 points
- **Header Security Issues**: -25 points
- **Auth Bypass Vulnerability**: -35 points
- **Parameter Tampering Vulnerability**: -30 points

### Risk Levels

| Score Range | Risk Level | Action Required |
|-------------|------------|-----------------|
| 90-100 | Low | Monitor regularly |
| 70-89 | Medium | Address within 30 days |
| 50-69 | High | Address within 7 days |
| 0-49 | Critical | Immediate action required |

## 🧪 Testing Methodology

### Authentication Testing

The scanner tests authentication by:

1. Sending requests with configured credentials
2. Analyzing HTTP response codes (200, 401, 403)
3. Verifying proper access control mechanisms
4. Testing for authentication bypass vulnerabilities

### SQL Injection Testing

The scanner tests for SQL injection by:

1. Sending baseline requests to establish normal response patterns
2. Testing with various SQL injection payloads
3. Analyzing response differences and error messages
4. Looking for indicators of successful injection

### HTTP Method Testing

The scanner validates HTTP method handling by:

1. Testing supported HTTP methods for each endpoint
2. Verifying proper handling of disallowed methods
3. Checking for method-based access control issues
4. Ensuring REST compliance

### XSS Vulnerability Testing

The scanner tests for XSS vulnerabilities by:

1. Sending baseline requests to establish normal response patterns
2. Testing with various XSS payloads
3. Analyzing response content for unsanitized payload reflection
4. Looking for indicators of successful XSS execution

### Header Security Analysis

The scanner analyzes HTTP response headers by:

1. Checking for presence of recommended security headers
2. Identifying insecure information disclosure headers
3. Validating cookie security attributes
4. Analyzing CORS policy configurations

### Authentication Bypass Testing

The scanner tests for authentication bypass by:

1. Sending requests without authentication credentials
2. Testing with invalid credentials
3. Checking for common bypass headers
4. Analyzing response codes for unauthorized access

### Parameter Tampering Detection

The scanner tests for parameter tampering by:

1. Modifying parameter values in requests
2. Adding extra parameters to requests
3. Testing for IDOR (Insecure Direct Object Reference)
4. Analyzing response behavior for parameter changes

## 🔧 Advanced Configuration

### Rate Limiting

To prevent overwhelming target APIs, configure rate limiting:

```yaml
rate_limiting:
  requests_per_second: 10
  max_concurrent_requests: 5
```

### Custom Headers

Add custom headers for requests:

```yaml
headers:
  "User-Agent": "API-Security-Scanner/1.0"
  "X-API-Key": "your-api-key"
  "Accept": "application/json"
```

### Proxy Configuration

Configure proxy settings for corporate environments:

```yaml
proxy:
  url: "http://proxy.company.com:8080"
  username: "proxy-user"
  password: "proxy-password"
```

## 🚨 Best Practices

### For Security Teams

1. **Run in Staging First**: Always test against staging environments before production
2. **Schedule Scans**: Run during off-peak hours to minimize impact
3. **Monitor Resources**: Watch for high CPU/memory usage during large scans
4. **Review Results**: Analyze findings and prioritize critical vulnerabilities

### For Development Teams

1. **Integrate into CI/CD**: Add security scans to your deployment pipeline
2. **Fix Findings Promptly**: Address security issues before deployment
3. **Update Configuration**: Regularly update API endpoints and test payloads
4. **Document Exceptions**: Maintain documentation for accepted risks

### For API Owners

1. **Understand Scope**: Clearly define which APIs can be tested
2. **Coordinate Testing**: Schedule scans with API maintenance windows
3. **Review Results**: Work with security teams to understand findings
4. **Implement Fixes**: Prioritize and deploy security patches

## 🐛 Troubleshooting

### Common Issues

#### Connection Timeouts

```bash
# Increase timeout
./api-security-scanner -timeout 30
```

#### Authentication Failures

```yaml
# Verify credentials in config.yaml
auth:
  username: "correct-username"
  password: "correct-password"
```

#### SSL Certificate Issues

```bash
# Disable SSL verification (not recommended for production)
export SSL_CERT_FILE=/path/to/cert.pem
```

### Debug Mode

Enable verbose logging for troubleshooting:

```bash
./api-security-scanner -verbose
```

## 🤝 Contributing

We welcome contributions! Please see our [Contributing Guidelines](CONTRIBUTING.md) for details.

### Development Setup

```bash
# Fork the repository
git clone https://github.com/elliotsecops/API-Security-Scanner.git
cd API-Security-Scanner

# Create a feature branch
git checkout -b feature/amazing-feature

# Make your changes
go build
go run main.go

# Test your changes
git add .
git commit -m "Add amazing feature"
git push origin feature/amazing-feature
```

### Coding Standards

- Follow Go standard formatting (`go fmt`)
- Write clear, concise code with proper documentation
- Add comprehensive error handling
- Include tests for new features
- Update documentation as needed

## 📄 License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## 🙏 Acknowledgments

- **OWASP API Security Top 10** - For security testing guidelines
- **Go Community** - For excellent tooling and libraries
- **Security Researchers** - For vulnerability research and disclosures

## 📞 Support

- **Issues**: [GitHub Issues](https://github.com/elliotsecops/API-Security-Scanner/issues)
- **Discussions**: [GitHub Discussions](https://github.com/elliotsecops/API-Security-Scanner/discussions)
- **Email**: [Security Team](mailto:security@elliotsecops.com)

## 🔮 Roadmap

### Phase 1: Core Infrastructure ✅
- [x] Basic authentication testing
- [x] SQL injection detection
- [x] HTTP method validation
- [x] Concurrent execution
- [x] Detailed reporting

### Phase 2: Enhanced Testing ✅
- [x] XSS vulnerability detection
- [x] Security header analysis
- [x] Authentication bypass testing
- [x] Parameter tampering detection

### Phase 3: Advanced Features (Planned)
- [ ] NoSQL injection testing
- [ ] OpenAPI/Swagger integration
- [ ] API discovery and crawling
- [ ] Historical comparison and trending

### Phase 4: Enterprise Features (Planned)
- [ ] Multi-tenant support
- [ ] SIEM integration
- [ ] Advanced authentication methods
- [ ] Performance metrics

---

**Made with ❤️ for the security community**

[![Star on GitHub](https://img.shields.io/github/stars/elliotsecops/API-Security-Scanner.svg?style=social&label=Star)](https://github.com/elliotsecops/API-Security-Scanner)
[![Fork on GitHub](https://img.shields.io/github/forks/elliotsecops/API-Security-Scanner.svg?style=social&label=Fork)](https://github.com/elliotsecops/API-Security-Scanner)
