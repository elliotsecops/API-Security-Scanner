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

Built with performance and reliability in mind, the scanner uses concurrent execution to efficiently test multiple endpoints simultaneously while providing detailed security reports.

## ✨ Key Features

### 🔒 Security Testing
- **Authentication Testing**: Validates basic auth credentials and identifies access control issues
- **SQL Injection Detection**: Comprehensive payload-based testing for SQL injection vulnerabilities
- **HTTP Method Validation**: Ensures proper HTTP method handling and prevents method-based attacks
- **Concurrent Execution**: High-performance parallel testing of multiple endpoints
- **Detailed Reporting**: Comprehensive security assessments with risk analysis

### 🚀 Performance & Reliability
- **Fast Execution**: Concurrent testing with configurable rate limiting
- **Robust Error Handling**: Graceful handling of network timeouts and connection issues
- **Memory Efficient**: Optimized for large-scale API testing
- **Configurable Timeouts**: Prevents hanging requests with configurable timeouts

### 📊 Reporting & Output
- **Multiple Output Formats**: Text-based detailed reports
- **Risk Assessment**: Automated risk scoring and remediation recommendations
- **Comprehensive Logging**: Structured logging for debugging and audit purposes
- **Score-based Metrics**: 100-point scoring system for security posture assessment

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
| `-output` | Output format (text, json) | `text` |
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
  - "<script>alert('XSS')</script>" # XSS payload (if testing web APIs)
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

Risk Assessment:
No significant risks detected.

Endpoint: https://api.example.com/v1/data
Overall Score: 50/100
Test Results:
- Auth Test: PASSED
  Details: Authentication successful
- HTTP Method Test: PASSED
  Details: Method validation successful
- Injection Test: FAILED
  Details: Potential SQL injection detected with payload: ' OR '1'='1

Risk Assessment:
- SQL injection vulnerabilities pose a significant data breach risk.

Overall Security Assessment:
Average Security Score: 75/100
Critical Vulnerabilities Detected: 1

Moderate security risks detected. Address identified vulnerabilities promptly.
```

## 🎯 Security Scoring

The scanner uses a 100-point scoring system:

- **Starting Score**: 100/100 for each endpoint
- **Authentication Failure**: -30 points
- **HTTP Method Failure**: -20 points
- **Injection Vulnerability**: -50 points

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
git clone https://github.com/your-username/API-Security-Scanner.git
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

### Phase 2: Enhanced Testing (Planned)
- [ ] XSS vulnerability detection
- [ ] NoSQL injection testing
- [ ] Security header analysis
- [ ] Rate limiting and throttling

### Phase 3: Advanced Features (Planned)
- [ ] OpenAPI/Swagger integration
- [ ] API discovery and crawling
- [ ] Multiple output formats (JSON, XML, HTML)
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
