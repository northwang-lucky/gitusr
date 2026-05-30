package xdgpath

import (
	"os"
	"path/filepath"
)

const (
	dataDir  = "gitusr"
	dataFile = "user-list.json"
)

// DataFilePath returns the full path to the user-list.json data file.
// It reads $XDG_DATA_HOME environment variable; if not set, it falls back to
// ~/.local/share. The parent directory is created if it does not exist.
func DataFilePath() (string, error) {
	dataHome := os.Getenv("XDG_DATA_HOME")
	if dataHome == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			return "", err
		}
		dataHome = filepath.Join(home, ".local", "share")
	}

	dir := filepath.Join(dataHome, dataDir)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return "", err
	}

	return filepath.Join(dir, dataFile), nil
}
