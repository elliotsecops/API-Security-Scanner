# 🚀 API Security Scanner - Enterprise Edition

[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)
[![Go Version](https://img.shields.io/badge/Go-1.24+-blue.svg)](https://golang.org/dl/)
[![Build Status](https://img.shields.io/badge/Build-Passing-green.svg)](https://github.com/elliotsecops/API-Security-Scanner)
[![Version](https://img.shields.io/badge/Version-4.0.0-orange.svg)](https://github.com/elliotsecops/API-Security-Scanner)
[![Enterprise](https://img.shields.io/badge/Enterprise-Ready-brightgreen.svg)](https://github.com/elliotsecops/API-Security-Scanner)

**⚡ One-Command Setup • Zero-Configuration GUI • Enterprise-Grade Security**

A comprehensive, enterprise-grade API security testing platform with an intuitive web interface. Get started in seconds with automated installation and smart dependency management.

## 🎯 Quick Start

Easiest way (one command):

```bash
# From the repo root
./start.sh
```

- This will automatically build the GUI if needed and pick free ports if defaults are busy.
- Default dashboard URL: http://localhost:8090 (the exact port will be printed on start).
- Run without the GUI: `./start.sh --no-gui`
- Override ports: `./start.sh --port 8082 --metrics-port 8095`

Alternative shortcuts:

```bash
# Using Make
make start            # with GUI
a make start-nogui     # backend only

# Using npm (from repo root)
npm start             # with GUI
npm run start:nogui   # backend only
```

Previous flow (kept for development):

```bash
# Clone and install everything automatically
git clone https://github.com/elliotsecops/API-Security-Scanner.git
cd API-Security-Scanner
./install.sh

# Or run development mode (GUI dev server)
./run.sh dev
```

Access the GUI:
- Development (run.sh dev): http://localhost:3000
- Production via start.sh: http://localhost:8090 (or the printed port)

Default Login: `admin` / `admin`

## 🌟 Why Choose This Scanner?

- **🚀 One-Command Installation** - Automated setup with dependency management
- **🎯 User-Friendly GUI** - Modern React interface with real-time dashboards
- **🛡️ Comprehensive Testing** - SQLi, XSS, NoSQL injection, auth bypass, and more
- **📊 Smart Reporting** - Risk assessment, trending, and multi-format exports
- **🏢 Enterprise Ready** - Multi-tenant, SIEM integration, API discovery
- **⚡ High Performance** - Concurrent testing with configurable rate limits

## 🎯 Overview

The API Security Scanner performs automated security assessments of API endpoints by testing for common vulnerabilities including:

- **Authentication Bypass Detection**
- **SQL Injection Vulnerabilities**
- **NoSQL Injection Vulnerabilities**
- **HTTP Method Validation**
- **Security Header Analysis**
- **Parameter Tampering Detection**
- **Cross-Site Scripting (XSS) Vulnerabilities**
- **Authentication Bypass Testing**

Built with enterprise-grade performance and reliability in mind, the scanner uses concurrent execution to efficiently test multiple endpoints simultaneously while providing detailed security reports. The platform now includes comprehensive enterprise features such as multi-tenant architecture, SIEM integration, real-time monitoring dashboards, and advanced authentication methods.

## ✨ Key Features

### 🔒 Security Testing
- **Authentication Testing**: Validates basic auth credentials and identifies access control issues
- **SQL Injection Detection**: Comprehensive payload-based testing for SQL injection vulnerabilities
- **HTTP Method Validation**: Ensures proper HTTP method handling and prevents method-based attacks
- **XSS Vulnerability Detection**: Tests for cross-site scripting vulnerabilities using common payloads
- **Header Security Analysis**: Analyzes HTTP response headers for security issues
- **Authentication Bypass Testing**: Tests for authentication vulnerabilities
- **Parameter Tampering Detection**: Tests for parameter manipulation vulnerabilities
- **NoSQL Injection Detection**: Tests for NoSQL injection vulnerabilities in MongoDB, CouchDB, etc. with comprehensive payload sets.
- **Concurrent Execution**: High-performance parallel testing of multiple endpoints
- **Detailed Reporting**: Comprehensive security assessments with risk analysis

### 🔧 Advanced Features
- **OpenAPI/Swagger Integration**: Import and test APIs from OpenAPI specifications with automatic endpoint generation and validation
- **API Discovery**: Automatically discover and crawl API endpoints from base URLs with configurable depth and exclusion patterns
- **Historical Comparison**: Track security posture over time with comprehensive trend analysis and vulnerability tracking
- **Endpoint Crawling**: Recursive discovery of API endpoints with intelligent link extraction and parameter discovery
- **Parameter Discovery**: Automatically discover and test API parameters from HTML forms and API responses
- **Score Trending**: Visualize security score changes over multiple scans with detailed reporting
- **Vulnerability Tracking**: Track new and resolved vulnerabilities across scans with comparative analysis
- **Multi-format Reports**: Generate reports in JSON, HTML, CSV, XML, and text formats with historical data

### 🏢 Enterprise Features (Phase 4)
- **Multi-Tenant Architecture**: Support for multiple organizations with complete data isolation and resource quotas
- **SIEM Integration**: Native integration with Wazuh, Splunk, ELK, QRadar, and ArcSight for centralized security monitoring
- **Advanced Authentication**: Support for OAuth2, JWT, API keys, bearer tokens, and mutual TLS authentication
- **Real-Time Monitoring Dashboard**: Web-based dashboard with WebSocket support for live metrics and visualization
- **Grafana Integration**: Built-in Prometheus metrics export for advanced visualization and alerting with Grafana
- **Performance Metrics**: Comprehensive monitoring of CPU, memory, network usage, and scan performance
- **Resource Management**: Configurable rate limits, concurrency controls, and resource quotas per tenant
- **Alert Management**: Configurable alerts with email, Slack, and webhook notifications for critical findings
- **Health Monitoring**: Built-in health checks and system monitoring with automatic failover
- **Historical Analytics**: Advanced trend analysis with time-series data and predictive insights
- **Enterprise Logging**: Structured JSON logging with tenant isolation and audit trails

### 🖥️ Web Interface (GUI)
- **React-Based GUI**: Modern, responsive web interface built with React and Material-UI
- **Real-Time Dashboard**: Live metrics, vulnerability tracking, and system health monitoring
- **Interactive Visualizations**: Charts and graphs for security data analysis using Chart.js
- **Grafana Integration**: Prometheus metrics endpoint for advanced visualization and alerting
- **Scan Management**: Configure, run, and monitor security scans through intuitive web interface
- **Results Analysis**: Detailed vulnerability reports with filtering, searching, and export capabilities
- **Multi-Tenant Management**: Complete tenant administration through web interface
- **Development & Production Modes**: Hot reload development and optimized production builds
- **Single Server Deployment**: GUI served by Go backend, no separate web server required

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

### ⚡ Automated Installation (Recommended)

**Zero-config setup - detects and installs everything automatically:**

```bash
# Full automated installation
git clone https://github.com/elliotsecops/API-Security-Scanner.git
cd API-Security-Scanner
./install.sh
```

The installation script will:
- ✅ Detect your operating system
- ✅ Install Go 1.24+ (if missing)
- ✅ Install Node.js v16+ (if missing)
- ✅ Install all GUI dependencies
- ✅ Build the application
- ✅ Create configuration files
- ✅ Set up desktop shortcuts (Linux)

### 🚀 Quick Start

**Already have dependencies installed? Start immediately:**

```bash
git clone https://github.com/elliotsecops/API-Security-Scanner.git
cd API-Security-Scanner
./run.sh dev
```

### 📋 Prerequisites (Manual Install)

If you prefer manual installation:

- **Go 1.24 or higher** - [Download Go](https://golang.org/dl/)
- **Node.js v16 or higher** - [Download Node.js](https://nodejs.org/)
- **Git** - [Download Git](https://git-scm.com/downloads)

### 🐳 Docker Installation and Integration Testing

The project includes a complete integration test environment with OWASP Juice Shop, allowing you to test the scanner in a controlled environment with known vulnerabilities.

### Quick Start with Complete Environment

```bash
# Clone the repository
git clone https://github.com/elliotsecops/API-Security-Scanner.git
cd API-Security-Scanner

# Build and start both the scanner and the test API (OWASP Juice Shop)
docker-compose up -d

# Verify both containers are running
docker ps

# The dashboard is accessible at: http://localhost:8090 (or the port printed by the launcher)
# The API is available on port: 8081 (for API calls)
# The test API (Juice Shop) is available on port: 3000
```

### Run a Test Scan

Once both containers are running, you can execute a security scan:

```bash
# Run a test scan against OWASP Juice Shop
docker exec api-security-scanner ./api-security-scanner -config config-test.yaml -scan

# Or run the scanner in dashboard mode
docker exec api-security-scanner ./api-security-scanner -config config-test.yaml -dashboard
```

### One-Command Launch with Docker

For the simplest setup with automatic port detection and GUI:

```bash
# From the repo root (when using Docker setup)
./start.sh --port 8081 --metrics-port 8090
```

Alternative shortcuts:
```bash
make start            # same as ./start.sh
npm start             # same as ./start.sh
```

### Manual Docker Build

```bash
# Build Docker image
docker build -t api-security-scanner .

# Run with test configuration
docker run -d --name api-security-scanner -p 8080-8081:8080-8081 \
  -v $(pwd)/config-test.yaml:/app/config-test.yaml \
  -v $(pwd)/reports:/app/reports \
  api-security-scanner ./api-security-scanner -config config-test.yaml -dashboard
```

## 🎮 Running the Application

### ⚡ Easy Commands

One-command launch (recommended):

```bash
./start.sh                       # build GUI if needed and start dashboard
./start.sh --no-gui              # backend only
./start.sh --port 8082           # override backend port
./start.sh --metrics-port 8095   # override dashboard port
```

Shortcuts:

```bash
make start           # same as ./start.sh
make start-nogui     # same as ./start.sh --no-gui
npm start            # same as ./start.sh
npm run start:nogui  # same as ./start.sh --no-gui
```

Legacy/dev commands (still available):

```bash
./run.sh dev         # GUI dev server at :3000, backend at :8080
./run.sh prod        # legacy production runner
./run.sh backend     # backend only
./run.sh gui         # GUI dev server
./run.sh build       # build everything
./run.sh stop        # stop processes
./run.sh help        # show help
```

## 🚀 Usage

### 🌐 Web Interface (GUI)

**Access Points:**
- Development (run.sh dev): http://localhost:3000
- Production (start.sh): http://localhost:8090 (or the port printed by the launcher)
- Default Login: `admin` / `admin`

**GUI Features:**
- **📊 Dashboard**: Real-time metrics and system health
- **🔍 Scanner**: Configure and run security scans
- **📈 Results**: View vulnerability reports and analysis
- **🏢 Tenants**: Multi-tenant management interface
- **⚙️ Settings**: System configuration and preferences

### 🖥️ Command Line Usage

```bash
# Show version information
./api-security-scanner -version

# Run security scan with default configuration
./api-security-scanner -scan

# Run with custom configuration file
./api-security-scanner -scan -config /path/to/custom-config.yaml

# Start monitoring dashboard
./api-security-scanner -dashboard

# Run scan for specific tenant
./api-security-scanner -scan -tenant "production"

# Generate historical comparison
./api-security-scanner -historical -output html

# Show trend analysis
./api-security-scanner -trend -output json
```

### 🔄 Development vs Production

| Mode | Development | Production |
|------|-------------|------------|
| **GUI URL** | http://localhost:3000 | http://localhost:8080 |
| **Process** | 2 separate processes | 1 integrated process |
| **Hot Reload** | ✅ Yes | ❌ No |
| **Performance** | Development optimized | Production optimized |
| **Best For** | Development & Testing | Regular Use |

## 🚨 First Steps

1. **🚀 Start the Application**
   ```bash
   ./run.sh dev  # Development mode
   # or
   ./run.sh prod # Production mode
   ```

2. **🌐 Open the GUI**
   - Development: http://localhost:3000
   - Production: http://localhost:8080

3. **🔑 Login**
   - Username: `admin`
   - Password: `admin`

4. **⚙️ Configure APIs**
   - Edit `config.yaml` to add your API endpoints
   - Or use the GUI to configure endpoints

5. **🔍 Run Security Scan**
   - Use the GUI to configure and run scans
   - View results and vulnerability reports

## 📚 Quick Links

- **🚀 Quick Start**: [QUICKSTART.md](QUICKSTART.md) - Get running in 60 seconds
- **📖 User Guide**: [GUIDE.md](GUIDE.md) - Comprehensive usage guide
- **⚙️ Configuration**: [CONFIGURATION.md](CONFIGURATION.md) - All configuration options
- **🔧 Troubleshooting**: [QUICKSTART.md](QUICKSTART.md#troubleshooting) - Common issues and solutions

### Command Line Options

| Option | Description | Default |
|--------|-------------|---------|
| `-config` | Path to configuration file | `config.yaml` |
| `-scan` | Run security scan immediately | `false` |
| `-dashboard` | Start monitoring dashboard | `false` |
| `-tenant` | Tenant ID for multi-tenant mode | `default` |
| `-output` | Output format (json, html, text) | `json` |
| `-historical` | Show historical comparison | `false` |
| `-trend` | Show trend analysis | `false` |
| `-version` | Show version information | `false` |

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

### NoSQL Injection Configuration

```yaml
nosql_payloads:
  - "{$ne: null}"
  - "{$gt: ''}"
  - "{$or: [1,1]}"
  - "{$where: 'sleep(100)'}"
  - "{$regex: '.*'}"
  - "{$exists: true}"
  - "{$in: [1,2,3]}"
```

### OpenAPI Integration Configuration

```yaml
openapi_spec: "path/to/openapi.yaml"
```

### API Discovery Configuration

```yaml
api_discovery:
  enabled: true
  max_depth: 3
  follow_links: true
  discover_params: true
  user_agent: "API-Security-Scanner-Discovery/1.0"
  exclude_patterns:
    - "/static/"
    - "/assets/"
    - ".css"
    - ".js"
```

### Historical Data Configuration

```yaml
historical_data:
  enabled: true
  storage_path: "./history"
  retention_days: 30
  compare_previous: true
  trend_analysis: true
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
- NoSQL Injection Test: PASSED
  Details: No NoSQL injection vulnerabilities detected

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
- NoSQL Injection Test: FAILED
  Details: Potential NoSQL injection detected with payload: {$ne: null}

Risk Assessment:
- SQL injection vulnerabilities pose a significant data breach risk.
- NoSQL injection vulnerabilities could allow unauthorized database access in document databases.
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
- **SQL Injection Vulnerability**: -50 points
- **NoSQL Injection Vulnerability**: -50 points
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

## 🔧 Enterprise Features

### Multi-Tenant Architecture

The Enterprise Edition supports multi-tenant deployments with complete data isolation:

- **Tenant Management**: Create and manage multiple security teams/organizations
- **Data Isolation**: Each tenant has isolated storage and configuration
- **Resource Quotas**: Configure limits per tenant (requests, endpoints, concurrent scans)
- **Custom Settings**: Tenant-specific notification settings and alert thresholds

### SIEM Integration

Send security events to major SIEM platforms for centralized monitoring:

- **Wazuh**: Native syslog integration with custom decoders and rules
- **Splunk**: HTTP Event Collector (HEC) integration
- **ELK Stack**: Elasticsearch indexing with Kibana dashboards
- **IBM QRadar**: LEA protocol and event forwarding
- **ArcSight**: CEF format and SmartConnector integration

### Advanced Authentication

Support for enterprise authentication standards:

- **OAuth 2.0**: Multiple grant types (client_credentials, authorization_code, password)
- **JWT**: JSON Web Token authentication with various signing algorithms
- **API Keys**: Custom header-based authentication
- **Mutual TLS**: Certificate-based authentication
- **Basic Auth**: Enhanced basic authentication with rate limiting

### Performance Monitoring

Real-time monitoring and metrics collection:

- **System Metrics**: CPU, memory, network usage monitoring
- **Security Metrics**: Vulnerability counts, scan success rates, threat trends
- **Business Metrics**: API availability, response times, compliance status
- **Dashboard**: Real-time WebSocket-based monitoring dashboard
- **Health Checks**: Automated system health monitoring

### Historical Analysis

Comprehensive historical data analysis and trending:

- **Trend Analysis**: Track vulnerability trends over time
- **Comparative Analysis**: Compare current scan results with historical data
- **Compliance Reporting**: Generate reports for regulatory compliance
- **Automated Reporting**: Scheduled reports and executive summaries

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

### NoSQL Injection Testing

The scanner tests for NoSQL injection by:

1. Sending baseline requests to establish normal response patterns
2. Testing with various NoSQL injection payloads for MongoDB, CouchDB, and other document databases
3. Analyzing response differences and error messages for NoSQL syntax
4. Looking for indicators of successful NoSQL injection including:
   - Response time anomalies
   - Error messages containing NoSQL syntax
   - Response body changes indicating unauthorized data access
   - Status code deviations from baseline responses

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

## 📊 Historical Reporting and Analysis

The scanner now includes comprehensive historical reporting capabilities to track security posture over time:

### Historical Comparison Reports

Generate comparison reports between current and previous scans:

```bash
# Generate comparison report (automatically included when historical data is enabled)
./api-security-scanner -output json
./api-security-scanner -output html
./api-security-scanner -output text
```

**Features:**
- Score changes between scans
- New and resolved vulnerability tracking
- Endpoint-specific change analysis
- Visual indicators for improvement/regression

### Trend Analysis

Track security trends over multiple scans:

```yaml
# Enable trend analysis in configuration
historical_data:
  enabled: true
  trend_analysis: true
  storage_path: "./history"
  retention_days: 30
```

**Features:**
- Security score progression over time
- Vulnerability count trends
- Time-based analysis with configurable periods
- Visual charts and graphs (HTML output)

### Data Management

- **Automatic Storage**: Scan results automatically saved with timestamps
- **Configurable Retention**: Set retention policies for historical data
- **Comparison Flexibility**: Compare with any previous scan
- **Export Capabilities**: Export historical data in multiple formats

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

## 🚀 Advanced Usage Examples

### OpenAPI Integration

```bash
# Scan using OpenAPI specification
./api-security-scanner -config config-with-openapi.yaml
```

```yaml
# Configuration with OpenAPI
openapi_spec: "./api-spec.yaml"
api_endpoints: []  # Leave empty to generate from spec
```

### API Discovery

```bash
# Run with API discovery enabled
./api-security-scanner -config config-with-discovery.yaml
```

```yaml
# Configuration with discovery
api_discovery:
  enabled: true
  max_depth: 3
  follow_links: true
  discover_params: true
```

### Historical Analysis

```bash
# Generate trend analysis
./api-security-scanner -output html -config config-with-history.yaml
```

```yaml
# Configuration with historical data
historical_data:
  enabled: true
  storage_path: "./security-history"
  retention_days: 90
  compare_previous: true
  trend_analysis: true
```

### Enterprise SIEM Integration

```bash
# Run with SIEM integration
./api-security-scanner -config config-siem.yaml -scan
```

```yaml
# Wazuh SIEM Configuration
siem:
  enabled: true
  type: "syslog"
  format: "json"
  config:
    host: "wazuh-manager.company.com"
    port: 514
    facility: "local0"
    severity: "info"

# Splunk SIEM Configuration
siem:
  enabled: true
  type: "splunk"
  format: "json"
  endpoint_url: "https://splunk.company.com:8088/services/collector"
  auth_token: "your-splunk-hec-token"
```

### Multi-Tenant Deployment

```bash
# Run scan for specific tenant
./api-security-scanner -config enterprise-config.yaml -tenant "acme-corp" -scan

# Start dashboard for specific tenant
./api-security-scanner -config enterprise-config.yaml -tenant "acme-corp" -dashboard
```

```yaml
# Enterprise multi-tenant configuration
tenant:
  id: "acme-corp"
  name: "Acme Corporation"
  description: "Enterprise security team"
  is_active: true
  settings:
    resource_limits:
      max_requests_per_day: 50000
      max_concurrent_scans: 10
      max_endpoints_per_scan: 200
    data_isolation:
      storage_path: "./data/acme-corp"
      enabled: true
    notification_settings:
      email_notifications: true
      webhook_url: "https://hooks.slack.com/services/xxx"
      alert_threshold: "medium"
```

### Advanced Authentication

```bash
# Run with OAuth2 authentication
./api-security-scanner -config config-oauth2.yaml -scan
```

```yaml
# OAuth2 Configuration
auth:
  enabled: true
  type: "oauth2"
  config:
    client_id: "security-scanner"
    client_secret: "your-client-secret"
    token_url: "https://auth.company.com/oauth/token"
    scopes: ["read", "write"]
    grant_type: "client_credentials"

# JWT Configuration
auth:
  enabled: true
  type: "jwt"
  config:
    secret_key: "your-jwt-secret"
    signing_method: "HS256"
    audience: "api-security-scanner"
    issuer: "auth.company.com"
```

### Performance Monitoring

```bash
# Start monitoring dashboard
./api-security-scanner -config config-metrics.yaml -dashboard

# Run scan with metrics enabled
./api-security-scanner -config config-metrics.yaml -scan
```

```yaml
# Metrics Configuration
metrics:
  enabled: true
  port: 8080
  update_interval: 30s
  retention_days: 30
  dashboard:
    enabled: true
    port: 8081
    host: "localhost"
    update_interval: 5s
    max_connections: 100
  health_check:
    enabled: true
    interval: 30s
    timeout: 10s
```

### Grafana Integration

The API Security Scanner now includes built-in Prometheus metrics export for integration with Grafana for advanced visualization and monitoring:

```yaml
# Metrics Configuration (for Prometheus/Grafana integration)
metrics:
  enabled: true
  port: 8090  # Prometheus will scrape metrics from this port
  update_interval: 30s
  retention_days: 30
```

The scanner exposes metrics in standard Prometheus format at the `/metrics` endpoint. To integrate with Grafana:

1. **Direct Integration**: Configure Prometheus to scrape from `http://<scanner-host>:8090/metrics`
2. **Docker Setup**: Use the provided `grafana-docker-compose.yml` for a complete setup with auto-provisioned dashboard
3. **Custom Dashboards**: All scanner metrics are available with proper labels for tenant isolation

Key metrics available include:
- `api_scanner_total_vulnerabilities` - Total vulnerabilities found
- `api_scanner_critical_vulnerabilities` - Critical vulnerabilities count
- `api_scanner_cpu_usage` - System CPU usage percentage
- `api_scanner_memory_usage` - Memory usage in MB
- `api_scanner_throughput` - Requests per second
- `api_scanner_error_rate` - Error rate percentage
- And many more metrics with tenant-specific labels

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

### Phase 3: Advanced Features ✅
- [x] NoSQL injection testing
- [x] OpenAPI/Swagger integration
- [x] API discovery and crawling
- [x] Historical comparison and trending

### Phase 4: Enterprise Features ✅
- [x] Multi-tenant support with data isolation
- [x] SIEM integration (Wazuh, Splunk, ELK, QRadar, ArcSight)
- [x] Advanced authentication methods (OAuth2, JWT, API keys, Mutual TLS)
- [x] Performance metrics and monitoring dashboard

### Phase 5: Future Enhancements (Planned)
- [ ] Machine learning-based vulnerability detection
- [ ] API behavior analysis and anomaly detection
- [ ] Cloud-native deployment options
- [ ] Advanced compliance reporting
- [ ] Integration with additional security tools

---

**Made with ❤️ for the security community**

[![Star on GitHub](https://img.shields.io/github/stars/elliotsecops/API-Security-Scanner.svg?style=social&label=Star)](https://github.com/elliotsecops/API-Security-Scanner)
[![Fork on GitHub](https://img.shields.io/github/forks/elliotsecops/API-Security-Scanner.svg?style=social&label=Fork)](https://github.com/elliotsecops/API-Security-Scanner)
