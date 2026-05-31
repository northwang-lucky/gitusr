package version

import "testing"

func TestVersionDefault(t *testing.T) {
	if Version != "dev" {
		t.Errorf("expected Version to be \"dev\", got %q", Version)
	}
}
