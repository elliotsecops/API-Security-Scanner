# API Security Scanner - API Documentation

## 📋 Table of Contents

1. [Overview](#overview)
2. [Authentication](#authentication)
3. [Configuration API](#configuration-api)
4. [Tenant Management API](#tenant-management-api)
5. [Scan Management API](#scan-management-api)
6. [Metrics API](#metrics-api)
7. [SIEM API](#siem-api)
8. [Health Check API](#health-check-api)
9. [WebSocket API](#websocket-api)
10. [Error Responses](#error-responses)
11. [Rate Limiting](#rate-limiting)
12. [Examples](#examples)

## 🎯 Overview

The API Security Scanner provides a comprehensive REST API for managing security scans, configurations, tenants, and enterprise features. This documentation covers all available endpoints, authentication methods, and usage examples.

### Base URL

```
http://localhost:8081/api
```

### API Version

Current API version: **v1**

### Content Types

All API requests and responses use JSON format:

- **Content-Type**: `application/json`
- **Accept**: `application/json`

## 🔐 Authentication

### Basic Authentication

```bash
curl -X GET http://localhost:8081/api/config \
  -H "Authorization: Basic $(echo -n 'admin:password' | base64)"
```

### Bearer Token Authentication

```bash
curl -X GET http://localhost:8081/api/config \
  -H "Authorization: Bearer your-token-here"
```

### API Key Authentication

```bash
curl -X GET http://localhost:8081/api/config \
  -H "X-API-Key: your-api-key-here"
```

### OAuth2 Authentication

```bash
# Get access token
curl -X POST https://auth.company.com/oauth/token \
  -H "Content-Type: application/json" \
  -d '{
    "grant_type": "client_credentials",
    "client_id": "your-client-id",
    "client_secret": "your-client-secret",
    "scope": "read write"
  }'

# Use access token
curl -X GET http://localhost:8081/api/config \
  -H "Authorization: Bearer access-token"
```

## ⚙️ Configuration API

### Get Configuration

Retrieve the current configuration.

**Endpoint:** `GET /api/config`

**Response:**
```json
{
  "scanner": {
    "api_endpoints": [
      {
        "url": "https://api.example.com/users",
        "method": "GET"
      }
    ],
    "injection_payloads": ["' OR '1'='1"],
    "xss_payloads": ["<script>alert('XSS')</script>"],
    "rate_limiting": {
      "requests_per_second": 10,
      "max_concurrent_requests": 5
    }
  },
  "tenant": {
    "id": "default",
    "name": "Default Tenant",
    "is_active": true
  },
  "server": {
    "port": 8081,
    "host": "localhost"
  }
}
```

**Example:**
```bash
curl -X GET http://localhost:8081/api/config \
  -H "Authorization: Bearer your-token"
```

### Update Configuration

Update the current configuration.

**Endpoint:** `POST /api/config`

**Request Body:**
```json
{
  "scanner": {
    "api_endpoints": [
      {
        "url": "https://api.example.com/users",
        "method": "GET"
      },
      {
        "url": "https://api.example.com/data",
        "method": "POST",
        "body": "{\"query\": \"value\"}"
      }
    ],
    "rate_limiting": {
      "requests_per_second": 15,
      "max_concurrent_requests": 8
    }
  }
}
```

**Response:**
```json
{
  "success": true,
  "message": "Configuration updated successfully",
  "config": {
    "scanner": {
      "api_endpoints": [
        {
          "url": "https://api.example.com/users",
          "method": "GET"
        },
        {
          "url": "https://api.example.com/data",
          "method": "POST",
          "body": "{\"query\": \"value\"}"
        }
      ],
      "rate_limiting": {
        "requests_per_second": 15,
        "max_concurrent_requests": 8
      }
    }
  }
}
```

**Example:**
```bash
curl -X POST http://localhost:8081/api/config \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer your-token" \
  -d '{
    "scanner": {
      "api_endpoints": [
        {
          "url": "https://api.example.com/users",
          "method": "GET"
        }
      ],
      "rate_limiting": {
        "requests_per_second": 15,
        "max_concurrent_requests": 8
      }
    }
  }'
```

### Validate Configuration

Validate the current configuration.

**Endpoint:** `POST /api/config/validate`

**Response:**
```json
{
  "valid": true,
  "errors": [],
  "warnings": [],
  "details": {
    "scanner": {
      "valid": true,
      "errors": [],
      "warnings": []
    },
    "tenant": {
      "valid": true,
      "errors": [],
      "warnings": []
    }
  }
}
```

**Example:**
```bash
curl -X POST http://localhost:8081/api/config/validate \
  -H "Authorization: Bearer your-token"
```

## 🏢 Tenant Management API

### List Tenants

Retrieve all tenants.

**Endpoint:** `GET /api/tenants`

**Query Parameters:**
- `page` (optional): Page number (default: 1)
- `limit` (optional): Items per page (default: 10)
- `active` (optional): Filter by active status (true/false)
- `search` (optional): Search by name or ID

**Response:**
```json
{
  "tenants": [
    {
      "id": "tenant-001",
      "name": "Acme Corporation",
      "description": "Enterprise security team",
      "is_active": true,
      "created_at": "2024-01-01T00:00:00Z",
      "updated_at": "2024-01-01T00:00:00Z",
      "settings": {
        "resource_limits": {
          "max_requests_per_day": 10000,
          "max_concurrent_scans": 5,
          "max_endpoints_per_scan": 100
        }
      }
    }
  ],
  "pagination": {
    "page": 1,
    "limit": 10,
    "total": 1,
    "total_pages": 1
  }
}
```

**Example:**
```bash
curl -X GET "http://localhost:8081/api/tenants?page=1&limit=10&active=true" \
  -H "Authorization: Bearer your-token"
```

### Get Tenant

Retrieve a specific tenant by ID.

**Endpoint:** `GET /api/tenants/{tenant_id}`

**Response:**
```json
{
  "id": "tenant-001",
  "name": "Acme Corporation",
  "description": "Enterprise security team",
  "is_active": true,
  "created_at": "2024-01-01T00:00:00Z",
  "updated_at": "2024-01-01T00:00:00Z",
  "settings": {
    "resource_limits": {
      "max_requests_per_day": 10000,
      "max_concurrent_scans": 5,
      "max_endpoints_per_scan": 100,
      "max_storage_mb": 1000,
      "scan_retention_days": 30
    },
    "data_isolation": {
      "storage_path": "./data/tenant-001",
      "enabled": true,
      "encryption_enabled": true
    },
    "notification_settings": {
      "email_notifications": true,
      "email_recipients": ["security@company.com"],
      "alert_threshold": "medium"
    }
  },
  "stats": {
    "total_scans": 150,
    "vulnerabilities_found": 25,
    "last_scan": "2024-01-15T10:30:00Z"
  }
}
```

**Example:**
```bash
curl -X GET http://localhost:8081/api/tenants/tenant-001 \
  -H "Authorization: Bearer your-token"
```

### Create Tenant

Create a new tenant.

**Endpoint:** `POST /api/tenants`

**Request Body:**
```json
{
  "id": "new-tenant",
  "name": "New Organization",
  "description": "Security operations center",
  "is_active": true,
  "settings": {
    "resource_limits": {
      "max_requests_per_day": 5000,
      "max_concurrent_scans": 3,
      "max_endpoints_per_scan": 50
    },
    "data_isolation": {
      "storage_path": "./data/new-tenant",
      "enabled": true
    },
    "notification_settings": {
      "email_notifications": true,
      "email_recipients": ["security@company.com"],
      "alert_threshold": "medium"
    }
  }
}
```

**Response:**
```json
{
  "success": true,
  "message": "Tenant created successfully",
  "tenant": {
    "id": "new-tenant",
    "name": "New Organization",
    "description": "Security operations center",
    "is_active": true,
    "created_at": "2024-01-15T10:30:00Z",
    "updated_at": "2024-01-15T10:30:00Z",
    "settings": {
      "resource_limits": {
        "max_requests_per_day": 5000,
        "max_concurrent_scans": 3,
        "max_endpoints_per_scan": 50
      }
    }
  }
}
```

**Example:**
```bash
curl -X POST http://localhost:8081/api/tenants \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer your-token" \
  -d '{
    "id": "new-tenant",
    "name": "New Organization",
    "description": "Security operations center",
    "is_active": true,
    "settings": {
      "resource_limits": {
        "max_requests_per_day": 5000,
        "max_concurrent_scans": 3,
        "max_endpoints_per_scan": 50
      }
    }
  }'
```

### Update Tenant

Update an existing tenant.

**Endpoint:** `PUT /api/tenants/{tenant_id}`

**Request Body:**
```json
{
  "name": "Updated Organization",
  "description": "Updated security operations center",
  "is_active": true,
  "settings": {
    "resource_limits": {
      "max_requests_per_day": 15000,
      "max_concurrent_scans": 8,
      "max_endpoints_per_scan": 150
    },
    "notification_settings": {
      "email_notifications": true,
      "email_recipients": ["security@company.com", "admin@company.com"],
      "alert_threshold": "high"
    }
  }
}
```

**Response:**
```json
{
  "success": true,
  "message": "Tenant updated successfully",
  "tenant": {
    "id": "tenant-001",
    "name": "Updated Organization",
    "description": "Updated security operations center",
    "is_active": true,
    "updated_at": "2024-01-15T11:00:00Z",
    "settings": {
      "resource_limits": {
        "max_requests_per_day": 15000,
        "max_concurrent_scans": 8,
        "max_endpoints_per_scan": 150
      }
    }
  }
}
```

**Example:**
```bash
curl -X PUT http://localhost:8081/api/tenants/tenant-001 \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer your-token" \
  -d '{
    "name": "Updated Organization",
    "settings": {
      "resource_limits": {
        "max_requests_per_day": 15000,
        "max_concurrent_scans": 8
      }
    }
  }'
```

### Delete Tenant

Delete a tenant.

**Endpoint:** `DELETE /api/tenants/{tenant_id}`

**Response:**
```json
{
  "success": true,
  "message": "Tenant deleted successfully",
  "data_backup": "tenant-001-backup-2024-01-15.tar.gz"
}
```

**Example:**
```bash
curl -X DELETE http://localhost:8081/api/tenants/tenant-001 \
  -H "Authorization: Bearer your-token"
```

### Get Tenant Statistics

Get statistics for a specific tenant.

**Endpoint:** `GET /api/tenants/{tenant_id}/stats`

**Query Parameters:**
- `period` (optional): Time period (1d, 7d, 30d, 90d)
- `start_date` (optional): Start date (YYYY-MM-DD)
- `end_date` (optional): End date (YYYY-MM-DD)

**Response:**
```json
{
  "tenant_id": "tenant-001",
  "period": "30d",
  "stats": {
    "total_scans": 150,
    "successful_scans": 145,
    "failed_scans": 5,
    "total_vulnerabilities": 25,
    "critical_vulnerabilities": 3,
    "high_vulnerabilities": 8,
    "medium_vulnerabilities": 10,
    "low_vulnerabilities": 4,
    "average_score": 75.5,
    "scan_trends": [
      {
        "date": "2024-01-01",
        "scans": 5,
        "vulnerabilities": 2,
        "average_score": 80
      },
      {
        "date": "2024-01-02",
        "scans": 7,
        "vulnerabilities": 3,
        "average_score": 75
      }
    ],
    "resource_usage": {
      "requests_used": 7500,
      "requests_limit": 10000,
      "storage_used": 450,
      "storage_limit": 1000,
      "concurrent_scans": 2,
      "concurrent_limit": 5
    }
  }
}
```

**Example:**
```bash
curl -X GET "http://localhost:8081/api/tenants/tenant-001/stats?period=30d" \
  -H "Authorization: Bearer your-token"
```

## 📊 Scan Management API

### Start Scan

Start a new security scan.

**Endpoint:** `POST /api/scans`

**Request Body:**
```json
{
  "tenant_id": "tenant-001",
  "name": "Daily Security Scan",
  "description": "Regular security assessment",
  "endpoints": [
    {
      "url": "https://api.example.com/users",
      "method": "GET",
      "headers": {
        "Accept": "application/json"
      }
    },
    {
      "url": "https://api.example.com/data",
      "method": "POST",
      "body": "{\"query\": \"value\"}",
      "headers": {
        "Content-Type": "application/json"
      }
    }
  ],
  "config": {
    "injection_payloads": ["' OR '1'='1"],
    "xss_payloads": ["<script>alert('XSS')</script>"],
    "rate_limiting": {
      "requests_per_second": 10,
      "max_concurrent_requests": 5
    }
  },
  "options": {
    "output_format": "json",
    "include_details": true,
    "save_to_history": true,
    "send_alerts": true
  }
}
```

**Response:**
```json
{
  "success": true,
  "message": "Scan started successfully",
  "scan": {
    "id": "scan-001",
    "tenant_id": "tenant-001",
    "name": "Daily Security Scan",
    "status": "running",
    "started_at": "2024-01-15T10:30:00Z",
    "estimated_duration": 120,
    "endpoints_count": 2,
    "progress": 0,
    "webhook_url": "ws://localhost:8081/api/scans/scan-001/websocket"
  }
}
```

**Example:**
```bash
curl -X POST http://localhost:8081/api/scans \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer your-token" \
  -d '{
    "tenant_id": "tenant-001",
    "name": "Daily Security Scan",
    "endpoints": [
      {
        "url": "https://api.example.com/users",
        "method": "GET"
      }
    ]
  }'
```

### Get Scan Status

Get the status of a running or completed scan.

**Endpoint:** `GET /api/scans/{scan_id}`

**Response:**
```json
{
  "id": "scan-001",
  "tenant_id": "tenant-001",
  "name": "Daily Security Scan",
  "description": "Regular security assessment",
  "status": "completed",
  "started_at": "2024-01-15T10:30:00Z",
  "completed_at": "2024-01-15T10:45:00Z",
  "duration": 900,
  "endpoints_count": 2,
  "completed_endpoints": 2,
  "progress": 100,
  "results": {
    "total_vulnerabilities": 5,
    "critical_vulnerabilities": 1,
    "high_vulnerabilities": 2,
    "medium_vulnerabilities": 1,
    "low_vulnerabilities": 1,
    "average_score": 65,
    "endpoints": [
      {
        "url": "https://api.example.com/users",
        "score": 80,
        "vulnerabilities": 2,
        "status": "completed"
      },
      {
        "url": "https://api.example.com/data",
        "score": 50,
        "vulnerabilities": 3,
        "status": "completed"
      }
    ]
  },
  "summary": {
    "overall_status": "completed",
    "risk_level": "high",
    "recommendations": [
      "Fix SQL injection vulnerability in /data endpoint",
      "Implement proper authentication"
    ]
  }
}
```

**Example:**
```bash
curl -X GET http://localhost:8081/api/scans/scan-001 \
  -H "Authorization: Bearer your-token"
```

### List Scans

List all scans for a tenant or all tenants.

**Endpoint:** `GET /api/scans`

**Query Parameters:**
- `tenant_id` (optional): Filter by tenant ID
- `status` (optional): Filter by status (running, completed, failed)
- `page` (optional): Page number (default: 1)
- `limit` (optional): Items per page (default: 10)
- `start_date` (optional): Start date (YYYY-MM-DD)
- `end_date` (optional): End date (YYYY-MM-DD)

**Response:**
```json
{
  "scans": [
    {
      "id": "scan-001",
      "tenant_id": "tenant-001",
      "name": "Daily Security Scan",
      "status": "completed",
      "started_at": "2024-01-15T10:30:00Z",
      "completed_at": "2024-01-15T10:45:00Z",
      "duration": 900,
      "endpoints_count": 2,
      "total_vulnerabilities": 5,
      "average_score": 65
    },
    {
      "id": "scan-002",
      "tenant_id": "tenant-001",
      "name": "API Discovery Scan",
      "status": "running",
      "started_at": "2024-01-15T11:00:00Z",
      "endpoints_count": 5,
      "completed_endpoints": 2,
      "progress": 40
    }
  ],
  "pagination": {
    "page": 1,
    "limit": 10,
    "total": 2,
    "total_pages": 1
  }
}
```

**Example:**
```bash
curl -X GET "http://localhost:8081/api/scans?tenant_id=tenant-001&status=completed&page=1&limit=10" \
  -H "Authorization: Bearer your-token"
```

### Stop Scan

Stop a running scan.

**Endpoint:** `POST /api/scans/{scan_id}/stop`

**Response:**
```json
{
  "success": true,
  "message": "Scan stopped successfully",
  "scan": {
    "id": "scan-001",
    "status": "stopped",
    "stopped_at": "2024-01-15T10:35:00Z",
    "progress": 75,
    "results": {
      "completed_endpoints": 1,
      "total_endpoints": 2,
      "vulnerabilities_found": 3
    }
  }
}
```

**Example:**
```bash
curl -X POST http://localhost:8081/api/scans/scan-001/stop \
  -H "Authorization: Bearer your-token"
```

### Delete Scan

Delete a scan and its results.

**Endpoint:** `DELETE /api/scans/{scan_id}`

**Response:**
```json
{
  "success": true,
  "message": "Scan deleted successfully",
  "scan_id": "scan-001"
}
```

**Example:**
```bash
curl -X DELETE http://localhost:8081/api/scans/scan-001 \
  -H "Authorization: Bearer your-token"
```

### Get Scan Results

Get detailed results for a completed scan.

**Endpoint:** `GET /api/scans/{scan_id}/results`

**Query Parameters:**
- `format` (optional): Output format (json, html, pdf)
- `include_details` (optional): Include detailed test results (default: true)

**Response:**
```json
{
  "scan_id": "scan-001",
  "tenant_id": "tenant-001",
  "scan_name": "Daily Security Scan",
  "started_at": "2024-01-15T10:30:00Z",
  "completed_at": "2024-01-15T10:45:00Z",
  "duration": 900,
  "summary": {
    "total_endpoints": 2,
    "completed_endpoints": 2,
    "total_vulnerabilities": 5,
    "average_score": 65,
    "risk_level": "high"
  },
  "endpoints": [
    {
      "url": "https://api.example.com/users",
      "method": "GET",
      "score": 80,
      "results": [
        {
          "test_name": "Auth Test",
          "passed": true,
          "message": "Authentication successful"
        },
        {
          "test_name": "Injection Test",
          "passed": false,
          "message": "SQL injection detected with payload: '\'' OR '\''1'\''='\''1'\'''"
        }
      ],
      "vulnerabilities": [
        {
          "type": "SQL Injection",
          "severity": "high",
          "description": "SQL injection vulnerability detected",
          "payload": "' OR '1'='1",
          "remediation": "Use parameterized queries or prepared statements"
        }
      ]
    }
  ],
  "recommendations": [
    "Implement proper input validation",
    "Use parameterized queries",
    "Add authentication to sensitive endpoints"
  ]
}
```

**Example:**
```bash
curl -X GET "http://localhost:8081/api/scans/scan-001/results?format=json&include_details=true" \
  -H "Authorization: Bearer your-token"
```

## 📈 Metrics API

### Get Current Metrics

Get current system metrics.

**Endpoint:** `GET /api/metrics`

**Response:**
```json
{
  "timestamp": "2024-01-15T10:30:00Z",
  "system": {
    "cpu_usage": 25.5,
    "memory_usage": 512.5,
    "memory_total": 2048.0,
    "memory_percent": 25.0,
    "disk_usage": 85.2,
    "disk_total": 500.0,
    "disk_percent": 17.0,
    "network_in": 1024.5,
    "network_out": 2048.0,
    "goroutines": 25,
    "uptime": 86400
  },
  "scanner": {
    "active_scans": 2,
    "total_scans": 150,
    "successful_scans": 145,
    "failed_scans": 5,
    "requests_per_second": 8.5,
    "average_response_time": 250.5,
    "error_rate": 0.5
  },
  "security": {
    "vulnerabilities_found": 25,
    "critical_vulnerabilities": 3,
    "high_vulnerabilities": 8,
    "medium_vulnerabilities": 10,
    "low_vulnerabilities": 4,
    "vulnerability_trend": "increasing",
    "average_score": 75.5
  },
  "tenants": {
    "active_tenants": 3,
    "total_tenants": 5,
    "resource_usage": [
      {
        "tenant_id": "tenant-001",
        "requests_used": 7500,
        "requests_limit": 10000,
        "storage_used": 450,
        "storage_limit": 1000
      }
    ]
  }
}
```

**Example:**
```bash
curl -X GET http://localhost:8081/api/metrics \
  -H "Authorization: Bearer your-token"
```

### Get Historical Metrics

Get historical metrics data.

**Endpoint:** `GET /api/metrics/history`

**Query Parameters:**
- `period` (optional): Time period (1h, 6h, 24h, 7d, 30d)
- `start_time` (optional): Start time (Unix timestamp)
- `end_time` (optional): End time (Unix timestamp)
- `interval` (optional): Data interval (1m, 5m, 15m, 1h, 1d)

**Response:**
```json
{
  "period": "24h",
  "interval": "1h",
  "metrics": [
    {
      "timestamp": "2024-01-15T00:00:00Z",
      "system": {
        "cpu_usage": 20.5,
        "memory_usage": 480.0,
        "memory_percent": 23.4
      },
      "scanner": {
        "active_scans": 1,
        "requests_per_second": 5.2,
        "average_response_time": 200.0
      },
      "security": {
        "vulnerabilities_found": 0,
        "average_score": 80.0
      }
    },
    {
      "timestamp": "2024-01-15T01:00:00Z",
      "system": {
        "cpu_usage": 22.0,
        "memory_usage": 490.0,
        "memory_percent": 23.9
      },
      "scanner": {
        "active_scans": 2,
        "requests_per_second": 8.5,
        "average_response_time": 250.0
      },
      "security": {
        "vulnerabilities_found": 2,
        "average_score": 75.0
      }
    }
  ]
}
```

**Example:**
```bash
curl -X GET "http://localhost:8081/api/metrics/history?period=24h&interval=1h" \
  -H "Authorization: Bearer your-token"
```

### Get Tenant Metrics

Get metrics for a specific tenant.

**Endpoint:** `GET /api/tenants/{tenant_id}/metrics`

**Query Parameters:**
- `period` (optional): Time period (1h, 6h, 24h, 7d, 30d)

**Response:**
```json
{
  "tenant_id": "tenant-001",
  "period": "24h",
  "metrics": {
    "resource_usage": {
      "requests_used": 7500,
      "requests_limit": 10000,
      "requests_percent": 75.0,
      "storage_used": 450,
      "storage_limit": 1000,
      "storage_percent": 45.0,
      "concurrent_scans": 2,
      "concurrent_limit": 5
    },
    "scan_activity": {
      "total_scans": 150,
      "successful_scans": 145,
      "failed_scans": 5,
      "average_scan_duration": 900,
      "scans_today": 5,
      "scans_this_week": 35
    },
    "security_metrics": {
      "total_vulnerabilities": 25,
      "critical_vulnerabilities": 3,
      "high_vulnerabilities": 8,
      "medium_vulnerabilities": 10,
      "low_vulnerabilities": 4,
      "vulnerability_trend": "increasing",
      "average_score": 75.5,
      "risk_level": "medium"
    }
  }
}
```

**Example:**
```bash
curl -X GET "http://localhost:8081/api/tenants/tenant-001/metrics?period=24h" \
  -H "Authorization: Bearer your-token"
```

## 🔌 SIEM API

### Test SIEM Connection

Test connection to SIEM platform.

**Endpoint:** `POST /api/siem/test`

**Request Body:**
```json
{
  "type": "syslog",
  "config": {
    "host": "wazuh.company.com",
    "port": 514,
    "facility": "local0",
    "severity": "info"
  }
}
```

**Response:**
```json
{
  "success": true,
  "message": "SIEM connection test successful",
  "details": {
    "connection_time": "2024-01-15T10:30:00Z",
    "response_time": 25,
    "protocol": "syslog",
    "host": "wazuh.company.com",
    "port": 514
  }
}
```

**Example:**
```bash
curl -X POST http://localhost:8081/api/siem/test \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer your-token" \
  -d '{
    "type": "syslog",
    "config": {
      "host": "wazuh.company.com",
      "port": 514
    }
  }'
```

### Send Test Event

Send a test event to SIEM.

**Endpoint:** `POST /api/siem/test-event`

**Request Body:**
```json
{
  "event_type": "test",
  "message": "Test event from API Security Scanner",
  "severity": "info",
  "data": {
    "test_field": "test_value",
    "timestamp": "2024-01-15T10:30:00Z"
  }
}
```

**Response:**
```json
{
  "success": true,
  "message": "Test event sent successfully",
  "event_id": "test-event-001",
  "sent_at": "2024-01-15T10:30:00Z"
}
```

**Example:**
```bash
curl -X POST http://localhost:8081/api/siem/test-event \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer your-token" \
  -d '{
    "event_type": "test",
    "message": "Test event from API Security Scanner",
    "severity": "info"
  }'
```

### Get SIEM Status

Get SIEM integration status.

**Endpoint:** `GET /api/siem/status`

**Response:**
```json
{
  "enabled": true,
  "type": "syslog",
  "status": "connected",
  "last_event_sent": "2024-01-15T10:29:00Z",
  "events_sent_today": 25,
  "events_sent_total": 1250,
  "connection_info": {
    "host": "wazuh.company.com",
    "port": 514,
    "protocol": "syslog",
    "last_test": "2024-01-15T09:00:00Z",
    "test_status": "success"
  },
  "event_queue": {
    "queue_size": 0,
    "max_queue_size": 1000,
    "processing_rate": 10
  }
}
```

**Example:**
```bash
curl -X GET http://localhost:8081/api/siem/status \
  -H "Authorization: Bearer your-token"
```

## ❤️ Health Check API

### System Health

Get overall system health status.

**Endpoint:** `GET /health`

**Response:**
```json
{
  "status": "healthy",
  "timestamp": "2024-01-15T10:30:00Z",
  "version": "4.0.0",
  "uptime": 86400,
  "components": {
    "database": {
      "status": "healthy",
      "response_time": 5,
      "connection_pool": {
        "active": 2,
        "idle": 8,
        "total": 10
      }
    },
    "cache": {
      "status": "healthy",
      "response_time": 1,
      "memory_usage": 256
    },
    "siem": {
      "status": "healthy",
      "last_connection": "2024-01-15T10:29:00Z",
      "queue_size": 0
    },
    "storage": {
      "status": "healthy",
      "available_space": 414.8,
      "total_space": 500.0,
      "usage_percent": 17.0
    }
  },
  "metrics": {
    "cpu_usage": 25.5,
    "memory_usage": 512.5,
    "memory_percent": 25.0,
    "active_connections": 15,
    "requests_per_minute": 120
  }
}
```

**Example:**
```bash
curl -X GET http://localhost:8081/health
```

### Component Health

Get health status for specific components.

**Endpoint:** `GET /health/{component}`

**Available Components:**
- `database`
- `cache`
- `siem`
- `storage`
- `network`

**Response:**
```json
{
  "component": "database",
  "status": "healthy",
  "timestamp": "2024-01-15T10:30:00Z",
  "details": {
    "connection_string": "postgresql://scanner:***@localhost:5432/api_scanner",
    "response_time": 5,
    "connection_pool": {
      "active": 2,
      "idle": 8,
      "total": 10,
      "max": 20
    },
    "last_query": "SELECT COUNT(*) FROM scans",
    "query_time": 2
  }
}
```

**Example:**
```bash
curl -X GET http://localhost:8081/health/database \
  -H "Authorization: Bearer your-token"
```

### Health Check History

Get health check history.

**Endpoint:** `GET /health/history`

**Query Parameters:**
- `period` (optional): Time period (1h, 6h, 24h, 7d)

**Response:**
```json
{
  "period": "24h",
  "checks": [
    {
      "timestamp": "2024-01-15T10:30:00Z",
      "overall_status": "healthy",
      "components": {
        "database": "healthy",
        "cache": "healthy",
        "siem": "healthy",
        "storage": "healthy"
      },
      "metrics": {
        "cpu_usage": 25.5,
        "memory_usage": 512.5,
        "memory_percent": 25.0
      }
    },
    {
      "timestamp": "2024-01-15T10:25:00Z",
      "overall_status": "degraded",
      "components": {
        "database": "healthy",
        "cache": "degraded",
        "siem": "healthy",
        "storage": "healthy"
      },
      "metrics": {
        "cpu_usage": 45.0,
        "memory_usage": 800.0,
        "memory_percent": 39.0
      }
    }
  ]
}
```

**Example:**
```bash
curl -X GET "http://localhost:8081/health/history?period=24h" \
  -H "Authorization: Bearer your-token"
```

## 🌐 WebSocket API

### Scan Progress WebSocket

Connect to real-time scan progress updates.

**Endpoint:** `ws://localhost:8081/api/scans/{scan_id}/websocket`

**Message Format:**
```json
{
  "type": "progress",
  "timestamp": "2024-01-15T10:30:00Z",
  "scan_id": "scan-001",
  "data": {
    "progress": 50,
    "current_endpoint": "https://api.example.com/data",
    "completed_endpoints": 1,
    "total_endpoints": 2,
    "vulnerabilities_found": 3,
    "estimated_time_remaining": 300
  }
}
```

**Example using JavaScript:**
```javascript
const ws = new WebSocket('ws://localhost:8081/api/scans/scan-001/websocket');

ws.onmessage = function(event) {
  const data = JSON.parse(event.data);
  console.log('Received message:', data);

  if (data.type === 'progress') {
    console.log(`Scan progress: ${data.data.progress}%`);
  } else if (data.type === 'completed') {
    console.log('Scan completed!');
    ws.close();
  } else if (data.type === 'error') {
    console.error('Scan error:', data.message);
  }
};

ws.onopen = function() {
  console.log('WebSocket connected');
};

ws.onclose = function() {
  console.log('WebSocket disconnected');
};

ws.onerror = function(error) {
  console.error('WebSocket error:', error);
};
```

### Metrics WebSocket

Connect to real-time metrics updates.

**Endpoint:** `ws://localhost:8081/api/metrics/websocket`

**Message Format:**
```json
{
  "type": "metrics",
  "timestamp": "2024-01-15T10:30:00Z",
  "data": {
    "system": {
      "cpu_usage": 25.5,
      "memory_usage": 512.5,
      "memory_percent": 25.0
    },
    "scanner": {
      "active_scans": 2,
      "requests_per_second": 8.5
    },
    "security": {
      "vulnerabilities_found": 25,
      "average_score": 75.5
    }
  }
}
```

**Example using JavaScript:**
```javascript
const ws = new WebSocket('ws://localhost:8081/api/metrics/websocket');

ws.onmessage = function(event) {
  const data = JSON.parse(event.data);
  console.log('Received metrics:', data);

  if (data.type === 'metrics') {
    console.log(`CPU Usage: ${data.data.system.cpu_usage}%`);
    console.log(`Memory Usage: ${data.data.system.memory_percent}%`);
    console.log(`Active Scans: ${data.data.scanner.active_scans}`);
  }
};

ws.onopen = function() {
  console.log('Metrics WebSocket connected');
};
```

## ❌ Error Responses

### Error Format

All error responses follow this format:

```json
{
  "error": {
    "code": "ERROR_CODE",
    "message": "Human readable error message",
    "details": "Additional error details",
    "timestamp": "2024-01-15T10:30:00Z",
    "request_id": "req-001"
  }
}
```

### Common Error Codes

| Code | HTTP Status | Description |
|------|-------------|-------------|
| `UNAUTHORIZED` | 401 | Authentication failed |
| `FORBIDDEN` | 403 | Access denied |
| `NOT_FOUND` | 404 | Resource not found |
| `VALIDATION_ERROR` | 400 | Request validation failed |
| `CONFIGURATION_ERROR` | 400 | Configuration error |
| `RATE_LIMITED` | 429 | Rate limit exceeded |
| `INTERNAL_ERROR` | 500 | Internal server error |

### Error Examples

**Unauthorized:**
```json
{
  "error": {
    "code": "UNAUTHORIZED",
    "message": "Invalid authentication credentials",
    "timestamp": "2024-01-15T10:30:00Z",
    "request_id": "req-001"
  }
}
```

**Validation Error:**
```json
{
  "error": {
    "code": "VALIDATION_ERROR",
    "message": "Request validation failed",
    "details": {
      "field": "tenant_id",
      "error": "required field is missing"
    },
    "timestamp": "2024-01-15T10:30:00Z",
    "request_id": "req-002"
  }
}
```

**Rate Limited:**
```json
{
  "error": {
    "code": "RATE_LIMITED",
    "message": "Rate limit exceeded",
    "details": {
      "limit": 100,
      "window": "1h",
      "remaining": 0,
      "reset_time": "2024-01-15T11:30:00Z"
    },
    "timestamp": "2024-01-15T10:30:00Z",
    "request_id": "req-003"
  }
}
```

## 🚦 Rate Limiting

### Rate Limit Headers

The API includes rate limiting information in response headers:

```
X-RateLimit-Limit: 1000
X-RateLimit-Remaining: 999
X-RateLimit-Reset: 1642249200
X-RateLimit-Window: 1h
```

### Rate Limits by Endpoint

| Endpoint | Limit | Window |
|----------|-------|--------|
| `POST /api/scans` | 10 per tenant | 1 hour |
| `GET /api/metrics` | 100 | 1 hour |
| `GET /api/tenants` | 50 | 1 hour |
| `POST /api/config` | 20 | 1 hour |
| All other endpoints | 1000 | 1 hour |

### Tenant-Specific Limits

Rate limits can be configured per tenant:

```json
{
  "tenant_id": "tenant-001",
  "rate_limits": {
    "scans_per_hour": 20,
    "api_requests_per_hour": 1000,
    "concurrent_requests": 5
  }
}
```

## 📝 Examples

### Complete Scan Example

```bash
# 1. Start a scan
SCAN_RESPONSE=$(curl -s -X POST http://localhost:8081/api/scans \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer your-token" \
  -d '{
    "tenant_id": "tenant-001",
    "name": "Security Assessment",
    "endpoints": [
      {
        "url": "https://api.example.com/users",
        "method": "GET"
      },
      {
        "url": "https://api.example.com/data",
        "method": "POST",
        "body": "{\"query\": \"value\"}"
      }
    ]
  }')

# Extract scan ID
SCAN_ID=$(echo $SCAN_RESPONSE | jq -r '.scan.id')

echo "Started scan: $SCAN_ID"

# 2. Monitor scan progress
while true; do
  STATUS=$(curl -s -X GET http://localhost:8081/api/scans/$SCAN_ID \
    -H "Authorization: Bearer your-token" |
    jq -r '.status')

  echo "Scan status: $STATUS"

  if [ "$STATUS" = "completed" ] || [ "$STATUS" = "failed" ]; then
    break
  fi

  sleep 10
done

# 3. Get scan results
curl -s -X GET "http://localhost:8081/api/scans/$SCAN_ID/results?format=json" \
  -H "Authorization: Bearer your-token" | jq '.'
```

### Tenant Management Example

```bash
# 1. Create a new tenant
curl -X POST http://localhost:8081/api/tenants \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer your-token" \
  -d '{
    "id": "new-customer",
    "name": "New Customer Corp",
    "description": "New customer security team",
    "is_active": true,
    "settings": {
      "resource_limits": {
        "max_requests_per_day": 5000,
        "max_concurrent_scans": 3,
        "max_endpoints_per_scan": 50
      }
    }
  }'

# 2. Update tenant configuration
curl -X PUT http://localhost:8081/api/tenants/new-customer \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer your-token" \
  -d '{
    "settings": {
      "resource_limits": {
        "max_requests_per_day": 10000,
        "max_concurrent_scans": 5
      },
      "notification_settings": {
        "email_notifications": true,
        "email_recipients": ["security@newcustomer.com"]
      }
    }
  }'

# 3. Get tenant statistics
curl -X GET "http://localhost:8081/api/tenants/new-customer/stats?period=7d" \
  -H "Authorization: Bearer your-token" | jq '.'
```

### SIEM Integration Example

```bash
# 1. Test SIEM connection
curl -X POST http://localhost:8081/api/siem/test \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer your-token" \
  -d '{
    "type": "syslog",
    "config": {
      "host": "wazuh.company.com",
      "port": 514,
      "facility": "local0",
      "severity": "info"
    }
  }'

# 2. Send test event
curl -X POST http://localhost:8081/api/siem/test-event \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer your-token" \
  -d '{
    "event_type": "vulnerability_detected",
    "message": "SQL injection vulnerability detected",
    "severity": "high",
    "data": {
      "target_url": "https://api.example.com/data",
      "payload": "1 OR 1=1",
      "vulnerability_type": "SQL Injection"
    }
  }'

# 3. Get SIEM status
curl -X GET http://localhost:8081/api/siem/status \
  -H "Authorization: Bearer your-token" | jq '.'
```

### WebSocket Example (Python)

```python
import asyncio
import websockets
import json

async def listen_to_scan_progress(scan_id):
    uri = f"ws://localhost:8081/api/scans/{scan_id}/websocket"

    async with websockets.connect(uri) as websocket:
        print(f"Connected to scan {scan_id}")

        async for message in websocket:
            data = json.loads(message)

            if data['type'] == 'progress':
                progress = data['data']['progress']
                print(f"Scan progress: {progress}%")

            elif data['type'] == 'completed':
                print("Scan completed!")
                print(f"Results: {data['data']}")
                break

            elif data['type'] == 'error':
                print(f"Error: {data['message']}")
                break

# Usage
asyncio.get_event_loop().run_until_complete(
    listen_to_scan_progress("scan-001")
)
```

### Configuration Management Example

```bash
# 1. Get current configuration
curl -X GET http://localhost:8081/api/config \
  -H "Authorization: Bearer your-token" > current_config.json

# 2. Modify configuration
jq '.scanner.rate_limiting.requests_per_second = 15' current_config.json > new_config.json

# 3. Update configuration
curl -X POST http://localhost:8081/api/config \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer your-token" \
  -d @new_config.json

# 4. Validate configuration
curl -X POST http://localhost:8081/api/config/validate \
  -H "Authorization: Bearer your-token" | jq '.'
```

---

This comprehensive API documentation covers all endpoints, authentication methods, and usage examples for the API Security Scanner. For additional help or to report issues, please visit our [GitHub repository](https://github.com/elliotsecops/API-Security-Scanner).