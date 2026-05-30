package domain

import (
	"encoding/json"
	"testing"
)

func TestUser_JSONMarshal(t *testing.T) {
	u := User{Name: "Alice", Email: "alice@example.com"}

	data, err := json.Marshal(u)
	if err != nil {
		t.Fatalf("json.Marshal failed: %v", err)
	}

	var result map[string]any
	if err := json.Unmarshal(data, &result); err != nil {
		t.Fatalf("json.Unmarshal failed: %v", err)
	}

	if result["name"] != "Alice" {
		t.Errorf("expected name='Alice', got %v", result["name"])
	}
	if result["email"] != "alice@example.com" {
		t.Errorf("expected email='alice@example.com', got %v", result["email"])
	}
}

func TestUser_JSONUnmarshal(t *testing.T) {
	raw := `{"name":"Bob","email":"bob@example.com"}`

	var u User
	if err := json.Unmarshal([]byte(raw), &u); err != nil {
		t.Fatalf("json.Unmarshal failed: %v", err)
	}

	if u.Name != "Bob" {
		t.Errorf("expected Name='Bob', got %q", u.Name)
	}
	if u.Email != "bob@example.com" {
		t.Errorf("expected Email='bob@example.com', got %q", u.Email)
	}
}
