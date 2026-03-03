package config

import (
	"encoding/json"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"time"
)

// ConnectionSettings for timeout configuration
type ConnectionSettings struct {
	ConnectTimeout int `json:"connectTimeout"` // ms, 0 = default (10s)
	ReadTimeout    int `json:"readTimeout"`    // ms, 0 = unlimited
}

// Variables represents a key-value map for script variables
type Variables map[string]interface{}

// ScriptConfig represents script configuration for a case or collection
type ScriptConfig struct {
	SetupScript     string `json:"setupScript"`
	SetupEnabled    bool   `json:"setupEnabled"`
	PreSendScript   string `json:"preSendScript"`
	PreSendEnabled  bool   `json:"preSendEnabled"`
	PostRecvScript  string `json:"postRecvScript"`
	PostRecvEnabled bool   `json:"postRecvEnabled"`
}

// FramingConfig represents framing configuration
type FramingConfig struct {
	Mode           string `json:"mode"`                     // collection, none, delimiter, length-prefix, fixed-length
	Delimiter      string `json:"delimiter,omitempty"`      // \n, \r\n, \0, custom
	LengthEncoding string `json:"lengthEncoding,omitempty"` // binary, ascii, hex, bcd
	LengthOffset   int    `json:"lengthOffset,omitempty"`
	LengthBytes    int    `json:"lengthBytes,omitempty"`    // 1, 2, 4
	Endian         string `json:"endian,omitempty"`         // big, little
	IncludeHeader  bool   `json:"includeHeader,omitempty"`
	LengthMode     string `json:"lengthMode,omitempty"`     // append, rewrite
	FixedSize      int    `json:"fixedSize,omitempty"`
}

// Collection represents a collection folder
type Collection struct {
	Name                     string              `json:"name"`
	Description              string              `json:"description,omitempty"`
	CreatedAt                string              `json:"createdAt"`
	UpdatedAt                string              `json:"updatedAt"`
	Order                    int                 `json:"order"`                              // Display order
	SharedFraming            *FramingConfig      `json:"sharedFraming,omitempty"`
	SharedCharset            string              `json:"sharedCharset,omitempty"`            // utf-8, ascii, euc-kr, etc.
	SharedConnectionSettings *ConnectionSettings `json:"sharedConnectionSettings,omitempty"` // Timeout settings
	Notes                    string              `json:"notes,omitempty"`                    // Markdown notes
	SharedScriptConfig       *ScriptConfig       `json:"sharedScriptConfig,omitempty"`       // Shared scripts
	SharedVariables          Variables           `json:"sharedVariables,omitempty"`          // Shared variables
}

// SavedCase represents a saved test case
type SavedCase struct {
	ID                 string              `json:"id"`
	Name               string              `json:"name"`
	Protocol           string              `json:"protocol"`
	Host               string              `json:"host"`
	Port               int                 `json:"port"`
	CreatedAt          string              `json:"createdAt"`
	UpdatedAt          string              `json:"updatedAt"`
	Order              int                 `json:"order"`                        // Display order
	Framing            *FramingConfig      `json:"framing,omitempty"`
	Charset            string              `json:"charset,omitempty"`            // global, collection, or specific charset
	ConnectionSettings *ConnectionSettings `json:"connectionSettings,omitempty"` // Timeout settings
	DraftMessage       string              `json:"draftMessage,omitempty"`
	DraftFormat        string              `json:"draftFormat,omitempty"`
	UseVariables       bool                `json:"useVariables"`              // Enable variable substitution
	Notes              string              `json:"notes,omitempty"`           // Markdown notes
	ScriptConfig       *ScriptConfig       `json:"scriptConfig,omitempty"`    // Script configuration
	LocalVariables     Variables           `json:"localVariables,omitempty"`
	PostRecvSample     string              `json:"postRecvSample,omitempty"`  // Sample message for post-recv dry run
}

// CollectionWithCases represents a collection with its cases
type CollectionWithCases struct {
	Collection Collection  `json:"collection"`
	Cases      []SavedCase `json:"cases"`
}

