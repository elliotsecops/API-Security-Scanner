package config

import (
	"api-security-scanner/metrics"
	"api-security-scanner/scanner"
	"api-security-scanner/types"
	"fmt"
	"os"
	"time"

	"gopkg.in/yaml.v3"
)

// TenantConfig represents tenant configuration
type TenantConfig struct {
	ID           string                 `json:"id" yaml:"id"`
	Name         string                 `json:"name" yaml:"name"`
	Description  string                 `json:"description" yaml:"description"`
	IsActive     bool                   `json:"is_active" yaml:"is_active"`
	Settings     TenantSettings         `json:"settings" yaml:"settings"`
}

// TenantSettings represents tenant-specific settings
type TenantSettings struct {
	DataIsolation  DataIsolationConfig  `yaml:"data_isolation"`
	ResourceLimits ResourceLimits       `yaml:"resource_limits"`
	Notifications  NotificationSettings  `yaml:"notifications"`
	CustomBranding CustomBrandingConfig `yaml:"custom_branding"`
	SIEMIntegration SIEMIntegrationConfig `yaml:"siem_integration"`
}

// DataIsolationConfig represents data isolation configuration
type DataIsolationConfig struct {
	Enabled          bool    `yaml:"enabled"`
	StoragePath      string  `yaml:"storage_path"`
	EncryptionEnabled bool    `yaml:"encryption_enabled"`
	RetentionDays    int     `yaml:"retention_days"`
}

// ResourceLimits represents resource limits
type ResourceLimits struct {
	MaxConcurrentScans  int       `yaml:"max_concurrent_scans"`
	MaxEndpointsPerScan int       `yaml:"max_endpoints_per_scan"`
	MaxStorageMB       int       `yaml:"max_storage_mb"`
	ScanQuota          int       `yaml:"scan_quota_monthly"`
	RateLimit          RateLimit `yaml:"rate_limit"`
}

// RateLimit represents rate limiting configuration
type RateLimit struct {
	RequestsPerSecond int `yaml:"requests_per_second"`
	BurstSize         int `yaml:"burst_size"`
}

// NotificationSettings represents notification settings
type NotificationSettings struct {
	Email   EmailNotificationConfig   `yaml:"email"`
	Webhook WebhookNotificationConfig `yaml:"webhook"`
	Slack   SlackNotificationConfig   `yaml:"slack"`
}

// EmailNotificationConfig represents email notification configuration
type EmailNotificationConfig struct {
	Enabled    bool     `yaml:"enabled"`
	SMTPHost   string   `yaml:"smtp_host"`
	SMTPPort   int      `yaml:"smtp_port"`
	Username   string   `yaml:"username"`
	Password   string   `yaml:"password"`
	FromEmail  string   `yaml:"from_email"`
	Recipients []string `yaml:"recipients"`
}

// WebhookNotificationConfig represents webhook notification configuration
type WebhookNotificationConfig struct {
	Enabled bool                   `yaml:"enabled"`
	URL     string                 `yaml:"url"`
	Method  string                 `yaml:"method"`
	Headers map[string]string      `yaml:"headers"`
	Secret  string                 `yaml:"secret"`
}

// SlackNotificationConfig represents Slack notification configuration
type SlackNotificationConfig struct {
	Enabled    bool   `yaml:"enabled"`
	WebhookURL string `yaml:"webhook_url"`
	Channel    string `yaml:"channel"`
	Username   string `yaml:"username"`
}

// CustomBrandingConfig represents custom branding configuration
type CustomBrandingConfig struct {
	Enabled      bool   `yaml:"enabled"`
	LogoURL      string `yaml:"logo_url"`
	PrimaryColor string `yaml:"primary_color"`
	SecondaryColor string `yaml:"secondary_color"`
	CompanyName  string `yaml:"company_name"`
	CustomFooter string `yaml:"custom_footer"`
}

// SIEMIntegrationConfig represents SIEM integration configuration
type SIEMIntegrationConfig struct {
	Enabled     bool                   `yaml:"enabled"`
	Type        string                 `yaml:"type"`
	Config      map[string]interface{} `yaml:"config"`
	Format      string                 `yaml:"format"`
	EndpointURL string                 `yaml:"endpoint_url"`
	AuthToken   string                 `yaml:"auth_token"`
}

