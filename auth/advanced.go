package auth

import (
	"context"
	"crypto/x509"
	"encoding/base64"
	"encoding/pem"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/clientcredentials"
)

// AdvancedAuthConfig represents advanced authentication configuration
type AdvancedAuthConfig struct {
	Enabled          bool                   `yaml:"enabled"`
	Type             AuthType               `yaml:"type"`
	BasicAuth        *BasicAuthConfig      `yaml:"basic_auth,omitempty"`
	BearerToken      *BearerTokenConfig    `yaml:"bearer_token,omitempty"`
	APIKey           *APIKeyConfig         `yaml:"api_key,omitempty"`
	OAuth2           *OAuth2Config         `yaml:"oauth2,omitempty"`
	JWT              *JWTConfig            `yaml:"jwt,omitempty"`
	MutualTLS        *MutualTLSConfig      `yaml:"mutual_tls,omitempty"`
}

// AuthType represents authentication types
type AuthType string

const (
	AuthTypeBasic     AuthType = "basic"
	AuthTypeBearer    AuthType = "bearer"
	AuthTypeAPIKey    AuthType = "api_key"
	AuthTypeOAuth2    AuthType = "oauth2"
	AuthTypeJWT       AuthType = "jwt"
	AuthTypeMutualTLS AuthType = "mutual_tls"
)

// BasicAuthConfig represents basic authentication configuration
type BasicAuthConfig struct {
	Username string `yaml:"username"`
	Password string `yaml:"password"`
}

// BearerTokenConfig represents bearer token configuration
type BearerTokenConfig struct {
	Token     string        `yaml:"token"`
	ExpiresIn time.Duration `yaml:"expires_in"`
}

// APIKeyConfig represents API key configuration
type APIKeyConfig struct {
	Key         string `yaml:"key"`
	Value       string `yaml:"value"`
	Location    string `yaml:"location"` // "header", "query", "cookie"
	Prefix      string `yaml:"prefix"`   // e.g., "Bearer ", "API-Key "
}

// OAuth2Config represents OAuth2 configuration
type OAuth2Config struct {
	GrantType    string            `yaml:"grant_type"`    // "client_credentials", "authorization_code", "password"
	ClientID     string            `yaml:"client_id"`
	ClientSecret string            `yaml:"client_secret"`
	TokenURL     string            `yaml:"token_url"`
	AuthURL      string            `yaml:"auth_url"`      // for authorization_code flow
	RedirectURL  string            `yaml:"redirect_url"`  // for authorization_code flow
	Scopes       []string          `yaml:"scopes"`
	Headers      map[string]string `yaml:"headers"`
}

// JWTConfig represents JWT configuration
type JWTConfig struct {
	Algorithm   string   `yaml:"algorithm"`   // "HS256", "RS256", "ES256", etc.
	Secret      string   `yaml:"secret"`      // for HMAC algorithms
	PrivateKey  string   `yaml:"private_key"` // for RSA/ECDSA algorithms
	PublicKey   string   `yaml:"public_key"`  // for RSA/ECDSA algorithms
	KeyFile     string   `yaml:"key_file"`    // path to key file
	Claims      JWTClaims `yaml:"claims"`
}

// JWTClaims represents JWT claims configuration
type JWTClaims struct {
	Issuer    string   `yaml:"iss"`
	Subject   string   `yaml:"sub"`
	Audience  string   `yaml:"aud"`
	ExpiresIn int      `yaml:"exp"` // seconds
	NotBefore int      `yaml:"nbf"` // seconds
	IssuedAt  int      `yaml:"iat"` // seconds
	Roles     []string `yaml:"roles"`
	Custom    map[string]interface{} `yaml:"custom"`
}

// MutualTLSConfig represents mutual TLS configuration
type MutualTLSConfig struct {
	CertFile string `yaml:"cert_file"`
	KeyFile  string `yaml:"key_file"`
	CAFile   string `yaml:"ca_file"`
}

// AuthResult represents authentication result
type AuthResult struct {
	Success    bool
	Token      string
	Headers    map[string]string
	ExpiresAt  time.Time
	Error      error
}

