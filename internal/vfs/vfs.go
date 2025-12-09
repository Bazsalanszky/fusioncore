package vfs

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/bazsalanszky/fusioncore/internal/config"
	"github.com/bazsalanszky/fusioncore/internal/mod"
	"github.com/bazsalanszky/fusioncore/internal/prefix"
)

// SyncLinks creates symlinks for all active mods.
func SyncLinks() error {
	mods, err := mod.LoadMods()
	if err != nil {
		return fmt.Errorf("failed to load mods: %w", err)
	}

	dataDir, err := prefix.FindFallout76DataDir()
	if err != nil {
		return fmt.Errorf("failed to find Fallout 76 data directory: %w", err)
	}

	// First, remove all existing symlinks to avoid dangling links
	files, err := os.ReadDir(dataDir)
	if err != nil {
		return fmt.Errorf("failed to read data directory: %w", err)
	}
	for _, file := range files {
		if file.Type()&os.ModeSymlink != 0 {
			symlinkPath := filepath.Join(dataDir, file.Name())
			if err := os.Remove(symlinkPath); err != nil {
				return fmt.Errorf("failed to remove existing symlink at %s: %w", symlinkPath, err)
			}
		}
	}

	// Then, create symlinks for all active mods
	for _, m := range mods {
		if m.Active {
			ba2Files, err := findBa2Files(m.Path)
			if err != nil {
				return fmt.Errorf("failed to find .ba2 files in mod %s: %w", m.Name, err)
			}
			for _, ba2File := range ba2Files {
				symlinkPath := filepath.Join(dataDir, ba2File)
				targetPath := filepath.Join(m.Path, ba2File)
				if err := os.Symlink(targetPath, symlinkPath); err != nil {
					return fmt.Errorf("failed to create symlink for %s: %w", ba2File, err)
				}
				fmt.Printf("Created symlink for %s\n", ba2File)
			}

			/*symlinkPath := filepath.Join(dataDir, m.Name)
			if err := os.Symlink(m.Path, symlinkPath); err != nil {
				return fmt.Errorf("failed to create symlink for %s: %w", m.Name, err)
			}
			fmt.Printf("Created symlink for %s\n", m.Name)*/
		}
	}

	return nil
}

// findBa2Files finds all .ba2 files in a directory.
func findBa2Files(dir string) ([]string, error) {
	var ba2Files []string
	err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() && filepath.Ext(path) == ".ba2" {
			ba2Files = append(ba2Files, info.Name())
		}
		return nil
	})
	return ba2Files, err
}

// Activate activates a mod.
func Activate(modName string) error {
	mods, err := mod.LoadMods()
	if err != nil {
		return err
	}

	for _, m := range mods {
		if m.Name == modName {
			m.Active = true
			if err := mod.SaveMods(mods); err != nil {
				return err
			}

			ba2Files, err := findBa2Files(m.Path)
			if err != nil {
				return err
			}
			for _, ba2File := range ba2Files {
				if err := config.AddArchiveToCustomIniWithPrefix(ba2File); err != nil {
					return err
				}
			}

			return SyncLinks()
		}
	}

	return fmt.Errorf("mod not found: %s", modName)
}

// Deactivate deactivates a mod.
func Deactivate(modName string) error {
	mods, err := mod.LoadMods()
	if err != nil {
		return err
	}

	for _, m := range mods {
		if m.Name == modName {
			m.Active = false
			if err := mod.SaveMods(mods); err != nil {
				return err
			}

			ba2Files, err := findBa2Files(m.Path)
			if err != nil {
				return err
			}
			for _, ba2File := range ba2Files {
				if err := config.RemoveArchiveFromCustomIniWithPrefix(ba2File); err != nil {
					return err
				}
			}

			return SyncLinks()
		}
	}

	return fmt.Errorf("mod not found: %s", modName)
}
