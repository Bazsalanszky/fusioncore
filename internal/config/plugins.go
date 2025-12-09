package config

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"

	"github.com/bazsalanszky/fusioncore/internal/prefix"
)

// GetPluginsTxtPath returns the path to the plugins.txt file.
func GetPluginsTxtPath(prefixPath string) string {
	return filepath.Join(prefixPath, "pfx", "drive_c", "users", "steamuser", "AppData", "Local", "Fallout76", "plugins.txt")
}

// ReadPlugins reads the list of plugins from plugins.txt.
func ReadPlugins(prefixPath string) ([]string, error) {
	pluginsPath := GetPluginsTxtPath(prefixPath)
	file, err := os.Open(pluginsPath)
	if err != nil {
		if os.IsNotExist(err) {
			return []string{}, nil
		}
		return nil, fmt.Errorf("failed to open plugins.txt: %w", err)
	}
	defer file.Close()

	var plugins []string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		// Ignore comments and empty lines
		if len(line) > 0 && line[0] != '#' {
			plugins = append(plugins, line)
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("failed to read plugins.txt: %w", err)
	}

	return plugins, nil
}

// WritePlugins writes the list of plugins to plugins.txt.
func WritePlugins(prefixPath string, plugins []string) error {
	pluginsPath := GetPluginsTxtPath(prefixPath)
	file, err := os.Create(pluginsPath)
	if err != nil {
		return fmt.Errorf("failed to create plugins.txt: %w", err)
	}
	defer file.Close()

	writer := bufio.NewWriter(file)
	for _, plugin := range plugins {
		_, err := writer.WriteString(plugin + "\n")
		if err != nil {
			return fmt.Errorf("failed to write to plugins.txt: %w", err)
		}
	}

	return writer.Flush()
}

// AddPlugin adds a new plugin to plugins.txt.
func AddPlugin(prefixPath, pluginName string) error {
	plugins, err := ReadPlugins(prefixPath)
	if err != nil {
		return err
	}

	// Avoid adding duplicate plugins
	for _, p := range plugins {
		if p == pluginName {
			return nil // Plugin already exists
		}
	}

	plugins = append(plugins, pluginName)
	return WritePlugins(prefixPath, plugins)
}

// AddPluginWithPrefix finds the prefix and adds a new plugin to plugins.txt.
func AddPluginWithPrefix(pluginName string) error {
    prefixPath, err := prefix.FindFallout76Prefix()
    if err != nil {
        return err
    }
    return AddPlugin(prefixPath, pluginName)
}
