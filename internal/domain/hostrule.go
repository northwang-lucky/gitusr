package domain

// HostRule maps a repository host (or a wildcard pattern) to a saved user
// email. Rules are stored as an ordered list: when multiple rules match the
// same clone URL, the one appearing earlier wins (first-match semantics).
type HostRule struct {
	Host  string `json:"host"`
	Email string `json:"email"`
}

// HostRuleStore persists the ordered host rule list used by the clone hook
// to pick a git user automatically.
type HostRuleStore interface {
	// List returns all host rules in configuration order.
	List() ([]HostRule, error)
	// SaveAll replaces the entire rule list, preserving the given order.
	SaveAll(rules []HostRule) error
}
