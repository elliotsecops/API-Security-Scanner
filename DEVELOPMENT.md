# API Security Scanner - Development Guide

## Overview

This guide provides comprehensive information for developers who want to contribute to the API Security Scanner project, extend its functionality, or understand its architecture and development workflows.

## Architecture

### Backend (Go)
The backend is built with Go 1.24+ and follows a modular architecture:

```
api-security-scanner/
├── main.go                 # Application entry point
├── config/                # Configuration management
├── scanner/               # Security scanning engine
├── auth/                  # Authentication and authorization
├── tenant/                # Multi-tenant management
├── siem/                  # SIEM integration
├── metrics/               # Metrics collection and dashboard
├── history/               # Historical data management
├── discovery/             # API discovery and crawling
├── logging/               # Logging infrastructure
└── types/                 # Common types and interfaces
```

### Frontend (React)
The frontend is a React-based SPA (Single Page Application):

```
gui/
├── src/
│   ├── components/        # Reusable UI components
│   ├── contexts/          # React contexts for state management
│   ├── App.js             # Main application component
│   └── index.js           # Application entry point
├── public/                # Static assets
└── package.json           # Dependencies and scripts
```

### Key Design Principles

1. **Separation of Concerns**: Clear separation between business logic, presentation, and data access
2. **Modularity**: Each feature is implemented as a separate module with well-defined interfaces
3. **Testability**: Code is written to be easily testable with comprehensive test coverage
4. **Scalability**: Architecture supports horizontal scaling and multi-tenant deployment
5. **Security**: Security is built into every layer of the application

## Development Setup

### Prerequisites
- Go 1.24+
- Node.js 16+
- npm or yarn
- Git
- Docker (optional)

### Backend Development

1. **Clone the Repository**
```bash
git clone https://github.com/your-username/api-security-scanner.git
cd api-security-scanner
```

2. **Set Up Go Environment**
```bash
# Download dependencies
go mod download
go mod tidy

# Verify installation
go version
go mod verify
```

3. **Build and Test**
```bash
# Build the application
go build -o api-security-scanner .

# Run tests
go test ./...

# Run tests with coverage
go test -cover ./...

# Run specific package tests
go test ./scanner/
```

### Frontend Development

1. **Navigate to GUI Directory**
```bash
cd gui
```

2. **Install Dependencies**
```bash
npm install
```

3. **Start Development Server**
```bash
npm start
```

4. **Build for Production**
```bash
npm run build
```

### Full Stack Development

For full-stack development with both frontend and backend:

```bash
# Terminal 1: Start backend
./api-security-scanner -dashboard

# Terminal 2: Start frontend
cd gui && npm start
```

## Development Workflows

### Backend Development Workflow

1. **Feature Development**
```bash
# Create feature branch
git checkout -b feature/new-feature-name

# Make changes
# Add tests for new functionality

# Run tests
go test ./...

# Build and test locally
go build -o api-security-scanner .
./api-security-scanner -scan
```

2. **Adding New Security Tests**
```go
// Example: Adding a new security test
package scanner

type NewSecurityTest struct {
    name string
    description string
    payloads []string
}

func (test *NewSecurityTest) Execute(endpoint APIEndpoint) TestResult {
    // Implementation
}
```

3. **Configuration Management**
```go
// Adding new configuration options
type Config struct {
    NewFeature NewFeatureConfig `yaml:"new_feature"`
}

type NewFeatureConfig struct {
    Enabled bool `yaml:"enabled"`
    Setting string `yaml:"setting"`
}
```

### Frontend Development Workflow

1. **Component Development**
```javascript
// Example: Creating a new component
import React from 'react';
import { Box, Card, Typography } from '@mui/material';

const NewComponent = ({ data }) => {
    return (
        <Card>
            <Box p={2}>
                <Typography variant="h6">
                    {data.title}
                </Typography>
            </Box>
        </Card>
    );
};

export default NewComponent;
```

2. **API Integration**
```javascript
// Example: API integration
import axios from 'axios';

const fetchScanResults = async () => {
    try {
        const response = await axios.get('/api/scans');
        return response.data;
    } catch (error) {
        console.error('Error fetching scan results:', error);
        throw error;
    }
};
```

3. **State Management**
```javascript
// Example: Using React Context
import React, { createContext, useContext, useState } from 'react';

const FeatureContext = createContext();

export const FeatureProvider = ({ children }) => {
    const [state, setState] = useState(null);

    const updateState = (newState) => {
        setState(newState);
    };

    return (
        <FeatureContext.Provider value={{ state, updateState }}>
            {children}
        </FeatureContext.Provider>
    );
};
```

### Testing

