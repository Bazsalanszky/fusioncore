package nexus

import (
	"fmt"
	"net/url"
	"strings"
)

// NxmInfo holds the parsed information from an nxm URL.
type NxmInfo struct {
	Game   string
	ModID  string
	FileID string
	Key    string
	Expires string
}

// ParseNxmURL parses an nxm URL and returns the extracted information.
func ParseNxmURL(nxmURL string) (*NxmInfo, error) {
	parsedURL, err := url.Parse(nxmURL)
	if err != nil {
		return nil, fmt.Errorf("failed to parse nxm URL: %w", err)
	}

	if parsedURL.Scheme != "nxm" {
		return nil, fmt.Errorf("invalid URL scheme: expected 'nxm', got '%s'", parsedURL.Scheme)
	}

	pathParts := strings.Split(strings.Trim(parsedURL.Path, "/"), "/")
	if len(pathParts) < 4 || pathParts[0] != "mods" || pathParts[2] != "files" {
		return nil, fmt.Errorf("invalid nxm URL path: %s", parsedURL.Path)
	}

	query := parsedURL.Query()

	info := &NxmInfo{
		Game:    parsedURL.Host,
		ModID:   pathParts[1],
		FileID:  pathParts[3],
		Key:     query.Get("key"),
		Expires: query.Get("expires"),
	}

	return info, nil
}
