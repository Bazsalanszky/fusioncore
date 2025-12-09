package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/bazsalanszky/fusioncore/internal/prefix"
	"gopkg.in/ini.v1"
)

// GetFallout76CustomIniPath returns the path to the Fallout76Custom.ini file.
func GetFallout76CustomIniPath(prefixPath string) string {
	return filepath.Join(prefixPath, "pfx", "drive_c", "users", "steamuser", "Documents", "My Games", "Fallout 76", "Fallout76Custom.ini")
}

// GetFallout76CustomIniPathWithPrefix finds the prefix and returns the path to the Fallout76Custom.ini file.
func GetFallout76CustomIniPathWithPrefix() (string, error) {
	prefixPath, err := prefix.FindFallout76Prefix()
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

	

	// AddArchiveToCustomIniWithPrefix finds the prefix and adds a new archive to the sResourceArchive2List in Fallout76Custom.ini.

	func AddArchiveToCustomIniWithPrefix(archiveName string) error {

		prefixPath, err := prefix.FindFallout76Prefix()

		if err != nil {

			return err

		}

		return AddArchiveToCustomIni(prefixPath, archiveName)

	}

	

	// RemoveArchiveFromCustomIniWithPrefix finds the prefix and removes an archive from the sResourceArchive2List in Fallout76Custom.ini.

	func RemoveArchiveFromCustomIniWithPrefix(archiveName string) error {

		prefixPath, err := prefix.FindFallout76Prefix()

		if err != nil {

			return err

		}

		return RemoveArchiveFromCustomIni(prefixPath, archiveName)

	}

	