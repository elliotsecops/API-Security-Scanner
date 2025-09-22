# API Security Scanner - Configuration Guide

## 📋 Table of Contents

1. [Configuration Overview](#configuration-overview)
2. [Basic Configuration](#basic-configuration)
3. [Enterprise Configuration](#enterprise-configuration)
4. [Scanner Configuration](#scanner-configuration)
5. [Multi-Tenant Configuration](#multi-tenant-configuration)
6. [SIEM Configuration](#siem-configuration)
7. [Authentication Configuration](#authentication-configuration)
8. [Metrics Configuration](#metrics-configuration)
9. [Server Configuration](#server-configuration)
10. [Environment Variables](#environment-variables)
11. [Configuration Validation](#configuration-validation)
12. [Examples](#examples)

## 🎯 Configuration Overview

The API Security Scanner uses YAML configuration files to define all aspects of security testing, enterprise features, and system behavior. The configuration is structured into logical sections for easy management.

### Configuration File Structure

```yaml
scanner:
  # Core scanner settings
  # ...

tenant:
  # Multi-tenant configuration
  # ...

siem:
  # SIEM integration settings
  # ...

auth:
  # Authentication settings
  # ...

metrics:
  # Performance metrics settings
  # ...

server:
  # Server settings
  # ...
```

### Configuration Files

- `config.yaml` - Main configuration file
- `config-enterprise.yaml` - Enterprise configuration example
- `config-wazuh.yaml` - Wazuh SIEM integration example
- `config-splunk.yaml` - Splunk SIEM integration example

## ⚙️ Basic Configuration

### Minimal Configuration

```yaml
# Basic configuration for simple scans
scanner:
  api_endpoints:
    - url: "https://api.example.com/users"
      method: "GET"
    - url: "https://api.example.com/data"
      method: "POST"
      body: '{"query": "value"}'

  injection_payloads:
    - "' OR '1'='1"
    - "'; DROP TABLE users;--"

  xss_payloads:
    - "<script>alert('XSS')</script>"
    - "'><script>alert('XSS')</script>"

  rate_limiting:
    requests_per_second: 10
    max_concurrent_requests: 5
```

### Complete Basic Configuration

```yaml
scanner:
  # API endpoints to test
  api_endpoints:
    - url: "https://api.example.com/v1/users"
      method: "GET"
      headers:
        "Accept": "application/json"
    - url: "https://api.example.com/v1/data"
      method: "POST"
      body: '{"query": "value", "limit": 10}'
      headers:
        "Content-Type": "application/json"

  # Test payloads for SQL injection
  injection_payloads:
    - "' OR '1'='1"
    - "'; DROP TABLE users;--"
    - "1' OR '1'='1"
    - "admin'--"
    - "' OR 1=1--"
    - "' UNION SELECT NULL--"
    - "1; DROP TABLE users--"
    - "'||(SELECT NULL FROM DUAL)||'"

  # Test payloads for XSS
  xss_payloads:
    - "<script>alert('XSS')</script>"
    - "'><script>alert('XSS')</script>"
    - "<img src=x onerror=alert('XSS')>"
    - "javascript:alert('XSS')"
    - "'\"<script>alert(String.fromCharCode(88,83,83))</script>"
    - "<svg onload=alert('XSS')>"
    - "<iframe src=javascript:alert('XSS')>"

  # Test payloads for NoSQL injection
  nosql_payloads:
    - "{$ne: null}"
    - "{$gt: ''}"
    - "{$or: [1,1]}"
    - "{$where: 'sleep(100)'}"
    - "{$regex: '.*'}"
    - "{$exists: true}"
    - "{$in: [1,2,3]}"
    - "{$lt: 100}"

  # Authentication configuration
  auth:
    username: "admin"
    password: "securepassword"
    type: "basic"

  # Rate limiting configuration
  rate_limiting:
    requests_per_second: 10
    max_concurrent_requests: 5
    burst_limit: 15
    cooldown_period: 5s

  # Custom headers
  headers:
    "User-Agent": "API-Security-Scanner/4.0"
    "X-Scanner": "true"
    "Accept": "application/json"
    "X-Request-ID": "{{uuid}}"

  # Timeout configuration
  timeout:
    connect_timeout: 30s
    read_timeout: 60s
    write_timeout: 30s
    overall_timeout: 120s

  # Retry configuration
  retry:
    max_attempts: 3
    backoff_factor: 2.0
    max_delay: 30s
    retryable_errors:
      - "connection refused"
      - "timeout"
      - "rate limit"
```

## 🏢 Enterprise Configuration

### Complete Enterprise Configuration

```yaml
# Core scanner configuration
scanner:
  api_endpoints:
    - url: "https://api.example.com/v1/users"
      method: "GET"
    - url: "https://api.example.com/v1/data"
      method: "POST"
      body: '{"query": "value"}'

  injection_payloads:
    - "' OR '1'='1"
    - "'; DROP TABLE users;--"

  xss_payloads:
    - "<script>alert('XSS')</script>"
    - "'><script>alert('XSS')</script>"

  rate_limiting:
    requests_per_second: 10
    max_concurrent_requests: 5

# Multi-tenant configuration
tenant:
  id: "enterprise-tenant"
  name: "Enterprise Corporation"
  description: "Enterprise security operations center"
  is_active: true
  settings:
    resource_limits:
      max_requests_per_day: 100000
      max_concurrent_scans: 20
      max_endpoints_per_scan: 500
      max_storage_mb: 10000
      scan_retention_days: 90
    data_isolation:
      storage_path: "./data/enterprise-tenant"
      enabled: true
      encryption_enabled: true
      backup_enabled: true
      backup_schedule: "0 2 * * *"
    notification_settings:
      email_notifications: true
      email_recipients:
        - "security-team@company.com"
        - "admin@company.com"
      webhook_url: "https://hooks.slack.com/services/xxx"
      webhook_events:
        - "scan_completed"
        - "vulnerability_detected"
        - "scan_failed"
      alert_threshold: "medium"
      quiet_hours:
        start: "22:00"
        end: "06:00"
        timezone: "UTC"

# SIEM integration configuration
siem:
  enabled: true
  type: "syslog"
  format: "json"
  endpoint_url: ""
  auth_token: ""
  config:
    host: "wazuh-manager.company.com"
    port: 514
    facility: "local0"
    severity: "info"
    protocol: "udp"
    timeout: 30s
    retry_attempts: 3
    buffer_size: 1000
    flush_interval: 10s

# Advanced authentication configuration
auth:
  enabled: true
  type: "oauth2"
  config:
    # OAuth2 configuration
    client_id: "security-scanner"
    client_secret: "your-client-secret"
    token_url: "https://auth.company.com/oauth/token"
    scopes: ["read", "write", "security"]
    grant_type: "client_credentials"
    audience: "api-security-scanner"

    # JWT configuration
    secret_key: "your-jwt-secret-key"
    signing_method: "HS256"
    audience: "api-security-scanner"
    issuer: "auth.company.com"
    expiration: 3600

    # API key configuration
    header_name: "X-API-Key"
    key_value: "your-api-key"
    key_location: "header"

    # Mutual TLS configuration
    cert_file: "/path/to/client.crt"
    key_file: "/path/to/client.key"
    ca_file: "/path/to/ca.crt"
    skip_verify: false

# Performance metrics configuration
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
    websocket_timeout: 60s
    enable_cors: true
  health_check:
    enabled: true
    interval: 30s
    timeout: 10s
    endpoints:
      - "https://api.example.com/health"
      - "https://auth.example.com/health"
    notification_channels:
      - "email:admin@company.com"
      - "webhook:https://hooks.slack.com/services/xxx"
    failure_threshold: 3

# Server configuration
server:
  port: 8081
  host: "localhost"
  read_timeout: 30s
  write_timeout: 30s
  max_header_bytes: 1048576
  enable_cors: true
  cors_origins:
    - "http://localhost:3000"
    - "https://dashboard.company.com"
  tls:
    enabled: false
    cert_file: "/path/to/server.crt"
    key_file: "/path/to/server.key"
```

## 🔍 Scanner Configuration

### API Endpoints

```yaml
scanner:
  api_endpoints:
    # Basic GET endpoint
    - url: "https://api.example.com/v1/users"
      method: "GET"
      headers:
        "Accept": "application/json"

    # POST endpoint with body
    - url: "https://api.example.com/v1/data"
      method: "POST"
      body: '{"query": "value", "limit": 10}'
      headers:
        "Content-Type": "application/json"

    # PUT endpoint with parameters
    - url: "https://api.example.com/v1/users/{id}"
      method: "PUT"
      body: '{"name": "John", "email": "john@example.com"}'
      headers:
        "Content-Type": "application/json"

    # DELETE endpoint
    - url: "https://api.example.com/v1/users/{id}"
      method: "DELETE"
      headers:
        "Authorization": "Bearer {{token}}"

    # Endpoint with custom timeout
    - url: "https://api.example.com/v1/slow-endpoint"
      method: "GET"
      timeout: 120s
      headers:
        "Accept": "application/json"
```

### Test Payloads

```yaml
scanner:
  # SQL injection payloads
  injection_payloads:
    # Classic SQL injection
    - "' OR '1'='1"
    - "'; DROP TABLE users;--"
    - "1' OR '1'='1"
    - "admin'--"

    # Union-based injection
    - "' UNION SELECT NULL--"
    - "' UNION SELECT username, password FROM users--"

    # Boolean-based injection
    - "' OR 1=1--"
    - "' AND 1=1--"

    # Time-based injection
    - "'; WAITFOR DELAY '0:0:5'--"
    - "'; SLEEP(5)--"

    # Stacked queries
    - "1; DROP TABLE users--"
    - "1; SELECT pg_sleep(5)--"

    # NoSQL injection payloads
  nosql_payloads:
    # MongoDB operators
    - "{$ne: null}"
    - "{$gt: ''}"
    - "{$or: [1,1]}"
    - "{$where: 'sleep(100)'}"
    - "{$regex: '.*'}"
    - "{$exists: true}"
    - "{$in: [1,2,3]}"
    - "{$lt: 100}"

    # JavaScript injection
    - "'; return true; //"
    - "'; while(true) {} //"

    # XSS payloads
  xss_payloads:
    # Basic script injection
    - "<script>alert('XSS')</script>"
    - "'><script>alert('XSS')</script>"

    # Image-based XSS
    - "<img src=x onerror=alert('XSS')>"
    - "<img src='x' onerror='alert(\"XSS\")'>"

    # Event-based XSS
    - "<svg onload=alert('XSS')>"
    - "<body onload=alert('XSS')>"

    # JavaScript protocols
    - "javascript:alert('XSS')"
    - "data:text/html,<script>alert('XSS')</script>"

    # Filter bypass attempts
    - "<ScrIpT>alert('XSS')</sCrIpT>"
    - "<img src=x onerror=alert(String.fromCharCode(88,83,83))>"
    - "'\"<script>alert(String.fromCharCode(88,83,83))</script>"
```

### Authentication Configuration

```yaml
scanner:
  auth:
    # Basic authentication
    type: "basic"
    username: "admin"
    password: "securepassword"

    # Bearer token authentication
    type: "bearer"
    token: "your-bearer-token"
    header: "Authorization"
    prefix: "Bearer"

    # Custom header authentication
    type: "custom"
    header_name: "X-API-Key"
    header_value: "your-api-key"

    # Form-based authentication
    type: "form"
    login_url: "https://api.example.com/login"
    username_field: "username"
    password_field: "password"
    form_data:
      "client_id": "scanner"
      "grant_type": "password"
```

### Rate Limiting

```yaml
scanner:
  rate_limiting:
    # Basic rate limiting
    requests_per_second: 10
    max_concurrent_requests: 5

    # Advanced rate limiting
    burst_limit: 15
    cooldown_period: 5s

    # Per-endpoint rate limiting
    endpoint_limits:
      "https://api.example.com/v1/users":
        requests_per_second: 5
        max_concurrent_requests: 2
      "https://api.example.com/v1/data":
        requests_per_second: 3
        max_concurrent_requests: 1

    # Time-based rate limiting
    time_windows:
      - window: "1m"
        limit: 100
      - window: "1h"
        limit: 1000
      - window: "1d"
        limit: 10000
```

## 🏢 Multi-Tenant Configuration

### Tenant Configuration

```yaml
tenant:
  # Basic tenant information
  id: "enterprise-tenant"
  name: "Enterprise Corporation"
  description: "Enterprise security operations center"
  is_active: true
  created_at: "2024-01-01T00:00:00Z"
  updated_at: "2024-01-01T00:00:00Z"

  # Tenant settings
  settings:
    # Resource limits
    resource_limits:
      max_requests_per_day: 100000
      max_concurrent_scans: 20
      max_endpoints_per_scan: 500
      max_storage_mb: 10000
      scan_retention_days: 90
      max_api_calls_per_minute: 1000
      max_concurrent_api_calls: 50

    # Data isolation
    data_isolation:
      storage_path: "./data/enterprise-tenant"
      enabled: true
      encryption_enabled: true
      encryption_key: "your-encryption-key"
      backup_enabled: true
      backup_schedule: "0 2 * * *"
      backup_retention_days: 30
      compression_enabled: true

    # Notification settings
    notification_settings:
      email_notifications: true
      email_recipients:
        - "security-team@company.com"
        - "admin@company.com"
        - "alerts@company.com"
      email_subject_prefix: "[Security Scanner]"
      webhook_url: "https://hooks.slack.com/services/xxx"
      webhook_events:
        - "scan_started"
        - "scan_completed"
        - "vulnerability_detected"
        - "scan_failed"
        - "quota_exceeded"
      alert_threshold: "medium"
      quiet_hours:
        start: "22:00"
        end: "06:00"
        timezone: "UTC"
      weekend_quiet_hours: true

    # Custom configuration
    custom_config:
      default_scan_profile: "comprehensive"
      auto_scan_enabled: false
      scan_schedule: "0 2 * * *"
      compliance_framework: "SOC2"
      custom_payloads: "./payloads/tenant-specific.txt"
      exclude_patterns:
        - "/internal/"
        - "/admin/"
        - "/health"
```

### Multi-Tenant Examples

```yaml
# Production tenant
tenant:
  id: "production"
  name: "Production Environment"
  description: "Production security monitoring"
  is_active: true
  settings:
    resource_limits:
      max_requests_per_day: 50000
      max_concurrent_scans: 10
      max_endpoints_per_scan: 200
    data_isolation:
      storage_path: "./data/production"
      enabled: true
      encryption_enabled: true
    notification_settings:
      email_notifications: true
      email_recipients:
        - "prod-security@company.com"
      alert_threshold: "high"

# Development tenant
tenant:
  id: "development"
  name: "Development Environment"
  description: "Development security testing"
  is_active: true
  settings:
    resource_limits:
      max_requests_per_day: 10000
      max_concurrent_scans: 5
      max_endpoints_per_scan: 50
    data_isolation:
      storage_path: "./data/development"
      enabled: true
      encryption_enabled: false
    notification_settings:
      email_notifications: false
      alert_threshold: "low"

# Testing tenant
tenant:
  id: "testing"
  name: "Testing Environment"
  description: "QA and testing security scans"
  is_active: true
  settings:
    resource_limits:
      max_requests_per_day: 20000
      max_concurrent_scans: 8
      max_endpoints_per_scan: 100
    data_isolation:
      storage_path: "./data/testing"
      enabled: true
      encryption_enabled: false
    notification_settings:
      email_notifications: true
      email_recipients:
        - "qa-team@company.com"
      alert_threshold: "medium"
```

## 🔌 SIEM Configuration

### Wazuh SIEM Configuration

```yaml
siem:
  enabled: true
  type: "syslog"
  format: "json"
  endpoint_url: ""
  auth_token: ""
  config:
    # Wazuh manager connection
    host: "wazuh-manager.company.com"
    port: 514
    facility: "local0"
    severity: "info"
    protocol: "udp"

    # Connection settings
    timeout: 30s
    retry_attempts: 3
    retry_delay: 5s
    buffer_size: 1000
    flush_interval: 10s

    # Syslog settings
    hostname: "api-scanner"
    app_name: "api-security-scanner"
    pid: "{{pid}}"
    message_id: "api-security-event"

    # Event filtering
    include_events:
      - "vulnerability_detected"
      - "scan_completed"
      - "scan_failed"
      - "auth_bypass_detected"
    exclude_events:
      - "debug"
      - "heartbeat"

    # Event formatting
    event_format:
      timestamp_format: "2006-01-02T15:04:05Z07:00"
      include_hostname: true
      include_app_name: true
      include_pid: true
      structured_data: true
```

### Splunk SIEM Configuration

```yaml
siem:
  enabled: true
  type: "splunk"
  format: "json"
  endpoint_url: "https://splunk.company.com:8088/services/collector"
  auth_token: "your-splunk-hec-token"
  config:
    # Splunk HEC settings
    host: "splunk.company.com"
    port: 8088
    token: "your-splunk-hec-token"
    index: "api_security"
    source: "api-security-scanner"
    sourcetype: "api:security:scanner"

    # Connection settings
    timeout: 30s
    retry_attempts: 3
    retry_delay: 5s
    verify_ssl: true
    compression: true

    # Event formatting
    event_format:
      timestamp_field: "timestamp"
      host_field: "host"
      source_field: "source"
      sourcetype_field: "sourcetype"
      index_field: "index"

    # Batch settings
    batch_size: 100
    batch_timeout: 10s
    max_queue_size: 10000
```

### ELK Stack Configuration

```yaml
siem:
  enabled: true
  type: "elk"
  format: "json"
  endpoint_url: "https://elasticsearch.company.com:9200/api-scanner-events/_doc"
  auth_token: "your-elasticsearch-token"
  config:
    # Elasticsearch settings
    hosts:
      - "https://elasticsearch.company.com:9200"
    username: "api-scanner"
    password: "your-password"
    index: "api-scanner-%{+YYYY.MM.dd}"
    index_template: "api-scanner-template"

    # Connection settings
    timeout: 30s
    retry_attempts: 3
    retry_delay: 5s
    verify_ssl: true
    compression: true

    # Document settings
    document_id: "{{uuid}}"
    timestamp_field: "@timestamp"
    pipeline: "api-security-pipeline"

    # Bulk settings
    bulk_size: 1000
    bulk_timeout: 10s
    flush_interval: 10s
```

### IBM QRadar Configuration

```yaml
siem:
  enabled: true
  type: "qradar"
  format: "cef"
  endpoint_url: "https://qradar.company.com:8418"
  auth_token: "your-qradar-token"
  config:
    # QRadar settings
    host: "qradar.company.com"
    port: 8418
    token: "your-qradar-token"
    endpoint: "/api/ariel/events"

    # Connection settings
    timeout: 30s
    retry_attempts: 3
    retry_delay: 5s
    verify_ssl: true

    # CEF formatting
    cef_format:
      version: 0
      device_vendor: "API-Security-Scanner"
      device_product: "API-Security-Scanner"
      device_version: "4.0.0"
      signature_id: "{{event_id}}"
      name: "{{event_name}}"
      severity: "{{severity}}"

      # CEF extensions
      extensions:
        cs1: "{{tenant_id}}"
        cs1Label: "TenantID"
        cs2: "{{vulnerability_type}}"
        cs2Label: "VulnerabilityType"
        cs3: "{{target_url}}"
        cs3Label: "TargetURL"
        cs4: "{{method}}"
        cs4Label: "Method"
        cn1: "{{score}}"
        cn1Label: "Score"
        msg: "{{description}}"
```

### ArcSight Configuration

```yaml
siem:
  enabled: true
  type: "arcsight"
  format: "cef"
  endpoint_url: "https://arcsight.company.com:8443"
  auth_token: "your-arcsight-token"
  config:
    # ArcSight settings
    host: "arcsight.company.com"
    port: 8443
    username: "api-scanner"
    password: "your-password"
    connector: "api-security-scanner-connector"

    # Connection settings
    timeout: 30s
    retry_attempts: 3
    retry_delay: 5s
    verify_ssl: true

    # CEF formatting
    cef_format:
      version: 0
      device_vendor: "API-Security-Scanner"
      device_product: "API-Security-Scanner"
      device_version: "4.0.0"
      signature_id: "{{event_id}}"
      name: "{{event_name}}"
      severity: "{{severity}}"

      # ArcSight specific fields
      device_event_category: "Security"
      device_external_id: "{{scan_id}}"
      facility: "Security"

      # CEF extensions
      extensions:
        cs1: "{{tenant_id}}"
        cs1Label: "TenantID"
        cs2: "{{vulnerability_type}}"
        cs2Label: "VulnerabilityType"
        cs3: "{{target_url}}"
        cs3Label: "TargetURL"
        cs4: "{{method}}"
        cs4Label: "Method"
        cn1: "{{score}}"
        cn1Label: "Score"
        msg: "{{description}}"
```

## 🔐 Authentication Configuration

### OAuth2 Configuration

```yaml
auth:
  enabled: true
  type: "oauth2"
  config:
    # OAuth2 client credentials
    client_id: "security-scanner"
    client_secret: "your-client-secret"
    token_url: "https://auth.company.com/oauth/token"
    refresh_url: "https://auth.company.com/oauth/refresh"

    # OAuth2 grant types
    grant_type: "client_credentials"
    scopes: ["read", "write", "security"]
    audience: "api-security-scanner"

    # Token settings
    token_lifetime: 3600
    refresh_lifetime: 86400
    auto_refresh: true

    # PKCE settings (for authorization code flow)
    use_pkce: false
    code_challenge_method: "S256"

    # Additional parameters
    additional_params:
      "resource": "https://api.company.com"
      "access_type": "offline"
```

### JWT Configuration

```yaml
auth:
  enabled: true
  type: "jwt"
  config:
    # JWT settings
    secret_key: "your-jwt-secret-key"
    signing_method: "HS256"
    audience: "api-security-scanner"
    issuer: "auth.company.com"
    expiration: 3600

    # Public key (for RS256)
    public_key: |
      -----BEGIN PUBLIC KEY-----
      MIIBIjANBgkqhkiG9w0BAQEFAAOCAQ8AMIIBCgKCAQEA...
      -----END PUBLIC KEY-----

    # Private key (for signing)
    private_key: |
      -----BEGIN PRIVATE KEY-----
      MIIEvgIBADANBgkqhkiG9w0BAQEFAASCBKgwggSkAgEA...
      -----END PRIVATE KEY-----

    # Claims
    claims:
      sub: "api-security-scanner"
      name: "API Security Scanner"
      email: "scanner@company.com"
      roles: ["scanner", "security"]

    # Token validation
    validate_claims: true
    validate_expiry: true
    validate_signature: true

    # Header settings
    header_name: "Authorization"
    header_prefix: "Bearer"
```

### API Key Configuration

```yaml
auth:
  enabled: true
  type: "api_key"
  config:
    # API key settings
    key_value: "your-api-key"
    header_name: "X-API-Key"
    key_location: "header"

    # Query parameter settings
    query_param: "api_key"

    # Multiple API keys
    api_keys:
      - key: "primary-key"
        description: "Primary API key"
        rate_limit: 1000
      - key: "secondary-key"
        description: "Secondary API key"
        rate_limit: 500

    # Key rotation
    key_rotation:
      enabled: true
      rotation_period: "30d"
      grace_period: "7d"

    # Key validation
    validation:
      length_min: 32
      length_max: 256
      require_complexity: true
      allowed_characters: "a-zA-Z0-9-._~+/="
```

### Mutual TLS Configuration

```yaml
auth:
  enabled: true
  type: "mutual_tls"
  config:
    # Certificate files
    cert_file: "/path/to/client.crt"
    key_file: "/path/to/client.key"
    ca_file: "/path/to/ca.crt"

    # Certificate settings
    cert_password: "your-cert-password"
    key_password: "your-key-password"

    # TLS settings
    skip_verify: false
    server_name: "api.company.com"
    min_version: "1.2"
    max_version: "1.3"

    # Certificate validation
    validate_cert: true
    validate_chain: true
    validate_hostname: true

    # Certificate rotation
    rotation:
      enabled: true
      check_interval: "24h"
      expiry_warning: "7d"

    # CA bundle
    ca_bundle:
      - "/path/to/ca1.crt"
      - "/path/to/ca2.crt"
```

### Basic Authentication Configuration

```yaml
auth:
  enabled: true
  type: "basic"
  config:
    # Basic auth credentials
    username: "api-scanner"
    password: "your-password"

    # Multiple credentials
    credentials:
      - username: "scanner1"
        password: "password1"
        description: "Primary scanner account"
      - username: "scanner2"
        password: "password2"
        description: "Secondary scanner account"

    # Credential rotation
    rotation:
      enabled: true
      rotation_period: "90d"
      notification_period: "7d"

    # Header settings
    header_name: "Authorization"
    encoding: "base64"

    # Security settings
    password_policy:
      min_length: 12
      require_uppercase: true
      require_lowercase: true
      require_numbers: true
      require_special: true
```

## 📊 Metrics Configuration

### Basic Metrics Configuration

```yaml
metrics:
  enabled: true
  port: 8080
  update_interval: 30s
  retention_days: 30

  # Metrics collection
  collection:
    system_metrics: true
    security_metrics: true
    business_metrics: true
    scanner_metrics: true

  # Storage
  storage:
    type: "file"
    path: "./metrics"
    compression: true
    rotation: true
    max_file_size: "100MB"
    max_files: 10
```

### Advanced Metrics Configuration

```yaml
metrics:
  enabled: true
  port: 8080
  update_interval: 30s
  retention_days: 30

  # Dashboard configuration
  dashboard:
    enabled: true
    port: 8081
    host: "localhost"
    update_interval: 5s
    max_connections: 100
    websocket_timeout: 60s
    enable_cors: true
    cors_origins:
      - "http://localhost:3000"
      - "https://dashboard.company.com"

    # Authentication
    auth:
      enabled: true
      type: "basic"
      username: "admin"
      password: "admin-password"

    # UI settings
    ui:
      theme: "dark"
      refresh_interval: 10s
      max_data_points: 1000
      default_time_range: "1h"

    # Widgets
    widgets:
      - name: "system_cpu"
        type: "gauge"
        title: "CPU Usage"
        unit: "%"
      - name: "system_memory"
        type: "gauge"
        title: "Memory Usage"
        unit: "%"
      - name: "vulnerability_count"
        type: "counter"
        title: "Vulnerabilities Detected"
      - name: "scan_success_rate"
        type: "percentage"
        title: "Scan Success Rate"

  # Health check configuration
  health_check:
    enabled: true
    interval: 30s
    timeout: 10s
    endpoints:
      - url: "https://api.example.com/health"
        name: "API Health"
        expected_status: 200
      - url: "https://auth.example.com/health"
        name: "Auth Health"
        expected_status: 200
        headers:
          "Authorization": "Bearer {{token}}"

    # Notification channels
    notification_channels:
      - type: "email"
        config:
          recipients:
            - "admin@company.com"
          subject: "Health Check Alert"
          body: "Health check failed for {{endpoint}}"
      - type: "webhook"
        config:
          url: "https://hooks.slack.com/services/xxx"
          message: "Health check failed for {{endpoint}}"

    # Failure handling
    failure_threshold: 3
    recovery_threshold: 2
    alert_on_recovery: true

  # Metrics export
  export:
    enabled: true
    formats:
      - type: "prometheus"
        endpoint: "/metrics"
        format: "prometheus"
      - type: "json"
        endpoint: "/metrics.json"
        format: "json"
      - type: "influxdb"
        config:
          url: "http://influxdb:8086"
          database: "api_scanner"
          username: "scanner"
          password: "your-password"
          retention_policy: "30d"

  # Alerts
  alerts:
    enabled: true
    rules:
      - name: "high_cpu_usage"
        condition: "system.cpu_usage > 80"
        duration: "5m"
        severity: "warning"
        message: "High CPU usage detected: {{value}}%"
      - name: "high_memory_usage"
        condition: "system.memory_usage > 85"
        duration: "5m"
        severity: "critical"
        message: "High memory usage detected: {{value}}%"
      - name: "vulnerability_spike"
        condition: "security.vulnerability_count > 10"
        duration: "1m"
        severity: "critical"
        message: "Vulnerability spike detected: {{value}} vulnerabilities"

    # Notification channels
    notification_channels:
      - type: "email"
        config:
          smtp_server: "smtp.company.com"
          smtp_port: 587
          username: "alerts@company.com"
          password: "your-password"
          recipients:
            - "security-team@company.com"
      - type: "webhook"
        config:
          url: "https://hooks.slack.com/services/xxx"
          headers:
            "Content-Type": "application/json"
      - type: "pagerduty"
        config:
          service_key: "your-pagerduty-key"
          severity: "critical"
```

## 🌐 Server Configuration

### Basic Server Configuration

```yaml
server:
  port: 8081
  host: "localhost"
  read_timeout: 30s
  write_timeout: 30s
  max_header_bytes: 1048576
```

### Advanced Server Configuration

```yaml
server:
  # Basic settings
  port: 8081
  host: "localhost"
  read_timeout: 30s
  write_timeout: 30s
  idle_timeout: 60s
  max_header_bytes: 1048576

  # TLS settings
  tls:
    enabled: true
    cert_file: "/path/to/server.crt"
    key_file: "/path/to/server.key"
    min_version: "1.2"
    max_version: "1.3"
    cipher_suites:
      - "TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384"
      - "TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256"
    client_auth: false
    client_ca_file: "/path/to/client-ca.crt"

  # CORS settings
  cors:
    enabled: true
    allowed_origins:
      - "http://localhost:3000"
      - "https://dashboard.company.com"
      - "https://app.company.com"
    allowed_methods:
      - "GET"
      - "POST"
      - "PUT"
      - "DELETE"
      - "OPTIONS"
    allowed_headers:
      - "Origin"
      - "Content-Type"
      - "Accept"
      - "Authorization"
      - "X-Requested-With"
    exposed_headers:
      - "Content-Length"
      - "Content-Type"
      - "X-Request-ID"
    allow_credentials: true
    max_age: 86400

  # Rate limiting
  rate_limiting:
    enabled: true
    requests_per_second: 100
    burst: 200
    cleanup_interval: 5m

  # Middleware
  middleware:
    # Request logging
    logging:
      enabled: true
      format: "json"
      level: "info"
      include_request_body: false
      include_response_body: false

    # Request ID
    request_id:
      enabled: true
      header: "X-Request-ID"
      length: 16

    # Compression
    compression:
      enabled: true
      level: 5
      min_length: 1024

    # Security headers
    security_headers:
      enabled: true
      headers:
        "X-Content-Type-Options": "nosniff"
        "X-Frame-Options": "DENY"
        "X-XSS-Protection": "1; mode=block"
        "Strict-Transport-Security": "max-age=31536000; includeSubDomains"
        "Content-Security-Policy": "default-src 'self'"
        "Referrer-Policy": "strict-origin-when-cross-origin"
        "Permissions-Policy": "camera=(), microphone=(), geolocation=()"

  # API endpoints
  endpoints:
    # Configuration API
    config:
      enabled: true
      path: "/api/config"
      methods: ["GET", "POST"]
      auth_required: true

    # Tenant API
    tenants:
      enabled: true
      path: "/api/tenants"
      methods: ["GET", "POST", "PUT", "DELETE"]
      auth_required: true

    # Metrics API
    metrics:
      enabled: true
      path: "/api/metrics"
      methods: ["GET"]
      auth_required: true

    # Scan API
    scan:
      enabled: true
      path: "/api/scan"
      methods: ["POST", "GET"]
      auth_required: true

    # Health check
    health:
      enabled: true
      path: "/health"
      methods: ["GET"]
      auth_required: false

  # Static files
  static_files:
    enabled: true
    path: "./static"
    strip_prefix: false
    index: "index.html"
    spa: true

  # Graceful shutdown
  shutdown:
    timeout: 30s
    drain_timeout: 10s
    wait_before_shutdown: 5s
```

## 🌍 Environment Variables

### Configuration Environment Variables

```bash
# Server configuration
SERVER_PORT=8081
SERVER_HOST=localhost
SERVER_READ_TIMEOUT=30s
SERVER_WRITE_TIMEOUT=30s

# Scanner configuration
SCANNER_RATE_LIMITING_REQUESTS_PER_SECOND=10
SCANNER_RATE_LIMITING_MAX_CONCURRENT_REQUESTS=5

# Authentication configuration
AUTH_ENABLED=true
AUTH_TYPE=basic
AUTH_USERNAME=admin
AUTH_PASSWORD=your-password

# SIEM configuration
SIEM_ENABLED=true
SIEM_TYPE=syslog
SIEM_HOST=localhost
SIEM_PORT=514

# Metrics configuration
METRICS_ENABLED=true
METRICS_PORT=8080
METRICS_DASHBOARD_ENABLED=true
METRICS_DASHBOARD_PORT=8081

# Tenant configuration
TENANT_ID=default
TENANT_NAME=Default Tenant
TENANT_ACTIVE=true

# Database configuration (if using database storage)
DATABASE_URL=postgresql://user:password@localhost:5432/api_scanner
DATABASE_SSL_MODE=disable
DATABASE_MAX_CONNECTIONS=25

# Logging configuration
LOG_LEVEL=info
LOG_FORMAT=json
LOG_FILE=logs/app.log

# Security configuration
SECRET_KEY=your-secret-key
JWT_SECRET=your-jwt-secret
ENCRYPTION_KEY=your-encryption-key

# TLS configuration
TLS_ENABLED=false
TLS_CERT_FILE=/path/to/server.crt
TLS_KEY_FILE=/path/to/server.key

# CORS configuration
CORS_ENABLED=true
CORS_ALLOWED_ORIGINS=http://localhost:3000,https://dashboard.company.com
```

### Environment Variable Override Order

1. Environment variables (highest priority)
2. Configuration file values
3. Default values (lowest priority)

## ✅ Configuration Validation

### Configuration Validation

```bash
# Validate configuration file
./api-security-scanner -config config.yaml -validate-config

# Validate configuration with verbose output
./api-security-scanner -config config.yaml -validate-config -verbose

# Test specific components
./api-security-scanner -config config.yaml -test-auth
./api-security-scanner -config config.yaml -test-siem
./api-security-scanner -config config.yaml -test-metrics
```

### Configuration Validation Rules

#### Required Fields

```yaml
# Required fields in configuration
required_fields:
  scanner:
    - "api_endpoints"
    - "rate_limiting"
  tenant:
    - "id"
    - "name"
    - "is_active"
  server:
    - "port"
    - "host"
```

#### Field Validation

```yaml
# Field validation rules
validation_rules:
  scanner:
    api_endpoints:
      - field: "url"
        type: "string"
        pattern: "^https?://.+"
        required: true
      - field: "method"
        type: "string"
        enum: ["GET", "POST", "PUT", "DELETE", "PATCH", "HEAD", "OPTIONS"]
        required: true
    rate_limiting:
      - field: "requests_per_second"
        type: "integer"
        min: 1
        max: 1000
        required: true
      - field: "max_concurrent_requests"
        type: "integer"
        min: 1
        max: 100
        required: true

  tenant:
    - field: "id"
      type: "string"
      pattern: "^[a-zA-Z0-9_-]+$"
      required: true
    - field: "name"
      type: "string"
      min_length: 1
      max_length: 100
      required: true
    - field: "is_active"
      type: "boolean"
      required: true

  server:
    - field: "port"
      type: "integer"
      min: 1
      max: 65535
      required: true
    - field: "host"
      type: "string"
      pattern: "^[a-zA-Z0-9.-]+$"
      required: true
```

## 📝 Configuration Examples

### Example 1: Basic Configuration

```yaml
scanner:
  api_endpoints:
    - url: "https://api.example.com/users"
      method: "GET"
    - url: "https://api.example.com/data"
      method: "POST"
      body: '{"query": "value"}'

  injection_payloads:
    - "' OR '1'='1"
    - "'; DROP TABLE users;--"

  xss_payloads:
    - "<script>alert('XSS')</script>"
    - "'><script>alert('XSS')</script>"

  rate_limiting:
    requests_per_second: 10
    max_concurrent_requests: 5

server:
  port: 8081
  host: "localhost"
```

### Example 2: Enterprise Configuration

```yaml
scanner:
  api_endpoints:
    - url: "https://api.company.com/v1/users"
      method: "GET"
    - url: "https://api.company.com/v1/data"
      method: "POST"
      body: '{"query": "value"}'

  injection_payloads:
    - "' OR '1'='1"
    - "'; DROP TABLE users;--"

  xss_payloads:
    - "<script>alert('XSS')</script>"
    - "'><script>alert('XSS')</script>"

  rate_limiting:
    requests_per_second: 10
    max_concurrent_requests: 5

tenant:
  id: "enterprise"
  name: "Enterprise Corp"
  description: "Enterprise security team"
  is_active: true
  settings:
    resource_limits:
      max_requests_per_day: 10000
      max_concurrent_scans: 5
      max_endpoints_per_scan: 100
    data_isolation:
      storage_path: "./data/enterprise"
      enabled: true

siem:
  enabled: true
  type: "syslog"
  format: "json"
  config:
    host: "wazuh.company.com"
    port: 514
    facility: "local0"
    severity: "info"

auth:
  enabled: true
  type: "oauth2"
  config:
    client_id: "scanner"
    client_secret: "your-secret"
    token_url: "https://auth.company.com/oauth/token"
    scopes: ["read", "write"]

metrics:
  enabled: true
  port: 8080
  dashboard:
    enabled: true
    port: 8081

server:
  port: 8081
  host: "localhost"
  tls:
    enabled: true
    cert_file: "/path/to/server.crt"
    key_file: "/path/to/server.key"
```

### Example 3: Multi-Tenant Configuration

```yaml
# Primary configuration for multi-tenant setup
scanner:
  api_endpoints:
    - url: "https://api.company.com/v1/health"
      method: "GET"

  rate_limiting:
    requests_per_second: 5
    max_concurrent_requests: 2

tenant:
  id: "management"
  name: "Management Tenant"
  description: "Multi-tenant management"
  is_active: true
  settings:
    resource_limits:
      max_requests_per_day: 5000
      max_concurrent_scans: 3
      max_endpoints_per_scan: 50
    data_isolation:
      storage_path: "./data/management"
      enabled: true

# Additional tenants would be managed through the API or separate config files
```

### Example 4: Development Configuration

```yaml
scanner:
  api_endpoints:
    - url: "http://localhost:3000/api/users"
      method: "GET"
    - url: "http://localhost:3000/api/data"
      method: "POST"
      body: '{"query": "test"}'

  injection_payloads:
    - "' OR '1'='1"
    - "test' OR '1'='1"

  xss_payloads:
    - "<script>alert('XSS')</script>"
    - "test<script>alert('XSS')</script>"

  rate_limiting:
    requests_per_second: 20
    max_concurrent_requests: 10

server:
  port: 8081
  host: "localhost"

metrics:
  enabled: false

siem:
  enabled: false

auth:
  enabled: false
```

### Example 5: Production Configuration

```yaml
scanner:
  api_endpoints:
    - url: "https://api.company.com/v1/users"
      method: "GET"
      headers:
        "Accept": "application/json"
    - url: "https://api.company.com/v1/data"
      method: "POST"
      body: '{"query": "value"}'
      headers:
        "Content-Type": "application/json"

  injection_payloads:
    - "' OR '1'='1"
    - "'; DROP TABLE users;--"
    - "' UNION SELECT NULL--"

  xss_payloads:
    - "<script>alert('XSS')</script>"
    - "'><script>alert('XSS')</script>"
    - "<img src=x onerror=alert('XSS')>"

  rate_limiting:
    requests_per_second: 10
    max_concurrent_requests: 5

tenant:
  id: "production"
  name: "Production Environment"
  description: "Production security monitoring"
  is_active: true
  settings:
    resource_limits:
      max_requests_per_day: 50000
      max_concurrent_scans: 10
      max_endpoints_per_scan: 200
    data_isolation:
      storage_path: "./data/production"
      enabled: true
      encryption_enabled: true
    notification_settings:
      email_notifications: true
      email_recipients:
        - "prod-security@company.com"
      alert_threshold: "high"

siem:
  enabled: true
  type: "splunk"
  format: "json"
  endpoint_url: "https://splunk.company.com:8088/services/collector"
  auth_token: "your-splunk-token"

auth:
  enabled: true
  type: "oauth2"
  config:
    client_id: "scanner"
    client_secret: "your-client-secret"
    token_url: "https://auth.company.com/oauth/token"
    scopes: ["read", "write", "security"]

metrics:
  enabled: true
  port: 8080
  dashboard:
    enabled: true
    port: 8081
  health_check:
    enabled: true
    interval: 30s
    endpoints:
      - "https://api.company.com/health"

server:
  port: 8081
  host: "0.0.0.0"
  tls:
    enabled: true
    cert_file: "/etc/ssl/certs/server.crt"
    key_file: "/etc/ssl/private/server.key"
  cors:
    enabled: true
    allowed_origins:
      - "https://dashboard.company.com"
```

---

This comprehensive configuration guide covers all aspects of the API Security Scanner's configuration system. For additional examples and specific use cases, refer to the example configuration files in the repository.