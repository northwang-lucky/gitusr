package hook

import (
	"github.com/northwang-lucky/gitusr/internal/domain"
)

// GenerateApplyRC reads .gitusrrc from the current directory and applies
// git config via store lookup. It is called by the shell hook wrapper.
// If no .gitusrrc file is found, it returns nil silently.
func GenerateApplyRC(store domain.UserStore) error {
	rc, err := ParseRC("")
	if err != nil {
		return err
	}

	if rc == nil {
		return nil
	}

	return MatchAndApplyRC(store, rc)
}
