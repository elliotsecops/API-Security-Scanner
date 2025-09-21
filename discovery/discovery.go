package discovery

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"

	"github.com/antchfx/htmlquery"
	"github.com/antchfx/xpath"
	"golang.org/x/net/html"

	"api-security-scanner/types"
	"api-security-scanner/logging"
)

// DiscoveryConfig represents API discovery configuration
type DiscoveryConfig struct {
	Enabled          bool     `yaml:"enabled"`
	MaxDepth         int      `yaml:"max_depth"`
	FollowLinks      bool     `yaml:"follow_links"`
	DiscoverParams   bool     `yaml:"discover_params"`
	UserAgent        string   `yaml:"user_agent"`
	ExcludePatterns  []string `yaml:"exclude_patterns"`
}

// APIDiscovery handles API endpoint discovery and crawling
type APIDiscovery struct {
	config       DiscoveryConfig
	visited      map[string]bool
	discovered   []types.APIEndpoint
	mutex        sync.RWMutex
	client       *http.Client
}

// NewAPIDiscovery creates a new API discovery instance
func NewAPIDiscovery(config DiscoveryConfig) *APIDiscovery {
	if config.MaxDepth <= 0 {
		config.MaxDepth = 3
	}
	if config.UserAgent == "" {
		config.UserAgent = "API-Security-Scanner-Discovery/1.0"
	}

	return &APIDiscovery{
		config:  config,
		visited: make(map[string]bool),
		client: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

// DiscoverEndpoints discovers API endpoints from a base URL
func (d *APIDiscovery) DiscoverEndpoints(baseURL string) ([]types.APIEndpoint, error) {
	if !d.config.Enabled {
		return nil, nil
	}

	logging.Info("Starting API discovery", map[string]interface{}{
		"base_url":   baseURL,
		"max_depth":  d.config.MaxDepth,
		"follow_links": d.config.FollowLinks,
	})

	// Start discovery from base URL
	err := d.crawl(baseURL, 0)
	if err != nil {
		return nil, fmt.Errorf("discovery failed: %v", err)
	}

	logging.Info("API discovery completed", map[string]interface{}{
		"endpoints_discovered": len(d.discovered),
		"urls_visited":         len(d.visited),
	})

	return d.discovered, nil
}

// crawl recursively crawls URLs to discover API endpoints
func (d *APIDiscovery) crawl(currentURL string, depth int) error {
	// Check if we've already visited this URL
	d.mutex.RLock()
	if d.visited[currentURL] {
		d.mutex.RUnlock()
		return nil
	}
	d.mutex.RUnlock()

	// Check if we've exceeded max depth
	if depth > d.config.MaxDepth {
		return nil
	}

	// Check if URL matches any exclude patterns
	for _, pattern := range d.config.ExcludePatterns {
		if strings.Contains(currentURL, pattern) {
			logging.Debug("Skipping excluded URL", map[string]interface{}{
				"url":     currentURL,
				"pattern": pattern,
			})
			return nil
		}
	}

	// Mark URL as visited
	d.mutex.Lock()
	d.visited[currentURL] = true
	d.mutex.Unlock()

	logging.Debug("Crawling URL", map[string]interface{}{
		"url":   currentURL,
		"depth": depth,
	})

	// Make HTTP request
	req, err := http.NewRequest("GET", currentURL, nil)
	if err != nil {
		return fmt.Errorf("failed to create request for %s: %v", currentURL, err)
	}

	req.Header.Set("User-Agent", d.config.UserAgent)

	resp, err := d.client.Do(req)
	if err != nil {
		logging.Warn("Failed to crawl URL", map[string]interface{}{
			"url":   currentURL,
			"error": err.Error(),
		})
		return nil
	}
	defer resp.Body.Close()

	// Only process successful responses
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		logging.Debug("Skipping non-successful response", map[string]interface{}{
			"url":    currentURL,
			"status": resp.StatusCode,
		})
		return nil
	}

	// Check if this is an API endpoint
	if d.isAPIEndpoint(currentURL, resp) {
		d.mutex.Lock()
		d.discovered = append(d.discovered, types.APIEndpoint{
			URL:    currentURL,
			Method: "GET",
		})
		d.mutex.Unlock()

		// Test other HTTP methods for API endpoints
		methods := []string{"POST", "PUT", "DELETE", "PATCH", "OPTIONS"}
		for _, method := range methods {
			if d.testHTTPMethod(currentURL, method) {
				d.mutex.Lock()
				d.discovered = append(d.discovered, types.APIEndpoint{
					URL:    currentURL,
					Method: method,
				})
				d.mutex.Unlock()
			}
		}
	}

	// If follow_links is enabled and this is HTML, parse for links
	if d.config.FollowLinks && strings.Contains(resp.Header.Get("Content-Type"), "text/html") {
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			logging.Warn("Failed to read response body", map[string]interface{}{
				"url":   currentURL,
				"error": err.Error(),
			})
		} else {
			doc, err := html.Parse(strings.NewReader(string(body)))
			if err != nil {
				logging.Warn("Failed to parse HTML", map[string]interface{}{
					"url":   currentURL,
					"error": err.Error(),
				})
			} else {
				links, err := d.extractLinks(doc, currentURL)
				if err != nil {
					logging.Warn("Failed to extract links", map[string]interface{}{
						"url":   currentURL,
						"error": err.Error(),
					})
				} else {
					// Crawl discovered links
					for _, link := range links {
						if err := d.crawl(link, depth+1); err != nil {
							logging.Warn("Failed to crawl discovered link", map[string]interface{}{
								"url":   link,
								"error": err.Error(),
							})
						}
					}
				}
			}
		}
	}

	return nil
}

