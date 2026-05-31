package hook

import (
	"errors"
	"fmt"

	"github.com/northwang-lucky/gitusr/internal/domain"
)

// GenerateEnv generates the environment script for the given shell type.
// It checks if any hooks are installed first and returns an error if none are installed.
func GenerateEnv(shell ShellType) (string, error) {
	state, err := LoadState()
	if err != nil {
		return "", err
	}

	if len(state.InstalledTypes) == 0 {
		return "", errors.New("请先执行 hook install")
	}

	switch shell {
	case ShellTypeBash:
		return GenerateBashEnv(), nil
	case ShellTypeZsh:
		return GenerateZshEnv(), nil
	default:
		return "", fmt.Errorf("unsupported shell type: %s", shell)
	}
}

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