// AdvancedAuthManager manages advanced authentication methods
type AdvancedAuthManager struct {
	config *AdvancedAuthConfig
	client *http.Client
	oauth2Config *oauth2.Config
	jwtKey      interface{}
}

// NewAdvancedAuthManager creates a new advanced authentication manager
func NewAdvancedAuthManager(config *AdvancedAuthConfig) (*AdvancedAuthManager, error) {
	if !config.Enabled {
		return nil, fmt.Errorf("advanced authentication is not enabled")
	}

	manager := &AdvancedAuthManager{
		config: config,
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
	}

	// Initialize specific authentication type
	switch config.Type {
	case AuthTypeOAuth2:
		if err := manager.initOAuth2(); err != nil {
			return nil, fmt.Errorf("failed to initialize OAuth2: %v", err)
		}
	case AuthTypeJWT:
		if err := manager.initJWT(); err != nil {
			return nil, fmt.Errorf("failed to initialize JWT: %v", err)
		}
	case AuthTypeMutualTLS:
		if err := manager.initMutualTLS(); err != nil {
			return nil, fmt.Errorf("failed to initialize mutual TLS: %v", err)
		}
	}

	return manager, nil
}

// Authenticate performs authentication based on the configured type
func (m *AdvancedAuthManager) Authenticate() *AuthResult {
	switch m.config.Type {
	case AuthTypeBasic:
		return m.authenticateBasic()
	case AuthTypeBearer:
		return m.authenticateBearer()
	case AuthTypeAPIKey:
		return m.authenticateAPIKey()
	case AuthTypeOAuth2:
		return m.authenticateOAuth2()
	case AuthTypeJWT:
		return m.authenticateJWT()
	case AuthTypeMutualTLS:
		return m.authenticateMutualTLS()
	default:
		return &AuthResult{
			Success: false,
			Error:   fmt.Errorf("unsupported authentication type: %s", m.config.Type),
		}
	}
}

// RefreshToken refreshes the authentication token if supported
func (m *AdvancedAuthManager) RefreshToken() *AuthResult {
	switch m.config.Type {
	case AuthTypeOAuth2:
		return m.refreshOAuth2()
	case AuthTypeJWT:
		return m.authenticateJWT() // Generate new JWT
	case AuthTypeBearer:
		if m.config.BearerToken.ExpiresIn > 0 {
			return m.authenticateBearer()
		}
		return &AuthResult{
			Success: true,
			Token:   m.config.BearerToken.Token,
		}
	default:
		return &AuthResult{
			Success: false,
			Error:   fmt.Errorf("token refresh not supported for authentication type: %s", m.config.Type),
		}
	}
}

// authenticateBasic performs basic authentication
func (m *AdvancedAuthManager) authenticateBasic() *AuthResult {
	if m.config.BasicAuth == nil {
		return &AuthResult{
			Success: false,
			Error:   fmt.Errorf("basic authentication configuration is missing"),
		}
	}

	headers := make(map[string]string)
	token := base64.StdEncoding.EncodeToString([]byte(
		fmt.Sprintf("%s:%s", m.config.BasicAuth.Username, m.config.BasicAuth.Password),
	))
	headers["Authorization"] = "Basic " + token

	return &AuthResult{
		Success: true,
		Token:   token,
		Headers: headers,
	}
}

// authenticateBearer performs bearer token authentication
func (m *AdvancedAuthManager) authenticateBearer() *AuthResult {
	if m.config.BearerToken == nil {
		return &AuthResult{
			Success: false,
			Error:   fmt.Errorf("bearer token configuration is missing"),
		}
	}

	var expiresAt time.Time
	if m.config.BearerToken.ExpiresIn > 0 {
		expiresAt = time.Now().Add(m.config.BearerToken.ExpiresIn)
	}

	headers := make(map[string]string)
	headers["Authorization"] = "Bearer " + m.config.BearerToken.Token

	return &AuthResult{
		Success:   true,
		Token:     m.config.BearerToken.Token,
		Headers:   headers,
		ExpiresAt: expiresAt,
	}
}