// isAPIEndpoint determines if a URL is an API endpoint
func (d *APIDiscovery) isAPIEndpoint(urlStr string, resp *http.Response) bool {
	// Check URL patterns
	apiPatterns := []string{
		"/api/",
		"/v1/",
		"/v2/",
		"/rest/",
		"/services/",
		"/graphql",
		"/swagger",
		"/openapi",
	}

	for _, pattern := range apiPatterns {
		if strings.Contains(strings.ToLower(urlStr), pattern) {
			return true
		}
	}

	// Check content type
	contentType := resp.Header.Get("Content-Type")
	apiContentTypes := []string{
		"application/json",
		"application/xml",
		"application/hal+json",
		"application/vnd.api+json",
		"text/xml",
	}

	for _, apiType := range apiContentTypes {
		if strings.Contains(contentType, apiType) {
			return true
		}
	}

	// Check response body content (for APIs that don't set proper content types)
	if d.config.DiscoverParams {
		// This would require reading the body, which might be expensive
		// For now, we'll rely on URL patterns and content types
	}

	return false
}

// testHTTPMethod tests if an HTTP method is supported for a URL
func (d *APIDiscovery) testHTTPMethod(urlStr, method string) bool {
	req, err := http.NewRequest(method, urlStr, nil)
	if err != nil {
		return false
	}

	req.Header.Set("User-Agent", d.config.UserAgent)

	resp, err := d.client.Do(req)
	if err != nil {
		return false
	}
	defer resp.Body.Close()

	// Consider method supported if we don't get a 405 Method Not Allowed
	return resp.StatusCode != http.StatusMethodNotAllowed
}

// extractLinks extracts links from HTML content
func (d *APIDiscovery) extractLinks(body *html.Node, baseURL string) ([]string, error) {
	var links []string

	// Use the parsed HTML document directly
	doc := body

	// Use xpath to find all links
	xpathExpr := xpath.MustCompile("//a/@href|//link/@href|//script/@src|//img/@src")
	nodes := htmlquery.QuerySelectorAll(doc, xpathExpr)

	for _, node := range nodes {
		if href := htmlquery.SelectAttr(node, "href"); href != "" {
			if absoluteURL := d.makeAbsoluteURL(href, baseURL); absoluteURL != "" {
				links = append(links, absoluteURL)
			}
		}
		if src := htmlquery.SelectAttr(node, "src"); src != "" {
			if absoluteURL := d.makeAbsoluteURL(src, baseURL); absoluteURL != "" {
				links = append(links, absoluteURL)
			}
		}
	}

	return d.filterLinks(links), nil
}

// makeAbsoluteURL converts relative URLs to absolute URLs
func (d *APIDiscovery) makeAbsoluteURL(href, baseURL string) string {
	if strings.HasPrefix(href, "http://") || strings.HasPrefix(href, "https://") {
		return href
	}

	if strings.HasPrefix(href, "/") {
		// Relative to domain
		parsedURL, err := url.Parse(baseURL)
		if err != nil {
			return ""
		}
		return parsedURL.Scheme + "://" + parsedURL.Host + href
	}

	// Relative to current path
	parsedURL, err := url.Parse(baseURL)
	if err != nil {
		return ""
	}

	basePath := parsedURL.Path
	if !strings.HasSuffix(basePath, "/") {
		basePath = basePath[:strings.LastIndex(basePath, "/")+1]
	}

	return parsedURL.Scheme + "://" + parsedURL.Host + basePath + href
}

// filterLinks filters and deduplicates links
func (d *APIDiscovery) filterLinks(links []string) []string {
	seen := make(map[string]bool)
	var filtered []string

	for _, link := range links {
		// Skip non-http URLs
		if !strings.HasPrefix(link, "http://") && !strings.HasPrefix(link, "https://") {
			continue
		}

		// Skip mailto, tel, etc.
		if strings.Contains(link, "mailto:") || strings.Contains(link, "tel:") {
			continue
		}

		// Skip file extensions that are not API-related
		skipExtensions := []string{".css", ".js", ".png", ".jpg", ".jpeg", ".gif", ".pdf", ".zip"}
		skip := false
		for _, ext := range skipExtensions {
			if strings.HasSuffix(strings.ToLower(link), ext) {
				skip = true
				break
			}
		}
		if skip {
			continue
		}

		// Remove fragments
		if idx := strings.Index(link, "#"); idx != -1 {
			link = link[:idx]
		}

		// Remove query parameters for deduplication
		if idx := strings.Index(link, "?"); idx != -1 {
			link = link[:idx]
		}

		if !seen[link] {
			seen[link] = true
			filtered = append(filtered, link)
		}
	}

	return filtered
}

// DiscoverParameters discovers parameters from API responses
func (d *APIDiscovery) DiscoverParameters(endpoint types.APIEndpoint) ([]string, error) {
	if !d.config.DiscoverParams {
		return nil, nil
	}

	var parameters []string

	// Test the endpoint to see what parameters it might accept
	testParams := []string{
		"?test=value",
		"?id=1",
		"?page=1",
		"?limit=10",
		"?search=test",
		"?format=json",
	}

	for _, param := range testParams {
		testURL := endpoint.URL + param
		req, err := http.NewRequest("GET", testURL, nil)
		if err != nil {
			continue
		}

		req.Header.Set("User-Agent", d.config.UserAgent)

		resp, err := d.client.Do(req)
		if err != nil {
			continue
		}
		defer resp.Body.Close()

		// If the response is successful, the parameter might be valid
		if resp.StatusCode >= 200 && resp.StatusCode < 300 {
			parameters = append(parameters, param[1:]) // Remove the '?'
		}
	}

	return parameters, nil
}