package hosts

import (
	"errors"
	"regexp"
	"strings"
)

// hostPattern matches a bare hostname (two or more labels, e.g. "github.com")
// or a wildcard pattern (".*.example.com" style). IPv4 addresses happen to
// match the same shape, so they are accepted without a separate rule.
var hostPattern = regexp.MustCompile(`^(\*\.)?[a-z0-9]([a-z0-9-]*[a-z0-9])?(\.[a-z0-9]([a-z0-9-]*[a-z0-9])?)+$`)

// ValidateHost reports whether host is a valid rule value: a bare hostname,
// an IPv4 address, or a "*.suffix" wildcard. URLs, ports, paths and schemes
// are rejected — the rule value is the host itself, not a repository URL.
func ValidateHost(host string) error {
	if hostPattern.MatchString(strings.ToLower(host)) {
		return nil
	}
	return errors.New("host must be a bare hostname like github.com or a wildcard like *.example.com")
}
