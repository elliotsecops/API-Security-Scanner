# API Security Scanner GUI Guide

## Overview

The API Security Scanner now includes a comprehensive React-based graphical user interface (GUI) that provides an intuitive web interface for managing security scans, viewing results, and monitoring system metrics.

## Features

### 🎯 Core Features
- **Dashboard**: Real-time metrics and system health monitoring
- **Scanner**: Configure and run security scans with advanced options
- **Results**: View and analyze scan results with detailed vulnerability reports
- **Tenants**: Multi-tenant management interface
- **Settings**: Configuration management and system preferences

### 🔧 Technical Features
- **Real-time Updates**: WebSocket support for live data updates
- **Responsive Design**: Mobile-friendly interface using Material-UI
- **Data Visualization**: Interactive charts and graphs using Chart.js
- **SPA Architecture**: Single Page Application with React Router
- **API Integration**: Full REST API integration with backend services

## Quick Start

### Prerequisites
- Node.js (v16 or higher)
- npm or yarn
- Go 1.24+ (for backend)

### Development Setup

1. **Start the Backend Server**
```bash
# Build and start the API Security Scanner
go build -o api-security-scanner .
./api-security-scanner --dashboard --config config.yaml
```

2. **Start the GUI Development Server**
```bash
# Navigate to GUI directory
cd gui

# Install dependencies
npm install

# Start development server
npm start
```

3. **Access the GUI**
- GUI Development Server: `http://localhost:3000`
- Backend API: `http://localhost:8080`
- Default credentials: `admin` / `admin`

### Production Deployment

1. **Build the GUI**
```bash
cd gui
npm run build
```

2. **Deploy with Backend**
```bash
# The GUI build files are automatically served by the backend
./api-security-scanner --dashboard
```

3. **Access the Application**
- Single server deployment: `http://localhost:8080`
- No separate GUI server required

## GUI Architecture

### Component Structure
```
gui/src/
├── components/          # Reusable UI components
│   ├── Dashboard.js    # Main dashboard with metrics
│   ├── Scanner.js      # Scan configuration interface
│   ├── Results.js      # Results viewing and analysis
│   ├── Tenants.js      # Tenant management
│   └── Settings.js     # Configuration settings
├── contexts/           # React contexts for state management
│   ├── AuthContext.js # Authentication state
│   ├── WebSocketContext.js # Real-time updates
│   └── MetricsContext.js # Metrics data management
├── App.js              # Main application component
└── index.js            # Application entry point
```

### Key Components

#### Dashboard (`Dashboard.js`)
- Real-time system metrics display
- Vulnerability distribution charts
- Scan trends and activity feeds
- Resource usage monitoring

#### Scanner (`Scanner.js`)
- Quick scan configuration
- Advanced scan options
- Scan templates and history
- Real-time progress tracking

#### Results (`Results.js`)
- Scan results visualization
- Detailed vulnerability reports
- Export functionality
- Filtering and search capabilities

#### Tenants (`Tenants.js`)
- Multi-tenant management
- Tenant configuration
- Resource allocation
- Usage monitoring

#### Settings (`Settings.js`)
- System configuration
- User management
- API key management
- Integration settings

## API Integration

### Authentication
```javascript
// Login API
POST /api/auth/login
{
  "username": "admin",
  "password": "admin"
}

// Response
{
  "token": "jwt-token",
  "user": {
    "id": "1",
    "username": "admin",
    "role": "administrator"
  }
}
```

### System Metrics
```javascript
// Get system metrics
GET /api/system

// Response
{
  "total_scans": 150,
  "active_tenants": 5,
  "total_endpoints": 1000,
  "vulnerabilities": {
    "critical": 12,
    "high": 45,
    "medium": 78,
    "low": 23
  }
}
```

### Scan Management
```javascript
// Get scan results
GET /api/scans

// Response
[{
  "id": "scan_001",
  "name": "Daily Security Scan",
  "tenant_id": "tenant-001",
  "started_at": "2024-01-15T10:00:00Z",
  "completed_at": "2024-01-15T10:15:00Z",
  "average_score": 85.5,
  "risk_level": "medium",
  "total_vulnerabilities": 12,
  "endpoints": [...]
}]
```

