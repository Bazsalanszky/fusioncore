package os

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
)

const desktopFileContent = `[Desktop Entry]
Name=Fusion Core
Exec=%s %%U
Type=Application
Terminal=false
MimeType=x-scheme-handler/nxm;
`

// RegisterProtocolHandler registers the application as a handler for nxm:// URLs.
func RegisterProtocolHandler() error {
	executablePath, err := os.Executable()
	if err != nil {
		return fmt.Errorf("failed to get executable path: %w", err)
	}

	desktopFile := fmt.Sprintf(desktopFileContent, executablePath)
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("failed to get home directory: %w", err)
	}

	applicationsDir := filepath.Join(homeDir, ".local", "share", "applications")
	if err := os.MkdirAll(applicationsDir, 0755); err != nil {
		return fmt.Errorf("failed to create applications directory: %w", err)
	}

	desktopFilePath := filepath.Join(applicationsDir, "fusion-core.desktop")
	if err := os.WriteFile(desktopFilePath, []byte(desktopFile), 0644); err != nil {
		return fmt.Errorf("failed to write desktop file: %w", err)
	}

	cmd := exec.Command("xdg-mime", "default", "fusion-core.desktop", "x-scheme-handler/nxm")
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to register protocol handler: %w", err)
	}

	fmt.Println("Registered nxm:// protocol handler.")
	return nil
}
