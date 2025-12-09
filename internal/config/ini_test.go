package config

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"
)

func TestAddArchiveToCustomIni(t *testing.T) {
	tmpDir, err := ioutil.TempDir("", "test-prefix")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Create the necessary subdirectories for the ini file
	iniDir := filepath.Join(tmpDir, "pfx", "drive_c", "users", "steamuser", "Documents", "My Games", "Fallout 76")
	if err := os.MkdirAll(iniDir, 0755); err != nil {
		t.Fatalf("Failed to create ini directory: %v", err)
	}

	// Test 1: Add a new archive to a non-existent file
	archiveName1 := "TestMod1.ba2"
	err = AddArchiveToCustomIni(tmpDir, archiveName1)
	if err != nil {
		t.Fatalf("Test 1 failed: %v", err)
	}

	iniPath := GetFallout76CustomIniPath(tmpDir)
	content, err := ioutil.ReadFile(iniPath)
	if err != nil {
		t.Fatalf("Test 1 failed: could not read ini file: %v", err)
	}

	expectedContent1 := "[Archive]\nsResourceArchive2List = TestMod1.ba2\n"
	if string(content) != expectedContent1 {
		t.Errorf("Test 1 failed: expected content %q, got %q", expectedContent1, string(content))
	}

	// Test 2: Add a second archive
	archiveName2 := "TestMod2.ba2"
	err = AddArchiveToCustomIni(tmpDir, archiveName2)
	if err != nil {
		t.Fatalf("Test 2 failed: %v", err)
	}

	content, err = ioutil.ReadFile(iniPath)
	if err != nil {
		t.Fatalf("Test 2 failed: could not read ini file: %v", err)
	}

	expectedContent2 := "[Archive]\nsResourceArchive2List = TestMod1.ba2, TestMod2.ba2\n"
	if string(content) != expectedContent2 {
		t.Errorf("Test 2 failed: expected content %q, got %q", expectedContent2, string(content))
	}

	// Test 3: Add a duplicate archive
	err = AddArchiveToCustomIni(tmpDir, archiveName1)
	if err != nil {
		t.Fatalf("Test 3 failed: %v", err)
	}

	content, err = ioutil.ReadFile(iniPath)
	if err != nil {
		t.Fatalf("Test 3 failed: could not read ini file: %v", err)
	}

	if string(content) != expectedContent2 {
		t.Errorf("Test 3 failed: expected content %q, got %q", expectedContent2, string(content))
	}
}
