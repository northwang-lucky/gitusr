package hook

import (
	"encoding/json"
	"testing"
)

func TestHookState_DisabledTypes(t *testing.T) {
	// Create a HookState with both InstalledTypes and DisabledTypes
	original := &HookState{
		InstalledTypes: []HookType{HookTypeClone, HookTypeCommit},
		DisabledTypes:  []HookType{HookTypeCD},
	}

	// Marshal to JSON
	data, err := json.Marshal(original)
	if err != nil {
		t.Fatalf("json.Marshal() returned error: %v", err)
	}

	// Verify JSON contains the disabled_types field
	var raw map[string]json.RawMessage
	if err := json.Unmarshal(data, &raw); err != nil {
		t.Fatalf("json.Unmarshal to map returned error: %v", err)
	}

	if _, ok := raw["disabled_types"]; !ok {
		t.Error("expected JSON to contain 'disabled_types' key, but it was missing")
	}

	if _, ok := raw["installed_types"]; !ok {
		t.Error("expected JSON to contain 'installed_types' key, but it was missing")
	}

	// Unmarshal back
	restored := &HookState{}
	if err := json.Unmarshal(data, restored); err != nil {
		t.Fatalf("json.Unmarshal() returned error: %v", err)
	}

	// Verify equality
	if len(restored.InstalledTypes) != len(original.InstalledTypes) {
		t.Errorf("InstalledTypes length mismatch: got %d, want %d", len(restored.InstalledTypes), len(original.InstalledTypes))
	}
	for i, ht := range original.InstalledTypes {
		if restored.InstalledTypes[i] != ht {
			t.Errorf("InstalledTypes[%d]: got %q, want %q", i, restored.InstalledTypes[i], ht)
		}
	}

	if len(restored.DisabledTypes) != len(original.DisabledTypes) {
		t.Errorf("DisabledTypes length mismatch: got %d, want %d", len(restored.DisabledTypes), len(original.DisabledTypes))
	}
	for i, ht := range original.DisabledTypes {
		if restored.DisabledTypes[i] != ht {
			t.Errorf("DisabledTypes[%d]: got %q, want %q", i, restored.DisabledTypes[i], ht)
		}
	}
}

func TestHookState_DisabledTypes_Empty(t *testing.T) {
	// When DisabledTypes is nil, marshaling should produce empty array or omit
	original := &HookState{
		InstalledTypes: []HookType{HookTypeClone},
		// DisabledTypes is nil
	}

	data, err := json.Marshal(original)
	if err != nil {
		t.Fatalf("json.Marshal() returned error: %v", err)
	}

	restored := &HookState{}
	if err := json.Unmarshal(data, restored); err != nil {
		t.Fatalf("json.Unmarshal() returned error: %v", err)
	}

	if restored.DisabledTypes == nil {
		// nil is acceptable - no disabled types
	} else if len(restored.DisabledTypes) != 0 {
		t.Errorf("expected empty DisabledTypes, got %v", restored.DisabledTypes)
	}
}
