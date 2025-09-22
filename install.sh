#!/bin/bash

# API Security Scanner - Installation Script
# This script automates the entire setup process

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

echo -e "${BLUE}🚀 API Security Scanner - Installation Script${NC}"
echo -e "${BLUE}==========================================${NC}"

# Check if script is run as root
if [ "$EUID" -eq 0 ]; then
    echo -e "${YELLOW}⚠️  Running as root. Some checks may behave differently.${NC}"
fi

# Detect operating system
detect_os() {
    if [[ "$OSTYPE" == "linux-gnu"* ]]; then
        OS="linux"
        if [ -f /etc/debian_version ]; then
            DISTRO="debian"
        elif [ -f /etc/redhat-release ]; then
            DISTRO="redhat"
        elif [ -f /etc/arch-release ]; then
            DISTRO="arch"
        else
            DISTRO="unknown"
        fi
    elif [[ "$OSTYPE" == "darwin"* ]]; then
        OS="macos"
        DISTRO="macos"
    elif [[ "$OSTYPE" == "msys" ]] || [[ "$OSTYPE" == "cygwin" ]]; then
        OS="windows"
        DISTRO="windows"
    else
        OS="unknown"
        DISTRO="unknown"
    fi

    echo -e "${GREEN}📋 Detected OS: $OS ($DISTRO)${NC}"
}

# Check and install Go
install_go() {
    echo -e "${YELLOW}🔍 Checking Go installation...${NC}"

    if command -v go &> /dev/null; then
        GO_VERSION=$(go version | awk '{print $3}' | sed 's/go//')
        echo -e "${GREEN}✅ Go is installed: version $GO_VERSION${NC}"

        # Check if version is 1.24 or higher
        if python3 -c "import sys; version_parts = '$GO_VERSION'.split('.'); major, minor = int(version_parts[0]), int(version_parts[1]); sys.exit(0 if major > 1 or (major == 1 and minor >= 24) else 1)"; then
            echo -e "${GREEN}✅ Go version is compatible (1.24+ required)${NC}"
        else
            echo -e "${RED}❌ Go version $GO_VERSION is too old. Please install Go 1.24 or higher.${NC}"
            echo -e "${YELLOW}💡 Download from: https://golang.org/dl/${NC}"
            exit 1
        fi
    else
        echo -e "${RED}❌ Go is not installed.${NC}"

        if [[ "$DISTRO" == "debian" ]]; then
            echo -e "${YELLOW}📦 Installing Go on Debian/Ubuntu...${NC}"
            sudo apt update
            sudo apt install -y golang-go
        elif [[ "$DISTRO" == "redhat" ]]; then
            echo -e "${YELLOW}📦 Installing Go on RedHat/CentOS...${NC}"
            sudo yum install -y golang
        elif [[ "$DISTRO" == "arch" ]]; then
            echo -e "${YELLOW}📦 Installing Go on Arch Linux...${NC}"
            sudo pacman -S go
        elif [[ "$OS" == "macos" ]]; then
            echo -e "${YELLOW}📦 Installing Go on macOS...${NC}"
            if command -v brew &> /dev/null; then
                brew install go
            else
                echo -e "${RED}❌ Homebrew not found. Please install Homebrew first:${NC}"
                echo -e "${YELLOW}/bin/bash -c \"\$(curl -fsSL https://raw.githubusercontent.com/Homebrew/install/HEAD/install.sh)\"${NC}"
                exit 1
            fi
        else
            echo -e "${RED}❌ Please install Go manually:${NC}"
            echo -e "${YELLOW}💡 Download from: https://golang.org/dl/${NC}"
            exit 1
        fi

        # Verify installation
        if command -v go &> /dev/null; then
            echo -e "${GREEN}✅ Go installed successfully${NC}"
        else
            echo -e "${RED}❌ Go installation failed. Please install manually.${NC}"
            exit 1
        fi
    fi
}