### Tenant Management
```javascript
// Get tenants
GET /api/tenants

// Response
[{
  "id": "tenant-001",
  "name": "Enterprise Corp",
  "description": "Main enterprise tenant",
  "is_active": true,
  "settings": {
    "max_endpoints": 100,
    "scan_frequency": "daily"
  }
}]
```

## Development Workflow

### 1. Local Development
```bash
# Terminal 1: Start backend
./api-security-scanner --dashboard

# Terminal 2: Start GUI development server
cd gui && npm start
```

### 2. Making Changes
- Backend changes: Restart Go server
- Frontend changes: Hot reload automatically
- API changes: Update both frontend API calls and backend endpoints

### 3. Testing
```bash
# Run frontend tests
cd gui
npm test

# Run backend tests
go test ./...

# Integration testing
# Start both servers and test full workflow
```

## Configuration

### Backend Configuration
The backend server automatically detects and serves GUI build files:
- Looks for `./gui/build/index.html`
- Falls back to legacy dashboard if not found
- Serves all API endpoints alongside GUI

### Frontend Configuration
Key configuration files:
- `gui/package.json`: Dependencies and scripts
- `gui/src/contexts/`: State management
- `gui/public/index.html`: HTML template

### Environment Variables
```bash
# Backend (config.yaml)
server:
  port: 8080

# Frontend (gui/package.json)
"proxy": "http://localhost:8080"
```

## Deployment

### Development Deployment
- Separate frontend and backend servers
- Hot reload enabled
- Debug mode active

### Production Deployment
- Single server deployment
- Static files served by Go backend
- Optimized build with minification

### Docker Deployment
```dockerfile
# Multi-stage build
FROM node:16 AS gui-build
WORKDIR /app/gui
COPY gui/package*.json ./
RUN npm install
COPY gui/ ./
RUN npm run build

FROM golang:1.24 AS backend-build
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
COPY --from=gui-build /app/gui/build ./gui/build
RUN go build -o api-security-scanner .

FROM alpine:latest
COPY --from=backend-build /app/api-security-scanner .
CMD ["./api-security-scanner", "--dashboard"]
```

## Troubleshooting

### Common Issues

#### GUI Not Loading
- Check if backend server is running on port 8080
- Verify GUI build exists in `./gui/build/`
- Check console for JavaScript errors

#### API Calls Failing
- Verify backend server is running
- Check CORS settings
- Verify authentication token

#### Build Errors
- Clear node_modules: `rm -rf node_modules && npm install`
- Check Node.js version (requires v16+)
- Verify all dependencies are installed

### Debug Mode
```bash
# Enable debug logging
export DEBUG=api-security-scanner:*

# Start with verbose output
./api-security-scanner --dashboard --log-level debug
```

## Contributing

### Frontend Development
1. Follow React best practices
2. Use Material-UI components
3. Implement responsive design
4. Add proper error handling

### Backend Development
1. Follow Go coding standards
2. Implement proper API documentation
3. Add comprehensive error handling
4. Include unit tests

### Design Guidelines
- Use the existing Material-UI theme
- Follow established component patterns
- Implement proper accessibility
- Add loading states and error handling

## Future Enhancements

### Planned Features
- [ ] Real-time WebSocket updates
- [ ] Advanced threat intelligence
- [ ] Automated report generation
- [ ] Multi-language support
- [ ] Dark mode theme
- [ ] Mobile app support

### Technical Improvements
- [ ] Performance optimization
- [ ] Offline capabilities
- [ ] Advanced analytics
- [ ] Machine learning integration
- [ ] Cloud deployment options

## Support

For issues and questions:
1. Check this documentation
2. Review existing GitHub issues
3. Create new issue with detailed description
4. Include error logs and system information

---

**Note**: This GUI is a powerful addition to the API Security Scanner, providing an intuitive interface for managing security operations. The integration maintains full compatibility with existing CLI workflows while adding modern web-based capabilities.