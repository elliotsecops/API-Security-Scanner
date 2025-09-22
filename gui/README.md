# API Security Scanner GUI

This is the React-based graphical user interface for the API Security Scanner.

## Setup Instructions

### Prerequisites
- Node.js (v16 or higher)
- npm or yarn

### Development Setup

1. Navigate to the GUI directory:
```bash
cd gui
```

2. Install dependencies:
```bash
npm install
```

3. Start the development server:
```bash
npm start
```

The GUI will be available at `http://localhost:3000` and will proxy API requests to the Go backend running on `http://localhost:8080`.

### Production Build

1. Build the React application:
```bash
npm run build
```

2. The built files will be placed in the `build/` directory
3. The Go backend will automatically serve these files when placed in the `./gui/build` directory

## Features

- **Dashboard**: Real-time metrics and system health monitoring
- **Scanner**: Configure and run security scans
- **Results**: View and analyze scan results
- **Tenants**: Multi-tenant management
- **Settings**: Configuration management

## API Integration

The GUI integrates with the Go backend through the following API endpoints:

- `GET /api/system` - System metrics
- `GET /api/tenant` - Tenant-specific metrics
- `GET /api/scans` - Scan results
- `POST /api/auth/login` - Authentication
- `GET /api/tenants` - Tenant management
- `GET /api/export` - Export metrics data

## Authentication

Default credentials: `admin` / `admin`

## Development Notes

- The GUI uses Material-UI components for a consistent look and feel
- Chart.js is used for data visualization
- WebSocket support for real-time updates (planned)
- The application is a Single Page Application (SPA) with React Router