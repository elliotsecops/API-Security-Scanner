# API Security Scanner - Installation and Setup Guide

## 📋 Table of Contents

1. [System Requirements](#system-requirements)
2. [Installation Methods](#installation-methods)
3. [Basic Setup](#basic-setup)
4. [Enterprise Setup](#enterprise-setup)
5. [Configuration](#configuration)
6. [Verification](#verification)
7. [Troubleshooting](#troubleshooting)
8. [Upgrade Guide](#upgrade-guide)
9. [Uninstallation](#uninstallation)

## 🖥️ System Requirements

### Minimum Requirements

- **OS**: Linux, macOS, or Windows
- **Go**: 1.24 or later
- **RAM**: 512MB minimum, 2GB recommended
- **CPU**: 1 core minimum, 2+ cores recommended
- **Disk**: 100MB minimum, 1GB+ recommended for enterprise use
- **Network**: Internet access for API scanning and SIEM integration

### Recommended Requirements (Enterprise)

- **OS**: Ubuntu 20.04+, CentOS 8+, or RHEL 8+
- **Go**: 1.24+ latest version
- **RAM**: 4GB minimum, 8GB+ recommended
- **CPU**: 4 cores minimum, 8+ cores recommended
- **Disk**: 10GB+ SSD storage
- **Network**: High-speed internet connection
- **SIEM**: Access to SIEM platform (Wazuh, Splunk, ELK, etc.)
- **Database**: PostgreSQL 12+ or MySQL 8+ (for data persistence)

## 📦 Installation Methods

### Method 1: From Source (Recommended for Development)

#### Prerequisites

```bash
# Install Go 1.24+
# Ubuntu/Debian
sudo apt update
sudo apt install -y golang-go

# CentOS/RHEL
sudo yum install -y golang

# macOS (using Homebrew)
brew install go

# Verify Go installation
go version
```

#### Installation Steps

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

# (Optional) Install to system path
sudo cp api-security-scanner /usr/local/bin/
sudo chmod +x /usr/local/bin/api-security-scanner

# Verify installation
api-security-scanner -version
```

### Method 2: Binary Release

#### Download Binary

```bash
# Download the latest release
wget https://github.com/elliotsecops/API-Security-Scanner/releases/latest/download/api-security-scanner-linux-amd64

# Make executable
chmod +x api-security-scanner-linux-amd64

# Move to system path
sudo mv api-security-scanner-linux-amd64 /usr/local/bin/api-security-scanner

# Verify installation
api-security-scanner -version
```

#### Available Platforms

- `api-security-scanner-linux-amd64` - Linux 64-bit
- `api-security-scanner-linux-arm64` - Linux ARM 64-bit
- `api-security-scanner-darwin-amd64` - macOS Intel 64-bit
- `api-security-scanner-darwin-arm64` - macOS Apple Silicon
- `api-security-scanner-windows-amd64.exe` - Windows 64-bit

### Method 3: Docker Installation

#### Dockerfile

```dockerfile
FROM golang:1.24-alpine AS builder

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN go build -o api-security-scanner .

FROM alpine:latest
RUN apk --no-cache add ca-certificates
WORKDIR /root/

COPY --from=builder /app/api-security-scanner .
COPY --from=builder /app/config.yaml ./config.yaml
COPY --from=builder /app/config-test.yaml ./config-test.yaml

EXPOSE 8080 8081

CMD ["./api-security-scanner"]
```

#### Build and Run

```bash
# Build Docker image
docker build -t api-security-scanner .

# Run container with test configuration
docker run -d --name api-security-scanner -p 8080-8081:8080-8081 \
  -v $(pwd)/config-test.yaml:/app/config-test.yaml \
  -v $(pwd)/reports:/app/reports \
  api-security-scanner ./api-security-scanner -config config-test.yaml -dashboard
```

#### Docker Compose for Integration Testing

The repository includes a complete integration test environment with OWASP Juice Shop:

```yaml
version: '3.8'

services:
  # Vulnerable test API - OWASP Juice Shop
  juice-shop:
    image: bkimminich/juice-shop:latest
    container_name: juice-shop
    ports:
      - "3000:3000"
    environment:
      - NODE_ENV=production
    healthcheck:
      test: ["CMD", "wget", "--spider", "http://localhost:3000/rest/admin/application-version"]
      interval: 30s
      timeout: 10s
      retries: 3
      start_period: 90s

  # API Security Scanner
  api-security-scanner:
    build: .
    container_name: api-security-scanner
    depends_on:
      juice-shop:
        condition: service_healthy
    ports:
      - "8080:8080"
      - "8081:8081"
    volumes:
      - ./config-test.yaml:/app/config.yaml
      - ./reports:/app/reports
    environment:
      - SERVER_PORT=8081
      - METRICS_PORT=8080
    command: ["./api-security-scanner", "-config", "config.yaml", "-dashboard"]
```

To run the complete test environment:

```bash
# Start both services (Juice Shop and API Security Scanner)
docker-compose up -d

# Verify both containers are running
docker ps

# Access the dashboard at http://localhost:8080
# Run a test scan:
docker exec api-security-scanner ./api-security-scanner -config config-test.yaml -scan

# View scan results in the history directory
docker exec api-security-scanner ls -la history/
```

### Method 4: Package Manager (Linux)

#### APT (Debian/Ubuntu)

```bash
# Add repository
echo "deb [trusted=yes] https://apt.api-security-scanner.com ./" | sudo tee /etc/apt/sources.list.d/api-security-scanner.list

# Update package list
sudo apt update

# Install
sudo apt install api-security-scanner

# Verify
api-security-scanner -version
```

#### YUM/DNF (CentOS/RHEL)

```bash
# Add repository
sudo rpm -Uvh https://yum.api-security-scanner.com/api-security-scanner-latest.noarch.rpm

# Install
sudo yum install api-security-scanner

# Verify
api-security-scanner -version
```

## ⚙️ Basic Setup

### Quick Start

```bash
# 1. Create a basic configuration file
cat > config.yaml << EOF
scanner:
  api_endpoints:
    - url: "https://httpbin.org/get"
      method: "GET"
    - url: "https://httpbin.org/post"
      method: "POST"
      body: '{"test": "data"}'

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
EOF

# 2. Run a test scan
./api-security-scanner -config config.yaml -scan

# 3. Start the dashboard
./api-security-scanner -config config.yaml -dashboard
```

### Directory Structure

```
API-Security-Scanner/
├── api-security-scanner          # Main executable
├── config.yaml                   # Configuration file
├── config-enterprise.yaml        # Enterprise configuration example
├── config-wazuh.yaml             # Wazuh SIEM configuration example
├── data/                         # Data directory
│   ├── tenants/                  # Multi-tenant data
│   ├── history/                  # Historical scan data
│   └── metrics/                  # Metrics data
├── logs/                         # Log files
│   ├── app.log                   # Application logs
│   ├── scanner.log               # Scanner logs
│   └── siem.log                  # SIEM integration logs
├── static/                       # Static files for dashboard
│   ├── index.html                # Dashboard UI
│   ├── css/                      # CSS files
│   └── js/                       # JavaScript files
├── docs/                         # Documentation
├── examples/                     # Example configurations
├── scripts/                      # Helper scripts
└── tests/                        # Test files
```

### Basic Configuration

Create a minimal configuration file:

```yaml
# config.yaml
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

### Service Setup (Linux)

#### Systemd Service

```bash
# Create systemd service file
sudo tee /etc/systemd/system/api-security-scanner.service > /dev/null <<EOF
[Unit]
Description=API Security Scanner
After=network.target

[Service]
Type=simple
User=scanner
Group=scanner
WorkingDirectory=/opt/api-security-scanner
ExecStart=/opt/api-security-scanner/api-security-scanner -config /opt/api-security-scanner/config.yaml
Restart=always
RestartSec=10
Environment=PATH=/usr/local/bin:/usr/bin:/bin
Environment=HOME=/opt/api-security-scanner

[Install]
WantedBy=multi-user.target
EOF

# Create user and directories
sudo useradd -r -s /bin/false scanner
sudo mkdir -p /opt/api-security-scanner
sudo mkdir -p /var/log/api-security-scanner
sudo mkdir -p /var/lib/api-security-scanner
sudo chown -R scanner:scanner /opt/api-security-scanner
sudo chown -R scanner:scanner /var/log/api-security-scanner
sudo chown -R scanner:scanner /var/lib/api-security-scanner

# Copy files
sudo cp api-security-scanner /opt/api-security-scanner/
sudo cp config.yaml /opt/api-security-scanner/
sudo chmod +x /opt/api-security-scanner/api-security-scanner

# Enable and start service
sudo systemctl daemon-reload
sudo systemctl enable api-security-scanner
sudo systemctl start api-security-scanner

# Check status
sudo systemctl status api-security-scanner
```

#### Logrotate Configuration

```bash
# Create logrotate configuration
sudo tee /etc/logrotate.d/api-security-scanner > /dev/null <<EOF
/var/log/api-security-scanner/*.log {
    daily
    missingok
    rotate 30
    compress
    delaycompress
    notifempty
    create 0644 scanner scanner
    postrotate
        systemctl reload api-security-scanner
    endscript
}
EOF
```

## 🏢 Enterprise Setup

### Enterprise Installation

```bash
# 1. Install enterprise dependencies
go get github.com/golang-jwt/jwt/v5
go get golang.org/x/oauth2
go get github.com/sirupsen/logrus
go get github.com/gorilla/websocket

# 2. Build with enterprise features
go build -tags=enterprise -o api-security-scanner

# 3. Create enterprise configuration
cp config-enterprise.yaml config.yaml

# 4. Edit configuration for your environment
nano config.yaml

# 5. Create necessary directories
mkdir -p data/{tenants,history,metrics}
mkdir -p logs
mkdir -p static

# 6. Set permissions
chmod 755 data data/* logs static
chmod 644 config.yaml

# 7. Test the installation
./api-security-scanner -config config.yaml -validate-config
```

### Database Setup (Optional)

#### PostgreSQL Setup

```bash
# Install PostgreSQL
sudo apt install postgresql postgresql-contrib

# Start PostgreSQL
sudo systemctl start postgresql
sudo systemctl enable postgresql

# Create database and user
sudo -u postgres psql <<EOF
CREATE DATABASE api_scanner;
CREATE USER scanner WITH PASSWORD 'your-password';
GRANT ALL PRIVILEGES ON DATABASE api_scanner TO scanner;
ALTER USER scanner CREATEDB;
EOF

# Test connection
psql -h localhost -U scanner -d api_scanner
```

#### MySQL Setup

```bash
# Install MySQL
sudo apt install mysql-server

# Secure MySQL
sudo mysql_secure_installation

# Create database and user
sudo mysql <<EOF
CREATE DATABASE api_scanner;
CREATE USER 'scanner'@'localhost' IDENTIFIED BY 'your-password';
GRANT ALL PRIVILEGES ON api_scanner.* TO 'scanner'@'localhost';
FLUSH PRIVILEGES;
EOF

# Test connection
mysql -h localhost -u scanner -p api_scanner
```

### SIEM Integration Setup

#### Wazuh Integration

```bash
# 1. Install Wazuh manager
# Follow Wazuh installation guide

# 2. Configure Wazuh for API Security Scanner
sudo tee /var/ossec/etc/ossec.conf > /dev/null <<EOF
<!-- API Security Scanner Integration -->
<remote>
  <connection>syslog</connection>
  <port>514</port>
  <protocol>udp</protocol>
  <allowed-ips>127.0.0.1</allowed-ips>
</remote>

<decoder name="api-security-scanner">
  <program_name>^api-security-scanner</program_name>
</decoder>

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
</group>
EOF

# 3. Restart Wazuh
sudo systemctl restart wazuh-manager

# 4. Configure API Security Scanner for Wazuh
cp config-wazuh.yaml config.yaml
nano config.yaml
```

#### Splunk Integration

```bash
# 1. Enable HEC in Splunk
# Navigate to Settings -> Data Inputs -> HTTP Event Collector
# Create new HEC token with appropriate permissions

# 2. Configure API Security Scanner for Splunk
tee config-splunk.yaml > /dev/null <<EOF
siem:
  enabled: true
  type: "splunk"
  format: "json"
  endpoint_url: "https://splunk.company.com:8088/services/collector"
  auth_token: "your-splunk-hec-token"
EOF

# 3. Test configuration
./api-security-scanner -config config-splunk.yaml -test-siem
```

### Reverse Proxy Setup (Nginx)

```bash
# Install Nginx
sudo apt install nginx

# Create Nginx configuration
sudo tee /etc/nginx/sites-available/api-security-scanner > /dev/null <<EOF
server {
    listen 80;
    server_name scanner.company.com;

    # Security headers
    add_header X-Frame-Options "SAMEORIGIN" always;
    add_header X-Content-Type-Options "nosniff" always;
    add_header X-XSS-Protection "1; mode=block" always;
    add_header Referrer-Policy "strict-origin-when-cross-origin" always;

    # API endpoints
    location /api/ {
        proxy_pass http://localhost:8081;
        proxy_set_header Host \$host;
        proxy_set_header X-Real-IP \$remote_addr;
        proxy_set_header X-Forwarded-For \$proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto \$scheme;
    }

    # Dashboard
    location / {
        proxy_pass http://localhost:8080;
        proxy_set_header Host \$host;
        proxy_set_header X-Real-IP \$remote_addr;
        proxy_set_header X-Forwarded-For \$proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto \$scheme;

        # WebSocket support
        proxy_http_version 1.1;
        proxy_set_header Upgrade \$http_upgrade;
        proxy_set_header Connection "upgrade";
    }
}
EOF

# Enable site
sudo ln -s /etc/nginx/sites-available/api-security-scanner /etc/nginx/sites-enabled/

# Test and restart Nginx
sudo nginx -t
sudo systemctl restart nginx
```

### SSL/TLS Setup

```bash
# Install Certbot
sudo apt install certbot python3-certbot-nginx

# Obtain SSL certificate
sudo certbot --nginx -d scanner.company.com

# Configure SSL in Nginx
sudo tee /etc/nginx/sites-available/api-security-scanner > /dev/null <<EOF
server {
    listen 443 ssl http2;
    server_name scanner.company.com;

    ssl_certificate /etc/letsencrypt/live/scanner.company.com/fullchain.pem;
    ssl_certificate_key /etc/letsencrypt/live/scanner.company.com/privkey.pem;
    ssl_protocols TLSv1.2 TLSv1.3;
    ssl_ciphers HIGH:!aNULL:!MD5;

    # Security headers
    add_header Strict-Transport-Security "max-age=31536000; includeSubDomains" always;
    add_header X-Frame-Options "SAMEORIGIN" always;
    add_header X-Content-Type-Options "nosniff" always;
    add_header X-XSS-Protection "1; mode=block" always;

    # API endpoints
    location /api/ {
        proxy_pass http://localhost:8081;
        proxy_set_header Host \$host;
        proxy_set_header X-Real-IP \$remote_addr;
        proxy_set_header X-Forwarded-For \$proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto \$scheme;
    }

    # Dashboard
    location / {
        proxy_pass http://localhost:8080;
        proxy_set_header Host \$host;
        proxy_set_header X-Real-IP \$remote_addr;
        proxy_set_header X-Forwarded-For \$proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto \$scheme;

        # WebSocket support
        proxy_http_version 1.1;
        proxy_set_header Upgrade \$http_upgrade;
        proxy_set_header Connection "upgrade";
    }
}

server {
    listen 80;
    server_name scanner.company.com;
    return 301 https://\$server_name\$request_uri;
}
EOF

# Test and restart Nginx
sudo nginx -t
sudo systemctl restart nginx
```

### Load Balancer Setup (Optional)

```bash
# Install HAProxy
sudo apt install haproxy

# Configure HAProxy
sudo tee /etc/haproxy/haproxy.cfg > /dev/null <<EOF
global
    log /dev/log local0
    chroot /var/lib/haproxy
    stats socket /run/haproxy/admin.sock mode 660 level admin
    stats timeout 30s
    user haproxy
    group haproxy
    daemon

defaults
    log global
    mode http
    option httplog
    option dontlognull
    timeout connect 5000
    timeout client 50000
    timeout server 50000
    errorfile 400 /etc/haproxy/errors/400.http
    errorfile 403 /etc/haproxy/errors/403.http
    errorfile 408 /etc/haproxy/errors/408.http
    errorfile 500 /etc/haproxy/errors/500.http
    errorfile 502 /etc/haproxy/errors/502.http
    errorfile 503 /etc/haproxy/errors/503.http
    errorfile 504 /etc/haproxy/errors/504.http

# API Security Scanner frontend
frontend api_scanner_frontend
    bind *:80
    bind *:443 ssl crt /etc/letsencrypt/live/scanner.company.com/fullchain.pem
    redirect scheme https if !{ ssl_fc }
    default_backend api_scanner_backend

# API Security Scanner backend
backend api_scanner_backend
    balance roundrobin
    option httpchk GET /health
    server scanner1 192.168.1.10:8081 check
    server scanner2 192.168.1.11:8081 check
    server scanner3 192.168.1.12:8081 check

# Dashboard frontend
frontend dashboard_frontend
    bind *:8080
    default_backend dashboard_backend

# Dashboard backend
backend dashboard_backend
    balance roundrobin
    server dashboard1 192.168.1.10:8080 check
    server dashboard2 192.168.1.11:8080 check
    server dashboard3 192.168.1.12:8080 check
EOF

# Start HAProxy
sudo systemctl enable haproxy
sudo systemctl start haproxy
```

## 🔧 Configuration

### Environment Variables

Create a `.env` file for environment-specific configuration:

```bash
# .env
# Server configuration
SERVER_PORT=8081
SERVER_HOST=0.0.0.0

# Scanner configuration
SCANNER_RATE_LIMITING_REQUESTS_PER_SECOND=10
SCANNER_RATE_LIMITING_MAX_CONCURRENT_REQUESTS=5

# Authentication configuration
AUTH_ENABLED=true
AUTH_TYPE=basic
AUTH_USERNAME=admin
AUTH_PASSWORD=your-secure-password

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

# Database configuration (if using database)
DATABASE_URL=postgresql://scanner:your-password@localhost:5432/api_scanner
DATABASE_SSL_MODE=disable

# Logging configuration
LOG_LEVEL=info
LOG_FORMAT=json
LOG_FILE=logs/app.log

# Security configuration
SECRET_KEY=your-secret-key-here
JWT_SECRET=your-jwt-secret-here
ENCRYPTION_KEY=your-encryption-key-here
```

### Configuration Files

#### Basic Configuration (config.yaml)

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

#### Enterprise Configuration (config-enterprise.yaml)

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
```

## ✅ Verification

### Installation Verification

```bash
# Check version
./api-security-scanner -version

# Check help
./api-security-scanner -help

# Validate configuration
./api-security-scanner -config config.yaml -validate-config

# Test authentication
./api-security-scanner -config config.yaml -test-auth

# Test SIEM integration
./api-security-scanner -config config.yaml -test-siem

# Test database connection (if configured)
./api-security-scanner -config config.yaml -test-database
```

### Service Verification

```bash
# Check systemd service status
sudo systemctl status api-security-scanner

# Check service logs
sudo journalctl -u api-security-scanner -f

# Check if listening on correct ports
sudo netstat -tlnp | grep :8081
sudo netstat -tlnp | grep :8080

# Check process
ps aux | grep api-security-scanner
```

### API Verification

```bash
# Health check
curl http://localhost:8081/health

# API configuration
curl http://localhost:8081/api/config

# Dashboard access
curl http://localhost:8080/

# Metrics endpoint
curl http://localhost:8080/metrics
```

### Test Scan Verification

```bash
# Run test scan
./api-security-scanner -config config.yaml -scan -output json > test-scan.json

# Check scan results
cat test-scan.json

# Run scan with dashboard
./api-security-scanner -config config.yaml -dashboard &

# Access dashboard
open http://localhost:8081
```

## 🛠️ Troubleshooting

### Common Issues

#### Build Errors

**Problem**: `go build` fails with missing dependencies

```bash
# Solution: Install dependencies
go mod tidy
go get github.com/golang-jwt/jwt/v5
go get golang.org/x/oauth2
```

**Problem**: Import cycle errors

```bash
# Solution: Clean and rebuild
go clean -modcache
go mod download
go build
```

#### Configuration Errors

**Problem**: Invalid YAML configuration

```bash
# Solution: Validate YAML
python -c "import yaml; yaml.safe_load(open('config.yaml'))"

# Use scanner validation
./api-security-scanner -config config.yaml -validate-config -verbose
```

**Problem**: Configuration file not found

```bash
# Solution: Check file path and permissions
ls -la config.yaml
./api-security-scanner -config /full/path/to/config.yaml -scan
```

#### Runtime Errors

**Problem**: Port already in use

```bash
# Solution: Find and kill process
sudo lsof -i :8081
sudo kill -9 <PID>

# Or change port
./api-security-scanner -config config.yaml -server-port 8082
```

**Problem**: Permission denied

```bash
# Solution: Check permissions
ls -la api-security-scanner
chmod +x api-security-scanner

# Check data directory permissions
sudo chown -R $USER:$USER data/
sudo chown -R $USER:$USER logs/
```

#### Network Issues

**Problem**: Cannot connect to API endpoints

```bash
# Solution: Test network connectivity
curl -I https://api.example.com/users

# Test with specific headers
curl -H "User-Agent: API-Security-Scanner/4.0" https://api.example.com/users

# Check DNS resolution
nslookup api.example.com
```

**Problem**: SIEM connection fails

```bash
# Solution: Test SIEM connectivity
telnet wazuh.company.com 514

# Test with scanner
./api-security-scanner -config config.yaml -test-siem -verbose
```

#### Performance Issues

**Problem**: High memory usage

```bash
# Solution: Monitor memory usage
ps aux | grep api-security-scanner
top -p <PID>

# Adjust rate limiting
echo "rate_limiting:
  requests_per_second: 5
  max_concurrent_requests: 3" >> config.yaml
```

**Problem**: Slow scan performance

```bash
# Solution: Enable debug logging
./api-security-scanner -config config.yaml -debug -scan

# Check network latency
ping api.example.com
curl -w "@curl-format.txt" -o /dev/null -s https://api.example.com/users
```

### Debug Mode

Enable debug mode for detailed logging:

```bash
# Enable all debug logging
./api-security-scanner -config config.yaml -debug -scan

# Enable specific debug components
./api-security-scanner -config config.yaml -debug=scanner,auth,siem -scan

# Enable debug with verbose output
./api-security-scanner -config config.yaml -debug -verbose -scan
```

### Log Analysis

Check log files for detailed error information:

```bash
# View application logs
tail -f logs/app.log

# View scanner logs
tail -f logs/scanner.log

# View SIEM logs
tail -f logs/siem.log

# View systemd logs
sudo journalctl -u api-security-scanner -f

# Filter logs by level
grep "ERROR" logs/app.log
grep "WARN" logs/app.log
```

### Performance Tuning

#### System Tuning

```bash
# Increase file descriptor limit
echo "* soft nofile 65536" | sudo tee -a /etc/security/limits.conf
echo "* hard nofile 65536" | sudo tee -a /etc/security/limits.conf

# Tune kernel parameters
echo "net.core.somaxconn = 65536" | sudo tee -a /etc/sysctl.conf
echo "net.ipv4.tcp_max_syn_backlog = 65536" | sudo tee -a /etc/sysctl.conf
echo "net.core.netdev_max_backlog = 65536" | sudo tee -a /etc/sysctl.conf
sudo sysctl -p
```

#### Application Tuning

```yaml
# Performance optimization in config.yaml
scanner:
  rate_limiting:
    requests_per_second: 20
    max_concurrent_requests: 10
    burst_limit: 30

  timeout:
    connect_timeout: 15s
    read_timeout: 30s
    write_timeout: 15s
    overall_timeout: 60s

  retry:
    max_attempts: 3
    backoff_factor: 2.0
    max_delay: 30s

metrics:
  update_interval: 60s
  retention_days: 7
  dashboard:
    update_interval: 30s
    max_connections: 50
```

## 🔄 Upgrade Guide

### Version Upgrade

```bash
# 1. Backup current configuration and data
cp config.yaml config.yaml.backup
cp -r data/ data.backup/
cp -r logs/ logs.backup/

# 2. Download new version
wget https://github.com/elliotsecops/API-Security-Scanner/releases/latest/download/api-security-scanner-linux-amd64

# 3. Stop service
sudo systemctl stop api-security-scanner

# 4. Replace binary
sudo cp api-security-scanner-linux-amd64 /usr/local/bin/api-security-scanner
sudo chmod +x /usr/local/bin/api-security-scanner

# 5. Update configuration (if needed)
# Check release notes for configuration changes

# 6. Start service
sudo systemctl start api-security-scanner

# 7. Verify upgrade
./api-security-scanner -version
sudo systemctl status api-security-scanner
```

### Database Migration

```bash
# For PostgreSQL
sudo -u postgres psql api_scanner < migration.sql

# For MySQL
mysql -u scanner -p api_scanner < migration.sql
```

### Configuration Migration

```bash
# Use configuration migration tool
./api-security-scanner -migrate-config -old config-old.yaml -new config-new.yaml

# Manual migration
# Compare old and new configuration files
diff config.yaml.backup config.yaml
```

## 🗑️ Uninstallation

### Manual Uninstallation

```bash
# 1. Stop service
sudo systemctl stop api-security-scanner
sudo systemctl disable api-security-scanner

# 2. Remove service file
sudo rm /etc/systemd/system/api-security-scanner.service
sudo systemctl daemon-reload

# 3. Remove binary
sudo rm /usr/local/bin/api-security-scanner

# 4. Remove configuration and data
sudo rm -rf /opt/api-security-scanner
sudo rm -rf /var/log/api-security-scanner
sudo rm -rf /var/lib/api-security-scanner

# 5. Remove user
sudo userdel scanner

# 6. Remove logrotate configuration
sudo rm /etc/logrotate.d/api-security-scanner
```

### Package Uninstallation

#### APT
```bash
sudo apt remove api-security-scanner
sudo apt autoremove
```

#### YUM/DNF
```bash
sudo yum remove api-security-scanner
sudo yum autoremove
```

### Docker Cleanup

```bash
# Stop and remove container
docker stop api-security-scanner
docker rm api-security-scanner

# Remove image
docker rmi api-security-scanner

# Remove volumes
docker volume rm api-security-scanner_data
docker volume rm api-security-scanner_logs
```

### Complete Cleanup

```bash
# Remove all related files and directories
sudo rm -rf /opt/api-security-scanner
sudo rm -rf /var/log/api-security-scanner
sudo rm -rf /var/lib/api-security-scanner
sudo rm -f /etc/logrotate.d/api-security-scanner
sudo rm -f /etc/nginx/sites-available/api-security-scanner
sudo rm -f /etc/nginx/sites-enabled/api-security-scanner

# Reload nginx
sudo systemctl reload nginx

# Remove user
sudo userdel scanner

# Remove database (if created)
sudo -u postgres dropdb api_scanner
sudo -u postgres dropuser scanner
```

---

This comprehensive installation guide covers all aspects of installing and setting up the API Security Scanner, from basic development setups to enterprise deployments. For additional help and support, refer to the [troubleshooting section](#troubleshooting) or open an issue on GitHub.