// authenticateAPIKey performs API key authentication
func (m *AdvancedAuthManager) authenticateAPIKey() *AuthResult {
	if m.config.APIKey == nil {
		return &AuthResult{
			Success: false,
			Error:   fmt.Errorf("API key configuration is missing"),
		}
	}

	headers := make(map[string]string)
	token := m.config.APIKey.Value

	switch strings.ToLower(m.config.APIKey.Location) {
	case "header":
		headerName := m.config.APIKey.Key
		if headerName == "" {
			headerName = "X-API-Key"
		}
		headers[headerName] = m.config.APIKey.Prefix + token
	case "query":
		// Query parameters are handled differently, not in headers
	default:
		headers["Authorization"] = m.config.APIKey.Prefix + token
	}

	return &AuthResult{
		Success: true,
		Token:   token,
		Headers: headers,
	}
}

// authenticateOAuth2 performs OAuth2 authentication
func (m *AdvancedAuthManager) authenticateOAuth2() *AuthResult {
	if m.oauth2Config == nil {
		return &AuthResult{
			Success: false,
			Error:   fmt.Errorf("OAuth2 configuration is not initialized"),
		}
	}

	var token *oauth2.Token
	var err error

	switch m.config.OAuth2.GrantType {
	case "client_credentials":
		config := clientcredentials.Config{
			ClientID:     m.config.OAuth2.ClientID,
			ClientSecret: m.config.OAuth2.ClientSecret,
			TokenURL:     m.config.OAuth2.TokenURL,
			Scopes:       m.config.OAuth2.Scopes,
		}
		token, err = config.Token(context.Background())
	case "authorization_code":
		// This requires user interaction, simplified for automated scanning
		return &AuthResult{
			Success: false,
			Error:   fmt.Errorf("authorization_code grant type requires user interaction"),
		}
	case "password":
		// OAuth2 password grant (resource owner password credentials)
		config := m.oauth2Config
		token, err = config.PasswordCredentialsToken(context.Background(), "", "")
	default:
		return &AuthResult{
			Success: false,
			Error:   fmt.Errorf("unsupported OAuth2 grant type: %s", m.config.OAuth2.GrantType),
		}
	}

	if err != nil {
		return &AuthResult{
			Success: false,
			Error:   fmt.Errorf("OAuth2 authentication failed: %v", err),
		}
	}

	headers := make(map[string]string)
	headers["Authorization"] = "Bearer " + token.AccessToken

	return &AuthResult{
		Success:   true,
		Token:     token.AccessToken,
		Headers:   headers,
		ExpiresAt: token.Expiry,
	}
}

// authenticateJWT performs JWT authentication
func (m *AdvancedAuthManager) authenticateJWT() *AuthResult {
	if m.jwtKey == nil {
		return &AuthResult{
			Success: false,
			Error:   fmt.Errorf("JWT key is not initialized"),
		}
	}

	// Create JWT claims
	claims := jwt.MapClaims{}
	now := time.Now()

	// Set standard claims
	if m.config.JWT.Claims.Issuer != "" {
		claims["iss"] = m.config.JWT.Claims.Issuer
	}
	if m.config.JWT.Claims.Subject != "" {
		claims["sub"] = m.config.JWT.Claims.Subject
	}
	if m.config.JWT.Claims.Audience != "" {
		claims["aud"] = m.config.JWT.Claims.Audience
	}
	if m.config.JWT.Claims.ExpiresIn > 0 {
		claims["exp"] = now.Add(time.Duration(m.config.JWT.Claims.ExpiresIn) * time.Second).Unix()
	}
	if m.config.JWT.Claims.NotBefore > 0 {
		claims["nbf"] = now.Add(time.Duration(m.config.JWT.Claims.NotBefore) * time.Second).Unix()
	}
	if m.config.JWT.Claims.IssuedAt > 0 {
		claims["iat"] = now.Add(time.Duration(m.config.JWT.Claims.IssuedAt) * time.Second).Unix()
	}
	if len(m.config.JWT.Claims.Roles) > 0 {
		claims["roles"] = m.config.JWT.Claims.Roles
	}

	// Set custom claims
	for key, value := range m.config.JWT.Claims.Custom {
		claims[key] = value
	}

	// Create token
	token := jwt.NewWithClaims(getJWTSigningMethod(m.config.JWT.Algorithm), claims)

	// Sign token
	tokenString, err := token.SignedString(m.jwtKey)
	if err != nil {
		return &AuthResult{
			Success: false,
			Error:   fmt.Errorf("failed to sign JWT: %v", err),
		}
	}

	headers := make(map[string]string)
	headers["Authorization"] = "Bearer " + tokenString

	var expiresAt time.Time
	if m.config.JWT.Claims.ExpiresIn > 0 {
		expiresAt = now.Add(time.Duration(m.config.JWT.Claims.ExpiresIn) * time.Second)
	}

	return &AuthResult{
		Success:   true,
		Token:     tokenString,
		Headers:   headers,
		ExpiresAt: expiresAt,
	}
}