// AppSettings represents global application settings
type AppSettings struct {
	DefaultCharset     string              `json:"defaultCharset"`               // utf-8, ascii, euc-kr, etc.
	ConnectionSettings *ConnectionSettings `json:"connectionSettings,omitempty"` // Timeout settings
}

// Storage handles case persistence
type Storage struct {
	configDir string
	mu        sync.RWMutex
}

// NewStorage creates a new storage instance
func NewStorage() (*Storage, error) {
	configDir, err := getConfigDir()
	if err != nil {
		return nil, err
	}

	// Create config directory if it doesn't exist
	if err := os.MkdirAll(configDir, 0755); err != nil {
		return nil, err
	}

	return &Storage{
		configDir: configDir,
	}, nil
}

// getConfigDir determines the config directory based on the build mode
// - Development (wails dev): ~/.wirey
// - Production (wails build): .wirey next to the exe
func getConfigDir() (string, error) {
	if !IsProduction {
		// Development: use home directory
		homeDir, err := os.UserHomeDir()
		if err != nil {
			return "", err
		}
		return filepath.Join(homeDir, ".wirey"), nil
	}

	// Production: use .wirey folder next to exe
	exePath, err := os.Executable()
	if err != nil {
		// Fallback to home directory
		homeDir, err := os.UserHomeDir()
		if err != nil {
			return "", err
		}
		return filepath.Join(homeDir, ".wirey"), nil
	}

	exeDir := filepath.Dir(exePath)
	return filepath.Join(exeDir, ".wirey"), nil
}

// LoadCollections loads all collections with their cases
func (s *Storage) LoadCollections() ([]CollectionWithCases, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var result []CollectionWithCases

	// List all directories in configDir
	entries, err := os.ReadDir(s.configDir)
	if err != nil {
		return nil, err
	}

	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		collectionDir := filepath.Join(s.configDir, entry.Name())
		collectionFile := filepath.Join(collectionDir, "collection.json")

		// Check if collection.json exists
		if _, err := os.Stat(collectionFile); os.IsNotExist(err) {
			continue
		}

		// Load collection metadata
		collection, err := s.loadCollectionFile(collectionFile)
		if err != nil {
			continue
		}

		// Load all cases in this collection
		cases, err := s.loadCasesInDir(collectionDir)
		if err != nil {
			cases = []SavedCase{}
		}

		result = append(result, CollectionWithCases{
			Collection: collection,
			Cases:      cases,
		})
	}

	// Sort collections by Order
	sort.Slice(result, func(i, j int) bool {
		return result[i].Collection.Order < result[j].Collection.Order
	})

	return result, nil
}

// CreateCollection creates a new collection
func (s *Storage) CreateCollection(name string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	collectionDir := filepath.Join(s.configDir, name)

	// Create collection directory
	if err := os.MkdirAll(collectionDir, 0755); err != nil {
		return err
	}

	// Create collection.json
	now := time.Now().UTC().Format(time.RFC3339)
	collection := Collection{
		Name:      name,
		CreatedAt: now,
		UpdatedAt: now,
	}

	return s.saveCollectionFile(filepath.Join(collectionDir, "collection.json"), collection)
}

// UpdateCollection updates a collection's metadata
func (s *Storage) UpdateCollection(name string, col Collection) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	collectionFile := filepath.Join(s.configDir, name, "collection.json")

	// Update timestamp
	col.UpdatedAt = time.Now().UTC().Format(time.RFC3339)

	return s.saveCollectionFile(collectionFile, col)
}

// DeleteCollection deletes a collection and all its cases
func (s *Storage) DeleteCollection(name string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	collectionDir := filepath.Join(s.configDir, name)
	return os.RemoveAll(collectionDir)
}

