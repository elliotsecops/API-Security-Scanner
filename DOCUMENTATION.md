# API Security Scanner - Comprehensive Documentation

## 📚 Table of Contents

1. [Overview](#overview)
2. [Installation](#installation)
3. [Configuration](#configuration)
4. [Enterprise Features](#enterprise-features)
5. [SIEM Integration](#siem-integration)
6. [Authentication Methods](#authentication-methods)
7. [Multi-Tenant Architecture](#multi-tenant-architecture)
8. [Performance Monitoring](#performance-monitoring)
9. [API Reference](#api-reference)
10. [Troubleshooting](#troubleshooting)
11. [Contributing](#contributing)

## 🎯 Overview

The API Security Scanner Enterprise Edition is a comprehensive security testing tool designed for organizations that need to test REST APIs for common vulnerabilities. It provides enterprise-grade features including multi-tenant support, SIEM integration, advanced authentication, and performance monitoring.

### Key Features

- **Comprehensive Security Testing**: SQL injection, XSS, NoSQL injection, authentication bypass, parameter tampering
- **Enterprise Architecture**: Multi-tenant support with complete data isolation
- **SIEM Integration**: Native integration with Wazuh, Splunk, ELK, QRadar, and ArcSight
- **Advanced Authentication**: OAuth2, JWT, API keys, Mutual TLS, and Basic auth
- **Performance Monitoring**: Real-time metrics collection and monitoring dashboard
- **Historical Analysis**: Trend analysis and comparative reporting
- **API Discovery**: Automatic endpoint discovery and crawling
- **OpenAPI Integration**: Generate tests from OpenAPI specifications

## 🚀 Installation

### Prerequisites

- Go 1.24 or later (for building from source)
- Docker and Docker Compose (for containerized deployment)
- YAML configuration files
- Network access to target APIs
- (Optional) SIEM platform access
- (Optional) Authentication credentials

### Quick Install with Docker (Recommended for Testing)

The easiest way to get started is using the provided Docker Compose setup that includes both the scanner and a test API (OWASP Juice Shop):

```bash
# Clone the repository
git clone https://github.com/elliotsecops/API-Security-Scanner.git
cd API-Security-Scanner

# Build and start the complete test environment
docker-compose up -d

# Verify both containers are running
docker ps

# Access the dashboard at: http://localhost:8080
# Run a test scan:
docker exec api-security-scanner ./api-security-scanner -config config-test.yaml -scan
```

### Build from Source (For Development)

```bash
# Clone the repository
git clone https://github.com/elliotsecops/API-Security-Scanner.git
cd API-Security-Scanner

# Install dependencies
go mod tidy

# Install enterprise dependencies
go get github.com/golang-jwt/jwt/v5
go get golang.org/x/oauth2
go get github.com/sirupsen/logrus
go get github.com/gorilla/websocket

# Build the application
go build -o api-security-scanner

# Verify installation
./api-security-scanner -version
```

### Docker Manual Build

```bash
# Build Docker image
docker build -t api-security-scanner .

# Run in dashboard mode
docker run -d --name api-security-scanner -p 8080-8081:8080-8081 \
  -v $(pwd)/config-test.yaml:/app/config-test.yaml \
  -v $(pwd)/reports:/app/reports \
  api-security-scanner ./api-security-scanner -config config-test.yaml -dashboard
```

## ⚙️ Configuration

### Basic Configuration

```yaml
# Core scanner configuration
scanner:
  api_endpoints:
    - url: "https://api.example.com/users"
      method: "GET"
    - url: "https://api.example.com/data"
      method: "POST"
      body: '{"query": "value"}'

  # Test payloads
  injection_payloads:
    - "' OR '1'='1"
    - "'; DROP TABLE users;--"
    - "1' OR '1'='1"
    - "admin'--"

  xss_payloads:
    - "<script>alert('XSS')</script>"
    - "'><script>alert('XSS')</script>"
    - "<img src=x onerror=alert('XSS')>"

  nosql_payloads:
    - "{$ne: null}"
    - "{$gt: ''}"
    - "{$or: [1,1]}"
    - "{$where: 'sleep(100)'}"

  # Rate limiting
  rate_limiting:
    requests_per_second: 10
    max_concurrent_requests: 5

  # Custom headers
  headers:
    "User-Agent": "API-Security-Scanner/4.0"
    "X-Scanner": "true"
```

### Enterprise Configuration

```yaml
# Multi-tenant configuration
tenant:
  id: "your-tenant-id"
  name: "Your Organization"
  description: "Security team"
  is_active: true
  settings:
    resource_limits:
      max_requests_per_day: 10000
      max_concurrent_scans: 5
      max_endpoints_per_scan: 100
    data_isolation:
      storage_path: "./data/your-tenant-id"
      enabled: true
    notification_settings:
      email_notifications: true
      webhook_url: "https://hooks.slack.com/services/xxx"
      alert_threshold: "medium"

# SIEM Integration
siem:
  enabled: true
  type: "syslog"  # syslog, splunk, elk, qradar, arcsight
  format: "json"
  endpoint_url: ""  # For HTTP-based SIEMs
  auth_token: ""   # For authenticated SIEMs
  config:
    host: "localhost"
    port: 514
    facility: "local0"
    severity: "info"

# Advanced Authentication
auth:
  enabled: true
  type: "oauth2"  # basic, oauth2, jwt, api_key, mutual_tls
  config:
    client_id: "your-client-id"
    client_secret: "your-client-secret"
    token_url: "https://auth.example.com/oauth/token"
    scopes: ["read", "write"]
    grant_type: "client_credentials"

# Performance Metrics
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

# Server configuration
server:
  port: 8081
  host: "localhost"
```

## 🔧 Enterprise Features

### Multi-Tenant Architecture

The Enterprise Edition supports multi-tenant deployments with complete data isolation:

#### Key Features

- **Tenant Management**: Create and manage multiple security teams/organizations
- **Data Isolation**: Each tenant has isolated storage and configuration
- **Resource Quotas**: Configure limits per tenant (requests, endpoints, concurrent scans)
- **Custom Settings**: Tenant-specific notification settings and alert thresholds

#### Tenant Configuration

```yaml
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

#### Multi-Tenant Usage

```bash
# Run scan for specific tenant
./api-security-scanner -config enterprise-config.yaml -tenant "acme-corp" -scan

# Start dashboard for specific tenant
./api-security-scanner -config enterprise-config.yaml -tenant "acme-corp" -dashboard

# List all tenants
./api-security-scanner -config enterprise-config.yaml -list-tenants
```

### SIEM Integration

The scanner can send security events to major SIEM platforms for centralized monitoring:

#### Supported SIEM Platforms

- **Wazuh**: Native syslog integration with custom decoders and rules
- **Splunk**: HTTP Event Collector (HEC) integration
- **ELK Stack**: Elasticsearch indexing with Kibana dashboards
- **IBM QRadar**: LEA protocol and event forwarding
- **ArcSight**: CEF format and SmartConnector integration

#### Wazuh Integration

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
```

#### Splunk Integration

```yaml
# Splunk SIEM Configuration
siem:
  enabled: true
  type: "splunk"
  format: "json"
  endpoint_url: "https://splunk.company.com:8088/services/collector"
  auth_token: "your-splunk-hec-token"
```

#### ELK Stack Integration

```yaml
# ELK Stack Configuration
siem:
  enabled: true
  type: "elk"
  format: "json"
  endpoint_url: "https://elasticsearch.company.com:9200/api-scanner-events/_doc"
  auth_token: "your-elasticsearch-token"
```

### Advanced Authentication

Support for enterprise authentication standards:

#### OAuth 2.0 Authentication

```yaml
auth:
  enabled: true
  type: "oauth2"
  config:
    client_id: "security-scanner"
    client_secret: "your-client-secret"
    token_url: "https://auth.company.com/oauth/token"
    scopes: ["read", "write"]
    grant_type: "client_credentials"
```

#### JWT Authentication

```yaml
auth:
  enabled: true
  type: "jwt"
  config:
    secret_key: "your-jwt-secret"
    signing_method: "HS256"
    audience: "api-security-scanner"
    issuer: "auth.company.com"
```

#### API Key Authentication

```yaml
auth:
  enabled: true
  type: "api_key"
  config:
    header_name: "X-API-Key"
    key_value: "your-api-key"
```

#### Mutual TLS Authentication

```yaml
auth:
  enabled: true
  type: "mutual_tls"
  config:
    cert_file: "/path/to/client.crt"
    key_file: "/path/to/client.key"
    ca_file: "/path/to/ca.crt"
```

### Performance Monitoring

Real-time monitoring and metrics collection:

#### Metrics Configuration

```yaml
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

#### Available Metrics

- **System Metrics**: CPU usage, memory usage, network I/O
- **Security Metrics**: Vulnerability counts, scan success rates
- **Business Metrics**: API availability, response times
- **Scanner Metrics**: Requests per second, error rates

#### Monitoring Dashboard

```bash
# Start monitoring dashboard
./api-security-scanner -config config-metrics.yaml -dashboard

# Access dashboard at http://localhost:8081
```

## 🔐 SIEM Integration Guide

### Wazuh SIEM Integration

#### Configuration

```yaml
siem:
  enabled: true
  type: "syslog"
  format: "json"
  config:
    host: "wazuh-manager.company.com"
    port: 514
    facility: "local0"
    severity: "info"
```

#### Wazuh Manager Configuration

Add the following to your Wazuh `/var/ossec/etc/ossec.conf`:

```xml
<!-- API Security Scanner Syslog Integration -->
<remote>
  <connection>syslog</connection>
  <port>514</port>
  <protocol>udp</protocol>
  <allowed-ips>127.0.0.1</allowed-ips>
</remote>

<!-- Custom decoder for API Security Scanner events -->
<decoder name="api-security-scanner">
  <program_name>^api-security-scanner</program_name>
</decoder>

<!-- Custom rules for API Security Scanner -->
<group name="api,security,vulnerability">
  <rule id="100100" level="5">
    <if_sid>5711</if_sid>
    <field name="program_name">^api-security-scanner</field>
    <description>API Security Scanner - Security Event</description>
    <group>api_security</group>
  </rule>

  <rule id="100101" level="8">
    <if_sid>100100</if_sid>
    <field name="vulnerability">SQL injection</field>
    <description>API Security Scanner - SQL Injection Detected</description>
    <group>sql_injection,attack</group>
  </rule>

  <rule id="100102" level="8">
    <if_sid>100100</if_sid>
    <field name="vulnerability">XSS</field>
    <description>API Security Scanner - XSS Vulnerability Detected</description>
    <group>xss,attack</group>
  </rule>

  <rule id="100103" level="10">
    <if_sid>100100</if_sid>
    <field name="vulnerability">authentication bypass</field>
    <description>API Security Scanner - Authentication Bypass Detected</description>
    <group>auth_bypass,critical</group>
  </rule>
</group>
```

### Splunk SIEM Integration

#### Configuration

```yaml
siem:
  enabled: true
  type: "splunk"
  format: "json"
  endpoint_url: "https://splunk.company.com:8088/services/collector"
  auth_token: "your-splunk-hec-token"
```

#### Splunk Setup

1. **Enable HTTP Event Collector (HEC)**:
   - Navigate to Settings → Data Inputs → HTTP Event Collector
   - Create new HEC token with appropriate permissions
   - Configure index for API security events

2. **Create Dashboard**:
   - Use Splunk Web to create security dashboards
   - Add visualizations for vulnerability trends
   - Set up alerts for critical findings

### ELK Stack Integration

#### Configuration

```yaml
siem:
  enabled: true
  type: "elk"
  format: "json"
  endpoint_url: "https://elasticsearch.company.com:9200/api-scanner-events/_doc"
  auth_token: "your-elasticsearch-token"
```

#### Kibana Dashboard

1. **Create Index Pattern**:
   - Pattern: `api-scanner-*`
   - Time field: `timestamp`

2. **Import Dashboard**:
   - Use provided Kibana dashboard export
   - Customize for your environment
   - Set up alerts and notifications

## 🏗️ Multi-Tenant Architecture

### Architecture Overview

The multi-tenant architecture provides complete isolation between different organizations or departments:

```
┌─────────────────────────────────────────────────────────────┐
│                    API Security Scanner                     │
│                                                             │
│  ┌─────────────┐  ┌─────────────┐  ┌─────────────┐        │
│  │   Tenant A  │  │   Tenant B  │  │   Tenant C  │        │
│  │             │  │             │  │             │        │
│  │ • Isolated  │  │ • Isolated  │  │ • Isolated  │        │
│  │   Storage   │  │   Storage   │  │   Storage   │        │
│  │ • Resource  │  │ • Resource  │  │ • Resource  │        │
│  │   Quotas    │  │   Quotas    │  │   Quotas    │        │
│  │ • Custom    │  │ • Custom    │  │ • Custom    │        │
│  │   Settings  │  │   Settings  │  │   Settings  │        │
│  └─────────────┘  └─────────────┘  └─────────────┘        │
│                                                             │
│  ┌─────────────────────────────────────────────────────────┐ │
│  │                Core Scanner Engine                     │ │
│  │  • Security Testing • SIEM Integration • Auth          │ │
│  │  • Metrics Collection • Dashboard • Health Checks     │ │
│  └─────────────────────────────────────────────────────────┘ │
└─────────────────────────────────────────────────────────────┘
```

### Tenant Management

#### Creating a Tenant

```yaml
tenant:
  id: "new-tenant"
  name: "New Organization"
  description: "Security operations center"
  is_active: true
  settings:
    resource_limits:
      max_requests_per_day: 10000
      max_concurrent_scans: 5
      max_endpoints_per_scan: 100
    data_isolation:
      storage_path: "./data/new-tenant"
      enabled: true
    notification_settings:
      email_notifications: true
      webhook_url: "https://hooks.slack.com/services/xxx"
      alert_threshold: "medium"
```

#### Resource Quotas

```yaml
settings:
  resource_limits:
    max_requests_per_day: 10000    # Daily request limit
    max_concurrent_scans: 5         # Concurrent scan limit
    max_endpoints_per_scan: 100     # Endpoints per scan limit
    max_storage_mb: 1000            # Storage limit in MB
    scan_retention_days: 30         # Retention period
```

#### Data Isolation

```yaml
settings:
  data_isolation:
    storage_path: "./data/tenant-id"    # Isolated storage path
    enabled: true                       # Enable data isolation
    encryption_enabled: true            # Encrypt data at rest
    backup_enabled: true                # Enable automatic backups
    backup_schedule: "0 2 * * *"        # Backup schedule (cron)
```

## 📊 Performance Monitoring

### Metrics Collection

The scanner collects comprehensive metrics for monitoring and analysis:

#### System Metrics

- **CPU Usage**: Processor utilization percentage
- **Memory Usage**: RAM consumption and allocation
- **Network I/O**: Network traffic and bandwidth usage
- **Disk Usage**: Storage utilization and I/O operations
- **Goroutines**: Active goroutines count
- **Uptime**: Application uptime in seconds

#### Security Metrics

- **Vulnerability Counts**: Total vulnerabilities detected
- **Scan Success Rate**: Percentage of successful scans
- **Threat Trends**: Vulnerability trends over time
- **Risk Distribution**: Risk level distribution
- **Compliance Status**: Compliance with security standards

#### Business Metrics

- **API Availability**: API uptime and availability
- **Response Times**: Average and peak response times
- **Error Rates**: HTTP error rate analysis
- **Throughput**: Requests per second metrics

### Monitoring Dashboard

The real-time dashboard provides comprehensive monitoring capabilities:

#### Features

- **Real-time Metrics**: Live metrics updates via WebSocket
- **Historical Charts**: Trend analysis and historical data
- **Alert Management**: Configure and manage alerts
- **Tenant Overview**: Multi-tenant metrics overview
- **Health Status**: System health monitoring

#### Accessing the Dashboard

```bash
# Start the dashboard
./api-security-scanner -config config.yaml -dashboard

# Access at http://localhost:8081
```

#### Dashboard Sections

1. **Overview**: System-wide metrics and health status
2. **Security Metrics**: Vulnerability statistics and trends
3. **Performance Metrics**: System performance indicators
4. **Tenant Metrics**: Multi-tenant resource usage
5. **Alerts**: Active alerts and notifications
6. **Configuration**: Dashboard settings and preferences

### Health Checks

Automated health monitoring ensures system reliability:

```yaml
metrics:
  health_check:
    enabled: true
    interval: 30s        # Check interval
    timeout: 10s         # Check timeout
    endpoints:           # Endpoints to check
      - "https://api.example.com/health"
      - "https://auth.example.com/health"
    notification_channels:  # Alert channels
      - "email:admin@company.com"
      - "webhook:https://hooks.slack.com/services/xxx"
```

## 📖 API Reference

### Configuration API

#### GET /api/config

Get current configuration.

```bash
curl -X GET http://localhost:8081/api/config
```

#### POST /api/config

Update configuration.

```bash
curl -X POST http://localhost:8081/api/config \
  -H "Content-Type: application/json" \
  -d '{"scanner": {"api_endpoints": [...]}}'
```

### Tenant Management API

#### GET /api/tenants

List all tenants.

```bash
curl -X GET http://localhost:8081/api/tenants
```

#### POST /api/tenants

Create new tenant.

```bash
curl -X POST http://localhost:8081/api/tenants \
  -H "Content-Type: application/json" \
  -d '{"id": "new-tenant", "name": "New Organization"}'
```

#### GET /api/tenants/{id}

Get tenant details.

```bash
curl -X GET http://localhost:8081/api/tenants/tenant-id
```

### Metrics API

#### GET /api/metrics

Get current metrics.

```bash
curl -X GET http://localhost:8081/api/metrics
```

#### GET /api/metrics/history

Get historical metrics.

```bash
curl -X GET "http://localhost:8081/api/metrics/history?period=24h"
```

### Scan API

#### POST /api/scan

Start security scan.

```bash
curl -X POST http://localhost:8081/api/scan \
  -H "Content-Type: application/json" \
  -d '{"tenant_id": "tenant-id", "endpoints": [...]}'
```

#### GET /api/scan/{id}

Get scan results.

```bash
curl -X GET http://localhost:8081/api/scan/scan-id
```

### SIEM API

#### POST /api/siem/test

Test SIEM connection.

```bash
curl -X POST http://localhost:8081/api/siem/test \
  -H "Content-Type: application/json" \
  -d '{"type": "syslog", "config": {...}}'
```

## 🛠️ Troubleshooting

### Common Issues

#### Build Errors

**Problem**: Build fails with missing dependencies

```bash
# Solution: Install missing dependencies
go mod tidy
go get github.com/golang-jwt/jwt/v5
go get golang.org/x/oauth2
```

#### Configuration Errors

**Problem**: Invalid YAML configuration

```bash
# Solution: Validate YAML syntax
go run main.go -validate-config
```

#### SIEM Connection Issues

**Problem**: Cannot connect to SIEM platform

```bash
# Solution: Test SIEM connection
./api-security-scanner -config config.yaml -test-siem

# Check network connectivity
telnet siem-host 514
```

#### Authentication Issues

**Problem**: Authentication failures

```bash
# Solution: Test authentication
./api-security-scanner -config config.yaml -test-auth

# Check credentials and permissions
```

#### Performance Issues

**Problem**: High memory usage

```bash
# Solution: Adjust rate limiting
rate_limiting:
  requests_per_second: 5
  max_concurrent_requests: 3
```

### Debug Mode

Enable debug logging for troubleshooting:

```bash
# Run with debug logging
./api-security-scanner -config config.yaml -debug -scan

# Enable specific debug components
./api-security-scanner -config config.yaml -debug=auth,siem,scanner -scan
```

### Log Files

Check log files for detailed error information:

```bash
# View application logs
tail -f logs/app.log

# View SIEM integration logs
tail -f logs/siem.log

# View authentication logs
tail -f logs/auth.log
```

## 🤝 Contributing

We welcome contributions! Please follow these guidelines:

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
go test ./...

# Commit your changes
git add .
git commit -m "Add amazing feature"

# Push to branch
git push origin feature/amazing-feature

# Create a Pull Request
```

### Code Standards

- Follow Go standard formatting (`go fmt`)
- Write clear, concise code with proper documentation
- Add comprehensive error handling
- Include tests for new features
- Update documentation as needed

### Testing

```bash
# Run all tests
go test ./...

# Run tests with coverage
go test -cover ./...

# Run specific test
go test -run TestFunctionName ./...

# Run benchmarks
go test -bench=. ./...
```

### Documentation

- Update README.md for user-facing changes
- Update this documentation for internal changes
- Add comments to complex code sections
- Include examples for new features

---

## 📞 Support

For support and questions:

- **GitHub Issues**: [Report bugs and request features](https://github.com/elliotsecops/API-Security-Scanner/issues)
- **GitHub Discussions**: [Community discussions](https://github.com/elliotsecops/API-Security-Scanner/discussions)
- **Email**: [Security Team](mailto:security@elliotsecops.com)
- **Documentation**: [Full documentation](https://github.com/elliotsecops/API-Security-Scanner/wiki)

## 📄 License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

---

**Made with ❤️ for the security community**