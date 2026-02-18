package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"sync"
)

// Config holds all user-configurable settings for the application.
// Persisted as JSON in the OS-appropriate app data directory.
type Config struct {
	PrivacyMode bool   `json:"privacy_mode"` // Hide file names from Discord presence
	CustomLabel string `json:"custom_label"` // Text shown instead of file name when privacy is on
	RPCEnabled  bool   `json:"rpc_enabled"`  // Whether Discord RPC connection is active
	FirstRun    bool   `json:"first_run"`    // Show settings window on first launch
	mu          sync.Mutex
}

// DefaultConfig returns sensible defaults for a fresh install.
func DefaultConfig() *Config {
	return &Config{
		PrivacyMode: false,
		CustomLabel: "Working on a project",
		RPCEnabled:  true,
		FirstRun:    true,
	}
}

// configDir returns the OS-appropriate directory for storing config files.
// Windows: %APPDATA%/FigmaRPC/
// macOS:   ~/.config/figma-rpc/
func configDir() (string, error) {
	switch runtime.GOOS {
	case "windows":
		appData := os.Getenv("APPDATA")
		if appData == "" {
			return "", fmt.Errorf("APPDATA environment variable not set")
		}
		return filepath.Join(appData, "FigmaRPC"), nil
	case "darwin":
		home, err := os.UserHomeDir()
		if err != nil {
			return "", fmt.Errorf("could not determine home directory: %w", err)
		}
		return filepath.Join(home, ".config", "figma-rpc"), nil
	default:
		home, err := os.UserHomeDir()
		if err != nil {
			return "", fmt.Errorf("could not determine home directory: %w", err)
		}
		return filepath.Join(home, ".config", "figma-rpc"), nil
	}
}

// configPath returns the full path to the config JSON file.
func configPath() (string, error) {
	dir, err := configDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, "config.json"), nil
}

// LoadConfig reads config from disk. If the file doesn't exist,
// it creates one with defaults and returns it.
func LoadConfig() (*Config, error) {
	path, err := configPath()
	if err != nil {
		return DefaultConfig(), err
	}

	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			// First time: create default config and save it
			cfg := DefaultConfig()
			if saveErr := cfg.Save(); saveErr != nil {
				fmt.Println("Warning: could not save default config:", saveErr)
			}
			return cfg, nil
		}
		return DefaultConfig(), fmt.Errorf("could not read config: %w", err)
	}

	cfg := DefaultConfig()
	if err := json.Unmarshal(data, cfg); err != nil {
		return DefaultConfig(), fmt.Errorf("could not parse config: %w", err)
	}

	return cfg, nil
}

// Save writes the current config to disk as formatted JSON.
func (c *Config) Save() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	path, err := configPath()
	if err != nil {
		return err
	}

	// Ensure the directory exists
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("could not create config directory: %w", err)
	}

	data, err := json.MarshalIndent(c, "", "  ")
	if err != nil {
		return fmt.Errorf("could not serialize config: %w", err)
	}

	if err := os.WriteFile(path, data, 0644); err != nil {
		return fmt.Errorf("could not write config: %w", err)
	}

	return nil
}

// SetPrivacyMode updates the privacy mode setting and saves.
func (c *Config) SetPrivacyMode(enabled bool) error {
	c.PrivacyMode = enabled
	return c.Save()
}

// SetCustomLabel updates the custom label and saves.
func (c *Config) SetCustomLabel(label string) error {
	c.CustomLabel = label
	return c.Save()
}

// SetRPCEnabled updates the RPC toggle and saves.
func (c *Config) SetRPCEnabled(enabled bool) error {
	c.RPCEnabled = enabled
	return c.Save()
}

// SetFirstRun updates the first-run flag and saves.
func (c *Config) SetFirstRun(firstRun bool) error {
	c.FirstRun = firstRun
	return c.Save()
}
