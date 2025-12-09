package mod

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

// Mod represents a single mod.
type Mod struct {
	Name   string `json:"name"`
	Path   string `json:"path"`
	Active bool   `json:"active"`
	ModID  string `json:"mod_id"`
	FileID string `json:"file_id"`
}

// GetModsConfigPath returns the path to the mods.json file.
func GetModsConfigPath() (string, error) {
	configDir, err := os.UserConfigDir()
	if err != nil {
		return "", fmt.Errorf("failed to get user config directory: %w", err)
	}
	return filepath.Join(configDir, "fusion-core", "mods.json"), nil
}

// LoadMods loads the list of mods from mods.json.
func LoadMods() ([]*Mod, error) {
	modsPath, err := GetModsConfigPath()
	if err != nil {
		return nil, err
	}

	file, err := os.Open(modsPath)
	if err != nil {
		if os.IsNotExist(err) {
			return []*Mod{}, nil // Return empty list if file doesn't exist
		}
		return nil, fmt.Errorf("failed to open mods.json: %w", err)
	}
	defer file.Close()

	var mods []*Mod
	if err := json.NewDecoder(file).Decode(&mods); err != nil {
		return nil, fmt.Errorf("failed to decode mods.json: %w", err)
	}

	return mods, nil
}

// SaveMods saves the list of mods to mods.json.
func SaveMods(mods []*Mod) error {
	modsPath, err := GetModsConfigPath()
	if err != nil {
		return err
	}

	// Create the directory if it doesn't exist.
	if err := os.MkdirAll(filepath.Dir(modsPath), 0755); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	file, err := os.Create(modsPath)
	if err != nil {
		return fmt.Errorf("failed to create mods.json: %w", err)
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(mods); err != nil {
		return fmt.Errorf("failed to encode mods.json: %w", err)
	}

	return nil
}