# Check and install Node.js
install_nodejs() {
    echo -e "${YELLOW}🔍 Checking Node.js installation...${NC}"

    if command -v node &> /dev/null; then
        NODE_VERSION=$(node --version | sed 's/v//')
        echo -e "${GREEN}✅ Node.js is installed: version $NODE_VERSION${NC}"

        # Check if version is 16 or higher
        if python3 -c "import sys; version_parts = '$NODE_VERSION'.split('.'); major, minor = int(version_parts[0]), int(version_parts[1]); sys.exit(0 if major > 16 or (major == 16 and minor >= 0) else 1)"; then
            echo -e "${GREEN}✅ Node.js version is compatible (v16+ required)${NC}"
        else
            echo -e "${RED}❌ Node.js version $NODE_VERSION is too old. Please install Node.js v16 or higher.${NC}"
            echo -e "${YELLOW}💡 Download from: https://nodejs.org/${NC}"
            exit 1
        fi
    else
        echo -e "${RED}❌ Node.js is not installed.${NC}"

        if [[ "$DISTRO" == "debian" ]]; then
            echo -e "${YELLOW}📦 Installing Node.js on Debian/Ubuntu...${NC}"
            sudo apt update
            sudo apt install -y nodejs npm
        elif [[ "$DISTRO" == "redhat" ]]; then
            echo -e "${YELLOW}📦 Installing Node.js on RedHat/CentOS...${NC}"
            sudo yum install -y nodejs npm
        elif [[ "$DISTRO" == "arch" ]]; then
            echo -e "${YELLOW}📦 Installing Node.js on Arch Linux...${NC}"
            sudo pacman -S nodejs npm
        elif [[ "$OS" == "macos" ]]; then
            echo -e "${YELLOW}📦 Installing Node.js on macOS...${NC}"
            if command -v brew &> /dev/null; then
                brew install node
            else
                echo -e "${RED}❌ Homebrew not found. Please install Homebrew first.${NC}"
                exit 1
            fi
        else
            echo -e "${RED}❌ Please install Node.js manually:${NC}"
            echo -e "${YELLOW}💡 Download from: https://nodejs.org/${NC}"
            exit 1
        fi

        # Verify installation
        if command -v node &> /dev/null && command -v npm &> /dev/null; then
            echo -e "${GREEN}✅ Node.js and npm installed successfully${NC}"
        else
            echo -e "${RED}❌ Node.js installation failed. Please install manually.${NC}"
            exit 1
        fi
    fi
}

# Install GUI dependencies
install_gui_deps() {
    echo -e "${YELLOW}📦 Installing GUI dependencies...${NC}"

    if [ ! -d "gui" ]; then
        echo -e "${RED}❌ GUI directory not found. Please run this script from the project root.${NC}"
        exit 1
    fi

    cd gui

    if [ ! -d "node_modules" ]; then
        echo -e "${YELLOW}📦 Running npm install...${NC}"
        npm install
        echo -e "${GREEN}✅ GUI dependencies installed${NC}"
    else
        echo -e "${GREEN}✅ GUI dependencies already installed${NC}"
    fi

    cd ..
}

# Build the application
build_application() {
    echo -e "${YELLOW}🔨 Building the application...${NC}"

    # Download Go dependencies
    echo -e "${YELLOW}📦 Downloading Go dependencies...${NC}"
    go mod download
    go mod tidy

    # Build the main application
    echo -e "${YELLOW}🔨 Compiling the application...${NC}"
    go build -o api-security-scanner .

    echo -e "${GREEN}✅ Application built successfully${NC}"
}

# Create desktop shortcut (Linux only)
create_desktop_shortcut() {
    if [[ "$OS" == "linux" ]] && [[ -d "$HOME/Desktop" ]]; then
        echo -e "${YELLOW}📋 Creating desktop shortcut...${NC}"

        cat > "$HOME/Desktop/api-security-scanner.desktop" << EOF
[Desktop Entry]
Version=1.0
Type=Application
Name=API Security Scanner
Comment=Enterprise-grade API security testing platform
Exec=$(pwd)/run.sh prod
Icon=$(pwd)/gui/src/favicon.ico
Terminal=true
Categories=Security;Development;
EOF

        chmod +x "$HOME/Desktop/api-security-scanner.desktop"
        echo -e "${GREEN}✅ Desktop shortcut created${NC}"
    fi
}

