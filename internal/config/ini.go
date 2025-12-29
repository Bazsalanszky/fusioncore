package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/bazsalanszky/fusioncore/internal/games"
	"gopkg.in/ini.v1"
)

// GetCustomIniPath returns the path to the custom ini file for any game.
func GetCustomIniPath(prefixPath, configFile string) string {
	return filepath.Join(prefixPath, "pfx", "drive_c", "users", "steamuser", "Documents", "My Games", "Fallout 76", configFile)
}

// GetFallout76CustomIniPath returns the path to the Fallout76Custom.ini file.
func GetFallout76CustomIniPath(prefixPath string) string {
	return GetCustomIniPath(prefixPath, "Fallout76Custom.ini")
}

// GetFallout76CustomIniPathWithPrefix finds the prefix and returns the path to the Fallout76Custom.ini file.
func GetFallout76CustomIniPathWithPrefix() (string, error) {
	cfg, err := LoadConfig()
	if err != nil {
		return "", err
	}
	game, err := games.GetGameByID("fallout76")
	if err != nil {
		return "", err
	}

	// Get custom compatdata path if configured
	customPath := ""
	if cfg.CompatdataPaths != nil {
		customPath = cfg.CompatdataPaths[game.ID]
	}

	prefixPath, err := game.FindCompatdataWithCustomPath(customPath)
	if err != nil {
		return "", fmt.Errorf("failed to find Fallout 76 prefix: %w", err)
	}
	return GetFallout76CustomIniPath(prefixPath), nil
}

// AddArchiveToCustomIni adds a new archive to the sResourceArchive2List in Fallout76Custom.ini.
func AddArchiveToCustomIni(prefixPath, archiveName string) error {
	iniPath := GetFallout76CustomIniPath(prefixPath)
	cfg, err := ini.Load(iniPath)
	if err != nil {
		if os.IsNotExist(err) {
			cfg = ini.Empty()
		} else {
			return fmt.Errorf("failed to load Fallout76Custom.ini: %w", err)
		}
	}

	section := cfg.Section("Archive")
	key := section.Key("sResourceArchive2List")
	currentValue := key.String()

	if currentValue == "" {
		key.SetValue(archiveName)
	} else {
		// Avoid adding duplicate entries
		if !strings.Contains(currentValue, archiveName) {
			key.SetValue(fmt.Sprintf("%s, %s", currentValue, archiveName))
		}
	}

	return cfg.SaveTo(iniPath)

}

// RemoveArchiveFromCustomIni removes an archive from the sResourceArchive2List in Fallout76Custom.ini.

func RemoveArchiveFromCustomIni(prefixPath, archiveName string) error {

	iniPath := GetFallout76CustomIniPath(prefixPath)

	cfg, err := ini.Load(iniPath)

	if err != nil {

		return fmt.Errorf("failed to load Fallout76Custom.ini: %w", err)

	}

	section := cfg.Section("Archive")

	key := section.Key("sResourceArchive2List")

	currentValue := key.String()

	if strings.Contains(currentValue, archiveName) {

		newValue := strings.ReplaceAll(currentValue, ", "+archiveName, "")

		newValue = strings.ReplaceAll(newValue, archiveName+", ", "")

		newValue = strings.ReplaceAll(newValue, archiveName, "")

		key.SetValue(newValue)

	}

	return cfg.SaveTo(iniPath)

}

// AddArchiveToCustomIniWithPrefix finds the prefix and adds a new archive to the sResourceArchive2List.
func AddArchiveToCustomIniWithPrefix(archiveName string) error {
	cfg, err := LoadConfig()
	if err != nil {
		return err
	}
	game, err := games.GetGameByID(cfg.CurrentGame)
	if err != nil {
		return err
	}

	// Get custom compatdata path if configured
	customPath := ""
	if cfg.CompatdataPaths != nil {
		customPath = cfg.CompatdataPaths[cfg.CurrentGame]
	}

	prefixPath, err := game.FindCompatdataWithCustomPath(customPath)
	if err != nil {
		return err
	}
	return AddArchiveToCustomIni(prefixPath, archiveName)
}

// RemoveArchiveFromCustomIniWithPrefix finds the prefix and removes an archive from the sResourceArchive2List.

func RemoveArchiveFromCustomIniWithPrefix(archiveName string) error {
	cfg, err := LoadConfig()
	if err != nil {
		return err
	}
	game, err := games.GetGameByID(cfg.CurrentGame)
	if err != nil {
		return err
	}

	// Get custom compatdata path if configured
	customPath := ""
	if cfg.CompatdataPaths != nil {
		customPath = cfg.CompatdataPaths[cfg.CurrentGame]
	}

	prefixPath, err := game.FindCompatdataWithCustomPath(customPath)
	if err != nil {
		return err
	}
	return RemoveArchiveFromCustomIni(prefixPath, archiveName)
}
