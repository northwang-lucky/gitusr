package hook

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadState_FileNotExist(t *testing.T) {
	tmpDir := t.TempDir()
	t.Setenv("XDG_DATA_HOME", tmpDir)

	state, err := LoadState()
	if err != nil {
		t.Fatalf("LoadState() returned error: %v", err)
	}
	if len(state.InstalledTypes) != 0 {
		t.Errorf("expected empty InstalledTypes, got %v", state.InstalledTypes)
	}
}

func TestLoadState_InvalidJSON(t *testing.T) {
	tmpDir := t.TempDir()
	t.Setenv("XDG_DATA_HOME", tmpDir)

	dataDir := filepath.Join(tmpDir, "gitusr")
	if err := os.MkdirAll(dataDir, 0755); err != nil {
		t.Fatalf("failed to create data dir: %v", err)
	}
	if err := os.WriteFile(filepath.Join(dataDir, "hook-state.json"), []byte("{invalid}"), 0644); err != nil {
		t.Fatalf("failed to write invalid state file: %v", err)
	}

	state, err := LoadState()
	if err != nil {
		t.Fatalf("LoadState() returned error: %v", err)
	}
	if len(state.InstalledTypes) != 0 {
		t.Errorf("expected empty InstalledTypes, got %v", state.InstalledTypes)
	}
}

func TestSaveAndLoad(t *testing.T) {
	tmpDir := t.TempDir()
	t.Setenv("XDG_DATA_HOME", tmpDir)

	state := &HookState{
		InstalledTypes: []HookType{HookTypeClone, HookTypeCommit},
	}

	if err := SaveState(state); err != nil {
		t.Fatalf("SaveState() returned error: %v", err)
	}

	loaded, err := LoadState()
	if err != nil {
		t.Fatalf("LoadState() returned error: %v", err)
	}

	if len(loaded.InstalledTypes) != 2 {
		t.Fatalf("expected 2 installed types, got %d", len(loaded.InstalledTypes))
	}
	if loaded.InstalledTypes[0] != HookTypeClone {
		t.Errorf("expected first type %q, got %q", HookTypeClone, loaded.InstalledTypes[0])
	}
	if loaded.InstalledTypes[1] != HookTypeCommit {
		t.Errorf("expected second type %q, got %q", HookTypeCommit, loaded.InstalledTypes[1])
	}
}

func TestIsInstalled_True(t *testing.T) {
	tmpDir := t.TempDir()
	t.Setenv("XDG_DATA_HOME", tmpDir)

	state := &HookState{
		InstalledTypes: []HookType{HookTypeClone},
	}
	if err := SaveState(state); err != nil {
		t.Fatalf("SaveState() returned error: %v", err)
	}

	installed, err := IsInstalled(HookTypeClone)
	if err != nil {
		t.Fatalf("IsInstalled() returned error: %v", err)
	}
	if !installed {
		t.Errorf("expected IsInstalled(%q) to be true", HookTypeClone)
	}
}

func TestIsInstalled_False(t *testing.T) {
	tmpDir := t.TempDir()
	t.Setenv("XDG_DATA_HOME", tmpDir)

	state := &HookState{
		InstalledTypes: []HookType{},
	}
	if err := SaveState(state); err != nil {
		t.Fatalf("SaveState() returned error: %v", err)
	}

	installed, err := IsInstalled(HookTypeCommit)
	if err != nil {
		t.Fatalf("IsInstalled() returned error: %v", err)
	}
	if installed {
		t.Errorf("expected IsInstalled(%q) to be false", HookTypeCommit)
	}
}
