package xdgpath

import (
	"os"
	"path/filepath"
	"testing"
)

func TestDataFilePath_WithCustomEnv(t *testing.T) {
	tmpDir := t.TempDir()
	customDataHome := filepath.Join(tmpDir, "xdgdata")

	t.Setenv("XDG_DATA_HOME", customDataHome)

	got, err := DataFilePath()
	if err != nil {
		t.Fatalf("DataFilePath() returned error: %v", err)
	}

	expected := filepath.Join(customDataHome, "gitusr", "user-list.json")
	if got != expected {
		t.Errorf("expected %q, got %q", expected, got)
	}

	// Verify the parent directory was created
	parentDir := filepath.Join(customDataHome, "gitusr")
	info, err := os.Stat(parentDir)
	if err != nil {
		t.Fatalf("parent dir %q was not created: %v", parentDir, err)
	}
	if !info.IsDir() {
		t.Errorf("%q is not a directory", parentDir)
	}
}

func TestDataFilePath_DefaultPath(t *testing.T) {
	// Unset XDG_DATA_HOME so the fallback is exercised
	t.Setenv("XDG_DATA_HOME", "")

	// Ensure HOME is predictable for the test
	home := t.TempDir()
	t.Setenv("HOME", home)

	got, err := DataFilePath()
	if err != nil {
		t.Fatalf("DataFilePath() returned error: %v", err)
	}

	expected := filepath.Join(home, ".local", "share", "gitusr", "user-list.json")
	if got != expected {
		t.Errorf("expected %q, got %q", expected, got)
	}

	// Verify parent directory was created
	parentDir := filepath.Join(home, ".local", "share", "gitusr")
	info, err := os.Stat(parentDir)
	if err != nil {
		t.Fatalf("parent dir %q was not created: %v", parentDir, err)
	}
	if !info.IsDir() {
		t.Errorf("%q is not a directory", parentDir)
	}
}
