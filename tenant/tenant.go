package tenant

import (
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	"gopkg.in/yaml.v3"

	"api-security-scanner/logging"
	"api-security-scanner/scanner"
	"api-security-scanner/types"
)

// Tenant represents a multi-tenant organization
type Tenant struct {
	ID           string                 `json:"id" yaml:"id"`
	Name         string                 `json:"name" yaml:"name"`
	Description  string                 `json:"description" yaml:"description"`
	Config       *scanner.Config        `json:"config" yaml:"config"`
	Settings     TenantSettings         `json:"settings" yaml:"settings"`
	CreatedAt    time.Time              `json:"created_at"`
	UpdatedAt    time.Time              `json:"updated_at"`
	IsActive     bool                   `json:"is_active" yaml:"is_active"`
}

// TenantSettings represents tenant-specific settings
type TenantSettings struct {
	DataIsolation    DataIsolationConfig   `yaml:"data_isolation"`
	ResourceLimits   ResourceLimits       `yaml:"resource_limits"`
	Notifications    NotificationConfig   `yaml:"notifications"`
	CustomBranding   CustomBrandingConfig `yaml:"custom_branding"`
	SIEMIntegration SIEMConfig          `yaml:"siem_integration"`
}

// DataIsolationConfig defines data isolation settings
type DataIsolationConfig struct {
	Enabled           bool   `yaml:"enabled"`
	StoragePath       string `yaml:"storage_path"`
	EncryptionEnabled bool   `yaml:"encryption_enabled"`
	RetentionDays     int    `yaml:"retention_days"`
}

// ResourceLimits defines tenant resource limits
type ResourceLimits struct {
	MaxConcurrentScans int           `yaml:"max_concurrent_scans"`
	MaxEndpointsPerScan int          `yaml:"max_endpoints_per_scan"`
	MaxStorageMB       int64         `yaml:"max_storage_mb"`
	ScanQuota           int           `yaml:"scan_quota_monthly"`
	RateLimit           RateLimit     `yaml:"rate_limit"`
}

// RateLimit defines tenant-specific rate limiting
type RateLimit struct {
	RequestsPerSecond int `yaml:"requests_per_second"`
	BurstSize         int `yaml:"burst_size"`
}

// NotificationConfig defines notification settings
type NotificationConfig struct {
	Email     EmailConfig     `yaml:"email"`
	Webhook   WebhookConfig   `yaml:"webhook"`
	Slack     SlackConfig     `yaml:"slack"`
	PagerDuty PagerDutyConfig `yaml:"pagerduty"`
}

// EmailConfig defines email notification settings
type EmailConfig struct {
	Enabled    bool     `yaml:"enabled"`
	SMTPHost   string   `yaml:"smtp_host"`
	SMTPPort   int      `yaml:"smtp_port"`
	Username   string   `yaml:"username"`
	Password   string   `yaml:"password"`
	FromEmail  string   `yaml:"from_email"`
	Recipients []string `yaml:"recipients"`
}

// WebhookConfig defines webhook notification settings
type WebhookConfig struct {
	Enabled  bool   `yaml:"enabled"`
	URL      string `yaml:"url"`
	Method   string `yaml:"method"`
	Headers  map[string]string `yaml:"headers"`
	Secret   string `yaml:"secret"`
}

// SlackConfig defines Slack notification settings
type SlackConfig struct {
	Enabled    bool     `yaml:"enabled"`
	WebhookURL string   `yaml:"webhook_url"`
	Channel    string   `yaml:"channel"`
	Username   string   `yaml:"username"`
}

// PagerDutyConfig defines PagerDuty notification settings
type PagerDutyConfig struct {
	Enabled      bool   `yaml:"enabled"`
	IntegrationKey string `yaml:"integration_key"`
	ServiceID    string `yaml:"service_id"`
}

// CustomBrandingConfig defines custom branding settings
type CustomBrandingConfig struct {
	Enabled           bool   `yaml:"enabled"`
	LogoURL           string `yaml:"logo_url"`
	PrimaryColor      string `yaml:"primary_color"`
	SecondaryColor    string `yaml:"secondary_color"`
	CompanyName       string `yaml:"company_name"`
	CustomFooter      string `yaml:"custom_footer"`
}

// SIEMConfig defines SIEM integration settings
type SIEMConfig struct {
	Enabled     bool              `yaml:"enabled"`
	Type        SIEMType          `yaml:"type"`
	Config      map[string]string `yaml:"config"`
	Format      SIEMFormat        `yaml:"format"`
	EndpointURL string            `yaml:"endpoint_url"`
	AuthToken   string            `yaml:"auth_token"`
}

// SIEMType represents different SIEM platforms
type SIEMType string

const (
	SIEMTypeSplunk   SIEMType = "splunk"
	SIEMTypeELK      SIEMType = "elk"
	SIEMTypeQRadar   SIEMType = "qradar"
	SIEMTypeArcSight SIEMType = "arcsight"
	SIEMTypeSyslog   SIEMType = "syslog"
)

