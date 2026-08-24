package cli

import (
	"strings"
	"testing"

	"github.com/northwang-lucky/gitusr/internal/domain"
	"github.com/northwang-lucky/gitusr/internal/i18n"
)

func TestHostsSet_AddsRule(t *testing.T) {
	i18n.ResetForTesting()
	i18n.InitWithLocale("en")

	store := &mockStore{initialized: true, users: []domain.User{{Name: "Alice", Email: "alice@example.com"}}}
	hostStore := newTestHostStore(t)

	cmd := NewHostsCmd(store, hostStore)
	stdout, stderr, err := executeCmd(cmd, "set", "github.com", "alice@example.com")

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if stderr != "" {
		t.Errorf("expected empty stderr, got %q", stderr)
	}
	if !strings.Contains(stdout, "github.com") {
		t.Errorf("expected success message, got %q", stdout)
	}

	rules, err := hostStore.List()
	if err != nil {
		t.Fatalf("list hosts: %v", err)
	}
	if len(rules) != 1 || rules[0].Host != "github.com" || rules[0].Email != "alice@example.com" {
		t.Errorf("unexpected rules: %+v", rules)
	}
}

// host 大小写规范化 + 已存在 host 就地更新且顺序不变。
func TestHostsSet_UpdatesInPlace(t *testing.T) {
	i18n.ResetForTesting()
	i18n.InitWithLocale("en")

	store := &mockStore{initialized: true, users: []domain.User{{Name: "Alice", Email: "alice@example.com"}}}
	hostStore := newTestHostStore(t)
	if err := hostStore.SaveAll([]domain.HostRule{
		{Host: "github.com", Email: "a@example.com"},
		{Host: "code.byted.org", Email: "b@byted.com"},
	}); err != nil {
		t.Fatalf("save hosts: %v", err)
	}

	cmd := NewHostsCmd(store, hostStore)
	if _, _, err := executeCmd(cmd, "set", "GITHUB.com", "alice@example.com"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	rules, err := hostStore.List()
	if err != nil {
		t.Fatalf("list hosts: %v", err)
	}
	if len(rules) != 2 {
		t.Fatalf("expected 2 rules, got %d: %+v", len(rules), rules)
	}
	if rules[0].Host != "github.com" || rules[0].Email != "alice@example.com" {
		t.Errorf("first rule should be updated in place, got %+v", rules[0])
	}
	if rules[1].Host != "code.byted.org" {
		t.Errorf("second rule should keep its position, got %+v", rules[1])
	}
}

func TestHostsSet_InvalidHost(t *testing.T) {
	i18n.ResetForTesting()
	i18n.InitWithLocale("en")

	store := &mockStore{initialized: true, users: []domain.User{{Name: "Alice", Email: "alice@example.com"}}}
	hostStore := newTestHostStore(t)

	cmd := NewHostsCmd(store, hostStore)
	_, _, err := executeCmd(cmd, "set", "https://github.com/a/b.git", "alice@example.com")
	if err == nil {
		t.Fatal("expected error for URL host input")
	}
}

func TestHostsSet_EmailNotFound(t *testing.T) {
	i18n.ResetForTesting()
	i18n.InitWithLocale("en")

	store := &mockStore{initialized: true, users: []domain.User{{Name: "Alice", Email: "alice@example.com"}}}
	hostStore := newTestHostStore(t)

	cmd := NewHostsCmd(store, hostStore)
	_, _, err := executeCmd(cmd, "set", "github.com", "ghost@example.com")
	if err == nil {
		t.Fatal("expected error for unknown email")
	}
}

func TestHostsList_Empty(t *testing.T) {
	i18n.ResetForTesting()
	i18n.InitWithLocale("en")

	cmd := NewHostsCmd(&mockStore{}, newTestHostStore(t))
	stdout, _, err := executeCmd(cmd, "list")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(stdout, "No host rules configured") {
		t.Errorf("expected empty message, got %q", stdout)
	}
}

func TestHostsList_Rows(t *testing.T) {
	i18n.ResetForTesting()
	i18n.InitWithLocale("en")

	hostStore := newTestHostStore(t)
	if err := hostStore.SaveAll([]domain.HostRule{
		{Host: "github.com", Email: "a@example.com"},
		{Host: "*.byted.org", Email: "b@byted.com"},
	}); err != nil {
		t.Fatalf("save hosts: %v", err)
	}

	cmd := NewHostsCmd(&mockStore{}, hostStore)
	stdout, _, err := executeCmd(cmd, "list")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(stdout, "0: github.com") || !strings.Contains(stdout, "1: *.byted.org") {
		t.Errorf("expected numbered rows, got %q", stdout)
	}
}

func TestHostsRemove_RemovesRule(t *testing.T) {
	i18n.ResetForTesting()
	i18n.InitWithLocale("en")

	hostStore := newTestHostStore(t)
	if err := hostStore.SaveAll([]domain.HostRule{
		{Host: "github.com", Email: "a@example.com"},
		{Host: "code.byted.org", Email: "b@byted.com"},
	}); err != nil {
		t.Fatalf("save hosts: %v", err)
	}

	cmd := NewHostsCmd(&mockStore{}, hostStore)
	if _, _, err := executeCmd(cmd, "remove", "github.com"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	rules, err := hostStore.List()
	if err != nil {
		t.Fatalf("list hosts: %v", err)
	}
	if len(rules) != 1 || rules[0].Host != "code.byted.org" {
		t.Errorf("expected only code.byted.org left, got %+v", rules)
	}
}

func TestHostsRemove_NotFound(t *testing.T) {
	i18n.ResetForTesting()
	i18n.InitWithLocale("en")

	cmd := NewHostsCmd(&mockStore{}, newTestHostStore(t))
	_, _, err := executeCmd(cmd, "remove", "github.com")
	if err == nil {
		t.Fatal("expected error for unknown host")
	}
}

func TestHostsMove_Reorder(t *testing.T) {
	i18n.ResetForTesting()
	i18n.InitWithLocale("en")

	cases := []struct {
		host   string
		target string
		want   []string
	}{
		{"code.byted.org", "first", []string{"code.byted.org", "github.com", "gitlab.com"}},
		{"github.com", "last", []string{"gitlab.com", "code.byted.org", "github.com"}},
		{"code.byted.org", "up", []string{"github.com", "code.byted.org", "gitlab.com"}},
		{"github.com", "down", []string{"gitlab.com", "github.com", "code.byted.org"}},
		{"gitlab.com", "2", []string{"github.com", "code.byted.org", "gitlab.com"}},
	}
	for _, c := range cases {
		hostStore := newTestHostStore(t)
		if err := hostStore.SaveAll([]domain.HostRule{
			{Host: "github.com", Email: "a@example.com"},
			{Host: "gitlab.com", Email: "b@example.com"},
			{Host: "code.byted.org", Email: "c@byted.com"},
		}); err != nil {
			t.Fatalf("save hosts: %v", err)
		}

		cmd := NewHostsCmd(&mockStore{}, hostStore)
		if _, _, err := executeCmd(cmd, "move", c.host, c.target); err != nil {
			t.Fatalf("move %s %s failed: %v", c.host, c.target, err)
		}

		rules, err := hostStore.List()
		if err != nil {
			t.Fatalf("list hosts: %v", err)
		}
		var got []string
		for _, r := range rules {
			got = append(got, r.Host)
		}
		if strings.Join(got, ",") != strings.Join(c.want, ",") {
			t.Errorf("move %s %s: got %v, want %v", c.host, c.target, got, c.want)
		}
	}
}

func TestHostsMove_InvalidTarget(t *testing.T) {
	i18n.ResetForTesting()
	i18n.InitWithLocale("en")

	hostStore := newTestHostStore(t)
	if err := hostStore.SaveAll([]domain.HostRule{
		{Host: "github.com", Email: "a@example.com"},
		{Host: "gitlab.com", Email: "b@example.com"},
	}); err != nil {
		t.Fatalf("save hosts: %v", err)
	}

	cmd := NewHostsCmd(&mockStore{}, hostStore)
	for _, target := range []string{"left", "5", "-1"} {
		if _, _, err := executeCmd(cmd, "move", "github.com", target); err == nil {
			t.Errorf("expected error for target %q", target)
		}
	}
	// up 已在首位 → 边界错误
	if _, _, err := executeCmd(cmd, "move", "github.com", "up"); err == nil {
		t.Error("expected boundary error when moving up from first")
	}
	// down 已在末位 → 边界错误
	if _, _, err := executeCmd(cmd, "move", "gitlab.com", "down"); err == nil {
		t.Error("expected boundary error when moving down from last")
	}
}

func TestHostsMove_NotFound(t *testing.T) {
	i18n.ResetForTesting()
	i18n.InitWithLocale("en")

	cmd := NewHostsCmd(&mockStore{}, newTestHostStore(t))
	_, _, err := executeCmd(cmd, "move", "github.com", "first")
	if err == nil {
		t.Fatal("expected error for unknown host")
	}
}