#### Backend Testing
```go
// Example: Unit test
package scanner

import "testing"

func TestNewSecurityTest(t *testing.T) {
    test := NewSecurityTest{
        name: "Test Security Test",
        payloads: []string{"test payload"},
    }

    result := test.Execute(APIEndpoint{
        URL: "https://example.com/api/test",
        Method: "GET",
    })

    if !result.Passed {
        t.Errorf("Expected test to pass")
    }
}
```

#### Frontend Testing
```javascript
// Example: Component test
import { render, screen } from '@testing-library/react';
import NewComponent from './NewComponent';

test('renders component with title', () => {
    const testData = { title: 'Test Title' };
    render(<NewComponent data={testData} />);

    expect(screen.getByText('Test Title')).toBeInTheDocument();
});
```

### Integration Testing

1. **API Contract Testing**
```bash
# Start test server
./api-security-scanner -dashboard -config test-config.yaml &

# Test API endpoints
curl http://localhost:8080/api/system
curl http://localhost:8080/api/scans
curl http://localhost:8080/api/tenants
```

2. **End-to-End Testing**
```javascript
// Example: E2E test with Cypress
describe('GUI Integration', () => {
    it('should display dashboard with metrics', () => {
        cy.visit('/');
        cy.get('[data-testid="dashboard"]').should('be.visible');
        cy.get('[data-testid="total-scans"]').should('be.visible');
    });
});
```

## Code Standards

### Go Code Standards

1. **Formatting**
```bash
# Format code
go fmt ./...

# Lint code
golint ./...

# Static analysis
go vet ./...
```

2. **Naming Conventions**
- Use `camelCase` for variable names
- Use `PascalCase` for exported names
- Use `snake_case` for configuration keys
- Use `UPPER_CASE` for constants

3. **Error Handling**
```go
// Good error handling
func processData(data string) error {
    if data == "" {
        return fmt.Errorf("empty data provided")
    }

    // Process data
    return nil
}

// Usage
if err := processData(input); err != nil {
    log.Printf("Error processing data: %v", err)
    return err
}
```

### JavaScript/React Code Standards

1. **ESLint Configuration**
```json
{
  "extends": [
    "react-app",
    "react-app/jest"
  ],
  "rules": {
    "semi": ["error", "always"],
    "quotes": ["error", "single"]
  }
}
```

2. **Component Structure**
```javascript
// Good component structure
import React from 'react';
import PropTypes from 'prop-types';
import { Box, Typography } from '@mui/material';

const MyComponent = ({ title, children, onClick }) => {
    const handleClick = () => {
        onClick();
    };

    return (
        <Box onClick={handleClick}>
            <Typography variant="h6">{title}</Typography>
            {children}
        </Box>
    );
};

MyComponent.propTypes = {
    title: PropTypes.string.isRequired,
    children: PropTypes.node,
    onClick: PropTypes.func
};

MyComponent.defaultProps = {
    children: null,
    onClick: () => {}
};

export default MyComponent;
```

### Documentation Standards

1. **Code Documentation**
```go
// Good Go documentation
// ProcessData processes the input data and returns the result.
// It validates the input format and applies security checks.
// Returns processed data or error if validation fails.
func ProcessData(data string) (string, error) {
    // Implementation
}
```

```javascript
// Good JavaScript documentation
/**
 * Processes the input data and returns the result.
 * @param {string} data - The input data to process
 * @returns {Promise<string>} The processed data
 * @throws {Error} If validation fails
 */
const processData = async (data) => {
    // Implementation
};
```

## Performance Optimization

### Backend Optimization

1. **Concurrent Processing**
```go
// Use goroutines for concurrent processing
func processEndpoints(endpoints []APIEndpoint) []Result {
    results := make([]Result, len(endpoints))
    var wg sync.WaitGroup
    wg.Add(len(endpoints))

    for i, endpoint := range endpoints {
        go func(idx int, ep APIEndpoint) {
            defer wg.Done()
            results[idx] = scanEndpoint(ep)
        }(i, endpoint)
    }

    wg.Wait()
    return results
}
```

2. **Memory Management**
```go
// Use buffers and object pooling
var bufferPool = sync.Pool{
    New: func() interface{} {
        return new(bytes.Buffer)
    },
}

func processData(data []byte) {
    buf := bufferPool.Get().(*bytes.Buffer)
    defer bufferPool.Put(buf)

    buf.Reset()
    buf.Write(data)
    // Process data
}
```

### Frontend Optimization

1. **Component Optimization**
```javascript
// Use React.memo for component memoization
const OptimizedComponent = React.memo(({ data }) => {
    return <div>{data.value}</div>;
});
```

2. **State Management**
```javascript
// Use useMemo and useCallback for performance
const MyComponent = ({ data }) => {
    const processedData = useMemo(() => {
        return expensiveProcessing(data);
    }, [data]);

    const handleClick = useCallback(() => {
        // Handle click
    }, []);

    return <div onClick={handleClick}>{processedData}</div>;
};
```

