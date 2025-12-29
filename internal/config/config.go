package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

// Config holds the application's configuration.
type Config struct {
	APIKey          string            `json:"api_key"`
	CurrentGame     string            `json:"current_game"`
	GamePaths       map[string]string `json:"game_paths,omitempty"`       // Maps game ID to custom game directory path
	CompatdataPaths map[string]string `json:"compatdata_paths,omitempty"` // Maps game ID to custom compatdata directory path
}

// GetConfigPath returns the path to the configuration file.
func GetConfigPath() (string, error) {
	configDir, err := os.UserConfigDir()
	if err != nil {
		return "", fmt.Errorf("failed to get user config directory: %w", err)
	}
	return filepath.Join(configDir, "fusion-core", "config.json"), nil
}

// LoadConfig loads the application's configuration from the config file.
func LoadConfig() (*Config, error) {
	configPath, err := GetConfigPath()
	if err != nil {
		return nil, err
	}

	file, err := os.Open(configPath)
	if err != nil {
		if os.IsNotExist(err) {
			return &Config{
				CurrentGame:     "fallout76",
				GamePaths:       make(map[string]string),
				CompatdataPaths: make(map[string]string),
			}, nil // Return empty config with default game
		}
		return nil, fmt.Errorf("failed to open config file: %w", err)
	}
	defer file.Close()

	var config Config
	if err := json.NewDecoder(file).Decode(&config); err != nil {
		return nil, fmt.Errorf("failed to decode config file: %w", err)
	}

	// Set default game if not set
	if config.CurrentGame == "" {
		config.CurrentGame = "fallout76"
	}

	// Initialize GamePaths if nil
	if config.GamePaths == nil {
		config.GamePaths = make(map[string]string)
	}

	// Initialize CompatdataPaths if nil
	if config.CompatdataPaths == nil {
		config.CompatdataPaths = make(map[string]string)
	}

	return &config, nil
}

// SaveConfig saves the application's configuration to the config file.
func SaveConfig(config *Config) error {
	configPath, err := GetConfigPath()
	if err != nil {
		return err
	}

	// Create the directory if it doesn't exist.
	if err := os.MkdirAll(filepath.Dir(configPath), 0755); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	file, err := os.Create(configPath)
	if err != nil {
		return fmt.Errorf("failed to create config file: %w", err)
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(config); err != nil {
		return fmt.Errorf("failed to encode config file: %w", err)
	}

	return nil
}
