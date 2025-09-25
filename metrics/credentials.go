package metrics

import (
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"
)

const (
	defaultDashboardUsername = "admin"
	defaultDashboardPassword = "admin"
	credentialsFileName      = "dashboard_credentials.json"
	saltBytes                = 16
	derivedKeyBytes          = 32
	defaultIterations        = 120000
)

type credentialRecord struct {
	Username     string    `json:"username"`
	PasswordHash string    `json:"password_hash"`
	UpdatedAt    time.Time `json:"updated_at"`
}

// CredentialManager manages dashboard credentials that can be rotated from the UI.
type CredentialManager struct {
	path   string
	mutex  sync.RWMutex
	record credentialRecord
}

// NewCredentialManager loads dashboard credentials from disk (or seeds defaults).
func NewCredentialManager(path string) (*CredentialManager, error) {
	cm := &CredentialManager{}
	if path == "" {
		path = filepath.Join("config", credentialsFileName)
	}
	cm.path = path

	if err := cm.loadFromDisk(); err != nil {
		if errors.Is(err, os.ErrNotExist) {
			if err := cm.seedDefault(); err != nil {
				return nil, err
			}
		} else {
			return nil, err
		}
	}

	return cm, nil
}

// NewInMemoryCredentialManager provides a fallback credential store that is not persisted.
func NewInMemoryCredentialManager() (*CredentialManager, error) {
	passwordHash, err := hashPassword(defaultDashboardPassword, defaultIterations)
	if err != nil {
		return nil, fmt.Errorf("failed to initialise in-memory credentials: %w", err)
	}

	return &CredentialManager{
		record: credentialRecord{
			Username:     defaultDashboardUsername,
			PasswordHash: passwordHash,
			UpdatedAt:    time.Now().UTC(),
		},
	}, nil
}

// Username returns the currently configured username.
func (cm *CredentialManager) Username() string {
	cm.mutex.RLock()
	defer cm.mutex.RUnlock()
	return cm.record.Username
}

// Authenticate validates a set of credentials against the stored record.
func (cm *CredentialManager) Authenticate(username, password string) bool {
	cm.mutex.RLock()
	record := cm.record
	cm.mutex.RUnlock()

	if username != record.Username {
		return false
	}

	if record.PasswordHash == "" {
		return false
	}

	return verifyPassword(record.PasswordHash, password)
}

// Update rotates the dashboard credentials, enforcing basic policy checks.
func (cm *CredentialManager) Update(username, password string) error {
	username = strings.TrimSpace(username)
	if username == "" {
		return fmt.Errorf("username cannot be empty")
	}

	if len(password) < 12 {
		return fmt.Errorf("password must be at least 12 characters long")
	}

	if strings.EqualFold(username, password) {
		return fmt.Errorf("password cannot match the username")
	}

	if err := cm.persist(username, password); err != nil {
		return err
	}

	return nil
}

// LastUpdated exposes the timestamp of the last rotation.
func (cm *CredentialManager) LastUpdated() time.Time {
	cm.mutex.RLock()
	defer cm.mutex.RUnlock()
	return cm.record.UpdatedAt
}

func (cm *CredentialManager) loadFromDisk() error {
	cm.mutex.Lock()
	defer cm.mutex.Unlock()

	data, err := os.ReadFile(cm.path)
	if err != nil {
		return err
	}

	var record credentialRecord
	if err := json.Unmarshal(data, &record); err != nil {
		return err
	}

	if record.Username == "" || record.PasswordHash == "" {
		return fmt.Errorf("invalid credential record")
	}

	cm.record = record
	return nil
}

func (cm *CredentialManager) seedDefault() error {
	return cm.persist(defaultDashboardUsername, defaultDashboardPassword)
}

func (cm *CredentialManager) persist(username, password string) error {
	cm.mutex.Lock()
	defer cm.mutex.Unlock()

	passwordHash, err := hashPassword(password, defaultIterations)
	if err != nil {
		return fmt.Errorf("failed to hash password: %w", err)
	}

	record := credentialRecord{
		Username:     username,
		PasswordHash: passwordHash,
		UpdatedAt:    time.Now().UTC(),
	}

	if err := cm.writeRecord(record); err != nil {
		return err
	}

	cm.record = record
	return nil
}

func (cm *CredentialManager) writeRecord(record credentialRecord) error {
	if err := os.MkdirAll(filepath.Dir(cm.path), 0o755); err != nil {
		return fmt.Errorf("failed to prepare credential path: %w", err)
	}

	data, err := json.MarshalIndent(record, "", "  ")
	if err != nil {
		return err
	}

	tmp := cm.path + ".tmp"
	if err := os.WriteFile(tmp, data, 0o600); err != nil {
		return err
	}

	return os.Rename(tmp, cm.path)
}

func hashPassword(password string, iterations int) (string, error) {
	salt := make([]byte, saltBytes)
	if _, err := rand.Read(salt); err != nil {
		return "", fmt.Errorf("failed to generate salt: %w", err)
	}

	derived := pbkdf2Key([]byte(password), salt, iterations, derivedKeyBytes)

	return fmt.Sprintf("%d:%s:%s", iterations,
		base64.StdEncoding.EncodeToString(salt),
		base64.StdEncoding.EncodeToString(derived),
	), nil
}

func verifyPassword(encodedHash, password string) bool {
	parts := strings.Split(encodedHash, ":")
	if len(parts) != 3 {
		return false
	}

	iterations, err := strconv.Atoi(parts[0])
	if err != nil || iterations <= 0 {
		return false
	}

	salt, err := base64.StdEncoding.DecodeString(parts[1])
	if err != nil {
		return false
	}

	stored, err := base64.StdEncoding.DecodeString(parts[2])
	if err != nil {
		return false
	}

	derived := pbkdf2Key([]byte(password), salt, iterations, len(stored))
	if len(derived) != len(stored) {
		return false
	}

	return subtle.ConstantTimeCompare(derived, stored) == 1
}

func pbkdf2Key(password, salt []byte, iterations, keyLen int) []byte {
	if iterations <= 0 {
		iterations = defaultIterations
	}

	hLen := sha256.Size
	numBlocks := (keyLen + hLen - 1) / hLen
	derived := make([]byte, numBlocks*hLen)

	for block := 1; block <= numBlocks; block++ {
		U := pbkdf2Block(password, salt, iterations, block)
		copy(derived[(block-1)*hLen:], U)
	}

	return derived[:keyLen]
}

func pbkdf2Block(password, salt []byte, iterations, blockNum int) []byte {
	mac := hmac.New(sha256.New, password)
	mac.Write(salt)
	mac.Write(intToBytes(blockNum))
	U := mac.Sum(nil)
	T := make([]byte, len(U))
	copy(T, U)

	for i := 1; i < iterations; i++ {
		mac = hmac.New(sha256.New, password)
		mac.Write(U)
		U = mac.Sum(nil)
		for j := range T {
			T[j] ^= U[j]
		}
	}

	return T
}

func intToBytes(i int) []byte {
	return []byte{
		byte(i >> 24),
		byte(i >> 16),
		byte(i >> 8),
		byte(i),
	}
}