# Create configuration file
create_config() {
    echo -e "${YELLOW}⚙️  Creating configuration file...${NC}"

    if [ ! -f "config.yaml" ]; then
        cat > config.yaml << EOF
# API Security Scanner Configuration
# Generated by install.sh

# API endpoints to test
api_endpoints:
  - url: "https://httpbin.org/get"
    method: "GET"
  - url: "https://httpbin.org/post"
    method: "POST"
    body: '{"test": "data"}'

# Authentication credentials
auth:
  username: "admin"
  password: "admin"

# Rate limiting configuration
rate_limiting:
  requests_per_second: 10
  max_concurrent_requests: 5

# Custom headers
headers:
  "User-Agent": "API-Security-Scanner/4.0"
  "X-Scanner": "true"

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

# NoSQL injection test payloads
nosql_payloads:
  - "{\$ne: null}"
  - "{\$gt: ''}"
  - "{\$or: [1,1]}"
  - "{\$where: 'sleep(100)'}"

# GUI configuration
gui:
  enabled: true
  development: false
  port: 8080

# Historical data configuration
historical_data:
  enabled: true
  storage_path: "./history"
  retention_days: 30
  compare_previous: true
  trend_analysis: true

# Metrics configuration
metrics:
  enabled: true
  port: 8080
  update_interval: 30s
  retention_days: 30
EOF

        echo -e "${GREEN}✅ Configuration file created: config.yaml${NC}"
    else
        echo -e "${GREEN}✅ Configuration file already exists${NC}"
    fi
}

# Run initial tests
run_tests() {
    echo -e "${YELLOW}🧪 Running initial tests...${NC}"

    # Test the application
    echo -e "${YELLOW}🧪 Testing application version...${NC}"
    ./api-security-scanner -version

    echo -e "${GREEN}✅ Application tests passed${NC}"
}

# Show success message
show_success() {
    echo ""
    echo -e "${GREEN}🎉 Installation completed successfully!${NC}"
    echo ""
    echo -e "${BLUE}🚀 Quick Start Commands:${NC}"
    echo -e "${GREEN}  ./run.sh dev     ${NC}- Start in development mode"
    echo -e "${GREEN}  ./run.sh prod    ${NC}- Start in production mode"
    echo -e "${GREEN}  ./run.sh help    ${NC}- Show all commands"
    echo ""
    echo -e "${BLUE}🔧 Default Configuration:${NC}"
    echo -e "${GREEN}  GUI URL:           http://localhost:8080${NC}"
    echo -e "${GREEN}  Backend API:       http://localhost:8080/api${NC}"
    echo -e "${GREEN}  Default Login:     admin / admin${NC}"
    echo ""
    echo -e "${BLUE}📚 Documentation:${NC}"
    echo -e "${GREEN}  README.md          - Full documentation${NC}"
    echo -e "${GREEN}  GUIDE.md           - User guide${NC}"
    echo -e "${GREEN}  CONFIGURATION.md   - Configuration options${NC}"
    echo ""
    echo -e "${YELLOW}⚠️  First run may take a few moments as the application initializes.${NC}"
}

# Main installation process
main() {
    detect_os
    install_go
    install_nodejs
    install_gui_deps
    build_application
    create_config
    create_desktop_shortcut
    run_tests
    show_success
}

# Handle command line arguments
case "${1:-install}" in
    "install")
        main
        ;;
    "deps")
        detect_os
        install_go
        install_nodejs
        echo -e "${GREEN}✅ Dependencies installed successfully${NC}"
        ;;
    "help")
        echo -e "${BLUE}API Security Scanner - Installation Script${NC}"
        echo ""
        echo "Usage: $0 [COMMAND]"
        echo ""
        echo "Commands:"
        echo -e "  ${GREEN}install${NC}    Full installation (default)"
        echo -e "  ${GREEN}deps${NC}       Install system dependencies only"
        echo -e "  ${GREEN}help${NC}        Show this help message"
        ;;
    *)
        echo -e "${RED}❌ Unknown command: $1${NC}"
        echo -e "${YELLOW}Use '$0 help' for available commands${NC}"
        exit 1
        ;;
esac