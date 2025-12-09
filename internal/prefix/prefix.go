package prefix

import (
	"errors"
	"os"
	"path/filepath"
)

// FindSteamRoot finds the root directory of the Steam installation.
// It checks common locations for the Steam installation on Linux.
func FindSteamRoot() (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}

	possiblePaths := []string{
		filepath.Join(homeDir, ".steam", "steam"),
		filepath.Join(homeDir, ".local", "share", "Steam"),
	}

	for _, path := range possiblePaths {
		if _, err := os.Stat(path); err == nil {
			return path, nil
		}
	}

	return "", errors.New("Steam root directory not found")
}

// FindCompatdata finds the compatdata directory for a given AppID.
func FindCompatdata(steamRoot, appID string) (string, error) {
	compatdataPath := filepath.Join(steamRoot, "steamapps", "compatdata", appID)
	if _, err := os.Stat(compatdataPath); err == nil {
		return compatdataPath, nil
	}

	return "", errors.New("compatdata directory not found for AppID: " + appID)
}

// FindFallout76Prefix finds the Proton prefix for Fallout 76.
func FindFallout76Prefix() (string, error) {
	steamRoot, err := FindSteamRoot()
	if err != nil {
		return "", err
	}

	return FindCompatdata(steamRoot, "1151340")
}

// FindFallout76GameDir finds the game directory for Fallout 76.
// This function currently only checks the default Steam library.
func FindFallout76GameDir() (string, error) {
	steamRoot, err := FindSteamRoot()
	if err != nil {
		return "", err
	}

	gameDir := filepath.Join(steamRoot, "steamapps", "common", "Fallout76")
	if _, err := os.Stat(gameDir); err == nil {
		return gameDir, nil
	}

	return "", errors.New("Fallout 76 game directory not found in default library")
}

// FindFallout76DataDir finds the Data directory for Fallout 76.
func FindFallout76DataDir() (string, error) {
	gameDir, err := FindFallout76GameDir()
	if err != nil {
		return "", err
	}

	dataDir := filepath.Join(gameDir, "Data")
	if _, err := os.Stat(dataDir); err == nil {
		return dataDir, nil
	}

	return "", errors.New("Fallout 76 Data directory not found")
}