// RenameCollection renames a collection folder
func (s *Storage) RenameCollection(oldName, newName string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	oldDir := filepath.Join(s.configDir, oldName)
	newDir := filepath.Join(s.configDir, newName)

	// Rename directory
	if err := os.Rename(oldDir, newDir); err != nil {
		return err
	}

	// Update collection.json with new name
	collectionFile := filepath.Join(newDir, "collection.json")
	collection, err := s.loadCollectionFile(collectionFile)
	if err != nil {
		return err
	}

	collection.Name = newName
	collection.UpdatedAt = time.Now().UTC().Format(time.RFC3339)

	return s.saveCollectionFile(collectionFile, collection)
}

// SaveCase saves a case to a collection
func (s *Storage) SaveCase(collectionName string, c SavedCase) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Ensure collection directory exists
	collectionDir := filepath.Join(s.configDir, collectionName)
	if _, err := os.Stat(collectionDir); os.IsNotExist(err) {
		return os.ErrNotExist
	}

	// Update timestamp
	c.UpdatedAt = time.Now().UTC().Format(time.RFC3339)
	if c.CreatedAt == "" {
		c.CreatedAt = c.UpdatedAt
	}

	// Save case file (use ID as filename)
	caseFile := filepath.Join(collectionDir, s.sanitizeFilename(c.ID)+".json")

	data, err := json.MarshalIndent(c, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(caseFile, data, 0644)
}

// DeleteCase deletes a case from a collection
func (s *Storage) DeleteCase(collectionName string, caseID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	caseFile := filepath.Join(s.configDir, collectionName, s.sanitizeFilename(caseID)+".json")
	return os.Remove(caseFile)
}

// LoadCase loads a single case from a collection
func (s *Storage) LoadCase(collectionName string, caseID string) (SavedCase, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	caseFile := filepath.Join(s.configDir, collectionName, s.sanitizeFilename(caseID)+".json")
	data, err := os.ReadFile(caseFile)
	if err != nil {
		return SavedCase{}, err
	}

	var c SavedCase
	if err := json.Unmarshal(data, &c); err != nil {
		return SavedCase{}, err
	}

	return c, nil
}

// MoveCase moves a case from one collection to another
func (s *Storage) MoveCase(fromCollection, toCollection, caseID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	srcFile := filepath.Join(s.configDir, fromCollection, s.sanitizeFilename(caseID)+".json")
	dstFile := filepath.Join(s.configDir, toCollection, s.sanitizeFilename(caseID)+".json")

	// Read the case
	data, err := os.ReadFile(srcFile)
	if err != nil {
		return err
	}

	// Write to new location
	if err := os.WriteFile(dstFile, data, 0644); err != nil {
		return err
	}

	// Remove from old location
	return os.Remove(srcFile)
}

// Helper functions

func (s *Storage) loadCollectionFile(path string) (Collection, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return Collection{}, err
	}

	var collection Collection
	if err := json.Unmarshal(data, &collection); err != nil {
		return Collection{}, err
	}

	return collection, nil
}

func (s *Storage) saveCollectionFile(path string, collection Collection) error {
	data, err := json.MarshalIndent(collection, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(path, data, 0644)
}

func (s *Storage) loadCasesInDir(dir string) ([]SavedCase, error) {
	var cases []SavedCase

	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		// Skip collection.json
		if entry.Name() == "collection.json" {
			continue
		}

		// Only process .json files
		if !strings.HasSuffix(entry.Name(), ".json") {
			continue
		}

		caseFile := filepath.Join(dir, entry.Name())
		data, err := os.ReadFile(caseFile)
		if err != nil {
			continue
		}

		var c SavedCase
		if err := json.Unmarshal(data, &c); err != nil {
			continue
		}

		cases = append(cases, c)
	}

	// Sort cases by Order
	sort.Slice(cases, func(i, j int) bool {
		return cases[i].Order < cases[j].Order
	})

	return cases, nil
}

func (s *Storage) sanitizeFilename(name string) string {
	// Replace invalid filename characters
	invalid := []string{"/", "\\", ":", "*", "?", "\"", "<", ">", "|"}
	result := name
	for _, char := range invalid {
		result = strings.ReplaceAll(result, char, "_")
	}
	return result
}