// SIEMFormat represents SIEM data formats
type SIEMFormat string

const (
	SIEMFormatJSON    SIEMFormat = "json"
	SIEMFormatCEF     SIEMFormat = "cef"
	SIEMFormatLEEF    SIEMFormat = "leef"
	SIEMFormatSyslog  SIEMFormat = "syslog"
)

// TenantManager manages multi-tenant operations
type TenantManager struct {
	tenants      map[string]*Tenant
	mutex        sync.RWMutex
	basePath     string
	defaultTenant *Tenant
}

// NewTenantManager creates a new tenant manager
func NewTenantManager(basePath string) (*TenantManager, error) {
	if basePath == "" {
		basePath = "./tenants"
	}

	// Create base directory if it doesn't exist
	if err := os.MkdirAll(basePath, 0755); err != nil {
		return nil, fmt.Errorf("failed to create tenant directory: %v", err)
	}

	tm := &TenantManager{
		tenants:  make(map[string]*Tenant),
		basePath: basePath,
	}

	// Load existing tenants
	if err := tm.loadTenants(); err != nil {
		logging.Warn("Failed to load existing tenants", map[string]interface{}{
			"error": err.Error(),
		})
	}

	// Create default tenant if none exists
	if len(tm.tenants) == 0 {
		if err := tm.createDefaultTenant(); err != nil {
			return nil, fmt.Errorf("failed to create default tenant: %v", err)
		}
	}

	return tm, nil
}

// CreateTenant creates a new tenant
func (tm *TenantManager) CreateTenant(tenant *Tenant) error {
	tm.mutex.Lock()
	defer tm.mutex.Unlock()

	// Validate tenant
	if tenant.ID == "" {
		return fmt.Errorf("tenant ID is required")
	}

	if tenant.Name == "" {
		return fmt.Errorf("tenant name is required")
	}

	// Check if tenant already exists
	if _, exists := tm.tenants[tenant.ID]; exists {
		return fmt.Errorf("tenant with ID '%s' already exists", tenant.ID)
	}

	// Set default values
	if tenant.CreatedAt.IsZero() {
		tenant.CreatedAt = time.Now()
	}
	tenant.UpdatedAt = time.Now()

	// Set default settings if not provided
	if tenant.Settings.ResourceLimits.MaxConcurrentScans == 0 {
		tenant.Settings.ResourceLimits.MaxConcurrentScans = 5
	}
	if tenant.Settings.ResourceLimits.MaxEndpointsPerScan == 0 {
		tenant.Settings.ResourceLimits.MaxEndpointsPerScan = 100
	}
	// RetentionDays is handled by DataIsolation config

	// Create tenant directory
	tenantPath := filepath.Join(tm.basePath, tenant.ID)
	if err := os.MkdirAll(tenantPath, 0755); err != nil {
		return fmt.Errorf("failed to create tenant directory: %v", err)
	}

	// Save tenant configuration
	if err := tm.saveTenant(tenant); err != nil {
		return fmt.Errorf("failed to save tenant: %v", err)
	}

	// Add to in-memory map
	tm.tenants[tenant.ID] = tenant

	logging.Info("Created new tenant", map[string]interface{}{
		"tenant_id":   tenant.ID,
		"tenant_name": tenant.Name,
	})

	return nil
}

// GetTenant retrieves a tenant by ID
func (tm *TenantManager) GetTenant(tenantID string) (*Tenant, error) {
	tm.mutex.RLock()
	defer tm.mutex.RUnlock()

	tenant, exists := tm.tenants[tenantID]
	if !exists {
		return nil, fmt.Errorf("tenant '%s' not found", tenantID)
	}

	if !tenant.IsActive {
		return nil, fmt.Errorf("tenant '%s' is not active", tenantID)
	}

	return tenant, nil
}

// UpdateTenant updates an existing tenant
func (tm *TenantManager) UpdateTenant(tenantID string, updates *Tenant) error {
	tm.mutex.Lock()
	defer tm.mutex.Unlock()

	tenant, exists := tm.tenants[tenantID]
	if !exists {
		return fmt.Errorf("tenant '%s' not found", tenantID)
	}

	// Update allowed fields
	if updates.Name != "" {
		tenant.Name = updates.Name
	}
	if updates.Description != "" {
		tenant.Description = updates.Description
	}
	if updates.Config != nil {
		tenant.Config = updates.Config
	}
	tenant.UpdatedAt = time.Now()

	// Save updated tenant
	if err := tm.saveTenant(tenant); err != nil {
		return fmt.Errorf("failed to save tenant updates: %v", err)
	}

	logging.Info("Updated tenant", map[string]interface{}{
		"tenant_id":   tenant.ID,
		"tenant_name": tenant.Name,
	})

	return nil
}

