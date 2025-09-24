# 🚀 API Security Scanner - Quick Start Guide

This guide will get you up and running in minutes with the optimized workflow. The easiest way to start is with the integrated Docker Compose setup that includes both the scanner and a vulnerable test API.

## ⚡ One-Command Setup with Complete Test Environment

### Option 1: Docker Compose (Recommended for Testing)
```bash
# Clone and start the complete environment in seconds
git clone https://github.com/elliotsecops/API-Security-Scanner.git
cd API-Security-Scanner

# Start both the scanner and OWASP Juice Shop (vulnerable test API)
docker-compose up -d

# Access the dashboard at: http://localhost:8080
# The API runs on: http://localhost:8081
# The test API runs on: http://localhost:3000

# Run a test scan:
docker exec api-security-scanner ./api-security-scanner -config config-test.yaml -scan
```

### Option 2: Quick Start (Dependencies Already Installed)
```bash
# Just run the application
./run.sh dev
```

## 🎯 Running the Application

### Development Mode (For Testing & Development)
```bash
./run.sh dev
```
- Starts GUI on `http://localhost:3000`
- Starts backend on `http://localhost:8080`
- Auto-reloads when code changes
- Best for development and testing

### Production Mode (For Regular Use)
```bash
./run.sh prod
```
- Built GUI served by backend
- Single process on `http://localhost:8080`
- Optimized for performance
- Best for regular use

## 🔧 Common Commands

```bash
# Build the application
./run.sh build

# Start only the backend
./run.sh backend

# Start only the GUI development server
./run.sh gui

# Install dependencies
./run.sh install

# Stop all running processes
./run.sh stop

# Clean build artifacts
./run.sh clean

# Show help
./run.sh help
```

## 🌐 Accessing the GUI

### Development Mode
- **GUI**: http://localhost:3000
- **Backend API**: http://localhost:8080

### Production Mode
- **Web Interface**: http://localhost:8080
- **Backend API**: http://localhost:8080/api

## 🔑 Default Login

- **Username**: `admin`
- **Password**: `admin`

## 📋 What's Included

### Automated Features
- ✅ **Dependency checking** - Validates Go and Node.js versions
- ✅ **Automatic installation** - Installs missing dependencies
- ✅ **Build automation** - Compiles Go app and builds React GUI
- ✅ **Process management** - Starts and stops services cleanly
- ✅ **Configuration generation** - Creates default config file
- ✅ **Desktop shortcuts** - Creates launcher icon (Linux)

### Smart Features
- 🚀 **Auto-detection** - Detects operating system and dependencies
- 🛡️ **Error handling** - Graceful handling of missing dependencies
- 📊 **Progress feedback** - Color-coded status messages
- 🔧 **Port management** - Handles port conflicts automatically
- 🧹 **Cleanup** - Properly stops processes and cleans up

## 🔨 Manual Installation (If Needed)

If the automated script doesn't work for your system:

### Prerequisites
- **Go 1.24+**: https://golang.org/dl/
- **Node.js v16+**: https://nodejs.org/

### Step-by-Step
```bash
# 1. Clone the repository
git clone https://github.com/elliotsecops/API-Security-Scanner.git
cd API-Security-Scanner

# 2. Install GUI dependencies
cd gui
npm install
cd ..

# 3. Build the Go application
go build -o api-security-scanner .

# 4. Run the application
./api-security-scanner -dashboard
```

## 🐛 Troubleshooting

### Port Already in Use
```bash
# Find what's using the port
lsof -i :3000  # GUI port
lsof -i :8080  # Backend port

# Kill the process
kill -9 <PID>
```

### Permission Issues
```bash
# Make scripts executable
chmod +x install.sh run.sh
```

### Go Version Issues
```bash
# Check Go version
go version

# Install correct version (example for Ubuntu)
sudo apt install golang-go
```

### Node.js Issues
```bash
# Check Node.js version
node --version
npm --version

# Install using NVM (recommended)
curl -o- https://raw.githubusercontent.com/nvm-sh/nvm/v0.39.0/install.sh | bash
nvm install 16
nvm use 16
```

## 📚 Next Steps

1. **Explore the GUI** - Navigate through dashboard, scanner, and results
2. **Configure APIs** - Edit `config.yaml` to add your API endpoints
3. **Run Security Scan** - Use the GUI to configure and run security tests
4. **View Results** - Analyze vulnerabilities and security reports
5. **Customize Settings** - Modify configuration for your specific needs

## 🆘 Getting Help

- **Documentation**: See `README.md`, `GUIDE.md`, `CONFIGURATION.md`
- **Issues**: https://github.com/elliotsecops/API-Security-Scanner/issues
- **Discussions**: https://github.com/elliotsecops/API-Security-Scanner/discussions

---

**🎉 You're ready to start securing your APIs!**