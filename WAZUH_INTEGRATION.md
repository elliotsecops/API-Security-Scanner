# Wazuh SIEM Integration Guide

This guide explains how to integrate the API Security Scanner with Wazuh SIEM for centralized security monitoring and alerting.

## Overview

The API Security Scanner can send security events to Wazuh via syslog integration. This allows you to:
- Centralize security findings in Wazuh
- Create custom rules and alerts
- Correlate API security events with other security data
- Generate automated responses to vulnerabilities

## Configuration

### 1. API Security Scanner Configuration

Use the `config-wazuh.yaml` file or modify your existing configuration:

```yaml
siem:
  enabled: true
  type: "syslog"  # Use syslog for Wazuh
  format: "json"   # JSON format for structured data
  config:
    host: "localhost"    # Wazuh manager IP
    port: 514           # Wazuh syslog port
    facility: "local0"  # Syslog facility
    severity: "info"    # Default severity
```

### 2. Wazuh Manager Configuration

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

  <rule id="100104" level="7">
    <if_sid>100100</if_sid>
    <field name="vulnerability">NoSQL injection</field>
    <description>API Security Scanner - NoSQL Injection Detected</description>
    <group>nosql_injection,attack</group>
  </rule>
</group>
```

### 3. Wazuh Agent Configuration (Optional)

If running the scanner on a different host, configure the Wazuh agent to forward logs:

```xml
<!-- Localfile configuration on agent -->
<localfile>
  <log_format>syslog</log_format>
  <location>/var/log/api-security-scanner.log</location>
</localfile>
```

## Event Format

The scanner sends structured JSON events via syslog:

```json
{
  "timestamp": "2025-09-22T01:44:43Z",
  "event_type": "vulnerability_detected",
  "severity": "high",
  "tenant_id": "security-team",
  "source_ip": "192.168.1.100",
  "target_url": "https://api.example.com/users",
  "method": "GET",
  "vulnerability": "SQL injection",
  "description": "Potential SQL injection detected with payload: ' OR '1'='1",
  "raw_data": {
    "payload": "' OR '1'='1",
    "response_status": 200,
    "test_type": "injection"
  },
  "tags": ["sql_injection", "api_security", "attack"]
}
```

## Testing the Integration

1. **Start Wazuh Manager**:
   ```bash
   systemctl start wazuh-manager
   ```

2. **Run API Security Scanner**:
   ```bash
   ./api-security-scanner --scan --config config-wazuh.yaml
   ```

3. **Check Wazuh Dashboard**:
   - Navigate to Security Events
   - Filter by `api_security` group
   - Look for events from `api-security-scanner`

## Advanced Configuration

### Custom Decoders

Add custom decoders in `/var/ossec/etc/decoders/local_decoder.xml`:

```xml
<decoder name="api-security-scanner-json">
  <parent>api-security-scanner</parent>
  <type>json</type>
  <field name="vulnerability">vulnerability</field>
  <field name="severity">severity</field>
  <field name="target_url">target_url</field>
  <field name="event_type">event_type</field>
</decoder>
```

### Active Response Rules

Create automated responses for critical vulnerabilities:

```xml
<rule id="100200" level="12">
  <if_sid>100103</if_sid>
  <field name="severity">critical</field>
  <description>API Security Scanner - Critical Auth Bypass - Block IP</description>
  <group>auth_bypass,critical,active_response</group>
  <action>firewall-drop</action>
</rule>
```

### Integration with Elasticsearch

For enhanced analytics, forward Wazuh events to Elasticsearch:

1. **Install Wazuh-Elasticstack integration**
2. **Configure Filebeat to read Wazuh alerts**
3. **Create Kibana dashboards for API security metrics**

## Monitoring and Alerting

### Kibana Dashboard Example

Create visualizations for:
- Vulnerability trends over time
- Most vulnerable endpoints
- Attack types distribution
- Tenant-specific security metrics

### Wazuh Alerts

Configure email alerts for critical findings:

```xml
<alert>
  <command>email-alert</command>
  <location>security-team@company.com</location>
  <level>8</level>
  <group>api_security</group>
</alert>
```

## Troubleshooting

### Common Issues

1. **Syslog Connection Failed**:
   - Check Wazuh manager firewall rules
   - Verify port 514 is open
   - Check network connectivity

2. **Events Not Appearing in Wazuh**:
   - Verify decoder configuration
   - Check Wazuh manager logs
   - Test syslog connectivity manually

3. **JSON Parsing Errors**:
   - Verify JSON format in events
   - Check decoder syntax
   - Test with simple events first

### Testing Syslog Connectivity

```bash
# Test syslog to Wazuh
echo "test message" | nc -u localhost 514

# Check Wazuh logs
tail -f /var/ossec/logs/archives/archives.log
```

## Security Considerations

1. **Network Security**:
   - Use encrypted syslog (TLS) for production
   - Restrict IP access to Wazuh manager
   - Monitor syslog traffic for anomalies

2. **Authentication**:
   - Configure Wazuh API authentication
   - Use certificate-based authentication where possible
   - Regularly rotate credentials

3. **Data Privacy**:
   - Anonymize sensitive data in events
   - Configure data retention policies
   - Comply with privacy regulations

## Performance Tuning

### High-Volume Environments

1. **Scale Wazuh Cluster**:
   - Add multiple Wazuh managers
   - Use load balancers for syslog traffic
   - Implement horizontal scaling

2. **Optimize Scanner Settings**:
   - Adjust scan frequency
   - Use targeted scanning
   - Implement rate limiting

3. **Wazuh Performance**:
   - Tune database parameters
   - Optimize index settings
   - Monitor resource usage

## Integration Benefits

### Advantages of Wazuh Integration

1. **Centralized Monitoring**: Single pane of glass for all security events
2. **Correlation**: Link API security events with other security data
3. **Automated Response**: Trigger automated actions for critical findings
4. **Compliance**: Meet regulatory requirements with audit trails
5. **Scalability**: Handle large volumes of security data
6. **Cost-Effective**: Leverage open-source security tools

### Use Cases

1. **Real-time Threat Detection**: Immediate alerts for active attacks
2. **Compliance Reporting**: Generate reports for auditors
3. **Incident Response**: Automated containment of security incidents
4. **Security Analytics**: Trend analysis and vulnerability tracking
5. **Multi-tenant Security**: Isolate events by customer/department

This integration provides a powerful, cost-effective solution for enterprise API security monitoring using Wazuh SIEM.