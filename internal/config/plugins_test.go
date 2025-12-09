package config

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

func TestPlugins(t *testing.T) {
	tmpDir, err := ioutil.TempDir("", "test-prefix-plugins")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	pluginsDir := filepath.Join(tmpDir, "pfx", "drive_c", "users", "steamuser", "AppData", "Local", "Fallout76")
	if err := os.MkdirAll(pluginsDir, 0755); err != nil {
		t.Fatalf("Failed to create plugins directory: %v", err)
	}

	pluginsPath := GetPluginsTxtPath(tmpDir)

	// Test 1: Read from non-existent file
	plugins, err := ReadPlugins(tmpDir)
	if err != nil {
		t.Fatalf("Test 1 failed: %v", err)
	}
	if len(plugins) != 0 {
		t.Errorf("Test 1 failed: expected 0 plugins, got %d", len(plugins))
	}

	// Test 2: Add a plugin
	plugin1 := "TestMod1.esp"
	err = AddPlugin(tmpDir, plugin1)
	if err != nil {
		t.Fatalf("Test 2 failed: %v", err)
	}

	plugins, err = ReadPlugins(tmpDir)
	if err != nil {
		t.Fatalf("Test 2 failed: %v", err)
	}
	expectedPlugins := []string{plugin1}
	if !reflect.DeepEqual(plugins, expectedPlugins) {
		t.Errorf("Test 2 failed: expected %v, got %v", expectedPlugins, plugins)
	}

	// Test 3: Add another plugin
	plugin2 := "TestMod2.esl"
	err = AddPlugin(tmpDir, plugin2)
	if err != nil {
		t.Fatalf("Test 3 failed: %v", err)
	}

	plugins, err = ReadPlugins(tmpDir)
	if err != nil {
		t.Fatalf("Test 3 failed: %v", err)
	}
	expectedPlugins = []string{plugin1, plugin2}
	if !reflect.DeepEqual(plugins, expectedPlugins) {
		t.Errorf("Test 3 failed: expected %v, got %v", expectedPlugins, plugins)
	}

	// Test 4: Add a duplicate plugin
	err = AddPlugin(tmpDir, plugin1)
	if err != nil {
		t.Fatalf("Test 4 failed: %v", err)
	}

	plugins, err = ReadPlugins(tmpDir)
	if err != nil {
		t.Fatalf("Test 4 failed: %v", err)
	}
	if !reflect.DeepEqual(plugins, expectedPlugins) {
		t.Errorf("Test 4 failed: expected %v, got %v", expectedPlugins, plugins)
	}

	// Test 5: Read a file with comments and empty lines
	fileContent := "# This is a comment\n\nTestMod1.esp\nTestMod2.esl\n"
	err = ioutil.WriteFile(pluginsPath, []byte(fileContent), 0644)
	if err != nil {
		t.Fatalf("Test 5 failed: could not write to plugins.txt: %v", err)
	}
	plugins, err = ReadPlugins(tmpDir)
	if err != nil {
		t.Fatalf("Test 5 failed: %v", err)
	}
	expectedPlugins = []string{"TestMod1.esp", "TestMod2.esl"}
	if !reflect.DeepEqual(plugins, expectedPlugins) {
		t.Errorf("Test 5 failed: expected %v, got %v", expectedPlugins, plugins)
	}
}
