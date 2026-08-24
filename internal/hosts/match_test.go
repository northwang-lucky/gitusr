package hosts

import (
	"testing"

	"github.com/northwang-lucky/gitusr/internal/domain"
)

func TestValidateHost(t *testing.T) {
	valid := []string{
		"github.com",
		"code.byted.org",
		"*.byted.org",
		"gitlab.example.co.uk",
		"1.2.3.4",
		"UPPER.CASE", // 大小写不敏感,存储前统一小写化
	}
	for _, host := range valid {
		if err := ValidateHost(host); err != nil {
			t.Errorf("ValidateHost(%q) = %v, want nil", host, err)
		}
	}

	invalid := []string{
		"",
		"https://github.com",
		"github.com:443",
		"git@github.com",
		"*.byted.org/path",
		"*.byted.org:22",
		"foo bar",
	}
	for _, host := range invalid {
		if err := ValidateHost(host); err == nil {
			t.Errorf("ValidateHost(%q) = nil, want error", host)
		}
	}
}

func TestParseHostSchemes(t *testing.T) {
	cases := []struct {
		raw  string
		host string
		ok   bool
	}{
		{"https://github.com/northwang-lucky/gitusr.git", "github.com", true},
		{"http://code.byted.org/a/b.git", "code.byted.org", true},
		{"https://GITHUB.COM/a/b.git", "github.com", true},
		{"https://github.com:443/a/b.git", "github.com", true},
		{"ssh://git@github.com/a/b.git", "github.com", true},
		{"ssh://git@code.byted.org:2222/a/b.git", "code.byted.org", true},
		{"git://github.com/a/b.git", "github.com", true},
		{"git@github.com:a/b.git", "github.com", true},
		{"git@code.byted.org:a/b.git", "code.byted.org", true},
		{"code.byted.org:a/b.git", "code.byted.org", true},
		{"ssh://git@[::1]:22/a/b.git", "::1", true},
		{"", "", false},
		{"file:///data/repo.git", "", false},
		{"/data/repo.git", "", false},
		{"git@github.com", "", false},
	}
	for _, c := range cases {
		host, ok := ParseHost(c.raw)
		if host != c.host || ok != c.ok {
			t.Errorf("ParseHost(%q) = (%q, %v), want (%q, %v)", c.raw, host, ok, c.host, c.ok)
		}
	}
}

func TestMatchHost(t *testing.T) {
	rules := []domain.HostRule{
		{Host: "github.com", Email: "a@example.com"},
		{Host: "*.byted.org", Email: "b@byted.com"},
	}

	cases := []struct {
		raw   string
		email string
		ok    bool
	}{
		{"https://github.com/x/y.git", "a@example.com", true},
		{"git@github.com:x/y.git", "a@example.com", true},
		{"https://code.byted.org/x/y.git", "b@byted.com", true},
		{"ssh://git@dev.one.byted.org/x/y.git", "b@byted.com", true},
		{"https://example.com/x/y.git", "", false},
		{"file:///data/repo.git", "", false},
	}
	for _, c := range cases {
		rule, ok := MatchHost(rules, c.raw)
		if rule.Email != c.email || ok != c.ok {
			t.Errorf("MatchHost(%q) = (%q, %v), want (%q, %v)", c.raw, rule.Email, ok, c.email, c.ok)
		}
	}
}

// 精确匹配优先于通配匹配,即使通配规则先配置。
func TestMatchRuleExactBeatsWildcard(t *testing.T) {
	rules := []domain.HostRule{
		{Host: "*.byted.org", Email: "wildcard@byted.com"},
		{Host: "code.byted.org", Email: "exact@byted.com"},
	}

	rule, ok := MatchRule(rules, "code.byted.org")
	if !ok || rule.Email != "exact@byted.com" {
		t.Errorf("exact match should win, got (%q, %v)", rule.Email, ok)
	}
}

// 同类型通配匹配时,先配置者胜。
func TestMatchRuleWildcardOrder(t *testing.T) {
	rules := []domain.HostRule{
		{Host: "*.byted.org", Email: "first@byted.com"},
		{Host: "*.org", Email: "second@example.com"},
	}

	rule, ok := MatchRule(rules, "a.byted.org")
	if !ok || rule.Email != "first@byted.com" {
		t.Errorf("first configured wildcard should win, got (%q, %v)", rule.Email, ok)
	}
}

// 裸域规则只匹配自身,不匹配子域。
func TestMatchRuleBareHostDoesNotMatchSubdomains(t *testing.T) {
	rules := []domain.HostRule{
		{Host: "byted.org", Email: "bare@byted.com"},
	}

	_, ok := MatchRule(rules, "code.byted.org")
	if ok {
		t.Error("bare host rule should not match subdomains")
	}

	rule, ok := MatchRule(rules, "byted.org")
	if !ok || rule.Email != "bare@byted.com" {
		t.Error("bare host rule should match itself")
	}
}

func TestMatchRuleCaseInsensitive(t *testing.T) {
	rules := []domain.HostRule{
		{Host: "GITHUB.com", Email: "a@example.com"},
	}

	_, ok := MatchRule(rules, "github.com")
	if !ok {
		t.Error("matching should be case-insensitive")
	}
}
