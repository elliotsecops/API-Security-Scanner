# Grafana Integration for API Security Scanner

This document explains how to integrate the API Security Scanner with Grafana for monitoring and visualization of security metrics.

## Overview

The API Security Scanner includes built-in support for Prometheus metrics export, which can be easily integrated with Grafana for comprehensive visualization and monitoring. This approach provides:

- Professional-grade dashboards with advanced visualization options
- Alerting capabilities
- Historical data analysis
- Multi-user support and permissions
- Integration with existing monitoring infrastructure

## Architecture

The integration uses a standard Prometheus -> Grafana architecture:

```
API Security Scanner
        |
        | (exposes /metrics endpoint in Prometheus format)
        |
    Prometheus
        |
        | (scrapes metrics)
        |
    Grafana
        |
        | (visualizes metrics)
        |
    Users/Operators
```

## Prerequisites

- API Security Scanner v4.0+
- Prometheus server (optional, if not using the docker-compose setup)
- Grafana (optional, if not using the docker-compose setup)

## Quick Start with Docker Compose

The easiest way to get started is using the provided docker-compose configuration, which ships with a background scan workload and Grafana auto-provisioning already wired up.

```bash
docker-compose -f grafana-docker-compose.yml up -d
```

This will start:
- API Security Scanner (dashboard/API) on ports 8081 and 8090
- API Security Scanner scan worker that runs recurring scans to populate metrics
- Prometheus on port 9090
- Grafana on port 3000 (default credentials: admin/admin) with the dashboard auto-provisioned

> **Heads up:** If you customise the stack or run the scanner outside docker-compose, make sure at least one scan job is active; the metrics endpoint only reflects actual scan history. The bundled scan worker sleeps five minutes between runs—tune that cadence to match your environment.
>
> Prometheus also runs inside Docker, so if you modify the compose topology ensure the scrape target still points to a reachable hostname (the default `api-security-scanner:8090` works for the provided stack).

## Manual Configuration

### 1. Configure the Scanner

Ensure your `config.yaml` has metrics enabled:

```yaml
metrics:
  enabled: true
  port: 8090  # This should match the target in Prometheus config
  update_interval: 30s
  retention_days: 30
```

### 2. Set up Prometheus

Configure Prometheus to scrape metrics from the scanner. Create or update `prometheus.yml`:

```yaml
scrape_configs:
  - job_name: 'api-security-scanner'
    static_configs:
      - targets: ['localhost:8090']  # Works for bare-metal or kind setups
    metrics_path: '/metrics'
    scrape_interval: 30s

  # When Prometheus runs in Docker alongside the scanner use the service name:
  # - job_name: 'api-security-scanner-docker'
  #   static_configs:
  #     - targets: ['api-security-scanner:8090']
  #   metrics_path: '/metrics'
  #   scrape_interval: 30s
```

### 3. Configure Grafana

#### Add Prometheus as a Data Source

