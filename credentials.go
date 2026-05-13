package launchpad

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

// Credentials bundles the consumer key and access token needed to sign
// API requests. It can be persisted to and loaded from a JSON file.
type Credentials struct {
	ConsumerKey string       `json:"consumer_key"`
	Token       *AccessToken `json:"access_token"`
}

// DefaultCredentialsDir returns ~/.config/launchpad.
func DefaultCredentialsDir() (string, error) {
	cfgDir, err := os.UserConfigDir()
	if err != nil {
		return "", fmt.Errorf("launchpad: config dir: %w", err)
	}
	return filepath.Join(cfgDir, "launchpad"), nil
}

// DefaultCredentialsPath returns ~/.config/launchpad/<consumerKey>.json.
func DefaultCredentialsPath(consumerKey string) (string, error) {
	dir, err := DefaultCredentialsDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, consumerKey+".json"), nil
}

// Save writes the credentials to path as JSON.
// Parent directories are created with mode 0700 if they don't exist.
// The file is written with mode 0600.
func (c *Credentials) Save(path string) error {
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0700); err != nil {
		return fmt.Errorf("launchpad: creating credentials dir: %w", err)
	}

	data, err := json.MarshalIndent(c, "", "  ")
	if err != nil {
		return fmt.Errorf("launchpad: marshalling credentials: %w", err)
	}

	if err := os.WriteFile(path, data, 0600); err != nil {
		return fmt.Errorf("launchpad: writing credentials: %w", err)
	}
	return nil
}

// LoadCredentials reads credentials from a JSON file at path.
func LoadCredentials(path string) (*Credentials, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("launchpad: reading credentials: %w", err)
	}

	var creds Credentials
	if err := json.Unmarshal(data, &creds); err != nil {
		return nil, fmt.Errorf("launchpad: parsing credentials: %w", err)
	}
	return &creds, nil
}
