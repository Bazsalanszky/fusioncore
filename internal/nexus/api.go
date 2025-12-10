package nexus

import (
	"encoding/json"
	"fmt"
	"io"
	"mime"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

// DownloadLinkInfo holds the download link information from the Nexus Mods API.
type DownloadLinkInfo struct {
	Name string `json:"name"`
	URI  string `json:"URI"`
}

func ValidateAPIKey(apiKey string) (string, error) {
	req, err := http.NewRequest("GET", "https://api.nexusmods.com/v1/users/validate.json", nil)
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("apikey", apiKey)
	req.Header.Set("accept", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to perform request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("API key validation failed with status: %s", resp.Status)
	}

	// Find username from response
	var result struct {
		UserName string `json:"name"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", fmt.Errorf("failed to decode response: %w", err)
	}

	return result.UserName, nil
}

// GetDownloadURL gets the download URL for a mod file from the Nexus Mods API.
func GetDownloadURL(info *NxmInfo, apiKey string) (string, error) {
	if info == nil {
		return "", fmt.Errorf("NxmInfo is nil")
	}

	apiURL := fmt.Sprintf("https://api.nexusmods.com/v1/games/%s/mods/%s/files/%s/download_link.json?key=%s&expires=%s", info.Game, info.ModID, info.FileID, info.Key, info.Expires)

	fmt.Println("Api key:", apiURL)

	req, err := http.NewRequest("GET", apiURL, nil)
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("apikey", apiKey)
	req.Header.Set("accept", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to perform request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("API request failed with status: %s", resp.Status)
	}

	var downloadLinks []DownloadLinkInfo
	if err := json.NewDecoder(resp.Body).Decode(&downloadLinks); err != nil {
		return "", fmt.Errorf("failed to decode response: %w", err)
	}

	if len(downloadLinks) == 0 {
		return "", fmt.Errorf("no download links found")
	}

	// The API returns a list of download links, we'll use the first one.
	// The 'URI' field contains the actual download URL.
	return downloadLinks[0].URI, nil
}

// DownloadFile downloads a file from a URL to a specified directory.
func DownloadFile(url, destDir string, progressCb func(float64)) (string, error) {
	resp, err := http.Get(url)
	if err != nil {
		return "", fmt.Errorf("failed to download file: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("download failed with status: %s", resp.Status)
	}

	// Create the destination directory if it doesn't exist.
	if err := os.MkdirAll(destDir, 0755); err != nil {
		return "", fmt.Errorf("failed to create destination directory: %w", err)
	}

	// Get the filename from the Content-Disposition header.
	disposition := resp.Header.Get("Content-Disposition")
	_, params, err := mime.ParseMediaType(disposition)
	if err != nil {
		// If the header is not present or invalid, extract the filename from the URL path.
		_, file := filepath.Split(url)
		params = map[string]string{"filename": strings.Split(file, "?")[0]}
	}
	filename := strings.Split(params["filename"], "?")[0] // Remove any query parameters.

	filePath := filepath.Join(destDir, filename)
	out, err := os.Create(filePath)
	if err != nil {
		return "", fmt.Errorf("failed to create file: %w", err)
	}
	defer out.Close()

	// Create a progress writer to track download progress.
	progressWriter := &ProgressWriter{
		Writer:         out,
		Total:          resp.ContentLength,
		ProgressReport: progressCb,
	}

	// Copy the file content.
	_, err = io.Copy(progressWriter, resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to save file: %w", err)
	}

	return filePath, nil
}

// ProgressWriter is a wrapper around an io.Writer that reports progress.
type ProgressWriter struct {
	Writer         io.Writer
	Total          int64
	Written        int64
	ProgressReport func(float64)
}

// Write implements the io.Writer interface.
func (pw *ProgressWriter) Write(p []byte) (int, error) {
	n, err := pw.Writer.Write(p)
	if err != nil {
		return n, err
	}

	pw.Written += int64(n)
	if pw.ProgressReport != nil {
		pw.ProgressReport(float64(pw.Written) / float64(pw.Total))
	}

	return n, nil
}

