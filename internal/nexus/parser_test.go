package nexus

import (
	"testing"
)

func TestParseNxmURL(t *testing.T) {
	testURL := "nxm://fallout4/mods/12345/files/67890?key=my-secret-key&expires=1234567890"
	expected := &NxmInfo{
		Game:    "fallout4",
		ModID:   "12345",
		FileID:  "67890",
		Key:     "my-secret-key",
		Expires: "1234567890",
	}

	actual, err := ParseNxmURL(testURL)
	if err != nil {
		t.Fatalf("ParseNxmURL failed: %v", err)
	}

	if actual.Game != expected.Game {
		t.Errorf("expected game %q, got %q", expected.Game, actual.Game)
	}
	if actual.ModID != expected.ModID {
		t.Errorf("expected mod ID %q, got %q", expected.ModID, actual.ModID)
	}
	if actual.FileID != expected.FileID {
		t.Errorf("expected file ID %q, got %q", expected.FileID, actual.FileID)
	}
	if actual.Key != expected.Key {
		t.Errorf("expected key %q, got %q", expected.Key, actual.Key)
	}
	if actual.Expires != expected.Expires {
		t.Errorf("expected expires %q, got %q", expected.Expires, actual.Expires)
	}
}
