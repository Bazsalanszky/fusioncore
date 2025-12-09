package extractor

import (
	"fmt"

	"github.com/gen2brain/go-unarr"
)

// Extract extracts an archive to a destination directory.
// It supports zip, tar, 7z, and rar formats.
func Extract(archivePath, destDir string) error {
	a, err := unarr.NewArchive(archivePath)
	if err != nil {
		return fmt.Errorf("failed to open archive: %w", err)
	}
	defer a.Close()

	_, err = a.Extract(destDir)
	if err != nil {
		return fmt.Errorf("failed to extract archive: %w", err)
	}

	return nil
}
