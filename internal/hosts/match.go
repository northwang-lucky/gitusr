// Package hosts implements the URL-to-host parsing and host rule matching
// used by the clone hook to pick a git user automatically.
//
// A host rule is either an exact hostname ("github.com") or a wildcard
// pattern ("*.byted.org") that matches any host under the given suffix.
// Rules are evaluated in configuration order: exact matches always beat
// wildcard matches; among wildcards, the first configured rule wins.
package hosts

import (
	"net/url"
	"strings"

	"github.com/northwang-lucky/gitusr/internal/domain"
)

// ParseHost extracts a normalized hostname (lowercase, no port) from a git
// clone URL. It supports https/http/ssh/git URLs and scp-like forms such as
// "git@host:path". It returns ok=false when no host can be determined, which
// the caller treats as "no rule applies".
func ParseHost(rawURL string) (host string, ok bool) {
	rawURL = strings.TrimSpace(rawURL)
	if rawURL == "" {
		return "", false
	}

	if strings.Contains(rawURL, "://") {
		u, err := url.Parse(rawURL)
		if err != nil {
			return "", false
		}
		host = strings.ToLower(u.Hostname())
		return host, host != ""
	}

	// scp-like form: [user@]host:path (no scheme, first colon splits host).
	if idx := strings.Index(rawURL, ":"); idx > 0 {
		prefix := rawURL[:idx]
		if !strings.Contains(prefix, "/") {
			host = prefix
			if at := strings.LastIndex(host, "@"); at >= 0 {
				host = host[at+1:]
			}
			host = strings.ToLower(host)
			return host, host != ""
		}
	}

	return "", false
}

// MatchHost finds the first rule matching the given clone URL.
func MatchHost(rules []domain.HostRule, rawURL string) (domain.HostRule, bool) {
	host, ok := ParseHost(rawURL)
	if !ok {
		return domain.HostRule{}, false
	}
	return MatchRule(rules, host)
}

// MatchRule finds the first rule matching the given normalized hostname.
// Exact matches take priority over wildcard matches; within each class the
// rule configured first wins.
func MatchRule(rules []domain.HostRule, host string) (domain.HostRule, bool) {
	host = strings.ToLower(host)

	// Exact pass: hostnames are unique in the rule list, so the first
	// exact match is also the only one.
	for _, rule := range rules {
		if strings.ToLower(rule.Host) == host {
			return rule, true
		}
	}

	// Wildcard pass: first configured wildcard whose suffix matches wins.
	for _, rule := range rules {
		suffix, ok := wildcardSuffix(rule.Host)
		if ok && strings.HasSuffix(host, suffix) {
			return rule, true
		}
	}

	return domain.HostRule{}, false
}

// wildcardSuffix converts a "*.example.com" pattern into its match suffix
// ".example.com". Non-wildcard patterns return ok=false.
func wildcardSuffix(pattern string) (string, bool) {
	lower := strings.ToLower(pattern)
	if strings.HasPrefix(lower, "*.") {
		return lower[1:], true
	}
	return "", false
}