// LoadSettings loads global application settings
func (s *Storage) LoadSettings() (AppSettings, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	settingsFile := filepath.Join(s.configDir, "settings.json")

	// Return defaults if file doesn't exist
	if _, err := os.Stat(settingsFile); os.IsNotExist(err) {
		return AppSettings{DefaultCharset: "utf-8"}, nil
	}

	data, err := os.ReadFile(settingsFile)
	if err != nil {
		return AppSettings{DefaultCharset: "utf-8"}, err
	}

	var settings AppSettings
	if err := json.Unmarshal(data, &settings); err != nil {
		return AppSettings{DefaultCharset: "utf-8"}, err
	}

	// Ensure default value
	if settings.DefaultCharset == "" {
		settings.DefaultCharset = "utf-8"
	}

	return settings, nil
}

// SaveSettings saves global application settings
func (s *Storage) SaveSettings(settings AppSettings) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	settingsFile := filepath.Join(s.configDir, "settings.json")

	data, err := json.MarshalIndent(settings, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(settingsFile, data, 0644)
}

// CaseOrder represents the order of a case
type CaseOrder struct {
	ID    string `json:"id"`
	Order int    `json:"order"`
}

// CollectionOrder represents the order of a collection
type CollectionOrder struct {
	Name  string `json:"name"`
	Order int    `json:"order"`
}

// ReorderCollections updates the order of collections
func (s *Storage) ReorderCollections(orders []CollectionOrder) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	for _, o := range orders {
		collectionFile := filepath.Join(s.configDir, o.Name, "collection.json")
		collection, err := s.loadCollectionFile(collectionFile)
		if err != nil {
			continue
		}
		collection.Order = o.Order
		if err := s.saveCollectionFile(collectionFile, collection); err != nil {
			return err
		}
	}
	return nil
}

// ReorderCases updates the order of cases in a collection
func (s *Storage) ReorderCases(collectionName string, orders []CaseOrder) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	collectionDir := filepath.Join(s.configDir, collectionName)

	for _, o := range orders {
		caseFile := filepath.Join(collectionDir, s.sanitizeFilename(o.ID)+".json")
		data, err := os.ReadFile(caseFile)
		if err != nil {
			continue
		}

		var c SavedCase
		if err := json.Unmarshal(data, &c); err != nil {
			continue
		}

		c.Order = o.Order

		newData, err := json.MarshalIndent(c, "", "  ")
		if err != nil {
			continue
		}

		if err := os.WriteFile(caseFile, newData, 0644); err != nil {
			return err
		}
	}
	return nil
}

// IsFirstRun checks if this is the first run (no collections exist)
func (s *Storage) IsFirstRun() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()

	entries, err := os.ReadDir(s.configDir)
	if err != nil {
		return true
	}

	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		collectionFile := filepath.Join(s.configDir, entry.Name(), "collection.json")
		if _, err := os.Stat(collectionFile); err == nil {
			return false
		}
	}
	return true
}

// InitializeSamples creates the "Getting Started" collection with sample cases
func (s *Storage) InitializeSamples() error {
	now := time.Now().UTC().Format(time.RFC3339)

	// Get sample data
	collection := SampleCollection()
	collection.CreatedAt = now
	collection.UpdatedAt = now

	// Create collection directory
	collectionDir := filepath.Join(s.configDir, collection.Name)
	if err := os.MkdirAll(collectionDir, 0755); err != nil {
		return err
	}

	// Save collection metadata
	if err := s.saveCollectionFile(filepath.Join(collectionDir, "collection.json"), collection); err != nil {
		return err
	}

	// Save sample cases
	for _, c := range SampleCases() {
		c.CreatedAt = now
		c.UpdatedAt = now
		caseFile := filepath.Join(collectionDir, s.sanitizeFilename(c.ID)+".json")
		data, err := json.MarshalIndent(c, "", "  ")
		if err != nil {
			return err
		}
		if err := os.WriteFile(caseFile, data, 0644); err != nil {
			return err
		}
	}

	return nil
}