1. Open Grafana in your browser (default: http://localhost:3000)
2. Login with your credentials (default: admin/admin)
3. Go to "Connections" → "Data sources"
4. Click "Add new data source"
5. Select "Prometheus"
6. Enter the URL: `http://prometheus:9090` (if using docker-compose) or `http://localhost:9090` (for local setup)
7. Click "Save & Test"

#### Provision the Dashboard

The docker-compose setup mounts `grafana-dashboards.yaml` (provider) and `grafana-dashboard.json` (raw dashboard) into Grafana, so the dashboard appears automatically on first boot.

For manual deployments you have two options:

1. **Auto-provision** – Copy `grafana-dashboards.yaml` and `grafana-dashboard.json` into your Grafana provisioning tree (for example `/etc/grafana/provisioning/dashboards/`), update the paths as needed, and restart Grafana.
2. **Manual import** – In Grafana go to "Create" → "Import", upload `grafana-dashboard.json`, pick the Prometheus data source, and click "Import".

### 4. Dashboard Overview

The provided dashboard includes:

- **Vulnerability Summary**: Total vulnerabilities by count and severity
- **System Resources**: CPU, memory, and goroutine usage
- **Performance Metrics**: Response time, throughput, and error rates
- **Tenant-Specific Metrics**: Vulnerability tracking per tenant
- **Historical Trends**: Time-based visualization of metrics

## Available Metrics

The scanner exposes the following metrics:

| Metric | Type | Description |
|--------|------|-------------|
| `api_scanner_total_scans` | Counter | Total number of scans completed |
| `api_scanner_total_endpoints` | Gauge | Total number of endpoints tested |
| `api_scanner_total_vulnerabilities` | Gauge | Total number of vulnerabilities found |
| `api_scanner_critical_vulnerabilities` | Gauge | Number of critical vulnerabilities |
| `api_scanner_high_vulnerabilities` | Gauge | Number of high vulnerabilities |
| `api_scanner_medium_vulnerabilities` | Gauge | Number of medium vulnerabilities |
| `api_scanner_low_vulnerabilities` | Gauge | Number of low vulnerabilities |
| `api_scanner_active_tenants` | Gauge | Number of active tenants |
| `api_scanner_avg_response_time` | Gauge | Average response time in milliseconds |
| `api_scanner_throughput` | Gauge | Requests per second |
| `api_scanner_error_rate` | Gauge | Percentage of errors |
| `api_scanner_cpu_usage` | Gauge | CPU usage percentage |
| `api_scanner_memory_usage` | Gauge | Memory usage in MB |
| `api_scanner_goroutines` | Gauge | Number of goroutines |
| `api_scanner_tenant_total_scans` | Gauge | Total scans for a specific tenant |
| `api_scanner_tenant_critical_vulnerabilities` | Gauge | Critical vulnerabilities for a specific tenant |
| `api_scanner_tenant_high_vulnerabilities` | Gauge | High vulnerabilities for a specific tenant |

### Known Gaps and Limitations

- Metrics require actual scans to populate meaningful data. The docker-compose stack includes a scheduled scan worker; if you disable it, ensure another process runs scans on a cadence that matches your monitoring needs.

## Custom Dashboards

You can create custom dashboards using the available metrics. Common use cases include:

- Creating tenant-specific dashboards
- Setting up alert rules based on vulnerability thresholds
- Creating executive dashboards with high-level security metrics
- Building compliance reporting dashboards

## Alerting

Grafana's alerting can be configured to trigger notifications when:

- Critical vulnerabilities exceed a threshold
- System resources reach dangerous levels
- Scan completion rates drop below expectations
- New vulnerability types are detected

To set up alerting:

1. Navigate to an existing panel or create a new one
2. Click on the panel title and select "Edit"
3. Go to the "Alert" tab
4. Define your alert conditions
5. Configure notification channels

## Security Considerations

- Protect the metrics endpoint (`/metrics`) with appropriate authentication if exposing externally
- Use secure connections between services
- Implement proper Grafana user management and permissions
- Regularly rotate Grafana admin passwords

## Troubleshooting

### Metrics endpoint not available
- Verify the scanner is running and the metrics service is enabled
- Check the correct port is exposed and accessible
- Ensure no firewall rules block the metrics endpoint

### Grafana not showing data
- Verify Prometheus can scrape the metrics endpoint
- Check Prometheus logs for scrape errors
- Confirm the data source configuration in Grafana

### Dashboard panels showing "No data"
- Ensure metrics have been collected (run a scan if necessary)
- Check that the time range in Grafana is appropriate
- Verify the metric names match what's exported by the scanner

### Run a scan workload
- Execute the scanner in `-scan` mode alongside the dashboard to populate metrics, for example:

  ```bash
  docker-compose -f grafana-docker-compose.yml exec api-security-scanner ./api-security-scanner -scan
  ```

- Alternatively add a dedicated service/container that runs scheduled scans and shares the metrics/report volumes with the dashboard container.
- In Prometheus, open `Status → Targets` and ensure the scrape target shows `UP` after scans run; otherwise adjust the hostname/port to reach the scanner.

## Advanced Configuration

### Custom Metrics Labels

Metrics include tenant information where applicable using labels like `{tenant_id="tenant-001"}`. You can use these labels in dashboard queries to filter or group data by tenant.

### Performance Tuning

- Adjust scrape intervals based on your monitoring needs vs. performance impact
- Configure Prometheus retention policies appropriately
- Monitor the scanner's performance during metric collection

## Next Steps

- Implement custom alert rules based on your security policies
- Create executive dashboards with key security metrics
- Integrate with existing incident management tools
- Set up automated report generation