// SIEMConfig represents SIEM configuration
type SIEMConfig struct {
	Enabled     bool                   `yaml:"enabled"`
	Type        string                 `yaml:"type"`
	Config      map[string]interface{} `yaml:"config"`
	Format      string                 `yaml:"format"`
	EndpointURL string                 `yaml:"endpoint_url"`
	AuthToken   string                 `yaml:"auth_token"`
}

// Config represents the complete application configuration
type Config struct {
	Scanner      *scanner.Config     `yaml:"scanner"`
	Tenant       *TenantConfig        `yaml:"tenant"`
	Metrics      *metrics.Dashboard   `yaml:"metrics"`
	SIEM         *SIEMConfig          `yaml:"siem"`
	Auth         struct {
		Enabled     bool               `yaml:"enabled"`
		Type        string             `yaml:"type"`
		Config      map[string]string  `yaml:"config"`
	} `yaml:"auth"`
	Server       struct {
		Port        int                `yaml:"port"`
		Host        string             `yaml:"host"`
	} `yaml:"server"`
}

// Load loads the configuration from a YAML file
func Load(filename string) (*Config, error) {
	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, err
	}

	var config Config
	err = yaml.Unmarshal(data, &config)
	if err != nil {
		return nil, err
	}

	// Validate the configuration
	if err := Validate(&config); err != nil {
		return nil, fmt.Errorf("configuration validation failed: %w", err)
	}

	return &config, nil
}

// Validate checks if the configuration is valid
func Validate(config *Config) error {
	// Validate scanner configuration
	if config.Scanner == nil {
		return fmt.Errorf("scanner configuration is required")
	}

	// Check if we have at least one API endpoint
	if len(config.Scanner.APIEndpoints) == 0 {
		return fmt.Errorf("at least one API endpoint is required")
	}

	// Validate each endpoint
	for i, endpoint := range config.Scanner.APIEndpoints {
		if endpoint.URL == "" {
			return fmt.Errorf("endpoint %d: URL is required", i)
		}
		if endpoint.Method == "" {
			return fmt.Errorf("endpoint %d: HTTP method is required", i)
		}
		// Validate HTTP method is one of the standard methods
		validMethods := map[string]bool{
			"GET": true, "POST": true, "PUT": true, "DELETE": true,
			"PATCH": true, "HEAD": true, "OPTIONS": true, "TRACE": true,
		}
		if !validMethods[endpoint.Method] {
			return fmt.Errorf("endpoint %d: invalid HTTP method '%s'", i, endpoint.Method)
		}
	}

	// Check if we have at least one injection payload
	if len(config.Scanner.InjectionPayloads) == 0 {
		return fmt.Errorf("at least one injection payload is required")
	}

	// If auth is provided, both username and password are required
	if (config.Scanner.Auth.Username != "" && config.Scanner.Auth.Password == "") ||
	   (config.Scanner.Auth.Username == "" && config.Scanner.Auth.Password != "") {
		return fmt.Errorf("both username and password are required for authentication, or neither")
	}

	// Set default XSS payloads if none provided
	if len(config.Scanner.XSSPayloads) == 0 {
		config.Scanner.XSSPayloads = []string{
			"<script>alert('XSS')</script>",
			"'><script>alert('XSS')</script>",
			"<img src=x onerror=alert('XSS')>",
			"javascript:alert('XSS')",
		}
	}

	// Set default NoSQL payloads if none provided
	if len(config.Scanner.NoSQLPayloads) == 0 {
		config.Scanner.NoSQLPayloads = []string{
			"{$ne: null}",
			"{$gt: ''}",
			"{$or: [1,1]}",
			"{$where: 'sleep(100)'}",
			"{$regex: '.*'}",
			"{$exists: true}",
			"{$in: [1,2,3]}",
		}
	}

	// Set default metrics configuration
	if config.Metrics == nil {
		config.Metrics = &metrics.Dashboard{
			Enabled:       true,
			Port:          8080,
			UpdateInterval: 30 * time.Second,
			RetentionDays: 30,
			Charts: []metrics.DashboardChart{
				{
					Title:      "Vulnerability Overview",
					Type:      "pie",
					Metrics:    []string{"critical", "high", "medium", "low"},
					TimeRange:  "24h",
					Refresh:    30,
				},
				{
					Title:      "Performance Metrics",
					Type:      "line",
					Metrics:    []string{"response_time", "throughput"},
					TimeRange:  "24h",
					Refresh:    30,
				},
			},
		}
	}

	// Set default server configuration
	if config.Server.Port == 0 {
		config.Server.Port = 8081
	}
	if config.Server.Host == "" {
		config.Server.Host = "localhost"
	}

	return nil
}