// DeleteTenant deactivates a tenant
func (tm *TenantManager) DeleteTenant(tenantID string) error {
	tm.mutex.Lock()
	defer tm.mutex.Unlock()

	tenant, exists := tm.tenants[tenantID]
	if !exists {
		return fmt.Errorf("tenant '%s' not found", tenantID)
	}

	// Don't allow deletion of default tenant
	if tenantID == "default" {
		return fmt.Errorf("cannot delete default tenant")
	}

	tenant.IsActive = false
	tenant.UpdatedAt = time.Now()

	if err := tm.saveTenant(tenant); err != nil {
		return fmt.Errorf("failed to save tenant deactivation: %v", err)
	}

	logging.Info("Deactivated tenant", map[string]interface{}{
		"tenant_id":   tenant.ID,
		"tenant_name": tenant.Name,
	})

	return nil
}

// ListTenants returns all active tenants
func (tm *TenantManager) ListTenants() []*Tenant {
	tm.mutex.RLock()
	defer tm.mutex.RUnlock()

	var activeTenants []*Tenant
	for _, tenant := range tm.tenants {
		if tenant.IsActive {
			activeTenants = append(activeTenants, tenant)
		}
	}

	return activeTenants
}

// ValidateTenantResourceUsage validates if tenant is within resource limits
func (tm *TenantManager) ValidateTenantResourceUsage(tenantID string, endpointCount int) error {
	tenant, err := tm.GetTenant(tenantID)
	if err != nil {
		return err
	}

	// Check endpoint limit
	if endpointCount > tenant.Settings.ResourceLimits.MaxEndpointsPerScan {
		return fmt.Errorf("tenant '%s' exceeded maximum endpoints per scan limit: %d",
			tenantID, tenant.Settings.ResourceLimits.MaxEndpointsPerScan)
	}

	// TODO: Add more resource validation (concurrent scans, storage quota, etc.)

	return nil
}

// GetTenantConfigPath returns the configuration file path for a tenant
func (tm *TenantManager) GetTenantConfigPath(tenantID string) string {
	return filepath.Join(tm.basePath, tenantID, "config.yaml")
}

// GetTenantDataPath returns the data directory path for a tenant
func (tm *TenantManager) GetTenantDataPath(tenantID string) string {
	return filepath.Join(tm.basePath, tenantID, "data")
}

// createDefaultTenant creates a default tenant
func (tm *TenantManager) createDefaultTenant() error {
	defaultTenant := &Tenant{
		ID:          "default",
		Name:        "Default Organization",
		Description: "Default tenant for single-tenant use",
		Config: &scanner.Config{
			APIEndpoints: []types.APIEndpoint{
				{
					URL:    "https://api.example.com/users",
					Method: "GET",
				},
			},
		},
		Settings: TenantSettings{
			DataIsolation: DataIsolationConfig{
				Enabled:       false,
				StoragePath:   "./history",
				RetentionDays: 30,
			},
			ResourceLimits: ResourceLimits{
				MaxConcurrentScans:  10,
				MaxEndpointsPerScan: 1000,
				MaxStorageMB:       10240,
				ScanQuota:           1000,
				RateLimit: RateLimit{
					RequestsPerSecond: 100,
					BurstSize:         50,
				},
			},
		},
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		IsActive:  true,
	}

	return tm.CreateTenant(defaultTenant)
}

// loadTenants loads all tenants from disk
func (tm *TenantManager) loadTenants() error {
	entries, err := os.ReadDir(tm.basePath)
	if err != nil {
		return err
	}

	for _, entry := range entries {
		if entry.IsDir() {
			tenantID := entry.Name()
			configPath := tm.GetTenantConfigPath(tenantID)

			data, err := os.ReadFile(configPath)
			if err != nil {
				logging.Warn("Failed to read tenant config", map[string]interface{}{
					"tenant_id": tenantID,
					"error":     err.Error(),
				})
				continue
			}

			var tenant Tenant
			if err := yaml.Unmarshal(data, &tenant); err != nil {
				logging.Warn("Failed to parse tenant config", map[string]interface{}{
					"tenant_id": tenantID,
					"error":     err.Error(),
				})
				continue
			}

			tm.tenants[tenantID] = &tenant
		}
	}

	return nil
}

// saveTenant saves a tenant configuration to disk
func (tm *TenantManager) saveTenant(tenant *Tenant) error {
	// Create tenant data directory
	dataPath := tm.GetTenantDataPath(tenant.ID)
	if err := os.MkdirAll(dataPath, 0755); err != nil {
		return err
	}

	// Save tenant configuration
	configPath := tm.GetTenantConfigPath(tenant.ID)
	data, err := yaml.Marshal(tenant)
	if err != nil {
		return err
	}

	if err := os.WriteFile(configPath, data, 0644); err != nil {
		return err
	}

	return nil
}