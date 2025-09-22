#!/bin/bash

# API Security Scanner - Easy Launcher
# This script provides an optimized workflow for running the application

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Configuration
GUI_PORT=3000
BACKEND_PORT=8080
PROJECT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
GUI_DIR="$PROJECT_DIR/gui"

echo -e "${BLUE}🚀 API Security Scanner - Easy Launcher${NC}"
echo -e "${BLUE}==========================================${NC}"

# Check dependencies
check_dependencies() {
    echo -e "${YELLOW}🔍 Checking dependencies...${NC}"

    if ! command -v go &> /dev/null; then
        echo -e "${RED}❌ Go is not installed. Please install Go 1.24 or higher.${NC}"
        exit 1
    fi

    if ! command -v node &> /dev/null; then
        echo -e "${RED}❌ Node.js is not installed. Please install Node.js v16 or higher.${NC}"
        exit 1
    fi

    if ! command -v npm &> /dev/null; then
        echo -e "${RED}❌ npm is not installed. Please install npm.${NC}"
        exit 1
    fi

    echo -e "${GREEN}✅ All dependencies are installed.${NC}"
}

# Install GUI dependencies if needed
install_gui_deps() {
    if [ ! -d "$GUI_DIR/node_modules" ]; then
        echo -e "${YELLOW}📦 Installing GUI dependencies...${NC}"
        cd "$GUI_DIR"
        npm install
        cd "$PROJECT_DIR"
        echo -e "${GREEN}✅ GUI dependencies installed.${NC}"
    else
        echo -e "${GREEN}✅ GUI dependencies already installed.${NC}"
    fi
}

# Build the Go application
build_app() {
    echo -e "${YELLOW}🔨 Building the application...${NC}"
    go build -o api-security-scanner .
    echo -e "${GREEN}✅ Application built successfully.${NC}"
}

# Start the GUI in development mode
start_gui_dev() {
    echo -e "${YELLOW}🌐 Starting GUI in development mode...${NC}"
    cd "$GUI_DIR"
    npm start &
    GUI_PID=$!
    cd "$PROJECT_DIR"
    echo $GUI_PID > .gui_dev_pid
    echo -e "${GREEN}✅ GUI started on http://localhost:$GUI_PORT${NC}"
}

# Start the GUI in production mode
build_gui_prod() {
    echo -e "${YELLOW}🔨 Building GUI for production...${NC}"
    cd "$GUI_DIR"
    npm run build
    cd "$PROJECT_DIR"
    echo -e "${GREEN}✅ GUI built for production.${NC}"
}

# Start the backend
start_backend() {
    echo -e "${YELLOW}🔧 Starting backend server...${NC}"
    ./api-security-scanner -dashboard &
    BACKEND_PID=$!
    echo $BACKEND_PID > .backend_pid
    echo -e "${GREEN}✅ Backend started on http://localhost:$BACKEND_PORT${NC}"
}

# Stop running processes
stop_processes() {
    echo -e "${YELLOW}🛑 Stopping processes...${NC}"

    if [ -f .gui_dev_pid ]; then
        GUI_PID=$(cat .gui_dev_pid)
        if ps -p $GUI_PID > /dev/null; then
            kill $GUI_PID
            echo -e "${GREEN}✅ GUI development server stopped.${NC}"
        fi
        rm .gui_dev_pid
    fi

    if [ -f .backend_pid ]; then
        BACKEND_PID=$(cat .backend_pid)
        if ps -p $BACKEND_PID > /dev/null; then
            kill $BACKEND_PID
            echo -e "${GREEN}✅ Backend server stopped.${NC}"
        fi
        rm .backend_pid
    fi
}

# Show help
show_help() {
    echo -e "${BLUE}Usage: $0 [COMMAND]${NC}"
    echo ""
    echo "Commands:"
    echo -e "  ${GREEN}dev${NC}         Start in development mode (GUI dev server + backend)"
    echo -e "  ${GREEN}prod${NC}        Start in production mode (built GUI + backend)"
    echo -e "  ${GREEN}backend${NC}     Start only the backend server"
    echo -e "  ${GREEN}gui${NC}         Start only the GUI development server"
    echo -e "  ${GREEN}build${NC}       Build the application and GUI"
    echo -e "  ${GREEN}install${NC}     Install dependencies only"
    echo -e "  ${GREEN}stop${NC}        Stop all running processes"
    echo -e "  ${GREEN}clean${NC}       Clean build artifacts"
    echo -e "  ${GREEN}help${NC}        Show this help message"
    echo ""
    echo "Examples:"
    echo -e "  $0 dev          # Start in development mode"
    echo -e "  $0 prod         # Start in production mode"
    echo -e "  $0 stop         # Stop all processes"
}

# Clean build artifacts
clean() {
    echo -e "${YELLOW}🧹 Cleaning build artifacts...${NC}"
    stop_processes
    rm -f api-security-scanner
    rm -rf gui/build
    rm -f .gui_dev_pid .backend_pid
    echo -e "${GREEN}✅ Build artifacts cleaned.${NC}"
}

# Trap Ctrl+C and call stop_processes
trap stop_processes EXIT

# Main script logic
case "${1:-dev}" in
    "dev")
        check_dependencies
        install_gui_deps
        build_app
        start_gui_dev
        start_backend
        echo -e "${BLUE}🎉 Development mode started!${NC}"
        echo -e "${BLUE}GUI: http://localhost:$GUI_PORT${NC}"
        echo -e "${BLUE}Backend API: http://localhost:$BACKEND_PORT${NC}"
        echo -e "${YELLOW}Press Ctrl+C to stop all services${NC}"
        wait
        ;;
    "prod")
        check_dependencies
        install_gui_deps
        build_app
        build_gui_prod
        start_backend
        echo -e "${BLUE}🎉 Production mode started!${NC}"
        echo -e "${BLUE}Web Interface: http://localhost:$BACKEND_PORT${NC}"
        echo -e "${YELLOW}Press Ctrl+C to stop the service${NC}"
        wait
        ;;
    "backend")
        check_dependencies
        build_app
        start_backend
        echo -e "${BLUE}🔧 Backend server started!${NC}"
        echo -e "${BLUE}Backend API: http://localhost:$BACKEND_PORT${NC}"
        echo -e "${YELLOW}Press Ctrl+C to stop the service${NC}"
        wait
        ;;
    "gui")
        check_dependencies
        install_gui_deps
        start_gui_dev
        echo -e "${BLUE}🌐 GUI development server started!${NC}"
        echo -e "${BLUE}GUI: http://localhost:$GUI_PORT${NC}"
        echo -e "${YELLOW}Press Ctrl+C to stop the service${NC}"
        wait
        ;;
    "build")
        check_dependencies
        install_gui_deps
        build_app
        build_gui_prod
        echo -e "${GREEN}✅ Application and GUI built successfully.${NC}"
        ;;
    "install")
        check_dependencies
        install_gui_deps
        echo -e "${GREEN}✅ All dependencies installed.${NC}"
        ;;
    "stop")
        stop_processes
        echo -e "${GREEN}✅ All processes stopped.${NC}"
        ;;
    "clean")
        clean
        ;;
    "help")
        show_help
        ;;
    *)
        echo -e "${RED}❌ Unknown command: $1${NC}"
        echo ""
        show_help
        exit 1
        ;;
esac