## Security Considerations

### Backend Security

1. **Input Validation**
```go
// Always validate input
func validateInput(input string) error {
    if len(input) > 1000 {
        return fmt.Errorf("input too long")
    }

    // Add more validation
    return nil
}
```

2. **SQL Injection Prevention**
```go
// Use parameterized queries
func getUserByID(id string) (*User, error) {
    query := "SELECT * FROM users WHERE id = ?"
    row := db.QueryRow(query, id)

    var user User
    err := row.Scan(&user.ID, &user.Name)
    if err != nil {
        return nil, err
    }

    return &user, nil
}
```

### Frontend Security

1. **XSS Prevention**
```javascript
// Use React's built-in XSS protection
const SafeComponent = ({ content }) => {
    return <div>{content}</div>; // React escapes content by default
};
```

2. **Authentication**
```javascript
// Secure token storage
const login = async (credentials) => {
    const response = await axios.post('/api/auth/login', credentials);
    localStorage.setItem('token', response.data.token);
    return response.data;
};
```

## Deployment

### Development Deployment

1. **Local Development**
```bash
# Start backend
./api-security-scanner -dashboard

# Start frontend
cd gui && npm start
```

2. **Docker Development**
```dockerfile
FROM golang:1.24-alpine AS backend
WORKDIR /app
COPY . .
RUN go build -o api-security-scanner .

FROM node:16-alpine AS frontend
WORKDIR /app/gui
COPY gui/package*.json ./
RUN npm install
COPY gui/ ./
RUN npm run build

FROM alpine:latest
COPY --from=backend /app/api-security-scanner .
COPY --from=frontend /app/gui/build ./gui/build
CMD ["./api-security-scanner", "--dashboard"]
```

### Production Deployment

1. **Build for Production**
```bash
# Build frontend
cd gui && npm run build

# Build backend
go build -o api-security-scanner .
```

2. **Docker Production**
```bash
# Build Docker image
docker build -t api-security-scanner .

# Run container
docker run -p 8080:8080 api-security-scanner
```

## Contributing

### Pull Request Process

1. **Fork the Repository**
```bash
git clone https://github.com/your-username/api-security-scanner.git
```

2. **Create Feature Branch**
```bash
git checkout -b feature/new-feature
```

3. **Make Changes and Test**
```bash
# Run tests
go test ./...
cd gui && npm test

# Build and test
go build -o api-security-scanner .
./api-security-scanner -scan
```

4. **Commit Changes**
```bash
git add .
git commit -m "Add new feature"
git push origin feature/new-feature
```

5. **Create Pull Request**
- Provide clear description
- Include test results
- Document breaking changes

### Issue Reporting

1. **Bug Reports**
- Use GitHub Issues
- Provide detailed steps to reproduce
- Include error logs and system information
- Add expected vs actual behavior

2. **Feature Requests**
- Describe the feature clearly
- Explain the use case
- Provide implementation suggestions

### Code Review Guidelines

1. **Review Checklist**
- [ ] Code follows project standards
- [ ] Tests are included and passing
- [ ] Documentation is updated
- [ ] No breaking changes (or clearly documented)
- [ ] Security considerations addressed

2. **Review Process**
- Assign reviewers based on expertise
- Address all review comments
- Ensure CI/CD pipeline passes
- Get approval before merging

## Troubleshooting

### Common Issues

1. **Build Failures**
```bash
# Clean and rebuild
go clean -modcache
go mod download
go build -o api-security-scanner .
```

2. **Test Failures**
```bash
# Run specific test with verbose output
go test -v ./scanner/

# Check test coverage
go test -cover ./...
```

3. **GUI Development Issues**
```bash
# Clear node_modules and reinstall
rm -rf node_modules package-lock.json
npm install
```

### Debug Mode

```bash
# Enable debug logging
export DEBUG=api-security-scanner:*

# Start with verbose output
./api-security-scanner -dashboard -log-level debug
```

## Resources

### Documentation
- [Main Documentation](./DOCUMENTATION.md)
- [Configuration Guide](./CONFIGURATION.md)
- [Installation Guide](./INSTALL.md)
- [GUI Guide](./GUIDE.md)

### Tools and Dependencies
- [Go Documentation](https://golang.org/doc/)
- [React Documentation](https://reactjs.org/docs/)
- [Material-UI Documentation](https://mui.com/material-ui/)
- [Chart.js Documentation](https://www.chartjs.org/docs/)

### Community
- GitHub Issues
- GitHub Discussions
- Stack Overflow (use appropriate tags)

---

This development guide provides comprehensive information for contributing to the API Security Scanner project. Following these guidelines ensures high-quality contributions and maintainable code.