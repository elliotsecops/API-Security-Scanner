package openapi

import (
	"fmt"
	"strings"

	"github.com/getkin/kin-openapi/openapi3"
	"api-security-scanner/types"
	"api-security-scanner/logging"
)

// OpenAPIIntegration handles OpenAPI specification parsing and testing
type OpenAPIIntegration struct {
	spec *openapi3.T
}

// NewOpenAPIIntegration creates a new OpenAPI integration instance
func NewOpenAPIIntegration(specPath string) (*OpenAPIIntegration, error) {
	loader := openapi3.NewLoader()
	doc, err := loader.LoadFromFile(specPath)
	if err != nil {
		return nil, fmt.Errorf("failed to load OpenAPI spec: %v", err)
	}

	if err := doc.Validate(loader.Context); err != nil {
		return nil, fmt.Errorf("OpenAPI spec validation failed: %v", err)
	}

	return &OpenAPIIntegration{
		spec: doc,
	}, nil
}

// GenerateEndpointsFromSpec generates API endpoints from OpenAPI specification
func (o *OpenAPIIntegration) GenerateEndpointsFromSpec() []types.APIEndpoint {
	var endpoints []types.APIEndpoint

	baseURL := o.spec.Servers[0].URL

	for path, pathItem := range o.spec.Paths.Map() {
		if pathItem.Get != nil {
			endpoints = append(endpoints, types.APIEndpoint{
				URL:    baseURL + path,
				Method: "GET",
			})
		}
		if pathItem.Post != nil {
			endpoints = append(endpoints, types.APIEndpoint{
				URL:    baseURL + path,
				Method: "POST",
			})
		}
		if pathItem.Put != nil {
			endpoints = append(endpoints, types.APIEndpoint{
				URL:    baseURL + path,
				Method: "PUT",
			})
		}
		if pathItem.Delete != nil {
			endpoints = append(endpoints, types.APIEndpoint{
				URL:    baseURL + path,
				Method: "DELETE",
			})
		}
		if pathItem.Patch != nil {
			endpoints = append(endpoints, types.APIEndpoint{
				URL:    baseURL + path,
				Method: "PATCH",
			})
		}
	}

	logging.Info("Generated endpoints from OpenAPI spec", map[string]interface{}{
		"endpoints_count": len(endpoints),
		"base_url": baseURL,
	})

	return endpoints
}

// ValidateEndpointAgainstSpec validates an endpoint against OpenAPI specification
func (o *OpenAPIIntegration) ValidateEndpointAgainstSpec(endpoint types.APIEndpoint) error {
	// Extract path from URL
	path := o.extractPathFromURL(endpoint.URL)

	pathItem := o.spec.Paths.Find(path)
	if pathItem == nil {
		return fmt.Errorf("path %s not found in OpenAPI spec", path)
	}

	// Check if method is defined in spec
	switch endpoint.Method {
	case "GET":
		if pathItem.Get == nil {
			return fmt.Errorf("GET method not defined for path %s", path)
		}
	case "POST":
		if pathItem.Post == nil {
			return fmt.Errorf("POST method not defined for path %s", path)
		}
	case "PUT":
		if pathItem.Put == nil {
			return fmt.Errorf("PUT method not defined for path %s", path)
		}
	case "DELETE":
		if pathItem.Delete == nil {
			return fmt.Errorf("DELETE method not defined for path %s", path)
		}
	case "PATCH":
		if pathItem.Patch == nil {
			return fmt.Errorf("PATCH method not defined for path %s", path)
		}
	default:
		return fmt.Errorf("unsupported HTTP method: %s", endpoint.Method)
	}

	return nil
}

// GenerateTestCasesFromSpec generates test cases based on OpenAPI specification
func (o *OpenAPIIntegration) GenerateTestCasesFromSpec(endpoint types.APIEndpoint) []types.APIEndpoint {
	var testCases []types.APIEndpoint

	path := o.extractPathFromURL(endpoint.URL)
	pathItem := o.spec.Paths.Find(path)
	if pathItem == nil {
		return testCases
	}

	var operation *openapi3.Operation
	switch endpoint.Method {
	case "GET":
		operation = pathItem.Get
	case "POST":
		operation = pathItem.Post
	case "PUT":
		operation = pathItem.Put
	case "DELETE":
		operation = pathItem.Delete
	case "PATCH":
		operation = pathItem.Patch
	default:
		return testCases
	}

	if operation == nil {
		return testCases
	}

	// Generate test cases based on parameters
	for _, parameter := range operation.Parameters {
		if parameter.Value.In == "query" {
			// Test with malformed query parameters
			testEndpoint := endpoint
			testEndpoint.URL = endpoint.URL + "?" + parameter.Value.Name + "=' OR '1'='1"
			testCases = append(testCases, testEndpoint)
		}
	}

	// Test with different content types if requestBody is defined
	if operation.RequestBody != nil {
		for contentType := range operation.RequestBody.Value.Content {
			if strings.Contains(contentType, "json") {
				// Test with malformed JSON
				testEndpoint := endpoint
				testEndpoint.Body = "{"
				testCases = append(testCases, testEndpoint)

				// Test with injection payload
				testEndpoint = endpoint
				testEndpoint.Body = fmt.Sprintf(`{"test": "' OR '1'='1"}`)
				testCases = append(testCases, testEndpoint)
			}
		}
	}

	return testCases
}

// extractPathFromURL extracts the API path from a full URL
func (o *OpenAPIIntegration) extractPathFromURL(url string) string {
	// Remove base URL to get the path
	baseURL := o.spec.Servers[0].URL
	if strings.HasPrefix(url, baseURL) {
		return url[len(baseURL):]
	}

	// If no base URL match, try to extract path manually
	if strings.HasPrefix(url, "http://") {
		url = url[7:]
	} else if strings.HasPrefix(url, "https://") {
		url = url[8:]
	}

	// Find the first / after the domain
	if idx := strings.Index(url, "/"); idx != -1 {
		return url[idx:]
	}

	return "/"
}

// GetSpecInfo returns information about the OpenAPI specification
func (o *OpenAPIIntegration) GetSpecInfo() map[string]interface{} {
	info := map[string]interface{}{
		"title":       o.spec.Info.Title,
		"version":     o.spec.Info.Version,
		"description": o.spec.Info.Description,
		"paths_count": len(o.spec.Paths.Map()),
		"servers":     []string{},
	}

	for _, server := range o.spec.Servers {
		info["servers"] = append(info["servers"].([]string), server.URL)
	}

	return info
}