// TestConfigReachability tests if the configured endpoints are reachable
func TestConfigReachability(config *scanner.Config) []string {
	var issues []string
	// In a real implementation, we would test endpoint reachability here
	// For now, we'll just return an empty slice
	return issues
}

// CreateDefaultConfig creates a default configuration for Phase 4
func CreateDefaultConfig() *Config {
	return &Config{
		Scanner: &scanner.Config{
			APIEndpoints: []types.APIEndpoint{
				{
					URL:    "https://api.example.com/users",
					Method: "GET",
				},
				{
					URL:    "https://api.example.com/users",
					Method: "POST",
					Body:   `{"name":"test","email":"test@example.com"}`,
				},
			},
			InjectionPayloads: []string{
				"' OR '1'='1",
				"' OR 1=1--",
				"' UNION SELECT NULL--",
				"1; DROP TABLE users--",
				"' AND (SELECT COUNT(*) FROM information_schema.tables)>0--",
			},
			Auth: scanner.Auth{
				Username: "",
				Password: "",
			},
			RateLimiting: scanner.RateLimiting{
				RequestsPerSecond:     10,
				MaxConcurrentRequests: 5,
			},
		},
		Tenant: &TenantConfig{
			ID:          "default",
			Name:        "Default Organization",
			Description: "Default tenant for single-tenant use",
			Settings: TenantSettings{
				DataIsolation: DataIsolationConfig{
					Enabled:           false,
					StoragePath:       "./history",
					EncryptionEnabled: false,
					RetentionDays:     30,
				},
				ResourceLimits: ResourceLimits{
					MaxConcurrentScans:  10,
					MaxEndpointsPerScan: 1000,
					MaxStorageMB:       10240,
					ScanQuota:          1000,
					RateLimit: RateLimit{
						RequestsPerSecond: 100,
						BurstSize:         50,
					},
				},
				Notifications: NotificationSettings{
					Email: EmailNotificationConfig{
						Enabled: false,
					},
					Webhook: WebhookNotificationConfig{
						Enabled: false,
					},
					Slack: SlackNotificationConfig{
						Enabled: false,
					},
				},
				CustomBranding: CustomBrandingConfig{
					Enabled: false,
				},
				SIEMIntegration: SIEMIntegrationConfig{
					Enabled: false,
				},
			},
			IsActive: true,
		},
		Metrics: &metrics.Dashboard{
			Enabled:       true,
			Port:          8080,
			UpdateInterval: 30 * time.Second,
			RetentionDays: 30,
			Charts: []metrics.DashboardChart{
				{
					Title:      "Vulnerability Overview",
					Type:      "pie",
					Metrics:    []string{"critical", "high", "medium", "low"},
					TimeRange:  "24h",
					Refresh:    30,
				},
				{
					Title:      "Performance Metrics",
					Type:      "line",
					Metrics:    []string{"response_time", "throughput"},
					TimeRange:  "24h",
					Refresh:    30,
				},
			},
		},
		SIEM: &SIEMConfig{
			Enabled: false,
			Type:    "syslog",
		},
		Auth: struct {
			Enabled bool              `yaml:"enabled"`
			Type    string            `yaml:"type"`
			Config  map[string]string `yaml:"config"`
		}{
			Enabled: false,
		},
		Server: struct {
			Port int    `yaml:"port"`
			Host string `yaml:"host"`
		}{
			Port: 8081,
			Host: "localhost",
		},
	}
}
