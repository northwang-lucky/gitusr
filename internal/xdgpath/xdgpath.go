package xdgpath

import (
	"os"
	"path/filepath"
)

const (
	dataDir   = "gitusr"
	dataFile  = "user-list.json"
	hostsFile = "hosts.json"
)

// dataHomeDir returns the gitusr data directory ($XDG_DATA_HOME/gitusr or
// ~/.local/share/gitusr), creating it if absent.
func dataHomeDir() (string, error) {
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

	return dir, nil
}

// DataFilePath returns the full path to the user-list.json data file.
func DataFilePath() (string, error) {
	dir, err := dataHomeDir()
	if err != nil {
		return "", err
	}

	return filepath.Join(dir, dataFile), nil
}

// HostsFilePath returns the full path to the hosts.json routing file.
func HostsFilePath() (string, error) {
	dir, err := dataHomeDir()
	if err != nil {
		return "", err
	}

	return filepath.Join(dir, hostsFile), nil
}
