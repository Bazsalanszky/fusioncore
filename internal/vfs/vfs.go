package vfs

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/bazsalanszky/fusioncore/internal/config"
	"github.com/bazsalanszky/fusioncore/internal/games"
	"github.com/bazsalanszky/fusioncore/internal/mod"
)

// SyncLinks creates symlinks for all active mods.
func SyncLinks() error {
	cfg, err := config.LoadConfig()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	game, err := games.GetGameByID(cfg.CurrentGame)
	if err != nil {
		return fmt.Errorf("failed to get current game: %w", err)
	}

	mods, err := mod.LoadMods(cfg.CurrentGame)
	if err != nil {
		return fmt.Errorf("failed to load mods: %w", err)
	}

	// Get custom path if configured
	customPath := ""
	if cfg.GamePaths != nil {
		customPath = cfg.GamePaths[cfg.CurrentGame]
	}

	dataDir, err := game.FindDataDirWithCustomPath(customPath)
	if err != nil {
		return fmt.Errorf("failed to find %s data directory: %w", game.Name, err)
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
			archiveFiles, err := findArchiveFiles(m.Path, game.ArchiveExt)
			if err != nil {
				return fmt.Errorf("failed to find %s files in mod %s: %w", game.ArchiveExt, m.Name, err)
			}
			for _, ba2File := range archiveFiles {
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

// findArchiveFiles finds all archive files with the given extension in a directory.
func findArchiveFiles(dir, ext string) ([]string, error) {
	var archiveFiles []string
	err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() && filepath.Ext(path) == ext {
			archiveFiles = append(archiveFiles, info.Name())
		}
		return nil
	})
	return archiveFiles, err
}

// Activate activates a mod.
func Activate(modName string) error {
	cfg, err := config.LoadConfig()
	if err != nil {
		return err
	}

	game, err := games.GetGameByID(cfg.CurrentGame)
	if err != nil {
		return err
	}

	mods, err := mod.LoadMods(cfg.CurrentGame)
	if err != nil {
		return err
	}

	for _, m := range mods {
		if m.Name == modName {
			m.Active = true
			if err := mod.SaveMods(mods, cfg.CurrentGame); err != nil {
				return err
			}

			archiveFiles, err := findArchiveFiles(m.Path, game.ArchiveExt)
			if err != nil {
				return err
			}
			for _, archiveFile := range archiveFiles {
				if err := config.AddArchiveToCustomIniWithPrefix(archiveFile); err != nil {
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
	cfg, err := config.LoadConfig()
	if err != nil {
		return err
	}

	game, err := games.GetGameByID(cfg.CurrentGame)
	if err != nil {
		return err
	}

	mods, err := mod.LoadMods(cfg.CurrentGame)
	if err != nil {
		return err
	}

	for _, m := range mods {
		if m.Name == modName {
			m.Active = false
			if err := mod.SaveMods(mods, cfg.CurrentGame); err != nil {
				return err
			}

			archiveFiles, err := findArchiveFiles(m.Path, game.ArchiveExt)
			if err != nil {
				return err
			}
			for _, archiveFile := range archiveFiles {
				if err := config.RemoveArchiveFromCustomIniWithPrefix(archiveFile); err != nil {
					return err
				}
			}

			return SyncLinks()
		}
	}

	return fmt.Errorf("mod not found: %s", modName)
}