// authenticateMutualTLS performs mutual TLS authentication
func (m *AdvancedAuthManager) authenticateMutualTLS() *AuthResult {
	// Mutual TLS is configured at the transport level, not token-based
	return &AuthResult{
		Success: true,
		Token:   "mutual_tls_configured",
		Headers: make(map[string]string),
	}
}

// refreshOAuth2 refreshes OAuth2 token
func (m *AdvancedAuthManager) refreshOAuth2() *AuthResult {
	if m.oauth2Config == nil {
		return &AuthResult{
			Success: false,
			Error:   fmt.Errorf("OAuth2 configuration is not initialized"),
		}
	}

	// For client credentials flow, just get a new token
	if m.config.OAuth2.GrantType == "client_credentials" {
		return m.authenticateOAuth2()
	}

	return &AuthResult{
		Success: false,
		Error:   fmt.Errorf("token refresh not implemented for OAuth2 grant type: %s", m.config.OAuth2.GrantType),
	}
}

// initOAuth2 initializes OAuth2 configuration
func (m *AdvancedAuthManager) initOAuth2() error {
	if m.config.OAuth2 == nil {
		return fmt.Errorf("OAuth2 configuration is missing")
	}

	m.oauth2Config = &oauth2.Config{
		ClientID:     m.config.OAuth2.ClientID,
		ClientSecret: m.config.OAuth2.ClientSecret,
		Endpoint: oauth2.Endpoint{
			TokenURL: m.config.OAuth2.TokenURL,
			AuthURL:  m.config.OAuth2.AuthURL,
		},
		RedirectURL: m.config.OAuth2.RedirectURL,
		Scopes:      m.config.OAuth2.Scopes,
	}

	return nil
}

// initJWT initializes JWT configuration
func (m *AdvancedAuthManager) initJWT() error {
	if m.config.JWT == nil {
		return fmt.Errorf("JWT configuration is missing")
	}

	switch strings.ToUpper(m.config.JWT.Algorithm) {
	case "HS256", "HS384", "HS512":
		// HMAC algorithms use a secret key
		if m.config.JWT.Secret == "" {
			return fmt.Errorf("JWT secret is required for HMAC algorithms")
		}
		m.jwtKey = []byte(m.config.JWT.Secret)
	case "RS256", "RS384", "RS512", "ES256", "ES384", "ES512":
		// RSA/ECDSA algorithms use public/private key pair
		key, err := m.loadJWTKey()
		if err != nil {
			return fmt.Errorf("failed to load JWT key: %v", err)
		}
		m.jwtKey = key
	default:
		return fmt.Errorf("unsupported JWT algorithm: %s", m.config.JWT.Algorithm)
	}

	return nil
}

// initMutualTLS initializes mutual TLS configuration
func (m *AdvancedAuthManager) initMutualTLS() error {
	if m.config.MutualTLS == nil {
		return fmt.Errorf("mutual TLS configuration is missing")
	}

	// Verify that certificate and key files exist
	if _, err := os.Stat(m.config.MutualTLS.CertFile); os.IsNotExist(err) {
		return fmt.Errorf("certificate file does not exist: %s", m.config.MutualTLS.CertFile)
	}
	if _, err := os.Stat(m.config.MutualTLS.KeyFile); os.IsNotExist(err) {
		return fmt.Errorf("key file does not exist: %s", m.config.MutualTLS.KeyFile)
	}

	// Configure HTTP client with TLS
	// Note: This would require custom TLS configuration in the http.Transport
	return nil
}

