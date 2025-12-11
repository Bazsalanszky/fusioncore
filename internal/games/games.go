package games

import (
	"fmt"
	"os"
	"path/filepath"
)

// Game represents a supported game with its configuration
type Game struct {
	ID          string
	Name        string
	AppID       string
	NexusName   string
	GameDir     string
	DataSubDir  string
	ConfigFile  string
	PluginsFile string
	ArchiveExt  string
}

// GetSupportedGames returns all supported games
func GetSupportedGames() []Game {
	return []Game{
		{
			ID:          "fallout76",
			Name:        "Fallout 76",
			AppID:       "1151340",
			NexusName:   "fallout76",
			GameDir:     "Fallout76",
			DataSubDir:  "Data",
			ConfigFile:  "Fallout76Custom.ini",
			PluginsFile: "plugins.txt",
			ArchiveExt:  ".ba2",
		},
		{
			ID:          "fallout4",
			Name:        "Fallout 4",
			AppID:       "377160",
			NexusName:   "fallout4",
			GameDir:     "Fallout 4",
			DataSubDir:  "Data",
			ConfigFile:  "Fallout4Custom.ini",
			PluginsFile: "plugins.txt",
			ArchiveExt:  ".ba2",
		},
		{
			ID:          "fallout3",
			Name:        "Fallout 3",
			AppID:       "22300",
			NexusName:   "fallout3",
			GameDir:     "Fallout 3 goty",
			DataSubDir:  "Data",
			ConfigFile:  "Fallout.ini",
			PluginsFile: "plugins.txt",
			ArchiveExt:  ".bsa",
		},
		{
			ID:          "falloutnv",
			Name:        "Fallout: New Vegas",
			AppID:       "22380",
			NexusName:   "newvegas",
			GameDir:     "Fallout New Vegas",
			DataSubDir:  "Data",
			ConfigFile:  "Fallout.ini",
			PluginsFile: "plugins.txt",
			ArchiveExt:  ".bsa",
		},
		{
			ID:          "skyrim",
			Name:        "The Elder Scrolls V: Skyrim",
			AppID:       "72850",
			NexusName:   "skyrim",
			GameDir:     "Skyrim",
			DataSubDir:  "Data",
			ConfigFile:  "Skyrim.ini",
			PluginsFile: "plugins.txt",
			ArchiveExt:  ".bsa",
		},
		{
			ID:          "skyrimse",
			Name:        "The Elder Scrolls V: Skyrim Special Edition",
			AppID:       "489830",
			NexusName:   "skyrimspecialedition",
			GameDir:     "Skyrim Special Edition",
			DataSubDir:  "Data",
			ConfigFile:  "Skyrim.ini",
			PluginsFile: "plugins.txt",
			ArchiveExt:  ".bsa",
		},
	}
}

// GetGameByID returns a game by its ID
func GetGameByID(id string) (*Game, error) {
	for _, game := range GetSupportedGames() {
		if game.ID == id {
			return &game, nil
		}
	}
	return nil, fmt.Errorf("game not found: %s", id)
}

// GetGameByNexusName returns a game by its Nexus name
func GetGameByNexusName(nexusName string) (*Game, error) {
	for _, game := range GetSupportedGames() {
		if game.NexusName == nexusName {
			return &game, nil
		}
	}
	return nil, fmt.Errorf("game not found for nexus name: %s", nexusName)
}

// FindSteamRoot finds the root directory of the Steam installation
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

	return "", fmt.Errorf("Steam root directory not found")
}

// FindGameDir finds the game directory for a specific game
func (g *Game) FindGameDir() (string, error) {
	steamRoot, err := FindSteamRoot()
	if err != nil {
		return "", err
	}

	gameDir := filepath.Join(steamRoot, "steamapps", "common", g.GameDir)
	if _, err := os.Stat(gameDir); err == nil {
		return gameDir, nil
	}

	return "", fmt.Errorf("%s game directory not found", g.Name)
}

// FindDataDir finds the Data directory for a specific game
func (g *Game) FindDataDir() (string, error) {
	gameDir, err := g.FindGameDir()
	if err != nil {
		return "", err
	}

	dataDir := filepath.Join(gameDir, g.DataSubDir)
	if _, err := os.Stat(dataDir); err == nil {
		return dataDir, nil
	}

	return "", fmt.Errorf("%s Data directory not found", g.Name)
}

// FindCompatdata finds the compatdata directory for a specific game
func (g *Game) FindCompatdata() (string, error) {
	steamRoot, err := FindSteamRoot()
	if err != nil {
		return "", err
	}

	compatdataPath := filepath.Join(steamRoot, "steamapps", "compatdata", g.AppID)
	if _, err := os.Stat(compatdataPath); err == nil {
		return compatdataPath, nil
	}

	return "", fmt.Errorf("compatdata directory not found for %s", g.Name)
}

// GetModsDir returns the mods directory for a specific game
func (g *Game) GetModsDir() (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(homeDir, "Games", "FusionCore", "Mods", g.Name), nil
}