// loadJWTKey loads JWT key from configuration or file
func (m *AdvancedAuthManager) loadJWTKey() (interface{}, error) {
	var keyData []byte
	var err error

	if m.config.JWT.KeyFile != "" {
		keyData, err = os.ReadFile(m.config.JWT.KeyFile)
		if err != nil {
			return nil, fmt.Errorf("failed to read key file: %v", err)
		}
	} else if m.config.JWT.PrivateKey != "" {
		keyData = []byte(m.config.JWT.PrivateKey)
	} else {
		return nil, fmt.Errorf("JWT key is not configured")
	}

	// Parse the key based on algorithm
	if strings.HasPrefix(m.config.JWT.Algorithm, "RS") {
		// RSA key
		block, _ := pem.Decode(keyData)
		if block == nil {
			return nil, fmt.Errorf("failed to decode PEM block")
		}

		key, err := x509.ParsePKCS1PrivateKey(block.Bytes)
		if err != nil {
			return nil, fmt.Errorf("failed to parse RSA private key: %v", err)
		}
		return key, nil
	} else if strings.HasPrefix(m.config.JWT.Algorithm, "ES") {
		// ECDSA key
		block, _ := pem.Decode(keyData)
		if block == nil {
			return nil, fmt.Errorf("failed to decode PEM block")
		}

		key, err := x509.ParseECPrivateKey(block.Bytes)
		if err != nil {
			return nil, fmt.Errorf("failed to parse ECDSA private key: %v", err)
		}
		return key, nil
	}

	return nil, fmt.Errorf("unsupported JWT algorithm for key loading: %s", m.config.JWT.Algorithm)
}

// getJWTSigningMethod returns JWT signing method based on algorithm
func getJWTSigningMethod(algorithm string) jwt.SigningMethod {
	switch strings.ToUpper(algorithm) {
	case "HS256":
		return jwt.SigningMethodHS256
	case "HS384":
		return jwt.SigningMethodHS384
	case "HS512":
		return jwt.SigningMethodHS512
	case "RS256":
		return jwt.SigningMethodRS256
	case "RS384":
		return jwt.SigningMethodRS384
	case "RS512":
		return jwt.SigningMethodRS512
	case "ES256":
		return jwt.SigningMethodES256
	case "ES384":
		return jwt.SigningMethodES384
	case "ES512":
		return jwt.SigningMethodES512
	default:
		return jwt.SigningMethodHS256 // Default
	}
}

// ApplyToRequest applies authentication to an HTTP request
func (m *AdvancedAuthManager) ApplyToRequest(req *http.Request) error {
	if m.config == nil || !m.config.Enabled {
		return nil
	}

	authResult := m.Authenticate()
	if !authResult.Success {
		return fmt.Errorf("authentication failed: %v", authResult.Error)
	}

	// Apply headers
	for key, value := range authResult.Headers {
		req.Header.Set(key, value)
	}

	// Apply API key in query parameter if configured
	if m.config.Type == AuthTypeAPIKey && m.config.APIKey != nil &&
		strings.ToLower(m.config.APIKey.Location) == "query" {

		// Parse existing query parameters
		q := req.URL.Query()
		q.Add(m.config.APIKey.Key, m.config.APIKey.Value)
		req.URL.RawQuery = q.Encode()
	}

	return nil
}

// ValidateToken validates a JWT token
func (m *AdvancedAuthManager) ValidateToken(tokenString string) (*jwt.MapClaims, error) {
	if m.config.Type != AuthTypeJWT || m.jwtKey == nil {
		return nil, fmt.Errorf("JWT validation is not configured")
	}

	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		// Validate signing method
		if token.Method != getJWTSigningMethod(m.config.JWT.Algorithm) {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return m.jwtKey, nil
	})

	if err != nil {
		return nil, fmt.Errorf("failed to parse JWT: %v", err)
	}

	if !token.Valid {
		return nil, fmt.Errorf("invalid JWT token")
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return nil, fmt.Errorf("invalid JWT claims format")
	}

	return &claims, nil
}

// GetHTTPClient returns HTTP client configured for this authentication method
func (m *AdvancedAuthManager) GetHTTPClient() *http.Client {
	// Return the configured client
	// For mutual TLS, this would include custom TLS configuration
	return m.client
}